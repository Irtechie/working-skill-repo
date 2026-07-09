---
kb_id: kb-2026-07-09-phoenix-routing-slicing-absorption
slice_id: slice-001
title: "Fix snapshot path drift and gate health"
blockers: []
verification: integration
test_level: integration
functional_risk: none
model_tier: small
model_route: local-5090-coder
hitl: false
expected_files:
  - path: ".github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1"
    op: edit
    scope: "Change default snapshot dir from .atv/snapshots to .kb/snapshots."
  - path: "cmd/kbcheck"
    op: edit
    scope: "Only if needed after diagnosing why go run ./cmd/kbcheck core --list hung in this shell."
  - path: "docs/context/operations/testing.md"
    op: edit
    scope: "Record any gate-health caveat or fixed command behavior."
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Start by fixing the snapshot path mismatch, then diagnose the Go gate hang before expanding the harness."
human_action: ""
can_continue_other_slices: false
---

# Slice 001: Fix Snapshot Path Drift And Gate Health

## What This Delivers

The existing KB proof spine is not expanded until its current maintenance gate
is trustworthy and the remaining `.atv/snapshots` default is removed.

## Acceptance Criteria

- `kb-regression-snapshot.ps1` defaults to `.kb/snapshots`.
- README/skills/scripts agree on the snapshot path.
- The `go run ./cmd/kbcheck core --list` hang observed on July 9 is either
  fixed or documented with a narrower reliable proof command.
- `git diff --check` passes.

## Test Scenarios

- Run a text search proving no active runtime snapshot default points to
  `.atv/snapshots`.
- Run the narrowest available Go proof command. Prefer
  `go test ./cmd/kbcheck -run TestProof -count=1`; then re-try
  `go run ./cmd/kbcheck core --list` if the hang is believed fixed.

## Scope Boundary

No new `done_check`, run-state, or route-history fields in this slice.
