---
kb_id: kb-2026-06-01-go-native-core-gate-rewrite
slice_id: slice-105
title: "Remove PS1 gate dependency after proof"
blockers: [slice-104]
verification: integration
test_level: functional-cli
functional_risk: broad
hitl: false
expected_files:
  - path: .github/skills/kb-check/scripts/kb-check.ps1
    op: delete
    scope: "Remove only if Go parity proof passes and docs/globals are updated."
  - path: scripts/kb-release-gate.ps1
    op: delete
    scope: "Remove only if Go release gate fully replaces it."
  - path: scripts/powershell-helpers.ps1
    op: edit
    scope: "Remove or retain only if other scripts still need it."
  - path: .github/skills/kb-check/SKILL.md
    op: edit
    scope: "Point users to the Go-native gate."
  - path: README.md
    op: edit
    scope: "Update platform claims after PS1 removal."
  - path: config/skill-quality.json
    op: edit
    scope: "Update sync/lint references if PS1 paths are removed."
protected_oracles: []
status: done
---

# Slice 105: Remove PS1 Gate Dependency After Proof

## What To Build

Only after slice 104 records passing parity proof, remove or demote PS1 gate
entrypoints and update docs/sync targets. This is the cleanup slice, not the
place to discover parity gaps.

## Acceptance Criteria

- PS1 gate removal happens only after recorded parity proof.
- `go run .\cmd\kbcheck local-release` is the canonical release command.
- `kb-check -All` equivalent runs via Go.
- Skill sync report passes after propagated skill copies are updated.
- README no longer overstates PowerShell-first tooling.

## Test Scenarios

- `go test ./...`
- `go run .\cmd\kbcheck core`
- `go run .\cmd\kbcheck local-release`
- `scripts\skill-sync-report.ps1` only if still present; otherwise Go
  equivalent.
- `git diff --check`

## Scope Boundary

- Stop if parity proof is missing or failing.
- Do not delete unrelated PS1 helper scripts still used by other tooling.

## Completion Proof

- Removed `.github/skills/kb-check/scripts/kb-check.ps1` and
  `scripts/kb-release-gate.ps1` after the parity report passed.
- Retained `scripts/powershell-helpers.ps1` because existing validator scripts
  still source it.
- Synced `kb-check` to Codex, Copilot, shared agents, ATV `.github`, ATV
  scaffold, and ATV plugin skill roots.
- Removed untracked `media-*` skills from global roots and this working skill
  root; they are not part of the global bundle.
- `go run .\cmd\kbcheck local-release --json` exited 0.
