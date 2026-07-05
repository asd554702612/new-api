package controller

import (
	"errors"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type privacyRequestCreateRequest struct {
	Type         string `json:"type"`
	RequestType  string `json:"request_type"`
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
	Content      string `json:"content"`
}

type privacyRequestAdminUpdateRequest struct {
	Status                 string `json:"status"`
	AdminNote              string `json:"admin_note"`
	ExecuteAccountDeletion bool   `json:"execute_account_deletion"`
}

func GetPersonalInfoSnapshot(c *gin.Context) {
	user, err := model.GetUserById(c.GetInt("id"), false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"id":                user.Id,
		"username":          user.Username,
		"display_name":      user.DisplayName,
		"role":              user.Role,
		"status":            user.Status,
		"email":             user.Email,
		"phone_number":      user.PhoneNumber,
		"github_id":         user.GitHubId,
		"discord_id":        user.DiscordId,
		"oidc_id":           user.OidcId,
		"wechat_id":         user.WeChatId,
		"telegram_id":       user.TelegramId,
		"group":             user.Group,
		"quota":             user.Quota,
		"used_quota":        user.UsedQuota,
		"request_count":     user.RequestCount,
		"aff_code":          user.AffCode,
		"aff_count":         user.AffCount,
		"aff_quota":         user.AffQuota,
		"aff_history_quota": user.AffHistoryQuota,
		"inviter_id":        user.InviterId,
		"linux_do_id":       user.LinuxDOId,
		"stripe_customer":   user.StripeCustomer,
		"created_at":        user.CreatedAt,
		"last_login_at":     user.LastLoginAt,
	})
}

func ListMyPrivacyRequests(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	records, total, err := model.ListUserPrivacyRequests(c.GetInt("id"), pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}

func CreatePrivacyRequest(c *gin.Context) {
	var req privacyRequestCreateRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	requestType := req.RequestType
	if requestType == "" {
		requestType = req.Type
	}
	record, err := model.CreatePrivacyRequest(model.PrivacyRequestInput{
		UserId:       c.GetInt("id"),
		Username:     c.GetString("username"),
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		RequestType:  requestType,
		Content:      req.Content,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, record)
}

func CancelMyPrivacyRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	record, err := model.CancelUserPrivacyRequest(id, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, record)
}

func AdminListPrivacyRequests(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	filter := model.PrivacyRequestFilter{
		Status:      c.Query("status"),
		RequestType: c.Query("request_type"),
	}
	if filter.RequestType == "" {
		filter.RequestType = c.Query("type")
	}
	if userIdStr := c.Query("user_id"); userIdStr != "" {
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			common.ApiErrorI18n(c, i18n.MsgInvalidParams)
			return
		}
		filter.UserId = userId
	}
	records, total, err := model.ListPrivacyRequests(pageInfo, filter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}

func AdminGetPrivacyRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	record, err := model.GetPrivacyRequestById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, record)
}

func AdminUpdatePrivacyRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	var req privacyRequestAdminUpdateRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	updated, err := model.UpdatePrivacyRequestByAdminAndDeleteUser(id, model.PrivacyRequestAdminUpdate{
		Status:                 req.Status,
		AdminId:                c.GetInt("id"),
		AdminName:              c.GetString("username"),
		AdminNote:              req.AdminNote,
		ExecuteAccountDeletion: req.ExecuteAccountDeletion,
	})
	if err != nil {
		if errors.Is(err, model.ErrPrivacyRequestRootDeletion) {
			common.ApiErrorI18n(c, i18n.MsgUserCannotDeleteRootUser)
			return
		}
		common.ApiError(c, err)
		return
	}
	recordManageAuditFor(c, updated.UserId, "compliance.privacy_update", map[string]interface{}{
		"id":           updated.Id,
		"request_type": updated.RequestType,
		"status":       updated.Status,
	})
	common.ApiSuccess(c, updated)
}
