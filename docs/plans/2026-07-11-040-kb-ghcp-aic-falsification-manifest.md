---
type: kb-manifest
kb_id: kb-2026-07-11-ghcp-aic-falsification
brainstorm_path: docs/brainstorms/2026-07-11-ghcp-aic-amr-falsification-requirements.md
created: 2026-07-11
status: reviewed
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
scope-verified-files:
  - cmd/amrbench/context.go
  - cmd/amrbench/context_test.go
  - cmd/amrbench/approval.go
  - cmd/amrbench/approval_test.go
  - cmd/amrbench/atomic_write.go
  - cmd/amrbench/atomic_write_unix.go
  - cmd/amrbench/atomic_write_windows.go
  - cmd/amrbench/bounded_writer.go
  - cmd/amrbench/fixture.go
  - cmd/amrbench/fixture_test.go
  - cmd/amrbench/grade.go
  - cmd/amrbench/grade_artifact_test.go
  - cmd/amrbench/grade_test.go
  - cmd/amrbench/main.go
  - cmd/amrbench/main_test.go
  - cmd/amrbench/runner.go
  - cmd/amrbench/runner_test.go
  - cmd/amrbench/sandbox.go
  - cmd/amrbench/sandbox_test.go
  - cmd/kbcheck/execution_telemetry.go
  - cmd/kbcheck/execution_telemetry_test.go
  - cmd/kbcheck/model_routing_ghcp_release.go
  - cmd/kbcheck/model_routing_ghcp_release_test.go
  - cmd/kbcheck/model_routing_release.go
  - docs/context/eval-map.md
  - docs/context/PROJECT.md
  - docs/context/architecture/README.md
  - docs/context/architecture/ghcp-aic-benchmark.md
  - docs/context/goals/ghcp-aic-falsification.md
  - docs/context/kb/instincts/scoped/model-routing/evaluation.yaml
  - docs/context/kb/kb-completions.txt
  - docs/context/memory-maintenance.md
  - docs/context/operations/testing.md
  - docs/handoffs/active/2026-07-11-ghcp-aic-no-paid-readiness.md
  - docs/handoffs/done/2026-07-11-ghcp-aic-snapshot-blocker.md
  - docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md
  - docs/plans/2026-07-11-041-ghcp-otel-accounting-plan.md
  - docs/plans/2026-07-11-042-amrbench-isolation-plan.md
  - docs/plans/2026-07-11-043-context-diet-ab-plan.md
  - docs/plans/2026-07-11-044-paired-amr-falsification-plan.md
  - docs/results/2026-07-11-ghcp-aic-plan-to-work-proof.md
  - docs/results/2026-07-12-ghcp-follow-on-no-paid.json
  - docs/results/2026-07-12-ghcp-aic-attended-preview.md
  - docs/results/2026-07-12-ghcp-aic-finalization-proof.md
  - docs/solutions/logic-errors/proof-spine-digest-check-semantics-2026-07-05.md
  - evals/amr-model-benchmark/README.md
  - evals/amr-model-benchmark/config.json
  - evals/amr-model-benchmark/context-contract.schema.json
  - evals/amr-model-benchmark/context-contracts/baseline.json
  - evals/amr-model-benchmark/context-contracts/minimal.json
  - evals/amr-model-benchmark/context-contracts/artifacts/baseline-packet.txt
  - evals/amr-model-benchmark/context-contracts/artifacts/minimal-packet.txt
  - evals/amr-model-benchmark/context-contracts/artifacts/reviewer-overlay.txt
  - evals/amr-model-benchmark/context-contracts/artifacts/worker-overlay.txt
  - evals/amr-model-benchmark/qualification/canonical-cache-key/go.mod
  - evals/amr-model-benchmark/qualification/canonical-cache-key/key/query.go
  - evals/amr-model-benchmark/qualification/canonical-cache-key/request/cache.go
  - evals/amr-model-benchmark/qualification/retry-after-parser/go.mod
  - evals/amr-model-benchmark/qualification/retry-after-parser/retry/retry.go
  - internal/ghcpotel/parser.go
  - internal/ghcpotel/parser_test.go
  - internal/ghcpotel/parser_graph_test.go
  - internal/ghcpotel/redact.go
  - internal/ghcpotel/redact_test.go
  - internal/ghcpotel/testdata/schema-probe-redacted.jsonl
  - internal/modelrouting/storage_acl_windows.go
  - todo.md
  - todo-done.md
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
      - docs/brainstorms/2026-07-11-ghcp-aic-amr-falsification-requirements.md
      - docs/plans/2026-07-10-030-kb-session-model-routing-manifest.md
      - docs/results/2026-07-11-ghcp-aic-plan-to-work-proof.md
    blockers: []
    passed_at: "2026-07-11T23:48:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "strict exact leaf-call accounting is implemented"
      - "protected oracle is preserved"
      - "focused and package checks pass"
      - "scope, QA, and regression snapshot gates pass"
      - "no paid or model call ran"
    proof:
      - internal/ghcpotel/parser.go
      - internal/ghcpotel/parser_test.go
      - internal/ghcpotel/redact.go
      - internal/ghcpotel/redact_test.go
      - internal/ghcpotel/testdata/schema-probe-redacted.jsonl
      - .kb/runs/ghcp-aic/slice-001-result.md
      - .kb/snapshots/ghcp-aic-slice-001.json
    blockers: []
    passed_at: "2026-07-11T23:59:00-04:00"
    allowed_next_action: "kb-work slice-002"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "DisabledRunner makes no-paid paths unable to start providers"
      - "network-disabled Podman proof containment is ready"
      - "fixtures prove RED, GREEN, and negative sensitivity with protected closure"
      - "budget, scope, QA, platform, and regression gates pass"
      - "no paid or model call ran"
    proof:
      - cmd/amrbench/runner.go
      - cmd/amrbench/runner_test.go
      - cmd/amrbench/sandbox.go
      - cmd/amrbench/sandbox_test.go
      - cmd/amrbench/fixture.go
      - cmd/amrbench/fixture_test.go
      - evals/amr-model-benchmark/config.json
      - .kb/runs/ghcp-aic/slice-002-result.md
      - .kb/snapshots/ghcp-aic-slice-002.json
    blockers: []
    passed_at: "2026-07-12T02:20:00-04:00"
    allowed_next_action: "kb-work slice-003"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "development and held-out routing corpora are disjoint"
      - "baseline and minimal controls differ only in ambient context"
      - "crossover orders and role overlays are frozen"
      - "winner rule rejects correctness or weak-savings regressions"
      - "integration, scope, QA, and regression gates pass without paid calls"
    proof:
      - cmd/amrbench/context.go
      - cmd/amrbench/context_test.go
      - evals/amr-model-benchmark/context-contract.schema.json
      - evals/amr-model-benchmark/context-contracts/baseline.json
      - evals/amr-model-benchmark/context-contracts/minimal.json
      - .kb/runs/ghcp-aic/slice-003-result.md
      - .kb/snapshots/ghcp-aic-slice-003.json
    blockers: []
    passed_at: "2026-07-12T02:35:00-04:00"
    allowed_next_action: "kb-work slice-004"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "direct and AMR evidence is paired by task, seed, and family"
      - "invalid accounting, route, oracle, isolation, or proof evidence is rejected"
      - "promotion gates require correctness, aggregate and median AIU, confidence, and intervention safety"
      - "GHCP follow-on validation preserves the initial cohort and no-paid evidence stays not-promoted"
      - "full tests, core, QA, done check, scope, and regression snapshot pass"
    proof:
      - cmd/amrbench/grade.go
      - cmd/amrbench/grade_test.go
      - cmd/kbcheck/model_routing_ghcp_release.go
      - cmd/kbcheck/model_routing_ghcp_release_test.go
      - docs/results/2026-07-12-ghcp-follow-on-no-paid.json
      - .kb/runs/ghcp-aic/slice-004-result.md
      - .kb/snapshots/ghcp-aic-slice-004.json
    blockers: []
    passed_at: "2026-07-12T03:25:00-04:00"
    allowed_next_action: "kb-work completion"
  - gate_id: work-to-complete
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "all four slice-to-done gates pass"
      - "terminal no-paid done check exits zero"
      - "full Go tests and core gate pass"
      - "scope-verified-files is populated"
      - "board and goal state point to post-work completion"
      - "no unresolved P0/P1 or paid call exists"
    proof:
      - docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md
      - .kb/runs/ghcp-aic/slice-001-result.md
      - .kb/runs/ghcp-aic/slice-002-result.md
      - .kb/runs/ghcp-aic/slice-003-result.md
      - .kb/runs/ghcp-aic/slice-004-result.md
      - docs/results/2026-07-12-ghcp-follow-on-no-paid.json
    blockers: []
    passed_at: "2026-07-12T03:25:00-04:00"
    allowed_next_action: "kb-complete docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md"
  - gate_id: complete-to-ship
    owner_skill: kb-complete
    status: passed
    required_evidence:
      - "final deterministic and objective checks pass"
      - "functional CLI and dry-run preview are proven with zero paid calls"
      - "multi-agent review has no unresolved reachable P0/P1"
      - "follow-up resolution and proof rerun complete"
      - "compound, learning, memory refresh, maintenance, and cleanup complete"
      - "attended preview is recorded and paid execution remains blocked"
    proof:
      - docs/results/2026-07-12-ghcp-aic-finalization-proof.md
      - docs/results/2026-07-12-ghcp-aic-attended-preview.md
      - docs/results/2026-07-12-ghcp-follow-on-no-paid.json
      - docs/context/architecture/ghcp-aic-benchmark.md
      - docs/context/kb/instincts/scoped/model-routing/evaluation.yaml
      - docs/solutions/logic-errors/proof-spine-digest-check-semantics-2026-07-05.md
    blockers: []
    passed_at: "2026-07-12T11:45:00Z"
    allowed_next_action: "human approval after trusted approval verifier and route availability"
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
    status: done
    owner: agent
    can_continue_other_slices: false
    notes: "scope-forecast: loaded 9 expected files + 0 convention-matched tests; scope-discovery: docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md - workflow gate/status; scope-discovery: docs/plans/2026-07-11-041-ghcp-otel-accounting-plan.md - protected oracle hash and amendment; scope-discovery: todo.md - board sync; scope-discovery: docs/results/2026-07-11-ghcp-aic-plan-to-work-proof.md - repaired gate evidence; scope-discovery: docs/handoffs/done/2026-07-11-ghcp-aic-snapshot-blocker.md - resolved pre-slice blocker; scope-check: forecast=9 changed=14 discovered=5 unexplained=0; TDD: RED undefined Parse/RedactSchema then GREEN strict trace-span accounting and redaction; protected-oracle: SHA256 35b9967ffbdaaad945435c470d79955bab798bb369898eb34b82d3dc0f5bf666 preserved; proof: focused and full affected package tests PASS; qa-lint: gofmt and go vet PASS; qa-browser: skipped - no UI-reachable behavior changed; snapshot: .kb/snapshots/ghcp-aic-slice-001.json PASS; memory-impact: durable; areas=GHCP exact accounting; docs=eval map candidate; refresh=pending; paid-calls=0"
    protected_oracles:
      - path: internal/ghcpotel/parser_test.go
        role: "strict complete leaf-call accounting and schema-drift oracle"
        sha256: "35b9967ffbdaaad945435c470d79955bab798bb369898eb34b82d3dc0f5bf666"
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
    status: done
    owner: agent
    can_continue_other_slices: false
    notes: "scope-forecast: loaded 12 expected files + 0 convention-matched tests; scope-discovery: evals/amr-model-benchmark/qualification/retry-after-parser/retry/retry.go - protected known solution used only by qualification; scope-discovery: evals/amr-model-benchmark/qualification/canonical-cache-key/key/query.go - protected known solution; scope-discovery: evals/amr-model-benchmark/qualification/canonical-cache-key/request/cache.go - protected known solution; scope-discovery: internal/modelrouting/storage_acl_windows.go - repaired reproducible prerequisite snapshot failure caused by concurrent ACL descriptor growth; scope-discovery: docs/context/goals/ghcp-aic-falsification.md - durable goal ledger; scope-discovery: docs/handoffs/active/2026-07-11-ghcp-aic-no-paid-readiness.md - Podman cleanup and restart state; scope-check: forecast=12 changed=18 discovered=6 unexplained=0; TDD: RED missing DisabledRunner/containment/fixture/budget contracts then GREEN; protected-oracle: SHA256 8ac3d397ab5ec204196880bc0f3fc216268e2f8fba6e80d85fd0031dc297a1ab preserved; proof: focused/full package tests PASS; functional-cli: conformance --no-paid --require-ready ready=true runner=disabled paid_calls=0; fixture proof: retry-after-parser and canonical-cache-key baseline RED, solution GREEN, negative RED; platform: Windows native PASS, Linux/Darwin compile PASS; qa-lint: gofmt/go vet/diff-check PASS; qa-browser: skipped - no UI-reachable behavior changed; snapshot: .kb/snapshots/ghcp-aic-slice-002.json PASS; memory-impact: durable; areas=no-paid isolation and fixture authority; docs=eval map and active handoff; refresh=pending; paid-calls=0"
    protected_oracles:
      - path: cmd/amrbench/main_test.go
        role: "no-paid isolation, oracle closure, route mismatch, and budget state-machine oracle"
        sha256: "8ac3d397ab5ec204196880bc0f3fc216268e2f8fba6e80d85fd0031dc297a1ab"
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
    status: done
    owner: agent
    can_continue_other_slices: false
    notes: "scope-forecast: loaded 5 expected files + 0 convention-matched tests; scope-discovery: cmd/amrbench/main.go - public no-paid conformance must load frozen contracts; scope-discovery: evals/amr-model-benchmark/config.json - binds baseline/minimal contracts into readiness; scope-discovery: docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md - status/gate; scope-discovery: docs/plans/2026-07-11-043-context-diet-ab-plan.md - protected hashes; scope-discovery: todo.md and docs/context/goals/ghcp-aic-falsification.md - board/goal sync; scope-check: forecast=5 changed=11 discovered=6 unexplained=0; RED undefined context contract functions then GREEN; protected-oracles: context_test dc11a98ba48fe6a270f35536f2792059ee8873703fb460e1a315b3c330cc0538 and schema f7d569516ea4a348d58f4afecb8a6a4855929512c2c0894e01d77dd71c853998 preserved; proof: context/contract/crossover/winner/holdout tests PASS; functional-integration: public no-paid conformance validates contracts and remains ready with paid_calls=0; qa-lint: gofmt/go vet/diff-check PASS; qa-browser: skipped - no UI-reachable behavior changed; snapshot: .kb/snapshots/ghcp-aic-slice-003.json PASS; memory-impact: durable; areas=context A-B contract; docs=eval map; refresh=pending; paid-calls=0"
    protected_oracles:
      - path: cmd/amrbench/context_test.go
        role: "corpus separation, comparison parity, and winner-rule behavior oracle"
        sha256: "dc11a98ba48fe6a270f35536f2792059ee8873703fb460e1a315b3c330cc0538"
        update_policy: "requires explicit plan update"
      - path: evals/amr-model-benchmark/context-contract.schema.json
        role: "predeclared disjoint context A-B contract and winner-rule oracle"
        sha256: "adc08ddc31d61a4930b91f8a2547605b3a222f9b3b12253fe1b0cfd7af1dea25"
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
    status: done
    owner: agent
    can_continue_other_slices: false
    notes: "scope-forecast: loaded 8 expected files + 0 convention-matched tests; scope-discovery: cmd/amrbench/runner.go and runner_test.go - runtime-only route resolution and three-level credit budget; scope-discovery: docs/results/2026-07-12-ghcp-follow-on-no-paid.json - deterministic follow-on evidence; scope-discovery: evals/amr-model-benchmark/qualification/*/go.mod - isolate known solutions from root module; scope-discovery: evals/amr-model-benchmark/qualification/*/*.go - known solutions required by no-paid fixture qualification; scope-discovery: docs/context/goals/ghcp-aic-falsification.md and active handoff - goal/cleanup state; scope-discovery: internal/modelrouting/storage_acl_windows.go - prerequisite snapshot repair; scope-check: forecast=8 changed=22 discovered=14 unexplained=0; RED undefined paired/GHCP validators then GREEN; protected-oracles: grade_test 3c5300b4d3fbaf9081eaaf024a035ba0117e2dd8d7fe64b40537e8578115e3db and GHCP release test d3f3251d48adf3b1ca4ec87ad79cf52453c56a51260534deb8b330ff04fc195f preserved; proof: paired/conformance/promotion/GHCP focused tests PASS; functional-cli: grade-paired separates strong/weak families and follow-on validator reports not-promoted; full go test and core 33/33 PASS; done-check: conformance ready=true runner=disabled paid_calls=0 release_decision=not-promoted; qa-lint: gofmt/go vet/diff-check PASS; qa-browser: skipped - no UI-reachable behavior changed; snapshot: .kb/snapshots/ghcp-aic-slice-004.json PASS; memory-impact: durable; areas=paired grader, budget, GHCP follow-on release; docs=README and eval map updated; refresh=pending; paid-calls=0"
    protected_oracles:
      - path: cmd/amrbench/grade_test.go
        role: "paired correctness, all-inclusive AIU, family admission, and not-promoted oracle"
        sha256: "3c5300b4d3fbaf9081eaaf024a035ba0117e2dd8d7fe64b40537e8578115e3db"
        update_policy: "requires explicit plan update"
      - path: cmd/kbcheck/model_routing_ghcp_release_test.go
        role: "GHCP follow-on promotion oracle that preserves the landed Codex-first cohort"
        sha256: "60acbdedfc079d59c94997c3ff4b70d487cecfdebfa846fdcdddbbabfffa8dfb"
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
| 1 | GHCP OTel accounting | - | tdd | medium | done |
| 2 | Fixture and process isolation | 1 | tdd | large | done |
| 3 | Predeclare context-diet A/B | 2 | functional | medium | done |
| 4 | Paired full-fallback grader | 3 | functional | large | done |
