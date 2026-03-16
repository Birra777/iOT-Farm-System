package models

import "time"

// GlobalThreshold stores operator-configurable alert thresholds for a metric.
// Nil means "no threshold defined for this bound".
type GlobalThreshold struct {
	Metric       string     `json:"metric"`
	WarningLow   *float64   `json:"warning_low"`
	CriticalLow  *float64   `json:"critical_low"`
	WarningHigh  *float64   `json:"warning_high"`
	CriticalHigh *float64   `json:"critical_high"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
