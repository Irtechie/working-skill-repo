---
kb_id: kb-2026-05-31-learning-landmines
slice_id: slice-001
title: "Define local landmine schema and lifecycle"
blockers: []
verification: docs-check
test_level: static
functional_risk: none
hitl: false
expected_files:
  - path: docs/context/landmines.md
    op: add
    scope: "Create only if active landmines exist; define schema and lifecycle."
  - path: docs/context/PROJECT.md
    op: edit
    scope: "Point to landmine storage when present."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document any proof command for landmine lifecycle checks."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Create a compact schema and lifecycle without turning landmines into generic advice."
human_action: ""
can_continue_other_slices: true
---

# Define Local Landmine Schema And Lifecycle

## What To Build

Define the repo-local format for active landmines and resolved landmine
archives. A landmine must be a verified high-risk trap with an owner surface,
not a generic lesson.

## Acceptance Criteria

- Schema includes owner surface, evidence, severity, fix condition,
  verification, status, and archive reason.
- Fixed and verified landmines archive immediately.
- Stale-but-unfixed landmines remain reviewable instead of silently disappearing.
- `PROJECT.md` tells future sessions where active landmines live.

## Test Scenarios

- Run markdown/static checks through `kb-check -All`.
- Confirm no generic landmine examples are added as active facts.

## Scope Boundary

Do not add `backlog.md`. Do not promote landmines into skills in this slice.
