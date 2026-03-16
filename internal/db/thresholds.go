package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/agristream/agristream/internal/models"
)

// ThresholdsRepo handles persistence of global alert thresholds.
type ThresholdsRepo struct {
	pool *pgxpool.Pool
}

// NewThresholdsRepo constructs a ThresholdsRepo.
func NewThresholdsRepo(pool *pgxpool.Pool) *ThresholdsRepo {
	return &ThresholdsRepo{pool: pool}
}

// List returns all global thresholds ordered by metric name.
func (r *ThresholdsRepo) List(ctx context.Context) ([]models.GlobalThreshold, error) {
	const q = `
		SELECT metric, warning_low, critical_low, warning_high, critical_high, updated_at
		FROM global_thresholds
		ORDER BY metric
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query thresholds: %w", err)
	}
	defer rows.Close()

	var out []models.GlobalThreshold
	for rows.Next() {
		var t models.GlobalThreshold
		if err := rows.Scan(&t.Metric, &t.WarningLow, &t.CriticalLow, &t.WarningHigh, &t.CriticalHigh, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan threshold: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Upsert creates or replaces the threshold row for a metric.
func (r *ThresholdsRepo) Upsert(ctx context.Context, t models.GlobalThreshold) (models.GlobalThreshold, error) {
	const q = `
		INSERT INTO global_thresholds (metric, warning_low, critical_low, warning_high, critical_high, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (metric) DO UPDATE SET
			warning_low   = EXCLUDED.warning_low,
			critical_low  = EXCLUDED.critical_low,
			warning_high  = EXCLUDED.warning_high,
			critical_high = EXCLUDED.critical_high,
			updated_at    = NOW()
		RETURNING metric, warning_low, critical_low, warning_high, critical_high, updated_at
	`
	var out models.GlobalThreshold
	err := r.pool.QueryRow(ctx, q, t.Metric, t.WarningLow, t.CriticalLow, t.WarningHigh, t.CriticalHigh).
		Scan(&out.Metric, &out.WarningLow, &out.CriticalLow, &out.WarningHigh, &out.CriticalHigh, &out.UpdatedAt)
	if err != nil {
		return models.GlobalThreshold{}, fmt.Errorf("upsert threshold: %w", err)
	}
	return out, nil
}
