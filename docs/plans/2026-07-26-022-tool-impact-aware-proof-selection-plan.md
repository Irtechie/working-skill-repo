---
kb_id: kb-2026-07-26-change-aware-proof-governor
slice_id: slice-002
title: "Select only invalidated proof and reuse passing supersets"
blockers: [slice-001]
verification: tdd
test_level: functional-cli
functional_risk: broad
execution_class: cli
model_tier: large
model_tier_reason: "Coverage subsumption and change impact are the core soundness boundary; a false reuse is worse than a conservative rerun."
model_requirements: ["set-subsumption reasoning", "change-impact explanation", "deterministic CLI decisions", "conservative invalidation fallbacks"]
escalation_triggers: ["one changed input does not invalidate every dependent check", "the decision lacks exact invalidating paths", "overlapping checks run twice against one fingerprint", "unknown dependency metadata is treated as unchanged"]
context_packet_path: docs/plans/2026-07-26-proof-governor-context/slice-002.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'ProofSelection|CoverageSubsumption|ImpactInvalidation' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/proof_selection.go
    op: create
    scope: "Compute RUN, REUSE, or BLOCK from requested checks, receipts, current inputs, and replay policy."
  - path: cmd/kbcheck/proof_selection_test.go
    op: create
    scope: "Protect coverage subsumption, changed-input invalidation, and explanation output."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Expose proof-plan JSON and human-readable decision output."
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "Provide stable child check identities and dependency metadata for discovered checks."
  - path: cmd/kbcheck/release.go
    op: edit
    scope: "Collapse child checks already proven in the same immutable release state while retaining fresh aggregate release proof."
protected_oracles:
  - path: cmd/kbcheck/proof_selection_test.go
    role: "changed-input invalidation and passing-superset reuse oracle"
    sha256: "79fe938cefa3070f9252e7e29573ac24a30e0023e0abfbb01ea2435d888ad439"
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: false
---

# Slice 002: Impact-Aware Proof Selection

## What to Build

Add a deterministic planning command that receives requested check IDs and
returns the smallest sound run set. It may reuse a fresh passing superset
receipt, but only under the slice-001 contract. It must never infer “unchanged”
from chat history or agent confidence.

## Acceptance Criteria

1. `proof-plan` returns `RUN`, `REUSE`, or `BLOCK` with stable machine-readable
   reason codes and concise human output.
2. `RUN` includes every check invalidated by changed relevant inputs and names
   the exact paths and old/new fingerprints responsible.
3. `REUSE` names the receipt and proves requested IDs are a subset of its
   enumerated passing checks; duplicate requested checks collapse to one run.
4. Ambiguous impact, stale metadata, missing environment fields, and partial
   results choose conservative execution.
5. Selection is deterministic across path aliases and Windows path casing after
   canonical project-root resolution.
6. Composite profiles execute each stable child check at most once per proof
   state, including children inherited through `core` and `local-release`.
7. Integration-side aggregate proof still runs after composition. Worker,
   routing, worktree, or lease receipts remain supporting evidence unless the
   exact integrated state and aggregate oracle are freshly bound.

## Test Scenarios

- A full receipt satisfies three later subset requests without executing.
- Editing one relevant file selects the affected check once and preserves reuse
  for unaffected checks.
- Editing a shared dependency invalidates every registered dependent.
- Unknown impact and registry drift produce `RUN` with an explanation, not a
  silent cache hit.
- A release profile that contains `core` and an already-enumerated child does
  not execute that child twice against the same state.
- A worker receipt from a pre-integration state cannot satisfy the integrated
  aggregate check.

## Scope Boundary

This slice decides what should run. It does not yet launch commands, impose
replay budgets, or change the KB skills.

## Proof

`go test ./cmd/kbcheck -run 'ProofSelection|CoverageSubsumption|ImpactInvalidation' -count=1`
