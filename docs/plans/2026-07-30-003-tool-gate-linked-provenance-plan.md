---
kb_id: kb-2026-07-30-bounded-graph-run-provenance
slice_id: slice-003
title: "Link node receipts to manifest gates"
blockers: [slice-002]
verification: tdd
test_level: integration
functional_risk: broad
model_tier: large
model_tier_reason: "Gate advancement is the completion authority, so receipt linkage must reject stale, failed, unordered, or unproven attempts."
model_requirements: ["manifest gate contract reasoning", "DAG dependency validation", "artifact hash verification", "backward-compatible YAML parsing"]
escalation_triggers: ["legacy manifests become unreadable", "gate passage can occur with failed or stale proof", "receipt linkage requires rewriting unrelated gate history", "fan-in evidence cannot identify contributing receipts"]
workspace_mode: shared-serial
conflict_domains: ["file:cmd/kbcheck/manifest_contract.go", "file:cmd/kbcheck/swarm.go", "namespace:gate-ledger"]
shared_resources: ["git:integration-owner"]
context_packet_path: docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-003.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'Manifest.*Receipt|Gate.*Receipt|NodeAttempt.*Gate' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "Validate optional node receipt path/hash references on terminal slice and fan-in gates."
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: "Protect freshness, dependency order, proof, hash, and backward-compatibility behavior."
  - path: cmd/kbcheck/swarm.go
    op: edit
    scope: "Expose the minimal dependency and terminal-state facts needed for gate receipt validation."
  - path: cmd/kbcheck/swarm_test.go
    op: edit
    scope: "Prove blocked dependencies and fan-in cannot advance from receipt presence alone."
protected_oracles:
  - path: cmd/kbcheck/manifest_contract_test.go
    role: "gate authority and receipt-link oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan update"
status: retired
retired: 2026-08-28
retired_reason: parent goal bounded-graph-run-provenance retired undelivered; kept as a design record
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Add receipt-reference gate fixtures, then enforce terminal, freshness, dependency, and proof bindings while preserving legacy manifests."
human_action: ""
can_continue_other_slices: true
---

# Link Node Receipts to Manifest Gates

## What to Build

Allow terminal slice and fan-in gates to reference node-attempt receipt paths
and hashes, then validate those references as evidence while retaining the gate
ledger as the sole completion authority.

## Acceptance Criteria

- Receipt references are optional for legacy manifests and required only when a
  manifest opts into the new graph-run provenance contract.
- Referenced files stay inside the owning `.kb/runs/<run-id>` directory and
  match their declared hashes.
- A terminal gate rejects nonterminal, failed, stale-revision, wrong-generation,
  missing-proof, or dependency-before-parent receipts.
- Fan-in evidence names every contributing receipt and rejects incomplete sets.
- A valid receipt never advances a gate whose other required evidence or
  blocker contract fails.

## Test Scenarios

- Valid passed attempt advances an otherwise valid terminal gate.
- Failed, running, stale, tampered, or wrong-slice receipt fails.
- Child completion before parent completion fails.
- Fan-in omitting one required parent fails.
- Legacy manifest without the opt-in remains valid.

## Scope Boundary

No automatic manifest mutation, no telemetry-owned completion, and no rewrite
of historical gate entries.
