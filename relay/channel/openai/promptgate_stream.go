package openai

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/service/promptgate"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func OaiStreamHandlerWithPromptGate(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response, client *promptgate.Client) (*dto.Usage, *types.NewAPIError, bool) {
	if client == nil || !client.StreamOutputEnabled() || info.RelayFormat != types.RelayFormatOpenAI || !promptGateSupportsOpenAIStreamOutput(info.RelayMode) {
		return nil, nil, false
	}
	session, err := client.StartStreamSession(c.Request.Context(), promptgate.CheckRequest{
		Content:        "stream output",
		Direction:      "output",
		ContentType:    "text",
		ConversationID: info.RequestId,
		SubjectUserID:  fmt.Sprintf("%d", info.UserId),
		Metadata: map[string]string{
			"source": "new-api",
			"route":  info.RequestURLPath,
			"model":  info.OriginModelName,
			"stream": "true",
		},
	})
	if err != nil {
		if !client.StreamFailClosed() {
			return nil, nil, false
		}
		return nil, types.NewErrorWithStatusCode(err, types.ErrorCodePromptBlocked, http.StatusForbidden, types.ErrOptionWithSkipRetry()), true
	}
	events, cancelEvents, err := client.SubscribeStreamEvents(c.Request.Context(), session.EventsURL)
	if err != nil {
		if !client.StreamFailClosed() {
			return nil, nil, false
		}
		return nil, types.NewErrorWithStatusCode(err, types.ErrorCodePromptBlocked, http.StatusForbidden, types.ErrOptionWithSkipRetry()), true
	}
	defer cancelEvents()

	usage, apiErr := oaiStreamHandlerPromptGateSession(c, info, resp, client, session, events)
	return usage, apiErr, true
}

func promptGateSupportsOpenAIStreamOutput(relayMode int) bool {
	switch relayMode {
	case relayconstant.RelayModeChatCompletions, relayconstant.RelayModeCompletions:
		return true
	default:
		return false
	}
}

