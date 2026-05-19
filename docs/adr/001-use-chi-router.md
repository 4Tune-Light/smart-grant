# ADR-001: Use chi router

## Context

Smart Grant Review Platform membutuhkan HTTP router untuk backend service. Pemilihan router akan mempengaruhi arsitektur middleware, testing strategy, dan maintainability kode.

## Options Considered

**chi (github.com/go-chi/chi):**
+ Idiomatic Go — kompatibel dengan `net/http` dan `http.Handler`
+ Middleware adalah `func(http.Handler) http.Handler` — reusable tanpa vendor lock-in
+ Ringan, zero reflection, zero magic
+ Route grouping dan middleware scoping
+ Maintained dan stable
- Kurang built-in validator/logger dibanding gin

**gin:**
+ Performa tinggi
+ Built-in binding dan validation
- Proprietary `gin.Context` — lock-in ke gin
- Reflection-based, error handling verbose
- Tidak kompatibel dengan standard library

**echo:**
+ Mirip gin, minimalis
+ Built-in middleware
- Proprietary context
- Komunitas lebih kecil dari chi/gin

**std net/http (mux):**
+ Zero dependency
- Routing manual — tidak praktis untuk production

**fiber:**
+ Performa sangat tinggi (fasthttp)
- API mirip Express.js
- Tidak kompatibel dengan `http.Handler` — gabisa pake middleware standar

## Decision

Gunakan chi karena idiomatic Go, kompatibel dengan standard library, dan zero lock-in.

## Consequences

+ Middleware bisa di-sharing dengan library Go lain (`otelhttp`, `promhttp`)
+ Setiap handler tetap `http.Handler` — mudah di-test dengan `httptest`
+ Mudah diganti framework jika nanti dibutuhkan (tidak ada proprietary code)
- Developer harus setup validation dan logger secara manual (tapi ini sengaja untuk flexibility)

## Status

Accepted
