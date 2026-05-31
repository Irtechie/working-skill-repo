---
kb_id: kb-2026-05-31-proof-pipeline-spike
slice_id: slice-003
title: "Add protected oracle pipeline contract"
blockers: [slice-002]
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Represent protected test-first oracle fields in manifests or slice plans."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Require work to preserve protected oracle files once declared."
  - path: .github/skills/kb-check/SKILL.md
    op: edit
    scope: "Point to deterministic protected-oracle verification."
  - path: .github/skills/kb-check/scripts/kb-check.ps1
    op: edit
    scope: "Run any new protected-oracle selftest."
  - path: scripts/skill-eval-manifest-selftest.ps1
    op: edit
    scope: "Reuse or extend protected-file SHA selftest if appropriate."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Add the anti-cheat oracle contract to pipeline artifacts."
human_action: ""
can_continue_other_slices: true
---

# Add Protected Oracle Pipeline Contract

## What To Build

Move the valuable TDD behavior into the KB pipeline: define the behavior oracle
before implementation when practical, prove RED when possible, protect the
oracle with SHA/manifest, then implement and rerun the same oracle.

## Acceptance Criteria

- A slice can declare protected oracle files.
- Work instructions forbid mutating protected oracle files after the protected
  manifest is created unless the plan explicitly updates the oracle.
- Verification can detect protected oracle tampering.
- Existing eval manifest SHA selftest remains green.

## Test Scenarios

- Add or reuse a selftest that protects a fixture/test file, mutates it, and
  fails verification.
- Run `git diff --check`.
- Run `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`.

## Scope Boundary

This does not require keeping standalone `tdd` as a globally loaded skill. It
only preserves the anti-cheat mechanism.
