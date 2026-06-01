---
kb_id: kb-2026-06-01-cold-storage-follow-through
slice_id: slice-093
title: "Add high-value Copilot path instructions"
blockers: []
verification: integration
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: .github/instructions/skills.instructions.md
    op: create
    scope: "Path-specific guidance for `.github/skills/**` only if it avoids repeated skill-edit mistakes."
  - path: .github/instructions/evals.instructions.md
    op: create
    scope: "Path-specific guidance for eval fixtures and result schemas."
  - path: .github/instructions/scripts.instructions.md
    op: create
    scope: "Path-specific guidance for PowerShell/Go gate scripts."
  - path: README.md
    op: edit
    scope: "Document any added instruction files as repo-local scaffolding."
  - path: config/skill-quality.json
    op: edit
    scope: "Add instruction paths to lint/reference checks if useful."
protected_oracles: []
status: done
---

# Slice 093: Add High-Value Copilot Path Instructions

## What To Build

Add Copilot path-specific instruction files only where a repo-local rule is more
precise than global guidance. Candidate file classes are skills, evals, and
scripts.

## Acceptance Criteria

- Each instruction file has path selectors that match only the intended files.
- Instructions are short and repo-specific.
- No instruction repeats generic "write tests" or "read code" advice.
- README or testing docs explain that these are development scaffolding, not
  installed skills.
- `kb-check -All` and `git diff --check` pass.

## Test Scenarios

- Inspect selectors and ensure they match the intended path class.
- Run `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`.
- Run `git diff --check`.

## Scope Boundary

- Do not create broad instructions that apply to every markdown file.
- Do not move behavior from skills into instructions unless it is truly
  file-class-specific.
- Do not add instructions for app/UI/book workflows in this portable bundle.

## Completion Proof

- Added path-scoped instructions for `.github/skills/**/SKILL.md`,
  `evals/**/*.json,evals/**/*.md`, and `scripts/**/*.ps1,cmd/**/*.go`.
- `go run .\cmd\kbcheck local-release --json` exited 0 after the final diff.
