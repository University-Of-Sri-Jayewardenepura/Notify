package service

import "testing"

type stubDispatcher struct {
	called    bool
	eventType string
}

func (s *stubDispatcher) DispatchGitHubEvent(eventType string, payload map[string]any) error {
	s.called = true
	s.eventType = eventType
	return nil
}

func TestHandleGitHubEventDispatchesPingWithoutOrgCheck(t *testing.T) {
	dispatcher := &stubDispatcher{}
	svc := New("usj", dispatcher)

	err := svc.HandleGitHubEvent("ping", map[string]any{
		"zen": "keep it logically awesome",
	})
	if err != nil {
		t.Fatalf("expected ping event to succeed, got error: %v", err)
	}

	if !dispatcher.called {
		t.Fatal("expected ping event to dispatch without org filtering")
	}
}

func TestHandleGitHubEventIgnoresEventsFromDifferentOrganizations(t *testing.T) {
	dispatcher := &stubDispatcher{}
	svc := New("usj", dispatcher)

	err := svc.HandleGitHubEvent("push", map[string]any{
		"repository": map[string]any{
			"owner": map[string]any{
				"login": "someone-else",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected non-matching org event to be ignored without error, got: %v", err)
	}

	if dispatcher.called {
		t.Fatal("expected non-matching org event not to dispatch")
	}
}

func TestHandleGitHubEventDispatchesMatchingOrganizationEvents(t *testing.T) {
	dispatcher := &stubDispatcher{}
	svc := New("usj", dispatcher)

	err := svc.HandleGitHubEvent("push", map[string]any{
		"repository": map[string]any{
			"owner": map[string]any{
				"login": "usj",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected matching org event to dispatch, got error: %v", err)
	}

	if !dispatcher.called {
		t.Fatal("expected matching org event to dispatch")
	}
}

func TestLoggingDispatcherHandlesEventsWithoutError(t *testing.T) {
	dispatcher := NewLoggingDispatcher()

	err := dispatcher.DispatchGitHubEvent("ping", map[string]any{
		"zen": "keep it logically awesome",
	})
	if err != nil {
		t.Fatalf("expected logging dispatcher to succeed, got error: %v", err)
	}
}
