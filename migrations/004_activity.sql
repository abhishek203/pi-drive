CREATE TABLE activity (
    id          BIGSERIAL PRIMARY KEY,
    agent_id    UUID NOT NULL REFERENCES agents(id),
    action      TEXT NOT NULL,
    path        TEXT,
    details     JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_activity_agent ON activity(agent_id, created_at DESC);
CREATE INDEX idx_activity_action ON activity(agent_id, action);
