# pi-drive: Google Drive for AI Agents

Private file storage for AI agents. Files on S3, agents use `ls`, `grep`, `cat`. Share via URLs. Quotas + billing built in.

---

## What the agent sees

```
/drive/
  ├── my/              ← agent's private files
  │   ├── docs/
  │   ├── data.csv
  │   └── scripts/
  └── shared/          ← files shared with this agent
      ├── report.pdf        (from agent-b)
      └── dataset.json      (from agent-c)
```

No trace of other agents' files. Private by default.

---

## Agent commands

### Standard unix (just works, agent doesn't know it's S3)

```bash
ls /drive/my/
cat /drive/my/config.json
grep -r "error" /drive/my/logs/
find /drive/my/ -name "*.csv" -mtime -7
echo "hello" > /drive/my/notes.txt
cp results.csv /drive/my/reports/
mv /drive/my/draft.txt /drive/my/final.txt
mkdir -p /drive/my/projects/new-thing
rm /drive/my/temp.txt
wc -l /drive/my/data/*.csv
jq '.users[]' /drive/my/data.json
```

### pidrive CLI

```bash
# ── Auth ──────────────────────────────────────────────
pidrive register --email agent-a@company.com --name "Agent A"
# → ✓ API key: pk_7f3x9k2m...
# → Saved to ~/.pidrive/credentials

pidrive login --email agent-a@company.com
# → sends verification code to email → enter code → saves API key

pidrive whoami
# → agent-a@company.com (Agent A)
# → Plan: free (1 GB / 1 GB used)

# ── Mount ─────────────────────────────────────────────
pidrive mount
# → ✓ Mounted at /drive/

pidrive unmount
# → ✓ Unmounted

pidrive status
# → Mounted at /drive/
# → Agent: agent-a@company.com
# → Storage: 2.3 GB / 10 GB
# → Cache: 1.1 GB / 5 GB
# → Server: https://drive.dalo.dev ✓

# ── Sharing ───────────────────────────────────────────

# Share with a specific agent
pidrive share /drive/my/report.pdf --to agent-b@company.com
# → ✓ Shared. agent-b can access at /drive/shared/report.pdf

# Share with a URL (anyone with the link can access)
pidrive share /drive/my/data.csv --link
# → https://<PIDRIVE_SERVER>/s/abc123

# Share with a URL that expires
pidrive share /drive/my/data.csv --link --expires 7d
# → https://<PIDRIVE_SERVER>/s/abc123 (expires in 7 days)

# Share with write permission
pidrive share /drive/my/collab/ --to agent-b@company.com --permission write

# List all shares
pidrive shared
# SHARED BY ME:
#   report.pdf     → agent-b@company.com (read)
#   data.csv       → link: https://<PIDRIVE_SERVER>/s/abc123
# SHARED WITH ME:
#   results.json   ← agent-c@company.com (read)

# Pull a shared link
pidrive pull https://<PIDRIVE_SERVER>/s/abc123
# → ✓ data.csv → /drive/my/incoming/data.csv

pidrive pull https://<PIDRIVE_SERVER>/s/abc123 /drive/my/reports/
# → ✓ data.csv → /drive/my/reports/data.csv

# Revoke
pidrive revoke /drive/my/report.pdf --from agent-b@company.com
pidrive revoke https://<PIDRIVE_SERVER>/s/abc123

# ── Search ────────────────────────────────────────────
pidrive search "quarterly revenue"
# → /drive/my/reports/q1.txt:14: quarterly revenue was $1.2M
# → /drive/shared/summary.pdf:3: quarterly revenue summary

pidrive search "error" --type log,txt
pidrive search "config" --modified 7d
pidrive search "data" --my-only
pidrive search "report" --shared-only

# ── Trash ─────────────────────────────────────────────
pidrive trash
# → report-old.pdf   deleted 2 hours ago   recoverable until Apr 12
# → temp/            deleted 1 day ago     recoverable until Apr 11

pidrive restore report-old.pdf
# → ✓ Restored to /drive/my/report-old.pdf

pidrive trash empty
# → ✓ Permanently deleted all trash

# ── Activity ──────────────────────────────────────────
pidrive activity
# → 10:03  wrote     /drive/my/output.csv
# → 10:01  read      /drive/shared/input.json
# → 09:58  received  dataset.csv from agent-b@company.com
# → 09:45  deleted   /drive/my/temp.txt
# → 09:30  shared    report.pdf with agent-b@company.com

pidrive activity --since 1h
pidrive activity --type share

# ── Usage + Billing ───────────────────────────────────
pidrive usage
# → Storage:    2.3 GB / 10 GB (23%)
# → Shared out: 450 MB
# → Bandwidth:  15 GB this month
# → Plan:       pro ($5/mo)

pidrive upgrade --plan pro
# → ✓ Upgraded to pro (100 GB storage, $5/mo)

pidrive plans
# → free    1 GB storage     100 MB bandwidth/day    $0/mo
# → pro     100 GB storage   10 GB bandwidth/day     $5/mo
# → team    1 TB storage     unlimited bandwidth     $20/mo
```

