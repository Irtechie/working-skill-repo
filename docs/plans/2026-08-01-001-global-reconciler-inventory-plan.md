---
kb_id: kb-2026-08-01-global-cleanup-reconciliation
slice_id: slice-001
title: "Inventory and plan a portfolio without deleting first"
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: broad
execution_class: cli
model_tier: large
model_tier_reason: "This slice defines the cross-repository evidence model, exact protection predicates, confidence semantics, risk budget, and compact exception packet consumed by later destructive actions."
model_requirements: ["cross-repository Git reasoning", "fail-closed policy design", "CLI contract design", "deterministic fixture construction"]
escalation_triggers: ["a classification lacks authoritative evidence", "a host/provider adapter would be treated as mandatory", "confidence could override a missing predicate"]
workspace_mode: shared-serial
conflict_domains: ["go:reconcile-core", "cli:kbreconcile", "config:reconcile-contract"]
shared_resources: ["git:common-directory", "filesystem:portfolio-ledger"]
proof_check:
  kind: command_exit
  command: "go test ./internal/reconcile ./cmd/kbreconcile -run 'Inventory|Plan|DecisionPacket|NoKBRepo' -count=1"
  expect: 0
hitl: false
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Implement the protected read-only portfolio and cutoff-bound planner."
human_action: ""
can_continue_other_slices: false
protected_oracles:
  - path: internal/reconcile/reconcile_test.go
    purpose: "A mixed portfolio proves exact classification, compact exceptions, and unique-work preservation."
expected_files:
  - path: cmd/kbreconcile/main.go
    op: create
    scope: "Expose dry-run and plan modes without a repo-local kbcheck dependency."
  - path: cmd/kbreconcile/main_test.go
    op: create
    scope: "Prove CLI parsing, JSON output, fail-closed exits, and non-KB repository operation."
  - path: internal/reconcile/reconcile.go
    op: create
    scope: "Define normalized evidence, lifecycle dimensions, modes, and orchestration."
  - path: internal/reconcile/inventory.go
    op: create
    scope: "Inventory canonical repositories, worktrees, refs, dirt, queue/receipt evidence, and optional adapter records."
  - path: internal/reconcile/policy.go
    op: create
    scope: "Load least-privilege policy, mandatory predicates, thresholds, risk caps, and protected path classes."
  - path: internal/reconcile/plan.go
    op: create
    scope: "Classify artifacts and produce cutoff-bound actions, quarantine records, and one compact exception packet."
  - path: internal/reconcile/reconcile_test.go
    op: create
    scope: "Prove at least 20 mixed artifacts, post-cutoff/current/dirty/protected preservation, no per-artifact prompts, and no-KB operation."
  - path: config/reconcile-predicates.json
    op: create
    scope: "Publish the versioned mandatory predicate and downgrade contract."
---

# Slice 001 - Protected Portfolio Planning

## Acceptance Criteria

- `kbreconcile dry-run` and `kbreconcile plan` work from the global binary in a
  repository with no KB files or `cmd/kbcheck`.
- Inventory records canonical repo/worktree/ref identities, cutoff, dirt classes,
  queue/receipt evidence, protection reasons, freshness, and adapter limits.
- Classification is deterministic and deletion-first behavior is impossible.
- Missing mandatory predicates preserve/quarantine; confidence and risk budget
  cannot override them.
- One packet contains at most five grouped human-required decisions; unanswered
  ambiguity preserves data.
- A fixture with at least 20 artifacts retains every annotated unique artifact.

## Test Scenarios

1. Inventory a plain disposable Git repository with primary and linked worktrees.
2. Protect current, primary, default, active, dirty, ignored, post-cutoff, and
   credential/model/learning/live-state artifacts.
3. Classify exact contained, unique, ambiguous, and unsupported-adapter cases.
4. Exhaust the risk budget and report backlog/oldest-age/runs-to-convergence.
5. Emit one bounded decision packet rather than one prompt per artifact.

## Scope Boundary

This slice performs no destructive mutation, PR publication, merge, ref
deletion, worktree removal, or host session retirement.
