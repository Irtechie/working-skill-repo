---
kb_id: kb-2026-07-11-ghcp-aic-falsification
slice_id: slice-003
title: "Predeclare the context-diet A-B contract"
blockers: [slice-002]
verification: functional
test_level: integration
functional_risk: broad
model_tier: medium
context_packet_path: docs/plans/2026-07-11-ghcp-aic-context/slice-003.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/amrbench -run 'Context|Contract|Crossover|Winner|Holdout' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/amrbench/context.go
    op: create
    scope: "Build and hash baseline/minimal base packets plus preregistered role overlays without route-specific advice."
  - path: cmd/amrbench/context_test.go
    op: create
    scope: "Prove disjoint development/routing corpora, crossover order, frozen overlays, and non-inferior winner rule."
  - path: evals/amr-model-benchmark/context-contract.schema.json
    op: create
    scope: "Version context variants, corpora, proof/tool/timeout/stopping parity, and exact winner criteria."
  - path: evals/amr-model-benchmark/context-contracts/baseline.json
    op: create
    scope: "Hash the baseline context contract used for comparison."
  - path: evals/amr-model-benchmark/context-contracts/minimal.json
    op: create
    scope: "Hash the candidate minimal contract without deleting task-critical context."
protected_oracles:
  - path: cmd/amrbench/context_test.go
    role: "corpus separation, comparison parity, and winner-rule behavior oracle"
    sha256: "filled after RED before implementation"
    update_policy: "requires explicit plan update"
  - path: evals/amr-model-benchmark/context-contract.schema.json
    role: "predeclared disjoint context A-B contract and winner-rule oracle"
    sha256: "filled after RED before implementation"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
can_continue_other_slices: false
---

# Slice 003 - Predeclare Context Diet A-B

## Acceptance

- Context development and held-out routing corpora are disjoint and hash-bound.
- Baseline and minimal contracts share task, observed route, proof, tool, timeout,
  stopping, and order controls; only ambient context differs.
- The contract predeclares that minimal can win only with zero right-to-wrong
  outcomes, non-inferior correctness, lower aggregate AIU, and an at-least-10%
  paired median reduction whose paired confidence interval excludes zero;
  otherwise baseline stays.
- This slice implements deterministic contracts and simulation only. A later
  attended approval is required to spend credits on the context A/B.
- No context winner is selected or frozen in this slice. The attended A/B
  evidence, if later approved, selects and then freezes a winner.

## Scope Boundary

No route matrix, adaptive policy, production skill trimming, or paid call.
