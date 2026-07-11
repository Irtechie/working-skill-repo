---
kb_id: kb-2026-07-05-model-agnostic-planner-economy
slice_id: slice-006
title: "Write KB-core vs KB-payload decision report"
blockers: [slice-003, slice-004, slice-005]
verification: verification-only
test_level: none
functional_risk: none
model_tier: large
model_tier_reason: "This is the evidence synthesis step after the spike; it should not be delegated to a cheap worker."
hitl: false
expected_files:
  - path: docs/context/decisions/2026-07-05-model-agnostic-core-vs-payload.md
    op: edit
    scope: "record spike evidence, recommendation, and rejected paths"
  - path: docs/context/research/2026-07-05-humanlayer-pinned-repos-planner-economy.md
    op: edit
    scope: "refresh stale claims after the spike evidence exists"
  - path: todo.md
    op: edit
    scope: "record the accepted next architecture route"
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: "slices 003-005 done"
next_agent_action: "Summarize pass/fail evidence and recommend core, payload, or replacement."
human_action: ""
can_continue_other_slices: false
notes: "Decision accepted from evidence: keep KB as lightweight core and payload; do not add a second state runtime."
---

# Slice 006 - Core vs Payload Decision Report

## What To Decide

Use spike evidence to decide one of three routes:

1. KB remains both payload and lightweight runtime core.
2. KB becomes the payload on a small runtime/state engine.
3. KB is replaced or forked from a named upstream only if the spike shows a
   deeper mismatch.

## Acceptance Criteria

- The report includes executable evidence, not only model judgment.
- It names what KB is better at and what HumanLayer-style runtime machinery is
  better at after the spike.
- It explains whether a future bakeoff is still needed and what it would measure.
- It records any adapter boundary failures or state recovery failures.
- The recommendation is tied to executable spike evidence.

## Scope Boundary

Do not recommend another runtime without a demonstrated recovery or adapter
failure.

## Verification

Proof commands from slices 002-005 plus `go run ./cmd/kbcheck core`.
