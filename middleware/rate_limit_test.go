package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func failingRedisClientForRateLimitTest() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  5 * time.Millisecond,
		ReadTimeout:  5 * time.Millisecond,
		WriteTimeout: 5 * time.Millisecond,
		PoolTimeout:  5 * time.Millisecond,
		MaxRetries:   0,
	})
}

func performRateLimitRequest(router http.Handler, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.RemoteAddr = "203.0.113.10:12345"
	router.ServeHTTP(recorder, request)
	return recorder
}

func resetInMemoryRateLimiterForTest() {
	inMemoryRateLimiter.Clear()
}

func TestGlobalRateLimiterFallsBackToMemoryWhenRedisFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalRedisEnabled := common.RedisEnabled
	originalRDB := common.RDB
	defer func() {
		common.RedisEnabled = originalRedisEnabled
		common.RDB = originalRDB
		resetInMemoryRateLimiterForTest()
	}()
	common.RedisEnabled = true
	common.RDB = failingRedisClientForRateLimitTest()
	resetInMemoryRateLimiterForTest()

	router := gin.New()
	router.Use(rateLimitFactory(10, 60, "TEST"))
	router.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	recorder := performRateLimitRequest(router, "/ok")

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "ok", recorder.Body.String())
}

func TestModelRateLimiterFallsBackToMemoryWhenRedisFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalRedisEnabled := common.RedisEnabled
	originalRDB := common.RDB
	originalEnabled := setting.ModelRequestRateLimitEnabled
	originalDuration := setting.ModelRequestRateLimitDurationMinutes
	originalCount := setting.ModelRequestRateLimitCount
	originalSuccessCount := setting.ModelRequestRateLimitSuccessCount
	defer func() {
		common.RedisEnabled = originalRedisEnabled
		common.RDB = originalRDB
		setting.ModelRequestRateLimitEnabled = originalEnabled
		setting.ModelRequestRateLimitDurationMinutes = originalDuration
		setting.ModelRequestRateLimitCount = originalCount
		setting.ModelRequestRateLimitSuccessCount = originalSuccessCount
		resetInMemoryRateLimiterForTest()
	}()
	common.RedisEnabled = true
	common.RDB = failingRedisClientForRateLimitTest()
	setting.ModelRequestRateLimitEnabled = true
	setting.ModelRequestRateLimitDurationMinutes = 1
	setting.ModelRequestRateLimitCount = 10
	setting.ModelRequestRateLimitSuccessCount = 10
	resetInMemoryRateLimiterForTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("id", 123)
		c.Next()
	})
	router.Use(ModelRequestRateLimit())
	router.GET("/relay", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	recorder := performRateLimitRequest(router, "/relay")

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "ok", recorder.Body.String())
}