---

## Architecture

```
Agent A                    Agent B                    Agent C
  │                          │                          │
  │ ls /drive/my/            │ grep /drive/shared/      │ pidrive share ...
  │                          │                          │
  ▼                          ▼                          ▼
┌──────────┐             ┌──────────┐             ┌──────────┐
│ JuiceFS  │             │ JuiceFS  │             │ JuiceFS  │
│ FUSE     │             │ FUSE     │             │ FUSE     │
│ client   │             │ client   │             │ client   │
└────┬─────┘             └────┬─────┘             └────┬─────┘
     │                        │                        │
     └────────────┬───────────┴────────────┬───────────┘
                  │                        │
                  ▼                        ▼
           ┌─────────────┐         ┌─────────────┐
           │   Redis     │         │     S3      │
           │  (metadata) │         │   (data)    │
           └─────────────┘         └─────────────┘

     ┌─────────────────────────────────────────────────┐
     │              pidrive server                      │
     │                                                  │
     │  ┌──────┐ ┌───────┐ ┌────────┐ ┌─────────────┐ │
     │  │ Auth │ │ Share │ │ Search │ │ Billing     │ │
     │  │ API  │ │ API   │ │ API    │ │ API         │ │
     │  └──┬───┘ └──┬────┘ └──┬─────┘ └──────┬──────┘ │
     │     │        │         │               │        │
     │  ┌──▼────────▼─────────▼───────────────▼─────┐  │
     │  │              Postgres                     │  │
     │  │  agents │ shares │ activity │ billing     │  │
     │  └───────────────────────────────────────────┘  │
     │                                                  │
     │  ENV: PIDRIVE_SERVER_URL                         │
     │  ENV: PIDRIVE_S3_BUCKET                          │
     │  ENV: PIDRIVE_REDIS_URL                          │
     │  ENV: PIDRIVE_DATABASE_URL                       │
     │  ENV: PIDRIVE_STRIPE_KEY (billing)               │
     └─────────────────────────────────────────────────┘
```

---

## Server API

### Auth

```
POST /api/register
  Body: { email, name }
  → { api_key: "pk_..." }

POST /api/login
  Body: { email }
  → sends verification code to email

POST /api/verify
  Body: { email, code }
  → { api_key: "pk_..." }

GET /api/me
  Header: Authorization: Bearer pk_...
  → { id, email, name, plan, used_bytes, quota_bytes }
```

### Mount

```
POST /api/mount
  Header: Authorization: Bearer pk_...
  → { agent_id, juicefs_config, s3_prefix, redis_url }
  Server creates agent's directory in JuiceFS if first mount.

POST /api/unmount
  Header: Authorization: Bearer pk_...
  → { ok: true }
```

### Sharing

```
POST /api/share
  Header: Authorization: Bearer pk_...
  Body: { path, to_email?, link?, permission, expires? }
  → { share_id, url? }

GET /api/shared
  Header: Authorization: Bearer pk_...
  → { shared_by_me: [...], shared_with_me: [...] }

DELETE /api/share/:id
  Header: Authorization: Bearer pk_...
  → { ok: true }

GET /s/:id
  (public endpoint for link shares)
  → redirects to S3 presigned URL (downloads file)
  If restricted: requires Authorization header
```

### Search

```
GET /api/search?q=revenue&type=csv,txt&modified=7d
  Header: Authorization: Bearer pk_...
  → { results: [{ path, filename, snippet, modified_at }] }

POST /api/index
  Header: Authorization: Bearer pk_...
  → triggers re-index for this agent
```

### Activity

```
GET /api/activity?since=1h&type=share&limit=50
  Header: Authorization: Bearer pk_...
  → { events: [{ timestamp, action, path, details }] }
```

### Billing

