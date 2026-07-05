package router

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service/authz"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestComplianceRoutesUseExpectedPermissions(t *testing.T) {
	assertComplianceRoutePermission(t, privacyAdminPermissionRoutes, http.MethodGet, "/requests", authz.ComplianceRead, controller.AdminListPrivacyRequests)
	assertComplianceRoutePermission(t, privacyAdminPermissionRoutes, http.MethodGet, "/requests/:id", authz.ComplianceRead, controller.AdminGetPrivacyRequest)
	assertComplianceRoutePermission(t, privacyAdminPermissionRoutes, http.MethodPatch, "/requests/:id", authz.ComplianceWrite, controller.AdminUpdatePrivacyRequest)

	assertComplianceRoutePermission(t, feedbackAdminPermissionRoutes, http.MethodGet, "", authz.ComplianceRead, controller.AdminListFeedback)
	assertComplianceRoutePermission(t, feedbackAdminPermissionRoutes, http.MethodGet, "/:id", authz.ComplianceRead, controller.AdminGetFeedback)
	assertComplianceRoutePermission(t, feedbackAdminPermissionRoutes, http.MethodPatch, "/:id", authz.ComplianceWrite, controller.AdminUpdateFeedback)
}

func TestComplianceAdminRoutePermissionDeniedReturns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newComplianceRouterAuthzDB(t)
	require.NoError(t, authz.Init(db))
	require.NoError(t, authz.SetUserPermissions(99, authz.PermissionsMap{
		authz.ResourceCompliance: {
			authz.ActionRead: false,
		},
	}))

	engine := gin.New()
	engine.Use(sessions.Sessions("session", cookie.NewStore([]byte("compliance-router-test"))))
	engine.GET("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("username", "admin")
		session.Set("role", common.RoleAdminUser)
		session.Set("id", 99)
		session.Set("status", common.UserStatusEnabled)
		session.Set("group", "default")
		require.NoError(t, session.Save())
		c.Status(http.StatusNoContent)
	})
	api := engine.Group("/api")
	registerComplianceRoutes(api)

	loginRecorder := httptest.NewRecorder()
	loginRequest := httptest.NewRequest(http.MethodGet, "/login", nil)
	engine.ServeHTTP(loginRecorder, loginRequest)
	require.Equal(t, http.StatusNoContent, loginRecorder.Code)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/privacy/admin/requests", nil)
	request.Header.Set("New-Api-User", "99")
	for _, cookie := range loginRecorder.Result().Cookies() {
		request.AddCookie(cookie)
	}
	engine.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestComplianceRoutesRegisterWithoutConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	api := engine.Group("/api")

	require.NotPanics(t, func() {
		registerComplianceRoutes(api)
	})
}

func assertComplianceRoutePermission(t *testing.T, routes []permissionRoute, method string, path string, permission authz.Permission, handler any) {
	t.Helper()
	for _, route := range routes {
		if route.method == method && route.path == path {
			assert.Equal(t, permission, route.permission)
			assert.Equal(t, reflect.ValueOf(handler).Pointer(), reflect.ValueOf(route.handler).Pointer())
			return
		}
	}
	t.Fatalf("route %s %s not found", method, path)
}

func newComplianceRouterAuthzDB(t *testing.T) *gorm.DB {
	t.Helper()
	wasMaster := common.IsMasterNode
	common.IsMasterNode = true
	t.Cleanup(func() {
		common.IsMasterNode = wasMaster
	})

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&model.CasbinRule{}, &model.AuthzRole{}))
	return db
}
