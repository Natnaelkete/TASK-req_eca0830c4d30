# Agricultural Research Data and Results Operation Platform API

**Project Type:** Backend

A backend API supporting administrators, researchers, reviewers, and customer service personnel collaborating under a unified data framework for agricultural research.

## Tech Stack

- **Language:** Go 1.22
- **Framework:** Gin
- **ORM:** GORM
- **Database:** MySQL 8.0
- **Containerization:** Docker + Docker Compose

## Quick Start

```bash
# Clone the repository
git clone <repo-url>
cd repo

# Copy environment config
cp .env.example .env

# Start all services (MySQL + API) — Docker Compose v2 syntax
docker compose up --build

# Or with legacy Docker Compose v1 syntax (same result):
docker-compose up --build

# Shorter form once the images are built:
# docker-compose up
```

The API will be available at `http://localhost:8080`.

## Demo Credentials

The stack ships with four fixed demo accounts covering every role. They are
seeded on first `docker compose up` via the registration endpoint so you can
log in immediately without creating users manually.

| Role | Username | Password | Email |
|------|----------|----------|-------|
| admin | `demo_admin` | `Admin1234` | `demo_admin@example.com` |
| researcher | `demo_researcher` | `Research1234` | `demo_researcher@example.com` |
| reviewer | `demo_reviewer` | `Review1234` | `demo_reviewer@example.com` |
| customer_service | `demo_cs` | `Support1234` | `demo_cs@example.com` |
| viewer | `demo_viewer` | `Viewer1234` | `demo_viewer@example.com` |

> The `admin`, `reviewer`, and `customer_service` accounts must be provisioned
> by an admin (they cannot be self-registered at `/v1/auth/register` since that
> path only accepts `researcher` and `viewer`). Use the [seed script](scripts/seed_demo_users.sh)
> to bootstrap them against a running stack, or create them through the
> admin-only user-creation path.

Seed them locally with the bundled helper:

```bash
./scripts/seed_demo_users.sh
```

Then log in as the role you need, e.g.:

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo_researcher","password":"Research1234"}'
```

## Services

| Service | Host | Port |
|---------|------|------|
| API     | localhost | 8080 |
| MySQL   | localhost | 3307 (mapped from 3306) |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | MySQL host |
| `DB_PORT` | `3306` | MySQL port |
| `DB_USER` | `root` | MySQL user |
| `DB_PASSWORD` | `pass` | MySQL password |
| `DB_NAME` | `agri` | Database name |
| `JWT_SECRET` | `change-me-in-production` | JWT signing secret |
| `SERVER_PORT` | `8080` | API server port |
| `ENCRYPTION_KEY` | `0123456789abcdef0123456789abcdef` | 32-byte AES key for field encryption |

## Verification

```bash
# Health check
curl http://localhost:8080/health

# Register a user
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","email":"admin@test.com","password":"Admin123!","role":"admin"}'

# Login and get token
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin123!"}'

# Use the token for protected endpoints (bash / zsh)
export TOKEN="<token-from-login-response>"

# Windows PowerShell variant
# $env:TOKEN = "<token-from-login-response>"

# List plots
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/plots

