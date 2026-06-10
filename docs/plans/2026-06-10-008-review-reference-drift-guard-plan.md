---
kb_id: kb-2026-06-10-review-reference-drift-guard
slice_id: slice-001
title: "Protect shared review-skill references from silent drift"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: config/skill-quality.json
    op: edit
    scope: "Declare shared review-reference pairs and intentional forks."
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "Wire the review-reference guard into the core gate."
  - path: cmd/kbcheck/review_reference_guard.go
    op: create
    scope: "Validate declared shared references hash-match and intentional forks remain documented."
  - path: cmd/kbcheck/review_reference_guard_test.go
    op: create
    scope: "Prove pass/fail behavior for shared-match and fork-documentation cases."
  - path: docs/context/memory-maintenance.md
    op: edit
    scope: "Record the review-skill fork ownership and drift policy."
  - path: docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md
    op: edit
    scope: "Update slice and gate status after proof."
  - path: docs/plans/2026-06-10-008-review-reference-drift-guard-plan.md
    op: edit
    scope: "Update slice status after proof."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Implement the guard and run focused Go/CLI proof."
human_action: ""
can_continue_other_slices: true
---

# Slice 001: Protect shared review-skill references from silent drift

## What To Build

Add a deterministic `kbcheck` guard for the `ce-review` / `kb-review` shared-reference contract:

- Declared shared reference files must hash-match.
- Declared intentional forks must carry owner and reason metadata.
- The guard must run in `go run ./cmd/kbcheck core`.
- Durable memory must name why the two public review skills stay separate.

## Acceptance Criteria

- `diff-scope.md`, `resolve-base.sh`, and any other declared shared review references are protected by a hash drift check.
- Intentional `ce-review` / `kb-review` differences are recorded with owner and reason.
- The quality gate catches accidental divergence in declared shared references.
- `kb-review` remains the KB completion review lane.
- `ce-review` remains the generalized CE review lane.

## Expected Files

- `config/skill-quality.json`
- `cmd/kbcheck/checks.go`
- `cmd/kbcheck/review_reference_guard.go`
- `cmd/kbcheck/review_reference_guard_test.go`
- `docs/context/memory-maintenance.md`
- `docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md`
- `docs/plans/2026-06-10-008-review-reference-drift-guard-plan.md`

## Test Scenarios

- Unit-level Go test proves matching shared references pass.
- Unit-level Go test proves divergent shared references fail.
- Unit-level Go test proves an intentional fork without owner/reason fails.
- CLI proof: `go test ./cmd/kbcheck`.
- CLI proof: `go run ./cmd/kbcheck core`.
- CLI proof: `go run ./cmd/kbcheck local-release`.

## Scope Boundary

Do not delete, merge, or centralize review skills or agents in this slice. Do not add broad content-similarity duplicate detection here; that is a separate minimality enhancement.

## Dependencies

None.

## Completion Proof

- Added `kbcheck review-reference-guard` and wired it into `core`.
- Declared shared review-reference pairs and intentional forks in `config/skill-quality.json`.
- Added Go tests for matching shared references, shared drift rejection, and missing fork metadata rejection.
- Proof passed: `go test ./cmd/kbcheck`, `go run ./cmd/kbcheck review-reference-guard`, `go run ./cmd/kbcheck core`, `go run ./cmd/kbcheck local-release`.
