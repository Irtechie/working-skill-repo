---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-08-20-surface-audit-and-amr-retirement
source: docs/brainstorms/2026-08-20-kb-runtime-cognitive-routing-requirements.md
source_sha256: 7407e7b016730501bac24e4d73361e408ef22eb8c2078736c533681beb2117f3
created: 2026-08-20
status: draft
workflow_shape: multi-stream-epic
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
  not_required_reason: Audit criteria are explicit; any skill retirement needs later evidence and separate user approval.
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: Proves removal parity, active-route absence, serial policy, and install drift.
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

# Skill-Surface Audit, Serial-First Routing, and AMR Retirement

## Objective

Measure active skill ownership, callers, runtime cost, and behavioral proof
before any consolidation; make serial execution and deterministic routing the
default; retire active AMR routes.

## Gate Ledger

| Gate | Status | Evidence required | Next action |
|---|---|---|---|
| plan-to-work | needs-human | prior manifests passed; epic approval | `kb-work <manifest>` |

## Slices

| ID | Outcome | Depends on | Tier | Proof |
|---|---|---|---|---|
| audit-001 | Deterministic compact router and serial-first executor policy | routing-cognitive-delivery | large | route/worktree fixtures |
| audit-002 | Audit TDD and other candidate skills; remove only active AMR routes | audit-001 | large | parity matrix and no-active-route tests |

## Constraints

- Do not merge eval setup into runtime checks or research into memory repair.
- Do not delete an entrypoint before its caller/behavior parity is tested and
  the user separately approves its removal.
- Preserve historical AMR artifacts as history, not live functionality.
