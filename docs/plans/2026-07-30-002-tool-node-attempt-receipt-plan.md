---
kb_id: kb-2026-07-30-bounded-graph-run-provenance
slice_id: slice-002
title: "Emit immutable node-attempt receipts"
blockers: [slice-001]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "The receipt must unify existing route, lease, revision, dependency, and proof evidence without trusting self-reported fields."
model_requirements: ["Go contract design", "hash and attestation validation", "existing execution telemetry fluency", "bounded metadata review"]
escalation_triggers: ["receipt requires raw prompts or outputs", "authoritative revision or lease fields cannot be bound", "attempt identity conflicts across existing ledgers", "emission mutates accepted receipts"]
workspace_mode: shared-serial
conflict_domains: ["file:internal/modelrouting/node_attempt_receipt.go", "file:cmd/kbcheck/execution_telemetry.go", "namespace:.kb/runs"]
shared_resources: ["git:integration-owner", "storage:.kb/runs"]
context_packet_path: docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-002.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck ./internal/modelrouting -run 'NodeAttemptReceipt|ExecutionTelemetry' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: internal/modelrouting/node_attempt_receipt.go
    op: create
    scope: "Define, hash, validate, and reject unsafe fields in the versioned node-attempt receipt."
  - path: internal/modelrouting/node_attempt_receipt_test.go
    op: create
    scope: "Protect identity, hash, timestamp, status, dependency, revision, route, lease, and proof bindings."
  - path: cmd/kbcheck/execution_telemetry.go
    op: edit
    scope: "Reuse validated telemetry evidence when emitting a terminal node-attempt receipt."
  - path: cmd/kbcheck/execution_telemetry_test.go
    op: edit
    scope: "Prove CLI emission, immutable collision handling, and metadata-only output."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Add the minimum receipt-emission arguments without creating a second telemetry command family."
protected_oracles:
  - path: internal/modelrouting/node_attempt_receipt_test.go
    role: "receipt contract and tamper-detection oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan update"
status: retired
retired: 2026-08-28
retired_reason: parent goal bounded-graph-run-provenance retired undelivered; kept as a design record
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Freeze the versioned bounded receipt contract, prove tampering and raw-payload rejection RED, then wire atomic per-attempt emission."
human_action: ""
can_continue_other_slices: true
---

# Emit Immutable Node-Attempt Receipts

## What to Build

Create one immutable receipt per terminal node attempt by joining existing
validated execution telemetry, route receipt, lease/revision state, dependency
identity, and proof hashes. Store it inside the marker-owned run directory.

## Acceptance Criteria

- Schema version 1 contains stable run/node/slice/attempt identity, dependencies,
  session, context hash, state/base/head revisions, actual route/model, lease
  generation, timestamps, terminal status, retry count, and proof hash.
- Integrity hash covers every semantic field.
- Existing route attestations and proof hashes are referenced, not recomputed
  from untrusted prose.
- The contract rejects prompts, outputs, transcripts, diffs, screenshots, and
  unknown oversized fields.
- Re-emitting byte-identical content is idempotent; conflicting content for an
  existing attempt fails closed.
- Each receipt remains below the existing 1 MiB evidence bound.

## Test Scenarios

- Valid passed and failed attempts emit deterministic receipts.
- Changed semantic field breaks integrity.
- Duplicate identical emission succeeds without mutation.
- Duplicate conflicting emission fails.
- Invalid dependency, revision, timestamp, status, or proof binding fails.
- Raw payload fields and oversized receipts fail.

## Scope Boundary

No OpenTelemetry backend, span exporter, transcript capture, model call, or new
attempt scheduler.
