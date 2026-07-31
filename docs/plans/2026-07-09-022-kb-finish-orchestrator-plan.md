---
kb_id: kb-2026-07-09-plan-to-pr-finish
slice_id: slice-002
title: "Superseded by state-aware kb-complete orchestration"
blockers: [slice-001]
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: medium
hitl: false
expected_files: []
status: skipped
owner: agent
can_continue_other_slices: true
protected_oracles: []
superseded_by: docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md
---

# Slice 002 - Superseded Legacy Finish Alias

This slice is superseded by
`docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md`.
`kb-complete` is the single user-facing orchestrator and successful internal
phases return to it automatically.

The replacement flow is:

```text
kb-work -> kb-finalize -> kb-complete -> kb-ship -> authorized kb-land
```

The replacement preserves the original shipping boundary: `kb-ship` commits,
pushes, and opens or updates a PR but never merges. Only `kb-land` may integrate
the resolved remote default branch.