```
GET /api/plans
  → { plans: [{ id, name, storage_bytes, bandwidth_bytes, price_cents }] }

GET /api/usage
  Header: Authorization: Bearer pk_...
  → { used_bytes, quota_bytes, bandwidth_this_month, plan }

POST /api/upgrade
  Header: Authorization: Bearer pk_...
  Body: { plan: "pro" }
  → redirects to Stripe checkout
  → webhook updates plan in DB

GET /api/billing
  Header: Authorization: Bearer pk_...
  → { plan, next_billing_date, invoices: [...] }
```

### Trash

```
GET /api/trash
  Header: Authorization: Bearer pk_...
  → { items: [{ path, deleted_at, recoverable_until }] }

POST /api/trash/restore
  Header: Authorization: Bearer pk_...
  Body: { path }
  → { ok: true }

DELETE /api/trash
  Header: Authorization: Bearer pk_...
  → { ok: true } (empties trash)
```

---

## Database schema

```sql
-- ── Agents ───────────────────────────────────────────

CREATE TABLE agents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    api_key         TEXT UNIQUE NOT NULL,
    plan            TEXT NOT NULL DEFAULT 'free',
    quota_bytes     BIGINT NOT NULL DEFAULT 1073741824,   -- 1 GB
    used_bytes      BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    verified        BOOLEAN NOT NULL DEFAULT FALSE,
    verification_code TEXT,
    verification_expires TIMESTAMPTZ
);

CREATE INDEX idx_agents_email ON agents(email);
CREATE INDEX idx_agents_api_key ON agents(api_key);

-- ── Plans ────────────────────────────────────────────

CREATE TABLE plans (
    id              TEXT PRIMARY KEY,                      -- 'free', 'pro', 'team'
    name            TEXT NOT NULL,
    storage_bytes   BIGINT NOT NULL,
    bandwidth_bytes BIGINT NOT NULL,                       -- per day
    price_cents     INTEGER NOT NULL,                      -- per month
    stripe_price_id TEXT                                   -- Stripe price ID
);

INSERT INTO plans VALUES
    ('free', 'Free',  1073741824,   104857600,   0,    NULL),           -- 1GB, 100MB/day
    ('pro',  'Pro',   107374182400, 10737418240, 500,  'price_xxx'),    -- 100GB, 10GB/day
    ('team', 'Team',  1099511627776, -1,         2000, 'price_yyy');    -- 1TB, unlimited

-- ── Shares ───────────────────────────────────────────

CREATE TABLE shares (
    id              TEXT PRIMARY KEY,                      -- nanoid
    owner_id        UUID NOT NULL REFERENCES agents(id),
    source_path     TEXT NOT NULL,                         -- path relative to agent's root
    share_type      TEXT NOT NULL,                         -- 'direct' or 'link'
    target_id       UUID REFERENCES agents(id),            -- null for link shares
    permission      TEXT NOT NULL DEFAULT 'read',          -- 'read' or 'write'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ,                           -- null = never
    revoked         BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at      TIMESTAMPTZ
);

CREATE INDEX idx_shares_owner ON shares(owner_id);
CREATE INDEX idx_shares_target ON shares(target_id);
CREATE INDEX idx_shares_link ON shares(id) WHERE share_type = 'link' AND NOT revoked;

-- ── Activity log ─────────────────────────────────────

CREATE TABLE activity (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        UUID NOT NULL REFERENCES agents(id),
    action          TEXT NOT NULL,                         -- 'write','read','delete','share','revoke','mount','unmount'
    path            TEXT,
    details         JSONB,                                 -- extra info (share target, file size, etc.)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_activity_agent ON activity(agent_id, created_at DESC);
CREATE INDEX idx_activity_action ON activity(agent_id, action);

-- ── Search index ─────────────────────────────────────

CREATE TABLE search_index (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        UUID NOT NULL REFERENCES agents(id),
    path            TEXT NOT NULL,
    filename        TEXT NOT NULL,
    content         TEXT,                                  -- extracted text
    size_bytes      BIGINT,
    modified_at     TIMESTAMPTZ,
    indexed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    content_hash    TEXT,                                  -- skip re-indexing unchanged files
    UNIQUE(agent_id, path)
);

CREATE INDEX idx_search_agent ON search_index(agent_id);

-- Full-text search (Postgres)
ALTER TABLE search_index ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(filename, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(content, '')), 'B')
    ) STORED;

CREATE INDEX idx_search_fts ON search_index USING GIN(search_vector);

-- ── Billing ──────────────────────────────────────────

CREATE TABLE billing (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        UUID NOT NULL REFERENCES agents(id),
    stripe_customer_id TEXT,
    stripe_subscription_id TEXT,
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    status          TEXT NOT NULL DEFAULT 'active'         -- 'active','past_due','cancelled'
);

CREATE TABLE bandwidth_usage (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        UUID NOT NULL REFERENCES agents(id),
    date            DATE NOT NULL,
    bytes_in        BIGINT NOT NULL DEFAULT 0,
    bytes_out       BIGINT NOT NULL DEFAULT 0,
    UNIQUE(agent_id, date)
);

CREATE INDEX idx_bandwidth_agent ON bandwidth_usage(agent_id, date DESC);
```

