---
type: kb-manifest
kb_id: kb-2026-07-21-adhd-output-profile
created: 2026-07-21
status: active
workflow_shape: skill-bundle-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "proves repo quality and required global skill copies are synchronized"
gate_ledger:
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence: ["manifest and slice plan exist", "slice has scope and proof"]
    proof:
      - docs/plans/2026-07-21-000-kb-adhd-output-profile-manifest.md
      - docs/plans/2026-07-21-001-skill-adhd-output-profile-plan.md
    blockers: []
    passed_at: "2026-07-21"
    allowed_next_action: "kb-work docs/plans/2026-07-21-000-kb-adhd-output-profile-manifest.md"
slices:
  - id: slice-001
    title: "Add compact ADHD-friendly response profile"
    path: docs/plans/2026-07-21-001-skill-adhd-output-profile-plan.md
    blockers: []
    verification: verification-only
    test_level: integration
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "Cross-install policy wording must preserve proof and safety exceptions."
    model_requirements: ["precise policy editing", "cross-install hash verification"]
    escalation_triggers: ["local-release fails outside known dirty-worktree scope", "installed copy contains newer useful drift"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:kb-compact", "file:AGENTS.md", "file:README.md"]
    shared_resources: ["git:integration-owner", "sync:global-skills"]
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck local-release"
      expect: 0
    hitl: false
    status: pending
    owner: agent
---

# KB: ADHD-Friendly Output Profile

Adapt AYGHRI's action-first response ideas into the existing compact-output
contract without adding an always-loaded skill or weakening evidence gates.

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Add compact ADHD-friendly response profile | - | verification-only | no | pending |
