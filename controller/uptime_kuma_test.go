package controller

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/console_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetUptimeKumaStatusUsesShortProcessCache(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var upstreamHits int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstreamHits, 1)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/status-page/demo":
			_, _ = w.Write([]byte(`{"publicGroupList":[{"id":1,"name":"core","monitorList":[{"id":11,"name":"api"}]}]}`))
		case "/api/status-page/heartbeat/demo":
			_, _ = w.Write([]byte(`{"heartbeatList":{"11":[{"status":1}]},"uptimeList":{"11_24":99.9}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	original := *console_setting.GetConsoleSetting()
	t.Cleanup(func() {
		*console_setting.GetConsoleSetting() = original
	})

	groups, err := common.Marshal([]map[string]any{{
		"url":          upstream.URL,
		"slug":         "demo",
		"categoryName": "Demo",
	}})
	require.NoError(t, err)
	console_setting.GetConsoleSetting().UptimeKumaGroups = string(groups)

	router := gin.New()
	router.GET("/api/uptime/status", GetUptimeKumaStatus)

	for i := 0; i < 2; i++ {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/uptime/status", nil))
		require.Equal(t, http.StatusOK, recorder.Code)

		var response struct {
			Success bool                `json:"success"`
			Data    []UptimeGroupResult `json:"data"`
		}
		require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
		require.True(t, response.Success)
		require.Len(t, response.Data, 1)
		require.Equal(t, "Demo", response.Data[0].CategoryName)
		require.Len(t, response.Data[0].Monitors, 1)
		require.Equal(t, "api", response.Data[0].Monitors[0].Name)
	}

	require.Equal(t, int32(2), atomic.LoadInt32(&upstreamHits), "status and heartbeat should each be fetched once")
}
