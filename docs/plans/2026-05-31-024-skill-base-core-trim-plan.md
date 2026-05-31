---
kb_id: kb-2026-05-31-routing-trim
slice_id: slice-004
title: "Trim base and core workflow skills behind measurements"
blockers: [slice-001, slice-003]
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-first-principles/SKILL.md
    op: edit
    scope: "Trim examples while preserving pushback protocol and stop-before-edit rules."
  - path: .github/skills/kb-start/SKILL.md
    op: edit
    scope: "Keep thin router; remove duplicated long guidance."
  - path: .github/skills/kb-map/SKILL.md
    op: edit
    scope: "Extract or trim bulky bootstrap/reference detail if evals pass."
  - path: .github/skills/kb-brainstorm/SKILL.md
    op: edit
    scope: "Reduce duplicated route/ceremony text without losing HITL behavior."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Extract repeated templates only when the generated output remains valid."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Trim generic prose while preserving scope ledger and repair loop."
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "Trim learning/evolve prose after preserving gates."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Use surface report and route fixtures to trim only safe text."
human_action: ""
can_continue_other_slices: true
---

# Trim Base And Core Workflow Skills Behind Measurements

## What To Build

Trim generic or duplicated prose from base/core KB skills without removing the
workflow gates that make the bundle reliable.

## Acceptance Criteria

- Base route retains `kb-map`, `kb-first-principles`, `kb-check`, and `kb-start`
  behavior.
- `kb-epic` still ends at planning complete, then asks whether to continue.
- `kb-plan` still generates valid manifests and updates `todo.md`.
- `kb-work` still owns scope ledger, verification, and repair loop.
- `kb-complete` still runs review/proof/learning/memory cleanup gates.
- Loaded-surface report shows reduction.
- Route and skill eval gates pass.

## Test Scenarios

- Run loaded-surface before and after.
- Run route-complexity eval.
- Run skill lint and `kb-check -All`.
- Run sync report and inspect required targets.

## Scope Boundary

Do not delete skills in this slice. Deletion needs separate reference and route
proof.
