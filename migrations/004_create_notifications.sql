-- Notifications: plain-English, user-facing messages derived from alerts.
-- Created when the anomaly detector fires an alert. Stored independently
-- so the dashboard can poll and display them without re-deriving the text.

CREATE TABLE IF NOT EXISTS notifications (
    id         BIGSERIAL PRIMARY KEY,
    alert_id   BIGINT      REFERENCES alerts(id) ON DELETE SET NULL,
    field_id   UUID        NOT NULL REFERENCES fields(id),
    title      TEXT        NOT NULL,
    body       TEXT        NOT NULL,
    severity   TEXT        NOT NULL CHECK (severity IN ('warning', 'critical')),
    is_read    BOOLEAN     NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_created ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_unread  ON notifications(is_read) WHERE is_read = false;
