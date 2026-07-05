package openai

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/service/promptgate"
	"github.com/QuantumNous/new-api/types"
)

type promptGateOutputChecker interface {
	OutputEnabled() bool
	SafeMessage() string
	CheckContent(context.Context, promptgate.CheckRequest) (promptgate.CheckResponse, error)
}

func checkPromptGateOutput(ctx context.Context, checker promptGateOutputChecker, request promptgate.CheckRequest) *types.NewAPIError {
	if checker == nil || !checker.OutputEnabled() || strings.TrimSpace(request.Content) == "" {
		return nil
	}
	response, err := checker.CheckContent(ctx, request)
	if err != nil {
		return types.NewErrorWithStatusCode(err, types.ErrorCodePromptBlocked, http.StatusForbidden, types.ErrOptionWithSkipRetry())
	}
	if response.Decision == "" || response.Decision == "allow" {
		return nil
	}
	message := checker.SafeMessage()
	if message == "" {
		message = "抱歉，该请求或回答可能涉及不适宜内容，已被安全策略拦截。"
	}
	if response.TraceID != "" {
		message = fmt.Sprintf("%s trace_id=%s", message, response.TraceID)
	}
	return types.NewErrorWithStatusCode(fmt.Errorf("%s", message), types.ErrorCodePromptBlocked, http.StatusForbidden, types.ErrOptionWithSkipRetry())
}

func openAIResponseOutputText(response dto.OpenAITextResponse) string {
	parts := make([]string, 0, len(response.Choices)*2)
	for _, choice := range response.Choices {
		if content := strings.TrimSpace(choice.Message.StringContent()); content != "" {
			parts = append(parts, content)
		}
		if reasoning := strings.TrimSpace(choice.Message.GetReasoningContent()); reasoning != "" {
			parts = append(parts, reasoning)
		}
	}
	return strings.Join(parts, "\n")
}
