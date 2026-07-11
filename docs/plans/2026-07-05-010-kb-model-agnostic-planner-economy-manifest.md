---
type: kb-manifest
kb_id: kb-2026-07-05-model-agnostic-planner-economy
brainstorm_path: docs/context/research/2026-07-05-humanlayer-pinned-repos-planner-economy.md
created: 2026-07-05
status: reviewed
workflow_shape: "pipeline-change"
context_packet_contract: true
scope-verified-files:
  - .github/copilot-instructions.md
  - .github/skills/kb-plan/SKILL.md
  - .github/skills/kb-work/SKILL.md
  - .github/skills/kb-work/references/execution-prompt.md
  - .gitignore
  - AGENTS.md
  - README.md
  - cmd/kbcheck/checks.go
  - cmd/kbcheck/checks_test.go
  - cmd/kbcheck/context_packet.go
  - cmd/kbcheck/context_packet_test.go
  - cmd/kbcheck/eval_adapters.go
  - cmd/kbcheck/execution_telemetry.go
  - cmd/kbcheck/execution_telemetry_test.go
  - cmd/kbcheck/main.go
  - cmd/kbcheck/main_test.go
  - cmd/kbcheck/manifest_contract.go
  - cmd/kbcheck/manifest_contract_test.go
  - cmd/kbcheck/provider_hygiene.go
  - cmd/kbcheck/provider_hygiene_test.go
  - cmd/kbcheck/report_validators.go
  - cmd/kbcheck/skill_eval.go
  - cmd/kbcheck/skill_repo_contract_test.go
  - cmd/kbcheck/surface_report_test.go
  - cmd/kbcheck/swarm.go
  - cmd/kbcheck/testdata/context-packet-invalid.json
  - cmd/kbcheck/testdata/context-packet-valid.json
  - cmd/kbcheck/testdata/execution-telemetry-valid.json
  - docs/context/PROJECT.md
  - docs/context/decisions/2026-07-05-kb-control-plane-blueprint.md
  - docs/context/decisions/2026-07-05-model-agnostic-core-vs-payload.md
  - docs/context/eval-map.md
  - docs/context/goals/finish-skill-repo-hardening.md
  - docs/context/goals/kb-native-scoped-learning.md
  - docs/context/kb/instincts/project.yaml
  - docs/context/kb/instincts/scoped/skill-bundle/provider-hygiene.yaml
  - docs/context/kb/kb-completions.txt
  - docs/context/kb/context-packet-schema.md
  - docs/context/memory-maintenance.md
  - docs/context/operations/testing.md
  - docs/context/research/2026-07-09-cross-runtime-token-efficiency.md
  - docs/context/research/README.md
  - docs/handoffs/done/2026-07-09-phoenix-routing-go-gate-blocker.md
  - docs/plans/2026-07-01-010-kb-native-scoped-learning-manifest.md
  - docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md
  - docs/plans/2026-07-05-012-task-state-context-packet-spike-plan.md
  - docs/plans/2026-07-05-013-kb-plan-work-packet-integration-plan.md
  - docs/plans/2026-07-05-014-observable-metrics-kill-resume-plan.md
  - docs/plans/2026-07-05-015-segmentation-adapter-boundary-plan.md
  - docs/plans/2026-07-05-016-spike-decision-report-plan.md
  - docs/plans/2026-07-05-017-docs-sync-release-plan.md
  - docs/solutions/workflow-issues/contributor-core-vs-release-sync-gates-2026-06-10.md
  - docs/solutions/workflow-issues/optional-provider-hygiene-2026-07-09.md
  - evals/skill-eval/result.schema.json
  - todo-done.md
  - todo.md
