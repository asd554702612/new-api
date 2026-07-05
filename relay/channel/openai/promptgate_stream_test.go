package openai

import (
	"testing"

	relayconstant "github.com/QuantumNous/new-api/relay/constant"
)

func TestPromptGateSupportsOpenAIStreamOutputOnlyForTextModes(t *testing.T) {
	allowed := []int{
		relayconstant.RelayModeChatCompletions,
		relayconstant.RelayModeCompletions,
	}
	for _, mode := range allowed {
		if !promptGateSupportsOpenAIStreamOutput(mode) {
			t.Fatalf("expected relay mode %d to support PromptGate stream output", mode)
		}
	}

	blocked := []int{
		relayconstant.RelayModeResponses,
		relayconstant.RelayModeResponsesCompact,
		relayconstant.RelayModeImagesGenerations,
		relayconstant.RelayModeImagesEdits,
		relayconstant.RelayModeAudioSpeech,
		relayconstant.RelayModeAudioTranscription,
		relayconstant.RelayModeAudioTranslation,
	}
	for _, mode := range blocked {
		if promptGateSupportsOpenAIStreamOutput(mode) {
			t.Fatalf("expected relay mode %d to skip PromptGate stream output", mode)
		}
	}
}
