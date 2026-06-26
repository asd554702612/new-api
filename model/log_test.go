package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func insertAdminUsageLog(t *testing.T, userId int, username string, createdAt int64, quota int) {
	t.Helper()
	require.NoError(t, LOG_DB.Create(&Log{
		UserId:           userId,
		Username:         username,
		CreatedAt:        createdAt,
		Type:             LogTypeConsume,
		ModelName:        "gpt-test",
		TokenName:        "token-test",
		PromptTokens:     10,
		CompletionTokens: 5,
		Quota:            quota,
	}).Error)
}

func TestGetAllLogsFiltersByUserIdWhenProvided(t *testing.T) {
	truncateTables(t)

	insertAdminUsageLog(t, 101, "alpha", 100, 100)
	insertAdminUsageLog(t, 202, "beta", 200, 200)

	allLogs, total, err := GetAllLogs(LogTypeConsume, 0, 0, "", "", "", 0, 10, 0, "", "", "", 0)
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, allLogs, 2)

	userLogs, total, err := GetAllLogs(LogTypeConsume, 0, 0, "", "", "", 0, 10, 0, "", "", "", 101)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, userLogs, 1)
	require.Equal(t, 101, userLogs[0].UserId)
	require.Equal(t, "alpha", userLogs[0].Username)
}

func TestSumUsedQuotaFiltersByUserIdWhenProvided(t *testing.T) {
	truncateTables(t)

	insertAdminUsageLog(t, 101, "alpha", 100, 100)
	insertAdminUsageLog(t, 202, "beta", 200, 200)

	allStat, err := SumUsedQuota(LogTypeConsume, 0, 0, "", "", "", 0, "", 0)
	require.NoError(t, err)
	require.Equal(t, 300, allStat.Quota)

	userStat, err := SumUsedQuota(LogTypeConsume, 0, 0, "", "", "", 0, "", 101)
	require.NoError(t, err)
	require.Equal(t, 100, userStat.Quota)
}
