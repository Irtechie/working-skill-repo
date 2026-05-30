---
kb_id: kb-2026-05-30-eval-map
slice_id: slice-001
title: "Add kb-eval-map skill"
blockers: []
verification: verification-only
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-eval-map/SKILL.md
    op: create
    scope: "new skill contract for repo eval-surface discovery, native harness selection, intent gate, scaffolding, and kb-check wiring"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 001: Add KB Eval Map Skill

## What To Build

Create `kb-eval-map` as the dedicated setup skill for repository-native eval
mapping.

## Acceptance Criteria

- The skill explains its boundary: policy in skills, executable eval judgment in
  repo-local scripts/tests/evals.
- It has an intent gate for new or unclear repos.
- It has a pattern matrix for website, internal website, API, CLI, LLM/agent app,
  skill repo, docs/process repo, mobile/native, and mixed repos.
- It defines output files, safe scaffolding rules, and `kb-check` wiring rules.
- It includes frontmatter with `name`, `description`, and `argument-hint`.

## Verification

- Run `scripts/skill-lint.ps1`.
- Run `.\.github\skills\kb-check\scripts\kb-check.ps1 -All` after all slices.

Result: `scripts/skill-lint.ps1` exited 0 with 19 known warnings.
