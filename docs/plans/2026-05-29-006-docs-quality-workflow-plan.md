---
kb_id: kb-2026-05-29-cross-runtime-skill-quality
slice_id: slice-006
title: "Document canonical quality workflow"
blockers: [slice-004, slice-005]
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: AGENTS.md
    op: edit
    scope: "clarify skill-bundle memory exception and quality checks"
  - path: .github/copilot-instructions.md
    op: edit
    scope: "clarify GHCP-compatible quality entry points"
  - path: README.md
    op: edit
    scope: "document canonical quality workflow and cross-runtime support"
  - path: docs/context/PROJECT.md
    op: edit
    scope: "refresh project map with new quality workflow"
  - path: docs/context/operations/testing.md
    op: edit
    scope: "finalize canonical check commands and expected clean result"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 006: Document Canonical Quality Workflow

## What To Build

Update the human and agent-facing docs so both Codex and GHCP sessions know how to validate this skill repo.

## Acceptance Criteria

- `AGENTS.md` states that repo-local KB memory is allowed here only for maintaining this skill bundle.
- `.github/copilot-instructions.md` points GHCP sessions to the same quality commands without Codex-only assumptions.
- `README.md` documents the canonical quality workflow and cross-runtime contract at a visible level.
- `docs/context/PROJECT.md` and `docs/context/operations/testing.md` point fresh sessions to the same commands.
- Docs distinguish read-only sync reporting from actual propagation.

## Expected Files

- `AGENTS.md`
- `.github/copilot-instructions.md`
- `README.md`
- `docs/context/PROJECT.md`
- `docs/context/operations/testing.md`

## Test Scenarios

- Run `git diff --check`.
- Run the canonical quality command from the docs.
- Search docs for contradictory statements about memory, quality checks, or sync mutation.

## Scope Boundary

This slice documents the workflow after tooling exists. It does not add new scripts.

## Dependencies

- Requires slice-004 `kb-check` integration.
- Requires slice-005 sync drift report.
