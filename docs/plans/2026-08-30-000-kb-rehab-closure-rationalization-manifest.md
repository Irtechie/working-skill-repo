---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-08-30-rehab-closure-rationalization
source: docs/brainstorms/2026-08-30-rehab-closure-rationalization-requirements.md
source_sha256: 72901ac025c9cae11324473243589722b899728254272a0d5756c6ecf32a74bb
created: 2026-08-30
status: reviewed
workflow_shape: pipeline-change
objective_contract: true
blocker_lifecycle_contract: true
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
  merge: auto-after-checks
  post_merge_sync: true
delivery_authority:
  source: explicit-same-run-user-authorization
  mode: pr
  merge: auto-after-checks
  post_merge_sync: true
  authorized_actions: [commit, push-topic, create-or-update-pr, merge-after-required-checks, sync-installed-skills]
  forbidden_actions: [force-push, bypass-branch-protection, bypass-required-checks, bypass-required-reviews, push-default]
final_audited_shipping_scope:
  - cmd/kbcheck/work_reality.go
  - cmd/kbcheck/work_reality_test.go
  - docs/brainstorms/2026-08-30-rehab-closure-rationalization-requirements.md
  - docs/plans/2026-08-30-000-kb-rehab-closure-rationalization-manifest.md
  - docs/plans/2026-08-30-001-work-reality-remove-safety-plan.md
  - docs/results/code-reviews/kb-rehab-closure-rationalization-20260830-cli.json
  - docs/results/proofs/kb-rehab-closure-20260830-aggregate.md
  - docs/results/proofs/kb-rehab-closure-20260830-closure-001.md
gate_ledger:
  - gate_id: complete-to-ship
    owner_skill: kb-finalize
    gate_scope: release
    status: passed
    required_evidence:
      - "The aggregate deterministic proof passed on the integrated implementation commit."
      - "The focused work-reality tests prove only fully eligible terminal rows are removable."
      - "One CLI-readiness review found no actionable P0-P3 findings."
      - "The final repository release gate passed after durable proof and review metadata were committed."
    proof:
      - docs/results/proofs/kb-rehab-closure-20260830-aggregate.md
      - docs/results/proofs/kb-rehab-closure-20260830-closure-001.md
      - docs/results/code-reviews/kb-rehab-closure-rationalization-20260830-cli.json
      - "go run ./cmd/kbcheck local-release"
    blockers: []
    passed_at: "2026-08-31T04:29:28Z"
    allowed_next_action: "kb-complete docs/plans/2026-08-30-000-kb-rehab-closure-rationalization-manifest.md"
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
| complete-to-ship | passed | Review receipt, durable aggregate proof, and `go run ./cmd/kbcheck local-release` passed on `a6eea37` | `kb-ship` |
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

## Finalization

- Aggregate proof: `docs/results/proofs/kb-rehab-closure-20260830-aggregate.md`.
- Semantic review: `docs/results/code-reviews/kb-rehab-closure-rationalization-20260830-cli.json`
  (`cli-readiness`; P0=0, P1=0, P2=0, P3=0).
- Follow-up: resolved 0, logged 0, blocked 0.
- Knowledge and memory: skipped; this is a bounded correction to existing
  work-reality behavior with no new durable architecture or workflow contract.
- Cleanup: no run-owned ephemeral repository artifacts remained.

Status: reviewed.
