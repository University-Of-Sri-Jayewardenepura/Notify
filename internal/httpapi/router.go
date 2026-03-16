package httpapi

import "net/http"

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/health", healthHandler)

	return mux
}
