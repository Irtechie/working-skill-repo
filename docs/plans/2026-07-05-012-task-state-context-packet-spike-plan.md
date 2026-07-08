---
kb_id: kb-2026-07-05-model-agnostic-planner-economy
slice_id: slice-002
title: "Spike repo-local task state and context-packet object"
blockers: [slice-001]
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: large
model_tier_reason: "This defines the durable state spine and must avoid locking KB to one runtime or brittle markdown-as-database behavior."
hitl: false
expected_files:
  - path: cmd/kbcheck/task_state.go
    op: create
    scope: "define and validate the smallest task-state/context-packet schema"
  - path: cmd/kbcheck/task_state_test.go
    op: create
    scope: "cover valid state, missing packet fields, invalid transitions, and recovery hints"
  - path: cmd/kbcheck/testdata/task-state-valid.json
    op: create
    scope: "fixture for a valid slice task with context packet and proof target"
  - path: cmd/kbcheck/testdata/task-state-stuck.json
    op: create
    scope: "fixture for an interrupted or waiting state that requires deterministic recovery"
  - path: docs/context/kb/task-state-schema.md
    op: create
    scope: "document the repo-local state schema and storage boundary"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: "slice-001 done"
next_agent_action: "Implement the minimal schema and validator before touching skill text."
human_action: ""
can_continue_other_slices: true
---

# Slice 002 - Task State and Context Packet Spike

## What To Build

Create the smallest structured object that can represent:

- task id, parent id, slice id, status, phase, owner, and timestamps;
- context packet fields needed by a cheaper worker;
- HITL, blocked, interrupted, failed, completed, and recovery states;
- predicted model tier, actual runtime/model, token/cost estimate if available;
- proof command/result and packet sufficiency notes.

This should be a repo-local schema and validator first, not a daemon.

## Acceptance Criteria

- `kbcheck` can validate a task-state fixture.
- Invalid transitions and missing packet fields fail deterministically.
- Stuck or interrupted state produces a concrete recovery hint.
- The schema has no Claude/Codex/vendor-specific required fields.
- The state object can be used by app repos without committing ephemeral run
  state by default.

## Context Packet Minimum

- repo memory files checked;
- source files/interfaces already read and why;
- deterministic prefetch outputs;
- constraints and out-of-scope boundaries;
- acceptance/proof target;
- predicted `model_tier` and `model_tier_reason`;
- allowed files/tools or broad-search policy;
- escalation triggers.

## Scope Boundary

Do not rewrite `kb-plan` or `kb-work` behavior yet. Prove the state object first.

## Verification

Run:

```shell
go test ./cmd/kbcheck/...
go run ./cmd/kbcheck core
```
