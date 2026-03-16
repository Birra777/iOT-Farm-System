package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/agristream/agristream/internal/models"
)

// NotificationsRepo handles persistence of user-facing notifications.
type NotificationsRepo struct {
	pool *pgxpool.Pool
}

// NewNotificationsRepo constructs a NotificationsRepo.
func NewNotificationsRepo(pool *pgxpool.Pool) *NotificationsRepo {
	return &NotificationsRepo{pool: pool}
}

// Insert persists a new notification and returns its generated ID.
func (r *NotificationsRepo) Insert(ctx context.Context, n models.Notification) (int64, error) {
	const q = `
		INSERT INTO notifications (alert_id, field_id, title, body, severity)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id int64
	err := r.pool.QueryRow(ctx, q, n.AlertID, n.FieldID, n.Title, n.Body, n.Severity).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert notification: %w", err)
	}
	return id, nil
}

// List returns the most recent notifications (newest first), up to limit.
// Pass onlyUnread=true to return only unread notifications.
func (r *NotificationsRepo) List(ctx context.Context, onlyUnread bool, limit int) ([]models.Notification, error) {
	q := `
		SELECT id, alert_id, field_id, title, body, severity, is_read, created_at
		FROM notifications
	`
	args := []any{}
	if onlyUnread {
		q += " WHERE is_read = false"
	}
	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d", limit)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query notifications: %w", err)
	}
	defer rows.Close()

	var out []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.AlertID, &n.FieldID, &n.Title, &n.Body, &n.Severity, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// MarkRead marks a single notification as read.
func (r *NotificationsRepo) MarkRead(ctx context.Context, id int64) error {
	const q = `UPDATE notifications SET is_read = true WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	return nil
}

// MarkAllRead marks all notifications as read.
func (r *NotificationsRepo) MarkAllRead(ctx context.Context) error {
	const q = `UPDATE notifications SET is_read = true WHERE is_read = false`
	_, err := r.pool.Exec(ctx, q)
	if err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}

// UnreadCount returns the number of unread notifications.
func (r *NotificationsRepo) UnreadCount(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(*) FROM notifications WHERE is_read = false`
	var count int64
	if err := r.pool.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}
	return count, nil
}
