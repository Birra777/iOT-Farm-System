package rules

import (
	"context"
	"math"

	"github.com/agristream/agristream/internal/db"
	"github.com/agristream/agristream/internal/models"
)

// LoadRules loads threshold rules from the DB, merging DB values on top of
// DefaultRules so every metric always has a rule even if absent from the DB.
func LoadRules(ctx context.Context, repo *db.ThresholdsRepo) ([]Rule, error) {
	rows, err := repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Build a lookup map keyed by metric.
	dbMap := make(map[string]models.GlobalThreshold, len(rows))
	for _, r := range rows {
		dbMap[r.Metric] = r
	}

	defaults := DefaultRules()
	out := make([]Rule, 0, len(defaults))
	for _, def := range defaults {
		tr := def.(ThresholdRule)
		if ov, ok := dbMap[tr.Metric]; ok {
			if ov.WarningLow != nil {
				tr.LowWarning = *ov.WarningLow
			} else {
				tr.LowWarning = math.NaN()
			}
			if ov.CriticalLow != nil {
				tr.LowCritical = *ov.CriticalLow
			} else {
				tr.LowCritical = math.NaN()
			}
			if ov.WarningHigh != nil {
				tr.HighWarning = *ov.WarningHigh
			} else {
				tr.HighWarning = math.NaN()
			}
			if ov.CriticalHigh != nil {
				tr.HighCritical = *ov.CriticalHigh
			} else {
				tr.HighCritical = math.NaN()
			}
		}
		out = append(out, tr)
	}
	return out, nil
}
