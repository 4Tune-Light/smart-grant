# Smart Grant Review Platform

> Production-like grant management platform with AI-assisted risk scoring.
> Built for CV/portfolio — demonstrates senior-level backend engineering.

[![Go Version](https://img.shields.io/badge/go-1.26-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-46%20passing-brightgreen)]()
[![Coverage](https://img.shields.io/badge/coverage-60%25-yellowgreen)]()
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.1.0-blue)](docs/openapi.yaml)

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
API Gateway (port 8081 — rate limiter, JWT passthrough, security headers)
    |
Backend Service (port 8080 HTTP + 9090 gRPC)
    |
    --- PostgreSQL 16  (source of truth)
    --- Redis 7       (cache + async queue via Streams)
    --- MinIO         (S3-compatible file storage)
    --- Worker        (async: notification delivery)
    |
OpenTelemetry Collector
    |
    --- Prometheus  (metrics, port 9090)
    --- Jaeger      (distributed tracing, port 16686)
```

## Quick Start

```bash
# 1. Clone
git clone https://github.com/rizky/smart-grant.git
cd smart-grant

# 2. Copy environment
cp .env.example .env

# 3. Start all infrastructure
docker compose -f deploy/docker-compose.yml up -d

# 4. Run database migrations
go run ./scripts/migrate.go up

# 5. Start services (3 terminals)
make run-backend       # terminal 1
make run-api-gateway   # terminal 2
make run-worker        # terminal 3

# 6. Test
curl http://localhost:8081/health
# {"status":"ok"}
```

## API Overview

| Method | Path | Role | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | public | Register |
| POST | `/api/v1/auth/login` | public | Login |
| POST | `/api/v1/auth/refresh` | public | Refresh token |
| GET | `/api/v1/users` | admin | List users |
| PATCH | `/api/v1/users/{id}/role` | admin | Update role |
| GET | `/api/v1/proposals` | all | List (cursor) |
| GET | `/api/v1/proposals/page` | all | List (page) |
| POST | `/api/v1/proposals` | applicant | Create |
| PUT | `/api/v1/proposals/{id}` | applicant | Update |
| POST | `/api/v1/proposals/{id}/submit` | applicant | Submit |
| POST | `/api/v1/proposals/{id}/documents` | applicant | Upload file |
| GET | `/api/v1/proposals/{id}/documents` | all | List files |
| GET | `/api/v1/proposals/{id}` | all | Get detail |
| POST | `/api/v1/reviews/{id}` | reviewer | Create review |
| GET | `/api/v1/reviews/{id}` | all | Get reviews |
| POST | `/api/v1/reviews/{id}/approve` | admin | Approve |
| POST | `/api/v1/reviews/{id}/reject` | admin | Reject |
| POST | `/api/v1/risk/{id}` | admin, reviewer | Score proposal |
| GET | `/api/v1/risk/{id}` | admin, reviewer | Get score |
| GET | `/api/v1/audit-logs` | admin | Audit trail |
| GET | `/api/v1/notifications` | all | My notifications |
| GET | `/api/v1/notifications/stream` | all | SSE stream |
| PATCH | `/api/v1/notifications/read?id=` | all | Mark read |

Full documentation: [docs/openapi.yaml](docs/openapi.yaml)

## Testing

```bash
# Unit tests (fast, no Docker)
make test

# Integration + E2E tests (requires Docker)
make test-integration

# Coverage report
make test-coverage

# Current: 46 tests across 6 domains
go test -short -v -count=1 ./...
```

### Test layers

| Layer | Tool | Scope |
|---|---|---|
| Unit (service) | Mock interfaces | Business logic — 34 tests |
| Unit (handler) | `httptest` | HTTP handlers — 8 tests |
| Integration | `testcontainers` | Real PostgreSQL — 3 tests |
| E2E | `testcontainers` | Full grant flow — 1 test |

## Example Flow

```bash
# Register users
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"applicant@test.com","password":"pass1234","name":"Alice","role":"applicant"}'

# Login
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"applicant@test.com","password":"pass1234"}' | jq -r '.data.access_token')

# Create proposal
curl -X POST http://localhost:8081/api/v1/proposals \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Research Grant","description":"Funding for AI research","nominal_amount":500000000,"organization":"AI Lab"}'

# Submit
curl -X POST http://localhost:8081/api/v1/proposals/{id}/submit \
  -H "Authorization: Bearer $TOKEN"
```

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP Router | chi |
| Database | PostgreSQL 16 (pgx v5) |
| Cache / Queue | Redis 7 (go-redis) |
| File Storage | MinIO (S3-compatible) |
| RPC | gRPC + protobuf |
| Auth | JWT + argon2id |
| Observability | OpenTelemetry → Prometheus + Jaeger |
| Risk Engine | C4.5 Decision Tree (from scratch) |
| Container | Docker + Docker Compose |
| Testing | testify + testcontainers |
| Base Template | [go-scaffold](https://github.com/rizky/go-scaffold) |

## Deployment

```bash
# Full stack with Docker
docker compose -f deploy/docker-compose.yml up --build -d

# Run migrations
docker compose -f deploy/docker-compose.yml run migrate
```

Production checklist:
- [ ] Set strong `JWT_SECRET` via environment variable
- [ ] Use managed PostgreSQL + Redis
- [ ] Add nginx reverse proxy with TLS
- [ ] Configure Prometheus + Grafana dashboards
- [ ] Enable structured JSON logging
- [ ] Database backups

## License

MIT
