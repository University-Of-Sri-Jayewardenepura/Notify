package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/University-Of-Sri-Jayewardenepura/Notify/internal/github"
)

type gitHubHandlerDependencies struct {
	secret  string
	service GitHubEventService
}

func newGitHubHandler(deps gitHubHandlerDependencies) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		eventType := r.Header.Get("X-GitHub-Event")
		if eventType == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		signature := r.Header.Get("X-Hub-Signature-256")
		if signature == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !github.VerifySignature(payload, signature, deps.secret) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var jsonPayload map[string]any
		if err := json.Unmarshal(payload, &jsonPayload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if deps.service != nil {
			if err := deps.service.HandleGitHubEvent(eventType, jsonPayload); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	})
}
