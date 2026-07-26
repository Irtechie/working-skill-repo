---
title: Manifest-owned plan-run worktrees with shared-serial slices
date: 2026-07-26
category: workflow-issues
module: kb-work
problem_type: workflow_issue
component: development_workflow
severity: high
applies_when:
  - "Executing a KB manifest with multiple mutating slices"
  - "Running multiple manifest groups concurrently in sibling Git worktrees"
  - "Coordinating overlapping file, conflict-domain, or shared-resource claims"
tags: [plan-run, worktree-isolation, manifest-groups, shared-serial, collision-prevention, slice-leases, integration-head]
---

# Manifest-owned plan-run worktrees with shared-serial slices

## Context

Multiple valid plan sets can collide when they mutate one checkout, but creating
a branch and worktree for every slice moves the collision into a growing merge
queue. The stable isolation unit is the manifest/workstream group: one group
owns one non-default integration branch and worktree, and its slices advance
there sequentially.

Earlier guidance allowed worktrees for independently parallel slices. That
history is useful evidence, but it is superseded by the July 26 contract after
real multi-plan collision pressure showed that slice-level isolation was too
granular. (auto memory)

## Guidance

For a plan-run manifest:

1. Prepare one manifest-owned worktree and non-default integration ref.
2. Require every slice to use `workspace_mode: shared-serial`.
3. Acquire a manifest lease over normalized file, prefix, conflict-domain, and
   shared-resource claims before mutation. Each slice acquires an exact subset
   under the same run and owner lineage.
4. Commit implementation and its lifecycle projection together. There is no
   later lifecycle-only commit and no per-slice merge.
5. Accept a slice commit only when worktree/ref/owner identity, the expected
   integration head, the exact commit write set, claims, and coordinator-replayed
   proof all agree.
6. Archive accepted proof bytes with their SHA-256 and revalidate the per-slice
   proof ledger at completion.
7. Keep default-branch integration, push, PR creation, and merge outside
   `kb-work`. Missing delivery policy remains local-only.

Worktrees isolate files; they do not establish distributed ownership. The lease
state coordinates sibling worktrees sharing one Git common directory. Separate
clones and machines still require branch protection, PR checks, and remote
coordination.

## Why This Matters

One worktree per manifest group prevents unrelated plan sets from trampling one
checkout without creating branch/worktree sprawl inside a plan. Shared-serial
slice commits make the integration head a compare-and-swap sequence instead of
a merge fan-in problem. Exact claims expose collisions before mutation, while
immutable proof archives prevent a mutable receipt from changing what was
accepted. Separating execution from delivery also keeps a successful local run
from becoming an unauthorized merge to the team default branch.

## When to Apply

- A mutating manifest opts into `plan_run_worktree_default: true`.
- Multiple manifest groups may run concurrently, but only while their live
  path, domain, and resource claims are disjoint.
- A dirty source checkout must remain untouched while planned work proceeds.
- Team delivery needs an explicit PR/manual or separately authorized landing
  boundary.

Do not use this contract as a distributed lock, and do not create per-slice
worktrees for an active plan run.

## Examples

The manifest contract is:

```yaml
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
  plan_run_worktree_default: true
  internal_integration_target: plan-run-branch
  default_branch_delivery_owner: kb-complete
  allowed_modes: [shared-serial]
```

The lifecycle is:

```text
prepare group worktree
  -> acquire plan claims
  -> acquire one slice lease
  -> commit implementation + lifecycle
  -> replay proof and advance integration head
  -> release slice lease
  -> repeat
  -> validate immutable proof ledger and complete
  -> release worktree without merging
```

Canonical regression proof:

```powershell
go test ./cmd/kbcheck -run 'PlanRun|CrossManifest|DefaultBranchBoundary|DirtyBaseAuthority|DeliveryOwner' -count=1
go run ./cmd/kbcheck plan-worktree-selftest
go run ./cmd/kbcheck local-release
```

## Related

- [Worktree isolation contract](../../../.github/skills/kb-work/references/worktree-isolation.md)
- [KB workflow architecture](../../context/architecture/kb-workflow.md)
- [Contributor core versus release sync gates](contributor-core-vs-release-sync-gates-2026-06-10.md)
- [Current plan-run manifest](../../plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md)
- [Superseded slice-worktree plan](../../plans/2026-07-19-003-tool-worktree-isolation-plan.md)
