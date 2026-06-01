---
kb_id: kb-2026-06-01-cold-storage-follow-through
slice_id: slice-094
title: "Decide Go harness rewrite scope"
blockers: []
verification: hitl
test_level: none
functional_risk: none
hitl: true
expected_files:
  - path: docs/context/epics/cold-storage-follow-through.md
    op: edit
    scope: "Record the chosen Go rewrite scope or parked rationale."
protected_oracles: []
status: completed
---

# Slice 094: Decide Go Harness Rewrite Scope

## Human Decision

Choose the scope before any implementation plan replaces PowerShell-backed
behavior:

1. Small native-Go parity spike for `kb-check` discovery/reporting while still
   delegating deeper checks to PowerShell.
2. Full non-PowerShell rewrite of the core local gate.
3. Keep parked until a non-Windows consumer needs it.

## Why This Blocks Planning

The expected files, acceptance criteria, and verification burden differ
substantially between a parity spike and a real rewrite. Planning both would add
noise and likely overbuild.

## Default Recommendation

Keep the full rewrite parked unless there is a concrete non-Windows consumer.
If work is desired now, choose the small native-Go parity spike.

## Decision

Answered 2026-06-01: build the full non-PowerShell rewrite if it works for
Windows+; remove PS1 only after proof.

Follow-up manifest:
`docs/plans/2026-06-01-100-kb-go-native-core-gate-rewrite-manifest.md`.
