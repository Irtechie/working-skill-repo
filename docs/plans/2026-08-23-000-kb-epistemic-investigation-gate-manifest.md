---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-08-23-epistemic-investigation-gate
source: docs/brainstorms/2026-08-23-epistemic-investigation-gate-requirements.md
source_sha256: 54a5ed6875dd22642deabe792e19c87189acd2acccead071711f806a0a8f1db6
created: 2026-08-23
status: planned
workflow_shape: pipeline-change
objective_contract: true
blocker_lifecycle_contract: true
model_tier_contract: true
workspace_isolation_contract: true
pre_slice_review_contract: true
pre_slice_review:
  status: not-required
  source: docs/brainstorms/2026-08-23-epistemic-investigation-gate-requirements.md
  source_sha256: 54a5ed6875dd22642deabe792e19c87189acd2acccead071711f806a0a8f1db6
  mode: requirements-wide
  not_required_reason: Requirements are explicit, independently scoreable, and contain no unresolved planning question.
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck skill-eval-regression --run-root .kb/eval-runs/epistemic-treatment --baseline evals/skill-eval/baselines/epistemic-investigation.json --output docs/results/epistemic-investigation-comparison.json"
  expect: 0
  why: Reuses the existing regression owner to enforce protected epistemic and workflow-ease metrics for an interleaved control/treatment comparison anchored by the untouched baseline.
model_tier_contract:
  allowed: [small, medium, large]
  default: large
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
      - "Hash-bound requirements source exists"
      - "Existing skill-eval owners and extension-first constraint are recorded"
      - "Ease, false-positive investigation, oracle isolation, claim bounds, and live authority are explicit"
    proof:
      - docs/brainstorms/2026-08-23-epistemic-investigation-gate-requirements.md
      - "Existing-owner constraint section"
      - "Cognitive-load, Fixture Integrity, and Scope Boundaries sections"
    blockers: []
    passed_at: "2026-08-23T20:10:53-04:00"
    allowed_next_action: "kb-plan docs/brainstorms/2026-08-23-epistemic-investigation-gate-requirements.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Manifest and five slice plans exist"
      - "DAG has no missing blockers or cycles"
      - "Every slice has acceptance criteria, expected files, proof, tier, escalation, and HITL classification"
      - "Planning-specific treatment is separated from post-promotion deterministic enforcement"
    proof:
      - docs/plans/2026-08-23-000-kb-epistemic-investigation-gate-manifest.md
      - docs/plans/2026-08-23-001-skill-eval-epistemic-extension-plan.md
      - docs/plans/2026-08-23-002-epistemic-untouched-baseline-plan.md
      - docs/plans/2026-08-23-003-epistemic-investigation-treatment-plan.md
      - docs/plans/2026-08-23-004-epistemic-matched-replay-plan.md
      - docs/plans/2026-08-23-005-planning-assurance-enforcement-plan.md
    blockers: []
    passed_at: "2026-08-23T20:39:46-04:00"
    allowed_next_action: "kb-work docs/plans/2026-08-23-000-kb-epistemic-investigation-gate-manifest.md"
