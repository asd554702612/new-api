package controller

import (
	"testing"

	relayconstant "github.com/QuantumNous/new-api/relay/constant"
)

func TestPromptGateSupportsTextInputOnlyForTextRelayModes(t *testing.T) {
	allowed := []int{
		relayconstant.RelayModeChatCompletions,
		relayconstant.RelayModeCompletions,
		relayconstant.RelayModeResponses,
		relayconstant.RelayModeResponsesCompact,
	}
	for _, mode := range allowed {
		if !promptGateSupportsTextInput(mode) {
			t.Fatalf("expected relay mode %d to support PromptGate text input", mode)
		}
	}

	blocked := []int{
		relayconstant.RelayModeImagesGenerations,
		relayconstant.RelayModeImagesEdits,
		relayconstant.RelayModeAudioSpeech,
		relayconstant.RelayModeAudioTranscription,
		relayconstant.RelayModeAudioTranslation,
		relayconstant.RelayModeEmbeddings,
		relayconstant.RelayModeRerank,
	}
	for _, mode := range blocked {
		if promptGateSupportsTextInput(mode) {
			t.Fatalf("expected relay mode %d to skip PromptGate text input", mode)
		}
	}
}
