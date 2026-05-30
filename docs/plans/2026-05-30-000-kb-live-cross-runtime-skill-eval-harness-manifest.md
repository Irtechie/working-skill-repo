---
type: kb-manifest
kb_id: kb-2026-05-30-live-cross-runtime-skill-eval-harness
brainstorm_path: docs/brainstorms/2026-05-30-live-cross-runtime-skill-eval-harness-requirements.md
created: 2026-05-30
status: active
scope-verified-files:
  - docs/brainstorms/2026-05-30-live-cross-runtime-skill-eval-harness-requirements.md
  - docs/context/eval-map.md
  - docs/context/operations/testing.md
  - docs/context/architecture/README.md
  - config/skill-quality.json
  - scripts/skill-eval.ps1
  - scripts/skill-eval-run-codex.ps1
  - evals/skill-eval/result.schema.json
  - evals/route-complexity/README.md
  - .github/skills/kb-check/scripts/kb-check.ps1
slices:
  - id: slice-001
    title: "Add GHCP live skill eval adapter"
    path: docs/plans/2026-05-30-001-tool-ghcp-live-skill-eval-adapter-plan.md
    blockers: []
    verification: functional
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "If Copilot auth is unavailable, adapter must emit a clear unavailable/skip state rather than pass."
  - id: slice-002
    title: "Wire adapter dry-runs into canonical checks"
    path: docs/plans/2026-05-30-002-tool-adapter-dry-run-kb-check-plan.md
    blockers: [slice-001]
    verification: functional
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: ""
  - id: slice-003
    title: "Add live cross-runtime corpus runner"
    path: docs/plans/2026-05-30-003-tool-live-cross-runtime-corpus-runner-plan.md
    blockers: [slice-001]
    verification: functional
    test_level: functional-cli
    functional_risk: broad
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "Live model calls remain explicit and outside kb-check -All."
  - id: slice-004
    title: "Expand deterministic trace rule scoring"
    path: docs/plans/2026-05-30-004-tool-trace-rule-scoring-plan.md
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: ""
  - id: slice-005
    title: "Add transcript-derived claim verifier"
    path: docs/plans/2026-05-30-005-tool-transcript-claim-verifier-plan.md
    blockers: [slice-004]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "Extraction can be LLM-assisted later, but pass/fail checks must remain deterministic."
  - id: slice-006
    title: "Add output quality rubric scorer"
    path: docs/plans/2026-05-30-006-tool-output-quality-rubric-plan.md
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "Rubric output is separate from deterministic route/proof/claim pass/fail."
  - id: slice-007
    title: "Add cost and regression reporting"
    path: docs/plans/2026-05-30-007-tool-cost-regression-report-plan.md
    blockers: [slice-003]
    verification: functional
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: ""
  - id: slice-008
    title: "Add eval-map scaffold negative validation"
    path: docs/plans/2026-05-30-008-skill-eval-map-negative-validation-plan.md
    blockers: []
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: ""
---

# KB: Live Cross-Runtime Skill Eval Harness

## Origin

Brainstorm: `docs/brainstorms/2026-05-30-live-cross-runtime-skill-eval-harness-requirements.md`

## Slice Overview

| # | Slice | Blocked By | Verification | Test Level | HITL | Status |
|---|---|---|---|---|---|---|
| 1 | Add GHCP live skill eval adapter | - | functional | functional-cli | no | pending |
| 2 | Wire adapter dry-runs into canonical checks | slice-001 | functional | functional-cli | no | pending |
| 3 | Add live cross-runtime corpus runner | slice-001 | functional | functional-cli | no | pending |
| 4 | Expand deterministic trace rule scoring | - | tdd | functional-cli | no | pending |
| 5 | Add transcript-derived claim verifier | slice-004 | tdd | functional-cli | no | pending |
| 6 | Add output quality rubric scorer | - | tdd | functional-cli | no | pending |
| 7 | Add cost and regression reporting | slice-003 | functional | functional-cli | no | pending |
| 8 | Add eval-map scaffold negative validation | - | verification-only | none | no | pending |

## Assumptions

- Local Copilot CLI remains the first GHCP runtime because it is installed and
  documented for programmatic prompts.
- GHCP output schema enforcement is unavailable unless future `copilot help` or
  official docs expose a schema flag.
- Live model calls stay explicit and are not part of `kb-check -All`.
- Local JSON and Markdown artifacts remain the source of truth; dashboard
  integrations are exporter-only.

## Completion Criteria

- GHCP and Codex adapters both have dry-run coverage under `kb-check -All`.
- A live corpus command can run all current route fixtures for both runtimes.
- Deterministic scoring can fail route, trace, proof, structured claim, and
  transcript-derived false-claim cases.
- Output quality and cost/regression reports exist without replacing
  deterministic pass/fail gates.
- `docs/context/eval-map.md`, `docs/context/operations/testing.md`, and
  `todo.md` reflect the new truth.
