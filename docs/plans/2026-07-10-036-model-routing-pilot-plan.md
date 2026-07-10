---
kb_id: kb-2026-07-10-session-model-routing
slice_id: slice-006
title: "Prove the Codex-first advisory pilot and promotion boundary"
blockers: [slice-004, slice-005]
verification: functional
test_level: full
functional_risk: full
model_tier: large
context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-006.json
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/model_routing_release.go
    op: create
    scope: "validate support cohort, live receipts, proof binding, baseline, and forbidden claims"
  - path: cmd/kbcheck/model_routing_release_test.go
    op: create
    scope: "initial-pilot pass/fail and overclaim fixtures"
  - path: evals/model-routing/initial-pilot.json
    op: create
    scope: "representative difficulty, fallback, trust, and current-model baseline corpus"
  - path: docs/results/2026-07-10-session-model-routing-initial-pilot.json
    op: create
    scope: "machine-readable deterministic/live/install evidence and advisory support matrix"
  - path: docs/context/eval-map.md
    op: edit
    scope: "register routing pilot and promotion proof commands"
protected_oracles:
  - path: cmd/kbcheck/model_routing_release_test.go
    role: "support cohort, baseline, and overclaim oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
can_continue_other_slices: false
---

# Model Routing Advisory Pilot

## What To Build

Add a release validator and evidence artifact that distinguish deterministic conformance, attended live canaries, supported cohorts, and parked claims.

## Acceptance Criteria

- Initial-pilot passes only with Codex native and one Codex-hosted OpenAI-compatible route receipt bound to packet and work proof.
- Current-model-only/missing-router behavior passes without claiming multi-model dispatch.
- Baseline shows no right-to-wrong proof regressions or increases in repeat work/interventions and records one material benefit.
- Missing usage/cost/model/session telemetry is `unavailable`, never zero or inferred.
- GHCP, TinyBoss control, MCP dispatch, and default automatic routing remain unsupported until separate evidence exists.

## Test Scenarios

- Deterministic fake-host corpus for small/medium/large, fallbacks, trust denial, mismatch, timeout, and partial handoff.
- Attended native Codex canary and configured LiteLLM/OpenAI-compatible-via-Codex canary in isolated scratch worktrees.
- Release evidence fails on prose-only proof, stale/mismatched receipt, missing install proof, regression, or forbidden supported claim.

## Tier Rationale

Large: live model execution, capability claims, baseline interpretation, and release gating require strongest synthesis.

## Scope Boundary

No automatic default promotion unless this exact evidence passes; no unsupported host label.
