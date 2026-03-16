CREATE TABLE IF NOT EXISTS global_thresholds (
    metric        TEXT PRIMARY KEY,
    warning_low   NUMERIC,
    critical_low  NUMERIC,
    warning_high  NUMERIC,
    critical_high NUMERIC,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed with current hardcoded defaults so the UI shows them immediately.
INSERT INTO global_thresholds (metric, warning_low, critical_low, warning_high, critical_high) VALUES
    ('soil.moisture',        30,   20,   NULL, NULL),
    ('soil.ph',              5.5,  NULL, 7.5,  NULL),
    ('soil.nitrogen',        100,  80,   NULL, NULL),
    ('weather.temperature',  NULL, NULL, 35,   38),
    ('weather.humidity',     NULL, NULL, 85,   90)
ON CONFLICT (metric) DO NOTHING;
