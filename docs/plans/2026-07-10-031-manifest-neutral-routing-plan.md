---
kb_id: kb-2026-07-10-session-model-routing
slice_id: slice-001
title: "Keep manifests model-neutral with legacy compatibility"
blockers: []
verification: tdd
test_level: unit
functional_risk: narrow
model_tier: medium
context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-001.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run ModelRoute"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "accept route-free new manifests and treat legacy model_route only as a nonbinding hint"
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: "protect new/legacy compatibility and reject route-as-completion claims"
  - path: cmd/kbcheck/swarm.go
    op: edit
    scope: "preserve legacy parsing without requiring route ownership in ready-set state"
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "freeze model tier, risk, context, and proof only"
protected_oracles:
  - path: cmd/kbcheck/manifest_contract_test.go
    role: "new/legacy manifest compatibility oracle"
    sha256: "cbfd6b3265e47b1bba0d97e0590780c8ebbc6978d62041d6e75021536e338b4b"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
can_continue_other_slices: true
---

# Model-Neutral Manifest Migration

## What To Build

Migrate manifests from planned model routes to planned task difficulty while reading old `model_route` values as advisory history.

## Acceptance Criteria

- A new objective manifest with `model_tier` and no `model_route` passes.
- A legacy manifest with a valid or stale `model_route` remains readable without freezing work-time selection.
- Templates and prose no longer require routes in plans; route evidence belongs to run receipts.
- Proof gates remain identical regardless of worker tier or route.

## Test Scenarios

- Prove RED on a route-free objective manifest under the current validator, then GREEN.
- Run legacy valid, legacy stale, and new route-free fixtures.
- Assert a routing receipt cannot substitute for a slice `proof_check`.

## Tier Rationale

Medium: bounded cross-file contract migration with clear fixtures; escalate if compatibility touches broader gate semantics.

## Scope Boundary

No catalog, selection, dispatch, installer, or public support implementation.
