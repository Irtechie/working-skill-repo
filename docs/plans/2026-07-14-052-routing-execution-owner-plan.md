---
kb_id: kb-2026-07-14-routing-owner-contract
slice_id: slice-002
title: "Default to delegated workers and label experimental routing"
blockers: [slice-001]
verification: tdd
test_level: unit
functional_risk: narrow
model_tier: medium
model_tier_reason: "Selection policy, CLI receipts, skill wording, and README claims form one bounded routing contract."
model_requirements: ["Go selector changes", "policy reasoning", "cross-install sync"]
escalation_triggers: ["current route cannot prove planned capability", "AMR becomes required", "sync drift remains"]
expected_files:
  - {path: "internal/modelrouting/selector.go", op: edit, scope: "Require explicit qualified current-reasoner execution."}
  - {path: "cmd/kbrouter/select.go", op: edit, scope: "Expose owner decision and reason receipt."}
  - {path: ".github/skills/kb-work/SKILL.md", op: edit, scope: "Make delegation default and AMR optional."}
  - {path: ".github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1", op: edit, scope: "Merge useful global-only native stderr and return handling before sync."}
  - {path: "README.md", op: edit, scope: "Mark DDR testing and AMR optional/testing."}
protected_oracles: []
status: done
hitl: false
---

Acceptance: ordinary work delegates when an eligible route exists; current
execution is explicit and qualified; AMR stays disabled by default; README
does not present DDR or AMR as promoted.
