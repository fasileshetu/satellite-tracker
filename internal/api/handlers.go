package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"satellite-tracker/internal/db"
	"satellite-tracker/internal/models"
)

type Server struct {
	Store *db.Store
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) CreateComponent(w http.ResponseWriter, r *http.Request) {
	var in models.NewComponentInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if in.SatelliteID == "" || in.Name == "" || in.PartNumber == "" {
		writeError(w, http.StatusBadRequest, "satellite_id, name, and part_number are required")
		return
	}

	c, err := s.Store.CreateComponent(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create component")
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) ListComponents(w http.ResponseWriter, r *http.Request) {
	satelliteID := r.URL.Query().Get("satellite_id")
	comps, err := s.Store.ListComponents(r.Context(), satelliteID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list components")
		return
	}
	writeJSON(w, http.StatusOK, comps)
}

func (s *Server) GetComponent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	c, err := s.Store.GetComponent(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "component not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch component")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

type updateStatusInput struct {
	Status string `json:"status"`
}

func (s *Server) UpdateComponentStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var in updateStatusInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Status == "" {
		writeError(w, http.StatusBadRequest, "status is required")
		return
	}

	c, err := s.Store.UpdateStatus(r.Context(), id, in.Status)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "component not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update component")
		return
	}
	writeJSON(w, http.StatusOK, c)
}
