package service

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/require"
)

func TestModelSquareDefaultDomesticDenyRulesHideClaudeModels(t *testing.T) {
	original := system_setting.GetModelSquareSettings()
	t.Cleanup(func() {
		system_setting.SetModelSquareSettingsForTest(original)
	})

	system_setting.SetModelSquareSettingsForTest(system_setting.ModelSquareSettings{
		Environment:       system_setting.ModelSquareEnvironmentDomestic,
		DomesticDenyRules: "",
	})

	for _, modelName := range []string{
		"claude-3-5-sonnet",
		"anthropic/claude-3-5-sonnet",
		"bedrock.anthropic.claude-v2",
	} {
		require.True(t, IsModelSquareDenied(modelName), modelName)
	}
}

func TestModelSquareDomesticDenyRulesDoNotApplyOverseas(t *testing.T) {
	original := system_setting.GetModelSquareSettings()
	t.Cleanup(func() {
		system_setting.SetModelSquareSettingsForTest(original)
	})

	system_setting.SetModelSquareSettingsForTest(system_setting.ModelSquareSettings{
		Environment:       system_setting.ModelSquareEnvironmentOverseas,
		DomesticDenyRules: "*claude*",
	})

	require.False(t, IsModelSquareDenied("claude-3-5-sonnet"))
	require.False(t, IsModelSquareDenied("anthropic/claude-3-5-sonnet"))
}
