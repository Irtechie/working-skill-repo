---
kb_id: kb-2026-07-01-live-steering-learning-loop
slice_id: slice-001
title: "Add long-goal control-loop steering contract"
blockers: []
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: ".github/skills/kb-goal/SKILL.md"
    op: edit
    scope: "Add optional live-steering fields to the goal ledger and routing loop."
  - path: ".github/skills/kb-plan/SKILL.md"
    op: edit
    scope: "Teach planning when to include control-loop fields for recurring improvement work."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Edit kb-goal and kb-plan with the optional live-steering contract."
human_action: ""
can_continue_other_slices: true
---

# Slice 001 - Long-Goal Control-Loop Steering Contract

## What To Build

Add an optional control-loop contract for `kb-goal` and `kb-plan` so long-lived
or recurring improvement work can be framed as a measurable loop without forcing
every goal into that shape.

## Acceptance Criteria

- `kb-goal` goal ledgers can record set point, sensor, controller, actuator,
  disturbances, dampener, scope gate, batch size, WIP bound, and steering memory
  path.
- `kb-goal` says to use this block only when the goal is recurring, scheduled,
  or trend-improvement work.
- `kb-plan` tells planners to include those fields for applicable slices and to
  avoid inventing fake separation when sensor/controller/actuator are fused.
- The text explicitly rejects importing HumanLayer-specific CI/tool defaults.

## Verification

- `go run ./cmd/kbcheck core`
- `git diff --check`

## Scope Boundary

No scripts, workflow YAML, or runner integration in this slice.
