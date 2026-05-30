---
kb_id: kb-2026-05-30-live-cross-runtime-skill-eval-harness
slice_id: slice-003
title: "Add live cross-runtime corpus runner"
blockers: [slice-001]
verification: functional
test_level: functional-cli
functional_risk: broad
hitl: false
expected_files:
  - path: scripts/skill-eval-run-live-corpus.ps1
    op: create
    scope: "run selected route fixtures across Codex and GHCP adapters and summarize outcomes"
  - path: evals/skill-eval/README.md
    op: edit
    scope: "document corpus runner usage and result categories"
  - path: docs/context/eval-map.md
    op: edit
    scope: "update live corpus coverage from one Codex fixture to cross-runtime command"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 003: Add Live Cross-Runtime Corpus Runner

## What To Build

Create an explicit runner that can execute all route fixtures across Codex and
GHCP adapters and emit a machine-readable summary. It must distinguish pass,
adapter failure, invalid JSON, deterministic score failure, runtime unavailable,
and skipped-by-platform states.

## Acceptance Criteria

- The runner supports all current route fixtures and both runtimes.
- The summary includes runtime, fixture id, expected route, actual route,
  pass/fail/skip category, duration, exit code, and artifact paths.
- Live calls are never run by default `kb-check -All`.
- Failed runs leave enough logs to debug without re-running.

## Verification

- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval-run-live-corpus.ps1 -FixtureId tiny-typo-fix -Runtime codex -DryRun`
- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval-run-live-corpus.ps1 -All -Runtime codex,ghcp -DryRun`
- One explicit live Codex smoke run if auth is available.
- `git diff --check`
