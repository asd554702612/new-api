package model

import (
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserModelSelection struct {
	UserId    int    `json:"user_id" gorm:"primaryKey;autoIncrement:false;index"`
	ModelName string `json:"model_name" gorm:"type:varchar(255);primaryKey;autoIncrement:false;index"`
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

type userModelSelectionCacheEntry struct {
	modelNames []string
	expiresAt  time.Time
}

var userModelSelectionCache = struct {
	sync.RWMutex
	items map[int]userModelSelectionCacheEntry
}{
	items: make(map[int]userModelSelectionCacheEntry),
}

func NormalizeModelSelectionNames(modelNames []string) []string {
	seen := make(map[string]struct{}, len(modelNames))
	normalized := make([]string, 0, len(modelNames))
	for _, modelName := range modelNames {
		modelName = strings.TrimSpace(modelName)
		if modelName == "" {
			continue
		}
		if _, ok := seen[modelName]; ok {
			continue
		}
		seen[modelName] = struct{}{}
		normalized = append(normalized, modelName)
	}
	return normalized
}

func GetUserModelSelections(userId int) ([]string, error) {
	if userId <= 0 {
		return []string{}, nil
	}
	if modelNames, ok := getUserModelSelectionsFromProcessCache(userId); ok {
		return modelNames, nil
	}
	var modelNames []string
	err := DB.Model(&UserModelSelection{}).
		Where("user_id = ?", userId).
		Order("model_name ASC").
		Pluck("model_name", &modelNames).Error
	if err != nil {
		return nil, err
	}
	setUserModelSelectionsProcessCache(userId, modelNames)
	return modelNames, nil
}

func GetUserModelSelectionMap(userId int) (map[string]bool, error) {
	modelNames, err := GetUserModelSelections(userId)
	if err != nil {
		return nil, err
	}
	selections := make(map[string]bool, len(modelNames))
	for _, modelName := range modelNames {
		selections[modelName] = true
	}
	return selections, nil
}

func ReplaceUserModelSelections(userId int, modelNames []string) error {
	if userId <= 0 {
		return nil
	}
	modelNames = NormalizeModelSelectionNames(modelNames)
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userId).Delete(&UserModelSelection{}).Error; err != nil {
			return err
		}
		if len(modelNames) == 0 {
			return nil
		}
		now := time.Now().Unix()
		selections := make([]UserModelSelection, 0, len(modelNames))
		for _, modelName := range modelNames {
			selections = append(selections, UserModelSelection{
				UserId:    userId,
				ModelName: modelName,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&selections).Error
	})
	if err != nil {
		return err
	}
	setUserModelSelectionsProcessCache(userId, modelNames)
	return nil
}

func getUserModelSelectionCacheTTL() time.Duration {
	seconds := common.SyncFrequency
	if seconds <= 0 {
		seconds = 60
	}
	return time.Duration(seconds) * time.Second
}

func getUserModelSelectionsFromProcessCache(userId int) ([]string, bool) {
	if !common.MemoryCacheEnabled {
		return nil, false
	}
	now := time.Now()
	userModelSelectionCache.RLock()
	entry, ok := userModelSelectionCache.items[userId]
	userModelSelectionCache.RUnlock()
	if !ok {
		return nil, false
	}
	if !entry.expiresAt.After(now) {
		userModelSelectionCache.Lock()
		if current, exists := userModelSelectionCache.items[userId]; exists && !current.expiresAt.After(now) {
			delete(userModelSelectionCache.items, userId)
		}
		userModelSelectionCache.Unlock()
		return nil, false
	}
	return append([]string(nil), entry.modelNames...), true
}

func setUserModelSelectionsProcessCache(userId int, modelNames []string) {
	if !common.MemoryCacheEnabled || userId <= 0 {
		return
	}
	userModelSelectionCache.Lock()
	userModelSelectionCache.items[userId] = userModelSelectionCacheEntry{
		modelNames: append([]string(nil), modelNames...),
		expiresAt:  time.Now().Add(getUserModelSelectionCacheTTL()),
	}
	userModelSelectionCache.Unlock()
}
