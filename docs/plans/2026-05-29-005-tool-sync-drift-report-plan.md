---
kb_id: kb-2026-05-29-cross-runtime-skill-quality
slice_id: slice-005
title: "Add read-only sync drift report"
blockers: [slice-001]
verification: functional-cli
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-sync-report.ps1
    op: create
    scope: "read-only skill and agent target hash report"
  - path: config/skill-quality.json
    op: edit
    scope: "sync target metadata and required/optional classifications"
  - path: docs/context/operations/testing.md
    op: edit
    scope: "document sync report command as diagnostic, not mutation"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 005: Add Read-Only Sync Drift Report

## What To Build

Add a read-only drift report that compares skill copies across the working source, Codex global, Copilot global, shared agents global, and ATV targets.

## Acceptance Criteria

- `scripts/skill-sync-report.ps1` reports per-skill hashes for configured targets.
- Required target drift is reported as failure or must-fix.
- Optional or intentional omissions are reported as warnings or informational rows.
- The script suggests copy direction but never writes to any target.
- The report handles missing optional paths without failing the whole command.

## Expected Files

- `scripts/skill-sync-report.ps1`
- `config/skill-quality.json`
- `docs/context/operations/testing.md`

## Test Scenarios

- Run `powershell -ExecutionPolicy Bypass -File scripts/skill-sync-report.ps1`.
- Confirm output includes Codex, Copilot/GHCP, shared agents, and ATV target classifications.
- Confirm the script does not mutate files by comparing `git status --short` before and after.

## Scope Boundary

This slice does not perform sync or copy operations.

## Dependencies

- Requires slice-001 config contract.
