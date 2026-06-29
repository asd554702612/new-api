package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseRedisOptionAppliesTimeoutAndPoolEnv(t *testing.T) {
	t.Setenv("REDIS_CONN_STRING", "redis://:password@10.66.66.1:6379/1")
	t.Setenv("REDIS_DIAL_TIMEOUT_MS", "800")
	t.Setenv("REDIS_READ_TIMEOUT_MS", "700")
	t.Setenv("REDIS_WRITE_TIMEOUT_MS", "600")
	t.Setenv("REDIS_POOL_TIMEOUT_MS", "500")
	t.Setenv("REDIS_MAX_RETRIES", "1")
	t.Setenv("REDIS_MIN_IDLE_CONNS", "10")
	t.Setenv("REDIS_POOL_SIZE", "30")

	opt := ParseRedisOption()

	require.Equal(t, "10.66.66.1:6379", opt.Addr)
	require.Equal(t, 1, opt.DB)
	require.Equal(t, 800*time.Millisecond, opt.DialTimeout)
	require.Equal(t, 700*time.Millisecond, opt.ReadTimeout)
	require.Equal(t, 600*time.Millisecond, opt.WriteTimeout)
	require.Equal(t, 500*time.Millisecond, opt.PoolTimeout)
	require.Equal(t, 1, opt.MaxRetries)
	require.Equal(t, 10, opt.MinIdleConns)
	require.Equal(t, 30, opt.PoolSize)
}
