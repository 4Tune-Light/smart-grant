# ADR-005: Use cursor-based pagination (with page-based fallback)

## Context

API endpoints yang mengembalikan list data (proposals, notifications, audit logs) membutuhkan pagination. Pemilihan strategi pagination akan mempengaruhi performance, konsistensi data, dan user experience.

## Options Considered

**Cursor-based (keyset pagination):**
+ Performance O(1) — tidak bergantung pada jumlah data
+ Consistent — insert/delete di halaman awal tidak mempengaruhi halaman selanjutnya
+ Cocok untuk infinite scroll dan real-time data
- Tidak bisa lompat ke halaman arbitrary (no "go to page 5")
- Frontend perlu menyimpan cursor (tidak bisa bookmark URL dengan page number)

**Offset-based (LIMIT + OFFSET):**
+ Intuitif — user bisa `?page=3` langsung
+ SEO-friendly
+ Cocok untuk admin panel
- Performance O(n) — OFFSET besar = lambat
- Inconsistent — insert baru di halaman 1 bisa menyebabkan duplikasi di halaman 2

**Cursor-only:**
+ Performance maksimal
- Frontend kehilangan halaman manual

**Dual support (Cursor + Page):**
+ Frontend bisa pilih: infinite scroll (cursor) atau admin panel (page)
+ Backend maintain dua method
- Repository lebih kompleks — perlu dua implementasi

## Decision

Implementasikan dual support: `GET /proposals?cursor=xxx&limit=10` (cursor-based) dan `GET /proposals/page?page=1&limit=10` (page-based).

## Consequences

+ Frontend infinite scroll menggunakan cursor — consistent dan performant
+ Admin panel menggunakan page — bookmarkable dan familiar
+ Tambah complexity di repository layer — method duplikasi (ListByApplicant + ListByApplicantPage)
+ Cursor menggunakan composite key `(created_at, id)` — tiebreaker untuk created_at yang sama

## Status

Accepted
