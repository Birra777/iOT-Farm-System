package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/agristream/agristream/internal/models"
)

// FieldsRepo handles queries against the fields table.
type FieldsRepo struct {
	pool *pgxpool.Pool
}

// NewFieldsRepo constructs a FieldsRepo.
func NewFieldsRepo(pool *pgxpool.Pool) *FieldsRepo {
	return &FieldsRepo{pool: pool}
}

// ZoneCodeToUUID returns a map of zone_code → UUID for all fields.
// Used by the simulator to resolve zone codes to real field IDs at startup.
func (r *FieldsRepo) ZoneCodeToUUID(ctx context.Context) (map[string]string, error) {
	const q = `SELECT id, zone_code FROM fields`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query zone codes: %w", err)
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var id, zoneCode string
		if err := rows.Scan(&id, &zoneCode); err != nil {
			return nil, fmt.Errorf("scan zone code: %w", err)
		}
		m[zoneCode] = id
	}
	return m, rows.Err()
}

// List returns all farm fields ordered by name.
func (r *FieldsRepo) List(ctx context.Context) ([]models.Field, error) {
	const q = `
		SELECT id, name, crop_type, hectares, zone_code, created_at
		FROM fields
		ORDER BY name
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query fields: %w", err)
	}
	defer rows.Close()

	var fields []models.Field
	for rows.Next() {
		var f models.Field
		if err := rows.Scan(&f.ID, &f.Name, &f.CropType, &f.Hectares, &f.ZoneCode, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan field: %w", err)
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}

// Get returns a single field by ID.
func (r *FieldsRepo) Get(ctx context.Context, id string) (models.Field, error) {
	const q = `
		SELECT id, name, crop_type, hectares, zone_code, created_at
		FROM fields WHERE id = $1
	`
	var f models.Field
	err := r.pool.QueryRow(ctx, q, id).Scan(&f.ID, &f.Name, &f.CropType, &f.Hectares, &f.ZoneCode, &f.CreatedAt)
	if err != nil {
		return models.Field{}, fmt.Errorf("get field: %w", err)
	}
	return f, nil
}

// Stats returns aggregate pipeline statistics from the database.
func (r *FieldsRepo) Stats(ctx context.Context) (models.PipelineStats, error) {
	const q = `
		SELECT
			COUNT(*)                                                          AS total_readings,
			COUNT(*) FILTER (WHERE recorded_at > NOW() - INTERVAL '1 minute') AS readings_last_minute,
			COUNT(*) FILTER (WHERE recorded_at > NOW() - INTERVAL '1 hour')   AS readings_last_hour
		FROM sensor_readings
	`
	var s models.PipelineStats
	if err := r.pool.QueryRow(ctx, q).Scan(
		&s.TotalReadingsArchived,
		&s.ReadingsLastMinute,
		&s.ReadingsLastHour,
	); err != nil {
		return models.PipelineStats{}, fmt.Errorf("query stats: %w", err)
	}

	const aq = `SELECT COUNT(*) FROM alerts WHERE status = 'active'`
	if err := r.pool.QueryRow(ctx, aq).Scan(&s.ActiveAlerts); err != nil {
		return models.PipelineStats{}, fmt.Errorf("query active alerts: %w", err)
	}

	return s, nil
}
