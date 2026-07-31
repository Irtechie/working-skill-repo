---
type: kb-manifest
kb_id: kb-2026-07-09-plan-to-pr-finish
brainstorm_path: direct-chat
created: 2026-07-09
status: superseded
superseded_by: docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md
workflow_shape: "skill-bundle-change"
gate_ledger:
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: superseded
    required_evidence:
      - "user requested one skill from plans to done-done and checked in"
      - "kb-complete applies configured delivery after successful internal phases"
      - "kb-ship is the explicit commit/push/PR boundary"
      - "two vertical slices have expected files and proof"
    proof:
      - docs/plans/2026-07-09-020-kb-plan-to-pr-finish-manifest.md
      - docs/plans/2026-07-09-021-kb-ship-check-in-plan.md
      - docs/plans/2026-07-09-022-kb-finish-orchestrator-plan.md
      - .github/skills/kb-complete/SKILL.md
    blockers: []
    passed_at: "2026-07-09T20:35:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md"
slices:
  - id: slice-001
    title: "Make kb-ship the explicit commit, push, and PR boundary"
    path: docs/plans/2026-07-09-021-kb-ship-check-in-plan.md
    blockers: []
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: medium
    hitl: false
    status: skipped
    owner: agent
    can_continue_other_slices: true
    protected_oracles: []
  - id: slice-002
    title: "Superseded by state-aware kb-complete orchestration"
    path: docs/plans/2026-07-09-022-kb-finish-orchestrator-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: medium
    hitl: false
    status: skipped
    owner: agent
    can_continue_other_slices: true
    protected_oracles: []
---

# KB Plan-to-PR Finish - Superseded

The legacy finish alias is removed. `kb-complete` now owns the uninterrupted
state-aware automation boundary:

```text
plan/manifest -> kb-work -> kb-finalize -> kb-complete -> kb-ship
              -> authorized kb-land
```

Delivery still follows configured policy or explicit run-scoped authorization.
`kb-work` never integrates default, `kb-ship` never merges, and only `kb-land`
may integrate remote default.