func oaiStreamHandlerPromptGateSession(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response, client *promptgate.Client, session promptgate.StreamSession, events <-chan promptgate.StreamEvent) (*dto.Usage, *types.NewAPIError) {
	if resp == nil || resp.Body == nil {
		return nil, types.NewOpenAIError(fmt.Errorf("invalid response"), types.ErrorCodeBadResponse, http.StatusInternalServerError)
	}

	var responseId string
	var createAt int64
	var systemFingerprint string
	model := info.UpstreamModelName
	usage := &dto.Usage{}
	var containStreamUsage bool
	var responseTextBuilder strings.Builder
	var toolCount int
	var lastStreamData string
	var sequence int
	var blocked bool
	var gateFailedOpen bool
	pending := map[int]string{}
	pendingControls := make([]promptGateControlFrame, 0)

	flushRawErr := func(data string) error {
		return HandleStreamFormat(c, info, data, info.ChannelSetting.ForceFormat, info.ChannelSetting.ThinkingToContent)
	}
	flushRaw := func(data string, sr *helper.StreamResult) bool {
		if err := flushRawErr(data); err != nil {
			sr.Stop(err)
			return false
		}
		return true
	}
	flushControlsThrough := func(maxSeq int, sr *helper.StreamResult) bool {
		remaining := pendingControls[:0]
		for _, frame := range pendingControls {
			if frame.afterSeq > maxSeq {
				remaining = append(remaining, frame)
				continue
			}
			if !flushRaw(frame.data, sr) {
				return false
			}
		}
		pendingControls = remaining
		return true
	}
	flushControlsThroughErr := func(maxSeq int) error {
		remaining := pendingControls[:0]
		for _, frame := range pendingControls {
			if frame.afterSeq > maxSeq {
				remaining = append(remaining, frame)
				continue
			}
			if err := flushRawErr(frame.data); err != nil {
				return err
			}
		}
		pendingControls = remaining
		return nil
	}
	flushBufferedFailOpen := func(sr *helper.StreamResult) {
		if !flushControlsThrough(0, sr) {
			return
		}
		for seq := 1; seq <= sequence; seq++ {
			data, ok := pending[seq]
			if ok {
				delete(pending, seq)
				if !flushRaw(data, sr) {
					return
				}
			}
			if !flushControlsThrough(seq, sr) {
				return
			}
		}
	}
	handleEvent := func(event promptgate.StreamEvent, sr *helper.StreamResult) {
		switch event.Type {
		case "allow":
			if !flushControlsThrough(event.FromSequence-1, sr) {
				return
			}
			for seq := event.FromSequence; seq <= event.ToSequence; seq++ {
				data, ok := pending[seq]
				if !ok {
					continue
				}
				delete(pending, seq)
				if !flushRaw(data, sr) {
					return
				}
			}
			_ = flushControlsThrough(event.ToSequence, sr)
		case "block":
			blocked = true
			sendPromptGateSafeStreamMessage(c, info, firstNonEmpty(event.SafeMessage, client.SafeMessage()))
			sr.Stop(fmt.Errorf("promptgate stream blocked trace_id=%s", event.TraceID))
		case "error":
			if client.StreamFailClosed() {
				blocked = true
				sendPromptGateSafeStreamMessage(c, info, client.SafeMessage())
				sr.Stop(fmt.Errorf("promptgate stream error: %s", event.Error))
				return
			}
			gateFailedOpen = true
			flushBufferedFailOpen(sr)
		}
	}
	handleEventsClosed := func(sr *helper.StreamResult) {
		if client.StreamFailClosed() {
			blocked = true
			sendPromptGateSafeStreamMessage(c, info, client.SafeMessage())
			sr.Stop(fmt.Errorf("promptgate stream events closed"))
			return
		}
		gateFailedOpen = true
		flushBufferedFailOpen(sr)
	}
	drainEvents := func(sr *helper.StreamResult) {
		for {
			select {
			case event, ok := <-events:
				if !ok {
					handleEventsClosed(sr)
					return
				}
				handleEvent(event, sr)
				if sr.IsStopped() {
					return
				}
			default:
				return
			}
		}
	}
	waitBrieflyForEvents := func(sr *helper.StreamResult) {
		timer := time.NewTimer(50 * time.Millisecond)
		defer timer.Stop()
		for {
			select {
			case event, ok := <-events:
				if !ok {
					handleEventsClosed(sr)
					return
				}
				handleEvent(event, sr)
				if sr.IsStopped() {
					return
				}
			case <-timer.C:
				return
			}
		}
	}
	waitUntilDoneFinal := func() bool {
		for {
			select {
			case event, ok := <-events:
				if !ok {
					if client.StreamFailClosed() {
						sendPromptGateSafeStreamMessage(c, info, client.SafeMessage())
						return false
					}
					gateFailedOpen = true
					if err := flushControlsThroughErr(0); err != nil {
						return false
					}
					for seq := 1; seq <= sequence; seq++ {
						if data, ok := pending[seq]; ok {
							delete(pending, seq)
							if err := flushRawErr(data); err != nil {
								return false
							}
						}
						if err := flushControlsThroughErr(seq); err != nil {
							return false
						}
					}
					return true
				}
				switch event.Type {
				case "allow":
					if err := flushControlsThroughErr(event.FromSequence - 1); err != nil {
						return false
					}
					for seq := event.FromSequence; seq <= event.ToSequence; seq++ {
						data, ok := pending[seq]
						if !ok {
							continue
						}
						delete(pending, seq)
						if err := flushRawErr(data); err != nil {
							return false
						}
					}
					if err := flushControlsThroughErr(event.ToSequence); err != nil {
						return false
					}
				case "block":
					sendPromptGateSafeStreamMessage(c, info, firstNonEmpty(event.SafeMessage, client.SafeMessage()))
					return false
				case "error":
					if client.StreamFailClosed() {
						sendPromptGateSafeStreamMessage(c, info, client.SafeMessage())
						return false
					}
					gateFailedOpen = true
					if err := flushControlsThroughErr(0); err != nil {
						return false
					}
					for seq := 1; seq <= sequence; seq++ {
						if data, ok := pending[seq]; ok {
							delete(pending, seq)
							if err := flushRawErr(data); err != nil {
								return false
							}
						}
						if err := flushControlsThroughErr(seq); err != nil {
							return false
						}
					}
				case "done":
					if err := flushControlsThroughErr(sequence); err != nil {
						return false
					}
					return true
				}
			case <-c.Request.Context().Done():
				return false
			}
		}
	}

	helper.StreamScannerHandler(c, resp, info, func(data string, sr *helper.StreamResult) {
		if blocked {
			sr.Stop(fmt.Errorf("promptgate stream already blocked"))
			return
		}
		lastStreamData = data
		if err := processTokenData(info.RelayMode, data, &responseTextBuilder, &toolCount); err != nil {
			logger.LogError(c, "error processing stream token data: "+err.Error())
			sr.Error(err)
		}
		if gateFailedOpen {
			_ = flushRaw(data, sr)
			return
		}
		text := streamDataText(info.RelayMode, data)
		if strings.TrimSpace(text) == "" {
			pendingControls = append(pendingControls, promptGateControlFrame{afterSeq: sequence, data: data})
			drainEvents(sr)
			return
		}
		sequence++
		pending[sequence] = data
		if err := client.SubmitStreamChunk(c.Request.Context(), session.ChunksURL, promptgate.StreamChunk{
			Sequence:     sequence,
			ContentDelta: text,
			Kind:         "text",
		}); err != nil {
			if client.StreamFailClosed() {
				blocked = true
				sendPromptGateSafeStreamMessage(c, info, client.SafeMessage())
				sr.Stop(err)
				return
			}
			gateFailedOpen = true
			flushBufferedFailOpen(sr)
			return
		}
		waitBrieflyForEvents(sr)
	})

	if blocked {
		return service.ResponseText2Usage(c, responseTextBuilder.String(), info.UpstreamModelName, info.GetEstimatePromptTokens()), nil
	}

	if gateFailedOpen {
		return finishPromptGateStream(c, info, lastStreamData, responseTextBuilder.String(), toolCount)
	}

	sequence++
	if err := client.SubmitStreamChunk(c.Request.Context(), session.ChunksURL, promptgate.StreamChunk{
		Sequence: sequence,
		Kind:     "text",
		Final:    true,
	}); err != nil {
		if !client.StreamFailClosed() {
			return finishPromptGateStream(c, info, lastStreamData, responseTextBuilder.String(), toolCount)
		}
		sendPromptGateSafeStreamMessage(c, info, client.SafeMessage())
		return service.ResponseText2Usage(c, responseTextBuilder.String(), info.UpstreamModelName, info.GetEstimatePromptTokens()), nil
	}
	if !waitUntilDoneFinal() {
		return service.ResponseText2Usage(c, responseTextBuilder.String(), info.UpstreamModelName, info.GetEstimatePromptTokens()), nil
	}

	if lastStreamData != "" {
		shouldSendLastResp := false
		if err := handleLastResponse(lastStreamData, &responseId, &createAt, &systemFingerprint, &model, &usage, &containStreamUsage, info, &shouldSendLastResp); err != nil {
			logger.LogError(c, fmt.Sprintf("error handling last response: %s, lastStreamData: [%s]", err.Error(), lastStreamData))
		}
	}
	if !containStreamUsage {
		usage = service.ResponseText2Usage(c, responseTextBuilder.String(), info.UpstreamModelName, info.GetEstimatePromptTokens())
		usage.CompletionTokens += toolCount * 7
	}

	applyUsagePostProcessing(info, usage, common.StringToByteSlice(lastStreamData))
	HandleFinalResponse(c, info, lastStreamData, responseId, createAt, model, systemFingerprint, usage, containStreamUsage)
	return usage, nil
}

