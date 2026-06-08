package model

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type RankingQuotaTotal struct {
	ModelName   string `json:"model_name"`
	TotalTokens int64  `json:"total_tokens"`
}

type RankingQuotaBucket struct {
	ModelName string `json:"model_name"`
	Bucket    int64  `json:"bucket"`
	Tokens    int64  `json:"tokens"`
}

type UserLeaderboardRow struct {
	Rank     int    `json:"rank" gorm:"-"`
	UserId   int    `json:"user_id" gorm:"column:user_id"`
	Username string `json:"username" gorm:"column:username"`
	Tokens   int64  `json:"tokens" gorm:"column:tokens"`
	Requests int64  `json:"requests" gorm:"column:requests"`
	Quota    int64  `json:"quota" gorm:"column:quota"`
}

type UserLeaderboardSummary struct {
	TotalTokens   int64 `json:"total_tokens" gorm:"column:total_tokens"`
	TotalRequests int64 `json:"total_requests" gorm:"column:total_requests"`
	TotalQuota    int64 `json:"total_quota" gorm:"column:total_quota"`
}

func GetRankingQuotaTotals(startTime int64, endTime int64) ([]RankingQuotaTotal, error) {
	var rows []RankingQuotaTotal
	query := DB.Table("quota_data").
		Select("model_name, sum(token_used) as total_tokens").
		Where("model_name <> ''").
		Group("model_name").
		Having("sum(token_used) > 0").
		Order("total_tokens DESC")
	query = applyRankingQuotaTimeRange(query, startTime, endTime)
	err := query.Find(&rows).Error
	return rows, err
}

func GetUserLeaderboardRows(startTime int64, endTime int64, limit int, ignoredUserIds []int) ([]UserLeaderboardRow, UserLeaderboardSummary, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	base := applyUserLeaderboardFilters(LOG_DB.Table("logs"), startTime, endTime, ignoredUserIds)

	var summary UserLeaderboardSummary
	if err := base.
		Select("COALESCE(SUM(prompt_tokens + completion_tokens), 0) as total_tokens, COUNT(*) as total_requests, COALESCE(SUM(quota), 0) as total_quota").
		Scan(&summary).Error; err != nil {
		return nil, UserLeaderboardSummary{}, err
	}

	var rows []UserLeaderboardRow
	if err := base.
		Select("user_id, COALESCE(MAX(username), '') as username, COALESCE(SUM(prompt_tokens + completion_tokens), 0) as tokens, COUNT(*) as requests, COALESCE(SUM(quota), 0) as quota").
		Group("user_id").
		Order("tokens DESC, requests DESC, user_id ASC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, UserLeaderboardSummary{}, err
	}

	for i := range rows {
		rows[i].Rank = i + 1
	}
	return rows, summary, nil
}

func GetRankingQuotaBuckets(startTime int64, endTime int64, bucketSize int64) ([]RankingQuotaBucket, error) {
	if bucketSize <= 0 {
		bucketSize = 3600
	}
	bucketExpr := rankingBucketExpr(bucketSize)
	var rows []RankingQuotaBucket
	query := DB.Table("quota_data").
		Select(fmt.Sprintf("model_name, %s as bucket, sum(token_used) as tokens", bucketExpr)).
		Where("model_name <> ''").
		Group(fmt.Sprintf("model_name, %s", bucketExpr)).
		Having("sum(token_used) > 0").
		Order("bucket ASC")
	query = applyRankingQuotaTimeRange(query, startTime, endTime)
	err := query.Find(&rows).Error
	return rows, err
}

func rankingBucketExpr(bucketSize int64) string {
	if common.UsingMySQL {
		return fmt.Sprintf("FLOOR(created_at / %d) * %d", bucketSize, bucketSize)
	}
	return fmt.Sprintf("(created_at / %d) * %d", bucketSize, bucketSize)
}

func applyRankingQuotaTimeRange(query *gorm.DB, startTime int64, endTime int64) *gorm.DB {
	if startTime > 0 {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_at <= ?", endTime)
	}
	return query
}

func applyUserLeaderboardFilters(query *gorm.DB, startTime int64, endTime int64, ignoredUserIds []int) *gorm.DB {
	query = query.Where("type = ?", LogTypeConsume)
	if startTime > 0 {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_at < ?", endTime)
	}
	if len(ignoredUserIds) > 0 {
		query = query.Where("user_id NOT IN ?", ignoredUserIds)
	}
	return query
}
