---
kb_id: kb-2026-07-05-model-agnostic-planner-economy
slice_id: slice-002
title: "Implement context-packet and execution-telemetry contracts"
blockers: [slice-001]
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: large
model_tier_reason: "This defines a vendor-neutral execution boundary without duplicating the existing manifest/run-state spine."
hitl: false
expected_files:
  - path: cmd/kbcheck/context_packet.go
    op: create
    scope: "define and validate the smallest context-packet schema"
  - path: cmd/kbcheck/context_packet_test.go
    op: create
    scope: "cover valid packets and missing authority-boundary fields"
  - path: cmd/kbcheck/testdata/context-packet-valid.json
    op: create
    scope: "fixture for a valid slice context packet and proof target"
  - path: cmd/kbcheck/testdata/context-packet-invalid.json
    op: create
    scope: "fixture missing required packet fields"
  - path: docs/context/kb/context-packet-schema.md
    op: create
    scope: "document the packet/telemetry schema and runtime boundary"
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: "slice-001 done"
next_agent_action: "Implement the minimal schema and validator before touching skill text."
human_action: ""
can_continue_other_slices: true
notes: "Implemented vendor-neutral context packet validation, fixtures, docs, and CLI/selftest. Execution telemetry is a separate typed result linked by packet_id. Deliberately reused existing lifecycle state."
---

# Slice 002 - Task State and Context Packet Spike

## What To Build

Create the smallest structured context packet a bounded worker can consume.
Reuse manifests, route history, goal ledgers, and proof traces for lifecycle
state; do not create a second task database.

## Acceptance Criteria

- `kbcheck` can validate a context-packet fixture.
- Missing packet fields fail deterministically.
- The schema has no Claude/Codex/vendor-specific required fields.
- Separate execution telemetry normalizes runtime/model, predicted/actual tier,
  turns, input/output/cache tokens, rework, escalation, proof result, and packet
  sufficiency.
- Raw usage values remain authoritative; no unversioned weighted score is
  treated as proof.

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

Do not build a daemon, task database, or duplicate state machine.

## Verification

Run:

```shell
go test ./cmd/kbcheck/...
go run ./cmd/kbcheck core
```
