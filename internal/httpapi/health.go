package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":      "UP",
		"serviceName": "GitHub Notification Service",
		"version":     "0.1.0",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}
