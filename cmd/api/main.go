package main

import (
	"log"
	"net/http"
	"os"

	"satellite-tracker/internal/api"
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
	router := api.NewRouter(server)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
