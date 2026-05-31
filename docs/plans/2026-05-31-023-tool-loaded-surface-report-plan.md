---
kb_id: kb-2026-05-31-routing-trim
slice_id: slice-003
title: "Implement loaded-surface reporting"
blockers: []
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-surface-report.ps1
    op: add
    scope: "Report route-level skill line/token estimates and hashes."
  - path: config/skill-quality.json
    op: edit
    scope: "Configure surface thresholds or allowlists."
  - path: .github/skills/kb-check/scripts/kb-check.ps1
    op: edit
    scope: "Run the report or selftest in kb-check."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document measurement commands."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Implement measurement before trimming text."
human_action: ""
can_continue_other_slices: true
---

# Implement Loaded-Surface Reporting

## What To Build

Create a deterministic report that estimates loaded lines/tokens by route and
captures content hashes so trims can be measured instead of guessed.

## Acceptance Criteria

- Report covers base layer and named KB routes.
- Output includes skill list, line count, rough token estimate, and hash.
- Report can compare before/after snapshots.
- `kb-check -All` can run a non-noisy validation path.

## Test Scenarios

- Run report for base route.
- Run report for `kb-epic`, `kb-plan`, and `kb-work`.
- Confirm before/after compare detects a changed skill hash.
- Run `kb-check -All`.

## Scope Boundary

This does not decide which skills to delete. It produces the measurement needed
for deletion and trim decisions.
