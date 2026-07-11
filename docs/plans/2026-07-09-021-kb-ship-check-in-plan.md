---
kb_id: kb-2026-07-09-plan-to-pr-finish
slice_id: slice-001
title: "Make kb-ship the explicit commit, push, and PR boundary"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: medium
hitl: false
expected_files:
  - path: .github/skills/kb-ship/SKILL.md
    op: edit
    scope: "require complete-to-ship, deliberate staging, commit, push, and PR creation/update without merge"
  - path: README.md
    op: edit
    scope: "document kb-ship as checked-in PR delivery"
status: pending
owner: agent
can_continue_other_slices: true
protected_oracles: []
---

# Slice 001 - Explicit Check-In Boundary

## Acceptance Criteria

- Invoking `kb-ship` explicitly authorizes commit, push, and PR creation/update.
- Shipping never merges, force-pushes, bypasses hooks, or stages unrelated files.
- Default-branch work moves to a topic branch before commit.
- Success requires pushed state and a PR URL, unless there is genuinely nothing
  to ship.

## Verification

Run `go run ./cmd/kbcheck core` and route fixture validation.
