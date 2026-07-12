---
kb_id: kb-2026-07-11-ghcp-aic-falsification
slice_id: slice-004
title: "Grade paired full-fallback AMR without promotion theater"
blockers: [slice-003]
verification: functional
test_level: full
functional_risk: full
model_tier: large
context_packet_path: docs/plans/2026-07-11-ghcp-aic-context/slice-004.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/amrbench ./cmd/kbcheck -run 'Conformance|Paired|Promotion|GHCP' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/amrbench/grade.go
    op: create
    scope: "Pair direct and full-fallback arms by task/seed and compute correctness, exact total AIU, latency, fallback tail, intervention, and invalid outcomes."
  - path: cmd/amrbench/grade_test.go
    op: create
    scope: "Reject phase buckets, mismatched routes, partial telemetry, pooled families, median-only wins, and insufficient correctness bounds."
  - path: cmd/amrbench/main.go
    op: edit
    scope: "Expose no-paid conformance and attended-stage transitions without launching models by default."
  - path: evals/amr-model-benchmark/config.json
    op: edit
    scope: "Version task-family admission, planned/attempt tiers, observed route tuples, corpus split, and stage budgets without hosted model lists."
  - path: evals/amr-model-benchmark/README.md
    op: edit
    scope: "State exact current capabilities, invalid first canary, full-fallback-only scope, and explicit approval commands."
  - path: cmd/kbcheck/model_routing_ghcp_release.go
    op: create
    scope: "Reuse shared promotion math through a separate later-cohort validator without changing active Codex-first behavior."
  - path: cmd/kbcheck/model_routing_ghcp_release_test.go
    op: create
    scope: "Prove family-specific aggregate-plus-median benefit, no correctness/intervention regression, and landed Codex-cohort non-regression."
  - path: docs/context/eval-map.md
    op: edit
    scope: "Document deterministic conformance and attended live stages separately."
protected_oracles:
  - path: cmd/amrbench/grade_test.go
    role: "paired correctness, all-inclusive AIU, family admission, and not-promoted oracle"
    sha256: "filled after RED before implementation"
    update_policy: "requires explicit plan update"
  - path: cmd/kbcheck/model_routing_ghcp_release_test.go
    role: "GHCP follow-on promotion oracle that preserves the landed Codex-first cohort"
    sha256: "filled after RED before implementation"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
can_continue_other_slices: false
---

# Slice 004 - Paired Falsification Grader

## Acceptance

- Direct and AMR are paired by task/seed, frozen observed routes, base packet,
  role overlays, terminal criteria, and proof; direct is planned-tier and AMR has
  exactly one next-lower attempt plus planned-tier-or-higher full fallback.
- Every leaf call across orchestration reconciles to the arm total. Requested/
  actual mismatch, partial telemetry, oracle mutation, isolation failure, or
  missing proof yields `invalid`, not a model failure or win.
- Promotion is per eligible family and requires the existing correctness bound,
  no increased human intervention, at least 20% paired median AIU reduction, and
  lower aggregate AIU; mean, median, total, p90, fallback, and repair tail report.
- Existing `qualified`/`suspended` labels remain disabled until a separately
  approved held-out matrix provides enough evidence.
- `conformance --no-paid --require-ready` is the terminal deliverable. It
  constructs only DisabledRunner, loads no profile/secret, spawns no provider,
  and returns nonzero unless every required deterministic readiness check passes.

## Scope Boundary

No live canary/matrix, production partial reuse, adaptive routing, or change to
the active Codex-first cohort.
