---
kb_id: kb-2026-07-26-plan-run-worktree-isolation
slice_id: slice-003
title: "Integrate slice receipts only into the owning plan-run branch"
blockers: [slice-002]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Serialized Git integration with expected concurrent head movement, conflict recovery, and proof replay is data-loss-sensitive architecture."
model_requirements:
  - "Three-way Git integration tests."
  - "Compare-and-swap integration-head state."
  - "Recoverable conflict handling."
  - "Post-integration proof enforcement."
escalation_triggers:
  - "Integration can run in an arbitrary source branch."
  - "The second disjoint receipt is rejected only because the first moved the integration head."
  - "A conflict destroys the worker branch or lease."
  - "Proof is accepted only from the worker checkout."
context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-003.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'PlanRunIntegrate|IntegrationHead|ParallelReceipt' -count=1"
  expect: 0
hitl: false
workspace_mode: shared-serial
conflict_domains: [file:cmd/kbcheck/worktree_isolation.go, git:plan-run-branch, skill:kb-work]
shared_resources: [git:integration-owner, git:plan-run-branch]
expected_files:
  - path: cmd/kbcheck/worktree_isolation.go
    op: edit
    scope: "Link child slice receipts to a parent plan run and integrate only into its explicit integration ref/head."
  - path: cmd/kbcheck/worktree_isolation_test.go
    op: edit
    scope: "Preserve legacy lifecycle behavior while migrating receipts to plan-run ownership."
  - path: cmd/kbcheck/plan_run_integration_test.go
    op: create
    scope: "Protect sequential integration of two disjoint same-base receipts, conflict preservation, CAS, and proof replay."
  - path: cmd/kbcheck/plan_run_workspace.go
    op: edit
    scope: "Atomically advance integration_head after coordinator-owned integration."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Require plan-run identity and expected integration head for child integration actions."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Route worker receipts to the owning plan-run coordinator and rerun proof after each integration."
  - path: .github/skills/kb-work/references/worktree-isolation.md
    op: edit
    scope: "Document child receipt, integration-head CAS, conflict recovery, and aggregate proof order."
  - path: .github/skills/kb-work/references/execution-prompt.md
    op: edit
    scope: "Return commit/diff, observed writes, and proof without merging or editing lifecycle state."
protected_oracles:
  - path: cmd/kbcheck/plan_run_integration_test.go
    role: "two-receipt serialized integration and conflict-recovery oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan amendment"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Replace source-HEAD-equals-base integration with serialized integration-head ownership on the explicit plan-run branch."
human_action: ""
can_continue_other_slices: false
---

# Integrate Slice Receipts into the Owning Plan-Run Branch

## What to Build

Optional child slice worktrees return committed receipts to their plan-run
coordinator. The coordinator validates owner/run lineage, observed scope, fork
base, clean worker state, and the current compare-and-swap integration head,
then integrates exactly one receipt into the plan-run branch.

Advancing the plan-run integration head after an earlier authorized receipt is
normal. Unexpected external movement fails closed. A conflict preserves the
worker worktree, branch, lease, and recovery receipt.

## Acceptance Criteria

- Integration requires an explicit parent plan-run receipt and integration ref.
- The command refuses to merge into whichever branch happens to be checked out.
- Two disjoint slice branches created from the same original run base integrate
  serially even though the first integration moves the run head.
- Unexpected integration-ref movement not recorded by the coordinator fails
  compare-and-swap validation.
- Conflicting receipts stop without deleting or force-resetting either branch.
- Worker proof is treated as a receipt; the exact slice proof reruns from the
  updated plan-run worktree.
- Aggregate proof can detect that individually passing slices fail together.
- Workers never update `todo.md`, manifests, active handoffs, or integration
  receipts.

## Test Scenarios

1. Two child branches edit disjoint files from one base; integrate A then B and
   prove both commits are present on the plan-run branch.
2. Two child branches edit the same lines; B conflicts, remains recoverable, and
   no cleanup occurs.
3. An outside actor moves the integration ref without advancing the receipt;
   integration fails closed.
4. A worker checkout is dirty or its receipt owner/run mismatches; integration
   fails before Git mutation.
5. Worker proof passes but aggregate proof fails after integration; the slice
   is not marked done.

## Proof Check

```powershell
go test ./cmd/kbcheck -run 'PlanRunIntegrate|IntegrationHead|ParallelReceipt' -count=1
```

## Scope Boundary

- Do not push or open/merge PRs.
- Do not resolve merge conflicts automatically.
- Do not integrate into the remote default branch.
- Do not force-remove dirty or unintegrated worktrees.

## Dependencies

Requires slice-002 so only the current run integration owner can advance the
plan-run branch.

