package api

import "net/http"

// withCORS lets the Next.js dashboard (a separate origin -- localhost:3000
// in dev, or wherever it's deployed) call this API from the browser. The
// allowed origin is configurable via ALLOWED_ORIGIN so it doesn't need to
// be "*" once the dashboard has a real deployed URL; "*" is a reasonable
// default for local dev only.
func withCORS(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
