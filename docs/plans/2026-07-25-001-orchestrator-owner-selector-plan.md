---
type: kb-slice-plan
kb_id: kb-2026-07-25-orchestrator-directed-ddr
slice_id: ddr-001
title: Orchestrator ownership selector
status: done
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: narrow
model_tier: large
model_tier_reason: "Selector semantics span runtime capability, CLI receipts, and compatibility validation."
model_requirements: ["Go implementation", "CLI contract reasoning", "compatibility judgment"]
escalation_triggers: ["archived manifests fail", "unrelated selector callers require semantic changes"]
workspace_mode: worktree-required
conflict_domains: ["internal/modelrouting", "cmd/kbrouter", "cmd/kbcheck"]
shared_resources: ["git:integration-owner"]
context_packet: docs/plans/2026-07-25-orchestrator-directed-ddr-context/ddr-001.json
expected_files:
  - path: internal/modelrouting/selector.go
    op: edit
    scope: add owner-first selection
  - path: internal/modelrouting/selector_test.go
    op: edit
    scope: prove current and delegated ownership semantics
  - path: cmd/kbrouter/select.go
    op: edit
    scope: add CLI ownership inputs and receipt fields
  - path: cmd/kbrouter/select_test.go
    op: edit
    scope: prove public CLI contract
  - path: cmd/kbrouter/catalog.go
    op: edit
    scope: make discovered current capability usable without fabricating delegated readiness
  - path: cmd/kbrouter/catalog_test.go
    op: edit
    scope: prove real discover-to-current selection
  - path: cmd/kbrouter/dispatch.go
    op: edit
    scope: bind dispatch to the delegated capability envelope
  - path: cmd/kbrouter/dispatch_test.go
    op: edit
    scope: reject dispatch that bypasses selector capability
  - path: cmd/kbrouter/dispatch_c1_test.go
    op: edit
    scope: keep dispatch contract fixture ownership metadata valid
  - path: cmd/kbrouter/dispatch_c2_test.go
    op: edit
    scope: keep dispatch contract fixture ownership metadata valid
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: validate opted-in ownership contract
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: preserve legacy compatibility and reject invalid new manifests
proof_check:
  kind: command_exit
  command: "go test ./internal/modelrouting ./cmd/kbrouter ./cmd/kbcheck"
  expect: 0
---

# DDR-001: Orchestrator ownership selector

## Goal

Add an owner-first selector contract and CLI receipt.

## Scope

- `internal/modelrouting/selector.go`
- selector tests
- `cmd/kbrouter/select.go`
- CLI tests
- `cmd/kbcheck/manifest_contract.go`
- manifest contract tests

## Contract

1. Require the orchestrator's execution owner, owner reason, required tier, and
   tier reason for the normal CLI path.
2. For `current`, validate the current route and do not inspect workers.
3. For `delegated`, return exactly one qualified worker and never fall back to
   current automatically.
4. Preserve lower-tier attempt fields only for explicit experimental AMR
   compatibility.
5. Validate new manifest fields only when `model_selection_contract` exists.

## Proof

- Focused Go tests for model routing, kbrouter selection, and manifest parsing.
- `git diff --check`.

## Acceptance criteria

- Eligible workers do not override a `current` ownership decision.
- Delegated selection emits one route and no current fallback.
- Dispatch revalidates the delegated capability envelope and fails closed on
  catalog drift or missing ownership evidence.
- Missing owner/tier reasons fail the new public contract.
- Archived manifests without the opt-in contract remain readable.
