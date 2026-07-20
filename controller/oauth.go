package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/oauth"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	casdoorIdentitySessionState      = "casdoor_identity_state"
	casdoorIdentitySessionUserID     = "casdoor_identity_user_id"
	casdoorIdentitySessionOIDCID     = "casdoor_identity_oidc_id"
	casdoorIdentitySessionReturnPath = "casdoor_identity_return_path"
	casdoorIdentityReturnToken       = "/console/token"
	casdoorIdentityReturnPersonal    = "/console/personal"
)

// providerParams returns map with Provider key for i18n templates
func providerParams(name string) map[string]any {
	return map[string]any{"Provider": name}
}

// GenerateOAuthCode generates a state code for OAuth CSRF protection
func GenerateOAuthCode(c *gin.Context) {
	session := sessions.Default(c)
	state := common.GetRandomString(12)
	affCode := c.Query("aff")
	if affCode != "" {
		session.Set("aff", affCode)
	}
	session.Set("oauth_state", state)
	err := session.Save()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    state,
	})
}

// HandleOAuth handles OAuth callback for all standard OAuth providers
func HandleOAuth(c *gin.Context) {
	providerName := c.Param("provider")
	provider := oauth.GetProvider(providerName)
	if provider == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": i18n.T(c, i18n.MsgOAuthUnknownProvider),
		})
		return
	}

	session := sessions.Default(c)

	// 1. Validate state (CSRF protection)
	state := c.Query("state")
	if state == "" || session.Get("oauth_state") == nil || state != session.Get("oauth_state").(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": i18n.T(c, i18n.MsgOAuthStateInvalid),
		})
		return
	}

	// 2. Check if user is already logged in (bind flow)
	username := session.Get("username")
	if username != nil {
		handleOAuthBind(c, provider)
		return
	}

	// 3. Check if provider is enabled
	if !provider.IsEnabled() {
		common.ApiErrorI18n(c, i18n.MsgOAuthNotEnabled, providerParams(provider.GetName()))
		return
	}

	// 4. Handle error from provider
	errorCode := c.Query("error")
	if errorCode != "" {
		errorDescription := c.Query("error_description")
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": errorDescription,
		})
		return
	}

	// 5. Exchange code for token
	code := c.Query("code")
	token, err := provider.ExchangeToken(c.Request.Context(), code, c)
	if err != nil {
		handleOAuthError(c, err)
		return
	}

	// 6. Get user info
	oauthUser, err := provider.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		handleOAuthError(c, err)
		return
	}

	var identity *service.CasdoorIdentity
	allowCreateWhenRegisterDisabled := shouldAllowCreateWhenRegisterDisabled(providerName)
	if shouldUseCasdoorIdentity(providerName) {
		identity, err = syncCasdoorIdentity(c, oauthUser.ProviderUserID)
		if err != nil {
			common.ApiErrorMsg(c, "实名状态同步失败，请稍后重试")
			return
		}
	}

	// 7. Find or create user
	user, err := findOrCreateOAuthUser(c, provider, oauthUser, session, allowCreateWhenRegisterDisabled)
	if err != nil {
		switch err.(type) {
		case *OAuthUserDeletedError:
			common.ApiErrorI18n(c, i18n.MsgOAuthUserDeleted)
		case *OAuthRegistrationDisabledError:
			common.ApiErrorI18n(c, i18n.MsgUserRegisterDisabled)
		default:
			common.ApiError(c, err)
		}
		return
	}

	if identity != nil {
		if err := user.UpdateIdentitySnapshot(identity.IsVerified, identity.AgeChecked, identity.IsOver16, common.GetTimestamp()); err != nil {
			common.ApiError(c, err)
			return
		}
		if !service.CanEnterCasdoorIdentityBusiness(identity) {
			redirectURL, err := prepareCasdoorIdentityRedirect(c, user, identity.UserID, casdoorIdentityReturnToken)
			if err != nil {
				common.ApiError(c, err)
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "",
				"data": gin.H{
					"action":       "identity_required",
					"redirect_url": redirectURL,
				},
			})
			return
		}
	}

	// 8. Check user status
	if user.Status != common.UserStatusEnabled {
		common.ApiErrorI18n(c, i18n.MsgOAuthUserBanned)
		return
	}

	// 9. Setup login
	setupLogin(user, c)
}

