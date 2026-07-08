---
kb_id: kb-2026-07-05-model-agnostic-planner-economy
slice_id: slice-003
title: "Wire kb-plan and kb-work through context packets"
blockers: [slice-002]
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: large
model_tier_reason: "This changes planner/worker contracts and determines whether cheaper worker packets are operationally useful."
hitl: false
expected_files:
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "require split artifacts and context packets for non-trivial or high-risk plans"
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "consume packet data before execution and treat missing packet data as a planning defect"
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "validate packet references or embedded packet fields where the manifest opts in"
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: "cover missing context packet and valid packet-backed slice fixtures"
  - path: evals/route-complexity/skill-bundle-change.json
    op: edit
    scope: "expect packet-backed planning for high-risk skill-bundle changes"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: "slice-002 done"
next_agent_action: "Update planner and worker contracts to use the structured packet."
human_action: ""
can_continue_other_slices: true
---

# Slice 003 - Planner/Worker Packet Integration

## What To Build

Make KB's expensive planner produce enough structured packet data that a cheaper
worker can execute a bounded slice without rediscovering the repo.

The planning ladder is:

1. questions;
2. objective research;
3. design concept;
4. structure outline;
5. tactical vertical slices;
6. per-slice context packet;
7. work execution and proof update.

## Acceptance Criteria

- `kb-plan` explains when split artifacts and packets are required.
- `kb-work` reads the packet first and escalates if the packet is insufficient.
- `kbcheck` catches at least one missing-packet or malformed-packet case.
- Low-risk doc-only tasks can stay compact; the spike must not add ceremony to
  every tiny edit.
- Existing gate-ledger behavior remains intact.

## Scope Boundary

Do not add multiple runtime adapters in this slice. Use the packet contract as
the shared boundary.

## Verification

Run:

```shell
go test ./cmd/kbcheck/...
go run ./cmd/kbcheck core
```
