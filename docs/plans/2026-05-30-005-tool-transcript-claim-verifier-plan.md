---
kb_id: kb-2026-05-30-live-cross-runtime-skill-eval-harness
slice_id: slice-005
title: "Add transcript-derived claim verifier"
blockers: [slice-004]
verification: tdd
test_level: functional-cli
functional_risk: broad
hitl: false
expected_files:
  - path: scripts/skill-eval-claims.ps1
    op: create
    scope: "extract/check claims from final responses or transcripts against filesystem/git/log evidence"
  - path: scripts/skill-eval.ps1
    op: edit
    scope: "invoke transcript-derived claim checks when claim artifacts are present"
  - path: evals/skill-eval/claims/*
    op: create
    scope: "self-test true, false, and ambiguous transcript claim cases"
  - path: evals/skill-eval/README.md
    op: edit
    scope: "document deterministic claim verification contract"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 005: Add Transcript-Derived Claim Verifier

## What To Build

Add a verifier that checks final-answer or transcript-derived claims against
actual files, git state, commands, logs, and artifacts. Candidate extraction may
start simple, but pass/fail must come from deterministic checks.

## Acceptance Criteria

- Self-tests include one true claim, one false claim, and one ambiguous claim.
- False claims fail deterministically.
- Ambiguous claims are reported without becoming proof.
- Structured `claim_checks` remain supported and stable.

## Verification

- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval-claims.ps1`
- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1`
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`
- `git diff --check`
