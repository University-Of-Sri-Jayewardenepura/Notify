package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendPostsToWebhookPath(t *testing.T) {
	var gotPath string
	var gotPayload WebhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("expected valid JSON payload, got error: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "123", "token", server.Client())
	payload := WebhookPayload{Username: "GitHub Notify"}

	if err := client.Send(context.Background(), payload); err != nil {
		t.Fatalf("expected send to succeed, got error: %v", err)
	}

	if gotPath != "/123/token" {
		t.Fatalf("expected webhook path /123/token, got %q", gotPath)
	}

	if gotPayload.Username != "GitHub Notify" {
		t.Fatalf("expected username to be sent, got %#v", gotPayload.Username)
	}
}

func TestSendReturnsErrorForNonSuccessResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad webhook", http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL, "123", "token", server.Client())

	err := client.Send(context.Background(), WebhookPayload{})
	if err == nil {
		t.Fatal("expected non-success response to return an error")
	}
}

func TestSendReturnsErrorForNonTwoXXResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMultipleChoices)
	}))
	defer server.Close()

	client := NewClient(server.URL, "123", "token", server.Client())

	err := client.Send(context.Background(), WebhookPayload{})
	if err == nil {
		t.Fatal("expected non-2xx response to return an error")
	}
}
