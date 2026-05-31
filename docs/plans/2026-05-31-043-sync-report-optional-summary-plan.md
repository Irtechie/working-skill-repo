---
kb_id: kb-2026-05-31-warning-quality-cleanup
slice_id: slice-003
title: "Compact optional sync warnings"
blockers: []
verification: integration
test_level: cli
functional_risk: none
hitl: false
expected_files:
  - path: scripts/skill-sync-report.ps1
    op: edit
    scope: "Summarize optional differences by default and expose detailed rows behind a switch."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document compact optional output and detailed switch."
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Patch sync report output, then run sync report and full gate."
human_action: ""
can_continue_other_slices: true
---

# Compact Optional Sync Warnings

## What To Build

Keep optional ATV scaffold/plugin drift visible without printing every optional
row by default.

## Acceptance Criteria

- Required sync issues still fail and print detailed errors.
- Optional differences are summarized by default.
- Detailed optional rows are available with an explicit switch.

## Test Scenarios

- `powershell -ExecutionPolicy Bypass -File scripts/skill-sync-report.ps1`
- `powershell -ExecutionPolicy Bypass -File scripts/skill-sync-report.ps1 -VerboseOptional`
- `powershell -ExecutionPolicy Bypass -File .\.github\skills\kb-check\scripts\kb-check.ps1 -All`

## Scope Boundary

Do not sync full KB surface into optional ATV scaffold/plugin bundles.
