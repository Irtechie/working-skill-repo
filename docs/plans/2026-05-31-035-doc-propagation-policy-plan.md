---
kb_id: kb-2026-05-31-lazy-lane-consolidation
slice_id: slice-005
title: "Codify propagation policy"
blockers: []
verification: docs-check
test_level: static
functional_risk: none
hitl: true
expected_files:
  - path: AGENTS.md
    op: edit
    scope: "Clarify required vs optional sync targets if needed."
  - path: README.md
    op: edit
    scope: "Document visible bundle propagation policy."
  - path: scripts/skill-sync-report.ps1
    op: edit
    scope: "Ensure required/optional target behavior matches policy."
status: pending
owner: agent
blocked_reason: ""
resume_when: "Human confirms ATV scaffold/plugin thin default or changes it."
next_agent_action: "Codify the propagation policy after confirmation."
human_action: "Confirm optional ATV scaffold/plugin should remain thin for now."
can_continue_other_slices: true
test_inputs:
  - name: propagation_policy
    source: user
    required_for: "README/AGENTS policy wording"
    value: "Default: required globals stay synced; ATV scaffold/plugin may remain thinner and warning-only."
---

# Codify Propagation Policy

## What To Build

Document and enforce which skill targets are required to match this repo and
which ATV scaffold/plugin targets are optional thinner bundles.

## Acceptance Criteria

- Required targets remain zero required issues in sync report.
- Optional ATV scaffold/plugin differences are explained, not accidental.
- README/AGENTS wording matches script behavior.
- Future sync work knows when to copy vs preserve drift.

## Test Scenarios

- Run `scripts/skill-sync-report.ps1`.
- Run `git diff --check`.
- Run `kb-check -All`.

## Scope Boundary

Do not force full KB surface into ATV scaffold/plugin unless the user overrides
the thin-bundle default.
