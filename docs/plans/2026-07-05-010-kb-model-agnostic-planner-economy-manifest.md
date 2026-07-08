---
type: kb-manifest
kb_id: kb-2026-07-05-model-agnostic-planner-economy
brainstorm_path: docs/context/research/2026-07-05-humanlayer-pinned-repos-planner-economy.md
created: 2026-07-05
status: active
workflow_shape: "pipeline-change"
decision_position:
  default: "Keep KB as the planning, proof, skill, and learning payload while an absorption spike proves whether KB should also own durable runtime state."
  absorption_threshold: "If a small repo-local task-state store and context-packet object compose cleanly with kb-plan, kb-work, and kbcheck in one bounded spike, keep KB as core. If markdown-as-state fights recovery, adapters, or proof, design a small runtime and let KB ride on it as payload."
  review_owner: "User approved the absorption spike scope on 2026-07-05T13:59:39-04:00; Fable critique incorporated on 2026-07-05."
safe_assumptions:
  - "HumanLayer/CodeLayer already has durable sessions, approvals, HITL-as-state, and state transitions; that is not the gap to claim."
  - "HumanLayer's public humanlayer repo is design archaeology and issue evidence; its README says the public code is mostly deprecated."
  - "The real gaps to test for KB are model-agnostic adapter boundaries, tier-calibration telemetry, and clean segmentation of custom instructions, commands, skills, agents, subagents, and tools."
  - "A state machine is only safer when recovery invariants are deterministic and tested; stuck states must have repair paths."
  - "Build one adapter boundary for the daily runtime first. Add a second adapter only when the first boundary is proven."
model_tier_contract:
  tiny: "Deterministic inventories, schema/frontmatter fill, status summaries, docs table updates."
  small: "Narrow mechanical edits, simple tests, fixture updates, command output summarization."
  medium: "Ordinary vertical slices with complete context packets and clear proof."
  large: "Decomposition, design, architecture/security, adapter boundary decisions, failed-loop diagnosis, final synthesis."
  proof_rule: "No model tier is proof. Proof is executable evidence from tests, commands, traces, snapshots, or kbcheck gates."
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "HumanLayer pinned-repo planner-economy research exists"
      - "Dex/HumanLayer harness research exists"
      - "project memory identifies this repo as the portable skill bundle"
      - "Fable critique has been incorporated into the decision posture"
      - "active todo records planner-economy hardening as approved active work"
    proof:
      - docs/context/research/2026-07-05-humanlayer-pinned-repos-planner-economy.md
      - docs/context/research/2026-07-05-dexhorthy-humanlayer-agent-harness-research.md
      - docs/context/PROJECT.md
      - docs/context/decisions/2026-07-05-model-agnostic-core-vs-payload.md
      - todo.md
    blockers: []
    passed_at: "2026-07-05T12:49:37-04:00"
    allowed_next_action: "Review the absorption-spike scope with the user before implementation."
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest path exists"
      - "all 7 slice plan paths exist"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk, model_tier"
      - "user authorizes the bounded absorption spike"
    proof:
      - docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md
      - docs/plans/2026-07-05-011-absorption-spike-scope-plan.md
      - docs/plans/2026-07-05-012-task-state-context-packet-spike-plan.md
      - docs/plans/2026-07-05-013-kb-plan-work-packet-integration-plan.md
      - docs/plans/2026-07-05-014-observable-metrics-kill-resume-plan.md
      - docs/plans/2026-07-05-015-segmentation-adapter-boundary-plan.md
      - docs/plans/2026-07-05-016-spike-decision-report-plan.md
      - docs/plans/2026-07-05-017-docs-sync-release-plan.md
      - docs/context/decisions/2026-07-05-kb-control-plane-blueprint.md
    dag_validation: "slice-001 gates implementation; slice-002 provides state schema; slice-003 integrates planning/workflow; slice-004 and slice-005 prove recovery, telemetry, and boundaries; slice-006 decides core vs payload; slice-007 releases only after decision."
    blockers: []
    passed_at: "2026-07-05T13:59:39-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "user approved the absorption spike scope"
      - "control-plane blueprint exists"
      - "manifest plan-to-work gate is passed"
      - "todo records slice-002 as next runnable work"
    proof:
      - "user approval: ok put it together"
      - docs/context/decisions/2026-07-05-kb-control-plane-blueprint.md
      - docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md
      - todo.md
    blockers: []
    passed_at: "2026-07-05T13:59:39-04:00"
    allowed_next_action: "kb-work continue"
