package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseUserLeaderboardLimitDefaultsBoundsAndRejectsBadInput(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int
		wantErr bool
	}{
		{name: "empty", raw: "", want: 20},
		{name: "zero", raw: "0", want: 20},
		{name: "negative", raw: "-5", want: 20},
		{name: "valid", raw: "35", want: 35},
		{name: "caps maximum", raw: "200", want: 100},
		{name: "bad", raw: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUserLeaderboardLimit(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserLeaderboardRangeSupportsTodayYesterdayAndTimezoneFallback(t *testing.T) {
	now := time.Date(2026, 6, 6, 15, 30, 0, 0, time.UTC)

	start, end, period, ok := userLeaderboardRange("today", "Asia/Shanghai", now)
	require.True(t, ok)
	require.Equal(t, "today", period)
	require.Equal(t, time.Date(2026, 6, 6, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)).Unix(), start)
	require.Equal(t, time.Date(2026, 6, 7, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)).Unix(), end)

	start, end, period, ok = userLeaderboardRange("yesterday", "UTC", now)
	require.True(t, ok)
	require.Equal(t, "yesterday", period)
	require.Equal(t, time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC).Unix(), start)
	require.Equal(t, time.Date(2026, 6, 6, 0, 0, 0, 0, time.UTC).Unix(), end)

	_, _, _, ok = userLeaderboardRange("week", "UTC", now)
	require.False(t, ok)
}

func TestMaskUserLeaderboardDisplayName(t *testing.T) {
	require.Equal(t, "a***", maskUserLeaderboardDisplayName("alpha", 1))
	require.Equal(t, "用***", maskUserLeaderboardDisplayName("用户甲", 2))
	require.Equal(t, "User #3", maskUserLeaderboardDisplayName("", 3))
}
