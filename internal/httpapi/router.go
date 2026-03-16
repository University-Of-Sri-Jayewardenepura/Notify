package httpapi

import "net/http"

type GitHubEventService interface {
	HandleGitHubEvent(eventType string, payload map[string]any) error
}

type RouterDependencies struct {
	GitHubWebhookSecret string
	GitHubService       GitHubEventService
}

func NewRouter(deps RouterDependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/health", healthHandler)
	mux.Handle("/webhook/github", newGitHubHandler(gitHubHandlerDependencies{
		secret:  deps.GitHubWebhookSecret,
		service: deps.GitHubService,
	}))

	return mux
}
