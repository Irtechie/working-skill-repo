---
kb_id: kb-2026-07-20-harness-validation-recovery
slice_id: slice-004
title: "Rerun combined release and harness review proof"
blockers: [slice-003]
verification: functional
test_level: functional-cli
functional_risk: broad
model_tier: medium
model_tier_reason: "This is final integration proof across the repaired Go gate, bounded diagnostics, and external harness runner."
model_requirements: ["release-gate interpretation", "external harness evidence capture", "project memory refresh", "dirty worktree scope accounting"]
escalation_triggers: ["local-release fails for unrelated active work", "global skill drift requires merge before sync", "harness runner passes but repository-review claim lacks target-boundary evidence", "proof output cannot distinguish target failure from harness setup failure"]
context_packet_path: docs/plans/2026-07-20-harness-validation-recovery-context/slice-004.json
proof_check:
  kind: command_exit
  command: "powershell -NoProfile -ExecutionPolicy Bypass -File scripts/harness-engineering-review.ps1 -TargetRoot E:/working-skill-repo -HarnessRoot E:/Data/Codex/tmp/harness-engineering-lf-20260720"
  expect: 0
hitl: false
expected_files:
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Update current proof commands and harness-engineering test route."
  - path: docs/context/PROJECT.md
    op: edit
    scope: "Refresh route map only if the proof workflow or commands changed."
  - path: todo.md
    op: edit
    scope: "Update this KB workflow status and any remaining blockers."
  - path: docs/results/2026-07-20-harness-engineering-recovery.md
    op: edit
    scope: "Append final combined proof outcome."
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Run the combined proof, update testing docs/results with exact commands and outcomes, and leave the manifest/todo state honest."
human_action: ""
can_continue_other_slices: false
---

# Slice 004: Combined Release And Harness Review Proof

## What To Build

Complete the recovery by running the repaired native skill-repo proof and the
external harness validation through one reproducible command.

## Acceptance Criteria

- `git diff --check` passes.
- `go run ./cmd/kbcheck core --list` returns normally.
- `go run ./cmd/kbcheck core` returns normally or fails with bounded diagnostics.
- `go test ./...` returns normally or fails with bounded diagnostics.
- Harness corpus validation passes against an LF-stable checkout under Python
  3.12.
- The final report distinguishes target proof failures from harness setup
  failures.
- Project memory is refreshed if the canonical proof workflow changed.

## Test Scenarios

- Combined runner with valid target and LF-stable harness exits 0.
- Combined runner with CRLF harness exits nonzero before target proof, with
  remediation.
- Combined runner with a forced target failure labels the target boundary.

## Proof Check

`powershell -NoProfile -ExecutionPolicy Bypass -File scripts/harness-engineering-review.ps1 -TargetRoot E:/working-skill-repo -HarnessRoot E:/Data/Codex/tmp/harness-engineering-lf-20260720`

## Scope Boundary

Do not sync global skills unless this workflow changes skill files and the
release/sync gate requires it. Do not commit or push unless separately asked.

## Dependencies

Depends on slice 003.
