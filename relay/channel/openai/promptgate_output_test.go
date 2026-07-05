package openai

import (
	"context"
	"testing"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/service/promptgate"
)

type fakePromptGateChecker struct {
	response promptgate.CheckResponse
}

func (f fakePromptGateChecker) OutputEnabled() bool { return true }
func (f fakePromptGateChecker) SafeMessage() string { return "safe" }
func (f fakePromptGateChecker) CheckContent(context.Context, promptgate.CheckRequest) (promptgate.CheckResponse, error) {
	return f.response, nil
}

func TestCheckPromptGateOutputBlocksUnsafeResponse(t *testing.T) {
	checker := fakePromptGateChecker{response: promptgate.CheckResponse{Decision: "block", TraceID: "pg_trace"}}

	err := checkPromptGateOutput(context.Background(), checker, promptgate.CheckRequest{
		Content:     "unsafe answer",
		Direction:   "output",
		ContentType: "text",
	})

	if err == nil {
		t.Fatal("expected output block error")
	}
	if err.GetErrorCode() != "prompt_blocked" {
		t.Fatalf("error code = %s", err.GetErrorCode())
	}
}

func TestOpenAIResponseOutputTextIncludesReasoning(t *testing.T) {
	reasoning := "reasoning"
	response := dto.OpenAITextResponse{
		Choices: []dto.OpenAITextResponseChoice{
			{Message: dto.Message{Content: "answer", ReasoningContent: &reasoning}},
		},
	}

	got := openAIResponseOutputText(response)

	if got != "answer\nreasoning" {
		t.Fatalf("output text = %q", got)
	}
}
