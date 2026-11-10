// Package httpapi exposes the calculator REST API.
package httpapi

import (
	"encoding/json"
	"net/http"
)

// NewRouter registers the API routes and returns the HTTP handler.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("POST /api/operations/{op}", handleOperation)
	mux.HandleFunc("GET /swagger", handleSwaggerUI)
	mux.HandleFunc("GET /swagger/openapi.yaml", handleOpenAPISpec)
	return mux
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
