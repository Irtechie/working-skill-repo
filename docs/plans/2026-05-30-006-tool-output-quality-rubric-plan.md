---
kb_id: kb-2026-05-30-live-cross-runtime-skill-eval-harness
slice_id: slice-006
title: "Add output quality rubric scorer"
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-eval-quality.ps1
    op: create
    scope: "score captured results for completeness, maintainability, relevance, proof quality, and ceremony"
  - path: evals/skill-eval/quality/*.json
    op: create
    scope: "rubric self-tests for good, incomplete, unmaintainable, irrelevant, and over-ceremonial outputs"
  - path: evals/skill-eval/README.md
    op: edit
    scope: "document rubric dimensions, thresholds, and deterministic versus judged fields"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 006: Add Output Quality Rubric Scorer

## What To Build

Add a separate quality scoring layer for output completeness, maintainability,
relevance, proof quality, and unnecessary ceremony. It must not replace route,
proof, trace, or claim pass/fail.

## Acceptance Criteria

- Quality scorer can flag incomplete proof, irrelevant output, unmaintainable
  output, and over-ceremonial tiny-task handling.
- Rubric output labels each dimension as deterministic, LLM-judged, or
  human-only.
- Default deterministic `skill-eval.ps1` behavior remains stable unless quality
  scoring is explicitly requested or documented as part of a later gate.

## Verification

- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval-quality.ps1`
- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1`
- `git diff --check`
