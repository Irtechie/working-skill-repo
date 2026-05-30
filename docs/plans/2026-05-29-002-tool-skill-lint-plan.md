---
kb_id: kb-2026-05-29-cross-runtime-skill-quality
slice_id: slice-002
title: "Add deterministic skill lint"
blockers: [slice-001]
verification: functional-cli
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-lint.ps1
    op: create
    scope: "deterministic skill repository linter"
  - path: config/skill-quality.json
    op: edit
    scope: "add or refine lint rule settings and allowlists discovered during implementation"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 002: Add Deterministic Skill Lint

## What To Build

Add a PowerShell linter for the skill repo. The linter should validate structural health of skill files without using Codex-only or GHCP-only APIs.

## Acceptance Criteria

- `scripts/skill-lint.ps1` runs from the repo root.
- It validates `SKILL.md` frontmatter has `name` and `description`.
- It reports missing `argument-hint` as warning or failure based on config.
- It validates `@./references/*` links and script references resolve.
- It detects unresolved conflict markers in skills, agents, references, scripts, and docs.
- It checks hot-path skill line budgets against `config/skill-quality.json`.
- It exits nonzero for configured failures and zero when only warnings are present.

## Expected Files

- `scripts/skill-lint.ps1`
- `config/skill-quality.json`

## Test Scenarios

- Run `powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1`.
- Run a mode that emits warnings without failing for known allowlisted long skills.
- Temporarily test one bad fixture or isolated sample if the script supports fixture paths; do not leave intentional bad files in the repo.

## Scope Boundary

This slice does not run route-complexity evals and does not modify `kb-check`.

## Dependencies

- Requires slice-001 config contract.