slices:
  - id: epistemic-001
    title: Extend skill-eval with one sealed epistemic path
    path: docs/plans/2026-08-23-001-skill-eval-epistemic-extension-plan.md
    blockers: []
    verification: tdd
    test_level: integration
    functional_risk: broad
    model_tier: large
    model_tier_reason: This slice changes the existing result, scorer, adapter, and regression contracts while preserving all current route-eval behavior.
    model_requirements: [Go compatibility design, deterministic scoring, isolated process workspaces, experimental-design reasoning]
    escalation_triggers: [existing v1 fixtures regress, oracle enters the actor prompt/workspace, a new top-level command appears without a RED necessity proof, current plan-run collision is detected]
    workspace_mode: shared-serial
    conflict_domains: [namespace:skill-eval, file:cmd/kbcheck/skill_eval.go, file:cmd/kbcheck/eval_adapters.go, file:evals/skill-eval/result.schema.json, path:evals/skill-eval/epistemic]
    shared_resources: [git:integration-owner, eval:skill-eval, worktree:plan-run]
    context_packet_path: docs/plans/2026-08-23-epistemic-investigation-context/epistemic-001.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'SkillEvalEpistemic|SkillEvalBaseline' -count=1"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Recheck worktree/lease overlap, write protected compatibility and epistemic tests, prove RED, then extend the existing owners."
    human_action: ""
    can_continue_other_slices: false
  - id: epistemic-002
    title: Freeze the corpus and capture the untouched baseline
    path: docs/plans/2026-08-23-002-epistemic-untouched-baseline-plan.md
    blockers: [epistemic-001]
    verification: functional
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: Baseline validity requires matched runtime provenance, oracle isolation, bounded sampling, and an honest inconclusive state.
    model_requirements: [live adapter operation, provenance capture, bounded experiment design, cost telemetry, statistical restraint]
    escalation_triggers: [treatment changes before baseline, corpus/scorer hashes drift, runtime identity is missing, live preview lacks explicit approval]
    workspace_mode: shared-serial
    conflict_domains: [namespace:skill-eval, path:.kb/eval-runs/epistemic-baseline, path:evals/skill-eval/baselines, path:docs/results]
    shared_resources: [git:integration-owner, eval:live-runtime, budget:explicit-live-model]
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck skill-eval-regression --run-root .kb/eval-runs/epistemic-baseline --output evals/skill-eval/baselines/epistemic-investigation.json"
      expect: 0
    hitl: true
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: "epistemic-001 passes and the user approves the exact no-run baseline preview"
    next_agent_action: "Freeze hashes and control instruction identities, generate a no-run preview, request exact live-run approval, capture the baseline, and verify it with existing skill-eval commands."
    human_action: "Approve or reject the exact runtime/model, repetition count, corpus hash, and bounded-cost preview."
    can_continue_other_slices: false
  - id: epistemic-003
    title: Add the narrow investigation behavior after baseline freeze
    path: docs/plans/2026-08-23-003-epistemic-investigation-treatment-plan.md
    blockers: [epistemic-002]
    verification: integration
    test_level: integration
    functional_risk: broad
    model_tier: large
    model_tier_reason: The treatment changes planning-specific instruction surfaces and must preserve authority, agent-owned research, progress-based termination, and the invisible proceed path.
    model_requirements: [instruction design, planning behavior, compatibility testing, cognitive-load preservation]
    escalation_triggers: [rule becomes an always-on checklist, user questions increase for agent-owned research, unchanged reassurance can advance, treatment requires editing kb-cognitive, kb-work, manifest-contract, or frozen eval owners]
    workspace_mode: shared-serial
    conflict_domains: [skill:kb-plan, skill:kb-gate, file:config/skill-guidance-audit.json]
    shared_resources: [git:integration-owner, instruction:planning-trust-brake]
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck skill-lint"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: "epistemic-002 baseline artifact and hashes are verified"
    next_agent_action: "Protect baseline/oracle/scorer/test hashes, add the smallest bounded instruction treatment, and run structural compatibility checks without editing frozen eval surfaces."
    human_action: ""
    can_continue_other_slices: false
  - id: epistemic-004
    title: Run interleaved control/treatment replay and enforce promotion boundaries
    path: docs/plans/2026-08-23-004-epistemic-matched-replay-plan.md
    blockers: [epistemic-003]
    verification: functional
    test_level: full
    functional_risk: broad
    model_tier: large
    model_tier_reason: Final comparison must preserve separate outcomes, reject ease regressions, and avoid forcing a conclusion from weak evidence.
    model_requirements: [live eval operation, regression analysis, release verification, bounded claim synthesis]
    escalation_triggers: [baseline/scorer/oracle hashes differ, unnecessary investigation or user interruption worsens, evidence is underpowered, local-release fails]
    workspace_mode: shared-serial
    conflict_domains: [namespace:skill-eval, path:.kb/eval-runs/epistemic-treatment, path:docs/results, file:docs/context/eval-map.md, file:docs/context/operations/testing.md]
    shared_resources: [git:integration-owner, eval:live-runtime, budget:explicit-live-model]
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck skill-eval-regression --run-root .kb/eval-runs/epistemic-treatment --baseline evals/skill-eval/baselines/epistemic-investigation.json --output docs/results/epistemic-investigation-comparison.json"
      expect: 0
    hitl: true
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: "epistemic-003 passes and the user approves the exact interleaved control/treatment replay preview"
    next_agent_action: "Preview, request bounded approval, interleave hash-bound control and treatment arms, compare, and return promote, reject, or inconclusive."
    human_action: "Approve or reject the exact matched live replay; separately authorize any later global sync or delivery."
    can_continue_other_slices: false
  - id: epistemic-005
    title: Enforce planning assurance after behavioral promotion
    path: docs/plans/2026-08-23-005-planning-assurance-enforcement-plan.md
    blockers: [epistemic-004]
    verification: tdd
    test_level: integration
    functional_risk: broad
    model_tier: large
    model_tier_reason: This slice changes the manifest phase boundary and must preserve legacy resumability while rejecting stale or inconclusive new-plan assurance.
    model_requirements: [Go contract design, backward compatibility, skill workflow design, hash-bound receipt validation]
    escalation_triggers: [matched replay is not promote, legacy manifests fail, self-referential hashing appears, ordinary supported plans require user interaction]
    workspace_mode: shared-serial
    conflict_domains: [skill:kb-plan, skill:kb-work, namespace:manifest-contract, path:.github/skills/kb-plan/references]
    shared_resources: [git:integration-owner, gate:plan-to-work]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'ManifestContract.*PlanningAssurance' -count=1"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: "epistemic-004 returns promote with protected ease and epistemic metrics"
    next_agent_action: "Protect manifest-contract tests, prove RED, add the smallest new-schema sidecar validator, then update kb-plan and kb-work contracts."
    human_action: ""
    can_continue_other_slices: false
