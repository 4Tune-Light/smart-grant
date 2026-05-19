# ADR-002: Use argon2id over bcrypt for password hashing

## Context

Sistem autentikasi membutuhkan password hashing algorithm. Pilihan algorithm akan mempengaruhi keamanan terhadap serangan GPU/ASIC dan waktu respons API login.

## Options Considered

**argon2id (golang.org/x/crypto/argon2):**
+ Memory-hard — tidak bisa dipercepat dengan GPU/ASIC
+ Pemenang Password Hashing Competition (2015)
+ Konfigurable: time, memory, parallelism
+ Parameter disimpan di hash string — upgrade parameter tanpa migrasi data
- Lebih lambat dari bcrypt (wajar, ini justru yang diinginkan)

**bcrypt:**
+ Sangat mature — sudah ada sejak 1999
+ Sederhana — satu parameter (cost factor)
+ Library built-in di Go (`golang.org/x/crypto/bcrypt`)
- Tidak memory-hard — bisa dipercepat dengan GPU
- Cost factor maksimal 31 — terbatas

**scrypt:**
+ Memory-hard
- Standarisasi kurang dibanding argon2
- Parameter lebih kompleks

**PBKDF2:**
+ Paling tua (2000), standar NIST
+ FIPS 140 compliant
- Tidak memory-hard

**MD5/SHA1:**
- Tidak ada salt built-in
- Sangat cepat — rentan brute force
- Tidak acceptable untuk production

## Decision

Gunakan argon2id dengan parameter: time=1, memory=64MB, parallelism=4, key length=32 byte.

## Consequences

+ Password hash aman terhadap GPU brute force (membutuhkan ~64MB per attempt)
+ Parameter dikodekan dalam hash format standard (`$argon2id$v=19$m=65536,t=1,p=4$...`)
+ Upgrade parameter tanpa migrasi data — cukup verifikasi hash lama, re-hash dengan parameter baru
- Login 5x lebih lambat dari bcrypt (~500ms vs ~100ms) — tradeoff acceptability untuk keamanan

## Status

Accepted
