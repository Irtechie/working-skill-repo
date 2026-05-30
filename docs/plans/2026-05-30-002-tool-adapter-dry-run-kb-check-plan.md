---
kb_id: kb-2026-05-30-live-cross-runtime-skill-eval-harness
slice_id: slice-002
title: "Wire adapter dry-runs into canonical checks"
blockers: [slice-001]
verification: functional
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-check/scripts/kb-check.ps1
    op: edit
    scope: "include GHCP adapter dry-run in the default deterministic quality gate"
  - path: config/skill-quality.json
    op: edit
    scope: "mark live adapter support states for Codex and GHCP"
  - path: docs/context/operations/testing.md
    op: edit
    scope: "document default dry-run checks versus explicit live model commands"
  - path: README.md
    op: edit
    scope: "summarize cross-runtime adapter dry-runs in visible quality workflow"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 002: Wire Adapter Dry-Runs Into Canonical Checks

## What To Build

Extend the default quality gate so adapter plumbing is checked without model
calls. Codex and GHCP dry-runs should both run under `kb-check -All`; live model
runs remain explicit commands.

## Acceptance Criteria

- `kb-check -List` shows Codex and GHCP adapter dry-run checks when scripts
  exist.
- `kb-check -All` runs both dry-runs and exits nonzero if either adapter plumbing
  breaks.
- Docs clearly separate deterministic dry-run coverage from explicit live evals.
- Required skill copies are synced after changing `kb-check`.

## Verification

- `.\.github\skills\kb-check\scripts\kb-check.ps1 -List`
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`
- `powershell -ExecutionPolicy Bypass -File scripts\skill-sync-report.ps1`
- `git diff --check`
