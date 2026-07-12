---
type: kb-manifest
kb_id: kb-2026-07-11-ghcp-aic-falsification
brainstorm_path: docs/brainstorms/2026-07-11-ghcp-aic-amr-falsification-requirements.md
created: 2026-07-11
status: planned
workflow_shape: pipeline-change
objective_contract: true
preconditions:
  - "docs/plans/2026-07-10-030-kb-session-model-routing-manifest.md reaches reviewed/landed slice-007 baseline"
done_check:
  kind: command_exit
  command: "go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json"
  expect: 0
  why: "Proves the benchmark is safe, paired, exact-accounting, and ready for an explicitly approved attended canary without spending AI credits."
model_tier_contract:
  allowed: [small, medium, large]
  default: medium
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-brainstorm
    status: passed
    required_evidence:
      - "requirements and question gate are complete"
      - "claims are separated into verified, provisional, and parked"
      - "security and coherence document review have no unresolved P0/P1"
    proof:
      - docs/brainstorms/2026-07-11-ghcp-aic-amr-falsification-requirements.md
      - "live invalid canary and official GHCP OTel evidence"
      - "review: repo-critic keep-consolidate-delete audit"
      - "review: final security and coherence reviews clear"
    blockers: []
    passed_at: "2026-07-11T06:37:11-04:00"
    allowed_next_action: "kb-plan"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "four vertical slices have bounded packets, proof, risks, tiers, and no route hints"
      - "DAG is valid"
      - "incoming amrbench draft is classified as RED/consolidate, not accepted implementation"
      - "no paid run is authorized"
      - "session-routing manifest slice-007 is done, complete-to-ship passed/quarantined, and its landed commit is recorded"
    proof:
      - docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md
      - docs/plans/2026-07-11-041-ghcp-otel-accounting-plan.md
      - docs/plans/2026-07-11-042-amrbench-isolation-plan.md
      - docs/plans/2026-07-11-043-context-diet-ab-plan.md
      - docs/plans/2026-07-11-044-paired-amr-falsification-plan.md
      - "validation: manifest contract and all four context packets pass"
      - "review: final security-feasibility plan review has no P0-P1"
      - "review: final coherence-DAG-tier plan review has no P0-P1"
      - "route contract: plans contain difficulty/proof only and authorize no paid run"
      - "landed baseline: PR #1 merged as 36736ec52258093cde1b898bd1710a7a6039061d with successful Windows, macOS, and Linux checks"
      - "remote-main ancestry: delivered topic HEAD 8f19f972c7106c071fda9fecf157d497d147a7cb is contained by origin/main"
      - "legacy completion metadata quarantined: the older baseline manifest predates complete-to-ship, while its slice-007 local-release proof, merged PR checks, and remote-main ancestry independently prove delivery"
    blockers: []
    passed_at: "2026-07-11T22:48:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md"
slices:
  - id: slice-001
    title: "Normalize complete GHCP leaf-call accounting"
    path: docs/plans/2026-07-11-041-ghcp-otel-accounting-plan.md
    blockers: []
    external_blockers:
      - "session-routing slice-007 landed baseline"
    verification: tdd
    test_level: unit
    functional_risk: broad
    model_tier: medium
    context_packet_path: docs/plans/2026-07-11-ghcp-aic-context/slice-001.json
    proof_check:
      kind: command_exit
      command: "go test ./internal/ghcpotel ./cmd/amrbench ./cmd/kbcheck -run 'OTel|AIU|Span|Telemetry|Mismatch' -count=1"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: internal/ghcpotel/parser_test.go
        role: "strict complete leaf-call accounting and schema-drift oracle"
        sha256: "filled after RED before implementation"
        update_policy: "requires explicit plan update"
  - id: slice-002
    title: "Contain model edits and proof around qualified fixtures"
    path: docs/plans/2026-07-11-042-amrbench-isolation-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: functional-cli
    functional_risk: full
    model_tier: large
    context_packet_path: docs/plans/2026-07-11-ghcp-aic-context/slice-002.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/amrbench -run 'Isolation|Oracle|Fixture|Budget|Containment|InvalidRoute' -count=1"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: cmd/amrbench/main_test.go
        role: "no-paid isolation, oracle closure, route mismatch, and budget state-machine oracle"
        sha256: "filled after RED before implementation"
        update_policy: "requires explicit plan update"
  - id: slice-003
    title: "Predeclare the context-diet A-B contract"
    path: docs/plans/2026-07-11-043-context-diet-ab-plan.md
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
    status: pending
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: cmd/amrbench/context_test.go
        role: "corpus separation, comparison parity, and winner-rule behavior oracle"
        sha256: "filled after RED before implementation"
        update_policy: "requires explicit plan update"
      - path: evals/amr-model-benchmark/context-contract.schema.json
        role: "predeclared disjoint context A-B contract and winner-rule oracle"
        sha256: "filled after RED before implementation"
        update_policy: "requires explicit plan update"
  - id: slice-004
    title: "Grade paired full-fallback AMR without promotion theater"
    path: docs/plans/2026-07-11-044-paired-amr-falsification-plan.md
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
    status: pending
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: cmd/amrbench/grade_test.go
        role: "paired correctness, all-inclusive AIU, family admission, and not-promoted oracle"
        sha256: "filled after RED before implementation"
        update_policy: "requires explicit plan update"
      - path: cmd/kbcheck/model_routing_ghcp_release_test.go
        role: "GHCP follow-on promotion oracle that preserves the landed Codex-first cohort"
        sha256: "filled after RED before implementation"
        update_policy: "requires explicit plan update"
---

# GHCP AIC and AMR Falsification

## Outcome

Consolidate the incoming `amrbench` draft into a no-paid conformance harness
that can later run an explicitly approved GHCP canary. The harness must reject
route mismatch, incomplete accounting, mutable proof inputs, weak fixtures, and
unsupported containment before any model call. It measures full-fallback AMR
economics only; production partial reuse remains parked.

## Model Selection Contract

Plans record only difficulty, constraints, proof, and risk. Live worker choice
happens immediately before execution. No model name, route alias, provider, or
transport is durable plan advice.

## Incoming Draft Status

`cmd/amrbench/` and `evals/amr-model-benchmark/` are RED/consolidate inputs, not
accepted implementation. Keep the bounded-fixture concept, process-tree work,
leaf-chat dedup, actual-model capture, user-local profiles, and final proof.
Rebuild prompt-only oracle protection, float accounting, mismatched attribution,
unpaired grading, weak fixtures, and executable-correction claims.

No further paid run is authorized by this manifest. `hitl: false` means every
slice stops at deterministic `--no-paid` readiness. A separate attended approval
is required before any canary, context A/B, or paired matrix.

## Slice Overview

| # | Slice | Blocked By | Verification | Tier | Status |
|---|---|---|---|---|---|
| 1 | GHCP OTel accounting | - | tdd | medium | pending |
| 2 | Fixture and process isolation | 1 | tdd | large | pending |
| 3 | Predeclare context-diet A/B | 2 | functional | medium | pending |
| 4 | Paired full-fallback grader | 3 | functional | large | pending |
