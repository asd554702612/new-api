package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupStatusActivityTestDB(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false

	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.PaymentActivityConfig{}))
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
}

func TestGetStatusExposesPaymentActivityEnabledFlags(t *testing.T) {
	setupStatusActivityTestDB(t)
	require.NoError(t, model.UpdateLuckyWheelConfig(false, defaultStatusLuckyWheelConfig()))
	require.NoError(t, model.UpdateRechargeActivityConfig(true, defaultStatusRechargeActivityConfig()))

	router := gin.New()
	router.GET("/api/status", GetStatus)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/status", nil))

	var response struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Equal(t, false, response.Data["lucky_wheel_enabled"])
	require.Equal(t, true, response.Data["recharge_activity_enabled"])
}

func defaultStatusLuckyWheelConfig() model.LuckyWheelConfig {
	return model.LuckyWheelConfig{
		EligibleOrderTypes:  []string{model.PaymentOrderTypeBalance},
		MultiplierStep:      0.1,
		GlobalMaxMultiplier: 1,
		AmountTiers: []model.LuckyWheelAmountTier{
			{
				Id:            "default",
				Name:          "默认",
				MinAmount:     0,
				MinMultiplier: 0.1,
				MaxMultiplier: 1,
				DrawCount:     1,
			},
		},
	}
}

func defaultStatusRechargeActivityConfig() model.RechargeActivityConfig {
	return model.RechargeActivityConfig{
		EligibleOrderTypes: []string{model.PaymentOrderTypeBalance},
		Prizes: []model.RechargeActivityPrize{
			{
				Id:                "default",
				Name:              "默认奖品",
				RewardDescription: "人工发放",
				Probability:       100,
				Enabled:           true,
			},
		},
	}
}
