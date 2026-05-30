---
kb_id: kb-2026-05-30-eval-map
slice_id: slice-002
title: "Wire eval mapping into bootstrap and docs"
blockers: [slice-001]
verification: verification-only
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-map-bootstrap/SKILL.md
    op: edit
    scope: "invoke kb-eval-map during bootstrap after repo inventory and before final testing docs"
  - path: docs/context/architecture/README.md
    op: edit
    scope: "add kb-eval-map to verification/setup lane"
  - path: docs/context/operations/testing.md
    op: edit
    scope: "document eval-map as bootstrap-owned eval setup"
  - path: README.md
    op: edit
    scope: "add visible workflow entry for kb-eval-map if the skill list or bootstrap docs mention verification setup"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 002: Wire Eval Mapping Into Bootstrap And Docs

## What To Build

Make bootstrap call `kb-eval-map` as part of the expensive project setup pass and
update docs so fresh sessions know where eval setup lives.

## Acceptance Criteria

- `kb-map-bootstrap` names `docs/context/eval-map.md` as a standard bootstrap
  output.
- Bootstrap invokes `kb-eval-map` after repo inventory and before final
  operations/testing docs.
- Architecture/testing docs describe `kb-eval-map` as setup intelligence, not as
  the runtime verifier.
- Existing `kb-check`, `kb-functional-test`, `kb-qa`, and `kb-complete` ownership
  remains intact.

## Verification

- Run `scripts/skill-lint.ps1`.
- Run `git diff --check`.

Result: docs and bootstrap wiring updated; final deterministic gate runs in
slice-003.