last_refreshed: 2026-07-09
stale_refresh:
  decision: "Keep KB as lightweight core; do not build a second orchestration runtime."
  evidence: "Run-state, proof traces, goal ledgers, and manifest gates landed after this plan. Remaining value is the packet boundary, telemetry, provider hygiene, and honest loaded-surface measurement."
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
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "vendor-neutral context packet schema and fixtures exist"
      - "separate usage telemetry fields validate without becoming proof"
    proof:
      - cmd/kbcheck/context_packet.go
      - cmd/kbcheck/context_packet_test.go
      - cmd/kbcheck/testdata/context-packet-valid.json
      - docs/context/kb/context-packet-schema.md
    blockers: []
    passed_at: "2026-07-09T19:00:00-04:00"
    allowed_next_action: "kb-work continue"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "kb-plan produces bounded packet guidance"
      - "kb-work validates and consumes packets before broad search"
    proof:
      - .github/skills/kb-plan/SKILL.md
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-work/references/execution-prompt.md
    blockers: []
    passed_at: "2026-07-09T19:00:00-04:00"
    allowed_next_action: "kb-work continue"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "provider hygiene rejects Phoenix activation"
      - "CCE remains allowed as an optional adapter"
      - "usage telemetry preserves raw fields"
    proof:
      - cmd/kbcheck/provider_hygiene.go
      - cmd/kbcheck/provider_hygiene_test.go
      - cmd/kbcheck/execution_telemetry.go
    blockers: []
    passed_at: "2026-07-09T19:00:00-04:00"
    allowed_next_action: "kb-work continue"
  - gate_id: slice-slice-005-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "ambient instructions are deduplicated"
      - "base and conditional loaded surfaces are separate"
      - "provider/runtime boundary is documented"
    proof:
      - .github/copilot-instructions.md
      - AGENTS.md
      - cmd/kbcheck/report_validators.go
      - docs/context/research/2026-07-09-cross-runtime-token-efficiency.md
    blockers: []
    passed_at: "2026-07-09T19:00:00-04:00"
    allowed_next_action: "kb-work continue"
  - gate_id: slice-slice-006-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "core-vs-payload decision cites executable evidence"
      - "second runtime is rejected unless real adapter/recovery evidence changes"
    proof:
      - docs/context/decisions/2026-07-05-model-agnostic-core-vs-payload.md
      - docs/context/decisions/2026-07-05-kb-control-plane-blueprint.md
    blockers: []
    passed_at: "2026-07-09T19:00:00-04:00"
    allowed_next_action: "kb-work continue"
  - gate_id: slice-slice-007-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "docs and project memory reflect the decision"
      - "required skill roots hash-match"
      - "release gates pass"
    proof:
      - README.md
      - docs/context/PROJECT.md
      - docs/context/eval-map.md
      - docs/context/operations/testing.md
    blockers: []
    passed_at: "2026-07-09T19:00:00-04:00"
    allowed_next_action: "kb-complete docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md"
  - gate_id: work-to-complete
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "all slices are done"
      - "manifest contract passes"
      - "core and local-release pass"
      - "required skill hashes match"
      - "working and ATV diff checks pass"
      - "no unresolved P0/P1 findings are known"
    proof:
      - docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md
      - docs/context/decisions/2026-07-05-model-agnostic-core-vs-payload.md
      - docs/context/eval-map.md
      - cmd/kbcheck/context_packet_test.go
      - cmd/kbcheck/provider_hygiene_test.go
      - README.md
    blockers: []
    passed_at: "2026-07-09T19:12:00-04:00"
    allowed_next_action: "kb-complete docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md"
  - gate_id: complete-to-ship
    owner_skill: kb-complete
    status: passed
    required_evidence:
      - "core and local-release pass after review fixes"
      - "goal done check passes"
      - "functional CLI packet, telemetry, provider, and manifest probes pass"
      - "multi-agent review ran"
      - "all P0/P1 review findings are resolved"
      - "follow-up resolution is complete"
      - "proof/demo evidence is recorded"
      - "compound and targeted refresh completed"
      - "learn completed and cadence was updated"
      - "project memory was refreshed"
      - "memory maintenance was updated"
      - "cleanup completed"
    proof:
      - cmd/kbcheck/context_packet_test.go
      - cmd/kbcheck/execution_telemetry_test.go
      - cmd/kbcheck/provider_hygiene_test.go
      - cmd/kbcheck/manifest_contract_test.go
      - docs/context/goals/finish-skill-repo-hardening.md
      - docs/context/decisions/2026-07-05-model-agnostic-core-vs-payload.md
      - docs/solutions/workflow-issues/optional-provider-hygiene-2026-07-09.md
      - docs/solutions/workflow-issues/contributor-core-vs-release-sync-gates-2026-06-10.md
      - docs/context/kb/instincts/scoped/skill-bundle/provider-hygiene.yaml
      - docs/context/PROJECT.md
      - docs/context/memory-maintenance.md
      - todo-done.md
    blockers: []
    passed_at: "2026-07-09T19:48:00-04:00"
    allowed_next_action: "kb-ship docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md"
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
    title: "Implement context-packet and execution-telemetry contracts"
    path: docs/plans/2026-07-05-012-task-state-context-packet-spike-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: large
    context_packet_path: cmd/kbcheck/testdata/context-packet-valid.json
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: "slice-001 done"
    next_agent_action: "Implement the smallest vendor-neutral packet and telemetry schema without a second task-state runtime."
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
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: "slice-002 done"
    next_agent_action: "Make planning produce packet data and work consume/update it."
    human_action: ""
    can_continue_other_slices: true
    protected_oracles: []
  - id: slice-004
    title: "Prove telemetry normalization and provider hygiene"
    path: docs/plans/2026-07-05-014-observable-metrics-kill-resume-plan.md
    blockers: [slice-002, slice-003]
    verification: integration
    test_level: integration
    functional_risk: narrow
    model_tier: large
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: "slice-003 done"
    next_agent_action: "Add deterministic fixtures for usage normalization and opt-in provider hygiene."
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
    status: done
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
    verification: verification-only
    test_level: none
    functional_risk: none
    model_tier: large
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: "slices 003-005 done"
    next_agent_action: "Summarize spike evidence and recommend keep-KB-core, KB-as-payload, or replacement."
    human_action: ""
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
    status: done
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
| 2 | Implement context-packet and execution-telemetry contracts | slice-001 | integration | no | done |
| 3 | Wire kb-plan and kb-work through context packets | slice-002 | integration | no | done |
| 4 | Prove telemetry normalization and provider hygiene | slice-002, slice-003 | integration | no | done |
| 5 | Tighten custom-instruction segmentation and first adapter boundary | slice-002, slice-003 | integration | no | done |
| 6 | Write KB-core vs KB-payload decision report | slices 003-005 | verification-only | no | done |
| 7 | Update docs, sync surfaces, and release gate | slice-006 | verification-only | no | done |

