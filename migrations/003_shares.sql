CREATE TABLE shares (
    id              TEXT PRIMARY KEY,
    owner_id        UUID NOT NULL REFERENCES agents(id),
    source_path     TEXT NOT NULL,
    share_type      TEXT NOT NULL CHECK (share_type IN ('direct', 'link')),
    target_id       UUID REFERENCES agents(id),
    permission      TEXT NOT NULL DEFAULT 'read' CHECK (permission IN ('read', 'write')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ,
    revoked         BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at      TIMESTAMPTZ
);

CREATE INDEX idx_shares_owner ON shares(owner_id);
CREATE INDEX idx_shares_target ON shares(target_id);
CREATE INDEX idx_shares_active_links ON shares(id) WHERE share_type = 'link' AND NOT revoked;
