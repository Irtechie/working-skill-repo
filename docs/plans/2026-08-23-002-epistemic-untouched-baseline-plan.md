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
    sha256: filled by kb-work from the corpus manifest before live execution
    update_policy: Any change invalidates the baseline and requires recapture.
  - path: cmd/kbcheck/skill_eval_epistemic.go
    role: Frozen epistemic scorer.
    sha256: filled by kb-work before live execution
    update_policy: Any semantic change invalidates baseline-treatment comparison.
test_inputs:
  - name: baseline_preview
    source: generated
    required_for: Exact live-run approval
    value: .kb/eval-runs/epistemic-baseline-preview.json
status: pending
owner: agent
blocked_reason: ""
resume_when: epistemic-001 passes and the user approves the exact no-run preview
next_agent_action: Freeze hashes, generate a no-run preview, request approval, capture baseline, and verify the existing regression artifact.
human_action: Approve or reject the exact runtime/model, repetitions, corpus hash, and bounded cost.
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
- Preview mode makes no model call and names exact runtimes/models,
  repetitions, hashes, and bounded cost when the host exposes cost.
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
