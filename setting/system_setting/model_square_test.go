package system_setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitModelSquareSettingsFromEnv(t *testing.T) {
	original := GetModelSquareSettings()
	t.Cleanup(func() {
		SetModelSquareSettings(original)
	})

	t.Setenv("MODEL_SQUARE_SELECTION_ENABLED", "true")
	t.Setenv("MODEL_SQUARE_ENVIRONMENT", ModelSquareEnvironmentDomestic)
	t.Setenv("MODEL_SQUARE_DOMESTIC_DENY_RULES", "gpt*\ncustom-openai*")

	InitModelSquareSettingsFromEnv()

	settings := GetModelSquareSettings()
	require.True(t, settings.SelectionEnabled)
	require.Equal(t, ModelSquareEnvironmentDomestic, settings.Environment)
	require.Equal(t, "gpt*\ncustom-openai*", settings.DomesticDenyRules)
}

func TestInitModelSquareSettingsFromEnvDefaultsInvalidEnvironmentToOverseas(t *testing.T) {
	original := GetModelSquareSettings()
	t.Cleanup(func() {
		SetModelSquareSettings(original)
	})

	t.Setenv("MODEL_SQUARE_ENVIRONMENT", "invalid")

	InitModelSquareSettingsFromEnv()

	require.Equal(t, ModelSquareEnvironmentOverseas, GetModelSquareSettings().Environment)
}
