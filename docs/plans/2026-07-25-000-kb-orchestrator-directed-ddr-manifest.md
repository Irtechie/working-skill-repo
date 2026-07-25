---
type: kb-manifest
kb_id: kb-2026-07-25-orchestrator-directed-ddr
title: Orchestrator-directed DDR manifest
created: 2026-07-25
status: reviewed
workflow_shape: skill-bundle-change
workspace_mode: shared-serial
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "Proves repo-local checks and all required global sync targets are release-ready."
model_tier_contract:
  allowed: [small, medium, large]
  default: medium
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  owner_choice: current-or-delegated
  max_owner_decisions_per_slice: 1
  catalog: active-host-plus-user-local
  delegated_fallback: same-tier-then-higher
  automatic_downward_routing: false
  amr_required: false
  automatic_cross_owner_fallback: false
gate_ledger:
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest path exists"
      - "all slice plan paths exist"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk, model_tier"
      - "objective contract has a done_check and each slice has a proof_check"
    proof:
      - docs/plans/2026-07-25-000-kb-orchestrator-directed-ddr-manifest.md
      - docs/plans/2026-07-25-001-orchestrator-owner-selector-plan.md
      - docs/plans/2026-07-25-002-host-aware-work-routing-plan.md
      - docs/plans/2026-07-25-003-ddr-release-sync-plan.md
      - docs/plans/2026-07-25-orchestrator-directed-ddr-context/ddr-001.json
    blockers: []
    passed_at: "2026-07-25"
    allowed_next_action: "kb-work docs/plans/2026-07-25-000-kb-orchestrator-directed-ddr-manifest.md"
  - gate_id: slice-ddr-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "owner-first selector and CLI contract pass focused tests"
      - "dispatch revalidates the complete delegated capability envelope"
      - "real discovered current route retains work without fabricated App capability"
    proof:
      - internal/modelrouting/selector_test.go
      - cmd/kbrouter/dispatch_test.go
      - cmd/kbrouter/catalog_test.go
    blockers: []
    passed_at: "2026-07-25"
    allowed_next_action: "kb-work ddr-002"
  - gate_id: slice-ddr-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "normal skills exclude AMR and use one owner decision"
      - "native App and CLI/user-local execution branches are distinct"
      - "manifest safety fields reject downward and cross-owner fallback"
    proof:
      - cmd/kbcheck/ddr_contract_test.go
      - cmd/kbcheck/manifest_contract_test.go
      - "independent ddr_final_review: all five findings resolved"
    blockers: []
    passed_at: "2026-07-25"
    allowed_next_action: "kb-work ddr-003"
  - gate_id: slice-ddr-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "core contributor gate passes"
      - "required global skill roots hash-match source"
      - "local release gate passes with AMR still not promoted"
    proof:
      - todo-done.md
      - config/skill-quality.json
      - docs/plans/2026-07-25-000-kb-orchestrator-directed-ddr-manifest.md
    blockers: []
    passed_at: "2026-07-25"
    allowed_next_action: "kb-finalize docs/plans/2026-07-25-000-kb-orchestrator-directed-ddr-manifest.md"
  - gate_id: complete-to-ship
    owner_skill: kb-finalize
    status: passed
    required_evidence:
      - "all slices have terminal proof"
      - "independent review has no remaining blocker"
      - "release and sync gates pass"
    proof:
      - "slice-ddr-001-to-done"
      - "ddr_final_review follow-up"
      - "local-release ok=true"
    blockers: []
    passed_at: "2026-07-25"
    allowed_next_action: "direct integration authorized by repository owner"
slices:
  - id: ddr-001
    title: "Orchestrator ownership selector"
    path: docs/plans/2026-07-25-001-orchestrator-owner-selector-plan.md
    status: done
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: narrow
    model_tier: large
    model_tier_reason: "Selector semantics span runtime capability, CLI receipts, and compatibility validation."
    model_requirements: ["Go implementation", "CLI contract reasoning", "compatibility judgment"]
    escalation_triggers: ["archived manifests fail", "unrelated selector callers require semantic changes"]
    workspace_mode: worktree-required
    conflict_domains: ["internal/modelrouting", "cmd/kbrouter", "cmd/kbcheck"]
    shared_resources: ["git:integration-owner"]
    proof_check:
      kind: command_exit
      command: "go test ./internal/modelrouting ./cmd/kbrouter ./cmd/kbcheck"
      expect: 0
  - id: ddr-002
    title: "Host-aware work routing"
    path: docs/plans/2026-07-25-002-host-aware-work-routing-plan.md
    status: done
    blockers: []
    verification: verification-only
    test_level: functional-cli
    functional_risk: narrow
    model_tier: large
    model_tier_reason: "The policy controls every later worker selection and must reconcile active host and CLI capability surfaces."
    model_requirements: ["cross-skill consistency", "host capability reasoning", "documentation contract review"]
    escalation_triggers: ["active host schema is unavailable", "global skill drift conflicts with source policy"]
    workspace_mode: worktree-required
    conflict_domains: [".github/skills", "README.md", "LOCAL_MODELS.example.md", "docs/context/architecture"]
    shared_resources: ["git:integration-owner", "sync:global-skills"]
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck core"
      expect: 0
  - id: ddr-003
    title: "DDR release and sync"
    path: docs/plans/2026-07-25-003-ddr-release-sync-plan.md
    status: done
    blockers: [ddr-001, ddr-002]
    verification: verification-only
    test_level: full
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Release reconciliation requires repository ownership, dirty-baseline judgment, and exact sync control."
    model_requirements: ["Git integration authority", "release gate interpretation", "exact-path sync"]
    escalation_triggers: ["release gate times out", "global drift conflicts", "push is rejected"]
    workspace_mode: shared-serial
    conflict_domains: ["git:main", "global skill installs"]
    shared_resources: ["git:integration-owner", "sync:global-skills"]
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck local-release"
      expect: 0
---

# Orchestrator-directed DDR

## Objective

Make the active orchestrator decide once per slice whether to retain execution
or delegate to one qualified worker. Keep AMR available only as an experimental
benchmark; it is not part of the normal production path.

## Slices

| ID | Plan | Tier | Context packet | Depends on | Done check |
| --- | --- | --- | --- | --- | --- |
| DDR-001 | [Selector contract](2026-07-25-001-orchestrator-owner-selector-plan.md) | large | [packet](2026-07-25-orchestrator-directed-ddr-context/ddr-001.json) | none | Focused selector, CLI, and manifest contract tests |
| DDR-002 | [Host-aware workflow](2026-07-25-002-host-aware-work-routing-plan.md) | large | [packet](2026-07-25-orchestrator-directed-ddr-context/ddr-002.json) | DDR-001 contract | Skill and documentation checks |
| DDR-003 | [Release and sync](2026-07-25-003-ddr-release-sync-plan.md) | large | [packet](2026-07-25-orchestrator-directed-ddr-context/ddr-003.json) | DDR-001, DDR-002 | `kbcheck core`, `kbcheck local-release`, exact-path sync hashes |

## Routing decision

The current orchestrator retains cross-file contract ownership because this
slice requires the session's accumulated reconciliation context and release
authority. The selector implementation is delegated once to one bounded worker.
No worker is asked to run an AMR trial or prove a lower tier before executing.

## Acceptance

- Every normal selection declares `current` or `delegated`.
- `current` is capability-checked and wins without searching workers.
- `delegated` selects exactly one qualified route.
- Neither owner silently falls back to the other.
- Active host capability comes from its callable schema or live user-local
  catalog, not from a model's remembered product names.
- AMR remains visibly experimental and never gates normal work.
