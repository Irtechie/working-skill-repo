---
kb_id: kb-2026-05-29-cross-runtime-skill-quality
slice_id: slice-003
title: "Add route and complexity eval fixtures"
blockers: [slice-001]
verification: functional-cli
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: evals/route-complexity/README.md
    op: create
    scope: "fixture schema and expected route behavior"
  - path: evals/route-complexity/*.json
    op: create
    scope: "initial deterministic route-complexity fixtures"
  - path: scripts/route-complexity-eval.ps1
    op: create
    scope: "fixture validator and deterministic rubric checker"
  - path: config/skill-quality.json
    op: edit
    scope: "register eval fixture directory and rubric settings"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 003: Add Route Complexity Eval Fixtures

## What To Build

Create a deterministic eval fixture corpus that catches route mis-sizing without requiring live model calls in the first pass.

## Acceptance Criteria

- Fixture schema includes prompt, repo state, platform applicability, expected route, expected questions, expected artifacts, expected proof, and complexity signals.
- Initial fixtures cover tiny fix, known failing test, unclear broken behavior, bounded feature with "don't ask many questions", stale handoff, broad migration, release flow, and cross-runtime instruction update.
- A runner validates every fixture has required fields and that rubric signals map to the expected route category.
- The fixture docs explain that live model benchmarking is a later layer, not part of the first deterministic harness.

## Expected Files

- `evals/route-complexity/README.md`
- `evals/route-complexity/*.json`
- `scripts/route-complexity-eval.ps1`
- `config/skill-quality.json`

## Test Scenarios

- Run `powershell -ExecutionPolicy Bypass -File scripts/route-complexity-eval.ps1`.
- Verify all fixtures parse.
- Verify at least one under-sizing and one over-planning guard fixture is present.

## Scope Boundary

This slice does not call external models and does not change `kb-start` or `kb-task` yet.

## Dependencies

- Requires slice-001 config contract.
