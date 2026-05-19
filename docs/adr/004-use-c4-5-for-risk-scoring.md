# ADR-004: Use C4.5 decision tree for risk scoring

## Context

Sistem perlu menilai tingkat risiko proposal (low/medium/high) berdasarkan fitur seperti nominal amount, frekuensi pengajuan, dan kelengkapan dokumen. Algoritma harus bisa dijelaskan (explainable) dan diimplementasikan dari nol.

## Options Considered

**C4.5 Decision Tree:**
+ Explainable — keputusan bisa dilacak: "amount > 1.25M → frequency > 3 → HIGH"
+ Implementasi dari nol — menunjukkan pemahaman ML algorithms
+ Gain ratio — menormalisasi bias informasi gain ke feature dengan banyak nilai
+ Continuous feature split otomatis
+ Tidak perlu external library
+ Cocok untuk tabular data dengan fitur terbatas
- Kurang akurat dibanding ensemble methods
- Overfitting jika tidak di-prune

**Random Forest:**
+ Akurasi lebih tinggi (ensemble of trees)
+ Lebih robust terhadap overfitting
- Implementasi lebih kompleks
- Sulit dijelaskan (black box)

**Logistic Regression:**
+ Sederhana, cepat, terdefinisi
- Asumsi linearitas
- Kurang cocok untuk data non-linear

**External API (OpenAI, Claude):**
+ Akurasi tinggi
+ Bisa analisis teks proposal
- Biaya per request
- Ketergantungan pada third-party
- Latency tinggi
- Data privacy concerns

**Rule-based engine:**
+ Paling sederhana: if-else chain
+ Deterministik
- Tidak bisa belajar dari data — harus maintain rules manual

## Decision

Gunakan C4.5 decision tree dari nol. Algoritma diimplementasikan sebagai library internal (`internal/risk/engine/`) untuk menunjukkan ML understanding.

## Consequences

+ Portfolio value tinggi — implementasi decision tree dari nol
+ Model bisa di-retrain dari data real
+ Kelemahan akurasi bisa di-addressed dengan ensemble (improvement roadmap)
+ Training data saat ini masih 50 contoh — perlu diperbanyak

## Status

Accepted
