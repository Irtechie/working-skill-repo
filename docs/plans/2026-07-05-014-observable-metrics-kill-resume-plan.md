---
kb_id: kb-2026-07-05-model-agnostic-planner-economy
slice_id: slice-004
title: "Prove recovery, stuck-state handling, and execution telemetry"
blockers: [slice-002, slice-003]
verification: integration
test_level: integration
functional_risk: narrow
model_tier: large
model_tier_reason: "This is the self-healing proof: durable state is only useful if stuck states and telemetry are externally observable and recoverable."
hitl: false
expected_files:
  - path: cmd/kbcheck/task_state.go
    op: edit
    scope: "add transition validation, recovery hints, and telemetry checks"
  - path: cmd/kbcheck/task_state_test.go
    op: edit
    scope: "cover blocked HITL, stale interrupting, half-written, failed, resumed, and completed states"
  - path: cmd/kbcheck/ready_set.go
    op: edit
    scope: "surface ready, blocked, waiting, and recoverable task states for manifests"
  - path: docs/context/eval-map.md
    op: edit
    scope: "record the spike's deterministic proof surfaces"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: "slice-003 done"
next_agent_action: "Add recovery and telemetry fixtures before claiming self-healing."
human_action: ""
can_continue_other_slices: true
---

# Slice 004 - Recovery and Telemetry Proof

## What To Build

Add deterministic evidence that the task-state object is safer than markdown
alone.

The explicit lesson from HumanLayer issue evidence is that a state machine can
wedge if recovery transitions are missing. This slice must treat stuck-state
repair as a first-class invariant.

## Acceptance Criteria

- `kbcheck` reports ready, blocked, waiting, stale, recoverable, and invalid
  states separately.
- A stale interrupted/waiting fixture fails with a recovery action, not a vague
  error.
- Telemetry captures predicted tier, actual tier/model when available, proof
  result, rework count, escalation, and packet sufficiency.
- A simulated kill/resume path proves the next action can be recovered from
  structured state.
- No recovery path depends on asking a model to infer state from prose.

## Scope Boundary

Do not build a UI or daemon. Keep the proof in deterministic fixtures and CLI
checks.

## Verification

Run:

```shell
go test ./cmd/kbcheck/...
go run ./cmd/kbcheck ready-set --manifest docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md --json
go run ./cmd/kbcheck core
```
