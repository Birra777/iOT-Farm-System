package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/agristream/agristream/internal/models"
)

// AlertsRepo handles persistence of anomaly alerts.
type AlertsRepo struct {
	pool *pgxpool.Pool
}

// NewAlertsRepo constructs an AlertsRepo.
func NewAlertsRepo(pool *pgxpool.Pool) *AlertsRepo {
	return &AlertsRepo{pool: pool}
}

// Insert persists a new alert and returns its generated ID.
func (r *AlertsRepo) Insert(ctx context.Context, a models.Alert) (int64, error) {
	const q = `
		INSERT INTO alerts (field_id, sensor_type, metric, value, threshold, severity, message, status, triggered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	var id int64
	err := r.pool.QueryRow(ctx, q,
		a.FieldID,
		string(a.SensorType),
		a.Metric,
		a.Value,
		a.Threshold,
		string(a.Severity),
		a.Message,
		string(a.Status),
		a.TriggeredAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert alert: %w", err)
	}
	return id, nil
}

// List returns all alerts, optionally filtered by status.
func (r *AlertsRepo) List(ctx context.Context, status string) ([]models.Alert, error) {
	q := `
		SELECT id, field_id, sensor_type, metric, value, threshold, severity, message, status, triggered_at, resolved_at
		FROM alerts
	`
	args := []any{}
	if status != "" {
		q += " WHERE status = $1"
		args = append(args, status)
	}
	q += " ORDER BY triggered_at DESC LIMIT 200"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var a models.Alert
		var sensorType, severity, status string
		if err := rows.Scan(&a.ID, &a.FieldID, &sensorType, &a.Metric, &a.Value, &a.Threshold, &severity, &a.Message, &status, &a.TriggeredAt, &a.ResolvedAt); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		a.SensorType = models.SensorType(sensorType)
		a.Severity = models.Severity(severity)
		a.Status = models.AlertStatus(status)
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// Resolve marks an alert as resolved.
func (r *AlertsRepo) Resolve(ctx context.Context, id int64) error {
	now := time.Now()
	const q = `UPDATE alerts SET status = 'resolved', resolved_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, now, id)
	if err != nil {
		return fmt.Errorf("resolve alert: %w", err)
	}
	return nil
}
