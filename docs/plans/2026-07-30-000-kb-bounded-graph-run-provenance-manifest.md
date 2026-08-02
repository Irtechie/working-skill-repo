---
type: kb-manifest
manifest_schema: 2
kb_id: kb-2026-07-30-bounded-graph-run-provenance
brainstorm_path: docs/context/goals/bounded-graph-run-provenance.md
created: 2026-07-30
status: active
workflow_shape: pipeline-change
objective_contract: true
blocker_lifecycle_contract: true
pre_slice_review_contract: true
context_packet_contract: true
pre_slice_review:
  status: not-required
  source: docs/context/goals/bounded-graph-run-provenance.md
  mode: requirements-wide
  not_required_reason: "The bounded scope was already challenged by adversarial and feasibility reviewers, reduced after the user rejected extra ceremony, and explicitly approved as items 2-6; no product, architecture, trust-boundary, or scope decision remains open."
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "Proves the exact delivery tree satisfies repo-local deterministic checks, release checks, diff hygiene, and required read-only drift reports."
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
    owner_skill: kb-goal
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The durable goal defines bounded storage, receipt, linkage, diagnostics, proof, exclusions, and terminal delivery."
      - "The graph research defines the completion predicate and minimum receipt without requiring a new orchestration lane."
      - "The user explicitly approved items 2-6 and authorized execution, commit, PR creation, and merge."
      - "No unresolved ask-now or research-first decision remains."
    proof:
      - docs/context/goals/bounded-graph-run-provenance.md
      - docs/context/research/2026-07-29-graph-engineering-definition-and-provenance.md
      - todo.md
      - README.md
    blockers: []
    passed_at: "2026-07-30T05:57:14Z"
    allowed_next_action: "kb-plan docs/context/goals/bounded-graph-run-provenance.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The manifest and five slice plans exist."
      - "All five context packets validate."
      - "The dependency DAG has no missing blockers or cycles."
      - "Every slice declares acceptance criteria, expected files, verification, test level, functional risk, model tier, proof check, and HITL classification."
      - "The objective done check and bounded exclusions are preserved."
    proof:
      - docs/plans/2026-07-30-000-kb-bounded-graph-run-provenance-manifest.md
      - docs/plans/2026-07-30-001-tool-graph-run-storage-plan.md
      - docs/plans/2026-07-30-002-tool-node-attempt-receipt-plan.md
      - docs/plans/2026-07-30-003-tool-gate-linked-provenance-plan.md
      - docs/plans/2026-07-30-004-tool-graph-run-diagnostics-plan.md
      - docs/plans/2026-07-30-005-eval-graph-run-fault-injection-plan.md
      - docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-001.json
      - docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-002.json
      - docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-003.json
      - docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-004.json
      - docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-005.json
    blockers: []
    passed_at: "2026-07-30T05:57:14Z"
    allowed_next_action: "kb-work docs/plans/2026-07-30-000-kb-bounded-graph-run-provenance-manifest.md"
