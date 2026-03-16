package httpapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/University-Of-Sri-Jayewardenepura/Notify/internal/service"
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

type stubDispatcher struct {
	called    bool
	eventType string
	payload   map[string]any
	err       error
}

func (d *stubDispatcher) DispatchGitHubEvent(eventType string, payload map[string]any) error {
	d.called = true
	d.eventType = eventType
	d.payload = payload
	return d.err
}

func TestGitHubHandlerProcessesPingFixture(t *testing.T) {
	dispatcher := &stubDispatcher{}
	body := readFixture(t, "ping.json")
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service.New("usj", dispatcher),
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", signWebhookBody(body, "super-secret"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !dispatcher.called || dispatcher.eventType != "ping" {
		t.Fatalf("expected ping fixture to dispatch, got called=%v event=%q", dispatcher.called, dispatcher.eventType)
	}
}

func TestGitHubHandlerIgnoresForeignOrganizationFixture(t *testing.T) {
	dispatcher := &stubDispatcher{}
	body := readFixture(t, "push-foreign.json")
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service.New("usj", dispatcher),
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", signWebhookBody(body, "super-secret"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if dispatcher.called {
		t.Fatal("expected foreign organization fixture not to dispatch")
	}
}

func TestGitHubHandlerDispatchesMatchingOrganizationFixture(t *testing.T) {
	dispatcher := &stubDispatcher{}
	body := readFixture(t, "push-org.json")
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service.New("usj", dispatcher),
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", signWebhookBody(body, "super-secret"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !dispatcher.called || dispatcher.eventType != "push" {
		t.Fatalf("expected matching organization fixture to dispatch push, got called=%v event=%q", dispatcher.called, dispatcher.eventType)
	}

	repository, ok := dispatcher.payload["repository"].(map[string]any)
	if !ok {
		t.Fatalf("expected dispatched payload to include repository map, got %#v", dispatcher.payload["repository"])
	}

	if repository["full_name"] != "usj/repo" {
		t.Fatalf("expected repository full_name usj/repo, got %#v", repository["full_name"])
	}

	commits, ok := dispatcher.payload["commits"].([]any)
	if !ok || len(commits) != 1 {
		t.Fatalf("expected dispatched payload to include one commit, got %#v", dispatcher.payload["commits"])
	}
}

func TestGitHubHandlerReturnsBadRequestForFixtureWhenEventHeaderMissing(t *testing.T) {
	dispatcher := &stubDispatcher{}
	body := readFixture(t, "push-org.json")
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service.New("usj", dispatcher),
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", signWebhookBody(body, "super-secret"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if dispatcher.called {
		t.Fatal("expected missing event header fixture not to dispatch")
	}
}

func TestGitHubHandlerReturnsUnauthorizedForFixtureWhenSignatureMissing(t *testing.T) {
	dispatcher := &stubDispatcher{}
	body := readFixture(t, "push-org.json")
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service.New("usj", dispatcher),
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "push")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	if dispatcher.called {
		t.Fatal("expected missing signature fixture not to dispatch")
	}
}

func TestGitHubHandlerReturnsUnauthorizedForFixtureWhenSignatureInvalid(t *testing.T) {
	dispatcher := &stubDispatcher{}
	body := readFixture(t, "push-org.json")
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service.New("usj", dispatcher),
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	if dispatcher.called {
		t.Fatal("expected invalid signature fixture not to dispatch")
	}
}

func TestGitHubHandlerReturnsInternalServerErrorForFixtureWhenServiceFails(t *testing.T) {
	dispatcher := &stubDispatcher{err: errors.New("discord send failed")}
	body := readFixture(t, "push-org.json")
	handler := newGitHubHandler(gitHubHandlerDependencies{
		secret:  "super-secret",
		service: service.New("usj", dispatcher),
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", signWebhookBody(body, "super-secret"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	if !dispatcher.called {
		t.Fatal("expected failing fixture to reach dispatcher before returning 500")
	}
}

func signWebhookBody(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	body, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatalf("expected fixture %s to load, got error: %v", name, err)
	}

	return body
}
