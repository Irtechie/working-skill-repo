---
kb_id: kb-2026-07-05-model-agnostic-planner-economy
slice_id: slice-004
title: "Prove telemetry normalization and provider hygiene"
blockers: [slice-002, slice-003]
verification: integration
test_level: integration
functional_risk: narrow
model_tier: large
model_tier_reason: "This turns token/provider policy into deterministic, cross-runtime checks."
hitl: false
expected_files:
  - path: cmd/kbcheck/context_packet.go
    op: edit
    scope: "add normalized telemetry validation"
  - path: cmd/kbcheck/context_packet_test.go
    op: edit
    scope: "cover optional/invalid usage fields and proof outcomes"
  - path: cmd/kbcheck/provider_hygiene.go
    op: create
    scope: "detect Phoenix provider activation"
  - path: cmd/kbcheck/provider_hygiene_test.go
    op: create
    scope: "cover forbidden Phoenix, unrelated providers, and disabled configurations"
  - path: docs/context/eval-map.md
    op: edit
    scope: "record the spike's deterministic proof surfaces"
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: "slice-003 done"
next_agent_action: "Add recovery and telemetry fixtures before claiming self-healing."
human_action: ""
can_continue_other_slices: true
notes: "Added normalized optional usage fields and provider-hygiene checks. Phoenix activation fails."
---

# Slice 004 - Recovery and Telemetry Proof

## What To Build

Add deterministic evidence that usage telemetry is comparable across runtimes
and provider integrations remain optional.

## Acceptance Criteria

- Telemetry captures predicted tier, actual tier/model when available, proof
  result, rework count, escalation, packet sufficiency, turns, input/output
  tokens, and cache read/write tokens when the host exposes them.
- Raw usage fields remain the source of truth. Any weighted effective-token
  score is versioned and reported beside correctness/proof outcomes.
- Telemetry is a separate runtime-result artifact linked by `packet_id`, not
  mutable packet input or model-authored usage. Host extraction remains
  adapter-specific until Codex/GHCP expose stable measured output.
- Phoenix activation in repo or standard user config fails provider hygiene.
- Unrelated provider configuration does not fail the Phoenix-specific audit.

## Scope Boundary

Do not build a UI, daemon, or provider proxy.

## Verification

Run:

```shell
go test ./cmd/kbcheck/...
go run ./cmd/kbcheck ready-set --manifest docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md --json
go run ./cmd/kbcheck core
```
