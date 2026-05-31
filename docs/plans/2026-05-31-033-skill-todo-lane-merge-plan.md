---
kb_id: kb-2026-05-31-lazy-lane-consolidation
slice_id: slice-003
title: "Merge todo skills around root todo.md"
blockers: []
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/todo-create/SKILL.md
    op: edit
    scope: "Align durable todo workflow with root todo.md for KB."
  - path: .github/skills/todo-triage/SKILL.md
    op: edit
    scope: "Merge or mark as legacy/non-KB after caller audit."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Keep root todo.md as active board."
  - path: README.md
    op: edit
    scope: "Document todo.md vs any legacy CE todo path."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Resolve mismatch between root todo.md and legacy file-per-todo instructions."
human_action: ""
can_continue_other_slices: true
---

# Merge Todo Skills Around Root todo.md

## What To Build

Align `todo-create` and `todo-triage` with the KB root `todo.md` active
backlog/board model. Avoid adding `backlog.md` unless real usage proves
`todo.md` cannot carry the job.

## Acceptance Criteria

- KB planning still updates root `todo.md`.
- Todo skills no longer confuse KB workflows with legacy file-per-todo storage.
- If legacy CE todos remain, they are clearly labeled non-KB or compatibility.
- `todo-triage` is merged, parked, or retained only if references justify it.

## Test Scenarios

- Run reference scanner.
- Run skill lint and `kb-check -All`.
- Confirm `kb-plan` todo instructions still produce a clear active board.

## Scope Boundary

Do not introduce `backlog.md` in this slice.
