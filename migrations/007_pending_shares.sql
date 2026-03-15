ALTER TABLE shares ADD COLUMN IF NOT EXISTS target_email TEXT;
ALTER TABLE shares ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'pending'));

CREATE INDEX IF NOT EXISTS idx_shares_pending_email ON shares(target_email) WHERE status = 'pending' AND NOT revoked;
