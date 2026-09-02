---
name: kb-configure
description: "Configure portable per-project delivery and plan-worktree concurrency policy."
argument-hint: "[show|delivery-local|delivery-pr|delivery-direct|reset]"
---

# KB Configure

Configure optional project execution policy without making ordinary KB startup
interactive.

Orchestrator-directed DDR is the normal execution path and needs no project
setup. This skill owns delivery and plan-worktree concurrency policy. Personal
source preference belongs to user-local `kb-models` state keyed by project
identity.

## Config Path

`docs/context/operations/kb-routing.yaml`

The file contains portable project policy only. Never write model endpoints,
auth environment-variable names, trust approvals, commands, or credentials
there. `kbrouter` continues to own host and user-local model configuration.

## Behavior

1. If the file exists, read it and show a compact delivery/concurrency summary. Ask only
   which setting the user wants to change when their request is ambiguous.
2. If the file is absent and an argument was supplied:
   - `delivery-local` keeps reviewed work local. This is an opt-out: it leaves
     finished work on a branch with no review path, so choose it only for
     scratch or private projects.
   - `delivery-pr` commits, pushes a topic/fork branch, and opens/updates a PR.
     This is the default when no policy file exists.
   - `delivery-direct` permits verified direct-default integration; protection
     or policy rejection falls back to PR or blocks.
   - `show` reports PR/manual delivery and the plan-worktree limit without
     creating a file.
   - `reset` removes only this project policy after explicit confirmation.
3. If the file is absent and no mode was supplied, show the defaults and the
   exact commands above. Do not start a setup questionnaire.
4. Do not ask model-by-model questions. Normal DDR reads the active host schema
   and user-local catalog at work time; `kb-models` configures optional
   user-local extras only.
5. Preserve unrelated project policy when updating an existing file.

## Canonical Schema

```yaml
schema_version: 1

delivery:
  mode: pr
  merge: manual
  post_merge_sync: false

execution:
  max_plan_run_worktrees: 2
```

`execution.max_plan_run_worktrees` caps live KB-owned plan-run worktrees per
repository, so it bounds concurrent manifest groups. `kbcheck plan-worktree
--action prepare` fails closed at the ceiling; `adopt` creates nothing and is
never capped. Harness-created session worktrees are not counted and never
removed by KB. Raise it only with evidence that merge cost stayed lower than the
wall-clock saved.

These safety rules are fixed rather than configurable:

- `model_tier` is the minimum execution capability, not the validator.
- Normal work uses one explicit `current` or `delegated` owner decision.
- Delegated work selects exactly one qualified same-tier-or-higher route.
- Normal work never routes below the planned tier or silently falls back across
  owners.
- Ordinary proof remains authoritative. Routing receipts are telemetry.
- Repository ownership/write access never selects direct delivery by itself.
- Direct delivery, automatic merge, and post-merge sync require explicit policy.
- PR/manual is the default: reviewed work becomes a pushed, review-ready PR
  without asking. It never authorizes merge. Reaching PR-ready is automatic;
  accepting a PR never is.
- Local common-directory leases coordinate sibling worktrees only; they are not
  cross-machine team locks.

## Defaults

When no config exists:

- Normal DDR is orchestrator-owned.
- Delivery is PR with manual merge.
- At most two KB-owned plan-run worktrees may be live per repository.
- `kb-start`, `kb-plan`, and `kb-work` do not ask configuration questions.

## Output

After writing, report the path and a one-line summary:

```text
KB configured: delivery PR/manual; max plan-run worktrees 2.
```
