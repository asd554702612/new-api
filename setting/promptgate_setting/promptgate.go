package promptgate_setting

import (
	"os"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

type Setting struct {
	Enabled             bool   `json:"enabled"`
	BaseURL             string `json:"base_url"`
	APIKey              string `json:"api_key"`
	InputEnabled        bool   `json:"input_enabled"`
	OutputEnabled       bool   `json:"output_enabled"`
	StreamOutputEnabled bool   `json:"stream_output_enabled"`
	StreamFailClosed    bool   `json:"stream_fail_closed"`
}

var promptGateSetting = Setting{
	Enabled:             envBool("PROMPTGATE_ENABLED", false),
	BaseURL:             strings.TrimRight(os.Getenv("PROMPTGATE_BASE_URL"), "/"),
	APIKey:              os.Getenv("PROMPTGATE_API_KEY"),
	InputEnabled:        envBool("PROMPTGATE_INPUT_ENABLED", true),
	OutputEnabled:       envBool("PROMPTGATE_OUTPUT_ENABLED", true),
	StreamOutputEnabled: envBool("PROMPTGATE_STREAM_OUTPUT_ENABLED", true),
	StreamFailClosed:    envBool("PROMPTGATE_STREAM_FAIL_CLOSED", true),
}

func init() {
	config.GlobalConfig.Register("promptgate", &promptGateSetting)
}

func GetSetting() *Setting {
	return &promptGateSetting
}

func SetSettingForTest(setting Setting) {
	promptGateSetting = setting
}

func (s Setting) Clone() Setting {
	return s
}

func (s Setting) NormalizedBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(s.BaseURL), "/")
}

func (s Setting) NormalizedAPIKey() string {
	return strings.TrimSpace(s.APIKey)
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
