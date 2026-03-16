package models

import "time"

// Notification is a plain-English, user-facing message derived from an alert.
type Notification struct {
	ID        int64      `json:"id"`
	AlertID   *int64     `json:"alert_id,omitempty"`
	FieldID   string     `json:"field_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Severity  string     `json:"severity"`
	IsRead    bool       `json:"is_read"`
	CreatedAt time.Time  `json:"created_at"`
}
