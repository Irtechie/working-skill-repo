---
kb_id: kb-2026-05-30-live-cross-runtime-skill-eval-harness
slice_id: slice-004
title: "Expand deterministic trace rule scoring"
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-eval.ps1
    op: edit
    scope: "score required files/commands/tools and forbidden shortcuts from fixture/result metadata"
  - path: evals/skill-eval/selftest/*.json
    op: create
    scope: "add passing and failing self-tests for trace rule scoring"
  - path: evals/skill-eval/README.md
    op: edit
    scope: "document trace rule fields and scoring behavior"
  - path: evals/skill-eval/result.schema.json
    op: edit
    scope: "extend result/schema fields only if needed for trace rules"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 004: Expand Deterministic Trace Rule Scoring

## What To Build

Teach the scorer to verify required workflow reads/tools/commands and forbidden
shortcuts where fixtures or result metadata define them. This should catch cases
where an agent routes correctly but skips the evidence-gathering discipline the
skill requires.

## Acceptance Criteria

- Self-tests include at least one pass and one intentional fail for required
  file reads or commands.
- Self-tests include at least one intentional fail for a forbidden shortcut.
- Existing self-tests continue to pass.
- The new rule format is documented and deterministic.

## Verification

- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1`
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`
- `git diff --check`