## Work Gate

`work-to-complete` passed on 2026-07-09. All slices are done. Run:

```shell
kb-complete docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md
```

## Completion Notes

- review-mode: multi-agent
- review: P0=0 P1=8(resolved) P2=0 P3=0
- follow-up-resolution: resolved 8, logged 0, blocked 0
- proof: `go test ./cmd/kbcheck`; context-packet, execution-telemetry,
  provider-hygiene `--include-user`, manifest-contract, `core`,
  `local-release`, and both repo diff checks passed
- functional-test: functional-cli probes passed; no UI/API/browser surface
- demo: skipped - maintainer CLI/tooling change with executable output
- compound: wrote optional-provider hygiene solution and refreshed contributor
  core-vs-release guidance
- learn: 1 new scoped instinct, 1 project instinct updated; cadence 13
- evolve: skipped - cadence not divisible by 5
- steering-feedback: current=1 memory=0 observations=0 landmine-candidates=0 instinct-evidence=1
- kb-map-refresh: done - PROJECT, decisions, eval map, testing, research,
  solutions, goal, todo, and handoff state updated
- memory-maintenance: 4 signals recorded; deep review recommended because
  completed cycle count is 13
- compact: skipped - startup memory remains route-oriented; no new bloat signal
- cleanup: resolved handoff archived; `.kb/runs` ignored; no QA screenshots or
  observations required pruning
- alerts: run `kb-memory-review docs/context/memory-maintenance.md` before the
  next large feature
