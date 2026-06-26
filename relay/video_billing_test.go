package relay

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/video_billing_setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestApplyConfiguredVideoBillingOverridesTaskPrice(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, video_billing_setting.UpdateRulesByJSONString(`{}`))
	})
	require.NoError(t, video_billing_setting.UpdateRulesByJSONString(`{
		"happyhorse-1.0-t2v": {"mode":"per_second","base_price":0.03}
	}`))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/videos", strings.NewReader(`{"duration":5}`))
	c.Set("task_request", relaycommon.TaskSubmitReq{Duration: 5})

	info := &relaycommon.RelayInfo{
		OriginModelName: "happyhorse-1.0-t2v",
		PriceData: types.PriceData{
			Quota: 999,
			GroupRatioInfo: types.GroupRatioInfo{
				GroupRatio: 2,
			},
		},
	}

	applied, err := applyConfiguredVideoBilling(c, info)

	require.NoError(t, err)
	require.True(t, applied)
	require.Equal(t, int(0.15*common.QuotaPerUnit*2), info.PriceData.Quota)
	require.True(t, info.PriceData.UsePrice)
	require.InDelta(t, 0.15, info.PriceData.ModelPrice, 0.000001)
	require.InDelta(t, 5.0, info.PriceData.OtherRatios["video_seconds"], 0.000001)
}

func TestApplyConfiguredVideoBillingPriceDataDoesNotRequireModelPrice(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, video_billing_setting.UpdateRulesByJSONString(`{}`))
	})
	require.NoError(t, video_billing_setting.UpdateRulesByJSONString(`{
		"happyhorse-1.0-t2v": {"mode":"per_second","base_price":0.03}
	}`))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/videos", strings.NewReader(`{"duration":5}`))
	c.Set("task_request", relaycommon.TaskSubmitReq{Duration: 5})

	info := &relaycommon.RelayInfo{
		OriginModelName: "happyhorse-1.0-t2v",
		UserGroup:       "default",
		UsingGroup:      "default",
	}

	applied, err := applyConfiguredVideoBillingPriceData(c, info)

	require.NoError(t, err)
	require.True(t, applied)
	require.Equal(t, int(0.15*common.QuotaPerUnit), info.PriceData.Quota)
	require.True(t, info.PriceData.UsePrice)
	require.InDelta(t, 5.0, info.PriceData.OtherRatios["video_seconds"], 0.000001)
}

func TestApplyConfiguredVideoBillingUsesChannelDefaultSeconds(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, video_billing_setting.UpdateRulesByJSONString(`{}`))
	})
	require.NoError(t, video_billing_setting.UpdateRulesByJSONString(`{
		"wan2.2-t2v-plus": {"mode":"per_second","base_price":0.03},
		"sora-2": {"mode":"per_second","base_price":0.03},
		"veo-3.0-generate-001": {"mode":"per_second","base_price":0.03}
	}`))

	cases := []struct {
		name        string
		model       string
		channelType int
		wantSeconds float64
	}{
		{name: "ali", model: "wan2.2-t2v-plus", channelType: constant.ChannelTypeAli, wantSeconds: 5},
		{name: "sora", model: "sora-2", channelType: constant.ChannelTypeSora, wantSeconds: 4},
		{name: "veo", model: "veo-3.0-generate-001", channelType: constant.ChannelTypeGemini, wantSeconds: 8},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/v1/videos", strings.NewReader(`{}`))
			c.Set("task_request", relaycommon.TaskSubmitReq{})

			info := &relaycommon.RelayInfo{
				OriginModelName: tc.model,
				ChannelMeta: &relaycommon.ChannelMeta{
					ChannelType: tc.channelType,
				},
				PriceData: types.PriceData{
					GroupRatioInfo: types.GroupRatioInfo{
						GroupRatio: 1,
					},
				},
			}

			applied, err := applyConfiguredVideoBilling(c, info)

			require.NoError(t, err)
			require.True(t, applied)
			require.InDelta(t, tc.wantSeconds, info.PriceData.OtherRatios["video_seconds"], 0.000001)
			require.Equal(t, int(0.03*tc.wantSeconds*common.QuotaPerUnit), info.PriceData.Quota)
		})
	}
}
