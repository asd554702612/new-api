package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service/authz"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupComplianceControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gin.SetMode(gin.TestMode)
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.RedisEnabled = false
	common.TurnstileCheckEnabled = false
	common.IsMasterNode = true
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{},
		&model.Log{},
		&model.PrivacyRequest{},
		&model.PublicFeedback{},
		&model.CasbinRule{},
		&model.AuthzRole{},
	))
	require.NoError(t, authz.Init(db))
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func newComplianceControllerRouter(user *model.User) *gin.Engine {
	router := gin.New()
	router.Use(sessions.Sessions("session", cookie.NewStore([]byte("compliance-test"))))
	if user != nil {
		router.Use(func(c *gin.Context) {
			c.Set("id", user.Id)
			c.Set("username", user.Username)
			c.Set("role", user.Role)
			c.Set("status", user.Status)
			c.Next()
		})
	}
	return router
}

func createComplianceTestUser(t *testing.T, username string, role int) *model.User {
	t.Helper()
	token := username + "-sensitive-token"
	user := &model.User{
		Username:    username,
		Password:    "hashed-password",
		DisplayName: username,
		Role:        role,
		Status:      common.UserStatusEnabled,
		Email:       username + "@example.com",
		PhoneNumber: "+8613800138000",
		Group:       "default",
		AffCode:     username + "-aff",
		AccessToken: &token,
	}
	require.NoError(t, model.DB.Create(user).Error)
	return user
}

