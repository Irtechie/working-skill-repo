---
type: kb-manifest
kb_id: kb-2026-05-31-routing-trim
brainstorm_path: docs/context/epics/skill-minimalism-and-proof-harness.md
created: 2026-05-31
status: reviewed
slices:
  - id: slice-001
    title: "Add workflow-shape route fixtures"
    path: docs/plans/2026-05-31-021-eval-workflow-shape-route-fixtures-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add route fixtures proving small edits stay small and pipeline work escalates."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=5 discovered=0 unexplained=0; proof: powershell -ExecutionPolicy Bypass -File scripts/route-complexity-eval.ps1; proof: git diff --check; memory-impact: durable; areas=route workflow-shape fixture coverage; docs=docs/context/eval-map.md"
  - id: slice-002
    title: "Add compact shape classifier to kb-start and manifests"
    path: docs/plans/2026-05-31-022-skill-kb-start-shape-classifier-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add tiny workflow-shape classifier without building a giant pipeline each time."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=3 discovered=0 unexplained=0; proof: powershell -ExecutionPolicy Bypass -File scripts/route-complexity-eval.ps1; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof: rg -n Workflow Shape Check/workflow_shape .github/skills/kb-start/SKILL.md .github/skills/kb-plan/SKILL.md .github/skills/kb-epic/SKILL.md; memory-impact: durable; areas=kb-start shape routing and manifest workflow_shape"
  - id: slice-003
    title: "Implement loaded-surface reporting"
    path: docs/plans/2026-05-31-023-tool-loaded-surface-report-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Measure route-level lines/tokens before trimming."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=4 changed=3 discovered=0 unexplained=0; scope-forecast-unused: config/skill-quality.json - route map hardcoded in MVP script until real use proves config needed; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-surface-report.ps1; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-surface-report.ps1 -Route base -Json; proof: git diff --check; memory-impact: operational; docs=docs/context/operations/testing.md"
  - id: slice-004
    title: "Trim base and core workflow skills behind measurements"
    path: docs/plans/2026-05-31-024-skill-base-core-trim-plan.md
    blockers: [slice-001, slice-003]
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Trim only measured duplicate/generic text after route fixtures protect behavior."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=7 changed=1 discovered=0 unexplained=0; scope-forecast-unused: kb-start,kb-map,kb-brainstorm,kb-plan,kb-work,kb-complete - left untrimmed after measurement to avoid removing gates in same pass; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-surface-report.ps1 -Route base showed base 657 lines / token_estimate 4449 after earlier baseline 680 lines / token_estimate 6126; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof: git diff --check; memory-impact: durable; areas=kb-first-principles trim"
  - id: slice-005
    title: "Clean reference graph and deletion blockers"
    path: docs/plans/2026-05-31-025-tool-reference-graph-cleanup-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Resolve known unknown refs and make scanner useful for deletion safety."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=4 changed=2 discovered=0 unexplained=0; scope-forecast-unused: scripts/skill-lint.ps1 - scanner already caught these refs; scope-forecast-unused: config/skill-quality.json - no allowlist needed after cleanup; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof: rg -n ce-ideate/kb-route .github/skills README.md evals config returned no matches; proof: git diff --check; memory-impact: operational"
---

# KB: Routing And Trim

## Origin

Brainstorms:

- `docs/brainstorms/2026-05-31-base-layer-contract-requirements.md`
- `docs/brainstorms/2026-05-31-base-skill-trim-requirements.md`
- `docs/brainstorms/2026-05-31-core-workflow-trim-requirements.md`
- `docs/brainstorms/2026-05-31-workflow-shape-routing-requirements.md`

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Add workflow-shape route fixtures | - | integration | no | done |
| 2 | Add compact shape classifier to kb-start and manifests | slice-001 | integration | no | done |
| 3 | Implement loaded-surface reporting | - | integration | no | done |
| 4 | Trim base and core workflow skills behind measurements | slice-001, slice-003 | integration | no | done |
| 5 | Clean reference graph and deletion blockers | - | integration | no | done |

## Workflow Shape

`skill-bundle-change`: changes routing, lint/eval proof, and loaded skill
surface. It is not a coded pipeline framework slice.

## Done Criteria

- Route fixtures distinguish direct edit, normal plan, pipeline, and epic work.
- `kb-start` classifies shape cheaply without escalating every request.
- Loaded-surface reductions are measured before and after trims.
- Known stale skill references are fixed or explicitly parked.
