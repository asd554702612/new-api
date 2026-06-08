package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func insertUsageRankingLog(t *testing.T, userId int, username string, createdAt int64, promptTokens int, completionTokens int, quota int) {
	t.Helper()
	require.NoError(t, LOG_DB.Create(&Log{
		UserId:           userId,
		Username:         username,
		CreatedAt:        createdAt,
		Type:             LogTypeConsume,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		Quota:            quota,
	}).Error)
}

func TestGetUserLeaderboardRowsAggregatesAndExcludesIgnoredUsers(t *testing.T) {
	truncateTables(t)

	start := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC).Unix()
	end := start + 24*3600

	insertUsageRankingLog(t, 1, "alpha", start+10, 100, 50, 300)
	insertUsageRankingLog(t, 1, "alpha", start+20, 30, 20, 100)
	insertUsageRankingLog(t, 2, "ignored", start+30, 1000, 1000, 5000)
	insertUsageRankingLog(t, 3, "gamma", start+40, 80, 20, 200)
	insertUsageRankingLog(t, 4, "outside", end+1, 900, 100, 1000)
	require.NoError(t, LOG_DB.Create(&Log{
		UserId:           5,
		Username:         "error",
		CreatedAt:        start + 50,
		Type:             LogTypeError,
		PromptTokens:     500,
		CompletionTokens: 500,
		Quota:            1000,
	}).Error)

	rows, summary, err := GetUserLeaderboardRows(start, end, 20, []int{2})
	require.NoError(t, err)
	require.Equal(t, UserLeaderboardSummary{
		TotalTokens:   200 + 100,
		TotalRequests: 3,
		TotalQuota:    600,
	}, summary)
	require.Equal(t, []UserLeaderboardRow{
		{Rank: 1, UserId: 1, Username: "alpha", Tokens: 200, Requests: 2, Quota: 400},
		{Rank: 2, UserId: 3, Username: "gamma", Tokens: 100, Requests: 1, Quota: 200},
	}, rows)
}

func TestGetUserLeaderboardRowsUsesStableSortAndTotalsBeforeLimit(t *testing.T) {
	truncateTables(t)

	start := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC).Unix()
	end := start + 24*3600

	insertUsageRankingLog(t, 5, "beta", start+10, 50, 50, 50)
	insertUsageRankingLog(t, 1, "alpha", start+20, 50, 50, 70)
	insertUsageRankingLog(t, 9, "gamma", start+30, 40, 60, 90)

	rows, summary, err := GetUserLeaderboardRows(start, end, 2, nil)
	require.NoError(t, err)
	require.Equal(t, UserLeaderboardSummary{
		TotalTokens:   300,
		TotalRequests: 3,
		TotalQuota:    210,
	}, summary)
	require.Equal(t, []UserLeaderboardRow{
		{Rank: 1, UserId: 1, Username: "alpha", Tokens: 100, Requests: 1, Quota: 70},
		{Rank: 2, UserId: 5, Username: "beta", Tokens: 100, Requests: 1, Quota: 50},
	}, rows)
}

func TestGetUserLeaderboardRowsEmptyResult(t *testing.T) {
	truncateTables(t)

	start := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC).Unix()
	end := start + 24*3600

	rows, summary, err := GetUserLeaderboardRows(start, end, 20, nil)
	require.NoError(t, err)
	require.Empty(t, rows)
	require.Equal(t, UserLeaderboardSummary{}, summary)
}
