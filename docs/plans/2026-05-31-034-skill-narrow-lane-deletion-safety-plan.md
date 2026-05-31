---
kb_id: kb-2026-05-31-lazy-lane-consolidation
slice_id: slice-004
title: "Audit narrow lanes for deletion safety"
blockers: [slice-003]
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: evals/route-complexity/
    op: edit
    scope: "Add or update fixtures for retained narrow lanes."
  - path: .github/skills/kb-fix/SKILL.md
    op: edit
    scope: "Trim duplicated proof prose if safe."
  - path: .github/skills/kb-troubleshoot/SKILL.md
    op: edit
    scope: "Trim duplicated proof prose if safe."
  - path: .github/skills/kb-functional-test/SKILL.md
    op: edit
    scope: "Decide standalone vs lazy reference."
  - path: .github/skills/kb-regression-snapshot/SKILL.md
    op: edit
    scope: "Decide standalone vs eval-baseline integration."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Audit narrow lanes with reference and route proof before any deletion."
human_action: ""
can_continue_other_slices: true
---

# Audit Narrow Lanes For Deletion Safety

## What To Build

Audit narrow KB lanes and keep only distinct failure-mode/proof value. Merge or
delete only after route fixtures and reference scans are clean.

## Acceptance Criteria

- Each retained narrow lane has a distinct trigger and proof value.
- `kb-regression-snapshot` is reconciled with persisted eval baselines.
- `kb-functional-test` is either retained as a lazy lane or moved into
  `kb-work`/`kb-check` references.
- No deleted/merged skill has dangling invocation references.

## Test Scenarios

- Run route-complexity eval.
- Run skill reference scanner.
- Run `kb-check -All`.

## Scope Boundary

Do not delete `kb-review`, `ce-review`, `ce-compound`, or
`ce-compound-refresh` unless callers are rewritten first.
