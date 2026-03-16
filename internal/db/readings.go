package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/agristream/agristream/internal/models"
)

// ReadingsRepo handles persistence of sensor readings.
type ReadingsRepo struct {
	pool *pgxpool.Pool
}

// NewReadingsRepo constructs a ReadingsRepo.
func NewReadingsRepo(pool *pgxpool.Pool) *ReadingsRepo {
	return &ReadingsRepo{pool: pool}
}

// Insert persists a single SensorReading to the database.
// The query is ON CONFLICT DO NOTHING for idempotency in at-least-once delivery.
func (r *ReadingsRepo) Insert(ctx context.Context, reading models.SensorReading) error {
	const q = `
		INSERT INTO sensor_readings (field_id, sensor_id, sensor_type, metric, value, unit, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, q,
		reading.FieldID,
		reading.SensorID,
		string(reading.SensorType),
		reading.Metric,
		reading.Value,
		reading.Unit,
		reading.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("insert reading: %w", err)
	}
	return nil
}

// ListByFieldAndMetric returns time-series readings for a field/metric in the given window.
func (r *ReadingsRepo) ListByFieldAndMetric(ctx context.Context, fieldID, metric string, from, to time.Time) ([]models.SensorReading, error) {
	const q = `
		SELECT sensor_id, field_id, sensor_type, metric, value, unit, recorded_at
		FROM sensor_readings
		WHERE field_id = $1 AND metric = $2 AND recorded_at BETWEEN $3 AND $4
		ORDER BY recorded_at DESC
		LIMIT 1000
	`
	rows, err := r.pool.Query(ctx, q, fieldID, metric, from, to)
	if err != nil {
		return nil, fmt.Errorf("query readings: %w", err)
	}
	defer rows.Close()

	var readings []models.SensorReading
	for rows.Next() {
		var rd models.SensorReading
		var sensorType string
		if err := rows.Scan(&rd.SensorID, &rd.FieldID, &sensorType, &rd.Metric, &rd.Value, &rd.Unit, &rd.Timestamp); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		rd.SensorType = models.SensorType(sensorType)
		readings = append(readings, rd)
	}
	return readings, rows.Err()
}

// LatestPerMetric returns the most recent reading per metric for a field.
func (r *ReadingsRepo) LatestPerMetric(ctx context.Context, fieldID string) ([]models.SensorReading, error) {
	const q = `
		SELECT DISTINCT ON (metric)
			sensor_id, field_id, sensor_type, metric, value, unit, recorded_at
		FROM sensor_readings
		WHERE field_id = $1
		ORDER BY metric, recorded_at DESC
	`
	rows, err := r.pool.Query(ctx, q, fieldID)
	if err != nil {
		return nil, fmt.Errorf("query latest readings: %w", err)
	}
	defer rows.Close()

	var readings []models.SensorReading
	for rows.Next() {
		var rd models.SensorReading
		var sensorType string
		if err := rows.Scan(&rd.SensorID, &rd.FieldID, &sensorType, &rd.Metric, &rd.Value, &rd.Unit, &rd.Timestamp); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		rd.SensorType = models.SensorType(sensorType)
		readings = append(readings, rd)
	}
	return readings, rows.Err()
}
