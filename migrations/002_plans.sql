CREATE TABLE plans (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    storage_bytes   BIGINT NOT NULL,
    bandwidth_bytes BIGINT NOT NULL,
    price_cents     INTEGER NOT NULL,
    stripe_price_id TEXT
);

INSERT INTO plans VALUES
    ('free', 'Free',  1073741824,    104857600,    0,    NULL),
    ('pro',  'Pro',   107374182400,  10737418240,  500,  NULL),
    ('team', 'Team',  1099511627776, -1,           2000, NULL);
