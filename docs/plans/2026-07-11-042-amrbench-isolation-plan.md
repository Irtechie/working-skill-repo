---
kb_id: kb-2026-07-11-ghcp-aic-falsification
slice_id: slice-002
title: "Contain model edits and proof around qualified fixtures"
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
expected_files:
  - path: cmd/amrbench/main.go
    op: edit
    scope: "Replace prompt-only safety with a no-paid state machine, immutable proof closure, mutable allowlist, and invalid outcomes."
  - path: cmd/amrbench/main_test.go
    op: edit
    scope: "Prove test mutation, new TestMain, dependency redirect, route mismatch, missing telemetry, and budget failure stop before another phase."
  - path: cmd/amrbench/runner.go
    op: create
    scope: "Separate live runner construction from a DisabledRunner that cannot load profiles, secrets, or spawn provider processes."
  - path: cmd/amrbench/runner_test.go
    op: create
    scope: "Prove no-paid conformance uses DisabledRunner and hostile configs cause zero spawn/profile/secret access."
  - path: cmd/amrbench/process_windows.go
    op: edit
    scope: "Enforce model and proof process-tree/time/resource containment or report unsupported before paid execution."
  - path: cmd/amrbench/process_unix.go
    op: edit
    scope: "Enforce portable process and network/filesystem containment or report unsupported before paid execution."
  - path: cmd/amrbench/sandbox.go
    op: create
    scope: "Separate model-write and proof-execution sandbox contracts with minimal environment and provider-only/no-network policies."
  - path: cmd/amrbench/sandbox_test.go
    op: create
    scope: "Prove unsupported containment is non-ready, nonzero under require-ready, and cannot start a paid phase."
  - path: cmd/amrbench/fixture.go
    op: create
    scope: "Canonical fixture containment, proof-dependency closure, mutable/new-file allowlist, and red-green-negative qualification."
  - path: cmd/amrbench/fixture_test.go
    op: create
    scope: "Prove proof-closure mutation, new oracle files, weak baseline, missing green solution, and negative mutations fail qualification."
  - path: evals/amr-model-benchmark/config.json
    op: edit
    scope: "Remove volatile hosted model ladder; retain abstract cohorts, qualified tasks, proof closure, and admission metadata; park HTML until browser sandbox proof exists."
  - path: .github/workflows/cross-platform.yml
    op: edit
    scope: "Run native Linux amrbench containment tests and compile Darwin/Windows platform implementations."
protected_oracles:
  - path: cmd/amrbench/main_test.go
    role: "no-paid isolation, oracle closure, route mismatch, and budget state-machine oracle"
    sha256: "filled after RED before implementation"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
can_continue_other_slices: false
---

# Slice 002 - Isolation and Fixture Authority

## Acceptance

- Model writes cannot touch proof inputs or files outside an explicit allowlist.
- Generated code proof runs without network/host filesystem access, with a
  minimal environment and process/time/resource bounds.
- Unsupported isolation returns `invalid-isolation` before a paid phase.
- `conformance --no-paid` always uses DisabledRunner. `--require-ready` returns
  nonzero when any required platform/sandbox/fixture is unsupported or failed;
  plain conformance may report honest `ready: false` without claiming canary readiness.
- Complete proof closure and red-fail/green-pass/negative-mutation preflight are
  hash-bound; subjective tasks are unrunnable in every scoring arm.
- HTML/UI stays ineligible in this milestone; rendered browser/a11y containment
  is a later slice, not simulated by substring tests.
- Paid state is explicit and attended, with per-call/arm/experiment ceilings and
  stop-before-correction on any invalid attempt.
- No model call is part of this slice.

## Platform Proof

- Native Windows package and malicious containment tests.
- Native Linux CI containment tests.
- Darwin and alternate-OS compile proof.
- Any unproved platform reports `unsupported` and cannot satisfy
  `--require-ready`.

## Internal Checkpoints

1. DisabledRunner and paid-state RED/GREEN.
2. Model/proof sandbox and cross-platform compile/native proof.
3. Fixture closure plus red/green/negative qualification.
4. No-paid readiness integration; no live runner constructed.

## Scope Boundary

No production correction runner, live canary, context winner, or route promotion.
