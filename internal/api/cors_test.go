package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithCORS_SetsHeadersAndCallsNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/components", nil)
	rec := httptest.NewRecorder()

	withCORS(next, "https://dashboard.example.com").ServeHTTP(rec, req)

	if !called {
		t.Error("expected the wrapped handler to run for a non-OPTIONS request")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://dashboard.example.com" {
		t.Errorf("Access-Control-Allow-Origin = %q, want the configured origin", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Access-Control-Allow-Methods header was not set")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("Access-Control-Allow-Headers header was not set")
	}
}

func TestWithCORS_PreflightRequestShortCircuits(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodOptions, "/components", nil)
	rec := httptest.NewRecorder()

	withCORS(next, "*").ServeHTTP(rec, req)

	if called {
		t.Error("an OPTIONS preflight request should never reach the wrapped handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d for a preflight request", rec.Code, http.StatusNoContent)
	}
}
