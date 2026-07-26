---
kb_id: kb-2026-07-26-plan-run-worktree-isolation
slice_id: slice-001
title: "Create a manifest-owned plan-run workspace"
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "This establishes Git ownership and recovery state around dirty user work, where an incorrect base or target can lose or silently omit changes."
model_requirements:
  - "Git worktree and branch-lifecycle reasoning."
  - "Fail-closed Go CLI and receipt contracts."
  - "Windows and Unix path handling."
  - "Backward-compatible migration of existing slice worktree receipts."
escalation_triggers:
  - "A plan-run workspace requires force cleanup."
  - "Dirty uncommitted changes must be copied implicitly."
  - "The default branch becomes the internal integration target."
  - "Legacy slice receipts cannot be migrated safely."
context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-001.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'PlanRunWorkspace|PlanRunManifestContract' -count=1"
  expect: 0
hitl: false
workspace_mode: shared-serial
conflict_domains: [file:cmd/kbcheck/plan_run_workspace.go, file:cmd/kbcheck/manifest_contract.go, skill:kb-plan, skill:kb-work]
shared_resources: [git:integration-owner]
expected_files:
  - path: cmd/kbcheck/plan_run_workspace.go
    op: create
    scope: "Own plan-run workspace receipts, immutable base identity, explicit integration ref, deterministic path, and safe status/release behavior."
  - path: cmd/kbcheck/plan_run_workspace_test.go
    op: create
    scope: "Protect plan-run creation, dirty-source preservation, receipt identity, and non-default integration target behavior."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Expose the plan-run workspace lifecycle through the maintainer CLI."
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "Validate manifest-level plan-run workspace intent without storing live paths or owner tokens."
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: "Reject incomplete or unsafe plan-run workspace contracts."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Make one plan-run branch/worktree the concurrency unit for mutating manifests."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Prepare or resume the manifest-owned workspace before any mutating slice."
  - path: .github/skills/kb-work/references/worktree-isolation.md
    op: edit
    scope: "Separate plan-run workspace preparation from optional child slice worktrees."
protected_oracles:
  - path: cmd/kbcheck/plan_run_workspace_test.go
    role: "manifest-owned workspace and immutable-base oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan amendment"
status: in_progress
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Protect the plan-run workspace oracle, then implement immutable base and explicit integration-ref receipts without touching the default branch."
human_action: ""
can_continue_other_slices: false
---

# Create a Manifest-Owned Plan-Run Workspace

## What to Build

Give each concurrently mutating KB manifest one stable plan-run identity,
descriptive topic branch, and visible worktree. Persist an immutable base ref/SHA,
explicit non-default integration ref, current integration head, manifest path,
owner token, and cleanup state under the Git common directory.

The source checkout is an orientation surface, not an implicit integration
target. Dirty source work is preserved exactly and is never copied into the run
unless an explicit checkpoint/patch contract authorizes it.

## Acceptance Criteria

- One manifest maps idempotently to one plan-run receipt, branch, and worktree.
- Receipt state distinguishes base ref/SHA, integration ref/head, source
  checkout, worktree path, owner, status, and limitations.
- The integration ref is a non-default topic branch owned by the plan run.
- Preparing from a dirty source checkout neither stashes, resets, cleans,
  commits, nor copies its uncommitted changes.
- If the requested work depends on dirty source changes, preparation fails with
  a specific checkpoint-or-shared-serial recovery instruction.
- Paths are deterministic and registered so worktrees do not appear as
  unexplained sibling repositories.
- Existing slice-only receipts remain readable or fail with an explicit
  migration path.
- Release remains non-force and refuses dirty, active, or unintegrated state.

## Test Scenarios

1. Prepare two different manifests from one clean base; each receives a distinct
   branch/worktree and shared common-dir registry.
2. Repeat prepare for the same manifest and owner; the same receipt is returned.
3. Prepare from a dirty source checkout; dirty files remain byte-identical and
   absent from the new run unless explicitly checkpointed.
4. Resolve a remote default branch and prove the run integration ref differs.
5. Attempt path collision, owner mismatch, or unsafe release; each fails closed.

## Proof Check

```powershell
go test ./cmd/kbcheck -run 'PlanRunWorkspace|PlanRunManifestContract' -count=1
```

## Scope Boundary

- Do not implement cross-manifest conflict arbitration; slice-002 owns it.
- Do not integrate child slice receipts; slice-003 owns it.
- Do not push, create a PR, merge default, or change delivery policy.
- Do not use force cleanup, reset, stash, or hidden commits.

## Dependencies

This is the smallest enabling slice for slices 002-005. Execution remains
blocked until the current overlapping dirty baseline is contained.
