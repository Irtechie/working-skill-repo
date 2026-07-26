---
kb_id: kb-2026-07-26-plan-run-worktree-isolation
slice_id: slice-003
title: "Advance slice commits only on the owning plan-run branch"
blockers: [slice-002]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Serialized commit acceptance, compare-and-swap head state, and proof replay on the sole plan-run branch are data-loss-sensitive architecture."
model_requirements:
  - "Plan-run branch identity tests."
  - "Compare-and-swap integration-head state."
  - "Single-worktree commit receipt validation."
  - "Post-integration proof enforcement."
escalation_triggers:
  - "A slice runs in another worktree or branch."
  - "A commit is accepted after unexpected integration-head movement."
  - "The plan-run worktree is dirty at receipt time."
  - "Proof is accepted only from worker self-report."
context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-003.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'PlanRunAdvance|IntegrationHead|SliceCommit' -count=1"
  expect: 0
hitl: false
workspace_mode: shared-serial
conflict_domains: [file:cmd/kbcheck/plan_run_workspace.go, git:plan-run-branch, skill:kb-work]
shared_resources: [git:integration-owner, git:plan-run-branch]
expected_files:
  - path: cmd/kbcheck/plan_run_integration_test.go
    op: create
    scope: "Protect sequential same-worktree slice commits, branch identity, CAS, clean state, and proof replay."
  - path: cmd/kbcheck/plan_run_workspace.go
    op: edit
    scope: "Atomically advance integration_head after a proven slice commit on the owning plan-run branch."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Require plan-run identity and expected integration head for slice advance actions."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Keep slices in the one plan-run worktree and rerun proof after each accepted commit."
  - path: .github/skills/kb-work/references/worktree-isolation.md
    op: edit
    scope: "Document same-worktree slice receipts, integration-head CAS, and aggregate proof order."
  - path: .github/skills/kb-work/references/execution-prompt.md
    op: edit
    scope: "Return the plan-run commit, observed writes, and proof without changing branches or editing lifecycle state."
protected_oracles:
  - path: cmd/kbcheck/plan_run_integration_test.go
    role: "serialized same-worktree slice commit and integration-head oracle"
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

# Advance Slice Commits on the Owning Plan-Run Branch

## What to Build

Every slice mutates the one worktree owned by its manifest group. After a slice
commits, the coordinator validates owner/run lineage, observed scope, exact
branch/worktree identity, clean state, prior compare-and-swap integration head,
and proof before advancing the recorded integration head.

Advancing after an earlier authorized slice commit is normal. Unexpected
movement fails closed. There is no per-slice branch merge path.

## Acceptance Criteria

- Slice advance requires the explicit plan-run receipt and integration ref.
- The command refuses commits from any other branch or worktree.
- Two serialized slice commits advance the same run head in order.
- Unexpected integration-ref movement not recorded by the coordinator fails
  compare-and-swap validation.
- Dirty or mismatched receipts stop without reset, stash, or cleanup.
- Worker proof is treated as a receipt; the exact slice proof reruns in the
  plan-run worktree.
- Aggregate proof can detect that individually passing slices fail together.
- Workers never update `todo.md`, manifests, active handoffs, or integration
  receipts.

## Test Scenarios

1. Two serialized slices commit in the plan-run worktree; accept A then B and
   prove the recorded head advances in order.
2. A receipt from another branch/worktree fails before state mutation.
3. An outside actor moves the integration ref without advancing the receipt;
   acceptance fails closed.
4. The plan-run checkout is dirty or its receipt owner/run mismatches; advance
   fails before state mutation.
5. Worker proof passes but aggregate proof fails after the commit; the slice
   is not marked done.

## Proof Check

```powershell
go test ./cmd/kbcheck -run 'PlanRunAdvance|IntegrationHead|SliceCommit' -count=1
```

## Scope Boundary

- Do not push or open/merge PRs.
- Do not create per-slice branches or worktrees.
- Do not integrate into the remote default branch.
- Do not reset, stash, or force-clean the plan-run worktree.

## Dependencies

Requires slice-002 so only the current run integration owner can advance the
plan-run branch.
