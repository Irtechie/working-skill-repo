---
kb_id: kb-2026-07-30-bounded-graph-run-provenance
slice_id: slice-004
title: "Explain graph-run failure and incompletion"
blockers: [slice-003]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Causal diagnosis must reconcile DAG state, attempt receipts, proof, gates, and fan-in without reporting downstream symptoms as root causes."
model_requirements: ["deterministic graph traversal", "compact CLI and JSON design", "completion predicate reasoning", "corrupt-evidence handling"]
escalation_triggers: ["multiple causal roots cannot be represented deterministically", "diagnostics need raw transcript ingestion", "completion output disagrees with manifest-contract", "corrupt evidence is silently ignored"]
workspace_mode: shared-serial
conflict_domains: ["file:cmd/kbcheck/graph_run.go", "file:cmd/kbcheck/main.go", "namespace:graph-run-cli"]
shared_resources: ["git:integration-owner"]
context_packet_path: docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-004.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'GraphRunInspect' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/graph_run.go
    op: create
    scope: "Load validated run evidence and compute failed and why-not-done projections."
  - path: cmd/kbcheck/graph_run_test.go
    op: create
    scope: "Protect causal ordering, completion reasons, corruption reporting, and compact output."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Register graph-run inspect with --failed, --why-not-done, and --json."
protected_oracles:
  - path: cmd/kbcheck/graph_run_test.go
    role: "causal diagnosis and completion projection oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan update"
status: retired
retired: 2026-08-28
retired_reason: parent goal bounded-graph-run-provenance retired undelivered; kept as a design record
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Protect text and JSON CLI scenarios, then implement first-causal-failure and why-not-done projections over validated receipts and gates."
human_action: ""
can_continue_other_slices: true
---

# Explain Graph-Run Failure and Incompletion

## What to Build

Implement `kbcheck graph-run inspect --failed` and
`kbcheck graph-run inspect --why-not-done` with compact text and stable JSON
outputs derived from validated manifest, gate, storage, and receipt evidence.

## Acceptance Criteria

- `--failed` reports the first causal failed attempt, its upstream inputs,
  downstream impact, retry/supersession state, and evidence pointers.
- `--why-not-done` evaluates the completion predicate and reports missing
  terminal nodes, invalid/missing proof, incomplete fan-in gates, and unresolved
  blocking edges.
- Ordering is deterministic when multiple failures exist.
- Corrupt or missing evidence is explicit and never treated as success.
- Default text is compact; JSON contains stable machine-readable fields.
- Invalid selector combinations fail with concise nonzero output.

## Test Scenarios

- Single root failure with downstream blocked nodes.
- Downstream symptom with earlier causal parent failure.
- Multiple independent failures with deterministic ordering.
- Missing terminal node, missing proof, blocked edge, and incomplete fan-in.
- Fully complete run returns done and no false diagnosis.
- Corrupt receipt returns evidence-invalid.

## Scope Boundary

No raw log rendering, transcript summarization, LLM diagnosis, or new graph
topology source.
