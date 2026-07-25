---
kb_id: kb-2026-07-20-harness-validation-recovery
slice_id: slice-002
title: "Fail stalled proof commands with useful diagnostics"
blockers: [slice-001]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: medium
model_tier_reason: "The core behavior is bounded command execution and actionable timeout reporting after the package-load surface is available."
model_requirements: ["Go command runner tests", "Windows process-tree cleanup", "release-gate output design", "backward-compatible CLI flags"]
escalation_triggers: ["timeout leaves child processes running", "core/local-release output truncates the failing command", "diagnostics require a daemon or persistent monitor", "existing release tests become flaky"]
context_packet_path: docs/plans/2026-07-20-harness-validation-recovery-context/slice-002.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'Timeout|Release|Core'"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Expose bounded timeout defaults or flags for native proof commands if needed."
  - path: cmd/kbcheck/release.go
    op: edit
    scope: "Ensure local-release/core failures preserve command, status, timeout, and useful diagnostics."
  - path: cmd/kbcheck/main_test.go
    op: edit
    scope: "Add or update CLI timeout/diagnostic assertions."
  - path: cmd/kbcheck/release_test.go
    op: edit
    scope: "Add release-output assertions for timed-out required checks."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document the bounded failure behavior and next probe sequence."
protected_oracles:
  - path: cmd/kbcheck/main_test.go
    role: "CLI timeout diagnostic oracle"
    sha256: "filled by kb-work after oracle protection"
    update_policy: "requires explicit plan amendment"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Add or tighten timeout diagnostics around native proof commands so future hangs fail closed with command, duration, and next probe."
human_action: ""
can_continue_other_slices: false
---

# Slice 002: Bounded Proof Diagnostics

## What To Build

Make stalled native proof commands fail closed with actionable diagnostics. The
user should never see only "timed out after 90s with no completion" without a
command, boundary, and next probe.

## Acceptance Criteria

- A timed-out check reports command, configured timeout, status, and the next
  recommended probe.
- Process cleanup is bounded and does not leave known child processes running.
- `core --list` remains cheap and does not execute checks.
- `local-release` surfaces required timeout failures without hiding other
  check results.
- Existing release JSON remains parseable.

## Test Scenarios

- A helper command that sleeps past timeout returns exit 124 with diagnostics.
- A release check timeout marks the required check failed.
- `core --list` prints discovered checks without invoking the runner.

## Proof Check

`go test ./cmd/kbcheck -run 'Timeout|Release|Core'`

## Scope Boundary

This slice does not fix harness-engineering invocation or global sync drift.

## Dependencies

Depends on slice 001 because tests cannot be trusted while the package-load
surface itself hangs.