type promptGateControlFrame struct {
	afterSeq int
	data     string
}

func finishPromptGateStream(c *gin.Context, info *relaycommon.RelayInfo, lastStreamData string, responseText string, toolCount int) (*dto.Usage, *types.NewAPIError) {
	var responseId string
	var createAt int64
	var systemFingerprint string
	model := info.UpstreamModelName
	usage := &dto.Usage{}
	var containStreamUsage bool
	if lastStreamData != "" {
		shouldSendLastResp := false
		if err := handleLastResponse(lastStreamData, &responseId, &createAt, &systemFingerprint, &model, &usage, &containStreamUsage, info, &shouldSendLastResp); err != nil {
			logger.LogError(c, fmt.Sprintf("error handling last response: %s, lastStreamData: [%s]", err.Error(), lastStreamData))
		}
	}
	if !containStreamUsage {
		usage = service.ResponseText2Usage(c, responseText, info.UpstreamModelName, info.GetEstimatePromptTokens())
		usage.CompletionTokens += toolCount * 7
	}
	applyUsagePostProcessing(info, usage, common.StringToByteSlice(lastStreamData))
	HandleFinalResponse(c, info, lastStreamData, responseId, createAt, model, systemFingerprint, usage, containStreamUsage)
	return usage, nil
}

func streamDataText(relayMode int, data string) string {
	var text strings.Builder
	var toolCount int
	if err := processTokenData(relayMode, data, &text, &toolCount); err != nil {
		return ""
	}
	return text.String()
}

func sendPromptGateSafeStreamMessage(c *gin.Context, info *relaycommon.RelayInfo, message string) {
	if message == "" {
		message = "抱歉，该请求或回答可能涉及不适宜内容，已被安全策略拦截。"
	}
	response := dto.ChatCompletionsStreamResponse{
		Id:      "promptgate_blocked",
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   info.UpstreamModelName,
		Choices: []dto.ChatCompletionsStreamResponseChoice{{
			Index: 0,
			Delta: dto.ChatCompletionsStreamResponseChoiceDelta{Content: &message},
		}},
	}
	_ = helper.ObjectData(c, response)
	helper.Done(c)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
