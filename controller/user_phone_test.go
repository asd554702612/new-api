package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type userPhoneAPIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func setupUserPhoneControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false
	common.BatchUpdateEnabled = false
	common.RegisterEnabled = true
	common.PasswordRegisterEnabled = true
	common.PasswordLoginEnabled = true
	common.EmailVerificationEnabled = false
	common.PhoneVerificationEnabled = true
	constant.GenerateDefaultToken = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Log{}, &model.TwoFA{}, &model.TwoFABackupCode{}))

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func newUserPhoneRouter(t *testing.T, userID int) *gin.Engine {
	t.Helper()
	router := gin.New()
	router.Use(sessions.Sessions("session", cookie.NewStore([]byte("user-phone-test"))))
	if userID != 0 {
		router.Use(func(c *gin.Context) {
			c.Set("id", userID)
			c.Set("role", common.RoleCommonUser)
			c.Next()
		})
	}
	return router
}

func performUserPhoneJSON(t *testing.T, router *gin.Engine, method string, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := common.Marshal(body)
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeUserPhoneAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder) userPhoneAPIResponse {
	t.Helper()
	var response userPhoneAPIResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	return response
}

func createPhoneTestUser(t *testing.T, username string, rawPassword string, phoneNumber string) *model.User {
	t.Helper()
	hashedPassword, err := common.Password2Hash(rawPassword)
	require.NoError(t, err)
	user := &model.User{
		Username:    username,
		Password:    hashedPassword,
		DisplayName: username,
		Status:      common.UserStatusEnabled,
		Role:        common.RoleCommonUser,
		PhoneNumber: common.NormalizePhoneNumber(phoneNumber, "86"),
		Group:       "default",
	}
	require.NoError(t, model.DB.Create(user).Error)
	return user
}

func TestRegisterRequiresPhoneNumber(t *testing.T) {
	setupUserPhoneControllerTestDB(t)
	router := newUserPhoneRouter(t, 0)
	router.POST("/api/user/register", Register)

	recorder := performUserPhoneJSON(t, router, http.MethodPost, "/api/user/register", gin.H{
		"username": "phoneuser",
		"password": "password123",
	})

	response := decodeUserPhoneAPIResponse(t, recorder)
	require.False(t, response.Success)
}

func TestRegisterStoresPhoneAfterVerification(t *testing.T) {
	setupUserPhoneControllerTestDB(t)
	router := newUserPhoneRouter(t, 0)
	router.POST("/api/user/register", Register)
	common.RegisterVerificationCodeWithKey("+8613800138000", "123456", common.PhoneVerificationPurposeRegister)

	recorder := performUserPhoneJSON(t, router, http.MethodPost, "/api/user/register", gin.H{
		"username":                "phoneuser",
		"password":                "password123",
		"phone_number":            "13800138000",
		"phone_verification_code": "123456",
	})

	response := decodeUserPhoneAPIResponse(t, recorder)
	require.True(t, response.Success, response.Message)
	var user model.User
	require.NoError(t, model.DB.Where("username = ?", "phoneuser").First(&user).Error)
	require.Equal(t, "+8613800138000", user.PhoneNumber)
}

func TestSMSLoginWithCorrectCodeSucceeds(t *testing.T) {
	setupUserPhoneControllerTestDB(t)
	createPhoneTestUser(t, "alice", "password123", "13800138000")
	router := newUserPhoneRouter(t, 0)
	router.POST("/api/user/login", Login)
	common.RegisterVerificationCodeWithKey("+8613800138000", "123456", common.PhoneVerificationPurposeLogin)

	recorder := performUserPhoneJSON(t, router, http.MethodPost, "/api/user/login", gin.H{
		"login_type":   "sms",
		"phone_number": "13800138000",
		"sms_code":     "123456",
	})

	response := decodeUserPhoneAPIResponse(t, recorder)
	require.True(t, response.Success, response.Message)
}

func TestUpdateSelfBindsPhoneWithVerificationCode(t *testing.T) {
	setupUserPhoneControllerTestDB(t)
	user := createPhoneTestUser(t, "alice", "password123", "")
	router := newUserPhoneRouter(t, user.Id)
	router.PUT("/api/user/self", UpdateSelf)
	common.RegisterVerificationCodeWithKey("+8613800138000", "123456", common.PhoneVerificationPurposeBind)

	recorder := performUserPhoneJSON(t, router, http.MethodPut, "/api/user/self", gin.H{
		"username":                "alice",
		"display_name":            "alice",
		"phone_number":            "13800138000",
		"phone_verification_code": "123456",
	})

	response := decodeUserPhoneAPIResponse(t, recorder)
	require.True(t, response.Success, response.Message)
	var updated model.User
	require.NoError(t, model.DB.Where("id = ?", user.Id).First(&updated).Error)
	require.Equal(t, "+8613800138000", updated.PhoneNumber)
}

func TestUpdateSelfChangesPasswordWithPhoneVerificationCode(t *testing.T) {
	setupUserPhoneControllerTestDB(t)
	user := createPhoneTestUser(t, "alice", "password123", "13800138000")
	router := newUserPhoneRouter(t, user.Id)
	router.PUT("/api/user/self", UpdateSelf)
	common.RegisterVerificationCodeWithKey("+8613800138000", "123456", common.PhoneVerificationPurposeChangePassword)

	recorder := performUserPhoneJSON(t, router, http.MethodPut, "/api/user/self", gin.H{
		"username":                "alice",
		"display_name":            "alice",
		"original_password":       "password123",
		"password":                "newpassword",
		"phone_verification_code": "123456",
	})

	response := decodeUserPhoneAPIResponse(t, recorder)
	require.True(t, response.Success, response.Message)
	var updated model.User
	require.NoError(t, model.DB.Where("id = ?", user.Id).First(&updated).Error)
	require.True(t, common.ValidatePasswordAndHash("newpassword", updated.Password))
}
