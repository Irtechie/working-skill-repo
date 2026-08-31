---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-08-30-rehab-closure-rationalization
source: docs/brainstorms/2026-08-30-rehab-closure-rationalization-requirements.md
source_sha256: 72901ac025c9cae11324473243589722b899728254272a0d5756c6ecf32a74bb
created: 2026-08-30
status: completed
workflow_shape: pipeline-change
objective_contract: true
model_tier_contract: true
workspace_isolation_contract: true
proof_governor_contract: true
pre_slice_review_contract: true
pre_slice_review:
  status: not-required
  source: docs/brainstorms/2026-08-30-rehab-closure-rationalization-requirements.md
  source_sha256: 72901ac025c9cae11324473243589722b899728254272a0d5756c6ecf32a74bb
  mode: requirements-wide
  not_required_reason: "The requirements resolve the only material tradeoff from executed evidence: the candidate contains no separable DDR behavior, so cut its documentation-only claim and evaluator rather than inventing a model-specific rule."
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "Proves the repaired work-reality mutation, DDR contract, repository tests, documentation checks, and three global skill copies."
model_tier_contract:
  allowed: [medium, large]
  default: medium
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  default_owner: current
  owner_choice: current-or-delegated
  max_owner_decisions_per_slice: 1
  catalog: active-host-plus-user-local
  delegated_fallback: same-tier-then-higher
  automatic_downward_routing: false
  automatic_cross_owner_fallback: false
  amr_required: false
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
  plan_run_worktree_default: true
  internal_integration_target: plan-run-branch
  default_branch_delivery_owner: kb-complete
  allowed_modes: [shared-serial]
plan_run_worktree:
  branch: rehab-closure-rationalization
  workspace_mode: shared-serial
  commit_authorized: true
  commit_authorized_by: user
  commit_approval_ref: "Explicit /w2d invocation on 2026-08-30"
delivery:
  mode: pr
  merge: manual
  post_merge_sync: true
---

# Rehab Closure and DDR Rationalization

## Objective

Remove false rehabilitation blockers without relaxing proof, reject PR #36's
unproven mixed evaluator/DDR claim, and finish with no mixed failing salvage
branch.

## Scope

- `cmd/kbcheck/work_reality.go` and its tests: removal-mode preservation.
- PR #36: close as rejected after the replacement merges.

The epistemic evaluator/corpus, Windows ACL change, a new authority subsystem,
provider-specific restrictions, live model calls, and unpaired-work deletion
are excluded.

## Delivery Contract

This branch is the only replacement delivery candidate. It contains only the
removal-safety slice below. After its exact audited head merges, close PR #36
as rejected and delete its remote branch; its commits remain recoverable by
SHA.

## Gate Ledger

| Gate | Status | Evidence required | Next action |
|---|---|---|---|
| brainstorm-to-plan | passed | Requirements source hash and executed evidence for existing authority, unsafe removal behavior, and PR #36 gate failures | `kb-plan` |
| plan-to-work | passed | One bounded slice, complete requirement mapping, explicit exclusions, and focused deterministic proof | `kb-work docs/plans/2026-08-30-000-kb-rehab-closure-rationalization-manifest.md` |
| slice-closure-001-to-done | passed | `WorkReality` and `RehabRemoval` tests prove ineligible rows are preserved and eligible landed supersession remains removable | `kb-work docs/plans/2026-08-30-000-kb-rehab-closure-rationalization-manifest.md` |
| work-to-complete | passed | `closure-001` proof plus `go run ./cmd/kbcheck core` on `11d23df` (`core: ok checks=39`) | `kb-finalize docs/plans/2026-08-30-000-kb-rehab-closure-rationalization-manifest.md` |
| complete-to-ship | pending | Final review, exact-tree proof, and `local-release` | `kb-ship` |
| delivery | pending | Exact branch head, mergeability, and required GitHub conditions | `kb-land` |

## Slices

| ID | Outcome | Depends on | Tier | Proof |
|---|---|---|---|---|
| closure-001 | Non-removable todo rows survive `--action remove` unchanged | none | medium | targeted `WorkReality` tests |

## Requirement Mapping

| Requirement | Slice / outcome |
|---|---|
| R1-R3 | closure-001 |
| R4, R7-R8 | delivery contract |
| R9 | Already satisfied; close the duplicate session task with the existing-authority tests |

## Constraints

- `remove` may report preservation but may not mutate a row lacking terminal,
  contained, and resolving-artifact proof.
