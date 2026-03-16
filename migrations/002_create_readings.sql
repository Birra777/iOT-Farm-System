CREATE TABLE IF NOT EXISTS sensor_readings (
    id          BIGSERIAL PRIMARY KEY,
    field_id    UUID NOT NULL REFERENCES fields(id),
    sensor_id   TEXT NOT NULL,
    sensor_type TEXT NOT NULL CHECK (sensor_type IN ('soil', 'weather')),
    metric      TEXT NOT NULL,
    value       NUMERIC(10,4) NOT NULL,
    unit        TEXT NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_readings_field_metric_time
    ON sensor_readings (field_id, metric, recorded_at DESC);
