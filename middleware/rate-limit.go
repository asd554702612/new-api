package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

var timeFormat = "2006-01-02T15:04:05.000Z"

var inMemoryRateLimiter common.InMemoryRateLimiter

var defNext = func(c *gin.Context) {
	c.Next()
}

func redisRateLimiter(c *gin.Context, maxRequestNum int, duration int64, mark string) {
	ctx := context.Background()
	rdb := common.RDB
	key := "rateLimit:" + mark + c.ClientIP()
	listLength, err := rdb.LLen(ctx, key).Result()
	if err != nil {
		fmt.Println(err.Error())
		memoryRateLimiterWithInit(c, maxRequestNum, duration, mark)
		return
	}
	if listLength < int64(maxRequestNum) {
		if err := rdb.LPush(ctx, key, time.Now().Format(timeFormat)).Err(); err != nil {
			fmt.Println(err.Error())
			memoryRateLimiterWithInit(c, maxRequestNum, duration, mark)
			return
		}
		_ = rdb.Expire(ctx, key, common.RateLimitKeyExpirationDuration).Err()
	} else {
		oldTimeStr, err := rdb.LIndex(ctx, key, -1).Result()
		if err != nil {
			fmt.Println(err.Error())
			memoryRateLimiterWithInit(c, maxRequestNum, duration, mark)
			return
		}
		oldTime, err := time.Parse(timeFormat, oldTimeStr)
		if err != nil {
			fmt.Println(err)
			memoryRateLimiterWithInit(c, maxRequestNum, duration, mark)
			return
		}
		nowTimeStr := time.Now().Format(timeFormat)
		nowTime, err := time.Parse(timeFormat, nowTimeStr)
		if err != nil {
			fmt.Println(err)
			memoryRateLimiterWithInit(c, maxRequestNum, duration, mark)
			return
		}
		// time.Since will return negative number!
		// See: https://stackoverflow.com/questions/50970900/why-is-time-since-returning-negative-durations-on-windows
		if int64(nowTime.Sub(oldTime).Seconds()) < duration {
			_ = rdb.Expire(ctx, key, common.RateLimitKeyExpirationDuration).Err()
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		} else {
			if err := rdb.LPush(ctx, key, time.Now().Format(timeFormat)).Err(); err != nil {
				fmt.Println(err.Error())
				memoryRateLimiterWithInit(c, maxRequestNum, duration, mark)
				return
			}
			_ = rdb.LTrim(ctx, key, 0, int64(maxRequestNum-1)).Err()
			_ = rdb.Expire(ctx, key, common.RateLimitKeyExpirationDuration).Err()
		}
	}
}

func memoryRateLimiter(c *gin.Context, maxRequestNum int, duration int64, mark string) {
	key := mark + c.ClientIP()
	if !inMemoryRateLimiter.Request(key, maxRequestNum, duration) {
		c.Status(http.StatusTooManyRequests)
		c.Abort()
		return
	}
}

func memoryRateLimiterWithInit(c *gin.Context, maxRequestNum int, duration int64, mark string) {
	inMemoryRateLimiter.Init(common.RateLimitKeyExpirationDuration)
	memoryRateLimiter(c, maxRequestNum, duration, mark)
}

func rateLimitFactory(maxRequestNum int, duration int64, mark string) func(c *gin.Context) {
	if common.RedisEnabled {
		return func(c *gin.Context) {
			redisRateLimiter(c, maxRequestNum, duration, mark)
		}
	} else {
		// It's safe to call multi times.
		inMemoryRateLimiter.Init(common.RateLimitKeyExpirationDuration)
		return func(c *gin.Context) {
			memoryRateLimiter(c, maxRequestNum, duration, mark)
		}
	}
}

func GlobalWebRateLimit() func(c *gin.Context) {
	if common.GlobalWebRateLimitEnable {
		return rateLimitFactory(common.GlobalWebRateLimitNum, common.GlobalWebRateLimitDuration, "GW")
	}
	return defNext
}

func GlobalAPIRateLimit() func(c *gin.Context) {
	if common.GlobalApiRateLimitEnable {
		return rateLimitFactory(common.GlobalApiRateLimitNum, common.GlobalApiRateLimitDuration, "GA")
	}
	return defNext
}

func CriticalRateLimit() func(c *gin.Context) {
	if common.CriticalRateLimitEnable {
		return rateLimitFactory(common.CriticalRateLimitNum, common.CriticalRateLimitDuration, "CT")
	}
	return defNext
}

