---
type: kb-manifest
manifest_schema: 2
kb_id: kb-2026-07-30-model-tier-qualification
brainstorm_path: docs/brainstorms/2026-07-30-model-tier-qualification-requirements.md
created: 2026-07-30
status: active
workflow_shape: pipeline-change
objective_contract: true
blocker_lifecycle_contract: true
pre_slice_review_contract: true
context_packet_contract: true
pre_slice_review:
  status: passed
  source: docs/brainstorms/2026-07-30-model-tier-qualification-requirements.md
  source_sha256: e145f8f291e90e19b283af1ca150723ceb45560a202d06da82fc2b136cd007a0
  mode: requirements-wide
  review_id: model-tier-qualification-requirements-20260730
  reviewed_at: "2026-07-30T17:58:46-04:00"
  review_artifact: docs/results/document-reviews/model-tier-qualification-requirements-e145f8f291e9.json
  review_artifact_sha256: 39a020e0c0cf9c56736d475eb66b4f1dc7738ff4e86c96442f426d5f1d444d32
  persona_evidence_json: '{"coherence-reviewer":"consistency-risk: admission, exclusion, proof, and qualification terms must produce one deterministic state machine","feasibility-reviewer":"delivery-risk: strict cross-repository receipts and plan-sufficiency records must be mechanically representable","product-lens-reviewer":"product-risk: a reusable classifier must not turn a bounded cohort into a broad routing claim","spec-flow-analyzer":"flow-risk: preregistration, admission gates, exclusions, proof, and final decisions form a multi-state evidence lifecycle","security-lens-reviewer":"security-risk: live route evidence crosses trust boundaries while secrets, protected oracles, and host state must remain excluded","scope-guardian-reviewer":"scope-risk: planning, scoring, fixtures, documentation, and TokenZoom integration need explicit independent milestones","adversarial-document-reviewer":"adversarial-risk: selective cohorts, replay, correlated fixtures, forged receipts, and provider drift could falsely qualify a route"}'
  selected_personas_json: '["coherence-reviewer","feasibility-reviewer","product-lens-reviewer","spec-flow-analyzer","security-lens-reviewer","scope-guardian-reviewer","adversarial-document-reviewer"]'
  completed_personas_json: '["coherence-reviewer","feasibility-reviewer","product-lens-reviewer","spec-flow-analyzer","security-lens-reviewer","scope-guardian-reviewer","adversarial-document-reviewer"]'
  failed_personas_json: '[]'
  findings_resolved: 28
  unresolved_p0: 0
  unresolved_p1: 0
  residual_findings: 0
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "Proves the qualification-plan contract, experimental tier scorer, deterministic corpus, docs, diff hygiene, and required skill-copy drift gates agree."
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
    owner_skill: kb-brainstorm
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The requirements define model, plan, oracle, schema, route, and infrastructure attribution."
      - "The Medium capability and 30-fixture threshold were frozen independently of Deepseek4 outcomes."
      - "Question Gate classification exists with no unresolved ask-now or research-first item."
      - "The requirements-wide review receipt has no unresolved P0/P1 and no failed selected persona."
    proof:
      - docs/brainstorms/2026-07-30-model-tier-qualification-requirements.md
      - docs/results/document-reviews/model-tier-qualification-requirements-e145f8f291e9.json
      - docs/context/epics/ddr-planning-validation.md
      - todo.md
    blockers: []
    passed_at: "2026-07-30T18:00:00-04:00"
    allowed_next_action: "kb-plan docs/brainstorms/2026-07-30-model-tier-qualification-requirements.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The manifest, two slice plans, and two bounded context packets exist."
      - "The serial dependency DAG has no missing blockers or cycles."
      - "Both slices declare acceptance criteria, expected files, verification, test level, functional risk, model tier, proof check, and HITL classification."
      - "The portable experimental milestone is independent of TokenZoom live evidence."
      - "The objective done check remains the canonical local-release gate."
    proof:
      - docs/plans/2026-07-30-010-kb-model-tier-qualification-manifest.md
      - docs/plans/2026-07-30-011-tool-qualification-plan-receipt-plan.md
      - docs/plans/2026-07-30-012-eval-model-tier-classifier-plan.md
      - docs/plans/2026-07-30-model-tier-qualification-context/slice-001.json
      - docs/plans/2026-07-30-model-tier-qualification-context/slice-002.json
    blockers: []
    passed_at: "2026-07-30T18:00:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-30-010-kb-model-tier-qualification-manifest.md"
slices:
  - id: slice-001
    title: "Emit and validate qualification-plan receipts"
    path: docs/plans/2026-07-30-011-tool-qualification-plan-receipt-plan.md
    blockers: []
    verification: tdd
    test_level: integration
    functional_risk: broad
    model_tier: large
    model_tier_reason: "This changes planning policy and the causal boundary between plan defects and model failures; weak validation would corrupt every later tier decision."
    model_requirements: ["Go contract design", "YAML and strict JSON validation", "planning-policy reasoning", "adversarial fixture design"]
    escalation_triggers: ["ordinary non-qualification plans become invalid", "semantic model judgment becomes a deterministic pass condition", "legacy manifests become unreadable", "a new DDR-specific planner is proposed"]
    context_packet_path: docs/plans/2026-07-30-model-tier-qualification-context/slice-001.json
    workspace_mode: shared-serial
    conflict_domains: ["file:.github/skills/kb-plan/SKILL.md", "file:cmd/kbcheck/manifest_contract.go", "namespace:qualification-plan"]
    shared_resources: ["git:integration-owner", "sync:global-skills"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'QualificationPlan|Manifest.*Qualification' -count=1"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Write protected qualification-plan contract tests, then add the opt-in kb-plan record and kbcheck validation."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
  - id: slice-002
    title: "Classify model tier evidence deterministically"
    path: docs/plans/2026-07-30-012-eval-model-tier-classifier-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "The scorer makes reusable capability claims from hostile cross-repository evidence; state, trust, statistics, redaction, and replay rules must fail closed."
    model_requirements: ["Go strict-input security", "receipt and signature verification", "deterministic state-machine design", "statistical threshold reasoning", "CLI contract testing"]
    escalation_triggers: ["the scorer needs network or inference access", "private keys or endpoints enter repository state", "unsupported attestations can produce qualified", "routing promotion is coupled to classification"]
    context_packet_path: docs/plans/2026-07-30-model-tier-qualification-context/slice-002.json
    workspace_mode: shared-serial
    conflict_domains: ["file:cmd/kbcheck/main.go", "namespace:model-tier-eval", "path:evals/model-tier-qualification"]
    shared_resources: ["git:integration-owner"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'ModelTierEval' -count=1"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Freeze RED fixtures for every decision and exclusion path, then implement the offline experimental scorer and docs."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
---

# Model Tier Qualification

## Origin

Brainstorm: `docs/brainstorms/2026-07-30-model-tier-qualification-requirements.md`

## Workflow Shape

`pipeline-change` - planning policy, deterministic validation, scorer behavior,
eval fixtures, documentation, and required skill-copy propagation must agree.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Emit and validate qualification-plan receipts | - | tdd / integration | no | pending |
| 2 | Classify model tier evidence deterministically | slice-001 | tdd / functional-cli | no | pending |

## External Integration

TokenZoom owns the Deepseek4 evidence export and bounded Medium decision. That
integration is tracked by `docs/context/epics/ddr-planning-validation.md` and is
not a blocker for the experimental portable scorer.
