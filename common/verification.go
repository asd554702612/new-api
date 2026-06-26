package common

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type verificationValue struct {
	code string
	time time.Time
}

const (
	EmailVerificationPurpose = "v"
	PasswordResetPurpose     = "r"
)

var verificationMutex sync.Mutex
var verificationMap map[string]verificationValue
var verificationMapMaxSize = 10000
var VerificationValidMinutes = 10

func GenerateVerificationCode(length int) string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	if length == 0 {
		return code
	}
	return code[:length]
}

func RegisterVerificationCodeWithKey(key string, code string, purpose string) {
	cacheKey := verificationCacheKey(key, purpose)
	if RedisEnabled && RDB != nil {
		_ = RedisSet(cacheKey, code, time.Duration(VerificationValidMinutes)*time.Minute)
	}
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	verificationMap[cacheKey] = verificationValue{
		code: code,
		time: time.Now(),
	}
	if len(verificationMap) > verificationMapMaxSize {
		removeExpiredPairs()
	}
}

func VerifyCodeWithKey(key string, code string, purpose string) bool {
	cacheKey := verificationCacheKey(key, purpose)
	if RedisEnabled && RDB != nil {
		value, err := RedisGet(cacheKey)
		if err == nil {
			if value != code {
				return false
			}
			_ = RedisDel(cacheKey)
			deleteVerificationCodeFromMemory(cacheKey)
			return true
		}
	}
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	value, okay := verificationMap[cacheKey]
	now := time.Now()
	if !okay || int(now.Sub(value.time).Seconds()) >= VerificationValidMinutes*60 {
		return false
	}
	return code == value.code
}

func DeleteKey(key string, purpose string) {
	cacheKey := verificationCacheKey(key, purpose)
	if RedisEnabled && RDB != nil {
		_ = RedisDel(cacheKey)
	}
	deleteVerificationCodeFromMemory(cacheKey)
}

func deleteVerificationCodeFromMemory(cacheKey string) {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	delete(verificationMap, cacheKey)
}

func verificationCacheKey(key string, purpose string) string {
	return "verification:" + purpose + ":" + key
}

// no lock inside, so the caller must lock the verificationMap before calling!
func removeExpiredPairs() {
	now := time.Now()
	for key := range verificationMap {
		if int(now.Sub(verificationMap[key].time).Seconds()) >= VerificationValidMinutes*60 {
			delete(verificationMap, key)
		}
	}
}

func init() {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	verificationMap = make(map[string]verificationValue)
}
