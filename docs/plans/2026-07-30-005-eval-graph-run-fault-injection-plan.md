---
kb_id: kb-2026-07-30-bounded-graph-run-provenance
slice_id: slice-005
title: "Prove causal diagnosis, bounded retry, and completion"
blockers: [slice-004]
verification: functional
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "The final proof must inject failures across storage, attempts, dependencies, and gates, then show accepted nodes are not replayed."
model_requirements: ["deterministic fixture design", "retry-state modeling", "release-gate interpretation", "documentation and proof synthesis"]
escalation_triggers: ["fixture expectations are not mechanically consumed", "retry replays an accepted node", "local-release fails outside the named Windows harness blocker", "proof depends on live model calls"]
workspace_mode: shared-serial
conflict_domains: ["path:evals/graph-run", "file:cmd/kbcheck/graph_run_selftest.go", "file:cmd/kbcheck/checks.go", "docs:graph-run"]
shared_resources: ["git:integration-owner", "release:local-release"]
context_packet_path: docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-005.json
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck graph-run-selftest"
  expect: 0
hitl: false
expected_files:
  - path: evals/graph-run/fixtures.json
    op: create
    scope: "Define fixed storage, failure, retry, fan-in, and completion scenarios with deterministic expected outputs."
  - path: cmd/kbcheck/graph_run_selftest.go
    op: create
    scope: "Consume every fixture expectation and fail mechanically on mismatch."
  - path: cmd/kbcheck/graph_run_selftest_test.go
    op: create
    scope: "Prove the selftest rejects dishonest expectations and covers bounded retry."
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "Add graph-run-selftest to the canonical deterministic check registry."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Register graph-run-selftest."
  - path: README.md
    op: edit
    scope: "Document the bounded graph-run provenance and compact inspection commands."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document focused graph-run proof and release commands."
protected_oracles:
  - path: evals/graph-run/fixtures.json
    role: "fault-injection and retry/completion oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Create mechanically consumed fault fixtures, prove repair and retry behavior, update command/testing docs, then run core and local-release."
human_action: ""
can_continue_other_slices: true
---

# Prove Causal Diagnosis, Bounded Retry, and Completion

## What to Build

Add a deterministic fixture corpus and selftest that injects graph-run storage,
attempt, dependency, proof, and gate failures, then repairs the causal node and
proves only the required retry path executes before completion becomes true.

## Acceptance Criteria

- Every fixture contains fixed inputs and mechanically consumed expected fields.
- Scenarios cover owned/unowned retention, causal parent failure, downstream
  blocking, retry supersession, accepted-node non-replay, missing proof,
  incomplete fan-in, corrupt evidence, and final completion.
- Repair retries the failed node and affected descendants only.
- Accepted independent nodes retain their original receipt identities.
- The selftest performs no live model calls and requires no daemon/backend.
- README and testing docs show the two inspect commands and bounded storage
  behavior.
- Focused tests, `core`, `local-release`, and `git diff --check` are attempted
  on the exact delivery tree; the known Windows harness issue remains a scoped
  delivery blocker if freshly reproduced.

## Scope Boundary

No benchmark, live route execution, external telemetry service, exhaustive
trace capture, or repair of unrelated harness failures.
