package service

import (
	"path"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

func ModelSquareSelectionEnabled() bool {
	return system_setting.GetModelSquareSettings().SelectionEnabled
}

func IsModelSquareDenied(modelName string) bool {
	settings := system_setting.GetModelSquareSettings()
	if settings.Environment != system_setting.ModelSquareEnvironmentDomestic {
		return false
	}
	modelName = strings.ToLower(strings.TrimSpace(modelName))
	if modelName == "" {
		return false
	}
	for _, rule := range parseModelSquareRules(settings.DomesticDenyRules) {
		if matchModelSquareRule(rule, modelName) {
			return true
		}
	}
	return false
}

func matchModelSquareRule(rule string, modelName string) bool {
	if matched, err := path.Match(rule, modelName); err == nil && matched {
		return true
	}
	if modelName == rule {
		return true
	}
	for _, separator := range []string{"/", ":"} {
		if index := strings.LastIndex(modelName, separator); index >= 0 && index+1 < len(modelName) {
			segment := modelName[index+1:]
			if matched, err := path.Match(rule, segment); err == nil && matched {
				return true
			}
			if segment == rule {
				return true
			}
		}
	}
	return false
}

func parseModelSquareRules(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, ",", "\n")
	lines := strings.Split(raw, "\n")
	rules := make([]string, 0, len(lines))
	seen := map[string]struct{}{}
	for _, line := range lines {
		line = strings.ToLower(strings.TrimSpace(line))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		rules = append(rules, line)
	}
	return rules
}

func FilterModelNamesBySquareAvailability(modelNames []string) []string {
	filtered := make([]string, 0, len(modelNames))
	for _, modelName := range modelNames {
		if IsModelSquareDenied(modelName) {
			continue
		}
		filtered = append(filtered, modelName)
	}
	return filtered
}

func FilterModelNamesByUserSelection(userId int, modelNames []string) ([]string, error) {
	modelNames = FilterModelNamesBySquareAvailability(modelNames)
	if !ModelSquareSelectionEnabled() {
		return modelNames, nil
	}
	selections, err := model.GetUserModelSelectionMap(userId)
	if err != nil {
		return nil, err
	}
	filtered := make([]string, 0, len(modelNames))
	for _, modelName := range modelNames {
		if selections[modelName] {
			filtered = append(filtered, modelName)
		}
	}
	return filtered, nil
}

func IsUserSelectedModel(userId int, modelName string) (bool, error) {
	if IsModelSquareDenied(modelName) {
		return false, nil
	}
	if !ModelSquareSelectionEnabled() {
		return true, nil
	}
	selections, err := model.GetUserModelSelectionMap(userId)
	if err != nil {
		return false, err
	}
	return selections[modelName], nil
}

func FilterPricingBySquareAvailability(pricing []model.Pricing) []model.Pricing {
	if len(pricing) == 0 {
		return pricing
	}
	filtered := make([]model.Pricing, 0, len(pricing))
	for _, item := range pricing {
		if IsModelSquareDenied(item.ModelName) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func MarkPricingSelections(userId int, pricing []model.Pricing) ([]model.Pricing, error) {
	if !ModelSquareSelectionEnabled() || userId <= 0 {
		return pricing, nil
	}
	selections, err := model.GetUserModelSelectionMap(userId)
	if err != nil {
		return nil, err
	}
	for i := range pricing {
		pricing[i].Selected = selections[pricing[i].ModelName]
	}
	return pricing, nil
}

func ReplaceUserModelSelectionsForUsableModels(userId int, requested []string) ([]string, error) {
	requested = model.NormalizeModelSelectionNames(requested)
	requestedSet := make(map[string]bool, len(requested))
	for _, modelName := range requested {
		if IsModelSquareDenied(modelName) {
			continue
		}
		requestedSet[modelName] = true
	}

	user, err := model.GetUserCache(userId)
	if err != nil {
		return nil, err
	}
	usableGroups := GetUserUsableGroups(user.Group)
	allowed := make([]string, 0, len(requestedSet))
	for group := range usableGroups {
		for _, modelName := range model.GetGroupEnabledModels(group) {
			if requestedSet[modelName] && !common.StringsContains(allowed, modelName) {
				allowed = append(allowed, modelName)
			}
		}
	}
	allowed = FilterModelNamesBySquareAvailability(allowed)
	if err := model.ReplaceUserModelSelections(userId, allowed); err != nil {
		return nil, err
	}
	return allowed, nil
}
