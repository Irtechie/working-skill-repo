---
kb_id: kb-2026-05-30-codex-live-skill-eval-adapter
slice_id: slice-001
title: "Add Codex live skill eval adapter"
blockers: []
verification: verification-only
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-eval-run-codex.ps1
    op: create
    scope: "run route fixtures through codex exec in disposable worktrees and emit skill-eval result JSON"
  - path: evals/skill-eval/result.schema.json
    op: create
    scope: "JSON schema used by Codex output-schema and scorer documentation"
  - path: .github/skills/kb-check/scripts/kb-check.ps1
    op: edit
    scope: "add deterministic dry-run adapter check to canonical quality gate"
  - path: evals/skill-eval/README.md
    op: edit
    scope: "document dry-run and live Codex adapter usage"
  - path: docs/context/eval-map.md
    op: edit
    scope: "update current eval map from live adapter gap to Codex adapter implemented plus GHCP gap"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 001: Add Codex Live Skill Eval Adapter

## What Changed

Added a Codex adapter that can run route fixtures through `codex exec`, capture a
schema-shaped result, and score it with `scripts/skill-eval.ps1`.

## Acceptance Criteria

- Dry-run mode emits a scorer-compatible result and is included in `kb-check -All`.
- Live mode creates a disposable git worktree and runs `codex exec` in read-only
  mode with `--output-schema`.
- Result JSON is scored by `scripts/skill-eval.ps1`.
- Docs state that GHCP adapter and broader live corpus remain future work.

## Verification

- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval-run-codex.ps1 -FixtureId tiny-typo-fix`
- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval-run-codex.ps1 -FixtureId tiny-typo-fix -DryRun`
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`
- `git diff --check`
