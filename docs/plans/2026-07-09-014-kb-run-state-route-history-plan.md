---
kb_id: kb-2026-07-09-phoenix-routing-slicing-absorption
slice_id: slice-004
title: "Define KB-native run state and route-history guards"
blockers: [slice-002]
verification: integration
test_level: integration
functional_risk: none
model_tier: large
model_route: hosted-large
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck run-state-selftest"
  expect: 0
hitl: false
expected_files:
  - path: ".github/skills/kb-goal/SKILL.md"
    op: edit
    scope: "Define .kb/runs/<goal>/ state and route-history expectations for durable goals."
  - path: ".github/skills/kb-start/SKILL.md"
    op: edit
    scope: "Add oscillation/confidence guard routing behavior when run-state exists."
  - path: ".github/skills/kb-work/SKILL.md"
    op: edit
    scope: "Record route-history and progress events for manifest execution when active."
  - path: "docs/context/architecture/kb-workflow.md"
    op: edit
    scope: "Document the KB-native run-state shape."
  - path: "cmd/kbcheck"
    op: edit
    scope: "Add lightweight route-history/run-state validation only if needed to make the guard executable."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Design the smallest .kb/runs state shape that gives Phoenix/Ralph-style persistence without an external loop driver."
human_action: ""
can_continue_other_slices: true
---

# Slice 004: Define KB-Native Run State And Route-History Guards

## What This Delivers

KB gets the useful Ralph/Phoenix persistence idea without adopting
`.phoenix-ralph/` or a Phoenix driver. The state is repo-local, git-ignored, and
used only for durable goals or autonomous loops.

## Acceptance Criteria

- `.kb/runs/<goal>/` shape is documented with at least:
  `goal.md`, `done-check.json`, `backlog.json`, `progress.md`, and
  `route-history.jsonl`.
- `route-history.jsonl` has enough fields to detect repeated route oscillation,
  repeated low-confidence choices, and no-progress loops.
- `kb-start`/`kb-goal` rules say when to stop and re-plan instead of looping.
- The design keeps `todo.md`, manifests, and handoffs as the durable human
  surfaces; `.kb/runs` stays ephemeral run state.

## Test Scenarios

- Fixture route history with A/B/A/B oscillation is rejected or flagged.
- Fixture route history with three low-confidence routes and no state change
  requires re-plan or human clarification.

## Scope Boundary

No external Copilot loop driver. No MCP server. No Phoenix `.phoenix-ralph`
state directory.
