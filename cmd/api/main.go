package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"satellite-tracker/internal/api"
	"satellite-tracker/internal/auth"
	"satellite-tracker/internal/db"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/satellite_tracker?sslmode=disable"
	}

	store, err := db.New(dsn)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer store.Close()

	server := &api.Server{Store: store}

	// OIDC_ISSUER_URL / OIDC_CLIENT_ID come from `terraform output
	// cognito_issuer_url` / `cognito_app_client_id`. Left unset, the API
	// runs with write endpoints unprotected -- e.g. local dev via
	// docker-compose, where standing up Cognito isn't worth the trouble.
	var verifier *auth.Verifier
	issuerURL := os.Getenv("OIDC_ISSUER_URL")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	if issuerURL != "" && clientID != "" {
		verifier, err = auth.NewVerifier(context.Background(), auth.Config{
			IssuerURL: issuerURL,
			ClientID:  clientID,
		})
		if err != nil {
			log.Fatalf("failed to set up OIDC verifier: %v", err)
		}
		log.Println("OAuth/OIDC enabled -- write endpoints require a valid Cognito access token")
	} else {
		log.Println("OIDC_ISSUER_URL/OIDC_CLIENT_ID not set -- running with no auth (local dev only)")
	}

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*" // fine for a portfolio project; a real deployment would pin this to the dashboard's real origin
	}
	router := api.NewRouter(server, verifier, allowedOrigin)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
