---
type: kb-manifest
kb_id: kb-2026-05-31-proof-pipeline-spike
brainstorm_path: docs/context/epics/skill-minimalism-and-proof-harness.md
created: 2026-05-31
status: reviewed
slices:
  - id: slice-001
    title: "Persist skill-eval regression baselines"
    path: docs/plans/2026-05-31-001-tool-skill-eval-regression-baseline-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Design and implement baseline persistence for skill-eval."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=5 changed=4 discovered=1 unexplained=0; scope-discovery: scripts/skill-eval-baseline-selftest.ps1 - required deterministic baseline regression selftest; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-eval-baseline-selftest.ps1; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-eval.ps1 -Json; proof: git diff --check; review-fix: baseline comparison now fails negative fixtures that incorrectly start passing; memory-impact: operational; docs=docs/context/operations/testing.md"
  - id: slice-002
    title: "Build tiny coded pipeline spike"
    path: docs/plans/2026-05-31-002-tool-kb-pipeline-spike-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Create the smallest coded pipeline runner and one pipeline definition."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=4 changed=6 discovered=2 unexplained=0; scope-discovery: .gitignore - required to keep generated pipeline runs out of git; scope-discovery: scripts/kb-pipeline-selftest.ps1 - required deterministic pipeline spike selftest; proof: powershell -ExecutionPolicy Bypass -File scripts/kb-pipeline-selftest.ps1; proof: powershell -ExecutionPolicy Bypass -File scripts/kb-pipeline.ps1 -Status; proof: git diff --check; memory-impact: operational; docs=docs/context/operations/testing.md"
  - id: slice-003
    title: "Add protected oracle pipeline contract"
    path: docs/plans/2026-05-31-003-workflow-protected-oracle-contract-plan.md
    blockers: [slice-002]
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Teach plan/work/check artifacts how to express locked test-first oracles."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=5 changed=3 discovered=0 unexplained=0; scope-forecast-unused: .github/skills/kb-check/scripts/kb-check.ps1 - already runs the manifest selftest from prior slice; scope-forecast-unused: scripts/skill-eval-manifest-selftest.ps1 - reused existing tamper selftest without changes; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-eval-manifest-selftest.ps1; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof: git diff --check; memory-impact: durable; areas=KB plan/work/check protected oracle contract; docs=docs/context/operations/testing.md"
---

# KB: Proof Pipeline Spike

## Origin

Epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Persist skill-eval regression baselines | - | integration | no | done |
| 2 | Build tiny coded pipeline spike | - | integration | no | done |
| 3 | Add protected oracle pipeline contract | slice-002 | integration | no | done |

## Workflow Shape

`pipeline-change`: touches scripts, config, eval proof behavior, docs, and KB
workflow contracts.

## Done Criteria

- Skill eval baselines persist and can fail on regression.
- Pipeline spike creates a run directory, selected pipeline record, phase
  context packet, protected-file manifest, and proof summary.
- Protected oracle behavior is represented in plan/work/check artifacts without
  keeping standalone TDD globally loaded.
