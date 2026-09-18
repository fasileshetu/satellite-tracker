package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"satellite-tracker/internal/models"
)

type Store struct {
	conn *sql.DB
}

// New opens a connection pool to Postgres using a standard DSN, e.g.
// "postgres://user:pass@localhost:5432/satellite_tracker?sslmode=disable"
func New(dsn string) (*Store, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("pinging db: %w", err)
	}
	return &Store{conn: conn}, nil
}

func (s *Store) Close() error {
	return s.conn.Close()
}

func (s *Store) CreateComponent(ctx context.Context, in models.NewComponentInput) (models.Component, error) {
	const q = `
		INSERT INTO components (satellite_id, name, part_number, status)
		VALUES ($1, $2, $3, 'received')
		RETURNING id, satellite_id, name, part_number, status, created_at, updated_at`

	var c models.Component
	err := s.conn.QueryRowContext(ctx, q, in.SatelliteID, in.Name, in.PartNumber).Scan(
		&c.ID, &c.SatelliteID, &c.Name, &c.PartNumber, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (s *Store) ListComponents(ctx context.Context, satelliteID string) ([]models.Component, error) {
	q := `SELECT id, satellite_id, name, part_number, status, created_at, updated_at FROM components`
	args := []any{}
	if satelliteID != "" {
		q += ` WHERE satellite_id = $1`
		args = append(args, satelliteID)
	}
	q += ` ORDER BY created_at DESC`

	rows, err := s.conn.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Component
	for rows.Next() {
		var c models.Component
		if err := rows.Scan(&c.ID, &c.SatelliteID, &c.Name, &c.PartNumber, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetComponent(ctx context.Context, id int64) (models.Component, error) {
	const q = `SELECT id, satellite_id, name, part_number, status, created_at, updated_at FROM components WHERE id = $1`
	var c models.Component
	err := s.conn.QueryRowContext(ctx, q, id).Scan(
		&c.ID, &c.SatelliteID, &c.Name, &c.PartNumber, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (s *Store) UpdateStatus(ctx context.Context, id int64, status string) (models.Component, error) {
	const q = `
		UPDATE components SET status = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, satellite_id, name, part_number, status, created_at, updated_at`

	var c models.Component
	err := s.conn.QueryRowContext(ctx, q, id, status).Scan(
		&c.ID, &c.SatelliteID, &c.Name, &c.PartNumber, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}