---

## JuiceFS filesystem layout (internal)

```
juicefs://                                (full filesystem, only server/daemon sees this)
  ├── agents/
  │   ├── <agent-uuid-a>/
  │   │   ├── files/                      ← bind-mounted to /drive/my/
  │   │   └── .trash/                     ← JuiceFS trash
  │   ├── <agent-uuid-b>/
  │   │   ├── files/
  │   │   └── .trash/
  │   └── ...
  ├── shared/
  │   ├── direct/
  │   │   ├── to-<agent-uuid-a>/          ← bind-mounted to agent-a's /drive/shared/
  │   │   ├── to-<agent-uuid-b>/
  │   │   └── ...
  │   └── links/
  │       ├── <share-id>/
  │       │   └── data.csv
  │       └── <share-id>/
  │           └── report.pdf
  └── meta/
      └── (internal metadata)
```

---

## Isolation mechanism

```bash
# 1. Master mount (done once by pidrive daemon, hidden from agents)
juicefs mount $PIDRIVE_REDIS_URL /mnt/pidrive-master \
  --cache-dir /var/cache/pidrive \
  --cache-size 20480 \
  --trash-days 30

# 2. Per-agent bind mounts (done by `pidrive mount`)
mount --bind /mnt/pidrive-master/agents/<uuid>/files/        /drive/my/
mount --bind /mnt/pidrive-master/shared/direct/to-<uuid>/    /drive/shared/

# 3. Identity file (read-only, agent can't modify)
echo "<uuid>" > /drive/.agent-id
mount -o bind,ro /drive/.agent-id /drive/.agent-id

# Agent sees:
#   /drive/my/        (own files, read-write)
#   /drive/shared/    (shared files, read-only by default)
#   /drive/.agent-id  (identity, read-only)
```

For containers (stronger isolation):

```bash
# Agent runs in a container with only these mounts:
docker run \
  -v /mnt/pidrive-master/agents/<uuid>/files:/drive/my \
  -v /mnt/pidrive-master/shared/direct/to-<uuid>:/drive/shared:ro \
  -e PIDRIVE_API_KEY=pk_... \
  -e PIDRIVE_SERVER=https://drive.dalo.dev \
  agent-image
```

---

## Share flow (detailed)

### Direct share: agent-a shares report.pdf with agent-b

```
1. Agent A runs:
   pidrive share /drive/my/report.pdf --to agent-b@company.com

2. CLI sends to server:
   POST /api/share
   Authorization: Bearer pk_agent_a_...
   { path: "report.pdf", to_email: "agent-b@company.com", permission: "read" }

3. Server:
   - Validates API key → gets agent-a's UUID
   - Looks up agent-b by email → gets agent-b's UUID
   - Checks agent-a owns the file (path exists in /mnt/pidrive-master/agents/<a-uuid>/files/)
   - Copies file: /mnt/pidrive-master/agents/<a-uuid>/files/report.pdf
              → /mnt/pidrive-master/shared/direct/to-<b-uuid>/report.pdf
   - Inserts share record in Postgres
   - Logs activity for both agents
   - Returns { share_id: "abc123" }

4. Agent B (already mounted):
   ls /drive/shared/
   → report.pdf     ← it's there immediately (same JuiceFS filesystem)
```

### Link share: agent-a creates a URL

