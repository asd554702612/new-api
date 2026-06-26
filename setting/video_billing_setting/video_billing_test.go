package video_billing_setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRulesJSONAcceptsHappyHorseRules(t *testing.T) {
	raw := `{
		"happyhorse-1.0-t2v": {"mode":"per_second","base_price":0.03},
		"happyhorse-1.0-i2v": {
			"mode":"matrix",
			"base_price":0.04,
			"multipliers":{"resolution":{"720p":1,"1080p":1.5}}
		},
		"happyhorse-1.0-r2v": {
			"mode":"expr",
			"expr":"seconds * 0.05 * (param(\"resolution\") == \"1080p\" ? 1.5 : 1)"
		}
	}`

	require.NoError(t, ValidateRulesJSON(raw))
}

func TestValidateRulesJSONRejectsInvalidMode(t *testing.T) {
	err := ValidateRulesJSON(`{"happyhorse-1.0-t2v":{"mode":"unknown","base_price":0.03}}`)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown video billing mode")
}

func TestUpdateRulesByJSONStringReplacesRules(t *testing.T) {
	t.Cleanup(func() {
		videoBillingSetting.Rules = map[string]Rule{}
	})

	require.NoError(t, UpdateRulesByJSONString(`{"happyhorse-1.0-t2v":{"mode":"per_second","base_price":0.03}}`))
	rule, ok := GetRule("happyhorse-1.0-t2v")
	require.True(t, ok)
	require.Equal(t, "per_second", rule.Mode)

	require.NoError(t, UpdateRulesByJSONString(`{}`))
	_, ok = GetRule("happyhorse-1.0-t2v")
	require.False(t, ok)
}
