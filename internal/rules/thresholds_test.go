package rules_test

import (
	"testing"
	"time"

	"github.com/agristream/agristream/internal/models"
	"github.com/agristream/agristream/internal/rules"
)

func reading(metric string, value float64, sensorType models.SensorType) models.SensorReading {
	return models.SensorReading{
		SensorID:   "test-sensor",
		FieldID:    "DR",
		SensorType: sensorType,
		Metric:     metric,
		Value:      value,
		Unit:       "",
		Timestamp:  time.Now(),
	}
}

func TestDefaultRules(t *testing.T) {
	ruleSet := rules.DefaultRules()

	tests := []struct {
		name             string
		reading          models.SensorReading
		wantTriggered    bool
		wantSeverity     models.Severity
	}{
		// --- soil.moisture ---
		{
			name:          "moisture normal — no alert",
			reading:       reading(models.MetricSoilMoisture, 55, models.SensorTypeSoil),
			wantTriggered: false,
		},
		{
			name:          "moisture at warning boundary — no alert (exclusive)",
			reading:       reading(models.MetricSoilMoisture, 30, models.SensorTypeSoil),
			wantTriggered: false,
		},
		{
			name:          "moisture just below warning",
			reading:       reading(models.MetricSoilMoisture, 29.9, models.SensorTypeSoil),
			wantTriggered: true,
			wantSeverity:  models.SeverityWarning,
		},
		{
			name:          "moisture at critical boundary — warning fires (20 < 30 warning, not < 20 critical)",
			reading:       reading(models.MetricSoilMoisture, 20, models.SensorTypeSoil),
			wantTriggered: true,
			wantSeverity:  models.SeverityWarning,
		},
		{
			name:          "moisture just below critical",
			reading:       reading(models.MetricSoilMoisture, 19.9, models.SensorTypeSoil),
			wantTriggered: true,
			wantSeverity:  models.SeverityCritical,
		},

		// --- soil.ph ---
		{
			name:          "pH in optimal range — no alert",
			reading:       reading(models.MetricSoilPH, 6.5, models.SensorTypeSoil),
			wantTriggered: false,
		},
		{
			name:          "pH at low warning boundary — no alert (exclusive)",
			reading:       reading(models.MetricSoilPH, 5.5, models.SensorTypeSoil),
			wantTriggered: false,
		},
		{
			name:          "pH just below low warning",
			reading:       reading(models.MetricSoilPH, 5.49, models.SensorTypeSoil),
			wantTriggered: true,
			wantSeverity:  models.SeverityWarning,
		},
		{
			name:          "pH at high warning boundary — no alert (exclusive)",
			reading:       reading(models.MetricSoilPH, 7.5, models.SensorTypeSoil),
			wantTriggered: false,
		},
		{
			name:          "pH just above high warning",
			reading:       reading(models.MetricSoilPH, 7.51, models.SensorTypeSoil),
			wantTriggered: true,
			wantSeverity:  models.SeverityWarning,
		},

		// --- soil.nitrogen ---
		{
			name:          "nitrogen normal — no alert",
			reading:       reading(models.MetricSoilNitrogen, 150, models.SensorTypeSoil),
			wantTriggered: false,
		},
		{
			name:          "nitrogen just below warning",
			reading:       reading(models.MetricSoilNitrogen, 99, models.SensorTypeSoil),
			wantTriggered: true,
			wantSeverity:  models.SeverityWarning,
		},
		{
			name:          "nitrogen just below critical",
			reading:       reading(models.MetricSoilNitrogen, 79, models.SensorTypeSoil),
			wantTriggered: true,
			wantSeverity:  models.SeverityCritical,
		},

		// --- weather.temperature ---
		{
			name:          "temperature normal — no alert",
			reading:       reading(models.MetricWeatherTemperature, 28, models.SensorTypeWeather),
			wantTriggered: false,
		},
		{
			name:          "temperature just above warning",
			reading:       reading(models.MetricWeatherTemperature, 35.1, models.SensorTypeWeather),
			wantTriggered: true,
			wantSeverity:  models.SeverityWarning,
		},
		{
			name:          "temperature just above critical",
			reading:       reading(models.MetricWeatherTemperature, 38.1, models.SensorTypeWeather),
			wantTriggered: true,
			wantSeverity:  models.SeverityCritical,
		},

		// --- weather.humidity ---
		{
			name:          "humidity normal — no alert",
			reading:       reading(models.MetricWeatherHumidity, 60, models.SensorTypeWeather),
			wantTriggered: false,
		},
		{
			name:          "humidity just above warning",
			reading:       reading(models.MetricWeatherHumidity, 85.1, models.SensorTypeWeather),
			wantTriggered: true,
			wantSeverity:  models.SeverityWarning,
		},
		{
			name:          "humidity just above critical",
			reading:       reading(models.MetricWeatherHumidity, 90.1, models.SensorTypeWeather),
			wantTriggered: true,
			wantSeverity:  models.SeverityCritical,
		},

		// --- wrong metric — no alert ---
		{
			name:          "unknown metric — no alert",
			reading:       reading("weather.wind_speed", 80, models.SensorTypeWeather),
			wantTriggered: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			alerts := rules.ApplyRules(tc.reading, ruleSet)

			if tc.wantTriggered && len(alerts) == 0 {
				t.Fatalf("expected alert but none triggered (metric=%s value=%.2f)", tc.reading.Metric, tc.reading.Value)
			}
			if !tc.wantTriggered && len(alerts) > 0 {
				t.Fatalf("expected no alert but got %d (metric=%s value=%.2f severity=%s)", len(alerts), tc.reading.Metric, tc.reading.Value, alerts[0].Severity)
			}
			if tc.wantTriggered {
				got := alerts[0].Severity
				if got != tc.wantSeverity {
					t.Errorf("severity mismatch: got %q, want %q", got, tc.wantSeverity)
				}
				if alerts[0].FieldID != tc.reading.FieldID {
					t.Errorf("field_id mismatch: got %q, want %q", alerts[0].FieldID, tc.reading.FieldID)
				}
			}
		})
	}
}