---

# Epistemic Investigation Gate

## Decision

Extend the existing `skill-eval` pipeline. Do not create another evaluator,
adapter, baseline engine, or top-level command family.

## Flow

`existing scorer/adapter extension -> untouched planning baseline -> planning-instruction treatment -> interleaved control/treatment replay -> deterministic enforcement after promote`

## Slices

| # | Outcome | Depends on | HITL |
|---|---|---|---|
| 1 | One compatible sealed epistemic path through existing skill-eval | — | no |
| 2 | Frozen corpus plus untouched live baseline | 1 | exact live-run approval |
| 3 | Narrow kb-plan and kb-gate treatment | 2 | no |
| 4 | Interleaved control/treatment replay with promote/reject/inconclusive | 3 | exact live-run approval |
| 5 | New-schema assurance receipt plus kb-work enforcement | 4 promote | no |

## Promotion Boundary

Promotion requires all of the following on matched evidence:

- missed investigations decrease or remain zero;
- unnecessary investigations do not increase;
- correct resolution and revision do not decrease;
- agent-resolvable cases add no user questions;
- supported `proceed` cases add no visible ceremony; and
- existing skill-eval fixtures and release checks remain green.

Conflicting or underpowered evidence returns `inconclusive`.

`reject` and `inconclusive` still produce valid reports, but they do not satisfy
the manifest `done_check`; they route back to the smallest justified treatment
or sampling change without altering the frozen oracle.

## Claim Boundary

Success supports only the evaluated runtime/model versions, instruction-loading
contexts, and fixture distribution. It does not establish a training cause,
general model reliability, or correct behavior by every model everywhere.

## Implementation Boundary

- Recheck current worktree/lease collision before the first edit; this checkout
  is currently on the cognitive-rename branch and `cmd/kbcheck` has active plan
  history.
- Use one manifest-owned worktree and shared-serial slices.
- Do not edit `kb-cognitive`; consume its existing response contract.
- Local plan-run commits are authorized. Do not sync global skills, push, open
  a PR, or deliver without the later owning authorization.

`plan-to-work` passed after the user approved the planning-specific enforcement
design and asked to fix the skills. Live baseline and replay calls retain their
separate exact-preview approval gates.