// handleOAuthBind handles binding OAuth account to existing user
func handleOAuthBind(c *gin.Context, provider oauth.Provider) {
	if !provider.IsEnabled() {
		common.ApiErrorI18n(c, i18n.MsgOAuthNotEnabled, providerParams(provider.GetName()))
		return
	}

	// Exchange code for token
	code := c.Query("code")
	token, err := provider.ExchangeToken(c.Request.Context(), code, c)
	if err != nil {
		handleOAuthError(c, err)
		return
	}

	// Get user info
	oauthUser, err := provider.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		handleOAuthError(c, err)
		return
	}

	// Check if this OAuth account is already bound (check both new ID and legacy ID)
	if provider.IsUserIDTaken(oauthUser.ProviderUserID) {
		common.ApiErrorI18n(c, i18n.MsgOAuthAlreadyBound, providerParams(provider.GetName()))
		return
	}
	// Also check legacy ID to prevent duplicate bindings during migration period
	if legacyID, ok := oauthUser.Extra["legacy_id"].(string); ok && legacyID != "" {
		if provider.IsUserIDTaken(legacyID) {
			common.ApiErrorI18n(c, i18n.MsgOAuthAlreadyBound, providerParams(provider.GetName()))
			return
		}
	}

	// Get current user from session
	session := sessions.Default(c)
	id := session.Get("id")
	user := model.User{Id: id.(int)}
	err = user.FillUserById()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Handle binding based on provider type
	if genericProvider, ok := provider.(*oauth.GenericOAuthProvider); ok {
		// Custom provider: use user_oauth_bindings table
		err = model.UpdateUserOAuthBinding(user.Id, genericProvider.GetProviderId(), oauthUser.ProviderUserID)
		if err != nil {
			common.ApiError(c, err)
			return
		}
	} else {
		// Built-in provider: update user record directly
		provider.SetProviderUserID(&user, oauthUser.ProviderUserID)
		err = user.Update(false)
		if err != nil {
			common.ApiError(c, err)
			return
		}
	}

	common.ApiSuccessI18n(c, i18n.MsgOAuthBindSuccess, gin.H{
		"action": "bind",
	})
}

