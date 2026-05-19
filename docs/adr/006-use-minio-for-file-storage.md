# ADR-006: Use MinIO for file storage

## Context

Sistem proposal menerima upload dokumen dari applicant (PDF, spreadsheet, dll). File storage perlu reliable, scalable, dan aman untuk production.

## Options Considered

**MinIO (S3-compatible):**
+ S3 API compatible — bisa diganti ke AWS S3, GCS, atau DigitalOcean Spaces tanpa kode berubah
+ Bisa di-self-host di Docker — cocok untuk development
+ Performa tinggi
+ Versioning, retention policy
+ Immutable object storage
- Infrastructure tambahan (tapi ringan — binary single file)

**Local filesystem (./uploads/):**
+ Sederhana — zero dependency
- Tidak bisa scale horizontal (file tersimpan di satu server)
- Backup lebih kompleks
- Tidak cocok untuk production

**Database BLOB:**
+ Data terkonsolidasi di database
+ Backup otomatis (barebone database backup)
- Performa buruk untuk file besar
- Membuat database backup menjadi besar
- Tidak bisa serve file langsung dari database

**Third-party (AWS S3, GCS):**
+ Managed, reliable, scalable
- Biaya bulanan
- Butuh koneksi internet
- Tidak bisa development tanpa koneksi

## Decision

Gunakan MinIO, di-deploy sebagai container Docker. Semua operasi file melalui `pkg/storage.MinioStorage` yang mengimplementasikan interface `FileStorage`.

## Consequences

+ S3 API compatibility — migration ke cloud storage adalah opsional, bukan revolusi kode
+ File disimpan sebagai object dengan path `proposals/{proposal_id}/{uuid}.{ext}` — no collision
+ Nama file asli disimpan di metadata database
+ `pkg/storage/` bisa diganti implementasinya tanpa mengubah kode di service layer
- MinioStorage perlu diinject dari main.go — butuh config

## Status

Accepted
