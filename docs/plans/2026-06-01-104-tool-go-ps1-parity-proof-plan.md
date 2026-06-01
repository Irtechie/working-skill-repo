---
kb_id: kb-2026-06-01-go-native-core-gate-rewrite
slice_id: slice-104
title: "Prove PS1 and Go parity on Windows"
blockers: [slice-103]
verification: integration
test_level: functional-cli
functional_risk: medium
hitl: false
expected_files:
  - path: scripts/go-ps1-parity-report.ps1
    op: create
    scope: "Run PS1 and Go gates side by side on Windows and write a parity report."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Record parity proof command and removal gate."
  - path: docs/reports/go-gate-parity-2026-06-01.md
    op: create
    scope: "Persist the first parity proof report."
protected_oracles: []
status: done
---

# Slice 104: Prove PS1 And Go Parity On Windows

## What To Build

Create and run a parity proof that compares PS1 and Go gate outputs before any
PS1 removal. This repo currently proves Windows locally; broader OS proof can
be added when a macOS/Linux environment is available.

## Acceptance Criteria

- Parity script runs old PS1 core/release gates and new Go core/release gates.
- Report records command, exit code, check names, and differences.
- Removal is blocked if required checks differ or any gate fails unexpectedly.
- The report is durable under `docs/reports/`.

## Test Scenarios

- `pwsh -NoProfile -File scripts/go-ps1-parity-report.ps1`
- `go run .\cmd\kbcheck core`
- `go run .\cmd\kbcheck local-release`
- Existing PS1 gates until removal.

## Scope Boundary

- Do not remove PS1 in this slice.
- Do not claim macOS/Linux proof from a Windows-only run.

## Completion Proof

- `scripts/go-ps1-parity-report.ps1` exited 0.
- `docs/reports/go-gate-parity-2026-06-01.md` records `Result: PASS`, no
  missing Go checks, no extra Go checks, and exit 0 for PS/Go core and local
  release on Windows.
