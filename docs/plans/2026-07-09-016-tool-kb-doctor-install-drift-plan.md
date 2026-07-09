---
kb_id: kb-2026-07-09-phoenix-routing-slicing-absorption
slice_id: slice-006
title: "Add kb-doctor install drift repair"
blockers: [slice-001]
verification: integration
test_level: integration
functional_risk: none
model_tier: small
model_route: local-5090-coder
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck doctor-selftest"
  expect: 0
hitl: false
expected_files:
  - path: "cmd/kbcheck/"
    op: edit
    scope: "Add doctor command or subcommand that consumes existing skill sync drift logic and offers explicit repair."
  - path: "config/skill-quality.json"
    op: inspect
    scope: "Reuse configured required and optional targets; do not duplicate target lists in code."
  - path: "docs/context/operations/skill-bundle-maintenance.md"
    op: edit
    scope: "Document doctor report and repair workflow."
  - path: "README.md"
    op: edit
    scope: "List the new doctor command once implemented."
protected_oracles:
  - "cmd/kbcheck/skill_validators_test.go"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Implement a repair-capable doctor wrapper around skill-sync-report after slice 001 proves the gate is healthy."
human_action: ""
can_continue_other_slices: true
---

# Slice 006: Add Kb-Doctor Install Drift Repair

## What This Delivers

Phoenix's useful doctor idea is absorbed without adding Rust, MCP, or a daemon.
The existing read-only `skill-sync-report` remains the evidence source.
`doctor` adds a safe repair path for stale installed skill copies.

## Acceptance Criteria

- `go run ./cmd/kbcheck doctor` reports required and optional skill drift using
  the same hash logic as `skill-sync-report`.
- `go run ./cmd/kbcheck doctor --fix` repairs stale required targets from
  `<working-skill-repo>/.github/skills` only after a clean source hash is known.
- The fix path refuses or pauses on possible global-only useful drift. It must
  not silently overwrite a changed target whose source has not been reviewed.
- The command is file-native and optional; no CCE, MCP server, Rust binary, or
  background app is required.
- Docs explain when to use `skill-sync-report`, `doctor`, and `doctor --fix`.

## Test Scenarios

- Unit fixture: required target matches source, command exits green.
- Unit fixture: required target is stale, report exits nonzero without `--fix`.
- Unit fixture: stale required target is repaired with `--fix` and hashes match.
- Unit fixture: target has unknown/newer drift metadata, `--fix` refuses and
  emits the merge-back instruction.
- Integration: `go run ./cmd/kbcheck doctor --root <fixture>` works without any
  global user directories.

## Scope Boundary

Do not turn `doctor` into an installer, marketplace promoter, or upstream ATV
sync tool. It repairs configured installed copies after review; it does not
decide which skills should exist.
