---
kb_id: kb-2026-06-01-go-native-core-gate-rewrite
slice_id: slice-101
title: "Define native Go gate parity contract"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: medium
hitl: false
expected_files:
  - path: cmd/kbcheck/main_test.go
    op: edit
    scope: "Add parity-contract tests for command shapes and missing-script behavior before rewriting execution."
  - path: cmd/kbcheck/parity_test.go
    op: create
    scope: "Define fixture repos and expected check lists for native Go parity."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document the parity contract and PS1 removal rule."
protected_oracles:
  - path: cmd/kbcheck/main_test.go
    role: "existing wrapper behavior oracle"
    sha256: "filled by kb-work before edits"
    update_policy: "requires explicit plan update"
status: done
---

# Slice 101: Define Native Go Gate Parity Contract

## What To Build

Before rewriting behavior, encode the current PS1 gate contract in Go tests:
check discovery, command names, required/optional release-gate behavior,
JSON/text output shape, and failure propagation.

## Acceptance Criteria

- Tests define expected `core` check discovery for this repo and simple fixture
  repos.
- Tests define `local-release` required/optional check semantics.
- Tests define that final native paths must not require `pwsh`/`powershell`.
- Docs state PS1 removal is blocked until parity proof passes.

## Test Scenarios

- `go test ./...`
- `go run .\cmd\kbcheck core --list` after the list mode exists in later
  slices, or equivalent fixture-level test before CLI exposure.

## Scope Boundary

- Do not remove PS1.
- Do not replace implementation yet except where needed to make tests compile.

## Completion Proof

- Added Go tests for core list/run/failure behavior, release failure reporting,
  and skill-repo check-name parity.
- `go test ./...` exited 0.
