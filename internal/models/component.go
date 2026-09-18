package models

import "time"

// Component represents a physical part being tracked through manufacturing
// and test, e.g. an avionics board or solar panel, tied to a satellite build.
type Component struct {
	ID          int64     `json:"id"`
	SatelliteID string    `json:"satellite_id"`
	Name        string    `json:"name"`
	PartNumber  string    `json:"part_number"`
	Status      string    `json:"status"` // e.g. "received", "in_test", "pass", "fail", "flight_ready"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewComponentInput is the shape expected on POST /components.
type NewComponentInput struct {
	SatelliteID string `json:"satellite_id"`
	Name        string `json:"name"`
	PartNumber  string `json:"part_number"`
}
