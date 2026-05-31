---
kb_id: kb-2026-05-31-routing-trim
slice_id: slice-005
title: "Clean reference graph and deletion blockers"
blockers: []
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-brainstorm/SKILL.md
    op: edit
    scope: "Fix or remove unknown ce-ideate reference."
  - path: README.md
    op: edit
    scope: "Fix or remove legacy kb-route reference."
  - path: scripts/skill-lint.ps1
    op: edit
    scope: "Keep scanner useful for invocation-shaped references."
  - path: config/skill-quality.json
    op: edit
    scope: "Adjust allowlists only with explicit rationale."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Resolve known unknown refs and keep deletion blockers meaningful."
human_action: ""
can_continue_other_slices: true
---

# Clean Reference Graph And Deletion Blockers

## What To Build

Clean known unknown skill references and make the reference scanner a reliable
deletion blocker before trimming or merging skills.

## Acceptance Criteria

- `ce-ideate` warning is fixed, renamed, or explicitly allowlisted with reason.
- `kb-route` warning is fixed, renamed, or explicitly allowlisted with reason.
- Scanner remains low-noise and catches invocation-shaped references.
- `kb-check -All` still exits 0.

## Test Scenarios

- Run `scripts/skill-lint.ps1`.
- Run `kb-check -All`.

## Scope Boundary

Do not remove a referenced skill until callers and fixtures are updated.
