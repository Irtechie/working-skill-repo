---
kb_id: kb-2026-07-05-model-agnostic-planner-economy
slice_id: slice-001
title: "Approve absorption spike scope and decision criteria"
blockers: []
verification: hitl
test_level: none
functional_risk: none
model_tier: large
model_tier_reason: "This is the architecture decision point: prove KB can absorb runtime state, or decide KB should become payload on a smaller runtime."
hitl: true
expected_files:
  - path: docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md
    op: edit
    scope: "record approval, edits, or rejection and unblock or redirect plan-to-work"
  - path: docs/context/decisions/2026-07-05-model-agnostic-core-vs-payload.md
    op: create
    scope: "capture the provisional architecture decision and spike pass/fail criteria"
protected_oracles: []
status: done
owner: human
blocked_reason: ""
resume_when: ""
next_agent_action: "Approved; continue with slice-002."
human_action: ""
can_continue_other_slices: true
---

# Slice 001 - Absorption Spike Scope

## What To Decide

Approve, edit, or reject a bounded spike that tests whether KB can absorb durable
task state and context packets cleanly.

This replaces the earlier bakeoff-first decision. A bakeoff against a mature
runtime is useful later only after KB has a comparable minimal state spine.

Decision: approved on 2026-07-05T13:59:39-04:00. The integrated blueprint lives
at `docs/context/decisions/2026-07-05-kb-control-plane-blueprint.md`.

## Acceptance Criteria

- The manifest records Fable's correction: HumanLayer already has durable
  session, approval, HITL, and state-machine mechanics.
- The plan chooses an absorption spike as the next decision point.
- Pass/fail criteria are explicit enough that the result cannot be decided by
  taste alone.
- `plan-to-work` is passed only after implementation is authorized.

## Pass Signals

- A structured task-state/context-packet object can be validated by `kbcheck`.
- `kb-plan` can produce packet data and `kb-work` can consume/update it.
- Recovery from blocked, interrupted, stale, or half-written state has tests.
- Telemetry records predicted tier, actual tier/model, proof, rework, and
  escalation.
- Vendor/runtime details stay outside the core slice contract.

## Fail Signals

- State updates require brittle markdown edits across multiple skills.
- Recovery relies on model judgment instead of deterministic checks.
- Claude/Codex/host assumptions leak into the core planning artifacts.
- Packet execution cannot be measured externally.
- The first adapter boundary cannot support a second adapter without redesign.

## Scope Boundary

No implementation happens in this slice. It is the human authorization gate.
This slice is complete; implementation begins with slice-002.

## Verification

Human review plus:

```shell
go run ./cmd/kbcheck manifest-contract --manifest docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md
```
