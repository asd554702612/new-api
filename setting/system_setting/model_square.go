package system_setting

import (
	"os"
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

const (
	ModelSquareEnvironmentOverseas = "overseas"
	ModelSquareEnvironmentDomestic = "domestic"
)

const DefaultModelSquareDomesticDenyRules = "gpt*\nchatgpt*\no1*\no3*\no4*\n*claude*"

type ModelSquareSettings struct {
	SelectionEnabled  bool   `json:"selection_enabled"`
	Environment       string `json:"environment"`
	DomesticDenyRules string `json:"domestic_deny_rules"`
}

var defaultModelSquareSettings = ModelSquareSettings{
	SelectionEnabled:  false,
	Environment:       ModelSquareEnvironmentOverseas,
	DomesticDenyRules: DefaultModelSquareDomesticDenyRules,
}

func init() {
	config.GlobalConfig.Register("model_square", &defaultModelSquareSettings)
}

func normalizeModelSquareSettings(settings ModelSquareSettings) ModelSquareSettings {
	settings.Environment = strings.ToLower(strings.TrimSpace(settings.Environment))
	if settings.Environment != ModelSquareEnvironmentDomestic {
		settings.Environment = ModelSquareEnvironmentOverseas
	}
	if strings.TrimSpace(settings.DomesticDenyRules) == "" {
		settings.DomesticDenyRules = DefaultModelSquareDomesticDenyRules
	}
	return settings
}

func GetModelSquareSettings() ModelSquareSettings {
	return normalizeModelSquareSettings(defaultModelSquareSettings)
}

func InitModelSquareSettingsFromEnv() {
	settings := defaultModelSquareSettings
	if value := strings.TrimSpace(os.Getenv("MODEL_SQUARE_SELECTION_ENABLED")); value != "" {
		settings.SelectionEnabled = strings.EqualFold(value, "true")
	}
	if value := strings.TrimSpace(os.Getenv("MODEL_SQUARE_ENVIRONMENT")); value != "" {
		settings.Environment = value
	}
	if value := strings.TrimSpace(os.Getenv("MODEL_SQUARE_DOMESTIC_DENY_RULES")); value != "" {
		settings.DomesticDenyRules = value
	}
	SetModelSquareSettings(settings)
}

func SetModelSquareSettings(settings ModelSquareSettings) {
	defaultModelSquareSettings = normalizeModelSquareSettings(settings)
}

func SetModelSquareSettingsForTest(settings ModelSquareSettings) {
	SetModelSquareSettings(settings)
}
