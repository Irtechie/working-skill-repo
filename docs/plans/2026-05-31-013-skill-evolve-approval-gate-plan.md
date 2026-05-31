---
kb_id: kb-2026-05-31-learning-landmines
slice_id: slice-003
title: "Gate evolve promotion and sync"
blockers: [slice-001]
verification: integration
test_level: cli
functional_risk: narrow
hitl: true
expected_files:
  - path: .github/skills/evolve/SKILL.md
    op: edit
    scope: "Require explicit approval before learned-* promotion/sync from this bundle."
  - path: scripts/skill-sync-report.ps1
    op: edit
    scope: "Optionally flag generated learned-* sync attempts."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document generated-skill approval proof."
status: pending
owner: agent
blocked_reason: ""
resume_when: "Human approves exact approval prompt or accepts the default."
next_agent_action: "Add a hard approval step after numeric gates pass."
human_action: "Approve wording for generated learned-* skill promotion."
can_continue_other_slices: true
test_inputs:
  - name: approval_prompt
    source: user
    required_for: "Human gate wording"
    value: "Default: 'Promote these generated learned-* skills and allow sync from this portable bundle? yes/no'"
---

# Gate Evolve Promotion And Sync

## What To Build

Keep `evolve` numeric maturity gates but add an explicit human approval gate
before generated `learned-*` skills are committed or synced from this portable
bundle.

## Acceptance Criteria

- Candidates still require confidence greater than 0.85, more than five
  observations, and last seen within 90 days.
- Generated skills are drafts until approved.
- Sync/propagation cannot silently include generated skills.
- The final report distinguishes generated drafts from approved skills.

## Test Scenarios

- Use a small synthetic instinct file to show a candidate reaches the approval
  prompt.
- Confirm declining approval creates no promoted/synced skill.
- Run `kb-check -All`.

## Scope Boundary

Do not change the scoring model unless tests prove it is wrong.
