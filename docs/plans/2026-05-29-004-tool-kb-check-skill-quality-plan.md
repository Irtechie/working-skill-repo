---
kb_id: kb-2026-05-29-cross-runtime-skill-quality
slice_id: slice-004
title: "Wire skill quality into kb-check"
blockers: [slice-002, slice-003]
verification: functional-cli
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-check/scripts/kb-check.ps1
    op: edit
    scope: "discover and run skill repo quality checks when skill repo files are present"
  - path: docs/context/operations/testing.md
    op: edit
    scope: "document canonical kb-check invocation and clean output"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 004: Wire Skill Quality Into KB Check

## What To Build

Make the existing KB check entry point discover and run the new skill quality checks in this repo.

## Acceptance Criteria

- `.\.github\skills\kb-check\scripts\kb-check.ps1 -List` reports skill lint and route-complexity eval checks in this repo.
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All` runs the skill lint and route-complexity eval commands.
- Existing app-oriented discovery still works for package, Python, .NET, and Makefile projects.
- Testing docs explain the canonical command for this repo.

## Expected Files

- `.github/skills/kb-check/scripts/kb-check.ps1`
- `docs/context/operations/testing.md`

## Test Scenarios

- Run `.\.github\skills\kb-check\scripts\kb-check.ps1 -List`.
- Run `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`.
- Confirm the command exits nonzero if one of the underlying quality commands fails.

## Scope Boundary

This slice does not add new lint rules or eval fixtures beyond invoking what earlier slices created.

## Dependencies

- Requires slice-002 skill lint.
- Requires slice-003 route-complexity eval runner.
