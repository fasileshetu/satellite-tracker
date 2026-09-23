package api

import (
	"net/http"

	"satellite-tracker/internal/auth"
)

// NewRouter wires up all routes using Go 1.22's method-aware ServeMux
// patterns, so no third-party router dependency is needed yet.
//
// authVerifier may be nil (local dev without Cognito configured, e.g.
// docker-compose) -- in that case every route is left unprotected,
// same as before OAuth existed. When it's set, the write endpoints
// require a valid Cognito access token; reads and /health stay open.
func NewRouter(s *Server, authVerifier *auth.Verifier) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.Health)
	mux.HandleFunc("GET /components", s.ListComponents)
	mux.HandleFunc("GET /components/{id}", s.GetComponent)

	createComponent := http.Handler(http.HandlerFunc(s.CreateComponent))
	updateStatus := http.Handler(http.HandlerFunc(s.UpdateComponentStatus))
	if authVerifier != nil {
		createComponent = authVerifier.Middleware(createComponent)
		updateStatus = authVerifier.Middleware(updateStatus)
	}
	mux.Handle("POST /components", createComponent)
	mux.Handle("PATCH /components/{id}/status", updateStatus)

	return mux
}
