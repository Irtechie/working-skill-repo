---
kb_id: kb-2026-07-20-harness-validation-recovery
slice_id: slice-003
title: "Make harness-engineering validation reproducible on Windows"
blockers: [slice-002]
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: medium
model_tier_reason: "The external failure mode is now known: Python 3.12 syntax plus CRLF-sensitive immutable hashes, both fixable with a bounded runner."
model_requirements: ["PowerShell-safe path handling", "Git clone config", "Python launcher discovery", "clear uv fallback semantics"]
escalation_triggers: ["Python 3.12 is unavailable", "LF-stable clone still fails manifest validation", "runner would mutate the external harness repo", "network clone fails after retries"]
context_packet_path: docs/plans/2026-07-20-harness-validation-recovery-context/slice-003.json
proof_check:
  kind: command_exit
  command: "powershell -NoProfile -ExecutionPolicy Bypass -File scripts/harness-engineering-review.ps1 -HarnessOnly -HarnessRoot E:/Data/Codex/tmp/harness-engineering-lf-20260720"
  expect: 0
hitl: false
expected_files:
  - path: scripts/harness-engineering-review.ps1
    op: create
    scope: "Clone or validate lopopolo/harness-engineering with core.autocrlf=false and Python 3.12."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document the harness-engineering runner and LF/Python requirements."
  - path: docs/results/2026-07-20-harness-engineering-recovery.md
    op: create
    scope: "Record the observed CRLF failure and LF-stable passing commands."
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Create a repo-owned runner that requires or discovers Python 3.12, clones with core.autocrlf=false when needed, uses uv only when available, and records the exact harness commit."
human_action: ""
can_continue_other_slices: false
---

# Slice 003: Harness-Engineering Runner

## What To Build

Add a repo-owned runner for applying `lopopolo/harness-engineering` validation
without depending on accidental machine defaults.

## Acceptance Criteria

- Existing harness checkout is validated without mutation when supplied.
- Missing harness checkout can be cloned with `git -c core.autocrlf=false`.
- Python 3.12 is required or discovered explicitly.
- If `uv` is present, the documented `uv run --script` path may be used.
- If `uv` is absent, the runner uses the verified `py -3.12` fallback.
- The runner records the harness commit and separates harness setup failure
  from target repo failure.

## Test Scenarios

- LF-stable harness checkout passes `validate_manifest.py`.
- LF-stable harness checkout passes `test_manifest.py`.
- CRLF-mismatched checkout fails with a message pointing to LF-stable clone
  remediation.
- Python 3.11 is rejected before script execution.

## Proof Check

`powershell -NoProfile -ExecutionPolicy Bypass -File scripts/harness-engineering-review.ps1 -HarnessOnly -HarnessRoot E:/Data/Codex/tmp/harness-engineering-lf-20260720`

## Scope Boundary

Do not patch the external harness repository or vendor its corpus into this
repo. This slice owns our invocation boundary only.

## Dependencies

Depends on slice 002 so runner failures can use the same diagnostic style as
native proof failures.
