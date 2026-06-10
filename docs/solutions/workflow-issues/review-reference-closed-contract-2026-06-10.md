# Review Reference Closed Contract

Date: 2026-06-10

## Problem

A pairwise review-reference guard protects known files but misses newly duplicated
filenames. If a future edit adds `new-rubric.md` under both `ce-review` and
`kb-review`, an enumeration-only guard never sees it.

## Solution

Configure review reference roots and sweep them for common relative filenames.
Every common file pair must be classified:

- `shared_pairs` must hash-match.
- `intentional_forks` must carry owner and reason.

`document-review` is part of the same review-reference family, so its overlapping
reference names are classified instead of treated as accidental exceptions.

## Proof

Commands:

```shell
go test ./cmd/kbcheck
go run ./cmd/kbcheck review-reference-guard
go run ./cmd/kbcheck core
go run ./cmd/kbcheck local-release
```

## Reuse

When a guard protects duplicated portable assets, do not only enumerate known
pairs. Also sweep configured roots so new duplication cannot appear unclassified.