```
1. Agent A runs:
   pidrive share /drive/my/data.csv --link --expires 7d

2. CLI sends to server:
   POST /api/share
   Authorization: Bearer pk_agent_a_...
   { path: "data.csv", link: true, expires: "7d" }

3. Server:
   - Validates API key
   - Generates share ID: "x7k9m2"
   - Copies file: /mnt/pidrive-master/agents/<a-uuid>/files/data.csv
              → /mnt/pidrive-master/shared/links/x7k9m2/data.csv
   - Inserts share record with expires_at
   - Returns { share_id: "x7k9m2", url: "$PIDRIVE_SERVER_URL/s/x7k9m2" }

4. CLI prints:
   → https://drive.dalo.dev/s/x7k9m2

5. Anyone with this URL:
   curl https://drive.dalo.dev/s/x7k9m2 -o data.csv
   # OR
   pidrive pull https://drive.dalo.dev/s/x7k9m2

6. Server handles GET /s/x7k9m2:
   - Looks up share record
   - Checks not revoked, not expired
   - Generates S3 presigned URL for the file
   - 302 redirect to presigned URL
   - Logs download in activity
   - Tracks bandwidth
```

---

## Config

All server config via environment variables:

```bash
# Server
PIDRIVE_SERVER_URL=https://drive.dalo.dev     # public URL
PIDRIVE_PORT=8080                              # server listen port

# Storage
PIDRIVE_S3_BUCKET=pidrive-data
PIDRIVE_S3_REGION=us-east-1
PIDRIVE_S3_ENDPOINT=                           # custom endpoint for R2/MinIO
PIDRIVE_S3_ACCESS_KEY=AKIA...
PIDRIVE_S3_SECRET_KEY=...

# Metadata
PIDRIVE_REDIS_URL=redis://localhost:6379/1     # JuiceFS metadata engine

# Database
PIDRIVE_DATABASE_URL=postgres://user:pass@localhost:5432/pidrive

# Billing
PIDRIVE_STRIPE_SECRET_KEY=sk_...
PIDRIVE_STRIPE_WEBHOOK_SECRET=whsec_...

# Email (for verification codes)
PIDRIVE_SMTP_HOST=smtp.sendgrid.net
PIDRIVE_SMTP_PORT=587
PIDRIVE_SMTP_USER=apikey
PIDRIVE_SMTP_PASS=SG...
PIDRIVE_FROM_EMAIL=noreply@pidrive.io

# JuiceFS
PIDRIVE_JUICEFS_CACHE_DIR=/var/cache/pidrive
PIDRIVE_JUICEFS_CACHE_SIZE_MB=20480            # 20 GB
PIDRIVE_JUICEFS_TRASH_DAYS=30
```

Client config at `~/.pidrive/credentials`:

```toml
api_key = "pk_7f3x9k2m..."
server = "https://drive.dalo.dev"
mount_path = "/drive"
```

---

## Tech stack

| Component | Technology | Why |
|---|---|---|
| CLI | Go | Single binary, FUSE bindings, fast |
| Server | Go (net/http + chi router) | Same language as CLI, share code |
| Database | Postgres | Robust, full-text search built in, JSONB |
| File metadata | Redis | JuiceFS metadata engine, fast |
| File storage | S3 (any compatible) | JuiceFS data backend |
| FUSE | JuiceFS | Full POSIX, caching, trash, strong consistency |
| Billing | Stripe | Standard |
| Email | SMTP (SendGrid/SES) | Verification codes |
| Search | Postgres tsvector | Already in our DB, good enough |

---

## Project structure

```
pi-drive/
  ├── cmd/
  │   ├── pidrive/              ← CLI binary
  │   │   └── main.go
  │   └── pidrive-server/       ← server binary
  │       └── main.go
  ├── internal/
  │   ├── auth/
  │   │   ├── auth.go           ← register, login, verify, API key validation
  │   │   └── email.go          ← send verification codes
  │   ├── mount/
  │   │   ├── mount.go          ← JuiceFS mount/unmount
  │   │   ├── isolation.go      ← bind mounts, per-agent namespace
  │   │   └── cache.go          ← cache management
  │   ├── share/
  │   │   ├── share.go          ← share/revoke logic
  │   │   ├── links.go          ← link-based sharing, presigned URLs
  │   │   └── copy.go           ← file copy between agent dirs
  │   ├── search/
  │   │   ├── indexer.go        ← background file indexer
  │   │   └── search.go         ← full-text search queries
  │   ├── billing/
  │   │   ├── plans.go          ← plan definitions
  │   │   ├── usage.go          ← track storage + bandwidth
  │   │   ├── stripe.go         ← Stripe integration
  │   │   └── quota.go          ← enforce limits
  │   ├── activity/
  │   │   └── activity.go       ← log + query events
  │   ├── trash/
  │   │   └── trash.go          ← list, restore, empty trash
  │   ├── db/
  │   │   ├── db.go             ← Postgres connection
  │   │   └── migrations/       ← SQL migration files
  │   └── config/
  │       └── config.go         ← env var parsing
  ├── api/
  │   ├── router.go             ← HTTP routes
  │   ├── middleware.go         ← auth middleware, rate limiting
  │   ├── auth_handler.go
  │   ├── share_handler.go
  │   ├── search_handler.go
  │   ├── billing_handler.go
  │   ├── activity_handler.go
  │   └── trash_handler.go
  ├── cli/
  │   ├── root.go               ← cobra root command
  │   ├── register.go
  │   ├── login.go
  │   ├── mount.go
  │   ├── share.go
  │   ├── search.go
  │   ├── trash.go
  │   ├── activity.go
  │   ├── usage.go
  │   └── pull.go
  ├── migrations/
  │   ├── 001_agents.sql
  │   ├── 002_shares.sql
  │   ├── 003_activity.sql
  │   ├── 004_search.sql
  │   └── 005_billing.sql
  ├── Dockerfile                ← server
  ├── docker-compose.yml        ← local dev (server + postgres + redis + minio)
  ├── Makefile
  ├── go.mod
  ├── go.sum
  ├── README.md
  └── PLAN.md
```

