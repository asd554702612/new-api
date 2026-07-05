package router

import (
	"net/http"

	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/service/authz"
	"github.com/gin-gonic/gin"
)

var privacyAdminPermissionRoutes = []permissionRoute{
	{method: http.MethodGet, path: "/requests", permission: authz.ComplianceRead, handler: controller.AdminListPrivacyRequests},
	{method: http.MethodGet, path: "/requests/:id", permission: authz.ComplianceRead, handler: controller.AdminGetPrivacyRequest},
	{method: http.MethodPatch, path: "/requests/:id", permission: authz.ComplianceWrite, handler: controller.AdminUpdatePrivacyRequest},
}

var feedbackAdminPermissionRoutes = []permissionRoute{
	{method: http.MethodGet, path: "", permission: authz.ComplianceRead, handler: controller.AdminListFeedback},
	{method: http.MethodGet, path: "/:id", permission: authz.ComplianceRead, handler: controller.AdminGetFeedback},
	{method: http.MethodPatch, path: "/:id", permission: authz.ComplianceWrite, handler: controller.AdminUpdateFeedback},
}

func registerComplianceRoutes(apiRouter *gin.RouterGroup) {
	anonymousRequestBodyLimit := middleware.AnonymousRequestBodyLimit()

	privacyRoute := apiRouter.Group("/privacy")
	{
		privacyRoute.GET("/personal-info", middleware.UserAuth(), controller.GetPersonalInfoSnapshot)
		privacyRoute.GET("/requests", middleware.UserAuth(), controller.ListMyPrivacyRequests)
		privacyRoute.POST("/requests", middleware.UserAuth(), middleware.CriticalRateLimit(), controller.CreatePrivacyRequest)
		privacyRoute.POST("/requests/:id/cancel", middleware.UserAuth(), controller.CancelMyPrivacyRequest)

		adminRoute := privacyRoute.Group("/admin")
		adminRoute.Use(middleware.AdminAuth())
		for _, route := range privacyAdminPermissionRoutes {
			adminRoute.Handle(route.method, route.path,
				middleware.RequirePermission(route.permission),
				route.handler,
			)
		}
	}

	feedbackRoute := apiRouter.Group("/feedback")
	{
		feedbackRoute.POST("", anonymousRequestBodyLimit, middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.CreatePublicFeedback)
		feedbackRoute.GET("/my", middleware.UserAuth(), controller.ListMyFeedback)

		adminRoute := feedbackRoute.Group("/admin")
		adminRoute.Use(middleware.AdminAuth())
		for _, route := range feedbackAdminPermissionRoutes {
			adminRoute.Handle(route.method, route.path,
				middleware.RequirePermission(route.permission),
				route.handler,
			)
		}
	}
}
