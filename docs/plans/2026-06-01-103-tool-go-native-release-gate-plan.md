---
kb_id: kb-2026-06-01-go-native-core-gate-rewrite
slice_id: slice-103
title: "Implement native Go release gate"
blockers: [slice-102]
verification: integration
test_level: functional-cli
functional_risk: medium
hitl: false
expected_files:
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Route release commands through native Go orchestration."
  - path: cmd/kbcheck/release.go
    op: create
    scope: "Implement local-release/live-release check composition and result summaries."
  - path: cmd/kbcheck/release_test.go
    op: create
    scope: "Test profile selection, skipped-explicit labeling, and failure propagation."
  - path: scripts/kb-release-gate.ps1
    op: edit
    scope: "Mark as transitional fallback if still present."
  - path: README.md
    op: edit
    scope: "Document Go release gate as preferred while PS1 remains fallback pending parity proof."
protected_oracles: []
status: done
---

# Slice 103: Implement Native Go Release Gate

## What To Build

Move `local-release` and `live-release` orchestration into Go. Preserve the
current semantics: required local checks block, optional static reports can
fail, live model corpus is explicit, and unavailable live surfaces are labeled
`skipped-explicit`.

## Acceptance Criteria

- `go run .\cmd\kbcheck local-release` does not invoke
  `scripts/kb-release-gate.ps1`.
- `go run .\cmd\kbcheck local-release --json` emits structured summary.
- `live-release` keeps live model calls explicit.
- Required failure and optional failure propagation are covered by tests.

## Test Scenarios

- `go test ./...`
- `go run .\cmd\kbcheck local-release`
- `go run .\cmd\kbcheck local-release --json`
- Failure fixture for a required check exits nonzero.

## Scope Boundary

- Do not remove PS1 until parity proof passes.
- Do not make live model calls part of default local release.

## Completion Proof

- `scripts/kb-release-gate-selftest.ps1` exited 0 against the Go release path.
- `go run .\cmd\kbcheck local-release --json` exited 0.
