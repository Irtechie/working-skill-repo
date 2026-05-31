---
kb_id: kb-2026-05-31-routing-trim
slice_id: slice-001
title: "Add workflow-shape route fixtures"
blockers: []
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: evals/route-complexity/
    op: edit
    scope: "Add fixtures for skill edit, skill-bundle change, pipeline change, and multi-stream epic."
  - path: scripts/route-complexity-eval.ps1
    op: edit
    scope: "Score workflow shape if current fixture schema needs it."
  - path: docs/context/eval-map.md
    op: edit
    scope: "Document route-shape fixture coverage."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Add fixtures before changing router text."
human_action: ""
can_continue_other_slices: true
---

# Add Workflow-Shape Route Fixtures

## What To Build

Add deterministic route fixtures that prove small work stays small and real
multi-surface work escalates to `kb-epic`.

## Acceptance Criteria

- A one-file skill edit does not route to epic.
- A skill plus sync/eval/docs change routes to planning.
- A proof harness or coded pipeline change routes to pipeline/epic shape.
- A multi-stream initiative routes to `kb-epic`.
- Existing route fixtures remain green.

## Test Scenarios

- Run `scripts/route-complexity-eval.ps1`.
- Run `kb-check -All`.

## Scope Boundary

Do not edit `kb-start` until fixtures exist.
