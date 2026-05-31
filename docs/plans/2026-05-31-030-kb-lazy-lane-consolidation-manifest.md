---
type: kb-manifest
kb_id: kb-2026-05-31-lazy-lane-consolidation
brainstorm_path: docs/context/epics/skill-minimalism-and-proof-harness.md
created: 2026-05-31
status: reviewed
slices:
  - id: slice-001
    title: "Create architecture deepening lazy lane"
    path: docs/plans/2026-05-31-031-skill-architecture-deepening-lane-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add route proof and compact lazy architecture deepening lane."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=4 discovered=1 unexplained=0; scope-discovery: config/skill-quality.json - allowed route list required for new route fixture; proof: powershell -ExecutionPolicy Bypass -File scripts/route-complexity-eval.ps1; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; memory-impact: durable; areas=kb-architecture-deepening lazy lane and route fixtures"
  - id: slice-002
    title: "Absorb TDD anti-cheat behavior into pipeline"
    path: docs/plans/2026-05-31-032-skill-tdd-absorb-delete-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Move protected-oracle behavior into KB pipeline, then park/delete standalone TDD if safe."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=4 changed=3 discovered=0 unexplained=0; scope-forecast-unused: scripts/skill-lint.ps1 - stale tdd reference detection already covered by reference scan; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-eval-manifest-selftest.ps1; memory-impact: durable; areas=tdd compatibility lane and protected-oracle pipeline wording"
  - id: slice-003
    title: "Merge todo skills around root todo.md"
    path: docs/plans/2026-05-31-033-skill-todo-lane-merge-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Align todo-create/todo-triage with root todo.md and remove backlog.md pressure."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=4 changed=4 discovered=1 unexplained=0; scope-discovery: kb-review/SKILL.md - review todo output still referenced file-per-todo storage; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof: rg -n backlog .github/skills README.md found no KB backlog adoption; memory-impact: durable; areas=root todo.md canonical workflow"
  - id: slice-004
    title: "Audit narrow lanes for deletion safety"
    path: docs/plans/2026-05-31-034-skill-narrow-lane-deletion-safety-plan.md
    blockers: [slice-003]
    verification: integration
    test_level: cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add route/reference proof before merging or deleting narrow lanes."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=5 changed=3 discovered=0 unexplained=0; scope-forecast-unused: kb-fix,kb-troubleshoot - retained unchanged after audit because route fixtures already distinguish them; proof: powershell -ExecutionPolicy Bypass -File scripts/route-complexity-eval.ps1; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; memory-impact: durable; docs=docs/context/research/2026-05-31-narrow-lane-deletion-safety.md"
  - id: slice-005
    title: "Codify propagation policy"
    path: docs/plans/2026-05-31-035-doc-propagation-policy-plan.md
    blockers: []
    verification: docs-check
    test_level: static
    functional_risk: none
    hitl: true
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Document thin optional ATV scaffold/plugin policy after confirmation."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=2 discovered=0 unexplained=0; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-sync-report.ps1 returned 0 required issues with optional ATV scaffold/plugin warning-only; proof: git diff --check; memory-impact: durable; docs=README.md AGENTS.md; default from discussion: keep optional ATV scaffold/plugin thin"
---

# KB: Lazy Lane Consolidation

## Origin

Brainstorms:

- `docs/brainstorms/2026-05-31-architecture-deepening-lane-requirements.md`
- `docs/brainstorms/2026-05-31-narrow-lane-trim-requirements.md`
- `docs/brainstorms/2026-05-31-questionable-global-skill-trim-requirements.md`

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Create architecture deepening lazy lane | - | integration | no | done |
| 2 | Absorb TDD anti-cheat behavior into pipeline | - | integration | no | done |
| 3 | Merge todo skills around root todo.md | - | integration | no | done |
| 4 | Audit narrow lanes for deletion safety | slice-003 | integration | no | done |
| 5 | Codify propagation policy | - | docs-check | yes | done |

## Workflow Shape

`skill-bundle-change`: consolidates lazy lanes and deletion candidates behind
route/reference proof.

## Done Criteria

- Architecture deepening is lazy and route-distinct.
- TDD anti-cheat behavior survives in pipeline artifacts.
- Todo skills align with root `todo.md`; no `backlog.md` added.
- Narrow lanes are kept, merged, or parked only with deletion proof.
- Propagation policy is explicit.