func DownloadRateLimit() func(c *gin.Context) {
	return rateLimitFactory(common.DownloadRateLimitNum, common.DownloadRateLimitDuration, "DW")
}

func UploadRateLimit() func(c *gin.Context) {
	return rateLimitFactory(common.UploadRateLimitNum, common.UploadRateLimitDuration, "UP")
}

// userRateLimitFactory creates a rate limiter keyed by authenticated user ID
// instead of client IP, making it resistant to proxy rotation attacks.
// Must be used AFTER authentication middleware (UserAuth).
func userRateLimitFactory(maxRequestNum int, duration int64, mark string) func(c *gin.Context) {
	if common.RedisEnabled {
		return func(c *gin.Context) {
			userId := c.GetInt("id")
			if userId == 0 {
				c.Status(http.StatusUnauthorized)
				c.Abort()
				return
			}
			key := fmt.Sprintf("rateLimit:%s:user:%d", mark, userId)
			userRedisRateLimiter(c, maxRequestNum, duration, key)
		}
	}
	// It's safe to call multi times.
	inMemoryRateLimiter.Init(common.RateLimitKeyExpirationDuration)
	return func(c *gin.Context) {
		userId := c.GetInt("id")
		if userId == 0 {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
		key := fmt.Sprintf("%s:user:%d", mark, userId)
		if !inMemoryRateLimiter.Request(key, maxRequestNum, duration) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}
	}
}

// userRedisRateLimiter is like redisRateLimiter but accepts a pre-built key
// (to support user-ID-based keys).
func userRedisRateLimiter(c *gin.Context, maxRequestNum int, duration int64, key string) {
	ctx := context.Background()
	rdb := common.RDB
	listLength, err := rdb.LLen(ctx, key).Result()
	if err != nil {
		fmt.Println(err.Error())
		userMemoryRateLimiter(c, maxRequestNum, duration, key)
		return
	}
	if listLength < int64(maxRequestNum) {
		if err := rdb.LPush(ctx, key, time.Now().Format(timeFormat)).Err(); err != nil {
			fmt.Println(err.Error())
			userMemoryRateLimiter(c, maxRequestNum, duration, key)
			return
		}
		_ = rdb.Expire(ctx, key, common.RateLimitKeyExpirationDuration).Err()
	} else {
		oldTimeStr, err := rdb.LIndex(ctx, key, -1).Result()
		if err != nil {
			fmt.Println(err.Error())
			userMemoryRateLimiter(c, maxRequestNum, duration, key)
			return
		}
		oldTime, err := time.Parse(timeFormat, oldTimeStr)
		if err != nil {
			fmt.Println(err)
			userMemoryRateLimiter(c, maxRequestNum, duration, key)
			return
		}
		nowTimeStr := time.Now().Format(timeFormat)
		nowTime, err := time.Parse(timeFormat, nowTimeStr)
		if err != nil {
			fmt.Println(err)
			userMemoryRateLimiter(c, maxRequestNum, duration, key)
			return
		}
		if int64(nowTime.Sub(oldTime).Seconds()) < duration {
			_ = rdb.Expire(ctx, key, common.RateLimitKeyExpirationDuration).Err()
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		} else {
			if err := rdb.LPush(ctx, key, time.Now().Format(timeFormat)).Err(); err != nil {
				fmt.Println(err.Error())
				userMemoryRateLimiter(c, maxRequestNum, duration, key)
				return
			}
			_ = rdb.LTrim(ctx, key, 0, int64(maxRequestNum-1)).Err()
			_ = rdb.Expire(ctx, key, common.RateLimitKeyExpirationDuration).Err()
		}
	}
}

func userMemoryRateLimiter(c *gin.Context, maxRequestNum int, duration int64, key string) {
	inMemoryRateLimiter.Init(common.RateLimitKeyExpirationDuration)
	if !inMemoryRateLimiter.Request(key, maxRequestNum, duration) {
		c.Status(http.StatusTooManyRequests)
		c.Abort()
	}
}

// SearchRateLimit returns a per-user rate limiter for search endpoints.
// Configurable via SEARCH_RATE_LIMIT_ENABLE / SEARCH_RATE_LIMIT / SEARCH_RATE_LIMIT_DURATION.
func SearchRateLimit() func(c *gin.Context) {
	if !common.SearchRateLimitEnable {
		return defNext
	}
	return userRateLimitFactory(common.SearchRateLimitNum, common.SearchRateLimitDuration, "SR")
}
