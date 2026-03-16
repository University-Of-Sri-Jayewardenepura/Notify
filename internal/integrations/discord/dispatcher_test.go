package discord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDispatcherSendsRenderedEventToDiscord(t *testing.T) {
	var gotPayload WebhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("expected valid JSON payload, got error: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	dispatcher := NewDispatcher(NewClient(server.URL, "123", "token", server.Client()))

	err := dispatcher.DispatchGitHubEvent("ping", map[string]any{
		"zen": "keep it logically awesome",
	})
	if err != nil {
		t.Fatalf("expected dispatcher to succeed, got error: %v", err)
	}

	if len(gotPayload.Embeds) != 1 || gotPayload.Embeds[0].Title != "GitHub Webhook Connected" {
		t.Fatalf("expected ping embed to be sent, got %#v", gotPayload.Embeds)
	}
}

func TestDispatcherSkipsUnsupportedEventWithoutSending(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	dispatcher := NewDispatcher(NewClient(server.URL, "123", "token", server.Client()))

	err := dispatcher.DispatchGitHubEvent("unknown", map[string]any{})
	if err != nil {
		t.Fatalf("expected unsupported event to be ignored without error, got: %v", err)
	}

	if called {
		t.Fatal("expected unsupported event not to send a Discord webhook")
	}
}

func TestClientUsesDefaultBaseURLWhenEmpty(t *testing.T) {
	client := NewClient("", "123", "token", nil)
	if client.baseURL != defaultBaseURL {
		t.Fatalf("expected default base URL %q, got %q", defaultBaseURL, client.baseURL)
	}
}
