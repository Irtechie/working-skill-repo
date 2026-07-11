---
type: kb-manifest
kb_id: kb-2026-07-09-plan-to-pr-finish
brainstorm_path: direct-chat
created: 2026-07-09
status: active
workflow_shape: "skill-bundle-change"
gate_ledger:
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "user requested one skill from plans to done-done and checked in"
      - "kb-complete must remain non-shipping when called internally"
      - "kb-ship is the explicit commit/push/PR boundary"
      - "two vertical slices have expected files and proof"
    proof:
      - docs/plans/2026-07-09-020-kb-plan-to-pr-finish-manifest.md
      - docs/plans/2026-07-09-021-kb-ship-check-in-plan.md
      - docs/plans/2026-07-09-022-kb-finish-orchestrator-plan.md
      - .github/skills/kb-complete/SKILL.md
    blockers: []
    passed_at: "2026-07-09T20:35:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-09-020-kb-plan-to-pr-finish-manifest.md"
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
    status: pending
    owner: agent
    can_continue_other_slices: true
    protected_oracles: []
  - id: slice-002
    title: "Add kb-finish plan-to-PR orchestration and routing"
    path: docs/plans/2026-07-09-022-kb-finish-orchestrator-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: medium
    hitl: false
    status: pending
    owner: agent
    can_continue_other_slices: true
    protected_oracles: []
---

# KB Plan-to-PR Finish

`kb-finish <manifest-or-plan>` is the explicit automation boundary:

```text
plan/manifest -> kb-work -> kb-complete -> kb-ship -> commit -> push -> PR
```

Ordinary `kb-work -> kb-complete` does not push unless the caller chose
`kb-finish`.
