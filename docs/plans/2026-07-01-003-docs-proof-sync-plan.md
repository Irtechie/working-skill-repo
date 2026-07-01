---
kb_id: kb-2026-07-01-live-steering-learning-loop
slice_id: slice-003
title: "Refresh docs and proof surface"
blockers: [slice-002]
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: "README.md"
    op: edit
    scope: "Mention live steering as part of the visible long-goal and learning workflow."
  - path: "docs/context/architecture/kb-workflow.md"
    op: edit
    scope: "Document the control-loop and steering-memory contract in the workflow architecture."
  - path: "docs/context/goals/live-steering-learning-loop.md"
    op: edit
    scope: "Record artifacts, slice status, and proof."
  - path: "todo.md"
    op: edit
    scope: "Keep active board pointer current."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Update visible docs and run deterministic proof."
human_action: ""
can_continue_other_slices: true
---

# Slice 003 - Docs And Proof Surface

## What To Build

Refresh user-visible and project-memory docs so future agents can discover the
new live steering contract without rereading this chat.

## Acceptance Criteria

- README names live steering for long goals and does not imply a scheduled CI
  runner was added.
- Architecture docs explain how live steering complements, not replaces,
  `kb-complete`, `learn`, and `evolve`.
- Goal ledger and todo reflect the current artifact and proof.
- Deterministic checks are run and failures are either fixed or recorded with
  exact blockers.

## Verification

- `go run ./cmd/kbcheck core`
- `git diff --check`

## Scope Boundary

No install propagation or ATV/global sync unless explicitly requested after the
repo-local change passes.
