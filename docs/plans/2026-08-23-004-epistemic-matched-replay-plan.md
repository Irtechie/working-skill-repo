---
kb_id: kb-2026-08-23-epistemic-investigation-gate
slice_id: epistemic-004
title: Run interleaved control/treatment replay and enforce promotion boundaries
blockers: [epistemic-003]
verification: functional
test_level: full
functional_risk: broad
model_tier: large
model_tier_reason: Final comparison must preserve separate outcomes, reject ease regressions, and avoid forcing a conclusion from weak evidence.
model_requirements: [live eval operation, regression analysis, release verification, bounded claim synthesis]
escalation_triggers: [hash mismatch, unnecessary investigation or user interruption worsens, evidence is underpowered, local-release fails]
workspace_mode: shared-serial
conflict_domains: [namespace:skill-eval, path:.kb/eval-runs/epistemic-treatment, path:docs/results, file:docs/context/eval-map.md, file:docs/context/operations/testing.md]
shared_resources: [git:integration-owner, eval:live-runtime, budget:explicit-live-model]
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck skill-eval-regression --run-root .kb/eval-runs/epistemic-treatment --baseline evals/skill-eval/baselines/epistemic-investigation.json --output docs/results/epistemic-investigation-comparison.json"
  expect: 0
hitl: true
expected_files:
  - path: docs/results/epistemic-investigation-comparison.json
    op: create
    scope: Store matched metric deltas, provenance checks, and promote/reject/inconclusive state.
  - path: docs/results/epistemic-investigation-comparison.md
    op: create
    scope: Explain the bounded decision without a combined flattering score or training-cause claim.
  - path: docs/context/eval-map.md
    op: edit
    scope: Register the skill-eval extension, commands, explicit live boundary, and limits.
  - path: docs/context/operations/testing.md
    op: edit
    scope: Document deterministic selftests, actor isolation, preview approval, baseline, and replay.
protected_oracles:
  - path: evals/skill-eval/baselines/epistemic-investigation.json
    role: Matched untouched regression baseline.
    sha256: filled by kb-work before treatment replay
    update_policy: Immutable during comparison.
  - path: cmd/kbcheck/skill_eval_epistemic.go
    role: Frozen scorer.
    sha256: filled by kb-work before treatment replay
    update_policy: Any semantic change invalidates both cohorts.
test_inputs:
  - name: treatment_preview
    source: generated
    required_for: Exact matched live replay approval
    value: .kb/eval-runs/epistemic-treatment-preview.json
status: pending
owner: agent
blocked_reason: ""
resume_when: epistemic-003 passes and the user approves the exact interleaved control/treatment replay preview
next_agent_action: Preview, request approval, interleave reconstructed control and treatment arms, compare, and return promote, reject, or inconclusive.
human_action: Approve or reject the exact interleaved replay; separately authorize any later global sync or delivery.
can_continue_other_slices: false
---

# Interleaved Matched Replay and Promotion

## Deliverable

A comparison produced by the existing regression command from interleaved
control and treatment arms on the same runtime/model window. It accepts the
treatment only when epistemic behavior improves without making supported work
harder or increasing user burden.

## Acceptance Criteria

- The untreated instruction bundle is reconstructed from baseline-bound content
  hashes or Git blob identities; the treatment bundle is separately hash-bound.
- Control and treatment cases are interleaved with matched runtime/model
  identity, fixture/scorer/schema hashes, repetitions, and run configuration.
- Every arm uses a fresh context. The report records per-fixture arm order,
  randomization seed or deterministic alternation rule, repetition count, and
  available inference settings.
- A missing stable model identity or fresh-context boundary makes the result
  `inconclusive`.
- Unexpected or unprovable repo/global/user instruction surfaces, cached
  sessions, or control/treatment bundle drift make that runtime result
  `inconclusive`.
- The earlier chronological baseline is reported as a drift anchor, not used as
  the sole causal comparator.
- Missed investigation, unnecessary investigation, resolution, revision, user
  questions, commands, turns, timing, and tokens stay separately visible.
- Promotion requires missed investigations to decrease or remain zero,
  unnecessary investigations not to increase, and resolution/revision not to
  decrease.
- Agent-resolvable fixtures add no user questions.
- Supported `proceed` cases add no visible ceremony.
- Hash drift, missing provenance, underpowered results, or conflicting metrics
  returns `inconclusive`, not a positive claim.
- `reject` and `inconclusive` preserve their reports but return a non-successful
  protected comparison, so neither can satisfy the manifest objective.
- Existing skill-eval selftests, `go run ./cmd/kbcheck core`, and
  `go run ./cmd/kbcheck local-release` pass before any later sync request.

## Decision States

| State | Meaning |
|---|---|
| `promote` | Protected epistemic outcomes improve or hold and ease does not regress |
| `reject` | A protected outcome or workflow-ease boundary worsens |
| `inconclusive` | Evidence cannot justify promotion or rejection |

## Scope Boundary

No automatic oracle revision, global sync, commit, push, PR, merge, universal
model claim, or training-cause claim. Deterministic `kb-work` enforcement is
owned by epistemic-005 and runs only after `promote`.
