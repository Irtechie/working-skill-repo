---
kb_id: kb-2026-08-23-epistemic-investigation-gate
slice_id: epistemic-002
title: Freeze the corpus and capture the untouched baseline
blockers: [epistemic-001]
verification: functional
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: Baseline validity requires provenance, matched sampling, oracle isolation, and an honest inconclusive state.
model_requirements: [live adapter operation, provenance capture, bounded experiment design, cost telemetry, statistical restraint]
escalation_triggers: [treatment changes before baseline, hashes drift, runtime identity is missing, live preview lacks approval]
workspace_mode: shared-serial
conflict_domains: [namespace:skill-eval, path:.kb/eval-runs/epistemic-baseline, path:evals/skill-eval/baselines, path:docs/results]
shared_resources: [git:integration-owner, eval:live-runtime, budget:explicit-live-model]
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck skill-eval-regression --run-root .kb/eval-runs/epistemic-baseline --output evals/skill-eval/baselines/epistemic-investigation.json"
  expect: 0
hitl: true
expected_files:
  - path: cmd/kbcheck/eval_adapters.go
    op: edit
    scope: Make Codex epistemic preview and live modes require the same router-selected contract, runtime identity, sealed actor, and ambient skill-isolation inventory without coordinator model selection.
  - path: cmd/kbcheck/skill_eval_epistemic_test.go
    op: edit
    scope: Protect no-call preview parity for runtime identity, route-contract binding, sealed prompt hashing, and ambient instruction-isolation inventory.
  - path: cmd/kbrouter/dispatch.go
    op: edit
    scope: Add a narrow evaluator dispatch seam that accepts an external actor root, sealed prompt, epistemic schema, and instruction-isolation configuration while retaining route/session attribution.
  - path: cmd/kbrouter/dispatch_test.go
    op: edit
    scope: Prove routed evaluator dispatch preserves CWD, schema, structured output, actual-model attribution, and fail-closed route eligibility without inference in tests.
  - path: cmd/kbrouter/main.go
    op: edit
    scope: Document only the bounded dispatch flags required by the evaluator seam; do not add AMR or a static tier-to-model catalog.
  - path: internal/modelrouting/storage_acl_windows.go
    op: edit
    scope: Avoid a redundant Windows owner write when the protected run root is already current-user-owned while retaining owner transfer for transitional Administrator/System roots.
  - path: internal/modelrouting/storage_acl_windows_test.go
    op: edit
    scope: Protect current-owner DACL-only security updates and required transitional owner transfer.
  - path: readme.md
    op: edit
    scope: Remove the stale lower-tier-attempt/AMR-derived routing description while preserving the current harness-aware DDR diagram and every other image.
  - path: evals/skill-eval/epistemic/visible/
    op: edit
    scope: Finalize a balanced development and holdout corpus before baseline capture.
  - path: evals/skill-eval/epistemic/oracles/
    op: edit
    scope: Finalize deterministic labels and expected resolving evidence, then freeze the corpus hash.
  - path: evals/skill-eval/baselines/epistemic-investigation.json
    op: create
    scope: Persist the existing regression report as the untouched comparison baseline.
  - path: docs/results/epistemic-investigation-baseline.md
    op: create
    scope: Record runtime/model provenance, corpus/scorer hashes, separate metrics, missing telemetry, and limitations.
protected_oracles:
  - path: evals/skill-eval/epistemic/
    role: Frozen visible corpus and hidden labels.
    sha256: 9c506933962c01d8d82a78a1a57c500def570db44298f50393f2cf8853813537
    update_policy: Any change invalidates the baseline and requires recapture.
  - path: cmd/kbcheck/skill_eval_epistemic.go
    role: Frozen epistemic scorer.
    sha256: 40e330ff5d2b8c3b0a5149e5b434baa31335f470572695c69da7682e819e2c0f
    update_policy: Any semantic change invalidates baseline-treatment comparison.
test_inputs:
  - name: baseline_preview
    source: generated
    required_for: Exact live-run approval
    value: .kb/eval-runs/epistemic-baseline-preview.json
status: in_progress
owner: agent
blocked_reason: ""
resume_when: the user approves the exact Terra app-native preview at .kb/eval-runs/epistemic-baseline-preview.json
next_agent_action: On approval, run eight no-history Terra fixture actors with no retries or model substitution, persist agent-id receipts and exact results, then score the untouched baseline.
human_action: Approve or reject the exact eight-call Terra preview; gpt-5.5 is excluded from this run.
can_continue_other_slices: false
---

# Untouched Baseline

## Deliverable

A baseline produced by the existing adapter and regression command before any
investigation-behavior instruction changes.

## Acceptance Criteria

- Corpus contains adequately supported `proceed` cases and unsupported
  `investigate` cases leading to supported, contradicted, and unresolved final
  states.
- The holdout corpus and scorer are frozen before the first model call.
- The baseline receipt records content hashes or Git blob identities for the
  untreated `kb-plan` and `kb-gate` surfaces so the final control arm can be
  reconstructed after treatment without copying stale ambient state.
- Preview mode makes no model call and names the selected route, minimum tier,
  runtime, repetitions, hashes, and bounded cost when the host exposes cost.
  Provider-reported actual model identity is mandatory on the live receipt and
  is never invented in the pre-run preview.
- `model_tier` classifies execution capability; it is not a model name or a
  reasoning-effort setting. The orchestrator owns work-time DDR selection,
  inspects the live callable harness and eligible user-local routes, classifies
  the task's complexity, and binds its actual route receipt into the preview.
  No durable skill or coordinator-maintained list maps Medium or Large to a
  closed set of models, and CLI catalog defaults are never translated into DDR.
- If the selected route cannot be dispatched by an isolated, hash-provable eval
  adapter, preview returns `adapter-unavailable`; it does not silently replace
  the route with an available Codex CLI model.
- The app-native Terra lane uses no-history agents bound to the sealed actor
  snapshots. Its receipt proves the host-harness route request and returned
  agent identity; because the app transport does not separately expose a
  provider-reported model field, claims remain bounded to harness-selected
  Terra rather than independent provider attestation.
- The evaluator/router seam now shares the sealed actor, prompt, schema,
  instruction configuration, structured output, and route/session attribution.
  Availability still fails closed when live discovery cannot create its
  protected run root or cannot select a dispatch-qualified route.
- Preview inventories all instruction surfaces capable of influencing the
  actor; a runtime without isolated or hash-provable instruction loading is
  excluded or marked inconclusive.
- Live execution requires explicit approval of that preview.
- Results preserve detection, resolution, revision, user questions, commands,
  turns, timing, and token fields separately; unavailable telemetry is marked
  unknown.
- Conflicting or underpowered samples are baseline evidence, not a forced
  ranking.
- Raw runs remain under `.kb`; the checked-in baseline contains normalized
  metrics and provenance needed for matched comparison.
- This chronological baseline is an anchor. Final promotion depends on a later
  interleaved control/treatment run using the same runtime/model window.

## Scope Boundary

No instruction treatment, skill sync, commit, push, or delivery.
