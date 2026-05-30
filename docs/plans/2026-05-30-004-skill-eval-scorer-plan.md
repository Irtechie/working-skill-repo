---
kb_id: kb-2026-05-30-skill-eval-scorer
slice_id: slice-001
title: "Add deterministic skill result scorer"
blockers: []
verification: verification-only
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-eval.ps1
    op: create
    scope: "score captured skill result JSON against route fixtures, trace evidence, and structured claim checks"
  - path: evals/skill-eval/README.md
    op: create
    scope: "document result schema and scorer usage"
  - path: evals/skill-eval/selftest/*.json
    op: create
    scope: "self-test pass and intentional failure cases for route, proof, and claim scoring"
  - path: .github/skills/kb-check/scripts/kb-check.ps1
    op: edit
    scope: "include skill-eval in the canonical skill repo quality gate"
  - path: config/skill-quality.json
    op: edit
    scope: "register skill eval self-test location and claim check types"
  - path: docs/context/eval-map.md
    op: edit
    scope: "update current eval map from planned gap to implemented deterministic scorer plus live-adapter gap"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 001: Add Deterministic Skill Result Scorer

## What Changed

Added a local skill-eval scorer outside the skills. It does not call Codex or
GHCP; it scores captured result JSON and self-tests the scorer with intentional
route, proof, and claim failures.

## Acceptance Criteria

- `scripts/skill-eval.ps1` exits 0 only when good self-test results pass and
  intentionally bad self-test results fail.
- `kb-check -All` runs `skill-eval`.
- Docs no longer say prompt/trace/claim evals are entirely unbuilt; they state
  that deterministic result scoring exists and live adapters remain future work.

## Verification

- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1`
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`
- `git diff --check`
