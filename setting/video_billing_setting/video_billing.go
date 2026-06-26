package video_billing_setting

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/videobilling"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/samber/lo"
)

const RulesField = "rules"

type Rule = videobilling.Rule

type VideoBillingSetting struct {
	Rules map[string]Rule `json:"rules"`
}

var videoBillingSetting = VideoBillingSetting{
	Rules: map[string]Rule{},
}

func init() {
	config.GlobalConfig.Register("video_billing_setting", &videoBillingSetting)
}

func GetRulesCopy() map[string]Rule {
	return lo.Assign(videoBillingSetting.Rules)
}

func GetRule(model string) (Rule, bool) {
	rule, ok := videoBillingSetting.Rules[model]
	return rule, ok
}

func Rules2JSONString() string {
	jsonBytes, err := common.Marshal(videoBillingSetting.Rules)
	if err != nil {
		common.SysError("error marshalling video billing rules: " + err.Error())
		return "{}"
	}
	return string(jsonBytes)
}

func UpdateRulesByJSONString(jsonStr string) error {
	var rules map[string]Rule
	if err := common.UnmarshalJsonStr(jsonStr, &rules); err != nil {
		return err
	}
	if rules == nil {
		rules = map[string]Rule{}
	}
	if err := ValidateRules(rules); err != nil {
		return err
	}
	videoBillingSetting.Rules = rules
	return nil
}

func ValidateRulesJSON(jsonStr string) error {
	var rules map[string]Rule
	if err := common.UnmarshalJsonStr(jsonStr, &rules); err != nil {
		return err
	}
	if rules == nil {
		rules = map[string]Rule{}
	}
	return ValidateRules(rules)
}

func ValidateRules(rules map[string]Rule) error {
	for model, rule := range rules {
		if strings.TrimSpace(model) == "" {
			return fmt.Errorf("model name cannot be empty")
		}
		if err := validateRule(rule); err != nil {
			return fmt.Errorf("model %s: %w", model, err)
		}
	}
	return nil
}

func validateRule(rule Rule) error {
	switch strings.TrimSpace(rule.Mode) {
	case videobilling.ModePerSecond:
		if rule.BasePrice < 0 {
			return fmt.Errorf("base_price must be non-negative")
		}
	case videobilling.ModeMatrix:
		if rule.BasePrice < 0 {
			return fmt.Errorf("base_price must be non-negative")
		}
		for field, values := range rule.Multipliers {
			if strings.TrimSpace(field) == "" {
				return fmt.Errorf("multiplier field cannot be empty")
			}
			for value, factor := range values {
				if strings.TrimSpace(value) == "" {
					return fmt.Errorf("multiplier value cannot be empty")
				}
				if factor <= 0 {
					return fmt.Errorf("multiplier %s.%s must be greater than 0", field, value)
				}
			}
		}
	case videobilling.ModeExpr:
		_, err := videobilling.Calculate(rule, videobilling.Input{
			Request: relaycommon.TaskSubmitReq{Duration: 5},
			Body:    []byte(`{"resolution":"1080p","mode":"standard"}`),
		})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown video billing mode: %s", rule.Mode)
	}
	return nil
}
