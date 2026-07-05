package promptgate

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	promptgatesetting "github.com/QuantumNous/new-api/setting/promptgate_setting"
)

const safeMessage = "抱歉，该请求或回答可能涉及不适宜内容，已被安全策略拦截。"

type Config struct {
	Enabled             bool
	BaseURL             string
	APIKey              string
	TimeoutMS           int
	InputEnabled        bool
	OutputEnabled       bool
	StreamOutputEnabled bool
	StreamFailClosed    bool
}

type Client struct {
	config     Config
	http       *http.Client
	streamHTTP *http.Client
}

type CheckRequest struct {
	Content        string            `json:"content"`
	Direction      string            `json:"direction,omitempty"`
	ContentType    string            `json:"content_type,omitempty"`
	ConversationID string            `json:"conversation_id,omitempty"`
	SubjectUserID  string            `json:"subject_user_id,omitempty"`
	SubjectIP      string            `json:"subject_ip,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type CheckResponse struct {
	Decision          string        `json:"decision"`
	RiskLevel         string        `json:"risk_level"`
	Categories        []string      `json:"categories"`
	MatchedRules      []MatchedRule `json:"matched_rules"`
	TraceID           string        `json:"trace_id"`
	LatencyMS         int64         `json:"latency_ms"`
	Direction         string        `json:"direction"`
	ContentType       string        `json:"content_type"`
	EnforcementAction string        `json:"enforcement_action"`
}

type MatchedRule struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Decision  string `json:"decision"`
	RiskLevel string `json:"risk_level"`
}

type StreamSession struct {
	SessionID string `json:"session_id"`
	ChunksURL string `json:"chunks_url"`
	EventsURL string `json:"events_url"`
	ExpiresAt string `json:"expires_at"`
}

type StreamChunk struct {
	Sequence     int    `json:"sequence"`
	ContentDelta string `json:"content_delta"`
	Kind         string `json:"kind"`
	Final        bool   `json:"final"`
}

type StreamEvent struct {
	Type         string `json:"-"`
	FromSequence int    `json:"from_sequence,omitempty"`
	ToSequence   int    `json:"to_sequence,omitempty"`
	Sequence     int    `json:"sequence,omitempty"`
	Content      string `json:"content,omitempty"`
	Decision     string `json:"decision,omitempty"`
	SafeMessage  string `json:"safe_message,omitempty"`
	TraceID      string `json:"trace_id,omitempty"`
	Error        string `json:"error,omitempty"`
}

func LoadConfigFromEnv() Config {
	timeoutMS := envInt("PROMPTGATE_TIMEOUT_MS", 3000)
	setting := promptgatesetting.GetSetting()
	baseURL := setting.NormalizedBaseURL()
	if baseURL == "" {
		baseURL = strings.TrimRight(os.Getenv("PROMPTGATE_BASE_URL"), "/")
	}
	apiKey := setting.NormalizedAPIKey()
	if apiKey == "" {
		apiKey = os.Getenv("PROMPTGATE_API_KEY")
	}
	return Config{
		Enabled:             setting.Enabled,
		BaseURL:             baseURL,
		APIKey:              apiKey,
		TimeoutMS:           timeoutMS,
		InputEnabled:        setting.InputEnabled,
		OutputEnabled:       setting.OutputEnabled,
		StreamOutputEnabled: setting.StreamOutputEnabled,
		StreamFailClosed:    setting.StreamFailClosed,
	}
}

func NewClient(config Config) *Client {
	if config.TimeoutMS <= 0 {
		config.TimeoutMS = 3000
	}
	return &Client{
		config: config,
		http: &http.Client{
			Timeout: time.Duration(config.TimeoutMS) * time.Millisecond,
		},
		streamHTTP: &http.Client{},
	}
}

func NewClientFromEnv() *Client {
	return NewClient(LoadConfigFromEnv())
}

func (c *Client) Enabled() bool {
	return c != nil && c.config.Enabled && c.config.BaseURL != "" && c.config.APIKey != ""
}

func (c *Client) InputEnabled() bool {
	return c.Enabled() && c.config.InputEnabled
}

func (c *Client) OutputEnabled() bool {
	return c.Enabled() && c.config.OutputEnabled
}

func (c *Client) StreamOutputEnabled() bool {
	return c.OutputEnabled() && c.config.StreamOutputEnabled
}

func (c *Client) StreamFailClosed() bool {
	return c.config.StreamFailClosed
}

func (c *Client) SafeMessage() string {
	return safeMessage
}

func (c *Client) CheckContent(ctx context.Context, request CheckRequest) (CheckResponse, error) {
	var response CheckResponse
	if !c.Enabled() {
		return CheckResponse{Decision: "allow"}, nil
	}
	if request.Direction == "" {
		request.Direction = "input"
	}
	if request.ContentType == "" {
		request.ContentType = "text"
	}
	err := c.postJSON(ctx, c.config.BaseURL+"/v1/moderation/check", request, &response)
	return response, err
}

func (c *Client) StartStreamSession(ctx context.Context, request CheckRequest) (StreamSession, error) {
	var response StreamSession
	if request.Direction == "" {
		request.Direction = "output"
	}
	if request.ContentType == "" {
		request.ContentType = "text"
	}
	err := c.postJSON(ctx, c.config.BaseURL+"/v1/moderation/check?stream=true", request, &response)
	if err != nil {
		return StreamSession{}, err
	}
	response.ChunksURL = c.resolveURL(response.ChunksURL)
	response.EventsURL = c.resolveURL(response.EventsURL)
	return response, nil
}

func (c *Client) SubmitStreamChunk(ctx context.Context, chunksURL string, chunk StreamChunk) error {
	if chunk.Kind == "" {
		chunk.Kind = "text"
	}
	var response map[string]any
	return c.postJSON(ctx, c.resolveURL(chunksURL), chunk, &response)
}

func (c *Client) SubscribeStreamEvents(ctx context.Context, eventsURL string) (<-chan StreamEvent, func(), error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.resolveURL(eventsURL), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("X-PromptGate-Key", c.config.APIKey)
	resp, err := c.streamHTTP.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, nil, fmt.Errorf("promptgate stream events status %d", resp.StatusCode)
	}

	events := make(chan StreamEvent, 16)
	cancel := func() { _ = resp.Body.Close() }
	go func() {
		defer close(events)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		eventType := ""
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event: ") {
				eventType = strings.TrimSpace(strings.TrimPrefix(line, "event: "))
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			var event StreamEvent
			if err := json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data: "))), &event); err != nil {
				events <- StreamEvent{Type: "error", Error: err.Error()}
				return
			}
			event.Type = eventType
			events <- event
			if event.Type == "done" || event.Type == "block" || event.Type == "error" {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			events <- StreamEvent{Type: "error", Error: err.Error()}
		}
	}()
	return events, cancel, nil
}

func (c *Client) postJSON(ctx context.Context, endpoint string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PromptGate-Key", c.config.APIKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("promptgate status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) resolveURL(value string) string {
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}
	base, err := url.Parse(c.config.BaseURL)
	if err != nil {
		return value
	}
	relative, err := url.Parse(value)
	if err != nil {
		return value
	}
	return base.ResolveReference(relative).String()
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

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
