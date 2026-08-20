---
kb_id: kb-2026-08-20-surface-audit-and-amr-retirement
slice_id: audit-002
title: Audit candidate skills and remove active AMR routes
blockers: [audit-001]
verification: tdd
test_level: integration
functional_risk: broad
model_tier: large
model_tier_reason: Removes public skills and benchmark paths while preserving required behavior and history.
model_requirements: [migration design, search-based impact analysis, compatibility testing]
escalation_triggers: [explicit TDD behavior regresses, portable trust brake is lost, active AMR reference remains]
token_budget: 7000
workspace_mode: shared-serial
conflict_domains: [skill:tdd, namespace:amr, config:removed-skills]
proof_check: {kind: command_exit, command: "go run ./cmd/kbcheck core", expect: 0}
hitl: false
status: pending
---

# Audit Candidate Skills and Remove Active AMR Routes

## Acceptance Criteria

- A parity matrix compares `tdd`'s explicit RED result, protected-oracle hash,
  GREEN result, and unchanged-oracle verification against normal KB work. A
  failing or incomplete parity result retains `tdd`.
- `kb-first-principles` remains a distinct portable conversational trust brake;
  it is excluded from execution inner loops, not retired.
- No active normal-work, acceptance, router, or benchmark route advertises AMR;
  DDR remains the lower-tier execution path.
- The audit ranks candidate skills by hot-path cost, caller count, duplicated
  behavior, and replacement proof. It proposes at most six candidates and does
  not delete any skill without separate user approval.
