---
kb_id: kb-2026-05-31-warning-quality-cleanup
slice_id: slice-002
title: "Codify local review fallback"
blockers: []
verification: integration
test_level: cli
functional_risk: none
hitl: false
expected_files:
  - path: .github/skills/kb-review/SKILL.md
    op: edit
    scope: "Define truthful local structured review fallback when reviewer subagents are unavailable."
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "Require review-mode recording and forbid claiming multi-agent review when fallback is used."
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Patch review fallback wording, then run lint/reference checks."
human_action: ""
can_continue_other_slices: true
---

# Codify Local Review Fallback

## What To Build

Make the review pipeline honest when reviewer subagents are unavailable or not
authorized by the runtime.

## Acceptance Criteria

- `kb-review` defines `review-mode: multi-agent` and `review-mode: local-fallback`.
- `kb-complete` records which review mode was used.
- The workflow does not claim a multi-agent review happened when only local
  review ran.

## Test Scenarios

- `powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1`
- `rg -n "local-fallback|review-mode" .github/skills/kb-review/SKILL.md .github/skills/kb-complete/SKILL.md`

## Scope Boundary

Do not change tool permissions or invent subagent access.
