---
kb_id: kb-2026-05-31-learning-landmines
slice_id: slice-002
title: "Add landmine evidence fields to learn flow"
blockers: [slice-001]
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/learn/SKILL.md
    op: edit
    scope: "Add landmine candidate capture fields and reject generic lessons."
  - path: config/skill-quality.json
    op: edit
    scope: "Add lint expectations if needed."
  - path: scripts/skill-lint.ps1
    op: edit
    scope: "Optionally detect generic landmine wording."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Review learn workflow and add landmine-specific capture only where it earns tokens."
human_action: ""
can_continue_other_slices: true
---

# Add Landmine Evidence Fields To Learn Flow

## What To Build

Constrain `learn` so landmine candidates carry evidence, owner surface, severity,
and fix condition. Keep the existing confidence and recency-decay model.

## Acceptance Criteria

- `learn` distinguishes ordinary instincts from landmine candidates.
- Landmine candidates require evidence and owner surface.
- Existing numeric scoring remains intact.
- Generic advice cannot become a landmine without a concrete failure mode.

## Test Scenarios

- Run skill lint.
- Add or update deterministic lint fixture if the linter grows landmine checks.
- Run `kb-check -All`.

## Scope Boundary

This does not generate `learned-*` skills. Promotion belongs to `evolve`.
