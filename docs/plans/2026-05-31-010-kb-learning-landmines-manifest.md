---
type: kb-manifest
kb_id: kb-2026-05-31-learning-landmines
brainstorm_path: docs/brainstorms/2026-05-31-repo-local-learning-landmines-requirements.md
created: 2026-05-31
status: reviewed
slices:
  - id: slice-001
    title: "Define local landmine schema and lifecycle"
    path: docs/plans/2026-05-31-011-doc-landmine-schema-lifecycle-plan.md
    blockers: []
    verification: docs-check
    test_level: static
    functional_risk: none
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add a local landmine schema with owner, fix condition, and verification fields."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=3 discovered=0 unexplained=0; proof: git diff --check; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof: rg -n Active Landmines/owner_surface docs/context/landmines.md docs/context/PROJECT.md docs/context/operations/testing.md; memory-impact: durable; areas=landmine local memory schema; docs=docs/context/landmines.md,docs/context/PROJECT.md"
  - id: slice-002
    title: "Add landmine evidence fields to learn flow"
    path: docs/plans/2026-05-31-012-skill-learn-landmine-fields-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Constrain learn to capture landmine candidates with evidence, severity, and owner."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=1 discovered=0 unexplained=0; scope-forecast-unused: config/skill-quality.json - no lint config change needed; scope-forecast-unused: scripts/skill-lint.ps1 - no deterministic generic-landmine scanner added yet; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof: rg -n Capture Landmine/owner_surface .github/skills/learn/SKILL.md; proof: git diff --check; memory-impact: durable; areas=learn landmine candidate contract"
  - id: slice-003
    title: "Gate evolve promotion and sync"
    path: docs/plans/2026-05-31-013-skill-evolve-approval-gate-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: true
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: "Human approves exact prompt wording or accepts default."
    next_agent_action: "Add explicit human approval before learned-* promotion/sync from this bundle."
    human_action: "Approve final wording for generated learned-* skill promotion."
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=2 discovered=0 unexplained=0; scope-forecast-unused: scripts/skill-sync-report.ps1 - existing required/optional sync drift check already blocks required target drift; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof: rg -n Human Approval Gate/Promote these generated .github/skills/evolve/SKILL.md docs/context/operations/testing.md; proof: git diff --check; memory-impact: durable; areas=evolve generated skill approval gate"
  - id: slice-004
    title: "Load and resolve active landmines through KB workflows"
    path: docs/plans/2026-05-31-014-skill-kb-map-landmine-loading-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Teach kb-map/kb-complete to load only active landmines and archive resolved ones."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=3 discovered=0 unexplained=0; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof: rg -n landmines/landmine-stale .github/skills/kb-map/SKILL.md .github/skills/kb-complete/SKILL.md .github/skills/kb-memory-review/SKILL.md; proof: git diff --check; memory-impact: durable; areas=KB memory landmine loading and resolution"
---

# KB: Learning Landmines

## Origin

Brainstorm: `docs/brainstorms/2026-05-31-repo-local-learning-landmines-requirements.md`

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Define local landmine schema and lifecycle | - | docs-check | no | done |
| 2 | Add landmine evidence fields to learn flow | slice-001 | integration | no | done |
| 3 | Gate evolve promotion and sync | slice-001 | integration | yes | done |
| 4 | Load and resolve active landmines through KB workflows | slice-001 | integration | no | done |

## Workflow Shape

`skill-bundle-change`: touches workflow skills and repo-local memory contracts.

## Done Criteria

- Landmines have owner, evidence, severity, fix condition, and verification.
- Fixed landmines archive immediately after proof.
- `evolve` keeps numeric maturity gates and adds explicit human approval before
  portable-bundle promotion or sync.
