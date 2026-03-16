package models

import "time"

// Severity levels for anomaly alerts.
type Severity string

const (
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// AlertStatus tracks the lifecycle of an alert.
type AlertStatus string

const (
	AlertStatusActive   AlertStatus = "active"
	AlertStatusResolved AlertStatus = "resolved"
)

// Alert represents a threshold breach detected by the anomaly detector.
type Alert struct {
	ID          int64       `json:"id"`
	FieldID     string      `json:"field_id"`
	SensorType  SensorType  `json:"sensor_type"`
	Metric      string      `json:"metric"`
	Value       float64     `json:"value"`
	Threshold   float64     `json:"threshold"`
	Severity    Severity    `json:"severity"`
	Message     string      `json:"message"`
	Status      AlertStatus `json:"status"`
	TriggeredAt time.Time   `json:"triggered_at"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty"`
}