slices:
  - id: slice-001
    title: "Bound graph-run storage and retention"
    path: docs/plans/2026-07-30-001-tool-graph-run-storage-plan.md
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Deletion safety crosses path containment, ownership markers, concurrent state, byte accounting, and active-run preservation; a false positive could destroy evidence."
    model_requirements: ["Go filesystem safety", "cross-platform locking", "atomic receipt persistence", "adversarial cleanup testing"]
    escalation_triggers: ["cleanup can reach an unowned path", "active or pinned state lacks an authoritative sensor", "symlink containment cannot be proven", "retention requires a daemon"]
    context_packet_path: docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-001.json
    workspace_mode: shared-serial
    conflict_domains: ["file:cmd/kbcheck/graph_run_storage.go", "file:cmd/kbcheck/main.go", "namespace:.kb/runs"]
    shared_resources: ["git:integration-owner", "storage:.kb/runs"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'GraphRunStorage' -count=1"
      expect: 0
    hitl: false
    status: in_progress
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Write protected cleanup-safety tests, then implement marker-owned accounting, dry-run retention, and guarded apply."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
  - id: slice-002
    title: "Emit immutable node-attempt receipts"
    path: docs/plans/2026-07-30-002-tool-node-attempt-receipt-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "The receipt must unify existing route, lease, revision, dependency, and proof evidence without trusting self-reported fields or duplicating orchestration state."
    model_requirements: ["Go contract design", "hash and attestation validation", "existing execution telemetry fluency", "bounded metadata review"]
    escalation_triggers: ["receipt requires raw prompts or outputs", "authoritative revision or lease fields cannot be bound", "attempt identity conflicts across existing ledgers", "emission mutates accepted receipts"]
    context_packet_path: docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-002.json
    workspace_mode: shared-serial
    conflict_domains: ["file:internal/modelrouting/node_attempt_receipt.go", "file:cmd/kbcheck/execution_telemetry.go", "namespace:.kb/runs"]
    shared_resources: ["git:integration-owner", "storage:.kb/runs"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck ./internal/modelrouting -run 'NodeAttemptReceipt|ExecutionTelemetry' -count=1"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Freeze the versioned bounded receipt contract, prove tampering and raw-payload rejection RED, then wire atomic per-attempt emission."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
  - id: slice-003
    title: "Link node receipts to manifest gates"
    path: docs/plans/2026-07-30-003-tool-gate-linked-provenance-plan.md
    blockers: [slice-002]
    verification: tdd
    test_level: integration
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Gate advancement is the completion authority, so receipt linkage must validate dependency order, revisions, hashes, terminal status, and proof without allowing telemetry to self-certify done."
    model_requirements: ["manifest gate contract reasoning", "DAG dependency validation", "artifact hash verification", "backward-compatible YAML parsing"]
    escalation_triggers: ["legacy manifests become unreadable", "gate passage can occur with failed or stale proof", "receipt linkage requires rewriting unrelated gate history", "fan-in evidence cannot identify contributing receipts"]
    context_packet_path: docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-003.json
    workspace_mode: shared-serial
    conflict_domains: ["file:cmd/kbcheck/manifest_contract.go", "file:cmd/kbcheck/swarm.go", "namespace:gate-ledger"]
    shared_resources: ["git:integration-owner"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'Manifest.*Receipt|Gate.*Receipt|NodeAttempt.*Gate' -count=1"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add receipt-reference gate fixtures, then enforce terminal, freshness, dependency, and proof bindings while preserving legacy manifests."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
  - id: slice-004
    title: "Explain graph-run failure and incompletion"
    path: docs/plans/2026-07-30-004-tool-graph-run-diagnostics-plan.md
    blockers: [slice-003]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Causal diagnosis must reconcile DAG state, attempt receipts, proof, gates, and fan-in without mistaking downstream symptoms for the first failure."
    model_requirements: ["deterministic graph traversal", "compact CLI and JSON design", "completion predicate reasoning", "corrupt-evidence handling"]
    escalation_triggers: ["multiple causal roots cannot be represented deterministically", "diagnostics need raw transcript ingestion", "completion output disagrees with manifest-contract", "corrupt evidence is silently ignored"]
    context_packet_path: docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-004.json
    workspace_mode: shared-serial
    conflict_domains: ["file:cmd/kbcheck/graph_run.go", "file:cmd/kbcheck/main.go", "namespace:graph-run-cli"]
    shared_resources: ["git:integration-owner"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'GraphRunInspect' -count=1"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Protect text and JSON CLI scenarios, then implement first-causal-failure and why-not-done projections over validated receipts and gates."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
  - id: slice-005
    title: "Prove causal diagnosis, bounded retry, and completion"
    path: docs/plans/2026-07-30-005-eval-graph-run-fault-injection-plan.md
    blockers: [slice-004]
    verification: functional
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "The final proof must inject failures across storage, attempts, dependencies, and gates, then show accepted nodes are not replayed and completion becomes true only after causal repair."
    model_requirements: ["deterministic fixture design", "retry-state modeling", "release-gate interpretation", "documentation and proof synthesis"]
    escalation_triggers: ["fixture expectations are not mechanically consumed", "retry replays an accepted node", "local-release fails outside the named Windows harness blocker", "proof depends on live model calls"]
    context_packet_path: docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-005.json
    workspace_mode: shared-serial
    conflict_domains: ["path:evals/graph-run", "file:cmd/kbcheck/graph_run_selftest.go", "file:cmd/kbcheck/checks.go", "docs:graph-run"]
    shared_resources: ["git:integration-owner", "release:local-release"]
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck graph-run-selftest"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Create mechanically consumed fault fixtures, prove repair and retry behavior, update command/testing docs, then run core and local-release."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
---

# KB: Bounded Graph-Run Provenance

## Origin

Goal: `docs/context/goals/bounded-graph-run-provenance.md`

Research: `docs/context/research/2026-07-29-graph-engineering-definition-and-provenance.md`

## Workflow Shape

`pipeline-change` - the work adds a bounded runtime evidence contract, safe
storage lifecycle, gate validation, operator diagnostics, and deterministic
release proof without creating a new workflow lane.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Bound graph-run storage and retention | - | tdd | no | in_progress |
| 2 | Emit immutable node-attempt receipts | slice-001 | tdd | no | pending |
| 3 | Link node receipts to manifest gates | slice-002 | tdd | no | pending |
| 4 | Explain graph-run failure and incompletion | slice-003 | tdd | no | pending |
| 5 | Prove causal diagnosis, bounded retry, and completion | slice-004 | functional | no | pending |

## Completion Predicate

```text
all required nodes terminal
+ every required acceptance criterion has valid proof
+ required fan-in/review gates passed
+ no unresolved blocking edge remains
```

## Scope Boundaries

- Store metadata and hashes only; never persist prompts, outputs, transcripts,
  diffs, screenshots, or exhaustive tool-call traces.
- Preserve active, pinned, corrupt, and unowned paths during retention.
- Reuse execution telemetry, leases, proof receipts, and gate ledgers rather
  than creating a second orchestration or observability platform.
- Do not add a collector, backend, daemon, runtime hook, persistent agent
  organization, context-economics benchmark, or `kb-graph` lane.
- Keep the separate Windows harness-validation recovery workstream outside this
  implementation unless its existing failure is the only remaining delivery
  blocker after fresh proof.
