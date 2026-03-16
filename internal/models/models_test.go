package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/agristream/agristream/internal/models"
)

// TestSensorReading_RoundTrip verifies JSON marshal/unmarshal preserves all fields.
func TestSensorReading_RoundTrip(t *testing.T) {
	original := models.SensorReading{
		SensorID:   "DR-SOIL-01-MOIST",
		FieldID:    "dab293c2-8931-40cd-803c-b0e41ad34976",
		SensorType: models.SensorTypeSoil,
		Metric:     models.MetricSoilMoisture,
		Value:      18.75,
		Unit:       "%",
		Timestamp:  time.Date(2026, 3, 13, 9, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got models.SensorReading
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.SensorID != original.SensorID {
		t.Errorf("SensorID: got %q, want %q", got.SensorID, original.SensorID)
	}
	if got.FieldID != original.FieldID {
		t.Errorf("FieldID: got %q, want %q", got.FieldID, original.FieldID)
	}
	if got.SensorType != original.SensorType {
		t.Errorf("SensorType: got %q, want %q", got.SensorType, original.SensorType)
	}
	if got.Metric != original.Metric {
		t.Errorf("Metric: got %q, want %q", got.Metric, original.Metric)
	}
	if got.Value != original.Value {
		t.Errorf("Value: got %f, want %f", got.Value, original.Value)
	}
	if got.Unit != original.Unit {
		t.Errorf("Unit: got %q, want %q", got.Unit, original.Unit)
	}
	if !got.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp: got %v, want %v", got.Timestamp, original.Timestamp)
	}
}

// TestAggregatedReading_RoundTrip verifies the aggregated message shape survives JSON.
func TestAggregatedReading_RoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	original := models.AggregatedReading{
		FieldID:     "37d24fb7-8149-44b6-8bfd-470298729a7d",
		SensorType:  models.SensorTypeWeather,
		Metric:      models.MetricWeatherTemperature,
		Avg:         31.5,
		Min:         28.0,
		Max:         35.2,
		Count:       12,
		WindowStart: now.Add(-time.Minute),
		WindowEnd:   now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got models.AggregatedReading
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Avg != original.Avg {
		t.Errorf("Avg: got %f, want %f", got.Avg, original.Avg)
	}
	if got.Count != original.Count {
		t.Errorf("Count: got %d, want %d", got.Count, original.Count)
	}
	if got.Metric != original.Metric {
		t.Errorf("Metric: got %q, want %q", got.Metric, original.Metric)
	}
}

// TestAlert_RoundTrip verifies Alert JSON round-trip including optional resolved_at.
func TestAlert_RoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	t.Run("active alert (no resolved_at)", func(t *testing.T) {
		a := models.Alert{
			ID:          7,
			FieldID:     "dab293c2-8931-40cd-803c-b0e41ad34976",
			SensorType:  models.SensorTypeSoil,
			Metric:      models.MetricSoilMoisture,
			Value:       15.3,
			Threshold:   20.0,
			Severity:    models.SeverityCritical,
			Message:     "Drought risk (value=15.30 %, threshold=20.00)",
			Status:      models.AlertStatusActive,
			TriggeredAt: now,
			ResolvedAt:  nil,
		}

		data, _ := json.Marshal(a)
		var got models.Alert
		json.Unmarshal(data, &got)

		if got.ResolvedAt != nil {
			t.Error("ResolvedAt should be nil for active alert")
		}
		if got.Severity != models.SeverityCritical {
			t.Errorf("Severity: got %q, want critical", got.Severity)
		}
		if got.Status != models.AlertStatusActive {
			t.Errorf("Status: got %q, want active", got.Status)
		}
	})

	t.Run("resolved alert (with resolved_at)", func(t *testing.T) {
		resolved := now.Add(10 * time.Minute)
		a := models.Alert{
			Status:      models.AlertStatusResolved,
			TriggeredAt: now,
			ResolvedAt:  &resolved,
		}

		data, _ := json.Marshal(a)
		var got models.Alert
		json.Unmarshal(data, &got)

		if got.ResolvedAt == nil {
			t.Fatal("ResolvedAt should not be nil for resolved alert")
		}
		if !got.ResolvedAt.Equal(resolved) {
			t.Errorf("ResolvedAt: got %v, want %v", got.ResolvedAt, resolved)
		}
	})
}
