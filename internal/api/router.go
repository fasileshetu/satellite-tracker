package api

import "net/http"

// NewRouter wires up all routes using Go 1.22's method-aware ServeMux
// patterns, so no third-party router dependency is needed yet.
func NewRouter(s *Server) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.Health)

	mux.HandleFunc("POST /components", s.CreateComponent)
	mux.HandleFunc("GET /components", s.ListComponents)
	mux.HandleFunc("GET /components/{id}", s.GetComponent)
	mux.HandleFunc("PATCH /components/{id}/status", s.UpdateComponentStatus)

	return mux
}
