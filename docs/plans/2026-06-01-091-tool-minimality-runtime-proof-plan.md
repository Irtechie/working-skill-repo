---
kb_id: kb-2026-06-01-cold-storage-follow-through
slice_id: slice-091
title: "Prove minimality candidates before deletion"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-surface-minimality.ps1
    op: edit
    scope: "Add evidence fields or a mode that separates static, example-only, dispatch, and runtime-use evidence for cold-storage candidates."
  - path: scripts/skill-surface-minimality-selftest.ps1
    op: edit
    scope: "Prove protected skills stay protected and cold-storage candidates carry evidence classifications."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document the deletion-proof gate and what evidence is insufficient."
  - path: todo.md
    op: edit
    scope: "Update remaining cold-storage candidates after proof report changes."
protected_oracles:
  - path: scripts/skill-surface-minimality-selftest.ps1
    role: "minimality classification oracle"
    sha256: "filled by kb-work before edits"
    update_policy: "requires explicit plan update"
status: done
---

# Slice 091: Prove Minimality Candidates Before Deletion

## What To Build

Extend the minimality report so a deletion candidate carries explicit evidence
quality instead of just "no static inbound reference." The output should make it
obvious whether a candidate has real dispatch/runtime evidence, static examples
only, docs-only mentions, or no evidence.

## Acceptance Criteria

- Protected skills cannot become deletion candidates.
- Cold-storage candidates include evidence-class fields that distinguish:
  `runtime`, `dispatch-static`, `example-only`, `docs-only`, and `none`.
- The report still refuses to approve deletion by itself.
- Selftest proves protected and evidence-class behavior.
- `kb-check -All` includes the selftest through the existing check list.

## Test Scenarios

- Run `pwsh -NoProfile -File scripts/skill-surface-minimality-selftest.ps1`.
- Run `pwsh -NoProfile -File scripts/skill-surface-minimality.ps1 -Json`.
- Run `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`.

## Scope Boundary

- Do not delete agents or skills.
- Do not count example prompts inside an agent file as real runtime use.
- Do not claim true runtime usage unless the evidence is externally captured or
  comes from a durable invocation log.

## Completion Proof

- `scripts/skill-surface-minimality-selftest.ps1` exited 0.
- `scripts/skill-surface-minimality.ps1 -Json` exited 0.
- The report now emits `evidence_class` and `evidence_sources`; output remains a
  candidate report, not deletion approval.
