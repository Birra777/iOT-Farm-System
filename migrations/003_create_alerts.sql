CREATE TABLE IF NOT EXISTS alerts (
    id           BIGSERIAL PRIMARY KEY,
    field_id     UUID NOT NULL REFERENCES fields(id),
    sensor_type  TEXT NOT NULL,
    metric       TEXT NOT NULL,
    value        NUMERIC(10,4) NOT NULL,
    threshold    NUMERIC(10,4) NOT NULL,
    severity     TEXT NOT NULL CHECK (severity IN ('warning', 'critical')),
    message      TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'resolved')),
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_alerts_field_status ON alerts (field_id, status);
