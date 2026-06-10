---
kb_id: kb-2026-06-10-review-reference-closed-contract
slice_id: slice-001
title: "Close review-reference guard contract"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: cmd/kbcheck/review_reference_guard.go
    op: edit
    scope: "Add root sweep mode that fails unclassified common reference filenames."
  - path: cmd/kbcheck/review_reference_guard_test.go
    op: edit
    scope: "Prove unclassified common references fail and classified common references pass."
  - path: config/skill-quality.json
    op: edit
    scope: "Declare review reference roots and document-review classifications."
  - path: docs/context/decisions/README.md
    op: edit
    scope: "Index durable decision files."
  - path: docs/context/decisions/contributor-core-vs-release-sync-gates-2026-06-10.md
    op: create
    scope: "File the contributor/release gate decision in the decisions tree."
  - path: docs/context/decisions/review-reference-drift-guard-2026-06-10.md
    op: create
    scope: "File the review-reference guard decision in the decisions tree."
  - path: docs/context/memory-maintenance.md
    op: edit
    scope: "Record the closed-contract follow-up."
  - path: docs/plans/2026-06-10-009-kb-review-reference-closed-contract-manifest.md
    op: edit
    scope: "Update status and gates after proof."
  - path: docs/plans/2026-06-10-010-review-reference-closed-contract-plan.md
    op: edit
    scope: "Update status after proof."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Implement discovery-based guard and run CLI proof."
human_action: ""
can_continue_other_slices: true
---

# Slice 001: Close Review-Reference Guard Contract

## What To Build

Convert the review-reference guard from a declared-pair watchlist into a closed contract:

- Sweep configured reference roots for common filenames.
- Fail any common filename pair not classified as shared or intentionally forked.
- Include `document-review` as a configured third review-reference root.
- Add durable decision docs under `docs/context/decisions/` while keeping compound notes under `docs/solutions/`.

## Acceptance Criteria

- Adding the same new filename under two configured review reference roots fails `kbcheck review-reference-guard` unless classified.
- Existing `ce-review` / `kb-review` common files remain classified.
- `document-review` overlaps are classified or explicitly exempted.
- `go run ./cmd/kbcheck core` includes and passes the closed-contract guard.
- Durable decisions are discoverable under `docs/context/decisions/`.

## Test Scenarios

- Unit-level Go test: classified common reference passes sweep mode.
- Unit-level Go test: unclassified common reference fails sweep mode.
- CLI proof: `go test ./cmd/kbcheck`.
- CLI proof: `go run ./cmd/kbcheck review-reference-guard`.
- CLI proof: `go run ./cmd/kbcheck core`.
- CLI proof: `go run ./cmd/kbcheck local-release`.

## Scope Boundary

Do not change review skill content. Do not add fuzzy similarity detection. Do not make agent deletion decisions.

## Completion Proof

- Added root sweep mode to `kbcheck review-reference-guard`.
- Configured `ce-review`, `kb-review`, and `document-review` reference roots.
- Classified `document-review` overlaps as intentional forks with owner/reason metadata.
- Added durable decision docs under `docs/context/decisions/`.
- Proof passed: `go test ./cmd/kbcheck`, `go run ./cmd/kbcheck review-reference-guard`, `go run ./cmd/kbcheck core`, `go run ./cmd/kbcheck local-release`, `git diff --check`.
