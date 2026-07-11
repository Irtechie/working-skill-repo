---
kb_id: kb-2026-07-10-session-model-routing
slice_id: slice-006
title: "Prove the Codex-first advisory pilot and promotion boundary"
blockers: [slice-004, slice-005b]
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
    scope: "deterministic no-paid planned-tier and exploratory next-lower conformance corpus; never live or efficiency evidence"
  - path: evals/model-routing/correction-pilot.json
    op: create
    scope: "protected localized seeded-fault stratum for correction safety, excluded from efficiency benefit"
  - path: docs/results/2026-07-10-session-model-routing-initial-pilot.json
    op: create
    scope: "machine-readable deterministic/live/install evidence and advisory support matrix"
  - path: docs/context/eval-map.md
    op: edit
    scope: "register routing pilot and promotion proof commands"
protected_oracles:
  - path: cmd/kbcheck/model_routing_release_test.go
    role: "support cohort, baseline, and overclaim oracle"
    sha256: "8a63ed271f8817d3efc13a7bc5f64e887e9dea024602d2080116203fe5883f45"
    update_policy: "requires explicit plan update"
status: done
owner: agent
can_continue_other_slices: false
---

# Model Routing Advisory Pilot

## What To Build

Add a release validator and evidence artifact that distinguish deterministic
conformance, attended live canaries, supported cohorts, and parked claims. A
successful validator exit means the artifact is internally honest and preserves
the safety defaults; it does not by itself mean a cohort or AMR is promoted.

## Acceptance Criteria

- The no-paid artifact may finish successfully as `not-promoted`, with zero
  supported cohorts and no live savings claim. Codex CLI plus one
  OpenAI-compatible/LiteLLM route becomes supported only after attended,
  route-bound receipts link the exact packet, work proof, and install proof.
- Current-model-only/missing-router behavior passes without claiming multi-model dispatch.
- Planned-tier host-native selection remains the zero-setup baseline. Next-lower attempts remain disabled by default unless the evidence meets the preregistered promotion contract.
- A deterministic exploratory corpus proves attempt/handoff mechanics,
  correction-dispatch refusal, and ineligible ordinary fallback without being
  mislabeled as live savings evidence.
- Promotion requires an independently held-out, power-justified live corpus and correction stratum. Missing comparable cost/latency or insufficient samples produces `not-promoted`, never a fabricated material benefit.
- Missing usage/cost/model/session telemetry is `unavailable`, never zero or inferred.
- GHCP, exact Codex App attribution, TinyBoss control, MCP/direct chat-completions dispatch, and public-default next-lower attempts remain unsupported until separate evidence exists.

## Test Scenarios

- Deterministic fake-host corpus for small/medium/large, fallbacks, trust denial,
  mismatch, timeout, partial handoff, localized proof failure, fail-closed
  correction dispatch, no-oracle ineligibility, and ordinary planned-tier fallback.
- An honest no-paid artifact records the attended native Codex and configured
  LiteLLM/OpenAI-compatible-via-Codex canaries as not run and not qualified.
  Running those canaries is a later HITL promotion activity, not a requirement
  for this `hitl: false` validator slice.
- Release evidence fails on prose-only proof, stale/mismatched receipt, missing install proof, regression, or forbidden supported claim.

## Tier Rationale

Large: live model execution, capability claims, baseline interpretation, and release gating require strongest synthesis.

## Scope Boundary

No next-lower default promotion from deterministic fixtures, insufficient live samples, or incomparable metrics; no unsupported host label.
