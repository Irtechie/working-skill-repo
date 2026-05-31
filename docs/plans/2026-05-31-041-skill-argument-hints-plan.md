---
kb_id: kb-2026-05-31-warning-quality-cleanup
slice_id: slice-001
title: "Add missing argument hints"
blockers: []
verification: integration
test_level: cli
functional_risk: none
hitl: false
expected_files:
  - path: .github/skills/ce-compound/SKILL.md
    op: edit
    scope: "Add argument-hint frontmatter only."
  - path: .github/skills/ce-compound-refresh/SKILL.md
    op: edit
    scope: "Add argument-hint frontmatter only."
  - path: .github/skills/evolve/SKILL.md
    op: edit
    scope: "Add argument-hint frontmatter only."
  - path: .github/skills/kb-first-principles/SKILL.md
    op: edit
    scope: "Add argument-hint frontmatter only."
  - path: .github/skills/kb-task/SKILL.md
    op: edit
    scope: "Add argument-hint frontmatter only."
  - path: .github/skills/kb-troubleshoot/SKILL.md
    op: edit
    scope: "Add argument-hint frontmatter only."
  - path: .github/skills/learn/SKILL.md
    op: edit
    scope: "Add argument-hint frontmatter only."
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Patch missing argument-hint fields, then run skill lint."
human_action: ""
can_continue_other_slices: true
---

# Add Missing Argument Hints

## What To Build

Add concise `argument-hint` frontmatter to skills that currently warn only
because the field is missing.

## Acceptance Criteria

- `scripts/skill-lint.ps1` no longer reports missing `argument-hint` for these
  skills.
- No skill body behavior changes.

## Test Scenarios

- `powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1`
- `git diff --check`

## Scope Boundary

Do not trim long skills in this slice.
