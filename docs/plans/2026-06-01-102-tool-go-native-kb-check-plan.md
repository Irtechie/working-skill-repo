---
kb_id: kb-2026-06-01-go-native-core-gate-rewrite
slice_id: slice-102
title: "Implement native Go kb-check core runner"
blockers: [slice-101]
verification: integration
test_level: functional-cli
functional_risk: medium
hitl: false
expected_files:
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Route `core` to native Go execution instead of PowerShell delegation."
  - path: cmd/kbcheck/checks.go
    op: create
    scope: "Implement check discovery equivalent to kb-check.ps1."
  - path: cmd/kbcheck/checks_test.go
    op: create
    scope: "Test check discovery and failure propagation."
  - path: .github/skills/kb-check/SKILL.md
    op: edit
    scope: "Document Go-native core gate as the preferred entrypoint while PS1 remains fallback."
  - path: README.md
    op: edit
    scope: "Update platform reality to describe native Go core gate status."
protected_oracles: []
status: done
---

# Slice 102: Implement Native Go kb-check Core Runner

## What To Build

Make `go run .\cmd\kbcheck core` discover and execute the same local checks as
`.github/skills/kb-check/scripts/kb-check.ps1 -All`, without invoking
PowerShell for the core runner itself.

## Acceptance Criteria

- `core --list` prints the discovered checks.
- `core` runs Go, package, Python, .NET, Makefile, and skill-repo checks using
  direct process execution.
- Check failure exits nonzero and names the failed check.
- Existing PS1 gate remains available as fallback during transition.

## Test Scenarios

- `go test ./...`
- `go run .\cmd\kbcheck core --list`
- `go run .\cmd\kbcheck core`
- Compare discovered check names against PS1 `kb-check.ps1 -List`.

## Scope Boundary

- This slice may still run existing PS1 child scripts as individual checks.
  It must not delegate the whole core gate to `kb-check.ps1`.
- Do not remove PS1.

## Completion Proof

- `go run .\cmd\kbcheck core --list` lists native-discovered checks.
- `go run .\cmd\kbcheck core` exited 0 after the concurrency fix.
