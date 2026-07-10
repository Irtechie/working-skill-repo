---
kb_id: kb-2026-07-10-session-model-routing
slice_id: slice-004
title: "Dispatch Codex and custom-provider workers with route-bound receipts"
blockers: [slice-001, slice-002, slice-003]
verification: integration
test_level: functional-cli
functional_risk: broad
model_tier: large
context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-004.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbrouter -run 'Dispatch|Receipt|Fallback'"
  expect: 0
hitl: false
expected_files:
  - path: internal/modelrouting/dispatch.go
    op: create
    scope: "typed host dispatch request, least-privilege authority, cancellation, and handoff state"
  - path: cmd/kbrouter/dispatch.go
    op: create
    scope: "Codex exec explicit-model/profile adapter and receipt writer"
  - path: cmd/kbrouter/dispatch_test.go
    op: create
    scope: "fake-host functional dispatch, timeout, fallback, mismatch, and partial-work fixtures"
  - path: cmd/kbcheck/execution_telemetry.go
    op: edit
    scope: "separate requested/actual route, run/session IDs, receipt status, and unavailable telemetry"
  - path: cmd/kbcheck/testdata/execution-telemetry-valid.json
    op: edit
    scope: "route-aware telemetry fixture"
protected_oracles:
  - path: cmd/kbrouter/dispatch_test.go
    role: "dispatch authority and receipt oracle"
    sha256: "a8c38d5d13edb9c9ed39e2c3752157af68ca9eb8548d99d7bfce1cfedb4a2eba"
    update_policy: "requires explicit plan update"
status: done
owner: agent
can_continue_other_slices: false
---

# Codex Model Dispatch

Continuation and recovery are split by difficulty in
`docs/plans/2026-07-10-038-session-model-routing-continuation-plan.md`.
The implementation and its public wiring are complete. C1/C2 evidence is in
`.kb/runs/session-model-routing/c1-final-report.md` and
`.kb/runs/session-model-routing/c2-final-report.md`.

## What To Build

Launch a bounded Codex worker with explicit `--model` and optional user-local provider profile. Go chooses/records the route; Codex owns the agent loop, tools, and workspace execution.

The initial external-process adapter is `dispatch-proven` on Windows through a
Job Object. Linux/macOS builds keep discovery and route selection portable but
return typed `dispatch-unavailable` before worker start; `kb-work` can continue
with the current model. Those platforms must not claim external worker dispatch
until a containment adapter passes native proof.

## Acceptance Criteria

- Native and custom-provider routes launch only through typed built-in adapters with no arbitrary command field.
- Context packet, cwd/worktree, sandbox, allowed roots/tools/network, timeout, output schema, and receipt paths are bounded.
- JSONL/session/provider-model evidence is parsed when present; unavailable/mismatch is explicit and never falsified.
- Fallback re-redacts diagnostics, rechecks trust, preserves partial diff/proof, and uses a finite attempt ledger.
- Correct work remains valid when route attribution is unknown.
- A platform without proven process-tree containment starts no worker, emits no
  execution receipt, consumes no model fallback, and reports
  `dispatch-unavailable` so current-model work can continue.

## Test Scenarios

- Fake Codex emits exact model/session evidence, missing evidence, mismatched model, timeout, nonzero exit, and partial diff.
- Custom profile is user-local; tracked project attempts to set provider/profile are rejected.
- Less-trusted fallback requires approval; raw provider output and credentials are never forwarded.

## Tier Rationale

Large: external process control, credentials, sandbox authority, fallback, and provenance are high-risk.

## Scope Boundary

No direct LLM agent loop, GHCP supported claim, TinyBoss control action, or generated named-agent fallback.
