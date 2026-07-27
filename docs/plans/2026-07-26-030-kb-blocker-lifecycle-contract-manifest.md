---
type: kb-manifest
kb_id: kb-2026-07-26-blocker-lifecycle-contract
created: 2026-07-26
status: completed
workflow_shape: pipeline-change
objective_contract: true
blocker_lifecycle_contract: true
done_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'TestBlockerLifecycle|TestManifestContract' -count=1"
  expect: 0
  why: "proves blocker ownership, pause state, stale rechecks, and scoped propagation are enforced"
gate_ledger:
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "manifest and both slice plans exist"
      - "blocker lifecycle failures from recent Agent127, release, visual, and proof runs are represented"
      - "existing low-cognitive-burden WIP is preserved"
    proof:
      - docs/plans/2026-07-26-030-kb-blocker-lifecycle-contract-manifest.md
      - docs/plans/2026-07-26-031-tool-blocker-lifecycle-contract-plan.md
      - docs/plans/2026-07-26-032-skill-blocker-ownership-propagation-plan.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-030-kb-blocker-lifecycle-contract-manifest.md"
  - gate_id: slice-slice-031-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Go lifecycle validator tests pass"
      - "portable Python parity tests pass"
    proof:
      - cmd/kbcheck/manifest_contract_test.go
      - .github/skills/kb-gate/scripts/check_gate_ledger_test.py
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work slice-032"
  - gate_id: slice-slice-032-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "communication contract test passes"
      - "fresh scenario review confirms narrow ownership and propagation"
    proof:
      - cmd/kbcheck/communication_contract_test.go
      - docs/plans/2026-07-26-032-skill-blocker-ownership-propagation-plan.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-finalize docs/plans/2026-07-26-030-kb-blocker-lifecycle-contract-manifest.md"
  - gate_id: work-to-complete
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "both slices are done"
      - "manifest aggregate proof passes"
    proof:
      - cmd/kbcheck/manifest_contract_test.go
      - cmd/kbcheck/communication_contract_test.go
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-finalize docs/plans/2026-07-26-030-kb-blocker-lifecycle-contract-manifest.md"
slices:
  - id: slice-031
    title: "Enforce blocker lifecycle metadata and stale rechecks"
    path: docs/plans/2026-07-26-031-tool-blocker-lifecycle-contract-plan.md
    blockers: []
    verification: tdd
    test_level: integration
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "The parser and validator affect every opt-in KB manifest and must remain backward compatible."
    model_requirements: ["Go parser changes", "negative contract fixtures", "portable Python parity", "backward compatibility"]
    escalation_triggers: ["legacy manifests become invalid without opting in", "release-only scope can block implementation", "stale timestamps pass", "Python and Go disagree"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'TestBlockerLifecycle|TestManifestContract' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
  - id: slice-032
    title: "Apply pause, ownership, and propagation rules across KB skills"
    path: docs/plans/2026-07-26-032-skill-blocker-ownership-propagation-plan.md
    blockers: [slice-031]
    verification: integration
    test_level: integration
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "The same status semantics must survive planning, execution, durable goals, completion, and user-facing summaries."
    model_requirements: ["workflow policy editing", "status transition reasoning", "contract-test coverage", "global skill drift reconciliation"]
    escalation_triggers: ["a real safety approval is weakened", "pause mutates after a plain stop", "optional gates still roll up", "new global drift appears"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'TestBlockerLifecycle|TestLowCognitiveBurdenCommunicationContract' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
---

# KB: Blocker Lifecycle and Acceptance Scope

Prevent safe work from stopping because a pause, agent-owned repair, optional
capability, stale dependency, or release-only receipt was mislabeled as a
whole-objective blocker.

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 31 | Enforce blocker lifecycle metadata and stale rechecks | - | tdd | no | done |
| 32 | Apply pause, ownership, and propagation rules across KB skills | 31 | integration | no | done |

Focused Go and portable Python proof passed. The broader `local-release` command
was attempted once with a 90-second bound and timed out in the already-tracked
harness validation defect; it is a delivery-harness issue, not failed evidence
for these two slices.
