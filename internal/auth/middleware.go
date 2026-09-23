// Package auth validates OAuth 2.0 access tokens issued by an OIDC
// provider (AWS Cognito, provisioned in terraform/cognito.tf) and
// protects HTTP handlers with them.
//
// This is the resource-server half of OIDC: it doesn't log anyone in or
// issue tokens itself -- it trusts tokens that were already issued by
// Cognito and just verifies "is this token real, current, and issued by
// the identity provider I trust." The actual login (redirecting a user
// to Cognito's Hosted UI, exchanging a code for a token) happens
// elsewhere -- a frontend, or the AWS CLI for testing.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// contextKey avoids collisions with any other package's context values.
type contextKey string

const subjectContextKey contextKey = "auth.subject"

// Config holds what's needed to validate tokens from one Cognito user
// pool. All three values come straight out of `terraform output` --
// see cognito_issuer_url and cognito_app_client_id.
type Config struct {
	// IssuerURL is https://cognito-idp.<region>.amazonaws.com/<user-pool-id>.
	// Every valid token's "iss" claim must match this exactly.
	IssuerURL string
	// ClientID is the Cognito app client ID. Checked against the token's
	// "client_id" claim so a token issued for some other app client
	// (even one in the same user pool) is rejected.
	ClientID string
}

// Verifier validates Cognito-issued JWTs. Build one with NewVerifier at
// startup and reuse it -- it caches the provider's public signing keys
// (JWKS) internally and refreshes them in the background, so it doesn't
// fetch them on every request.
type Verifier struct {
	cfg  Config
	jwks keyfunc.Keyfunc
}

// NewVerifier fetches and starts caching the identity provider's JWKS
// (its public signing keys). Cognito rotates these occasionally; keyfunc
// handles re-fetching automatically so a middleware built once at
// startup keeps working indefinitely.
func NewVerifier(ctx context.Context, cfg Config) (*Verifier, error) {
	jwksURL := cfg.IssuerURL + "/.well-known/jwks.json"

	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("fetch JWKS from %s: %w", jwksURL, err)
	}

	return &Verifier{cfg: cfg, jwks: jwks}, nil
}

// cognitoClaims is the subset of a Cognito access token's claims this
// service cares about. Cognito's ACCESS token (not the ID token) is what
// a caller should send here -- it carries "token_use": "access" and a
// "client_id" claim rather than the ID token's "aud".
type cognitoClaims struct {
	jwt.RegisteredClaims
	TokenUse string `json:"token_use"`
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
}

// Middleware rejects any request without a valid Cognito access token in
// its Authorization header, and makes the token's subject (the user's
// Cognito "sub" -- a stable UUID) available to handlers via SubjectFromContext.
func (v *Verifier) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := bearerToken(r)
		if !ok {
			writeUnauthorized(w, "missing or malformed Authorization header")
			return
		}

		claims := &cognitoClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, v.jwks.Keyfunc,
			jwt.WithValidMethods([]string{"RS256"}),
			jwt.WithIssuer(v.cfg.IssuerURL),
		)
		if err != nil || !token.Valid {
			writeUnauthorized(w, "invalid or expired token")
			return
		}

		// jwt.WithIssuer above already checked "iss". Cognito access
		// tokens don't carry a standard "aud" claim, so client_id is
		// checked separately -- otherwise any access token from *any*
		// app client in the pool would be accepted here.
		if claims.TokenUse != "access" {
			writeUnauthorized(w, "token is not an access token")
			return
		}
		if claims.ClientID != v.cfg.ClientID {
			writeUnauthorized(w, "token was not issued for this client")
			return
		}

		ctx := context.WithValue(r.Context(), subjectContextKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SubjectFromContext returns the authenticated user's Cognito subject
// (a stable UUID, not their email/username) from a request that's
// already passed through Middleware.
func SubjectFromContext(ctx context.Context) (string, bool) {
	sub, ok := ctx.Value(subjectContextKey).(string)
	return sub, ok
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	return strings.TrimPrefix(header, prefix), true
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="satellite-tracker"`)
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
