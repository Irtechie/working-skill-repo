---
type: kb-manifest
kb_id: kb-2026-07-14-routing-owner-contract
created: 2026-07-14
status: completed
workflow_shape: skill-bundle-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck core"
  expect: 0
  why: "Proves skill, README, routing, and manifest contracts remain valid."
model_tier_contract:
  allowed: [small, medium, large]
  default: medium
model_selection_contract:
  timing: work-time
  fallback: same-tier-then-higher-then-explicit-current
  automatic_downgrade: false
scope-verified-files: [".github/skills/kb-plan/SKILL.md", ".github/skills/kb-work/SKILL.md", ".github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1", "README.md", "cmd/kbcheck/manifest_contract.go", "cmd/kbcheck/manifest_contract_test.go", "cmd/kbcheck/swarm.go", "cmd/kbrouter/select.go", "cmd/kbrouter/select_test.go", "internal/modelrouting/selector.go", "internal/modelrouting/selector_test.go"]
gate_ledger:
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence: ["manifest and slice plans exist", "DAG is acyclic", "each slice has proof and model justification"]
    proof: ["docs/plans/2026-07-14-050-kb-routing-owner-contract-manifest.md", "docs/plans/2026-07-14-051-routing-plan-contract-plan.md", "docs/plans/2026-07-14-052-routing-execution-owner-plan.md"]
    blockers: []
    passed_at: "2026-07-14T19:00:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-14-050-kb-routing-owner-contract-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence: ["manifest validator rejects unjustified tiers", "focused tests pass"]
    proof: ["go test -run TestManifestContract -count=1 ./cmd/kbcheck", "go run ./cmd/kbcheck manifest-contract --manifest docs/plans/2026-07-14-050-kb-routing-owner-contract-manifest.md"]
    blockers: []
    passed_at: "2026-07-14T21:00:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-14-050-kb-routing-owner-contract-manifest.md"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence: ["selector proof passes", "README claims are bounded", "global copies hash-match"]
    proof: ["go test -count=1 ./internal/modelrouting", "go test -count=1 -timeout=20m ./cmd/kbrouter", "go run ./cmd/kbcheck skill-sync-report"]
    blockers: []
    passed_at: "2026-07-14T21:00:00-04:00"
    allowed_next_action: "kb-finalize docs/plans/2026-07-14-050-kb-routing-owner-contract-manifest.md"
  - gate_id: work-to-complete
    owner_skill: kb-work
    status: passed
    required_evidence: ["all slices done", "core passes", "local-release passes", "scope is verified"]
    proof: ["go run ./cmd/kbcheck core", "go run ./cmd/kbcheck local-release"]
    blockers: []
    passed_at: "2026-07-14T21:00:00-04:00"
    allowed_next_action: "kb-finalize docs/plans/2026-07-14-050-kb-routing-owner-contract-manifest.md"
slices:
  - id: slice-001
    title: "Enforce complexity-based planning metadata"
    path: docs/plans/2026-07-14-051-routing-plan-contract-plan.md
    blockers: []
    verification: tdd
    test_level: unit
    functional_risk: none
    model_tier: medium
    model_tier_reason: "Parser, validator, template, and compatibility fixtures must change together."
    model_requirements: ["Go parser changes", "skill contract editing", "objective unit proof"]
    escalation_triggers: ["existing manifest compatibility breaks", "schema requires nested YAML parsing"]
    proof_check:
      kind: command_exit
      command: "go test -run TestManifestContract -count=1 ./cmd/kbcheck"
      expect: 0
    hitl: false
    status: done
    notes: "manifest-contract and focused kbcheck tests pass; tier reason, requirements, and escalation triggers are now enforced"
  - id: slice-002
    title: "Default to delegated workers and label experimental routing"
    path: docs/plans/2026-07-14-052-routing-execution-owner-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: unit
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "Selection policy, CLI receipts, skill wording, and README claims form one bounded routing contract."
    model_requirements: ["Go selector changes", "policy reasoning", "cross-install sync"]
    escalation_triggers: ["current route cannot prove planned capability", "AMR becomes required", "sync drift remains"]
    proof_check:
      kind: command_exit
      command: "go test -count=1 ./internal/modelrouting"
      expect: 0
    hitl: false
    status: done
    notes: "selector and full kbrouter tests pass; AMR remains optional; global sync is 129/129 clean; useful shared-agent regression-snapshot fix merged"
---

# Routing Owner Contract

Keep DDR complexity classification and work-time worker selection mandatory,
while AMR lower-tier attempts remain optional, disabled by default, and testing.
Current-reasoner execution requires an explicit runtime reason and capability
floor instead of silently replacing unavailable delegation.
