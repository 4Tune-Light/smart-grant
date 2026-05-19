# ADR-008: Use OpenTelemetry for observability

## Context

Aplikasi multi-service (gateway, backend, worker) membutuhkan observability untuk debugging, monitoring, dan performance analysis. Tools harus bisa meng-handle distributed tracing antar service dan metrics collection.

## Options Considered

**OpenTelemetry (OTel):**
+ Standard industri — adopted by Google, Microsoft, AWS
+ Traces + Metrics + Logs dalam satu SDK
+ Export ke berbagai backend (Prometheus, Jaeger, Datadog, NewRelic)
+ Propagation context otomatis antar service (HTTP headers, gRPC metadata)
+ Go SDK mature
- Menambah dependensi (~10 packages)
- Complexity setup collector

**Prometheus client + manual tracing:**
+ Ringan, zero dependency selain promhttp
- Tidak ada distributed tracing
- Harus propagate trace context manual
- Tidak ada standard instrumentation

**Datadog / NewRelic agent:**
+ Managed — zero code integration
+ Lengkap (APM, tracing, metrics)
- Biaya lisensi
- Vendor lock-in
- Tidak bisa development tanpa akun

**Zerolog structured logging only:**
+ Ringan, zero dependency tambahan
- Tidak ada tracing
- Debugging membutuhkan correlation ID manual

## Decision

Gunakan OpenTelemetry SDK untuk tracing dan metrics. OTel Collector sebagai gateway ke Prometheus (metrics) dan Jaeger (traces).

## Consequences

+ Distributed tracing dari Gateway ke Backend ke Worker — trace di Jaeger UI
+ Metrics (request count, duration, error rate) di Prometheus
+ Custom spans di business logic (risk.Score, review.Create, proposal.Submit)
+ Setup collector + prometheus + jaeger di docker-compose
+ Tambah dependency: OTel SDK, OTel HTTP middleware, OTel gRPC interceptor

## Status

Accepted
