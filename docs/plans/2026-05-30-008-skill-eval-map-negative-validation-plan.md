---
kb_id: kb-2026-05-30-live-cross-runtime-skill-eval-harness
slice_id: slice-008
title: "Add eval-map scaffold negative validation"
blockers: []
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: .github/skills/kb-eval-map/SKILL.md
    op: edit
    scope: "tighten scaffold validation output requirements for pass and intentional negative checks"
  - path: docs/context/eval-map.md
    op: edit
    scope: "document current negative-validation contract for future consuming repos"
  - path: docs/context/operations/testing.md
    op: edit
    scope: "document how scaffolded smoke eval validation should be recorded"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 008: Add Eval-Map Scaffold Negative Validation

## What To Build

Strengthen `kb-eval-map` so future scaffolded smoke evals must record both a
passing command and an intentional negative check that failed as expected. This
turns "smoke eval exists" into "smoke eval can actually fail."

## Acceptance Criteria

- `kb-eval-map` requires pass-command and negative-check evidence in eval-map or
  testing docs for scaffolded smoke evals.
- The negative check language covers UI, API, CLI, file/config, and generated
  artifact workflows.
- Docs explain that failed negative validation means the smoke eval must be
  rewritten or deleted before reporting success.

## Verification

- Manual review of `kb-eval-map` scaffold policy against this slice.
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`
- `git diff --check`
