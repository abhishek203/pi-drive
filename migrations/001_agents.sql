CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE agents (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                 TEXT UNIQUE NOT NULL,
    name                  TEXT NOT NULL,
    api_key_hash          TEXT UNIQUE NOT NULL,
    api_key_prefix        TEXT NOT NULL,
    plan                  TEXT NOT NULL DEFAULT 'free',
    quota_bytes           BIGINT NOT NULL DEFAULT 1073741824,
    used_bytes            BIGINT NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    verified              BOOLEAN NOT NULL DEFAULT FALSE,
    verification_code     TEXT,
    verification_expires  TIMESTAMPTZ
);

CREATE INDEX idx_agents_email ON agents(email);
CREATE INDEX idx_agents_api_key_hash ON agents(api_key_hash);
CREATE INDEX idx_agents_api_key_prefix ON agents(api_key_prefix);
