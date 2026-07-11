---
kb_id: kb-2026-07-10-session-model-routing
slice_id: slice-005a
title: "Consolidate AMR into one proof-triggered attempt and correction loop"
blockers: [slice-005]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-005a.json
proof_check:
  kind: command_exit
  command: "go test ./internal/modelrouting ./cmd/kbrouter ./cmd/kbcheck"
  expect: 0
hitl: false
expected_files:
  - path: internal/modelrouting/selector.go
    op: edit
    scope: "Represent planned correction tier separately from an optional bounded lower-tier attempt."
  - path: internal/modelrouting/selector_test.go
    op: edit
    scope: "Replace the no-downward oracle with attempt-tier, trust, override, and escalation cases."
  - path: cmd/kbrouter/select.go
    op: edit
    scope: "Expose planned tier and optional attempt tier without inferring task suitability."
  - path: cmd/kbrouter/select_test.go
    op: edit
    scope: "Prove CLI attempt selection and planned-tier fallback metadata."
  - path: cmd/kbrouter/catalog_test.go
    op: edit
    scope: "Keep the private-route selection fixture on the exact require-pin contract after use becomes a bounded preference."
  - path: cmd/kbrouter/dispatch.go
    op: edit
    scope: "Carry planned and attempt tiers through strict dispatch packets without re-inferring eligibility."
  - path: cmd/kbrouter/dispatch_test.go
    op: edit
    scope: "Prove a selected lower-tier attempt survives actual dispatch and preserves correction-tier metadata."
  - path: internal/modelrouting/receipt.go
    op: edit
    scope: "Link attempt and correction phases to proof evidence without turning route telemetry into work proof."
  - path: internal/modelrouting/correction.go
    op: create
    scope: "Validate the minimal pilot correction packet and hunk-local acceptance boundary."
  - path: internal/modelrouting/correction_test.go
    op: create
    scope: "Prove unaccepted hunks, unrelated edits, worker-supplied authority, and unlocalizable broadening fail closed."
  - path: cmd/kbcheck/context_packet.go
    op: edit
    scope: "Keep planner difficulty authoritative and validate the bounded evidence AMR needs for an attempt."
  - path: cmd/kbcheck/context_packet_test.go
    op: edit
    scope: "Prove valid bounded-attempt and driver-only packet shapes."
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "Remove DDR-only self-authored manifest validation."
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: "Replace DDR receipt tests with observable AMR attempt/escalation contract tests."
  - path: cmd/kbcheck/swarm.go
    op: edit
    scope: "Remove DDR-only slice fields while preserving generic execution telemetry."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Define planned tier as correction authority and keep execution policy out of planning."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Own one lower-tier attempt, proof, surgical correction, and escalation loop; delete DDR duplication."
  - path: .github/skills/kb-work/references/execution-prompt.md
    op: edit
    scope: "Use one attempt/correction packet and preserve accepted hunks."
  - path: .github/skills/kb-functional-test/SKILL.md
    op: edit
    scope: "Own proof-level classification without contradicting AMR implementation attempts."
  - path: .github/skills/kb-configure/SKILL.md
    op: edit
    scope: "Delete DDR modes and retain delivery plus a simple lower-tier-attempt opt-out."
  - path: .github/skills/kb-models/SKILL.md
    op: edit
    scope: "Document beginner-default and advanced run overrides for the one AMR loop."
  - path: .github/skills/kb-map/SKILL.md
    op: edit
    scope: "During explicit project setup only, offer an optional handoff to kb-models without collecting connection details in project memory."
  - path: docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md
    op: edit
    scope: "Replace the hard tier floor with planned correction tier and bounded attempt policy."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Document one AMR loop and remove DDR as a separate subsystem."
  - path: README.md
    op: edit
    scope: "Explain the novice and advanced AMR surfaces without DDR ceremony."
  - path: evals/skill-eval/selftest/pass-session-model-routing.json
    op: edit
    scope: "Protect the one-loop skill behavior and proof authority."
  - path: .github/skills/kb-configure/references/kb-routing-example.yaml
    op: edit
    scope: "Remove DDR configuration and show lower-tier-attempt plus delivery policy."
protected_oracles:
  - path: internal/modelrouting/selector_test.go
    role: "attempt, correction-tier, trust, override, and fallback oracle"
    sha256: "ec77785683c690399d7eadbfc65cdb81768d87743406d03e647c0dace01dd7eb"
    update_policy: "this plan explicitly authorizes replacing the obsolete no-downward assertions"
  - path: internal/modelrouting/correction_test.go
    role: "driver-owned correction scope, preserved-hunk, and exact-proof oracle"
    sha256: "5f6cd0365c639f894be0c0f210d5410ecb32f2e0b23c508cbc239edb82226a59"
    update_policy: "requires explicit plan update"
status: done
owner: agent
can_continue_other_slices: false
---

# Slice 005a — One AMR Loop

## Outcome

The planner records difficulty and proof. At work time, AMR may try a lower-tier
worker only for a bounded, objectively provable packet. Passing work continues
without a stronger-model rewrite. Failure produces a surgical correction packet
for the planned tier. The current CLI fails closed before correction dispatch:
an isolated workspace, host-owned proof runner, and compare-and-swap apply path
are required before accepted hunks can be preserved safely. Failed pilots use
separate ordinary planned-tier execution and claim no savings. DDR no longer exists as a
separate planner, configuration, receipt, or workflow concept.

## Acceptance Criteria

1. Planned tier remains the required correction/authority tier; no model name is
   frozen in durable plans.
2. A bounded Medium code packet can explicitly request a Small attempt; selector
   output preserves the Medium correction tier and never infers suitability from
   “code” or a file extension alone.
3. Missing proof, sensitive/high-risk ambiguity, authority expansion, or an
   explicit opt-out starts at the planned tier.
4. Failed attempts hand the planned-tier model a compact correction contract:
   accepted result, failed criterion/location, allowed change, invariants,
   relevant interfaces, exact proof, current diff, and attempt ledger.
   A hunk is preserved only when an independent hunk-local oracle accepted it;
   otherwise the attempt is ineligible for the surgical pilot. Unlocalizable
   failures become separately measured ordinary planned-tier execution.
   Until isolated correction execution exists, the CLI refuses before worker
   launch or receipt creation and all failures use that ordinary path.
5. `use`, `require`, `prefer local|hosted`, and `ignore model routing` remain
   understandable advanced controls. New users receive no questionnaire and at
   most one compact attempt/fallback preview.
6. DDR skill/config/manifest language is removed or migrated without weakening
   ordinary proof, trust, scope, or delivery gates.
7. Runtime, CLI, manifest, skill-eval, core, and release/sync checks pass, except
   independently proven pre-existing failures recorded precisely.

## Test Scenarios

- Medium bounded request + explicit Small attempt selects proven Small routes
  first and reports Medium as correction tier.
- The same request without attempt eligibility begins at Medium.
- Small proof failure yields a surgical Medium correction handoff; current CLI
  execution refuses before launch and mints no acceptance receipt.
- A failing result with no independent hunk-local acceptance records zero
  accepted hunks and cannot claim surgical savings.
- Unsafe trust/tool/context/risk candidates remain ineligible at every tier.
- `require` unavailable pauses instead of silently substituting.
- Router unavailable degrades to current driver with ordinary proof.
- DDR-only manifest fields no longer create a cosmetic proof gate.

## Scope Boundary

No image regeneration, automatic global promotion, new model purchase, hidden
provider inference, or weakening of tests/review. Images remain gated until this
slice and the existing pilot proof pass.