func performComplianceJSON(t *testing.T, router *gin.Engine, method string, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := common.Marshal(body)
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeComplianceResponse(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var response map[string]any
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	return response
}

func TestGetPersonalInfoSnapshotDoesNotExposeSensitiveFields(t *testing.T) {
	setupComplianceControllerTestDB(t)
	user := createComplianceTestUser(t, "alice", common.RoleCommonUser)
	router := newComplianceControllerRouter(user)
	router.GET("/api/privacy/personal-info", GetPersonalInfoSnapshot)

	recorder := performComplianceJSON(t, router, http.MethodGet, "/api/privacy/personal-info", nil)
	response := decodeComplianceResponse(t, recorder)

	require.Equal(t, true, response["success"])
	data, ok := response["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "alice", data["username"])
	require.NotContains(t, data, "password")
	require.NotContains(t, data, "access_token")
}

func TestCreatePublicFeedbackAllowsAnonymousSubmission(t *testing.T) {
	setupComplianceControllerTestDB(t)
	router := newComplianceControllerRouter(nil)
	router.POST("/api/feedback", CreatePublicFeedback)

	recorder := performComplianceJSON(t, router, http.MethodPost, "/api/feedback", gin.H{
		"feedback_type": model.PublicFeedbackTypeComplaint,
		"title":         "Billing issue",
		"content":       "Please review this problem.",
		"contact_email": "visitor@example.com",
	})
	response := decodeComplianceResponse(t, recorder)

	require.Equal(t, true, response["success"])
	data, ok := response["data"].(map[string]any)
	require.True(t, ok)
	require.NotEmpty(t, data["tracking_code"])

	var saved model.PublicFeedback
	require.NoError(t, model.DB.First(&saved).Error)
	require.Equal(t, 0, saved.UserId)
	require.NotEmpty(t, saved.IpHash)
}

func TestTrackPublicFeedbackReturnsSanitizedRecord(t *testing.T) {
	setupComplianceControllerTestDB(t)
	feedback, err := model.CreatePublicFeedback(model.PublicFeedbackInput{
		UserId:       901,
		Username:     "alice",
		ContactName:  "Alice",
		ContactEmail: "alice@example.com",
		ContactPhone: "+8613800138000",
		FeedbackType: model.PublicFeedbackTypeComplaint,
		Title:        "Service issue",
		Content:      "Please resolve this issue.",
		IpHash:       "hash-sensitive",
	})
	require.NoError(t, err)
	_, err = model.UpdatePublicFeedbackStatus(feedback.Id, model.PublicFeedbackStatusResolved, 7, "admin", "resolved safely")
	require.NoError(t, err)

	router := newComplianceControllerRouter(nil)
	router.GET("/api/feedback/track/:tracking_code", TrackPublicFeedback)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/feedback/track/"+feedback.TrackingCode, nil)
	router.ServeHTTP(recorder, request)
	response := decodeComplianceResponse(t, recorder)

	require.Equal(t, true, response["success"])
	data, ok := response["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, feedback.TrackingCode, data["tracking_code"])
	require.Equal(t, model.PublicFeedbackStatusResolved, data["status"])
	require.Equal(t, "resolved safely", data["admin_note"])
	require.NotZero(t, data["handled_at"])
	require.NotContains(t, data, "user_id")
	require.NotContains(t, data, "username")
	require.NotContains(t, data, "contact_email")
	require.NotContains(t, data, "contact_phone")
	require.NotContains(t, data, "admin_id")
	require.NotContains(t, data, "ip_hash")
}

func TestGetMyFeedbackDetailRequiresOwner(t *testing.T) {
	setupComplianceControllerTestDB(t)
	alice := createComplianceTestUser(t, "alice", common.RoleCommonUser)
	bob := createComplianceTestUser(t, "bob", common.RoleCommonUser)
	ownFeedback, err := model.CreatePublicFeedback(model.PublicFeedbackInput{
		UserId:       alice.Id,
		Username:     alice.Username,
		FeedbackType: model.PublicFeedbackTypeFeedback,
		Title:        "Own feedback",
		Content:      "Visible to Alice.",
		IpHash:       "hash-a",
	})
	require.NoError(t, err)
	otherFeedback, err := model.CreatePublicFeedback(model.PublicFeedbackInput{
		UserId:       bob.Id,
		Username:     bob.Username,
		FeedbackType: model.PublicFeedbackTypeComplaint,
		Title:        "Other feedback",
		Content:      "Hidden from Alice.",
		IpHash:       "hash-b",
	})
	require.NoError(t, err)

	router := newComplianceControllerRouter(alice)
	router.GET("/api/feedback/my/:id", GetMyFeedbackDetail)

	ownRecorder := httptest.NewRecorder()
	ownRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/feedback/my/%d", ownFeedback.Id), nil)
	router.ServeHTTP(ownRecorder, ownRequest)
	ownResponse := decodeComplianceResponse(t, ownRecorder)
	require.Equal(t, true, ownResponse["success"])
	ownData, ok := ownResponse["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(ownFeedback.Id), ownData["id"])

	otherRecorder := httptest.NewRecorder()
	otherRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/feedback/my/%d", otherFeedback.Id), nil)
	router.ServeHTTP(otherRecorder, otherRequest)
	otherResponse := decodeComplianceResponse(t, otherRecorder)
	require.Equal(t, false, otherResponse["success"])
}

func TestAdminUpdatePrivacyRequestDoesNotDeleteRootUser(t *testing.T) {
	setupComplianceControllerTestDB(t)
	rootUser := createComplianceTestUser(t, "root-user", common.RoleRootUser)
	admin := createComplianceTestUser(t, "admin", common.RoleAdminUser)
	request, err := model.CreatePrivacyRequest(model.PrivacyRequestInput{
		UserId:      rootUser.Id,
		Username:    rootUser.Username,
		RequestType: model.PrivacyRequestTypeDeletion,
		Content:     "Please delete my account.",
	})
	require.NoError(t, err)

	router := newComplianceControllerRouter(admin)
	router.PATCH("/api/privacy/admin/requests/:id", AdminUpdatePrivacyRequest)

	recorder := performComplianceJSON(t, router, http.MethodPatch, fmt.Sprintf("/api/privacy/admin/requests/%d", request.Id), gin.H{
		"status":                   model.PrivacyRequestStatusCompleted,
		"execute_account_deletion": true,
	})
	response := decodeComplianceResponse(t, recorder)

	require.Equal(t, false, response["success"])
	_, err = model.GetUserById(rootUser.Id, false)
	require.NoError(t, err)
}

func TestAdminUpdatePrivacyRequestSoftDeletesCommonUser(t *testing.T) {
	setupComplianceControllerTestDB(t)
	targetUser := createComplianceTestUser(t, "target-user", common.RoleCommonUser)
	admin := createComplianceTestUser(t, "delete-admin", common.RoleAdminUser)
	request, err := model.CreatePrivacyRequest(model.PrivacyRequestInput{
		UserId:      targetUser.Id,
		Username:    targetUser.Username,
		RequestType: model.PrivacyRequestTypeDeletion,
		Content:     "Please delete my account.",
	})
	require.NoError(t, err)

	router := newComplianceControllerRouter(admin)
	router.PATCH("/api/privacy/admin/requests/:id", AdminUpdatePrivacyRequest)

	recorder := performComplianceJSON(t, router, http.MethodPatch, fmt.Sprintf("/api/privacy/admin/requests/%d", request.Id), gin.H{
		"status":                   model.PrivacyRequestStatusCompleted,
		"admin_note":               "deletion completed",
		"execute_account_deletion": true,
	})
	response := decodeComplianceResponse(t, recorder)

	require.Equal(t, true, response["success"])
	_, err = model.GetUserById(targetUser.Id, false)
	require.Error(t, err)

	updated, err := model.GetPrivacyRequestById(request.Id)
	require.NoError(t, err)
	require.Equal(t, model.PrivacyRequestStatusCompleted, updated.Status)
	require.True(t, updated.ExecuteAccountDeletion)
	require.NotZero(t, updated.HandledAt)
}

func TestAdminUpdatePrivacyRequestRollbackKeepsUserWhenUpdateInvalid(t *testing.T) {
	setupComplianceControllerTestDB(t)
	targetUser := createComplianceTestUser(t, "rollback-user", common.RoleCommonUser)
	admin := createComplianceTestUser(t, "rollback-admin", common.RoleAdminUser)
	request, err := model.CreatePrivacyRequest(model.PrivacyRequestInput{
		UserId:      targetUser.Id,
		Username:    targetUser.Username,
		RequestType: model.PrivacyRequestTypeDeletion,
		Content:     "Please delete my account.",
	})
	require.NoError(t, err)

	router := newComplianceControllerRouter(admin)
	router.PATCH("/api/privacy/admin/requests/:id", AdminUpdatePrivacyRequest)

	recorder := performComplianceJSON(t, router, http.MethodPatch, fmt.Sprintf("/api/privacy/admin/requests/%d", request.Id), gin.H{
		"status":                   model.PrivacyRequestStatusCompleted,
		"admin_note":               strings.Repeat("x", 2100),
		"execute_account_deletion": true,
	})
	response := decodeComplianceResponse(t, recorder)

	require.Equal(t, false, response["success"])
	_, err = model.GetUserById(targetUser.Id, false)
	require.NoError(t, err)

	unchanged, err := model.GetPrivacyRequestById(request.Id)
	require.NoError(t, err)
	require.Equal(t, model.PrivacyRequestStatusPending, unchanged.Status)
	require.False(t, unchanged.ExecuteAccountDeletion)
	require.Zero(t, unchanged.HandledAt)
}
