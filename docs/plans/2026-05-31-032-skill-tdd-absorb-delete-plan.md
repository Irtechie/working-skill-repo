---
kb_id: kb-2026-05-31-lazy-lane-consolidation
slice_id: slice-002
title: "Absorb TDD anti-cheat behavior into pipeline"
blockers: []
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/tdd/SKILL.md
    op: edit
    scope: "Trim, park, or delete only after anti-cheat behavior moves."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Require protected oracle fields when known before implementation."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Preserve protected oracle during implementation."
  - path: scripts/skill-lint.ps1
    op: edit
    scope: "Detect stale tdd references if deletion occurs."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Move protected test-first oracle behavior into the pipeline before deleting TDD."
human_action: ""
can_continue_other_slices: true
---

# Absorb TDD Anti-Cheat Behavior Into Pipeline

## What To Build

Move the valuable TDD landmine into KB planning/work: define oracle first when
practical, prove RED when possible, protect the oracle with SHA/manifest,
implement, rerun. Then decide whether standalone `tdd` can be deleted or parked.

## Acceptance Criteria

- Protected oracle behavior exists outside standalone `tdd`.
- Standalone `tdd` is no longer required for normal KB work.
- References to `tdd` are updated or intentionally preserved.
- Route fixtures cover explicit test-first requests if `tdd` remains lazy.

## Test Scenarios

- Run skill reference scanner.
- Run protected-oracle selftest.
- Run `kb-check -All`.

## Scope Boundary

Do not delete `tdd` until references are clean and the anti-cheat behavior is
available in the main pipeline.
