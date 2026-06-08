package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetUserRankingsRejectsInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/rankings/users?limit=bad", nil)

	GetUserRankings(ctx)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var payload struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
	require.False(t, payload.Success)
	require.Contains(t, payload.Message, "invalid leaderboard limit")
}

func TestGetUserRankingsReturnsLeaderboardPayload(t *testing.T) {
	setupModelListControllerTestDB(t)
	require.NoError(t, model.LOG_DB.AutoMigrate(&model.Log{}))
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	require.NoError(t, model.LOG_DB.Create(&model.Log{
		UserId:           1,
		Username:         "alpha",
		CreatedAt:        todayStart.Unix() + 3600,
		Type:             model.LogTypeConsume,
		PromptTokens:     10,
		CompletionTokens: 20,
		Quota:            50,
	}).Error)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/rankings/users?period=today&timezone=UTC&limit=20", nil)

	GetUserRankings(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			Period        string `json:"period"`
			StartDate     string `json:"start_date"`
			EndDate       string `json:"end_date"`
			TotalTokens   int64  `json:"total_tokens"`
			TotalRequests int64  `json:"total_requests"`
			TotalQuota    int64  `json:"total_quota"`
			Ranking       []struct {
				Rank        int    `json:"rank"`
				UserId      int    `json:"user_id"`
				DisplayName string `json:"display_name"`
				Tokens      int64  `json:"tokens"`
				Requests    int64  `json:"requests"`
				Quota       int64  `json:"quota"`
			} `json:"ranking"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
	require.True(t, payload.Success)
	require.Equal(t, "today", payload.Data.Period)
	require.Equal(t, todayStart.Format("2006-01-02"), payload.Data.StartDate)
	require.Equal(t, todayStart.Format("2006-01-02"), payload.Data.EndDate)
	require.EqualValues(t, 30, payload.Data.TotalTokens)
	require.EqualValues(t, 1, payload.Data.TotalRequests)
	require.EqualValues(t, 50, payload.Data.TotalQuota)
	require.Len(t, payload.Data.Ranking, 1)
	require.Equal(t, "a***", payload.Data.Ranking[0].DisplayName)
}