slices:
  - id: slice-001
    title: "Approve absorption spike scope and decision criteria"
    path: docs/plans/2026-07-05-011-absorption-spike-scope-plan.md
    blockers: []
    verification: hitl
    test_level: none
    functional_risk: none
    model_tier: large
    hitl: true
    status: done
    owner: human
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Approved; slice-002 is the next runnable implementation slice."
    human_action: ""
    can_continue_other_slices: false
    notes: "User approved the integrated blueprint on 2026-07-05T13:59:39-04:00."
    protected_oracles: []
  - id: slice-002
    title: "Spike repo-local task state and context-packet object"
    path: docs/plans/2026-07-05-012-task-state-context-packet-spike-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: large
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: "slice-001 done"
    next_agent_action: "Implement the smallest structured task-state/context-packet schema and validator."
    human_action: ""
    can_continue_other_slices: true
    protected_oracles: []
  - id: slice-003
    title: "Wire kb-plan and kb-work through context packets"
    path: docs/plans/2026-07-05-013-kb-plan-work-packet-integration-plan.md
    blockers: [slice-002]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: large
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: "slice-002 done"
    next_agent_action: "Make planning produce packet data and work consume/update it."
    human_action: ""
    can_continue_other_slices: true
    protected_oracles: []
  - id: slice-004
    title: "Prove recovery, stuck-state handling, and execution telemetry"
    path: docs/plans/2026-07-05-014-observable-metrics-kill-resume-plan.md
    blockers: [slice-002, slice-003]
    verification: integration
    test_level: integration
    functional_risk: narrow
    model_tier: large
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: "slice-003 done"
    next_agent_action: "Add deterministic fixtures for resume, blocked HITL, stale interrupting state, telemetry, and recovery."
    human_action: ""
    can_continue_other_slices: true
    protected_oracles: []
  - id: slice-005
    title: "Tighten custom-instruction segmentation and first adapter boundary"
    path: docs/plans/2026-07-05-015-segmentation-adapter-boundary-plan.md
    blockers: [slice-002, slice-003]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: large
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: "slice-003 done"
    next_agent_action: "Document and lint custom-instruction/command/skill/agent/subagent/tool ownership plus one daily-runtime adapter contract."
    human_action: ""
    can_continue_other_slices: true
    protected_oracles: []
  - id: slice-006
    title: "Write KB-core vs KB-payload decision report"
    path: docs/plans/2026-07-05-016-spike-decision-report-plan.md
    blockers: [slice-003, slice-004, slice-005]
    verification: hitl
    test_level: none
    functional_risk: none
    model_tier: large
    hitl: true
    status: pending
    owner: human
    blocked_reason: ""
    resume_when: "slices 003-005 done"
    next_agent_action: "Summarize spike evidence and recommend keep-KB-core, KB-as-payload, or replacement."
    human_action: "Accept or override the final architecture decision."
    can_continue_other_slices: false
    protected_oracles: []
  - id: slice-007
    title: "Update docs, sync surfaces, and release gate"
    path: docs/plans/2026-07-05-017-docs-sync-release-plan.md
    blockers: [slice-006]
    verification: verification-only
    test_level: functional-cli
    functional_risk: narrow
    model_tier: medium
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: "slice-006 done"
    next_agent_action: "Refresh docs, propagate approved skill changes, and run release gates."
    human_action: ""
    can_continue_other_slices: true
    protected_oracles: []
---

# KB: Model-Agnostic Planner Economy Absorption Spike

## Decision Summary

Fable's critique is accepted. The earlier "keep KB as core" recommendation was
directionally useful but too confident.

The current decision is:

```text
Use KB as the planning/proof/skill/learning payload now.
Run one bounded absorption spike to prove whether KB should also own durable runtime state.
```

HumanLayer/CodeLayer already proves useful runtime mechanics: durable sessions,
approvals as state, event history, session lineage, status transitions, and
telemetry. The question is not whether those ideas matter. The question is
whether KB can absorb the smallest useful version without turning markdown into
a brittle database.

The integrated blueprint is recorded in
`docs/context/decisions/2026-07-05-kb-control-plane-blueprint.md`.

## Spike Questions

1. Can a repo-local task-state store and context-packet object compose with
   `kb-plan`, `kb-work`, and `kbcheck` without fighting the skill bundle shape?
2. Can recovery from blocked, interrupted, stale, or half-written states be
   deterministic and tested?
3. Can the core stay model-agnostic while one daily-runtime adapter is wired
   first?
4. Can model-tier telemetry measure whether tiny/small/medium/large slices were
   sized correctly?
5. Can custom instruction, command, skill, agent, subagent, and tool ownership
   stay clear enough that cheaper workers receive packets instead of broad
   authority?

## Absorption Threshold

Keep KB as core if the spike proves:

- structured state can be created, validated, resumed, and repaired by `kbcheck`;
- context packets reduce broad rediscovery and are consumed by `kb-work`;
- stuck states have explicit recovery paths and tests;
- adapter details do not leak Claude/Codex/vendor assumptions into slice plans;
- telemetry captures predicted tier, actual tier/model, proof outcome, rework,
  escalation, and packet sufficiency.

Move toward a small runtime with KB as payload if the spike shows:

- state updates require brittle markdown surgery across multiple skills;
- recovery depends on model judgment instead of deterministic checks;
- adapter details leak into planning artifacts;
- packet execution cannot be measured externally;
- a second host/runtime adapter cannot be added without redesigning the core.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Approve absorption spike scope and decision criteria | - | hitl | yes | done |
| 2 | Spike repo-local task state and context-packet object | slice-001 | integration | no | pending |
| 3 | Wire kb-plan and kb-work through context packets | slice-002 | integration | no | pending |
| 4 | Prove recovery, stuck-state handling, and execution telemetry | slice-002, slice-003 | integration | no | pending |
| 5 | Tighten custom-instruction segmentation and first adapter boundary | slice-002, slice-003 | integration | no | pending |
| 6 | Write KB-core vs KB-payload decision report | slices 003-005 | hitl | yes | pending |
| 7 | Update docs, sync surfaces, and release gate | slice-006 | verification-only | no | pending |

## Work Gate

`plan-to-work` passed on 2026-07-05T13:59:39-04:00. The next runnable slice is
slice-002. Run:

```shell
kb-work docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md
```
