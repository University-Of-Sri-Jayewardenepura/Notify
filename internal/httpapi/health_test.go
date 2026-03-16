package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthRouteReturnsServiceStatus(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/webhook/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON response, got error: %v", err)
	}

	if payload["status"] != "UP" {
		t.Fatalf("expected status to be UP, got %#v", payload["status"])
	}

	if payload["serviceName"] != "GitHub Notification Service" {
		t.Fatalf("expected serviceName to be GitHub Notification Service, got %#v", payload["serviceName"])
	}

	if payload["version"] != "0.1.0" {
		t.Fatalf("expected version to be 0.1.0, got %#v", payload["version"])
	}

	if payload["timestamp"] == nil || payload["timestamp"] == "" {
		t.Fatalf("expected timestamp to be present, got %#v", payload["timestamp"])
	}

	timestamp, ok := payload["timestamp"].(string)
	if !ok {
		t.Fatalf("expected timestamp to be a string, got %#v", payload["timestamp"])
	}

	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		t.Fatalf("expected RFC3339 timestamp, got %q: %v", timestamp, err)
	}
}

func TestHealthRouteRejectsNonGetRequests(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/webhook/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}
