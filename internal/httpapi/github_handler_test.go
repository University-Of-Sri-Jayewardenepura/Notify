package httpapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubGitHubService struct {
	called    bool
	eventType string
}

func (s *stubGitHubService) HandleGitHubEvent(eventType string, payload map[string]any) error {
	s.called = true
	s.eventType = eventType
	return nil
}

func TestGitHubHandlerReturnsBadRequestWhenEventHeaderMissing(t *testing.T) {
	service := &stubGitHubService{}
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service,
	})

	body := []byte(`{"repository":{"name":"notify"}}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", signWebhookBody(body, "super-secret"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGitHubHandlerReturnsUnauthorizedWhenSignatureMissing(t *testing.T) {
	service := &stubGitHubService{}
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service,
	})

	body := []byte(`{"repository":{"name":"notify"}}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "ping")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGitHubHandlerReturnsUnauthorizedWhenSignatureInvalid(t *testing.T) {
	service := &stubGitHubService{}
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service,
	})

	body := []byte(`{"repository":{"name":"notify"}}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGitHubHandlerReturnsBadRequestWhenJSONInvalid(t *testing.T) {
	service := &stubGitHubService{}
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service,
	})

	body := []byte(`{"repository":`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", signWebhookBody(body, "super-secret"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGitHubHandlerDelegatesToServiceForValidRequest(t *testing.T) {
	service := &stubGitHubService{}
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service,
	})

	body := []byte(`{"zen":"keep it logically awesome"}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", signWebhookBody(body, "super-secret"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !service.called {
		t.Fatal("expected service to be called for valid webhook")
	}

	if service.eventType != "ping" {
		t.Fatalf("expected event type ping, got %q", service.eventType)
	}
}

func signWebhookBody(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
