---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-08-30-rehab-closure-rationalization
source: docs/brainstorms/2026-08-30-rehab-closure-rationalization-requirements.md
source_sha256: 5e7dc7db8a1c09b6580e644f0e9e45ec951702dd230e6389dde70a8c4ef81ee1
created: 2026-08-30
status: planned
workflow_shape: pipeline-change
objective_contract: true
model_tier_contract: true
workspace_isolation_contract: true
proof_governor_contract: true
pre_slice_review_contract: true
pre_slice_review:
  status: not-required
  source: docs/brainstorms/2026-08-30-rehab-closure-rationalization-requirements.md
  source_sha256: 5e7dc7db8a1c09b6580e644f0e9e45ec951702dd230e6389dde70a8c4ef81ee1
  mode: requirements-wide
  not_required_reason: "The requirements resolve the only material tradeoff from executed evidence: retain bounded complexity-aware DDR, and cut the unproven evaluator rather than invent a model-specific rule."
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

Remove false rehabilitation blockers without relaxing proof, preserve only the
useful complexity-aware DDR behavior from PR #36, and finish with no mixed,
failing salvage branch.

## Scope

- `cmd/kbcheck/work_reality.go` and its tests: removal-mode preservation.
- `cmd/kbrouter/dispatch.go` and tests: bounded complexity-aware owner choice.
- `README.md`, `kb-work`, `kb-workflow`, and `ddr_contract_test.go`: one DDR
  policy statement.
- PR #36: close as superseded after the replacement merges.

The epistemic evaluator/corpus, Windows ACL change, a new authority subsystem,
provider-specific restrictions, live model calls, and unpaired-work deletion
are excluded.

## Delivery Contract

This branch is the only replacement delivery candidate. It contains only the
two slices below. After its exact audited head merges, close PR #36 as
superseded and delete its remote branch; its commits remain recoverable by SHA.

## Gate Ledger

| Gate | Status | Evidence required | Next action |
|---|---|---|---|
| brainstorm-to-plan | passed | Requirements source hash and executed evidence for existing authority, unsafe removal behavior, and PR #36 gate failures | `kb-plan` |
| plan-to-work | passed | Two serial slices, complete requirement mapping, explicit exclusions, and focused deterministic proof | `kb-work docs/plans/2026-08-30-000-kb-rehab-closure-rationalization-manifest.md` |
| work-to-complete | pending | Both slice proofs plus `go run ./cmd/kbcheck core` | `kb-finalize` |
| complete-to-ship | pending | Final review, exact-tree proof, and `local-release` | `kb-ship` |
| delivery | pending | Exact branch head, mergeability, and required GitHub conditions | `kb-land` |

## Slices

| ID | Outcome | Depends on | Tier | Proof |
|---|---|---|---|---|
| closure-001 | Non-removable todo rows survive `--action remove` unchanged | none | medium | targeted `WorkReality` tests |
| closure-002 | Complexity-aware DDR is extracted without a second-local retry or evaluator surface | closure-001 | medium | router and DDR contract tests |

## Requirement Mapping

| Requirement | Slice / outcome |
|---|---|
| R1-R3 | closure-001 |
| R4-R6 | closure-002 |
| R7-R9 | closure-002 plus delivery contract |
| R10 | Already satisfied; close the duplicate session task with the existing-authority tests |

## Constraints

- `remove` may report preservation but may not mutate a row lacking terminal,
  contained, and resolving-artifact proof.
- The parent, not a fallback roulette, owns recovery after a local route fails.
- No model name, provider, or historical complaint is a normal-path routing
  rule.
- Do not copy code from PR #36 without a focused test that proves its retained
  behavior.
