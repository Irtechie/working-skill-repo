---
kb_id: kb-2026-07-09-phoenix-routing-slicing-absorption
slice_id: slice-002
title: "Add manifest done_check and proof_check schema enforcement"
blockers: [slice-001]
verification: tdd
test_level: unit
functional_risk: none
model_tier: medium
model_route: hosted-sonnet
hitl: false
expected_files:
  - path: "cmd/kbcheck/manifest_contract.go"
    op: edit
    scope: "Parse and validate top-level done_check, per-slice proof_check, and model_route."
  - path: "cmd/kbcheck/swarm.go"
    op: edit
    scope: "Extend manifestSlice parsing for proof_check and model_route if shared parser owns slice metadata."
  - path: "cmd/kbcheck/manifest_contract_test.go"
    op: edit
    scope: "Add RED/GREEN tests for missing done_check, missing proof_check, invalid model_route, and explicit no-check exceptions."
protected_oracles:
  - path: "cmd/kbcheck/manifest_contract_test.go"
    role: "schema enforcement oracle"
    sha256: "2e28ab6a67da6e0ae9d553ad1387a0f7d9bed477da90554fee703c30e58f543a"
    update_policy: "intentional oracle changes only"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Write failing validator tests first, then extend the parser/validator until the new schema passes."
human_action: ""
can_continue_other_slices: true
---

# Slice 002: Add Manifest Done/Proof Schema Enforcement

## What This Delivers

Phoenix's strongest routing/slicing contribution becomes a KB-native manifest
contract: every long-running goal has a top-level `done_check`, every runnable
slice has a `proof_check` or explicit no-check reason, and every model-tiered
slice has a valid `model_route`.

## Acceptance Criteria

- `cmd/kbcheck` rejects a manifest that opts into the Phoenix-style contract but
  lacks a top-level `done_check`.
- `cmd/kbcheck` rejects a runnable slice that lacks `proof_check` unless the
  slice records an explicit no-check exception and verification mode justifies
  it.
- `cmd/kbcheck` validates `model_route` against the manifest's route contract.
- Existing manifests without the opt-in contract do not break unless they are
  marked active for this new contract.

## Test Scenarios

- Missing `done_check` fixture fails.
- Missing per-slice `proof_check` fixture fails.
- Invalid `model_route` fixture fails.
- Valid manifest with `done_check`, `proof_check`, and route contract passes.

## Scope Boundary

This slice only adds code-level schema enforcement. Skill text and README wiring
come in later slices.
