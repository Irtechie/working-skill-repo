---
kb_id: kb-2026-07-19-evidence-graph-routing
slice_id: slice-006
title: "Carry impact and isolation contracts through KB lifecycle"
blockers: [slice-004, slice-005]
verification: integration
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Changes semantics across synchronized routing, planning, work, review, and manifest validation while preserving old manifests and consent/proof gates."
model_requirements: ["cross-skill lifecycle architecture", "manifest compatibility", "impact reconciliation", "coordinator/worker state separation"]
escalation_triggers: ["old manifests fail", "workers overwrite canonical lifecycle files", "expected_files becomes an allowlist", "graph evidence substitutes for source or functional proof"]
context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-006.json
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck graph-routing-lifecycle-selftest"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-map/SKILL.md
    op: edit
    scope: "Return recipe, packet, freshness, and fallback summary."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Record impact forecast, isolation intent, conflict domains, and integration dependencies."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Consume packet, acquire claim, reconcile observed impact, and separate coordinator/worker state."
  - path: .github/skills/kb-review/SKILL.md
    op: edit
    scope: "Review missed consumers/tests/docs and unexplained impact expansion."
  - path: .github/skills/kb-review/references/diff-scope.md
    op: edit
    scope: "Compare actual diff with file and symbol impact forecasts."
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "Validate impact/isolation fields while keeping legacy manifests readable."
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: "Positive, negative, and compatibility fixtures."
  - path: cmd/kbcheck/graph_routing_lifecycle.go
    op: create
    scope: "Deterministic plan/work/review lifecycle selftest."
  - path: evals/skill-eval/selftest/pass-evidence-graph-routing.json
    op: create
    scope: "Protected end-to-end skill routing trace."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Durable packet and workspace lifecycle."
protected_oracles:
  - path: evals/skill-eval/selftest/pass-evidence-graph-routing.json
    role: "end-to-end lifecycle behavior oracle"
    sha256: "80B306772B62E1035BD39DB36D4239A498E37198830F39AAFCAE63A872ACF786"
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Proceed to slice 007 routing correctness and concurrency evals."
human_action: ""
can_continue_other_slices: false
---

# Slice 006: KB Lifecycle Integration

## What To Build

Make the impact packet and workspace contract operational across KB:

- `kb-map` returns seeds, evidence, impact, limits, freshness, and fallback.
- `kb-plan` records impact forecast, conflict domains, workspace mode, and
  integration dependencies without machine-specific paths.
- `kb-work` validates freshness, acquires a claim, resolves live workspace,
  reconciles discovered files/symbols, and emits a receipt.
- `kb-review` checks for missed consumers/tests/docs and unexplained scope growth.
- One coordinator projects worker receipts into the canonical board/manifest.

## Why This Slice Exists

A better index changes nothing if later agents still plan from guessed files,
edit without ownership, or review only the final diff. This slice closes the
loop from orientation to verification.

## Acceptance Criteria

- New manifests may declare impact/isolation fields; legacy manifests remain readable.
- Missing/stale packet blocks graph-dependent claims but permits explicit file-native fallback.
- `expected_files` stays a forecast and observed evidence may expand it with explanation.
- A worker receipt cannot mark a slice done; the coordinator verifies integration and proof first.
- Canonical lifecycle files have one writer during parallel work.
- Review fails an unexplained impacted consumer, test, doc, or conflict domain.
- Routing/lease receipts remain evidence and cannot replace functional proof.
- Pre-edit drift for all four changed skills is merged before edits.

## Test Scenarios

- Valid packet flows map -> plan -> work -> review.
- Stale packet downgrades to source inspection and records the decision.
- Discovered impacted file is admitted with provenance; unrelated cleanup is rejected.
- Two worker receipts cannot independently complete the same slice.
- Legacy manifest executes serially without new optional fields.

## Proof Check

`go run ./cmd/kbcheck graph-routing-lifecycle-selftest`

## Scope Boundary

No automatic provider installation, model routing changes, publish/PR delivery,
or work outside this skill bundle and required global sync targets.

## Dependencies

Slices 004 and 005 provide the exact and structural/flow evidence paths being
wired into lifecycle behavior.

## Concurrency

Serial integration slice. One coordinator edits all shared skills and lifecycle
fixtures after the parallel adapter branches are integrated.
