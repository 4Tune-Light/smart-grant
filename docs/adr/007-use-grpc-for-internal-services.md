# ADR-007: Use gRPC for internal service communication

## Context

Platform memiliki service backend (REST + gRPC) dan worker (notification consumer). Service perlu berkomunikasi secara internal dengan contract yang ketat.

## Options Considered

**gRPC + protobuf:**
+ Contract-driven — .proto file sebagai source of truth
+ Code generation — tidak ada human error di serialization/deserialization
+ Performa tinggi (binary protocol, HTTP/2)
+ Streaming built-in
+ Strongly typed — compile-time safety
- Butuh protoc + code generation setup
- Browser tidak bisa consume gRPC langsung
- Lebih kompleks dari REST

**REST internal:**
+ Sederhana, familiar
+ JSON — bisa di-debug langsung
- No contract enforcement — request/response bisa berubah tanpa disadari
- Performa lebih rendah (text protocol)
- No built-in streaming

**GraphQL:**
+ Flexible query
- Overkill untuk internal microservice
- Kompleksitas tambahan

**Message Queue (RPC-style):**
+ Async by default
- Latency lebih tinggi
- Debug lebih sulit

## Decision

Gunakan gRPC untuk internal services (RiskService, NotificationService). REST tetap menjadi public API untuk frontend.

## Consequences

+ Service interface strict — perubahan contract terdeteksi di compile-time
+ gRPC service bisa dipanggil oleh klien dari bahasa lain (Python, Java)
+ Digunakan untuk RiskService (CalculateRisk) dan NotificationService (SendNotification)
- Code generation dari proto — perlu `make proto` di development workflow
- Saat ini gRPC belum benar-benar digunakan oleh internal caller (masih via direct Go function call) — ini adalah improvement yang bisa dilakukan di masa depan

## Status

Accepted
