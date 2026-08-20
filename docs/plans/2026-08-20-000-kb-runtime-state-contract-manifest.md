---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-08-20-runtime-state-contract
source: docs/brainstorms/2026-08-20-kb-runtime-cognitive-routing-requirements.md
source_sha256: 7407e7b016730501bac24e4d73361e408ef22eb8c2078736c533681beb2117f3
created: 2026-08-20
status: draft
workflow_shape: pipeline-change
objective_contract: true
model_tier_contract: true
workspace_isolation_contract: true
proof_governor_contract: true
pre_slice_review_contract: true
pre_slice_review:
  status: not-required
  source: docs/brainstorms/2026-08-20-kb-runtime-cognitive-routing-requirements.md
  source_sha256: 7407e7b016730501bac24e4d73361e408ef22eb8c2078736c533681beb2117f3
  mode: requirements-wide
  not_required_reason: User requirements are explicit, internally consistent, and contain no unresolved ask-now or research-first decision.
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: Proves the runtime contract, skills, documentation, and required global-copy drift agree.
model_tier_contract:
  allowed: [small, medium, large]
  default: large
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
  plan_run_worktree_default: true
  internal_integration_target: plan-run-branch
  default_branch_delivery_owner: kb-complete
  allowed_modes: [shared-serial]
delivery:
  mode: pr
  merge: manual
plan_run_worktree:
  default: plan-run-serial
  parallel_opt_in_required: true
---

# Runtime State Contract

## Objective

Extract inner-loop lifecycle state from conversational skill prose into a
versioned, deterministic runtime contract while preserving thin phase adapters.

## Gate Ledger

| Gate | Status | Evidence required | Next action |
|---|---|---|---|
| brainstorm-to-plan | passed | user requirements source; no open product decision | `kb-plan` complete |
| plan-to-work | needs-human | approval of this manifest; context packets; contract validation | `kb-work <manifest>` |

## Slices

| ID | Outcome | Depends on | Tier | Proof |
|---|---|---|---|---|
| runtime-001 | Versioned run-state and transition schema | — | large | focused Go contract tests |
| runtime-002 | Tool-facing transition executor and Markdown boundary projection | runtime-001 | large | CLI/integration fixtures |

## Constraints

- Preserve separate planning, work, finalization, and delivery authority.
- Structured records own heartbeats, blockers, proof reuse, and state changes.
- Markdown is produced only at user/reviewer boundaries.
- No provider routes, credentials, or model lists enter repository state.

## Approval Needed Before Work

Approve this manifest together with the sibling manifests in
`docs/context/epics/kb-runtime-cognitive-routing.md`.
