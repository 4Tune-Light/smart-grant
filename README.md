# Smart Grant Review Platform

> Production-like grant management platform with AI-assisted risk scoring.
> Built for CV/portfolio — demonstrates senior-level backend engineering.

[![Go Version](https://img.shields.io/badge/go-1.26-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Highlights

- **C4.5 Decision Tree** — risk scoring algorithm implemented from scratch
- **Async Processing** — Redis Streams for decoupled task execution
- **Full Observability** — OpenTelemetry traces + Prometheus metrics + Jaeger tracing
- **RBAC** — JWT-based role access control (admin, reviewer, applicant)
- **gRPC** — Internal service-to-service communication
- **Immutable Audit Trail** — append-only audit log for compliance
- **Clean Architecture** — handler → service → repository pattern

## Architecture

```
Client (SPA)
    |
API Gateway (REST — rate limiter, JWT passthrough)
    |
Backend Service (HTTP + gRPC)
    |
    --- PostgreSQL  (source of truth)
    --- Redis       (cache / async queue via Streams)
    --- Worker      (async: risk scoring, notifications)
    |
OpenTelemetry Collector
    |
    --- Prometheus  (metrics)
    --- Jaeger      (distributed tracing)
```

## Project Structure

```
├── cmd/
│   ├── api-gateway/      REST API gateway with rate limiting and auth
│   ├── backend/          Core backend service (REST handlers + gRPC)
│   └── worker/           Async worker (Redis Streams consumer)
├── internal/
│   ├── auth/             Authentication, JWT, RBAC
│   ├── proposal/         Proposal CRUD and lifecycle management
│   ├── review/           Review scoring and approval workflow
│   ├── risk/             C4.5 decision tree risk scoring engine
│   ├── audit/            Immutable audit trail logger
│   ├── notification/     Notification system with SSE
│   ├── config/           Config loader and type definitions
│   ├── middleware/        HTTP middleware (auth, logger, CORS, recovery)
│   ├── server/            Server abstractions (HTTP, gRPC, Manager)
│   └── telemetry/         OpenTelemetry SDK initialization
├── pkg/
│   ├── database/          PostgreSQL and Redis client factories
│   └── response/          Standard JSON API response helpers
├── proto/                 Protocol Buffer definitions (gRPC)
├── configs/               YAML configuration files
├── deploy/                Docker and infrastructure files
├── migrations/            SQL migration files
├── scripts/               Utility scripts (migration runner, seed data)
├── .env.example           Environment variable template
├── Makefile               Build automation
└── README.md
```

## Quick Start

```bash
# 1. Clone
git clone https://github.com/rizky/smart-grant.git
cd smart-grant

# 2. Copy environment
cp .env.example .env

# 3. Start dependencies (PostgreSQL, Redis, OTel)
docker compose -f deploy/docker-compose.yml up postgres redis -d

# 4. Run migrations
go run ./scripts/migrate.go up

# 5. Start backend (terminal 1)
make run-backend

# 6. Start API gateway (terminal 2)
make run-api-gateway

# 7. Start worker (terminal 3)
make run-worker

# 8. Verify
curl http://localhost:8081/health
# {"status":"ok"}
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build all service binaries |
| `make run-api-gateway` | Run API Gateway locally |
| `make run-backend` | Run Backend Service locally |
| `make run-worker` | Run Worker locally |
| `make dev` | Start full stack with Docker Compose |
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Rollback database migrations |
| `make test` | Run all tests with race detection |
| `make proto` | Generate protobuf stubs |

## API Endpoints

### Auth

| Method | Path | Role | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | public | Register new user |
| POST | `/api/v1/auth/login` | public | Login |
| POST | `/api/v1/auth/refresh` | public | Refresh access token |

### Proposals

| Method | Path | Role | Description |
|--------|------|------|-------------|
| GET | `/api/v1/proposals` | all | List proposals (filtered by role) |
| POST | `/api/v1/proposals` | applicant | Create proposal |
| GET | `/api/v1/proposals/:id` | all | Get proposal detail |
| PUT | `/api/v1/proposals/:id` | applicant | Update proposal |
| POST | `/api/v1/proposals/:id/submit` | applicant | Submit for review |
| POST | `/api/v1/proposals/:id/documents` | applicant | Upload document |
| GET | `/api/v1/proposals/:id/documents` | all | List documents |

### Reviews

| Method | Path | Role | Description |
|--------|------|------|-------------|
| POST | `/api/v1/reviews/:proposal_id` | reviewer | Submit review |
| GET | `/api/v1/reviews/:proposal_id` | all | Get review |
| POST | `/api/v1/reviews/:id/approve` | admin | Approve proposal |
| POST | `/api/v1/reviews/:id/reject` | admin | Reject proposal |

### Risk Scoring

| Method | Path | Role | Description |
|--------|------|------|-------------|
| GET | `/api/v1/risk/:proposal_id` | admin, reviewer | Get risk score |

### Audit & Notifications

| Method | Path | Role | Description |
|--------|------|------|-------------|
| GET | `/api/v1/audit-logs` | admin | Query audit trail |
| GET | `/api/v1/notifications` | all | List notifications |
| GET | `/api/v1/notifications/stream` | all | SSE notification stream |

### Analytics

| Method | Path | Role | Description |
|--------|------|------|-------------|
| GET | `/api/v1/analytics/dashboard` | admin | Dashboard statistics |

## Deployment

### Docker Compose

```bash
# Deploy full stack
docker compose -f deploy/docker-compose.yml up --build -d

# Run migrations
docker compose -f deploy/docker-compose.yml run migrate
```

### Production Checklist

- [ ] Set strong `JWT_SECRET` via environment variable
- [ ] Configure PostgreSQL with proper credentials and SSL
- [ ] Use managed Redis (Upstash / ElastiCache) for production
- [ ] Set up Prometheus + Grafana dashboards
- [ ] Add nginx reverse proxy with TLS termination
- [ ] Configure database backups and point-in-time recovery
- [ ] Set rate limiting values appropriate for production traffic
- [ ] Enable structured JSON logging and log aggregation

## Tech Stack

- **Language:** Go 1.26
- **HTTP Router:** chi
- **Database:** PostgreSQL 16 (pgx v5)
- **Cache / Queue:** Redis 7 (go-redis)
- **RPC:** gRPC (with protobuf)
- **Observability:** OpenTelemetry → Prometheus + Jaeger
- **Risk Engine:** C4.5 Decision Tree (from scratch)
- **Container:** Docker + Docker Compose
- **Base Template:** [go-scaffold](https://github.com/rizky/go-scaffold)

## License

MIT
