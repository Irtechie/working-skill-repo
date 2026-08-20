---
kb_id: kb-2026-08-20-runtime-state-contract
slice_id: runtime-001
title: Define a versioned runtime state and transition contract
blockers: []
verification: tdd
test_level: integration
functional_risk: broad
model_tier: large
model_tier_reason: State semantics cross planning, execution, proof, and delivery.
model_requirements: [Go contract design, state-machine reasoning, compatibility testing]
escalation_triggers: [phase authority collapses, state cannot be resumed, secrets enter records]
token_budget: 6000
workspace_mode: shared-serial
conflict_domains: [namespace:runtime-state, file:cmd/kbcheck/main.go, file:docs/context/architecture/kb-workflow.md]
proof_check: {kind: command_exit, command: "go test ./cmd/kbcheck -run RuntimeState -count=1", expect: 0}
hitl: false
status: pending
---

# Define a Versioned Runtime State and Transition Contract

## Acceptance Criteria

- A strict, versioned record represents run identity, phase, heartbeats,
  blockers, route receipts, proof receipts, and allowed transitions.
- Invalid, stale, contradictory, or authority-escalating records fail closed.
- Existing manifest gate semantics remain the authority for phase advancement.

## Expected Files

- `cmd/kbcheck/*runtime_state*.go` — contract/parser and tests.
- `cmd/kbcheck/main.go` — bounded public validation command.
- `docs/context/architecture/kb-workflow.md` — runtime/adapter boundary.

## Test Scenarios

- Valid serial transition sequence passes.
- Illegal phase skip, stale heartbeat, forged receipt, and invalid blocker fail.
- Legacy manifest proof remains readable without granting new transitions.

## Scope Boundary

No worker dispatch, PR delivery, model selection, or skill deletion in this slice.
