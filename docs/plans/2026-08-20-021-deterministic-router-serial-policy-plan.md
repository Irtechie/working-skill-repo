---
kb_id: kb-2026-08-20-surface-audit-and-amr-retirement
slice_id: audit-001
title: Replace verbose route table with deterministic classifier and serial-first policy
blockers: [routing-003]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: Changes default task routing and worktree concurrency semantics.
model_requirements: [classifier design, workflow safety, Git worktree testing]
escalation_triggers: [classifier loses route explainability, parallel execution lacks isolation receipt]
token_budget: 6500
workspace_mode: shared-serial
conflict_domains: [skill:kb-start, namespace:route-classifier, namespace:worktree-policy]
proof_check: {kind: command_exit, command: "go run ./cmd/kbcheck route-eval", expect: 0}
hitl: false
status: pending
---

# Deterministic Router and Serial-First Policy

## Acceptance Criteria

- Route classification is compact, deterministic, explainable, and fixture-tested.
- `kb-start` becomes a thin adapter instead of a large ranking table.
- One branch/worktree and serial mutation are defaults; parallel execution needs
  explicit opt-in plus a disjointness and benefit receipt.
