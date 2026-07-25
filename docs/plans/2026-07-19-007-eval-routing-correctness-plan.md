---
kb_id: kb-2026-07-19-evidence-graph-routing
slice_id: slice-007
title: "Falsify routing correctness and multi-session safety claims"
blockers: [slice-006]
verification: functional
test_level: functional-cli
functional_risk: broad
model_tier: medium
model_tier_reason: "The architecture is settled; this is bounded adversarial fixture and scorer work with deterministic promotion criteria."
model_requirements: ["adversarial fixtures", "two-process race harness", "impact recall/false-positive scoring", "stale and dirty-state tests"]
escalation_triggers: ["race tests are flaky", "token savings can mask missed impact", "required gate depends on an optional provider", "a safety claim lacks a negative fixture"]
context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-007.json
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck graph-routing-eval --require-ready"
  expect: 0
hitl: false
expected_files:
  - path: evals/graph-routing/README.md
    op: create
    scope: "Corpus, metrics, promotion rules, and provider-optional behavior."
  - path: evals/graph-routing/expected-results.json
    op: create
    scope: "Protected correctness and concurrency expectations."
  - path: evals/graph-routing/fixtures/symbol-impact.json
    op: create
    scope: "Definitions, implementations, generated/config consumers, tests, and docs."
  - path: evals/graph-routing/fixtures/stale-index.json
    op: create
    scope: "Revision/worktree mismatch and fallback cases."
  - path: evals/graph-routing/fixtures/multisession-race.json
    op: create
    scope: "Same-slice race, disjoint worktrees, prefix collisions, and stale recovery."
  - path: cmd/kbcheck/graph_routing_eval.go
    op: create
    scope: "Score correctness, retrieval cost, and concurrency separately."
  - path: cmd/kbcheck/graph_routing_eval_test.go
    op: create
    scope: "Positive, negative, unavailable-provider, and anti-gaming tests."
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "Add required deterministic readiness selftest without optional providers."
  - path: docs/context/eval-map.md
    op: edit
    scope: "Canonical graph-routing proof commands and open live-provider gaps."
protected_oracles:
  - path: evals/graph-routing/expected-results.json
    role: "routing correctness and concurrency promotion oracle"
    sha256: "0A01D0543B363E41EF801A2E823AC908286A6F674A10EDF52DE426F719519CBF"
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Proceed to slice 008 documentation, sync, and release proof."
human_action: ""
can_continue_other_slices: false
---

# Slice 007: Routing And Concurrency Falsification

## What To Build

Create a deterministic corpus and CLI gate that scores routing correctness and
multi-session safety separately from retrieval cost. Required checks run with
no optional providers; provider-specific live evaluations remain explicit.

## Why This Slice Exists

A benchmark that only measures tokens can reward incomplete impact sets.
Likewise, worktree creation can look safe while duplicate claims or mergeback
still race. Promotion requires negative fixtures that would expose both lies.

## Acceptance Criteria

- Report impacted-symbol/file recall and missed tests/docs.
- Report false positives per retrieved token separately from correctness.
- Fail on stale index acceptance, hidden dynamic limitations, or uncited exact edges.
- Deterministically prove one winner for a same-slice two-process race.
- Prove disjoint worktrees can complete and integrate serially.
- Fail path-prefix/generated/browser/index conflict collisions before mutation.
- Prove dirty source checkout preservation and no-force cleanup behavior.
- Optional provider absence is `skipped-unavailable`, not a failed core gate or invented result.
- Promotion requires zero safety invariant failures and a stated correctness threshold.

## Test Scenarios

- Interface implementation plus same-name distractor.
- Generated consumer and config-string registration.
- Reflection/dynamic dispatch with explicit incomplete result.
- Stale clean-index used in a dirty worktree.
- Same-slice race, stale owner recovery, base drift, and merge conflict.
- Token-cheap packet that misses a required consumer fails correctness.

## Proof Check

`go run ./cmd/kbcheck graph-routing-eval --require-ready`

## Scope Boundary

No paid/model calls, performance claims from deterministic fixtures, external
provider installation, or broad benchmark campaign.

## Dependencies

Slice 006 supplies the end-to-end lifecycle being falsified.

## Concurrency

Serial against the integration branch. Race behavior is exercised only inside
temporary repositories controlled by the test harness.
