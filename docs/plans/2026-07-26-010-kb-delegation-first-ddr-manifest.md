---
type: kb-manifest
kb_id: kb-2026-07-26-delegation-first-ddr
title: Delegation-first DDR owner gate
created: 2026-07-26
status: completed
workflow_shape: skill-bundle-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "Proves the runtime owner gate, skill contract, repository checks, and required global skill copies agree."
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
  amr_required: false
  automatic_cross_owner_fallback: false
gate_ledger:
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest and slice plan exist"
      - "slice has acceptance criteria, expected files, proof, test level, risk, and model tier"
      - "current execution has bounded exception reasons"
      - "no-qualified-route requires live route lookup"
    proof:
      - docs/plans/2026-07-26-010-kb-delegation-first-ddr-manifest.md
      - docs/plans/2026-07-26-011-delegation-first-owner-gate-plan.md
      - docs/plans/2026-07-26-delegation-first-ddr-context/ddr-001.json
      - internal/modelrouting/selector_test.go
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-010-kb-delegation-first-ddr-manifest.md"
  - gate_id: slice-ddr-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "current-owner reason gate passes focused selector and CLI tests"
      - "no-qualified-route is rejected when an eligible route exists"
      - "portable per-slice tier selection and parallel ready-set policy are explicit"
      - "Windows junction and canonical project identity are equivalent"
      - "required global skill copies match source"
    proof:
      - internal/modelrouting/selector_test.go
      - internal/modelrouting/identity_windows_test.go
      - cmd/kbrouter/select_test.go
      - cmd/kbcheck/ddr_contract_test.go
      - "focused canonical and junction tests passed"
      - "skill-sync-report: 138 comparisons, 0 required issues"
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-finalize docs/plans/2026-07-26-010-kb-delegation-first-ddr-manifest.md"
  - gate_id: complete-to-ship
    owner_skill: kb-finalize
    status: passed
    required_evidence:
      - "all slices have terminal proof"
      - "core, diff, sync, and model-routing release checks pass"
    proof:
      - "local-release: ok=true required_failures=0"
      - "kb-check-all: passed"
      - "git-diff-check: passed"
      - "skill-sync-report: passed"
      - "model-routing-initial-pilot: passed, not promoted"
    blockers: []
    passed_at: "2026-07-26T20:32:53-04:00"
    allowed_next_action: "direct integration authorized by repository owner"
slices:
  - id: ddr-001
    title: "Enforce delegation-first ownership with bounded current exceptions"
    path: docs/plans/2026-07-26-011-delegation-first-owner-gate-plan.md
    context_packet_path: docs/plans/2026-07-26-delegation-first-ddr-context/ddr-001.json
    status: done
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "The behavior is bounded to an existing Go selector and documented CLI contract, with deterministic tests."
    model_requirements: ["Go implementation", "CLI contract reasoning", "focused deterministic tests"]
    escalation_triggers: ["compatibility callers require unbounded reason text", "native-host and CLI catalogs cannot preserve distinct authority"]
    workspace_mode: shared-serial
    conflict_domains: ["internal/modelrouting", "cmd/kbrouter", ".github/skills/kb-work", ".github/skills/kb-plan", "routing docs"]
    shared_resources: ["git:integration-owner", "sync:global-skills"]
    proof_check:
      kind: command_exit
      command: "go test ./internal/modelrouting ./cmd/kbrouter ./cmd/kbcheck"
      expect: 0
    hitl: false
    owner: agent
    can_continue_other_slices: true
---

# Delegation-first DDR owner gate

## Objective

Keep planning and supervision on the orchestrator while making one qualified
subagent the normal execution owner. Let the orchestrator retain execution only
for a recognized, evidence-backed exception.

## Acceptance

- The planned tier remains the minimum execution capability.
- The active host resolves that portable tier to an exact qualified model only
  when it picks up the slice; the plan never freezes a provider or model.
- A bounded slice normally selects exactly one qualified same-tier-or-higher
  subagent.
- Independent ready slices select their own tier-qualified subagents and run in
  parallel when dependencies, writes, and shared resources are isolated.
- Current execution accepts only recognized reason codes.
- `no-qualified-route` cannot pass while the CLI catalog contains an eligible
  route; the skill also requires checking the active host surface.
- Windows junction and canonical repo paths resolve to the same project
  identity, while missing paths still fail closed.
- There is no automatic downward route or silent cross-owner fallback.
- Deterministic proof, not the routing receipt, accepts the work.
