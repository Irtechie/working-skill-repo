---
kb_id: kb-2026-05-31-proof-pipeline-spike
slice_id: slice-001
title: "Persist skill-eval regression baselines"
blockers: []
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-eval.ps1
    op: edit
    scope: "Add baseline load/save/compare flags and regression failure behavior."
  - path: scripts/skill-eval-run-codex.ps1
    op: edit
    scope: "Pass baseline options through for dry-run/live result scoring if needed."
  - path: scripts/skill-eval-run-ghcp.ps1
    op: edit
    scope: "Mirror baseline option support for GHCP adapter if needed."
  - path: evals/skill-eval/baselines/
    op: add
    scope: "Persist baseline JSON for deterministic fixture/scorer expectations."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document baseline update and regression commands."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Inspect current skill-eval output model and add a minimal baseline compare path."
human_action: ""
can_continue_other_slices: true
---

# Persist Skill-Eval Regression Baselines

## What To Build

Add a deterministic baseline mode to `skill-eval` so a run can be compared
against a persisted baseline and fail when a score, pass/fail result, issue
count, or required proof field regresses.

## Acceptance Criteria

- A baseline file can be created from known passing selftests.
- A future run can compare against that baseline.
- Meaningful regressions fail the command, not just report text.
- Intentional baseline updates require an explicit flag.
- `kb-check -All` includes a selftest proving regression failure.

## Test Scenarios

- Create or use a baseline from existing selftest fixtures.
- Compare unchanged results and pass.
- Mutate a copied result to simulate worse status or missing proof and fail.
- Run `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`.

## Scope Boundary

This does not run live model calls by default. It only covers deterministic
captured-result scoring and dry-run paths already included in `kb-check -All`.
