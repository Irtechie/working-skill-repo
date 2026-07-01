---
kb_id: kb-2026-07-01-live-steering-learning-loop
slice_id: slice-002
title: "Add durable feedback classification and steering memory"
blockers: [slice-001]
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: ".github/skills/kb-complete/SKILL.md"
    op: edit
    scope: "Feed durable review/iteration feedback into steering memory and observations before learn/evolve."
  - path: ".github/skills/learn/SKILL.md"
    op: edit
    scope: "Classify feedback into current-only, steering memory, observation, landmine, or instinct evidence."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Edit kb-complete and learn with feedback classification and steering-memory rules."
human_action: ""
can_continue_other_slices: true
---

# Slice 002 - Feedback Classification And Steering Memory

## What To Build

Teach the KB learning loop to use in-flight feedback without weakening the
post-work completion pipeline.

## Acceptance Criteria

- `kb-complete` has a steering-feedback step before `learn` that records durable
  feedback in a small steering memory file when a manifest or goal names one.
- `learn` classifies feedback and refuses to turn one-off PR comments into
  durable instincts or memory.
- Steering memory is explicitly concise, human-readable, and separate from raw
  logs/transcripts.
- Landmine criteria remain strict and evidence-based.

## Verification

- `go run ./cmd/kbcheck core`
- `git diff --check`

## Scope Boundary

No generated `learned-*` skill promotion and no global sync in this slice.
