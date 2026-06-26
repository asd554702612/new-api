package videobilling

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/expr-lang/expr"
	"github.com/tidwall/gjson"
)

const (
	ModePerSecond = "per_second"
	ModeMatrix    = "matrix"
	ModeExpr      = "expr"
)

type Rule struct {
	Mode        string                        `json:"mode"`
	BasePrice   float64                       `json:"base_price,omitempty"`
	Multipliers map[string]map[string]float64 `json:"multipliers,omitempty"`
	Expr        string                        `json:"expr,omitempty"`
}

type Input struct {
	Model       string
	Action      string
	ChannelType int
	Request     relaycommon.TaskSubmitReq
	Body        []byte
}

type Result struct {
	AmountUSD float64
	Seconds   int
	Factors   map[string]float64
}

func Calculate(rule Rule, input Input) (Result, error) {
	seconds := ResolveSeconds(input.Request, input.ChannelType)
	if seconds <= 0 {
		return Result{}, fmt.Errorf("video duration must be greater than 0")
	}

	switch strings.TrimSpace(rule.Mode) {
	case ModePerSecond:
		if rule.BasePrice < 0 {
			return Result{}, fmt.Errorf("base_price must be non-negative")
		}
		return Result{
			AmountUSD: rule.BasePrice * float64(seconds),
			Seconds:   seconds,
			Factors:   map[string]float64{},
		}, nil
	case ModeMatrix:
		if rule.BasePrice < 0 {
			return Result{}, fmt.Errorf("base_price must be non-negative")
		}
		factors := resolveFactors(rule.Multipliers, input, seconds)
		amount := rule.BasePrice * float64(seconds)
		for _, factor := range factors {
			amount *= factor
		}
		return Result{AmountUSD: amount, Seconds: seconds, Factors: factors}, nil
	case ModeExpr:
		amount, err := runExpr(rule.Expr, input, seconds)
		if err != nil {
			return Result{}, err
		}
		if amount < 0 {
			return Result{}, fmt.Errorf("expr result must be non-negative")
		}
		return Result{
			AmountUSD: amount,
			Seconds:   seconds,
			Factors:   map[string]float64{},
		}, nil
	default:
		return Result{}, fmt.Errorf("unknown video billing mode: %s", rule.Mode)
	}
}

func ResolveSeconds(req relaycommon.TaskSubmitReq, channelType int) int {
	if req.Metadata != nil {
		if seconds := anyToPositiveInt(req.Metadata["durationSeconds"]); seconds > 0 {
			return seconds
		}
		if seconds := anyToPositiveInt(req.Metadata["duration"]); seconds > 0 {
			return seconds
		}
	}
	if req.Duration > 0 {
		return req.Duration
	}
	if seconds, err := strconv.Atoi(strings.TrimSpace(req.Seconds)); err == nil && seconds > 0 {
		return seconds
	}
	return defaultSeconds(channelType)
}

func defaultSeconds(channelType int) int {
	switch channelType {
	case constant.ChannelTypeAli:
		return 5
	case constant.ChannelTypeGemini, constant.ChannelTypeVertexAi:
		return 8
	case constant.ChannelTypeSora, constant.ChannelTypeOpenAI:
		return 4
	case constant.ChannelTypeVidu:
		return 5
	case constant.ChannelTypeMiniMax:
		return 6
	case constant.ChannelTypeKling:
		return 5
	case constant.ChannelTypeJimeng:
		return 5
	default:
		return 1
	}
}

func resolveFactors(multipliers map[string]map[string]float64, input Input, seconds int) map[string]float64 {
	factors := make(map[string]float64)
	for field, values := range multipliers {
		value := resolveFieldValue(field, input, seconds)
		if value == "" {
			continue
		}
		if factor, ok := values[value]; ok && factor > 0 {
			factors[field] = factor
			continue
		}
		if factor, ok := values[strings.ToLower(value)]; ok && factor > 0 {
			factors[field] = factor
		}
	}
	return factors
}

func resolveFieldValue(field string, input Input, seconds int) string {
	field = strings.TrimSpace(field)
	switch field {
	case "seconds", "duration":
		return strconv.Itoa(seconds)
	case "model":
		return input.Model
	case "action":
		return input.Action
	case "size":
		return strings.TrimSpace(input.Request.Size)
	case "mode":
		return strings.TrimSpace(input.Request.Mode)
	case "resolution":
		return normalizeString(readMetadataString(input.Request.Metadata, "resolution"))
	case "has_image":
		return strconv.FormatBool(input.Request.HasImage() || strings.TrimSpace(input.Request.Image) != "")
	case "has_input_reference":
		return strconv.FormatBool(strings.TrimSpace(input.Request.InputReference) != "")
	}
	if input.Request.Metadata != nil {
		return normalizeString(readMetadataString(input.Request.Metadata, field))
	}
	return ""
}

func runExpr(exprStr string, input Input, seconds int) (float64, error) {
	exprStr = strings.TrimSpace(exprStr)
	if exprStr == "" {
		return 0, fmt.Errorf("expr is required")
	}

	resolution := resolveFieldValue("resolution", input, seconds)
	env := map[string]interface{}{
		"seconds":             float64(seconds),
		"duration":            float64(seconds),
		"size":                strings.TrimSpace(input.Request.Size),
		"resolution":          resolution,
		"mode":                strings.TrimSpace(input.Request.Mode),
		"action":              input.Action,
		"has_image":           input.Request.HasImage() || strings.TrimSpace(input.Request.Image) != "",
		"has_input_reference": strings.TrimSpace(input.Request.InputReference) != "",
		"param": func(path string) interface{} {
			return resolveParam(path, input, seconds)
		},
		"max":   math.Max,
		"min":   math.Min,
		"abs":   math.Abs,
		"ceil":  math.Ceil,
		"floor": math.Floor,
	}
	prog, err := expr.Compile(exprStr, expr.Env(env), expr.AsFloat64())
	if err != nil {
		return 0, fmt.Errorf("expr compile error: %w", err)
	}
	out, err := expr.Run(prog, env)
	if err != nil {
		return 0, fmt.Errorf("expr run error: %w", err)
	}
	f, ok := out.(float64)
	if !ok {
		return 0, fmt.Errorf("expr result is %T, want float64", out)
	}
	return f, nil
}

func resolveParam(path string, input Input, seconds int) interface{} {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	if len(input.Body) > 0 {
		if result := gjson.GetBytes(input.Body, path); result.Exists() {
			return result.Value()
		}
	}
	if value := resolveFieldValue(path, input, seconds); value != "" {
		return value
	}
	if input.Request.Metadata != nil {
		metadataBytes, err := common.Marshal(input.Request.Metadata)
		if err == nil {
			if result := gjson.GetBytes(metadataBytes, path); result.Exists() {
				return result.Value()
			}
		}
	}
	return nil
}

func readMetadataString(metadata map[string]interface{}, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func normalizeString(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func anyToPositiveInt(value interface{}) int {
	switch typed := value.(type) {
	case int:
		if typed > 0 {
			return typed
		}
	case int64:
		if typed > 0 {
			return int(typed)
		}
	case float64:
		if typed > 0 {
			return int(typed)
		}
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 0
}
