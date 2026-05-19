# Risk Scoring Engine — C4.5 Decision Tree

## Overview

C4.5 decision tree classifier built from scratch to predict proposal risk levels: `low`, `medium`, `high`.

```
internal/risk/
├── engine/
│   ├── types.go     Data structures (Example, TreeNode, DecisionTree)
│   ├── c45.go       C4.5 algorithm (entropy, gain ratio, recursive split)
│   └── seed.go      Training data (15 synthetic examples)
├── repository.go     Persist risk scores to PostgreSQL
├── service.go        Score a proposal, retrieve scores
├── handler.go        HTTP endpoints
├── dto.go            Response types
└── errors.go         Domain errors
```

## Algorithm Flow

### 1. Training — `BuildTree()`

```
Data training (15 contoh) → Hitung entropy parent → Cari split terbaik
→ Hitung Gain Ratio tiap feature → Pilih feature dengan GR tertinggi
→ Split data → Rekursi ke child node → Stop jika murni / habis feature
```

### 2. Classification — `Classify()`

```
Features proposal → Mulai dari root → Bandingkan nilai vs threshold
→ Ke kiri (≤) atau kanan (>) → Ulangi sampai leaf node
→ Return label + confidence
```

## Features

| Feature | Type | Description |
|---|---|---|
| `nominal_amount` | continuous | Amount requested (IDR) |
| `funding_frequency_30d` | continuous | How many times org received funds in last 30 days |
| `document_completeness` | continuous | Ratio of uploaded vs required documents (0.0 - 1.0) |

## How Splits Are Found

1. Sort examples by feature value
2. Test threshold at midpoint between every adjacent pair with **different labels**
3. For each threshold, calculate:

   **Entropy**: `H(S) = -Σ p_i × log₂(p_i)`
   **Information Gain**: `IG = H(parent) - Σ (|Sv|/|S| × H(Sv))`
   **Split Info**: `SI = -Σ (|Sv|/|S| × log₂(|Sv|/|S|))`
   **Gain Ratio**: `GR = IG / SI`

4. Pick threshold + feature with highest Gain Ratio

## Example Tree (from current seed data)

```
Root: nominal_amount ≤ 1.25B ?
├── YES:
│   ├── document_completeness ≤ 0.45 ? → HIGH
│   └── NO: frequency ≤ 2.5 ?
│       ├── YES → LOW
│       └── NO  → MEDIUM
└── NO:
    ├── frequency ≤ 3.5 ? → MEDIUM
    └── NO  → HIGH
```

## API

| Method | Path | Role | Description |
|---|---|---|---|
| POST | `/api/v1/risk/{id}` | admin, reviewer | Score a proposal |
| GET | `/api/v1/risk/{id}` | admin, reviewer | Get existing score |

## Limitations & Planned Improvements

| Issue | Current | Improvement |
|---|---|---|
| Training data | 15 synthetic examples | Add real/augmented data |
| Features | 3 features | Add `proposal_similarity`, `org_history_score` |
| `frequency` | Hardcoded 0.0 | Query from DB: `COUNT(*) WHERE organization = X AND created_at > 30d` |
| `completeness` | Hardcoded 1.0 | Compute from actual documents |
| Pruning | None | Reduced error pruning to prevent overfitting |
| Ensemble | Single tree | Random Forest for higher accuracy |
| Async scoring | Synchronous | Trigger scoring via Redis Streams on proposal submit |

## References

- Quinlan, J. R. (1993). *C4.5: Programs for Machine Learning*
- [Information Gain Ratio explanation](https://en.wikipedia.org/wiki/Information_gain_ratio)
