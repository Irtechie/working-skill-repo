# Review Reference Drift Guard

Date: 2026-06-10

## Problem

`ce-review` and `kb-review` intentionally keep separate public entry points, but
some reference files are shared doctrine duplicated on disk. Marking the
review-skill canonicalization slice done without a drift guard let byte-identical
shared references silently diverge later.

## Solution

Declare the relationship in `config/skill-quality.json`:

- `shared_pairs` must hash-match.
- `intentional_forks` must carry `owner` and `reason`.

Enforce that contract with `kbcheck review-reference-guard`, and include it in
`go run ./cmd/kbcheck core`.

## Proof

Commands:

```shell
go test ./cmd/kbcheck
go run ./cmd/kbcheck review-reference-guard
go run ./cmd/kbcheck core
go run ./cmd/kbcheck local-release
```

## Reuse

When a portable skill bundle duplicates files for install portability, make the
duplication explicit in config and fail the contributor gate on accidental drift.
