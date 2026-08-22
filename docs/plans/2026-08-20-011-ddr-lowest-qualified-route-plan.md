---
kb_id: kb-2026-08-20-routing-cognitive-delivery
slice_id: routing-001
title: Enforce exact-tier-or-higher DDR selection or typed retain-current exception
blockers: [runtime-002]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: Selection affects trust, capability, host-local catalogs, and mutation authority.
model_requirements: [routing policy, Go CLI integration, adversarial fixture design]
escalation_triggers: [route lacks required tools or context, no qualified route is unproven, fallback retries]
token_budget: 7000
workspace_mode: shared-serial
conflict_domains: [namespace:kbrouter-selection, skill:kb-work, file:cmd/kbcheck/model_routing_release.go]
proof_check: {kind: command_exit, command: "go test ./cmd/kbrouter -run 'Select|DDR' -count=1", expect: 0}
hitl: false
status: pending
---

# Enforce Exact-Tier-or-Higher DDR Selection

## Acceptance Criteria

- Execution cannot mutate without a qualified delegated route receipt or one of
  the recognized, evidence-backed retain-current reasons.
- Selection prefers an exact-tier qualified callable route, then a qualified
  higher-tier route, from the active host surface and user-local catalog with
  deterministic tie-breaking. It never automatically routes downward.
- No concrete route catalog is committed to the repository.

## Test Scenarios

- A small qualified route delegates a small slice only when planning classified
  that slice as `small`.
- Missing tool/context and no-qualified-route claims require a valid receipt.
- An ineligible cheaper route, a stale route, and a second retry are rejected.
