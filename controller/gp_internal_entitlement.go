package controller

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type GpInternalEntitlementUserDTO struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Quota     int    `json:"quota"`
	UsedQuota int    `json:"used_quota"`
	Status    int    `json:"status"`
	Group     string `json:"group"`
}

type GpInternalEntitlementDTO struct {
	User              GpInternalEntitlementUserDTO `json:"user"`
	BillingPreference string                       `json:"billing_preference"`
	Subscriptions     struct {
		Active []model.SubscriptionSummary `json:"active"`
		All    []model.SubscriptionSummary `json:"all"`
	} `json:"subscriptions"`
	Plans      []SubscriptionPlanDTO `json:"plans"`
	SnapshotAt int64                 `json:"snapshot_at"`
}

func requireGpInternalSecret(c *gin.Context) bool {
	expected := strings.TrimSpace(os.Getenv("GP_INTERNAL_SHARED_SECRET"))
	if expected == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "GP internal secret is not configured",
		})
		return false
	}

	actual := strings.TrimSpace(c.GetHeader("X-GP-Internal-Secret"))
	if actual == "" {
		auth := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			actual = strings.TrimSpace(auth[len("bearer "):])
		}
	}
	if actual != expected {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "GP internal secret is missing or invalid",
		})
		return false
	}
	return true
}

func GetGpInternalEntitlement(c *gin.Context) {
	if !requireGpInternalSecret(c) {
		return
	}

	userID, err := strconv.Atoi(strings.TrimSpace(c.Param("user_id")))
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid user_id",
		})
		return
	}

	var user model.User
	if err := model.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		common.ApiError(c, err)
		return
	}

	settingMap, _ := model.GetUserSetting(userID, true)
	activeSubscriptions, err := model.GetAllActiveUserSubscriptions(userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	allSubscriptions, err := model.GetAllUserSubscriptions(userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	var plans []model.SubscriptionPlan
	if err := model.DB.Where("enabled = ?", true).Order("sort_order desc, id desc").Find(&plans).Error; err != nil {
		common.ApiError(c, err)
		return
	}

	now := time.Now()
	planDTOs := make([]SubscriptionPlanDTO, 0, len(plans))
	for _, plan := range plans {
		plan.NormalizeDefaults()
		availability, err := model.GetSubscriptionPlanSaleAvailability(userID, &plan, now)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		planDTOs = append(planDTOs, SubscriptionPlanDTO{
			Plan:             plan,
			SaleAvailability: availability,
		})
	}

	payload := GpInternalEntitlementDTO{
		User: GpInternalEntitlementUserDTO{
			ID:        user.Id,
			Username:  user.Username,
			Quota:     user.Quota,
			UsedQuota: user.UsedQuota,
			Status:    user.Status,
			Group:     user.Group,
		},
		BillingPreference: common.NormalizeBillingPreference(settingMap.BillingPreference),
		Plans:             planDTOs,
		SnapshotAt:        common.GetTimestamp(),
	}
	payload.Subscriptions.Active = activeSubscriptions
	payload.Subscriptions.All = allSubscriptions

	common.ApiSuccess(c, payload)
}
