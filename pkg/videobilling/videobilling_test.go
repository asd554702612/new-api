package videobilling

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/require"
)

func TestCalculatePerSecond(t *testing.T) {
	result, err := Calculate(Rule{
		Mode:      ModePerSecond,
		BasePrice: 0.03,
	}, Input{
		Request: relaycommon.TaskSubmitReq{Duration: 5},
	})

	require.NoError(t, err)
	require.InDelta(t, 0.15, result.AmountUSD, 0.000001)
	require.Equal(t, 5, result.Seconds)
}

func TestCalculateMatrix(t *testing.T) {
	result, err := Calculate(Rule{
		Mode:      ModeMatrix,
		BasePrice: 0.04,
		Multipliers: map[string]map[string]float64{
			"resolution": {
				"1080p": 1.5,
			},
			"duration": {
				"10": 2,
			},
		},
	}, Input{
		Request: relaycommon.TaskSubmitReq{
			Duration: 10,
			Metadata: map[string]interface{}{
				"resolution": "1080p",
			},
		},
	})

	require.NoError(t, err)
	require.InDelta(t, 1.2, result.AmountUSD, 0.000001)
	require.Equal(t, 10, result.Seconds)
	require.InDelta(t, 1.5, result.Factors["resolution"], 0.000001)
	require.InDelta(t, 2.0, result.Factors["duration"], 0.000001)
}

func TestCalculateExpr(t *testing.T) {
	result, err := Calculate(Rule{
		Mode: ModeExpr,
		Expr: `seconds * 0.05 * (param("resolution") == "1080p" ? 1.5 : 1)`,
	}, Input{
		Request: relaycommon.TaskSubmitReq{Duration: 4},
		Body:    []byte(`{"resolution":"1080p"}`),
	})

	require.NoError(t, err)
	require.InDelta(t, 0.3, result.AmountUSD, 0.000001)
	require.Equal(t, 4, result.Seconds)
}

func TestResolveSecondsPriority(t *testing.T) {
	req := relaycommon.TaskSubmitReq{
		Duration: 5,
		Seconds:  "6",
		Metadata: map[string]interface{}{
			"duration":        7,
			"durationSeconds": 8,
		},
	}

	require.Equal(t, 8, ResolveSeconds(req, constant.ChannelTypeAli))
}

func TestResolveSecondsExplicitFields(t *testing.T) {
	require.Equal(t, 5, ResolveSeconds(relaycommon.TaskSubmitReq{Duration: 5}, constant.ChannelTypeAli))
	require.Equal(t, 6, ResolveSeconds(relaycommon.TaskSubmitReq{Seconds: "6"}, constant.ChannelTypeAli))
	require.Equal(t, 7, ResolveSeconds(relaycommon.TaskSubmitReq{
		Metadata: map[string]interface{}{
			"duration": 7,
		},
	}, constant.ChannelTypeAli))
	require.Equal(t, 8, ResolveSeconds(relaycommon.TaskSubmitReq{
		Metadata: map[string]interface{}{
			"durationSeconds": 8,
		},
	}, constant.ChannelTypeAli))
}

func TestResolveSecondsUsesProviderDefaults(t *testing.T) {
	cases := []struct {
		name        string
		channelType int
		want        int
	}{
		{name: "ali", channelType: constant.ChannelTypeAli, want: 5},
		{name: "gemini", channelType: constant.ChannelTypeGemini, want: 8},
		{name: "vertex", channelType: constant.ChannelTypeVertexAi, want: 8},
		{name: "sora", channelType: constant.ChannelTypeSora, want: 4},
		{name: "vidu", channelType: constant.ChannelTypeVidu, want: 5},
		{name: "hailuo", channelType: constant.ChannelTypeMiniMax, want: 6},
		{name: "kling", channelType: constant.ChannelTypeKling, want: 5},
		{name: "jimeng", channelType: constant.ChannelTypeJimeng, want: 5},
		{name: "unknown", channelType: constant.ChannelTypeUnknown, want: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, ResolveSeconds(relaycommon.TaskSubmitReq{}, tc.channelType))
		})
	}
}
