---
kb_id: kb-2026-06-10-skill-bundle-cleanup
slice_id: slice-001
title: "Reproduce and classify audit findings"
blockers: []
verification: verification-only
test_level: functional-cli
functional_risk: broad
hitl: false
expected_files:
  - path: docs/context/research/2026-06-10-skill-bundle-cleanup-audit-refresh.md
    op: create
    scope: "Record each pasted audit claim as confirmed, stale, already fixed, or needs-human with evidence."
  - path: docs/context/memory-maintenance.md
    op: edit
    scope: "Add only confirmed durable maintenance signals that remain after the audit refresh."
  - path: todo.md
    op: edit
    scope: "Reflect any confirmed blockers or retired stale claims that change execution status."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Reproduce every audit claim cheaply before allowing downstream edits."
human_action: ""
can_continue_other_slices: false
---

# Slice 001: Reproduce and classify audit findings

## What To Build

Create a current evidence table for the pasted audit. The result must separate
confirmed issues from stale claims before implementation touches code, skills,
or docs.

## Acceptance Criteria

- Every audit finding has a row with status: `confirmed`, `stale`,
  `already-fixed`, `not-reproducible`, or `needs-human`.
- Each row includes exact command, file path, or repo observation evidence.
- Confirmed durable issues are added to `docs/context/memory-maintenance.md`.
- Downstream slices are scoped to confirmed issues only.
- Claims contradicted by current memory, such as missing handoff directories, are
  retired instead of planned as work.

## Expected Files

- `docs/context/research/2026-06-10-skill-bundle-cleanup-audit-refresh.md`
- `docs/context/memory-maintenance.md`
- `todo.md`

## Test Scenarios

- Run exact-path checks for `todo.md`, `docs/context/PROJECT.md`, and
  `docs/handoffs/{active,parked,done}`.
- Run targeted checks for Go version, install profiles, selftest output wording,
  hook references, `.ps1` files, process-doc volume, and review-skill reference
  duplication.
- Where feasible, reproduce fresh-clone behavior in a temporary clone or a
  missing-target fixture without mutating global install directories.

## Scope Boundary

Do not fix code or delete files in this slice. Evidence only.

## Dependencies

None.
