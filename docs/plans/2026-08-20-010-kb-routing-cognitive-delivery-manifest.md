---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-08-20-routing-cognitive-delivery
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
  not_required_reason: User requirements define the delivery and routing outcomes without an unresolved product decision.
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: Proves routing, cognitive-skill propagation, delivery policy, and sync drift.
model_tier_contract:
  allowed: [small, medium, large]
  default: large
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
  plan_run_worktree_default: true
  internal_integration_target: plan-run-branch
  default_branch_delivery_owner: kb-complete
  allowed_modes: [shared-serial]
delivery: {mode: pr, merge: manual}
plan_run_worktree: {default: plan-run-serial, parallel_opt_in_required: true}
---

# DDR Routing, Cognitive Output, and Delivery

## Objective

Enforce lowest-qualified live DDR delegation, rename the cognitive-format owner,
and make delivery output reviewer-friendly without adding inner-loop prose.

## Gate Ledger

| Gate | Status | Evidence required | Next action |
|---|---|---|---|
| plan-to-work | needs-human | epic approval; runtime-state contract available | `kb-work <manifest>` |

## Slices

| ID | Outcome | Depends on | Tier | Proof |
|---|---|---|---|---|
| routing-001 | Exact-tier-or-higher DDR route/exception receipt is enforced | runtime-state manifest | large | router and kbcheck fixtures |
| routing-002 | `kb-cognitive` replaces `kb-compact` and drives boundary presentation | routing-001 | medium | communication and sync tests |
| routing-003 | Solo/collaborative delivery endpoint contract is made explicit | routing-002 | large | delivery-route fixtures |

## Constraints

- Plans name capabilities, never concrete models or private transports.
- Native and CLI catalogs remain separate unless an adapter proves callability.
- `kb-simplify` remains explicit-only.
