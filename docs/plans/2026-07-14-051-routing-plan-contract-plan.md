---
kb_id: kb-2026-07-14-routing-owner-contract
slice_id: slice-001
title: "Enforce complexity-based planning metadata"
blockers: []
verification: tdd
test_level: unit
functional_risk: none
model_tier: medium
model_tier_reason: "Parser, validator, template, and compatibility fixtures must change together."
model_requirements: ["Go parser changes", "skill contract editing", "objective unit proof"]
escalation_triggers: ["existing manifest compatibility breaks", "schema requires nested YAML parsing"]
expected_files:
  - {path: ".github/skills/kb-plan/SKILL.md", op: edit, scope: "Require tier justification and escalation metadata."}
  - {path: "cmd/kbcheck/swarm.go", op: edit, scope: "Parse routing justification fields."}
  - {path: "cmd/kbcheck/manifest_contract.go", op: edit, scope: "Reject unjustified tier labels."}
  - {path: "cmd/kbcheck/manifest_contract_test.go", op: edit, scope: "Prove missing fields fail."}
protected_oracles: []
status: done
hitl: false
---

Acceptance: every new model-tier contract carries a reason, requirements, and
observable escalation triggers; invalid manifests fail deterministically.
