---
kb_id: kb-2026-06-10-skill-bundle-cleanup
slice_id: slice-002
title: "Make the fresh-clone gate and Go toolchain contributor-safe"
blockers: [slice-001]
verification: integration
test_level: functional-cli
functional_risk: broad
hitl: false
expected_files:
  - path: go.mod
    op: edit
    scope: "Lower or justify the Go directive and toolchain requirements based on confirmed build needs."
  - path: cmd/kbcheck
    op: edit
    scope: "Make core gate handling explicit for missing optional sync targets or fresh-clone environments."
  - path: config/skill-quality.json
    op: edit
    scope: "Classify required versus optional sync targets so core and release gates behave intentionally."
  - path: README.md
    op: edit
    scope: "Document the contributor-safe core gate and any setup prerequisites."
  - path: .github/copilot-instructions.md
    op: edit
    scope: "Keep Copilot startup guidance aligned with the gate behavior."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Update canonical proof commands and expected fresh-clone behavior."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Use slice 001 evidence to patch only confirmed gate/toolchain friction."
human_action: ""
can_continue_other_slices: true
---

# Slice 002: Make the fresh-clone gate and Go toolchain contributor-safe

## What To Build

Make the documented quality gate usable by a new contributor without requiring
preinstalled global skill roots or an adjacent ATV checkout, unless the command
explicitly opts into release/sync enforcement.

## Acceptance Criteria

- `core` clearly distinguishes repo-quality failures from missing optional local
  install state.
- Required sync drift remains blocking where configured as required.
- Fresh-clone docs explain which command is contributor-safe and which command
  requires installed targets.
- Go version policy is the minimum actually required, or the higher version is
  justified in docs.

## Expected Files

- `go.mod`
- `cmd/kbcheck`
- `config/skill-quality.json`
- `README.md`
- `.github/copilot-instructions.md`
- `docs/context/operations/testing.md`

## Test Scenarios

- `go test ./...`
- `go run ./cmd/kbcheck core`
- Missing-target fixture or temporary clone proving fresh-clone behavior.
- Existing local-release or sync-drift check still blocks real required drift.

## Scope Boundary

Do not weaken release-grade sync enforcement. If a target must remain required,
make the contributor command separate instead of silently ignoring drift.

## Dependencies

Blocked by `slice-001` so stale audit claims do not drive unnecessary changes.
