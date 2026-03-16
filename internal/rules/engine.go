package rules

import (
	"github.com/agristream/agristream/internal/models"
)

// Rule evaluates a sensor reading and optionally produces an alert.
type Rule interface {
	Evaluate(reading models.SensorReading) (models.Alert, bool)
}

// ApplyRules runs every rule against the reading and returns all triggered alerts.
func ApplyRules(reading models.SensorReading, rules []Rule) []models.Alert {
	var alerts []models.Alert
	for _, r := range rules {
		if alert, triggered := r.Evaluate(reading); triggered {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}
