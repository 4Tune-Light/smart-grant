# ADR-003: Use Redis Streams for async notification delivery

## Context

Sistem notifikasi perlu mengirim notifikasi ke user secara real-time. Proses ini harus decoupled dari request-response cycle agar tidak memblokir response API.

## Options Considered

**Redis Streams (XADD / XREADGROUP):**
+ Persistent queue — data tidak hilang jika consumer crash
+ Consumer group — multiple worker bisa bagi beban
+ XACK — guaranteed delivery (at-least-once)
+ XREADGROUP blocking — polling efisien
+ Ordering terjamin — FIFO per stream
- Butuh Redis terpisah (Redis sudah ada di stack)

**Redis Pub/Sub:**
+ Sederhana — publish langsung deliver ke subscriber
+ Latensi sangat rendah
- No persistence — kalau subscriber offline, pesan hilang
- No consumer group — hanya fan-out
- Tidak cocok untuk async processing

**RabbitMQ:**
+ Mature, fitur lengkap
+ AMQP standard
- Tambah infrastructure baru (maintenance overhead)
- Overkill untuk use case sederhana

**Channel in-memory:**
+ Zero infrastructure
- Data hilang jika service restart
- Tidak bisa scale horizontal

## Decision

Gunakan Redis Streams untuk queue provider karena Redis sudah ada di infrastructure stack. Worker menggunakan consumer group untuk horizontal scaling.

## Consequences

+ Redis Streams memberikan persistence + consumer groups — fitur yang cukup untuk async notification
+ Tidak perlu tambahan infrastructure (RabbitMQ, Kafka)
+ Worker bisa di-scale ke N instance tanpa double delivery
+ Notifikasi tetap tersimpan walau worker sedang sibuk
- Tambah kompleksitas di worker (perlu handle ACK, retry, dead letter)
- Masih single-node Redis (perlu Redis Cluster untuk high availability)

## Status

Accepted
