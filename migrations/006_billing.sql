CREATE TABLE billing (
    id                      BIGSERIAL PRIMARY KEY,
    agent_id                UUID NOT NULL REFERENCES agents(id) UNIQUE,
    stripe_customer_id      TEXT,
    stripe_subscription_id  TEXT,
    current_period_start    TIMESTAMPTZ,
    current_period_end      TIMESTAMPTZ,
    status                  TEXT NOT NULL DEFAULT 'active'
);

CREATE TABLE bandwidth_usage (
    id          BIGSERIAL PRIMARY KEY,
    agent_id    UUID NOT NULL REFERENCES agents(id),
    date        DATE NOT NULL,
    bytes_in    BIGINT NOT NULL DEFAULT 0,
    bytes_out   BIGINT NOT NULL DEFAULT 0,
    UNIQUE(agent_id, date)
);

CREATE INDEX idx_bandwidth_agent ON bandwidth_usage(agent_id, date DESC);
