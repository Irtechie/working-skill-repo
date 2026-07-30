---
type: kb-manifest
kb_id: kb-2026-07-29-cargo-build-storage-contract
created: 2026-07-29
status: reviewed
workflow_shape: skill-bundle-change
objective_contract: true
blocker_lifecycle_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "Proves the skill contract, repository quality gates, and required global sync targets all pass."
model_tier_contract:
  allowed: [small, medium, large]
  default: medium
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  default_owner: delegated
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
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The user supplied measured duplicate Cargo target directories and named the owning skills."
      - "No unresolved product, architecture, or safety decision remains."
    proof:
      - "user report in session 519f58bf-b4b1-44f3-80f2-1a6c0f24fa43"
      - docs/plans/2026-07-29-001-skill-cargo-build-storage-contract-plan.md
    blockers: []
    passed_at: "2026-07-29T16:02:00Z"
    allowed_next_action: "kb-plan cargo build storage contract"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "docs/plans/2026-07-29-000-kb-cargo-build-storage-contract-manifest.md exists"
      - "docs/plans/2026-07-29-001-skill-cargo-build-storage-contract-plan.md exists"
      - "DAG has no missing blockers or cycles"
      - "The slice has acceptance criteria, expected_files, verification, test_level, functional_risk, model_tier, and proof_check"
    proof:
      - docs/plans/2026-07-29-000-kb-cargo-build-storage-contract-manifest.md
      - docs/plans/2026-07-29-001-skill-cargo-build-storage-contract-plan.md
      - "DAG validation: one slice with no blockers"
      - "Slice contract validation: all required fields present"
    blockers: []
    passed_at: "2026-07-29T16:02:00Z"
    allowed_next_action: "kb-work docs/plans/2026-07-29-000-kb-cargo-build-storage-contract-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Native Cargo storage behavior and public CLI contract pass."
      - "All owning skills preserve the exact project-keyed target."
      - "Required global skill roots match the reviewed source."
    proof:
      - cmd/kbcheck/cargo_storage.go
      - cmd/kbcheck/cargo_storage_test.go
      - cmd/kbcheck/skill_repo_contract_test.go
    proof_commands:
      - "go test ./cmd/kbcheck -run 'CargoStorage|CargoBuildStorage' -count=1"
      - "go vet ./cmd/kbcheck"
      - "go run ./cmd/kbcheck skill-sync-report"
    blockers: []
    passed_at: "2026-07-29T17:43:20Z"
    allowed_next_action: "kb-finalize docs/plans/2026-07-29-000-kb-cargo-build-storage-contract-manifest.md"
  - gate_id: work-to-complete
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Every slice is done or intentionally skipped."
      - "Final verification command/result is recorded."
      - "Scope-verified files and required global synchronization are complete."
    proof:
      - docs/plans/2026-07-29-000-kb-cargo-build-storage-contract-manifest.md
      - docs/plans/2026-07-29-001-skill-cargo-build-storage-contract-plan.md
      - docs/context/operations/testing.md
    proof_commands:
      - "go run ./cmd/kbcheck core"
      - "go run ./cmd/kbcheck local-release --json"
      - "git diff --check"
    blockers: []
    passed_at: "2026-07-29T17:43:20Z"
    allowed_next_action: "kb-finalize docs/plans/2026-07-29-000-kb-cargo-build-storage-contract-manifest.md"
  - gate_id: complete-to-ship
    owner_skill: kb-finalize
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Final kb-check and objective done_check passed."
      - "Multi-agent review has no unresolved P0/P1 findings."
      - "Follow-up resolution, compound, learn, memory refresh, maintenance, and cleanup are recorded."
      - "Required global runtime roots match the reviewed source."
    proof:
      - docs/plans/2026-07-29-000-kb-cargo-build-storage-contract-manifest.md
      - docs/solutions/workflow-issues/preventing-cargo-target-proliferation-in-kb-workflows-2026-07-29.md
      - docs/context/kb/instincts/scoped/skill-maintenance/cargo-storage.yaml
      - docs/context/memory-maintenance.md
      - todo-done.md
    proof_commands:
      - "go run ./cmd/kbcheck local-release --json"
      - "go run ./cmd/kbcheck cargo-storage --action validate --run-id cargo-build-storage-contract --root . --json"
      - "git diff --check"
    blockers: []
    passed_at: "2026-07-29T17:49:00Z"
    allowed_next_action: "kb-complete docs/plans/2026-07-29-000-kb-cargo-build-storage-contract-manifest.md"
slices:
  - id: slice-001
    title: "Enforce stable Cargo build storage across KB workflows"
    path: docs/plans/2026-07-29-001-skill-cargo-build-storage-contract-plan.md
    blockers: []
    verification: tdd
    test_level: unit
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "The edit is text-focused but crosses seven workflow owners, delegated prompts, cleanup semantics, and global propagation."
    model_requirements: ["cross-skill contract reasoning", "deterministic Go test execution", "safe required-root synchronization"]
    escalation_triggers: ["unknown global-only skill drift appears", "the release gate reveals an incompatible cleanup or proof contract"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:cargo-build-storage", "file:cmd/kbcheck/skill_repo_contract_test.go"]
    shared_resources: ["git:integration-owner", "sync:global-skills"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'CargoStorage|CargoBuildStorage' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: changed the seven owning skills, kb-work execution prompt, native cargo-storage command/tests, cross-skill contract, quality exception, README, project/testing/architecture/maintenance docs, plan, manifest, solution/instinct memory, and todo state; review: mode=multi-agent P0=0 P1=17(resolved after dedup) P2=0 P3=0; review-resolution: resolved deterministic resolution, portable runtime, receipt locking/collisions, ownership containment, crash-safe deletion intent, idempotency, exact accounting, capability probing, config fingerprints, authoritative resume, final validation, no-Cargo state, and atomic global propagation; follow-up-resolution: resolved 17, logged 0, blocked 0; proof: focused Cargo tests, full cmd/kbcheck tests, vet, core 38/38, local-release green, sync 138/138, git diff --check; proof-demo: public CLI Go tests, no visual demo required; functional-test: skipped - workflow and CLI policy are covered by public-command Go tests and no rendered UI changed; steering-feedback: current=0 memory=0 observations=5 landmine-candidates=0 instinct-evidence=1; compound: wrote Cargo target proliferation solution and refreshed contributor/release gate cross-link; learn: 1 new scoped instinct, 0 updated; evolve: skipped - completion count 16; memory-impact: durable; kb-map-refresh: done - README and context architecture/operations docs updated; memory-maintenance: counters updated, no new signals; compact: skipped - no startup bloat; cleanup: Cargo not-applicable receipt validated, retained_bytes=0 removed_bytes=0; alerts: memory review recommended because completed-cycle threshold is exceeded"
---

# KB: Cargo Build Storage Contract

## Origin

Source: measured user report showing 19.05 GB of duplicated Cargo outputs under one temporary agent run.

## Workflow Shape

`skill-bundle-change` - the behavior spans verification, debugging, delegated work, repair, finalization, completion, tests, and required global installs.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Enforce stable Cargo build storage across KB workflows | - | tdd | no | in_progress |