// findOrCreateOAuthUser finds existing user or creates new user
func findOrCreateOAuthUser(c *gin.Context, provider oauth.Provider, oauthUser *oauth.OAuthUser, session sessions.Session, allowCreateWhenRegisterDisabled bool) (*model.User, error) {
	user := &model.User{}

	// Check if user already exists with new ID
	if provider.IsUserIDTaken(oauthUser.ProviderUserID) {
		err := provider.FillUserByProviderID(user, oauthUser.ProviderUserID)
		if err != nil {
			return nil, err
		}
		// Check if user has been deleted
		if user.Id == 0 {
			return nil, &OAuthUserDeletedError{}
		}
		return user, nil
	}

	// Try to find user with legacy ID (for GitHub migration from login to numeric ID)
	if legacyID, ok := oauthUser.Extra["legacy_id"].(string); ok && legacyID != "" {
		if provider.IsUserIDTaken(legacyID) {
			err := provider.FillUserByProviderID(user, legacyID)
			if err != nil {
				return nil, err
			}
			if user.Id != 0 {
				// Found user with legacy ID, migrate to new ID
				common.SysLog(fmt.Sprintf("[OAuth] Migrating user %d from legacy_id=%s to new_id=%s",
					user.Id, legacyID, oauthUser.ProviderUserID))
				if err := user.UpdateGitHubId(oauthUser.ProviderUserID); err != nil {
					common.SysError(fmt.Sprintf("[OAuth] Failed to migrate user %d: %s", user.Id, err.Error()))
					// Continue with login even if migration fails
				}
				return user, nil
			}
		}
	}

	// User doesn't exist, create new user if registration is enabled
	if !common.RegisterEnabled && !allowCreateWhenRegisterDisabled {
		return nil, &OAuthRegistrationDisabledError{}
	}

	// Set up new user
	user.Username = provider.GetProviderPrefix() + strconv.Itoa(model.GetMaxUserId()+1)

	if oauthUser.Username != "" {
		if exists, err := model.CheckUserExistOrDeleted(oauthUser.Username, ""); err == nil && !exists {
			// 防止索引退化
			if len(oauthUser.Username) <= model.UserNameMaxLength {
				user.Username = oauthUser.Username
			}
		}
	}

	if oauthUser.DisplayName != "" {
		user.DisplayName = oauthUser.DisplayName
	} else if oauthUser.Username != "" {
		user.DisplayName = oauthUser.Username
	} else {
		user.DisplayName = provider.GetName() + " User"
	}
	if oauthUser.Email != "" {
		user.Email = oauthUser.Email
	}
	user.Role = common.RoleCommonUser
	user.Status = common.UserStatusEnabled

	// Handle affiliate code
	var affCode any
	if session != nil {
		affCode = session.Get("aff")
	}
	inviterId := 0
	if affCode != nil {
		inviterId, _ = model.GetUserIdByAffCode(affCode.(string))
	}

	// Use transaction to ensure user creation and OAuth binding are atomic
	if genericProvider, ok := provider.(*oauth.GenericOAuthProvider); ok {
		// Custom provider: create user and binding in a transaction
		err := model.DB.Transaction(func(tx *gorm.DB) error {
			// Create user
			if err := user.InsertWithTx(tx, inviterId); err != nil {
				return err
			}

			// Create OAuth binding
			binding := &model.UserOAuthBinding{
				UserId:         user.Id,
				ProviderId:     genericProvider.GetProviderId(),
				ProviderUserId: oauthUser.ProviderUserID,
			}
			if err := model.CreateUserOAuthBindingWithTx(tx, binding); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		// Perform post-transaction tasks (logs, sidebar config, inviter rewards)
		user.FinalizeOAuthUserCreation(inviterId)
	} else {
		// Built-in provider: create user and update provider ID in a transaction
		err := model.DB.Transaction(func(tx *gorm.DB) error {
			// Create user
			if err := user.InsertWithTx(tx, inviterId); err != nil {
				return err
			}

			// Set the provider user ID on the user model and update
			provider.SetProviderUserID(user, oauthUser.ProviderUserID)
			if err := tx.Model(user).Updates(map[string]interface{}{
				"github_id":   user.GitHubId,
				"discord_id":  user.DiscordId,
				"oidc_id":     user.OidcId,
				"linux_do_id": user.LinuxDOId,
				"wechat_id":   user.WeChatId,
				"telegram_id": user.TelegramId,
			}).Error; err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		// Perform post-transaction tasks
		user.FinalizeOAuthUserCreation(inviterId)
	}

	return user, nil
}

func shouldUseCasdoorIdentity(providerName string) bool {
	return providerName == "oidc" && setting.CasdoorIdentityEnabled
}

func shouldAllowCreateWhenRegisterDisabled(providerName string) bool {
	return providerName == "oidc"
}

func syncCasdoorIdentity(c *gin.Context, casdoorUserID string) (*service.CasdoorIdentity, error) {
	client := newCasdoorIdentityClientForRequest(c)
	return client.SyncUser(c.Request.Context(), casdoorUserID)
}

func newCasdoorIdentityClientForRequest(c *gin.Context) *service.CasdoorIdentityClient {
	clientID := strings.TrimSpace(setting.CasdoorClientID)
	clientSecret := strings.TrimSpace(setting.CasdoorClientSecret)
	if clientID == "" || clientSecret == "" {
		if resolved, ok := system_setting.GetOIDCSettings().ResolveClientForRequest(c.Request); ok {
			clientID = resolved.ClientId
			clientSecret = resolved.ClientSecret
		}
	}
	return service.NewCasdoorIdentityClient(setting.GetCasdoorBaseURL(), clientID, clientSecret)
}

func normalizeCasdoorIdentityReturnPath(returnPath string) string {
	switch returnPath {
	case casdoorIdentityReturnPersonal:
		return casdoorIdentityReturnPersonal
	default:
		return casdoorIdentityReturnToken
	}
}

func casdoorIdentitySnapshotData(user *model.User) gin.H {
	return gin.H{
		"identity_verified":    user.IdentityVerified,
		"identity_age_checked": user.IdentityAgeChecked,
		"identity_over16":      user.IdentityOver16,
		"identity_synced_at":   user.IdentitySyncedAt,
	}
}

func casdoorIdentityActionData(action string, user *model.User) gin.H {
	data := casdoorIdentitySnapshotData(user)
	data["action"] = action
	return data
}

func prepareCasdoorIdentityRedirect(c *gin.Context, user *model.User, casdoorUserID string, returnPath string) (string, error) {
	state := common.GetRandomString(32)
	session := sessions.Default(c)
	session.Set(casdoorIdentitySessionState, state)
	session.Set(casdoorIdentitySessionUserID, user.Id)
	session.Set(casdoorIdentitySessionOIDCID, casdoorUserID)
	session.Set(casdoorIdentitySessionReturnPath, normalizeCasdoorIdentityReturnPath(returnPath))
	if err := session.Save(); err != nil {
		return "", err
	}
	origin := system_setting.OIDCRequestOrigin(c.Request)
	redirectURI := setting.GetCasdoorIdentityCallbackURL(origin)
	client := newCasdoorIdentityClientForRequest(c)
	return client.BuildVerificationURL(casdoorUserID, redirectURI, state)
}

func getCurrentCasdoorIdentityUser(c *gin.Context) (*model.User, bool) {
	if !setting.CasdoorIdentityEnabled {
		common.ApiErrorMsg(c, "实名认证未启用")
		return nil, false
	}
	user, err := model.GetUserById(c.GetInt("id"), true)
	if err != nil {
		common.ApiError(c, err)
		return nil, false
	}
	if user.OidcId == "" {
		common.ApiErrorMsg(c, "请先绑定登录中心账号")
		return nil, false
	}
	return user, true
}

func syncCurrentCasdoorIdentity(c *gin.Context, user *model.User) (*service.CasdoorIdentity, bool) {
	identity, err := syncCasdoorIdentity(c, user.OidcId)
	if err != nil {
		common.SysLog(fmt.Sprintf("Casdoor identity sync failed for user %d: %v", user.Id, err))
		common.ApiErrorMsg(c, "实名状态同步失败，请稍后重试")
		return nil, false
	}
	if err := user.UpdateIdentitySnapshot(identity.IsVerified, identity.AgeChecked, identity.IsOver16, common.GetTimestamp()); err != nil {
		common.ApiError(c, err)
		return nil, false
	}
	return identity, true
}

func HandleUserIdentitySync(c *gin.Context) {
	user, ok := getCurrentCasdoorIdentityUser(c)
	if !ok {
		return
	}
	if _, ok = syncCurrentCasdoorIdentity(c, user); !ok {
		return
	}
	common.ApiSuccess(c, casdoorIdentitySnapshotData(user))
}

func HandleUserIdentityVerification(c *gin.Context) {
	user, ok := getCurrentCasdoorIdentityUser(c)
	if !ok {
		return
	}
	identity, ok := syncCurrentCasdoorIdentity(c, user)
	if !ok {
		return
	}
	if service.CanEnterCasdoorIdentityBusiness(identity) {
		common.ApiSuccess(c, casdoorIdentityActionData("verified", user))
		return
	}

	redirectURL, err := prepareCasdoorIdentityRedirect(c, user, user.OidcId, casdoorIdentityReturnPersonal)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	data := casdoorIdentityActionData("identity_required", user)
	data["redirect_url"] = redirectURL
	common.ApiSuccess(c, data)
}

func HandleCasdoorIdentityCallback(c *gin.Context) {
	session := sessions.Default(c)
	state := c.Query("state")
	savedState, _ := session.Get(casdoorIdentitySessionState).(string)
	if state == "" || savedState == "" || state != savedState {
		c.String(http.StatusBadRequest, "invalid identity state")
		return
	}
	userID, ok := session.Get(casdoorIdentitySessionUserID).(int)
	if !ok || userID == 0 {
		c.String(http.StatusBadRequest, "invalid identity session")
		return
	}
	casdoorUserID, _ := session.Get(casdoorIdentitySessionOIDCID).(string)
	if casdoorUserID == "" {
		c.String(http.StatusBadRequest, "invalid identity session")
		return
	}

	identity, err := syncCasdoorIdentity(c, casdoorUserID)
	if err != nil {
		common.ApiErrorMsg(c, "实名状态同步失败，请稍后重试")
		return
	}
	user, err := model.GetUserById(userID, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user.OidcId != casdoorUserID {
		c.String(http.StatusBadRequest, "invalid identity session")
		return
	}
	if err := user.UpdateIdentitySnapshot(identity.IsVerified, identity.AgeChecked, identity.IsOver16, common.GetTimestamp()); err != nil {
		common.ApiError(c, err)
		return
	}
	if user.Status != common.UserStatusEnabled {
		common.ApiErrorI18n(c, i18n.MsgOAuthUserBanned)
		return
	}
	if !service.CanEnterCasdoorIdentityBusiness(identity) {
		c.Redirect(http.StatusFound, "/identity-required")
		return
	}

	session.Delete(casdoorIdentitySessionState)
	session.Delete(casdoorIdentitySessionUserID)
	session.Delete(casdoorIdentitySessionOIDCID)
	returnPath := normalizeCasdoorIdentityReturnPath("")
	if savedReturnPath, _ := session.Get(casdoorIdentitySessionReturnPath).(string); savedReturnPath != "" {
		returnPath = normalizeCasdoorIdentityReturnPath(savedReturnPath)
	}
	session.Delete(casdoorIdentitySessionReturnPath)
	if err := establishLoginSession(user, c); err != nil {
		common.ApiErrorI18n(c, i18n.MsgUserSessionSaveFailed)
		return
	}
	c.Redirect(http.StatusFound, common.ThemeAwarePath(returnPath))
}

func IdentityRequired(c *gin.Context) {
	c.String(http.StatusOK, "实名认证未完成或年龄不满足要求，请返回登录中心完成实名认证后重试。")
}

// Error types for OAuth
type OAuthUserDeletedError struct{}

func (e *OAuthUserDeletedError) Error() string {
	return "user has been deleted"
}

type OAuthRegistrationDisabledError struct{}

func (e *OAuthRegistrationDisabledError) Error() string {
	return "registration is disabled"
}

// handleOAuthError handles OAuth errors and returns translated message
func handleOAuthError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *oauth.OAuthError:
		if e.Params != nil {
			common.ApiErrorI18n(c, e.MsgKey, e.Params)
		} else {
			common.ApiErrorI18n(c, e.MsgKey)
		}
	case *oauth.AccessDeniedError:
		common.ApiErrorMsg(c, e.Message)
	case *oauth.TrustLevelError:
		common.ApiErrorI18n(c, i18n.MsgOAuthTrustLevelLow)
	default:
		common.ApiError(c, err)
	}
}
