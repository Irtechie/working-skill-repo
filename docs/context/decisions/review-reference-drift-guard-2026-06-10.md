# Review Reference Drift Guard

Date: 2026-06-10

## Decision

Review-skill reference duplication must be explicit in `config/skill-quality.json`.
Configured review reference roots are swept for common filenames. Any common
filename pair must be classified as either:

- `shared_pairs` - the files must hash-match.
- `intentional_forks` - the files may differ but must carry owner and reason.

## Reason

`ce-review`, `kb-review`, and `document-review` duplicate some reference names
for portability and because their public lanes differ. An allowlist-only guard
protects known files but misses newly duplicated filenames. A root sweep turns
the guard into a closed contract.

## Consequences

- Adding the same new reference filename under two configured review roots fails
  `kbcheck review-reference-guard` until classified.
- `document-review` overlaps are first-class review-family entries, not
  undocumented exceptions.
- `go run ./cmd/kbcheck core` runs the guard.

## Related

- Compound note: `docs/solutions/workflow-issues/review-reference-drift-guard-2026-06-10.md`
- Config: `config/skill-quality.json`
- Guard: `cmd/kbcheck/review_reference_guard.go`