---

## Build phases

### Phase 1: Server + Auth (days 1-3)

**Goal**: Agents can register and authenticate.

#### Day 1: Project scaffolding
- [ ] Init Go module
- [ ] Cobra CLI setup with root command
- [ ] Chi router setup for server
- [ ] Postgres connection + migration runner
- [ ] docker-compose.yml for local dev (postgres + redis + minio)
- [ ] Config from env vars
- [ ] Makefile (build cli, build server, run dev, migrate)

#### Day 2: Registration + login
- [ ] `POST /api/register` → create agent, generate API key, send verification email
- [ ] `POST /api/login` → send verification code to email
- [ ] `POST /api/verify` → verify code, return API key
- [ ] `GET /api/me` → return agent info
- [ ] API key generation: `pk_` + 32 random chars, hashed in DB
- [ ] Auth middleware: extract Bearer token, lookup agent
- [ ] SMTP integration for verification emails

#### Day 3: CLI auth commands
- [ ] `pidrive register --email --name`
- [ ] `pidrive login --email`
- [ ] `pidrive whoami`
- [ ] Credentials file read/write at `~/.pidrive/credentials`
- [ ] HTTP client wrapper that adds Authorization header
- [ ] Integration test: register → verify → whoami

### Phase 2: Mount + Trash (days 4-6)

**Goal**: Agent can mount and use `ls`/`grep`/`cat`. Deleted files recoverable.

#### Day 4: JuiceFS setup
- [ ] Server: `POST /api/mount` endpoint
  - Verify API key
  - Create agent dir in JuiceFS if doesn't exist: `/agents/<uuid>/files/`
  - Create shared dir: `/shared/direct/to-<uuid>/`
  - Return JuiceFS connection info
- [ ] Server startup: format JuiceFS filesystem if not exists
  - `juicefs format $PIDRIVE_REDIS_URL pidrive --storage s3 --bucket $PIDRIVE_S3_BUCKET`
- [ ] Server startup: mount master JuiceFS at `/mnt/pidrive-master`

#### Day 5: CLI mount
- [ ] `pidrive mount`
  - Call `POST /api/mount` to register mount + get config
  - Mount JuiceFS client on agent's machine OR set up bind mounts
  - Create `/drive/my/` and `/drive/shared/`
  - Write `/drive/.agent-id` (read-only)
  - Write PID file at `~/.pidrive/mount.pid`
- [ ] `pidrive unmount`
  - Unmount bind mounts
  - Call `POST /api/unmount`
  - Clean up PID file
- [ ] `pidrive status`

#### Day 6: Trash + testing
- [ ] Enable JuiceFS trash: `--trash-days 30`
- [ ] `GET /api/trash` → list trash items
- [ ] `POST /api/trash/restore` → restore file
- [ ] `DELETE /api/trash` → empty trash
- [ ] `pidrive trash`, `pidrive restore`, `pidrive trash empty`
- [ ] Integration test: mount → write file → ls → grep → cat → rm → trash → restore → unmount → remount → files still there
- [ ] Isolation test: mount agent-a and agent-b, verify they can't see each other

### Phase 3: Sharing (days 7-9)

**Goal**: Agents can share files directly and via URLs.

#### Day 7: Direct sharing
- [ ] `POST /api/share` (direct)
  - Validate owner has file
  - Copy file to target's shared dir
  - Insert share record
  - Log activity for both agents
