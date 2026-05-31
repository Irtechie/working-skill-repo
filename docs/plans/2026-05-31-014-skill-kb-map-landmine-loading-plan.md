---
kb_id: kb-2026-05-31-learning-landmines
slice_id: slice-004
title: "Load and resolve active landmines through KB workflows"
blockers: [slice-001]
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-map/SKILL.md
    op: edit
    scope: "Load only concise active landmines during memory preflight."
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "Capture and resolve landmines after verified work."
  - path: .github/skills/kb-memory-review/SKILL.md
    op: edit
    scope: "Refresh stale or resolved landmine entries."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Wire active landmine lookup and resolved archive behavior into existing KB memory flow."
human_action: ""
can_continue_other_slices: true
---

# Load And Resolve Active Landmines Through KB Workflows

## What To Build

Teach KB memory workflows to surface only active landmines and resolve them when
the owning surface is fixed and verified.

## Acceptance Criteria

- Startup does not load archived or generic landmine notes.
- `kb-complete` can add a landmine only with evidence and owner surface.
- A fixed landmine moves to resolved/archive with proof.
- `kb-map refresh` can identify stale unresolved entries for review.

## Test Scenarios

- Create a sample active landmine in a temp fixture or doc section.
- Confirm instructions route it into startup context only when active.
- Run skill lint and `kb-check -All`.

## Scope Boundary

Do not make landmine learning global. This remains repo-local memory.
