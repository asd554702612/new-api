package controller

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetWeeklyQuotaStatus(c *gin.Context) {
	userId := c.GetInt("id")
	status, err := model.GetWeeklyQuotaStatus(userId, time.Now().Unix())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, status)
}

func ClaimWeeklyQuota(c *gin.Context) {
	userId := c.GetInt("id")
	claim, err := model.ClaimWeeklyQuota(userId, time.Now().Unix())
	if err != nil {
		common.ApiError(c, err)
		return
	}

	model.RecordLog(userId, model.LogTypeSystem, fmt.Sprintf("用户领取周额度，获得额度 %s", logger.LogQuota(claim.QuotaAwarded)))

	var user model.User
	newQuota := 0
	if err := model.DB.Select("quota").Where("id = ?", userId).First(&user).Error; err == nil {
		newQuota = user.Quota
	}

	common.ApiSuccess(c, gin.H{
		"quota_awarded":     claim.QuotaAwarded,
		"new_quota":         newQuota,
		"claimed_at":        claim.CreatedAt,
		"window_started_at": claim.WindowStart,
		"window_ends_at":    claim.WindowEnd,
	})
}