- [ ] `GET /api/shared`
  - Query shares by owner (outgoing) and target (incoming)
- [ ] `DELETE /api/share/:id`
  - Remove file from shared dir
  - Mark revoked
- [ ] CLI: `pidrive share <path> --to <email>`
- [ ] CLI: `pidrive shared`
- [ ] CLI: `pidrive revoke <path> --from <email>`

#### Day 8: Link sharing
- [ ] `POST /api/share` (link)
  - Copy file to `/shared/links/<id>/`
  - Return URL: `$PIDRIVE_SERVER_URL/s/<id>`
- [ ] `GET /s/:id` (public endpoint)
  - Lookup share, check not revoked/expired
  - Generate S3 presigned URL
  - 302 redirect
  - Track bandwidth
- [ ] CLI: `pidrive share <path> --link [--expires 7d]`
- [ ] CLI: `pidrive pull <url> [dest]`

#### Day 9: Sharing edge cases
- [ ] Share directories (recursive copy)
- [ ] Write permission shares (bind mount with write)
- [ ] Filename conflicts: `report.pdf` → `report (from agent-a).pdf`
- [ ] Quota check before sharing (does target have space?)
- [ ] Share expiry cron job: clean up expired link shares
- [ ] Integration test: share → receive → pull link → revoke → can't access

### Phase 4: Search + Activity (days 10-12)

**Goal**: Full-text search across files. Activity log.

#### Day 10: Activity log
- [ ] Log all operations in activity table:
  - mount/unmount
  - share/revoke
  - file write/delete (from FUSE events or periodic scan)
- [ ] `GET /api/activity` with filters (since, type, limit)
- [ ] CLI: `pidrive activity [--since 1h] [--type share]`

#### Day 11: Search indexer
- [ ] Background indexer (runs on server):
  - Walks each agent's files directory periodically (every 60s)
  - Extracts text from .txt, .md, .csv, .json, .yaml, .log, .py, .js, .ts
  - Hashes content, skips unchanged files
  - Upserts into search_index table (triggers tsvector update)
  - Also indexes shared files accessible to each agent
- [ ] `POST /api/index` → trigger re-index for an agent

#### Day 12: Search command
- [ ] `GET /api/search?q=...&type=...&modified=...`
  - Query: `search_vector @@ plainto_tsquery('english', $query)`
  - Filter by agent_id (own files + files shared with agent)
  - Filter by extension, modification date
  - Return path, filename, snippet (ts_headline), modified_at
- [ ] CLI: `pidrive search <query> [--type csv,txt] [--modified 7d] [--my-only] [--shared-only]`
- [ ] Integration test: write files → wait for index → search → results match

### Phase 5: Billing (days 13-15)

**Goal**: Usage tracking, quotas, paid plans.

#### Day 13: Usage tracking
- [ ] Track storage: update `agents.used_bytes` on file write/delete
  - Option A: periodic scan of agent directory (`du -sb`)
  - Option B: JuiceFS quota/stats commands
- [ ] Track bandwidth: increment `bandwidth_usage` on file download/share pull
- [ ] `GET /api/usage` → return storage + bandwidth stats
- [ ] CLI: `pidrive usage`
- [ ] Quota enforcement: reject writes when over quota
  - On `POST /api/share`: check target has space
  - On mount: set JuiceFS quota if supported, or check server-side

#### Day 14: Stripe integration
- [ ] `GET /api/plans` → return available plans
- [ ] `POST /api/upgrade` → create Stripe checkout session → redirect
- [ ] Stripe webhook handler:
  - `checkout.session.completed` → update plan + quota
  - `invoice.payment_failed` → mark past_due
  - `customer.subscription.deleted` → downgrade to free
- [ ] CLI: `pidrive plans`, `pidrive upgrade --plan pro`

#### Day 15: Billing polish
- [ ] `GET /api/billing` → current plan, next billing date, invoices
- [ ] Downgrade handling: if usage > new quota, block new writes but don't delete
- [ ] Grace period for past_due (7 days before downgrade)
- [ ] CLI: `pidrive billing`



### Phase 6: Install script (day 16)

**Goal**: One command to install pidrive CLI on any machine.

#### Day 16: Install script
- [ ] `install.sh` script hosted at server URL
- [ ] Detects OS (linux/darwin) and arch (amd64/arm64)
- [ ] Downloads `pidrive` binary from server or GitHub releases
- [ ] Downloads `juicefs` binary
- [ ] Installs FUSE if missing:
  - Linux: `apt install -y fuse3` or `yum install -y fuse3`
  - macOS: prompts to install macFUSE
