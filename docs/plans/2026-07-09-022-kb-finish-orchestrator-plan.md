---
kb_id: kb-2026-07-09-plan-to-pr-finish
slice_id: slice-002
title: "Add kb-finish plan-to-PR orchestration and routing"
blockers: [slice-001]
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: medium
hitl: false
expected_files:
  - path: .github/skills/kb-finish/SKILL.md
    op: create
    scope: "route plan/manifest state through kb-plan, kb-work, kb-complete, and kb-ship until PR delivery"
  - path: .github/skills/kb-start/SKILL.md
    op: edit
    scope: "route explicit done-done/checked-in requests to kb-finish"
  - path: config/skill-quality.json
    op: edit
    scope: "allow kb-finish route fixtures"
  - path: evals/route-complexity/finish-plan-flow.json
    op: create
    scope: "deterministically require kb-finish for plan-to-PR requests"
  - path: README.md
    op: edit
    scope: "document the one-command plan-to-PR lane"
  - path: docs/context/architecture/README.md
    op: edit
    scope: "index kb-finish as the explicit checked-in orchestration lane"
status: pending
owner: agent
can_continue_other_slices: true
protected_oracles: []
---

# Slice 002 - Plan-to-PR Orchestrator

## Acceptance Criteria

- `kb-finish` accepts a plan source or manifest.
- Unplanned input routes through `kb-plan`; unfinished manifests route through
  `kb-work`; completed manifests route through `kb-complete`.
- Shipping runs only after `complete-to-ship` passes.
- The lane ends only with a pushed branch and PR URL, or an honest blocker.
- Ordinary internal `kb-complete` calls remain non-shipping.

## Verification

Run route eval, skill lint, core, sync report, and `git diff --check`.
