package rules

import (
	"fmt"
	"math"
	"time"

	"github.com/agristream/agristream/internal/models"
)

// ThresholdRule fires when a metric crosses a low or high threshold.
// Either low or high bounds can be omitted by using math.NaN().
//
//   - LowWarning / LowCritical  — fires when value < threshold
//   - HighWarning / HighCritical — fires when value > threshold
//
// Critical takes precedence over warning when both apply.
type ThresholdRule struct {
	Metric      string
	LowWarning  float64
	LowCritical float64
	HighWarning float64
	HighCritical float64
	Message     string
}

// nan is a sentinel meaning "no threshold defined for this bound".
var nan = math.NaN()

// Evaluate checks the reading against all configured thresholds.
// Returns the most severe triggered alert, or (zero, false) if none.
func (r ThresholdRule) Evaluate(reading models.SensorReading) (models.Alert, bool) {
	if reading.Metric != r.Metric {
		return models.Alert{}, false
	}

	v := reading.Value
	severity, threshold, triggered := r.check(v)
	if !triggered {
		return models.Alert{}, false
	}

	return models.Alert{
		FieldID:     reading.FieldID,
		SensorType:  reading.SensorType,
		Metric:      reading.Metric,
		Value:       v,
		Threshold:   threshold,
		Severity:    severity,
		Message:     fmt.Sprintf("%s (value=%.2f %s, threshold=%.2f)", r.Message, v, reading.Unit, threshold),
		Status:      models.AlertStatusActive,
		TriggeredAt: time.Now().UTC(),
	}, true
}

// check returns the most severe triggered severity, its threshold, and whether any triggered.
func (r ThresholdRule) check(v float64) (models.Severity, float64, bool) {
	// Critical low
	if !math.IsNaN(r.LowCritical) && v < r.LowCritical {
		return models.SeverityCritical, r.LowCritical, true
	}
	// Critical high
	if !math.IsNaN(r.HighCritical) && v > r.HighCritical {
		return models.SeverityCritical, r.HighCritical, true
	}
	// Warning low
	if !math.IsNaN(r.LowWarning) && v < r.LowWarning {
		return models.SeverityWarning, r.LowWarning, true
	}
	// Warning high
	if !math.IsNaN(r.HighWarning) && v > r.HighWarning {
		return models.SeverityWarning, r.HighWarning, true
	}
	return "", 0, false
}

// DefaultRules returns the full set of agricultural threshold rules.
func DefaultRules() []Rule {
	return []Rule{
		// Soil moisture — drought risk
		ThresholdRule{
			Metric:      models.MetricSoilMoisture,
			LowWarning:  30,
			LowCritical: 20,
			HighWarning: nan,
			HighCritical: nan,
			Message:     "Drought risk",
		},
		// Soil pH — optimal range 5.5–7.5
		ThresholdRule{
			Metric:      models.MetricSoilPH,
			LowWarning:  5.5,
			LowCritical: nan,
			HighWarning: 7.5,
			HighCritical: nan,
			Message:     "pH out of optimal range",
		},
		// Soil nitrogen — deficiency
		ThresholdRule{
			Metric:      models.MetricSoilNitrogen,
			LowWarning:  100,
			LowCritical: 80,
			HighWarning: nan,
			HighCritical: nan,
			Message:     "Nitrogen deficiency",
		},
		// Temperature — heat stress
		ThresholdRule{
			Metric:      models.MetricWeatherTemperature,
			LowWarning:  nan,
			LowCritical: nan,
			HighWarning: 35,
			HighCritical: 38,
			Message:     "Heat stress risk",
		},
		// Humidity — fungal disease conditions
		ThresholdRule{
			Metric:      models.MetricWeatherHumidity,
			LowWarning:  nan,
			LowCritical: nan,
			HighWarning: 85,
			HighCritical: 90,
			Message:     "Fungal disease conditions",
		},
	}
}