- [ ] Puts binaries in `/usr/local/bin/`
- [ ] Prints next steps

```bash
# Agent installs with one command:
curl -sSL https://<PIDRIVE_SERVER>/install.sh | sh

# Then:
pidrive register --email agent@company.com --name "My Agent"
pidrive mount
ls /drive/my/
```

install.sh:
```bash
#!/bin/bash
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
esac

PIDRIVE_VERSION="latest"
JUICEFS_VERSION="1.2.2"

echo "Installing pidrive for $OS/$ARCH..."

# Install FUSE
if [ "$OS" = "linux" ]; then
  if ! command -v fusermount3 &>/dev/null; then
    echo "Installing FUSE..."
    if command -v apt &>/dev/null; then
      sudo apt update && sudo apt install -y fuse3
    elif command -v yum &>/dev/null; then
      sudo yum install -y fuse3
    fi
  fi
elif [ "$OS" = "darwin" ]; then
  if ! test -d /Library/Filesystems/macfuse.fs; then
    echo "macFUSE required. Install from: https://osxfuse.github.io/"
    echo "Then re-run this script."
    exit 1
  fi
fi

# Install JuiceFS
if ! command -v juicefs &>/dev/null; then
  echo "Installing JuiceFS $JUICEFS_VERSION..."
  curl -sSL https://d.juicefs.com/install | sh -
fi

# Install pidrive
echo "Installing pidrive..."
curl -sSLo /usr/local/bin/pidrive \
  "${PIDRIVE_SERVER:-https://github.com/your-org/pi-drive/releases/download/$PIDRIVE_VERSION}/pidrive-$OS-$ARCH"
chmod +x /usr/local/bin/pidrive

echo ""
echo "✓ pidrive installed!"
echo ""
echo "Next steps:"
echo "  pidrive register --email you@company.com --name \"My Agent\" --server https://your-server.com"
echo "  pidrive mount"
echo "  ls /drive/my/"
```

- [ ] Server endpoint `GET /install.sh` serves the script with `PIDRIVE_SERVER` baked in
- [ ] `make release` builds CLI for all OS/arch combos, uploads to GitHub releases
- [ ] Makefile targets: `build-all` (cross-compile), `release` (tag + upload)

---

## Dev setup (local, no cloud)

```bash
# 1. Clone
git clone https://github.com/your-org/pi-drive
cd pi-drive

# 2. Start infra
docker compose up -d   # postgres + redis + minio

# 3. Run migrations
make migrate

# 4. Start server
make run-server
# → listening on :8080

# 5. Build CLI
make build-cli
# → ./bin/pidrive

# 6. Register + mount
./bin/pidrive register --email test@test.com --name "Test Agent"
./bin/pidrive mount
ls /drive/my/
```

docker-compose.yml:
```yaml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: pidrive
      POSTGRES_USER: pidrive
      POSTGRES_PASSWORD: pidrive
    ports: ["5432:5432"]

  redis:
    image: redis:7
    ports: ["6379:6379"]

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports: ["9000:9000", "9001:9001"]
```

---

## Open questions

1. **How does the agent machine get JuiceFS access?**
   - Option A: Agent machine mounts JuiceFS directly (needs Redis URL + S3 creds)
   - Option B: Server provides a scoped JuiceFS config via `/api/mount`
   - Option C: Server runs a WebDAV/NFS gateway, agent mounts that instead
   - **Recommendation**: Option B for now. Server returns JuiceFS config with scoped creds.

2. **Copy vs hardlink for sharing?**
   - Copy: safe, revoke = delete copy, uses extra storage
   - Hardlink: instant, no extra storage, but revoke is messy
   - **Decision**: Copy. Storage is cheap. Safety > efficiency.

3. **Search: server-side vs client-side?**
   - Server-side: indexer runs on server, search via API. Works even when not mounted.
   - Client-side: `pidrive search` runs grep locally on mounted files.
   - **Decision**: Server-side for structured search. Agent can also just `grep` locally.

4. **Email provider for verification?**
   - SendGrid, AWS SES, or any SMTP server.
   - For dev: log codes to console, skip actual email.

5. **Do we need rate limiting?**
   - Yes. Per API key. 100 req/min for free, 1000 for pro.
   - Use Redis for rate limit counters.
