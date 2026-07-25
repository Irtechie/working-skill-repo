---
kb_id: kb-2026-07-19-evidence-graph-routing
slice_id: slice-004
title: "Prefer exact-symbol evidence through an optional adapter"
blockers: [slice-003]
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: medium
model_tier_reason: "Provider and packet boundaries are settled; fixture-driven exact-symbol ingestion is bounded and objectively rejectable."
model_requirements: ["SCIP/LSP semantics", "adapter fixtures", "optional-tool detection", "source-span provenance"]
escalation_triggers: ["new required dependency or daemon", "unstable symbol identity", "source verification bypass", "cross-repo semantics exceed the current local contract"]
context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-004.json
proof_check:
  kind: command_exit
  command: "go test ./internal/graphrouting -run 'ExactSymbol|SCIP|Fallback'"
  expect: 0
hitl: false
expected_files:
  - path: internal/graphrouting/provider.go
    op: create
    scope: "Optional provider capability and result interface."
  - path: internal/graphrouting/scip.go
    op: create
    scope: "Consume an already-generated exact-symbol index or deterministic snapshot."
  - path: internal/graphrouting/scip_test.go
    op: create
    scope: "Definition, reference, implementation, source-span, stale, and unavailable fixtures."
  - path: evals/graph-routing/exact-symbol-index.json
    op: create
    scope: "Protected polyglot exact-symbol fixture."
  - path: .github/skills/kb-map/references/graph-routing.md
    op: edit
    scope: "Exact-symbol precedence and optional-adapter invocation."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Exact-symbol adapter fixture command and fallback expectations."
protected_oracles:
  - path: evals/graph-routing/exact-symbol-index.json
    role: "definition/reference/implementation resolution fixture"
    sha256: "8D3296B66A3E045803E0B0C337611FAF88EADFB351F374DD81A794B8CEA2729D"
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Proceed to slice 005 structural and flow traversal recipes."
human_action: ""
can_continue_other_slices: true
---

# Slice 004: Optional Exact-Symbol Evidence

## What To Build

Add one optional exact-symbol adapter behind slice 001's contract. Prefer an
already-produced SCIP/LSP-grade index and normalize definitions, references,
implementations, overrides, diagnostics, and source spans into the impact packet.

## Why This Slice Exists

Exact symbol identity is the highest-value precision improvement. Structural
patterns and embeddings can find plausible text while still confusing scopes,
overloads, aliases, or implementations.

## Acceptance Criteria

- Exact-symbol evidence outranks structural, semantic, and inferred candidates.
- Adapter absence/staleness is observable and falls back to file-native lookup.
- No tool is downloaded, installed, started, or required automatically.
- Symbol and source identities bind to repo/revision/worktree fingerprint.
- Cross-repo references are not claimed unless the index explicitly proves them.
- Load-bearing edges retain source locations for later verification.

## Test Scenarios

- Resolve definition/reference/implementation across fixture files.
- Distinguish same-spelling symbols in different scopes.
- Reject mismatched revision and out-of-root source spans.
- Missing adapter/index returns explicit fallback without failing base lookup.

## Proof Check

`go test ./internal/graphrouting -run 'ExactSymbol|SCIP|Fallback'`

## Scope Boundary

No auto-install, language server lifecycle, global graph store, flow analysis,
or embedding/vector search.

## Dependencies

Slice 003 is required so this and slice 005 can execute concurrently without
sharing a mutable checkout or lifecycle files.

## Concurrency

May run with slice 005 only after acquiring a disjoint lease and isolated
worktree. Integration remains serialized.

Execution note: completed under one coordinator in the canonical checkout
because the slice 001-003 baseline is uncommitted and a fresh worktree from
`HEAD` would not contain it. No other mutating slice ran concurrently.
