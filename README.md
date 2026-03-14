# pidrive — Google Drive for AI Agents

Private file storage for AI agents. Files on S3, agents use `ls`, `grep`, `cat`. Share via URLs. Quotas + billing built in.

## How it works

```
Agent A (any VM)              Agent B (any VM)
  │                             │
  │ ls ~/drive/                 │ cat ~/drive/report.pdf
  │ grep -r "error" ~/drive/    │
  ▼                             ▼
  WebDAV mount ── HTTPS ──────► pidrive server
                                  │
                                  ▼
                          ┌──────────────┐
                          │ JuiceFS FUSE │
                          └──────┬───────┘
                                 │
                          ┌──────┴──────┐
                          ▼             ▼
                     ┌─────────┐  ┌──────────┐
                     │  Redis  │  │  AWS S3   │
                     │  (meta) │  │  (data)   │
                     └─────────┘  └──────────┘
```

Agents mount `~/drive/` via WebDAV over HTTPS. Standard unix commands just work.
No extra drivers needed — WebDAV is built into macOS and Linux.
Each agent is isolated. Sharing is explicit.

## Install

```bash
curl -sSL https://pidrive.ressl.ai/install.sh | bash
```

Or build from source:

```bash
git clone https://github.com/your-org/pi-drive
cd pi-drive
make build
# → bin/pidrive (CLI)
# → bin/pidrive-server (server)
```

## Quick start

### 1. Register

```bash
pidrive register --email agent@company.com --name "My Agent" --server https://pidrive.ressl.ai
pidrive verify --email agent@company.com --code 123456
```

### 2. Mount

```bash
pidrive mount
ls ~/drive/
echo "hello world" > ~/drive/test.txt
grep "hello" ~/drive/test.txt
```

### 3. Share

```bash
# Share with another agent
pidrive share report.pdf --to agent-b@company.com

# Create a link
pidrive share data.csv --link
# → https://pidrive.ressl.ai/s/abc123

# See shares
pidrive shared
```

## Commands

| Command | Description |
|---|---|
| `pidrive register` | Create a new agent account |
| `pidrive login` | Login to existing account |
| `pidrive verify` | Verify with email code |
| `pidrive whoami` | Show current agent info |
| `pidrive mount` | Mount drive via WebDAV |
| `pidrive unmount` | Unmount drive |
| `pidrive status` | Show mount + connection status |
| `pidrive share <path> --to <email>` | Share with agent |
| `pidrive share <path> --link` | Create shareable URL |
| `pidrive shared` | List all shares |
| `pidrive pull <url> [dest]` | Download shared file |
| `pidrive revoke <id>` | Revoke a share |
| `pidrive search <query>` | Full-text search |
| `pidrive activity` | Recent activity log |
| `pidrive trash` | List deleted files |
| `pidrive restore <path>` | Restore from trash |
| `pidrive usage` | Storage + bandwidth stats |
| `pidrive plans` | Show available plans |
| `pidrive upgrade --plan pro` | Upgrade plan |

## Server API

### Public
- `POST /api/register` — register agent
- `POST /api/login` — send verification code
- `POST /api/verify` — verify code, get API key
- `GET /api/plans` — list plans
- `GET /s/:id` — download shared file
- `GET /skill.md` — agent skill doc
- `GET /install.sh` — install script

### Authenticated (Bearer token)
- `GET /api/me` — agent info
- `GET /api/whoami` — alias for /api/me
- `POST /api/mount` — register mount, create agent dirs
- `POST /api/unmount` — unregister mount
- `POST /api/share` — share a file (direct)
- `POST /api/share/link` — share a file (link)
- `GET /api/shared` — list shares
- `DELETE /api/share/:id` — revoke share
- `GET /api/search?q=...` — search files
- `GET /api/activity` — activity log
- `GET /api/trash` — list trash
- `POST /api/trash/restore` — restore file
- `DELETE /api/trash` — empty trash
- `GET /api/usage` — storage stats
- `GET /api/billing` — billing info
- `POST /api/upgrade` — upgrade plan

### WebDAV (Basic Auth with API key as password)
- `/webdav/` — full WebDAV filesystem access (PROPFIND, GET, PUT, DELETE, MKCOL, LOCK, UNLOCK)

## Architecture

### Agent side
- `pidrive` CLI binary (~10 MB)
- WebDAV mount (built into macOS and Linux — no extra drivers)
- Recently accessed files cached locally for performance

### Server side
- `pidrive-server` — HTTP API (:8080) + WebDAV handler
- JuiceFS FUSE mount at `/mnt/pidrive` — stores data in S3, metadata in Redis
- Postgres — agents, shares, search index, activity, billing
- Redis — JuiceFS metadata (DB 1), app cache (DB 0)
- Caddy — HTTPS reverse proxy with auto Let's Encrypt

### Data flow
```
Agent: echo "hello" > ~/drive/test.txt
  → WebDAV PUT over HTTPS
  → pidrive server authenticates via API key (Basic Auth)
  → writes to /mnt/pidrive/agents/{agent-id}/files/test.txt
  → JuiceFS splits into chunks, stores in S3
  → metadata updated in Redis
```

## Environment variables

```bash
# Server
PIDRIVE_SERVER_URL=https://pidrive.ressl.ai
PIDRIVE_PORT=8080
PIDRIVE_DB_URL=postgres://pidrive:pidrive@localhost:5432/pidrive?sslmode=disable
PIDRIVE_REDIS_URL=redis://localhost:6379/0
PIDRIVE_JUICEFS_MOUNT_PATH=/mnt/pidrive
PIDRIVE_S3_BUCKET=pidrive-storage-prod
PIDRIVE_S3_REGION=us-east-2
PIDRIVE_RESEND_API_KEY=re_...
PIDRIVE_FROM_EMAIL=noreply@agents.ressl.ai
```

## Tech stack

- **CLI + Server**: Go
- **Filesystem**: JuiceFS (FUSE) on server, WebDAV mount on agent
- **WebDAV**: golang.org/x/net/webdav
- **Database**: Postgres 16 (agents, shares, search, billing)
- **Metadata**: Redis 7 (JuiceFS metadata engine, AOF persistence)
- **Storage**: AWS S3 (us-east-2)
- **Search**: Postgres tsvector full-text search
- **Email**: Resend API
- **HTTPS**: Caddy with auto Let's Encrypt
- **Billing**: Stripe (stub)

## License

MIT
