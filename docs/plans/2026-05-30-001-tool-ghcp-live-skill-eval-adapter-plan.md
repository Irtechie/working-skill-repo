---
kb_id: kb-2026-05-30-live-cross-runtime-skill-eval-harness
slice_id: slice-001
title: "Add GHCP live skill eval adapter"
blockers: []
verification: functional
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-eval-run-ghcp.ps1
    op: create
    scope: "run route fixtures through GitHub Copilot CLI and emit skill-eval result JSON"
  - path: evals/skill-eval/README.md
    op: edit
    scope: "document GHCP adapter dry-run/live usage and schema limitations"
  - path: docs/context/eval-map.md
    op: edit
    scope: "update GHCP adapter gap to implemented or runtime-auth-aware state"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 001: Add GHCP Live Skill Eval Adapter

## What To Build

Create a GHCP adapter mirroring the Codex adapter shape. It runs a route fixture
through `copilot -p`, captures stdout/stderr and transcript output where
available, extracts strict result JSON, writes `result.json`, and scores it with
`scripts/skill-eval.ps1`.

The adapter must fail hard on malformed JSON. If Copilot CLI auth/runtime is not
available, it must report an explicit unavailable state rather than pretending
the live eval passed.

## Acceptance Criteria

- `scripts/skill-eval-run-ghcp.ps1 -FixtureId tiny-typo-fix -DryRun` emits a
  scorer-compatible result.
- Live mode uses non-interactive Copilot CLI with no user questions and minimal
  permissions.
- Invalid or missing result JSON exits nonzero and leaves logs in the run
  directory.
- Runtime/auth unavailable is distinguishable from route/scoring failure.
- Docs state GHCP uses prompt-level JSON constraints because no local
  `--output-schema` flag is available.

## Verification

- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval-run-ghcp.ps1 -FixtureId tiny-typo-fix -DryRun`
- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval-run-ghcp.ps1 -FixtureId tiny-typo-fix`
- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1`
- `git diff --check`
