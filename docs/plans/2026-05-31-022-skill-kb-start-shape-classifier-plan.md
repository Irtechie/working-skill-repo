---
kb_id: kb-2026-05-31-routing-trim
slice_id: slice-002
title: "Add compact shape classifier to kb-start and manifests"
blockers: [slice-001]
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-start/SKILL.md
    op: edit
    scope: "Add cheap workflow-shape classifier."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Add optional workflow_shape manifest field if justified."
  - path: .github/skills/kb-epic/SKILL.md
    op: edit
    scope: "Reference shape routing for multi-stream initiatives."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Add a tiny classifier that rarely escalates ceremony."
human_action: ""
can_continue_other_slices: true
---

# Add Compact Shape Classifier To kb-start And Manifests

## What To Build

Teach `kb-start` to classify work as direct chat, single skill edit,
skill-bundle change, pipeline change, or multi-stream epic. It should not build
a pipeline unless pipeline signals are present.

## Acceptance Criteria

- Classifier is concise enough for base-layer loading.
- Pipeline signals include multi-surface changes, proof harness changes,
  propagation/sync rules, independent workstreams, and deletion/eval risk.
- Route fixtures pass.
- Manifests can record workflow shape when useful.

## Test Scenarios

- Run route-complexity eval.
- Run skill lint.
- Run `kb-check -All`.

## Scope Boundary

Do not create a new workflow-shape skill unless measurement proves `kb-start`
would become bloated.
