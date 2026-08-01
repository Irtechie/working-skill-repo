---
kb_id: kb-2026-08-01-global-cleanup-reconciliation
slice_id: slice-002
title: "Apply and verify only unchanged high-confidence actions"
blockers: [slice-001]
verification: tdd
test_level: functional-cli
functional_risk: destructive
execution_class: cli
model_tier: large
model_tier_reason: "This slice mutates Git worktrees, exact local refs, and queue metadata under races, so it requires deep trust-boundary and recovery reasoning."
model_requirements: ["Git worktree internals", "compare-and-swap refs", "remote containment", "partial failure recovery", "cross-platform locking"]
escalation_triggers: ["an action needs force", "remote authority is unresolved", "the plan cutoff or identity changed", "terminal-cleanup parity cannot be proven"]
workspace_mode: shared-serial
conflict_domains: ["go:reconcile-core", "go:terminal-cleanup", "git:worktrees", "git:refs", "state:work-queue"]
shared_resources: ["git:common-directory", "lock:work-queue", "remote:default"]
proof_check:
  kind: command_exit
  command: "go test ./internal/reconcile ./cmd/kbreconcile ./cmd/kbcheck -run 'Apply|Verify|Reconcile|TerminalCleanup' -count=1"
  expect: 0
hitl: false
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Implement checked apply/verify and terminal-cleanup conformance."
human_action: ""
can_continue_other_slices: false
protected_oracles:
  - path: cmd/kbcheck/terminal_cleanup_test.go
    purpose: "Existing current/primary/dirty/ignored/remote/CAS protections may not regress."
  - path: internal/reconcile/apply_test.go
    purpose: "Every mutation reacquires lock/evidence and fails closed on drift."
expected_files:
  - path: internal/reconcile/apply.go
    op: create
    scope: "Apply only allowlisted unchanged actions with fresh predicate checks."
  - path: internal/reconcile/git.go
    op: create
    scope: "Provide canonical Git evidence, non-force removal, and exact-SHA CAS."
  - path: internal/reconcile/receipt.go
    op: create
    scope: "Write cutoff/action-bound append-only receipts and verify independent outcome dimensions."
  - path: internal/reconcile/apply_test.go
    op: create
    scope: "Prove drift, dirt, ignored files, locks, current/primary/default, rewritten remote, partial residual, and idempotency behavior."
  - path: cmd/kbcheck/terminal_cleanup.go
    op: edit
    scope: "Bind the existing guard to the shared versioned safety contract without weakening predicates."
  - path: cmd/kbcheck/terminal_cleanup_test.go
    op: edit
    scope: "Prove shared contract version and unchanged terminal safety."
  - path: cmd/kbcheck/reconcile_contract_test.go
    op: create
    scope: "Require global and repo-native entrypoints to satisfy one predicate conformance corpus."
---

# Slice 002 - Checked Convergence

## Acceptance Criteria

- Apply accepts only an unexpired plan whose cutoff, target identity, mandatory
  predicates, and policy version still match under the shared repo lock.
- Worktree removal is non-force and preserves tracked, untracked, ignored,
  locked, current, primary, default, moved/recreated, and unresolved targets.
- Local ref deletion uses exact-SHA compare-and-swap only after integrated
  delivery and physical retirement are verified.
- Missing host/forge adapters prohibit PR mutation, merge, remote-ref deletion,
  and destructive host session-record retirement.
- Verify reports delivery, physical cleanup, ref retirement, and session record
  states separately and reconciles only exact safe partial outcomes.
- Repeated apply/verify is idempotent.

## Test Scenarios

1. Change a branch, worktree, claim, receipt, or remote after plan and observe a
   protected/contended outcome with no mutation.
2. Retire one clean, different-session, exact-contained fixture worktree.
3. Reject dirty/ignored/current/primary/default/locked/moved fixtures.
4. Delete only an exact merged local ref through CAS; preserve mismatch.
5. Recover an exact empty residual and preserve every non-empty residual.

## Scope Boundary

No force removal, recursive deletion, remote ref deletion, PR close/merge,
credential inspection, or host record deletion is allowed.

