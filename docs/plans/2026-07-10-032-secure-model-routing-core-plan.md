---
kb_id: kb-2026-07-10-session-model-routing
slice_id: slice-002
title: "Select routes conservatively from a secure session catalog"
blockers: []
verification: tdd
test_level: unit
functional_risk: broad
model_tier: large
context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-002.json
proof_check:
  kind: command_exit
  command: "go test ./internal/modelrouting"
  expect: 0
hitl: false
expected_files:
  - path: internal/modelrouting/catalog.go
    op: create
    scope: "versioned route, readiness, session catalog, fingerprint, and support-cohort types"
  - path: internal/modelrouting/storage.go
    op: create
    scope: "permission-preserving atomic JSON storage with path/link/size guards"
  - path: internal/modelrouting/policy.go
    op: create
    scope: "user/project trust merge, project identity, run override, and sensitive-data eligibility"
  - path: internal/modelrouting/selector.go
    op: create
    scope: "difficulty floor, evidence fit, same-class fallback, qualified escalation, and current-model degradation"
  - path: internal/modelrouting/receipt.go
    op: create
    scope: "attempt ledger, route evidence, mismatch, and capability-credit rules"
  - path: internal/modelrouting/identity_windows.go
    op: create
    scope: "replacement-resistant canonical project identity on Windows"
  - path: internal/modelrouting/identity_unix.go
    op: create
    scope: "replacement-resistant canonical project identity on macOS and Linux"
  - path: internal/modelrouting/selector_test.go
    op: create
    scope: "protected selection, trust, fallback, and no-false-credit oracle"
protected_oracles:
  - path: internal/modelrouting/selector_test.go
    role: "security and conservative selection oracle"
    sha256: "3e241a0c74d2020fcc1b535c45b45604b042c912292698be3b66bdae3e9a5465"
    update_policy: "requires explicit plan update"
status: done
owner: agent
can_continue_other_slices: true
---

# Secure Routing Core

Oracle update (slice-003 security remediation): the protected selector oracle was extended with opaque source identity, host-derived evidence, bounded DNS, ACL, storage-profile, concurrency, and current-route tamper regressions. Existing selector/fallback assertions were retained. New SHA-256: `3e241a0c74d2020fcc1b535c45b45604b042c912292698be3b66bdae3e9a5465`.

## What To Build

Create a provider-neutral Go package that merges native discovery, user-local extras, project policy, and run overrides into one redacted session catalog, then selects only dispatch-proven eligible routes.

## Acceptance Criteria

- Catalog readiness is cumulative and automatic selection uses only `dispatch-proven` routes.
- Stronger eligible routes may handle lower tiers; no automatic downward fallback occurs.
- Unknown capability, stale evidence, weak trust metadata, or model mismatch never earns automatic eligibility/credit.
- User/project policy cannot activate private routes or cross a trust boundary silently.
- Storage rejects traversal, unsafe links, oversized input, credential values, unsafe endpoints, metadata targets, and cross-origin auth forwarding.

## Test Scenarios

- Same-class fallback, evidence-qualified escalation, exact `require`, preferred `use`, ignore routing, and current-model degradation.
- Canonical project identity across normal path/worktree, approval mismatch across unrelated clone, and project policy narrowing.
- Loopback/private HTTP approval, public TLS, link-local rejection, DNS rebind defense seam, and auth-origin binding.
- Exact route receipt credits success; missing/mismatched identity records observation only.

## Tier Rationale

Large: security, policy, provenance, and fallback rules are architecture-shaping and repeat-work-sensitive.

## Scope Boundary

No host process launch, TinyBoss state changes, MCP dispatch, or generated agents.
