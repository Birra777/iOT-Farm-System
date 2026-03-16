CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS fields (
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name      TEXT NOT NULL,
    crop_type TEXT NOT NULL,
    hectares  NUMERIC(6,2),
    zone_code TEXT UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Seed the four farm zones.
INSERT INTO fields (name, crop_type, hectares, zone_code) VALUES
    ('Field 1', 'Maize',      42.00, 'NB'),
    ('Field 2', 'Sorghum',    28.00, 'RB'),
    ('Field 3', 'Millet',     19.00, 'DR'),
    ('Field 4', 'Groundnuts', 33.00, 'SP')
ON CONFLICT (zone_code) DO NOTHING;
