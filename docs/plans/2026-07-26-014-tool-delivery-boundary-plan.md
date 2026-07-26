---
kb_id: kb-2026-07-26-plan-run-worktree-isolation
slice_id: slice-004
title: "Keep default-branch delivery and dirty-WIP authority outside kb-work"
blockers: [slice-003]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "This is the team-safety boundary between reversible internal integration and authorized PR or direct-default delivery."
model_requirements:
  - "Git remote/default-branch detection."
  - "Delivery-policy reasoning."
  - "Dirty-work and commit-authority safety."
  - "Cross-skill contract testing."
escalation_triggers:
  - "kb-work can merge or push the resolved remote default."
  - "Worktree execution silently commits user-owned dirty files."
  - "Absence of delivery policy becomes direct delivery."
  - "Team mode depends on local common-dir locks."
context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-004.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'DefaultBranchBoundary|DirtyBaseAuthority|DeliveryOwner' -count=1"
  expect: 0
hitl: false
workspace_mode: shared-serial
conflict_domains: [git:default-branch, skill:kb-work, skill:kb-complete, skill:kb-ship, skill:kb-land]
shared_resources: [git:integration-owner, git:delivery-owner]
expected_files:
  - path: cmd/kbcheck/delivery_boundary_test.go
    op: create
    scope: "Protect default-branch refusal, local/PR/direct ownership, and dirty-base commit authority."
  - path: cmd/kbcheck/worktree_isolation.go
    op: edit
    scope: "Refuse internal integration when the target resolves to a local or remote default branch."
  - path: cmd/kbcheck/plan_run_workspace.go
    op: edit
    scope: "Record delivery mode as context while keeping it outside internal integration authority."
  - path: cmd/kbcheck/skill_repo_contract_test.go
    op: edit
    scope: "Require kb-work and delivery skills to preserve their separate authorities."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Forbid default-branch delivery and hidden dirty-WIP checkpoint commits during execution."
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "Accept a reviewed plan-run branch as the only delivery candidate and preserve local default."
  - path: .github/skills/kb-ship/SKILL.md
    op: edit
    scope: "Ship the plan-run topic branch to a correctly based manual PR without merging."
  - path: .github/skills/kb-land/SKILL.md
    op: edit
    scope: "Remain the only skill authorized to integrate the resolved remote default under explicit policy."
  - path: .github/skills/kb-configure/SKILL.md
    op: edit
    scope: "Document local default and PR/manual team recommendation without auto-enabling direct delivery."
protected_oracles:
  - path: cmd/kbcheck/delivery_boundary_test.go
    role: "default-branch refusal and dirty-work authority oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan amendment"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Fail closed on default-branch internal integration and route completed plan branches through configured local, PR, or explicitly authorized direct delivery."
human_action: ""
can_continue_other_slices: false
---

# Keep Delivery and Dirty-WIP Authority Outside `kb-work`

## What to Build

Make internal plan execution and repository delivery separate authorities.
`kb-work` may create local plan/slice commits required by an explicitly selected
worktree mode and integrate them into the plan-run branch. It cannot merge,
push, or retarget the resolved default branch.

If source work depends on uncommitted user changes, the run blocks and requests
an explicit checkpoint/patch decision or shared-serial execution; it never
silently commits or omits that WIP.

## Acceptance Criteria

- `kb-work` integration refuses local and remote default branch targets.
- No delivery policy remains local-only.
- PR/manual is documented as the recommended team policy.
- `kb-ship` receives the plan-run topic branch, audits its full diff, pushes a
  non-default ref, and opens/updates a PR without merging.
- Only `kb-land` may perform default-branch integration after explicit direct or
  authorized auto-merge policy.
- Worktree mode states exactly what local branch commits it needs; it does not
  imply authority to commit user-owned source-checkout changes.
- Dirty WIP required by the task produces an explicit recovery decision.
- Local common-dir leases are never presented as cross-machine team locks.

## Test Scenarios

1. Internal integration target resolves to `main`, `master`, or the fetched
   remote default; it is refused before merge.
2. Delivery config is absent; the completed plan-run branch remains local.
3. Delivery mode is PR/manual; shipping targets a non-default topic ref and
   stops with an open PR.
4. Direct delivery is not explicitly configured; `kb-land` refuses.
5. The source checkout has relevant uncommitted changes; worktree preparation
   blocks without creating a checkpoint commit.
6. The source checkout has unrelated dirty changes; the isolated run preserves
   them byte-for-byte.

## Proof Check

```powershell
go test ./cmd/kbcheck -run 'DefaultBranchBoundary|DirtyBaseAuthority|DeliveryOwner' -count=1
```

## Scope Boundary

- No automatic PR merge.
- No force push, admin bypass, branch-protection bypass, stash, reset, or
  destructive history rewrite.
- No distributed lock service.
- No change to repository write permissions.

## Dependencies

Requires slice-003 so the internal integration target is already the plan-run
branch before the default-branch guard becomes mandatory.

