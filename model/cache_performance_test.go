package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupModelPerformanceCacheTestDB(t *testing.T, tables ...any) *gorm.DB {
	t.Helper()

	previousMemoryCacheEnabled := common.MemoryCacheEnabled
	previousUsingSQLite := common.UsingSQLite
	previousUsingMySQL := common.UsingMySQL
	previousUsingPostgreSQL := common.UsingPostgreSQL
	previousDB := DB
	previousLogDB := LOG_DB
	channelSyncLock.RLock()
	previousGroup2model2channels := group2model2channels
	previousChannelsIDM := channelsIDM
	previousEnabledModelsByGroup := enabledModelsByGroup
	previousEnabledModels := enabledModels
	channelSyncLock.RUnlock()
	t.Cleanup(func() {
		common.MemoryCacheEnabled = previousMemoryCacheEnabled
		common.UsingSQLite = previousUsingSQLite
		common.UsingMySQL = previousUsingMySQL
		common.UsingPostgreSQL = previousUsingPostgreSQL
		DB = previousDB
		LOG_DB = previousLogDB
		channelSyncLock.Lock()
		group2model2channels = previousGroup2model2channels
		channelsIDM = previousChannelsIDM
		enabledModelsByGroup = previousEnabledModelsByGroup
		enabledModels = previousEnabledModels
		channelSyncLock.Unlock()
		userModelSelectionCache.Lock()
		userModelSelectionCache.items = make(map[int]userModelSelectionCacheEntry)
		userModelSelectionCache.Unlock()
		clearQuotaDataQueryCache()
	})

	common.MemoryCacheEnabled = true
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	initCol()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	DB = db
	LOG_DB = db
	require.NoError(t, db.AutoMigrate(tables...))
	return db
}

func TestEnabledModelListsUseChannelCacheSnapshot(t *testing.T) {
	db := setupModelPerformanceCacheTestDB(t, &Ability{}, &Channel{})
	require.NoError(t, db.Create(&[]Ability{
		{Group: "default", Model: "cached-a", ChannelId: 1, Enabled: true},
		{Group: "default", Model: "cached-b", ChannelId: 1, Enabled: true},
		{Group: "vip", Model: "cached-b", ChannelId: 1, Enabled: true},
		{Group: "vip", Model: "disabled-model", ChannelId: 1, Enabled: false},
	}).Error)

	InitChannelCache()
	require.NoError(t, db.Exec("DELETE FROM abilities").Error)

	require.ElementsMatch(t, []string{"cached-a", "cached-b"}, GetGroupEnabledModels("default"))
	require.ElementsMatch(t, []string{"cached-b"}, GetGroupEnabledModels("vip"))
	require.ElementsMatch(t, []string{"cached-a", "cached-b"}, GetEnabledModels())
}

func TestUserModelSelectionsUseShortProcessCache(t *testing.T) {
	db := setupModelPerformanceCacheTestDB(t, &UserModelSelection{})
	require.NoError(t, ReplaceUserModelSelections(7, []string{"model-a", "model-b"}))

	first, err := GetUserModelSelections(7)
	require.NoError(t, err)
	require.Equal(t, []string{"model-a", "model-b"}, first)

	require.NoError(t, db.Exec("DELETE FROM user_model_selections").Error)
	second, err := GetUserModelSelections(7)
	require.NoError(t, err)
	require.Equal(t, []string{"model-a", "model-b"}, second)
}

func TestQuotaDataQueriesUseShortProcessCache(t *testing.T) {
	db := setupModelPerformanceCacheTestDB(t, &QuotaData{})
	require.NoError(t, db.Create(&QuotaData{
		UserID:    42,
		Username:  "alice",
		ModelName: "gpt-test",
		CreatedAt: 3600,
		Count:     2,
		Quota:     100,
		TokenUsed: 20,
	}).Error)

	first, err := GetAllQuotaDates(0, 7200, "")
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.Equal(t, "gpt-test", first[0].ModelName)

	require.NoError(t, db.Exec("DELETE FROM quota_data").Error)
	second, err := GetAllQuotaDates(0, 7200, "")
	require.NoError(t, err)
	require.Len(t, second, 1)
	require.Equal(t, "gpt-test", second[0].ModelName)
}
