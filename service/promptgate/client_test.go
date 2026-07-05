package promptgate

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	promptgatesetting "github.com/QuantumNous/new-api/setting/promptgate_setting"
)

func TestClientCheckContentSendsPromptGateKeyAndParsesDecision(t *testing.T) {
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-PromptGate-Key")
		if r.URL.Path != "/v1/moderation/check" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var body CheckRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body.Content != "hello" || body.Direction != "output" || body.ContentType != "text" {
			t.Fatalf("body = %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"decision":"block","risk_level":"high","categories":["risk"],"matched_rules":[],"trace_id":"pg_trace","latency_ms":3,"direction":"output","content_type":"text","enforcement_action":"none"}`))
	}))
	defer server.Close()

	client := NewClient(Config{Enabled: true, BaseURL: server.URL, APIKey: "pg_test", TimeoutMS: 1000})
	got, err := client.CheckContent(t.Context(), CheckRequest{
		Content:     "hello",
		Direction:   "output",
		ContentType: "text",
	})
	if err != nil {
		t.Fatalf("check content: %v", err)
	}
	if gotKey != "pg_test" {
		t.Fatalf("X-PromptGate-Key = %q", gotKey)
	}
	if got.Decision != "block" || got.TraceID != "pg_trace" {
		t.Fatalf("response = %#v", got)
	}
}

func TestClientStartStreamSessionReturnsSessionURLs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/moderation/check" || r.URL.Query().Get("stream") != "true" {
			t.Fatalf("url = %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"session_id":"pgs_1","chunks_url":"/v1/moderation/check/pgs_1/chunks","events_url":"/v1/moderation/check/pgs_1/events","expires_at":"2026-07-03T00:00:00Z"}`))
	}))
	defer server.Close()

	client := NewClient(Config{Enabled: true, BaseURL: server.URL, APIKey: "pg_test", TimeoutMS: 1000})
	got, err := client.StartStreamSession(t.Context(), CheckRequest{
		Content:     "stream",
		Direction:   "output",
		ContentType: "text",
	})
	if err != nil {
		t.Fatalf("start stream session: %v", err)
	}
	if got.SessionID != "pgs_1" || got.ChunksURL == "" || got.EventsURL == "" {
		t.Fatalf("session = %#v", got)
	}
}

func TestClientSubscribeStreamEventsDoesNotUseRequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/moderation/check/pgs_1/events" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected flusher")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher.Flush()
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte("event: allow\n"))
		_, _ = w.Write([]byte(`data: {"from_sequence":1,"to_sequence":1,"content":"hello"}` + "\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	client := NewClient(Config{Enabled: true, BaseURL: server.URL, APIKey: "pg_test", TimeoutMS: 10})
	events, cancel, err := client.SubscribeStreamEvents(t.Context(), "/v1/moderation/check/pgs_1/events")
	if err != nil {
		t.Fatalf("subscribe stream events: %v", err)
	}
	defer cancel()

	select {
	case event := <-events:
		if event.Type != "allow" || event.FromSequence != 1 || event.ToSequence != 1 {
			t.Fatalf("event = %#v", event)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for stream event")
	}
}

func TestLoadConfigFromEnvUsesPromptGateSettingsWithEnvFallback(t *testing.T) {
	t.Setenv("PROMPTGATE_ENABLED", "false")
	t.Setenv("PROMPTGATE_BASE_URL", "http://env-promptgate")
	t.Setenv("PROMPTGATE_API_KEY", "env-key")
	t.Setenv("PROMPTGATE_INPUT_ENABLED", "false")
	t.Setenv("PROMPTGATE_OUTPUT_ENABLED", "false")
	t.Setenv("PROMPTGATE_STREAM_OUTPUT_ENABLED", "false")
	t.Setenv("PROMPTGATE_STREAM_FAIL_CLOSED", "false")

	original := promptgatesetting.GetSetting().Clone()
	t.Cleanup(func() {
		promptgatesetting.SetSettingForTest(original)
	})

	promptgatesetting.SetSettingForTest(promptgatesetting.Setting{
		Enabled:             true,
		BaseURL:             "http://db-promptgate/",
		APIKey:              "db-key",
		InputEnabled:        true,
		OutputEnabled:       true,
		StreamOutputEnabled: true,
		StreamFailClosed:    true,
	})

	got := LoadConfigFromEnv()

	if !got.Enabled || got.BaseURL != "http://db-promptgate" || got.APIKey != "db-key" {
		t.Fatalf("config endpoint = %#v", got)
	}
	if !got.InputEnabled || !got.OutputEnabled || !got.StreamOutputEnabled || !got.StreamFailClosed {
		t.Fatalf("config toggles = %#v", got)
	}

	promptgatesetting.SetSettingForTest(promptgatesetting.Setting{
		Enabled:             true,
		BaseURL:             "",
		APIKey:              "",
		InputEnabled:        true,
		OutputEnabled:       true,
		StreamOutputEnabled: true,
		StreamFailClosed:    true,
	})

	got = LoadConfigFromEnv()
	if got.BaseURL != "http://env-promptgate" || got.APIKey != "env-key" {
		t.Fatalf("expected env fallback for empty endpoint fields, got %#v", got)
	}
}

func TestLoadConfigFromEnvAllowsSettingsToDisableEnvEnabledPromptGate(t *testing.T) {
	t.Setenv("PROMPTGATE_ENABLED", "true")
	t.Setenv("PROMPTGATE_BASE_URL", "http://env-promptgate")
	t.Setenv("PROMPTGATE_API_KEY", "env-key")

	original := promptgatesetting.GetSetting().Clone()
	t.Cleanup(func() {
		promptgatesetting.SetSettingForTest(original)
	})

	promptgatesetting.SetSettingForTest(promptgatesetting.Setting{
		Enabled:             false,
		BaseURL:             "http://db-promptgate",
		APIKey:              "db-key",
		InputEnabled:        true,
		OutputEnabled:       true,
		StreamOutputEnabled: true,
		StreamFailClosed:    true,
	})

	got := LoadConfigFromEnv()

	if got.Enabled {
		t.Fatalf("expected setting to disable env-enabled PromptGate, got %#v", got)
	}
}
