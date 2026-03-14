CREATE TABLE search_index (
    id            BIGSERIAL PRIMARY KEY,
    agent_id      UUID NOT NULL REFERENCES agents(id),
    path          TEXT NOT NULL,
    filename      TEXT NOT NULL,
    content       TEXT,
    size_bytes    BIGINT,
    modified_at   TIMESTAMPTZ,
    indexed_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    content_hash  TEXT,
    UNIQUE(agent_id, path)
);

CREATE INDEX idx_search_agent ON search_index(agent_id);

ALTER TABLE search_index ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(filename, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(content, '')), 'B')
    ) STORED;

CREATE INDEX idx_search_fts ON search_index USING GIN(search_vector);