# Check system capacity
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/system/capacity
```

## Role & Permission Matrix

The API enforces role-based access control (RBAC) via JWT claims. Every
protected endpoint group lists the roles allowed to call it below.

| Endpoint Group | admin | researcher | reviewer | customer_service | viewer |
|---|---|---|---|---|---|
| `POST /v1/auth/register` (public) | self-register blocked | yes | blocked | blocked | yes |
| `GET /v1/auth/me` | yes | yes | yes | yes | yes |
| `/v1/plots/*` write | yes | yes | — | — | — |
| `/v1/plots/*` read | yes | yes | yes | yes | yes |
| `/v1/devices/*` write | yes | yes | — | — | — |
| `/v1/metrics/*` write | yes | yes | — | — | — |
| `/v1/monitor/*` write | yes | yes | — | — | — |
| `/v1/monitoring/ingest` | yes | yes | — | — | — |
| `/v1/monitoring/data` read | yes | yes | yes | yes | yes |
| `/v1/dashboards/*` | yes | yes | yes | yes | yes (user-scoped) |
| `/v1/analysis/*` | yes | yes | yes | — | — |
| `/v1/indicators/*` write | yes | yes | — | — | — |
| `/v1/indicators/:id` DELETE | yes | — | — | — | — |
| `/v1/orders/*` | yes | yes | yes | yes | — |
| `/v1/templates` write | yes | — | — | yes | — |
| `/v1/tasks/*` write | yes | yes | — | — | — |
| `/v1/tasks/:id/review` \| `/complete` | yes | yes | yes | — | — |
| `/v1/results/*` write | yes | yes | — | — | — |
| `/v1/results/:id/transition` \| `/notes` | yes | yes | yes | — | — |
| `/v1/results/:id/invalidate` | yes | — | — | — | — |
| `/v1/results/field-rules` write | yes | — | — | — | — |
| `/v1/system/*` | yes | — | — | — | — |

"yes" = allowed, "—" = forbidden (returns 403).

## API Endpoints

All endpoints except `/health` and `/v1/auth/register|login` require a Bearer JWT token.

### Authentication
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/auth/register` | Register new user |
| POST | `/v1/auth/login` | Login, returns JWT |
| GET | `/v1/auth/me` | Current user profile |

### Plots (Master Data)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/plots` | Create plot |
| GET | `/v1/plots` | List plots |
| GET | `/v1/plots/:id` | Get plot |
| PUT | `/v1/plots/:id` | Update plot |
| DELETE | `/v1/plots/:id` | Delete plot |

### Devices (Master Data)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/devices` | Create device |
| GET | `/v1/devices` | List devices |
| GET | `/v1/devices/:id` | Get device |
| PUT | `/v1/devices/:id` | Update device |
| DELETE | `/v1/devices/:id` | Delete device |

### Metrics (Indicator System)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/metrics` | Create metric |
| POST | `/v1/metrics/batch` | Batch create metrics |
| GET | `/v1/metrics` | List metrics |
| GET | `/v1/metrics/:id` | Get metric |
| DELETE | `/v1/metrics/:id` | Delete metric |

### Monitoring (Device Health & Alerts)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/monitor/device` | Enqueue device health check |
| POST | `/v1/monitor/threshold` | Enqueue threshold check |
| GET | `/v1/monitor/jobs/:id` | Get job status |
| GET | `/v1/monitor/queue/status` | Queue statistics |
| GET | `/v1/monitor/alerts` | List alerts |
| PATCH | `/v1/monitor/alerts/:id/resolve` | Resolve alert |

### Monitoring Data (Ingest, Query, Export)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/monitoring/ingest` | Batch ingest via async queue (idempotent) |
| GET | `/v1/monitoring/data` | List monitoring data (filterable) |
| GET | `/v1/monitoring/data/:id` | Get single record |
| POST | `/v1/monitoring/aggregate` | Multi-dimensional aggregation (count/min/max/avg/sum) |
| POST | `/v1/monitoring/curve` | Real-time curves (last N minutes) |
| POST | `/v1/monitoring/trends` | Trend statistics (daily/weekly/monthly + YoY/MoM) |
| GET | `/v1/monitoring/export/json` | Export data as JSON |
| GET | `/v1/monitoring/export/csv` | Export data as CSV |
| GET | `/v1/monitoring/jobs/:id` | Ingest job status |

### Dashboards
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/dashboards` | Save dashboard config |
| GET | `/v1/dashboards` | List user dashboards |
| GET | `/v1/dashboards/:id` | Load dashboard |
| PUT | `/v1/dashboards/:id` | Update dashboard |
| DELETE | `/v1/dashboards/:id` | Delete dashboard |

### Analysis (Trends, Funnels, Retention)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/analysis/trends` | Trend analysis with drill-down |
| POST | `/v1/analysis/funnels` | Funnel analysis with conversion rates |
| POST | `/v1/analysis/retention` | Cohort retention analysis |

### Orders & Conversations (Communication)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/orders` | Create order |
| GET | `/v1/orders` | List orders |
| GET | `/v1/orders/:id` | Get order |
| POST | `/v1/orders/:id/messages` | Post message (rate limited: 20/min, sensitive word filtered) |
| GET | `/v1/orders/:id/messages` | List conversation messages |
| PATCH | `/v1/orders/:id/messages/:msg_id/read` | Mark message as read |
| POST | `/v1/orders/:id/transfer` | Transfer ticket to another user |
| POST | `/v1/orders/:id/templates/:template_id` | Send template message |

### Templates
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/templates` | Create template |
| GET | `/v1/templates` | List templates |

### Tasks (Evaluation)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/tasks` | Create task |
| POST | `/v1/tasks/generate` | Generate tasks by object/cycle |
| GET | `/v1/tasks` | List tasks (visibility window filtered) |
| GET | `/v1/tasks/:id` | Get task |
| PUT | `/v1/tasks/:id` | Update task |
| DELETE | `/v1/tasks/:id` | Delete task |
| PATCH | `/v1/tasks/:id/submit` | Submit task |
| PATCH | `/v1/tasks/:id/review` | Move to review |
| PATCH | `/v1/tasks/:id/complete` | Complete task |

### Results (Management)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/results` | Create result (paper/project/patent) |
| GET | `/v1/results` | List results |
| GET | `/v1/results/:id` | Get result |
| PUT | `/v1/results/:id` | Update result (blocked if archived) |
| DELETE | `/v1/results/:id` | Delete result |
| PATCH | `/v1/results/:id/transition` | Status transition (draft->submitted->returned->approved->archived) |
| POST | `/v1/results/:id/notes` | Append notes (archived only) |
| POST | `/v1/results/:id/invalidate` | Invalidate with reason (archived only) |
| POST | `/v1/results/field-rules` | Create field validation rule |
| GET | `/v1/results/field-rules` | List field rules |

### System
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/system/capacity` | Check disk usage |
| GET | `/v1/system/notifications` | List system notifications |

### Chat (Legacy)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/chat` | Send message |
| GET | `/v1/chat` | List messages |
| PATCH | `/v1/chat/:id/read` | Mark as read |

## Key Features

- **Batch ingestion** via async queue with idempotent keys (`source_id`, `event_time`, `metric_code`)
- **Time-series indexes** on `(device_id, plot_id, metric_code, event_time)`
- **Result state machine**: `draft -> submitted -> returned -> approved -> archived` (strict)
- **Post-archive controls**: only retrospective notes; invalidation requires reason + full traceability
- **Sensitive word filtering**: messages containing blocked words are intercepted and logged
- **Rate limiting**: 20 messages per minute per user
- **Global audit logging**: every API request recorded with operator, timestamp, resource, action
- **Capacity monitoring**: background check every 60s, alerts when disk > 80%
- **Overdue task handling**: tasks auto-marked as delayed after 7 days past deadline

## Running Tests

```bash
# Using the test runner script (recommended)
./run_tests.sh

# Or via Makefile
make test

# Or directly with Go
go test ./... -coverprofile=cov.out -cover -v

# Inside Docker build environment (no local Go required)
docker build --target builder -t agri-build-test .
docker run --rm agri-build-test sh -c "cd /app && go test ./... -v -cover"
```

## Project Structure

```
repo/
├── cmd/server/main.go          # Entry point, router setup, graceful shutdown
├── internal/config/             # Environment-based configuration
├── pkg/
│   ├── handlers/                # HTTP request handlers (Gin)
│   ├── middleware/               # Auth (JWT), Audit logging
│   ├── models/                  # GORM models + DB init
│   └── services/                # Business logic layer
├── migrations/
│   ├── 001_create.sql           # Initial database schema
│   └── 002_indicator_versions_and_partitioning.sql  # Indicators, partitioning, archive
├── Dockerfile                   # Multi-stage build
├── docker-compose.yml           # MySQL + API orchestration
├── Makefile                     # Build/test/docker commands
├── run_tests.sh                 # Shell script to run all tests with coverage
├── go.mod / go.sum              # Go module dependencies
└── .env.example                 # Environment variable template
```
