---
kb_id: kb-2026-07-30-model-tier-qualification
slice_id: slice-001
title: "Emit and validate qualification-plan receipts"
blockers: []
verification: tdd
test_level: integration
functional_risk: broad
model_tier: large
model_tier_reason: "This changes planning policy and the causal boundary between plan defects and model failures."
model_requirements: ["Go contract design", "YAML and strict JSON validation", "planning-policy reasoning", "adversarial fixture design"]
escalation_triggers: ["ordinary non-qualification plans become invalid", "semantic model judgment becomes a deterministic pass condition", "legacy manifests become unreadable", "a new DDR-specific planner is proposed"]
workspace_mode: shared-serial
conflict_domains: ["file:.github/skills/kb-plan/SKILL.md", "file:cmd/kbcheck/manifest_contract.go", "namespace:qualification-plan"]
shared_resources: ["git:integration-owner", "sync:global-skills"]
context_packet_path: docs/plans/2026-07-30-model-tier-qualification-context/slice-001.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'QualificationPlan|Manifest.*Qualification' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Define the opt-in qualification-evidence plan record and keep ordinary plans outside the stricter gate."
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "Validate qualification-plan paths, hashes, invariant coverage, source specificity, tier raises, and review bindings."
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: "Protect valid, missing, restated, generic, stale-hash, tier-raise, and legacy-manifest cases."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Document qualification-plan ownership without creating a DDR-specific planner."
protected_oracles:
  - path: cmd/kbcheck/manifest_contract_test.go
    role: "qualification-plan and legacy-manifest contract oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Write protected qualification-plan contract tests, then add the opt-in kb-plan record and kbcheck validation."
human_action: ""
can_continue_other_slices: true
---

# Emit and Validate Qualification-Plan Receipts

## What to Build

Add an opt-in qualification-evidence record to `kb-plan`. It enumerates stable
invariants and binds each to repository-specific mechanism/hazard guidance or
an explicit uncertainty-driven tier raise. `kbcheck` validates and hashes that
record; ordinary plans remain backward-compatible.

## Acceptance Criteria

- Only plans explicitly admitted as qualification evidence require the stricter
  record.
- Every nontrivial invariant has a stable ID, requirement, repository-specific
  source/hazard, executor action, and proof target, or a tier raise.
- Normalized restatements and generic hazard prose fail.
- The record binds the exact plan and applicable plan-wide review receipt.
- Missing, escaped, unreadable, stale-hash, duplicate, or unknown data fails
  closed.
- Existing manifests without the opt-in remain readable and valid.
- `kb-plan`, `document-review`, and `kbcheck` ownership stays explicit; no new
  DDR-specific planner is added.

## Test Scenarios

- Valid mechanism and valid hazard records pass.
- Explicit uncertainty with a higher target tier passes.
- Acceptance-criterion restatement, generic "be careful" guidance, empty
  executor action, and missing proof target fail.
- Duplicate invariant IDs and stale plan/review hashes fail.
- Escaped or symlinked record paths fail.
- Ordinary schema-2 and legacy manifests remain valid.

## Scope Boundary

No live model call, model classification, route selection, permanent model
assignment, or general semantic grading of every KB plan.

