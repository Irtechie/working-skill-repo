---
kb_id: kb-2026-08-23-epistemic-investigation-gate
slice_id: epistemic-001
title: Extend skill-eval with one sealed epistemic path
blockers: []
verification: tdd
test_level: integration
functional_risk: broad
model_tier: large
model_tier_reason: The slice changes existing result, scorer, adapter, and regression contracts while preserving all current behavior.
model_requirements: [Go compatibility design, deterministic scoring, isolated process workspaces, experimental-design reasoning]
escalation_triggers: [existing v1 fixtures regress, oracle enters actor context, new command appears without RED proof, live worktree collision is detected]
workspace_mode: shared-serial
conflict_domains: [namespace:skill-eval, file:cmd/kbcheck/skill_eval.go, file:cmd/kbcheck/eval_adapters.go, file:evals/skill-eval/result.schema.json, path:evals/skill-eval/epistemic]
shared_resources: [git:integration-owner, eval:skill-eval, worktree:plan-run]
context_packet_path: docs/plans/2026-08-23-epistemic-investigation-context/epistemic-001.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'SkillEvalEpistemic|SkillEvalBaseline' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: evals/skill-eval/result.schema.json
    op: edit
    scope: Add optional epistemic actor-result and available-telemetry fields while preserving existing fixtures unchanged.
  - path: evals/skill-eval/epistemic/visible/
    op: create
    scope: Add actor-visible mini-repository tasks and evidence without labels or expected paths.
  - path: evals/skill-eval/epistemic/oracles/
    op: create
    scope: Add separately loaded decision, resolving-evidence, and final-state labels.
  - path: cmd/kbcheck/skill_eval_epistemic.go
    op: create
    scope: Compute detection, resolution, revision, and cost metrics inside the existing skill-eval owner.
  - path: cmd/kbcheck/skill_eval_epistemic_test.go
    op: create
    scope: Protect compatibility, confusion-matrix, resolution, revision, leakage, and non-aggregate behavior.
  - path: cmd/kbcheck/skill_eval.go
    op: edit
    scope: Dispatch optional epistemic results to the focused scorer and carry separate metrics into existing baselines/regression reports.
  - path: cmd/kbcheck/eval_adapters.go
    op: edit
    scope: Select epistemic fixtures and run from a disposable actor-only workspace while the parent retains the oracle.
  - path: config/skill-quality.json
    op: edit
    scope: Register the epistemic corpus under the existing skill-eval contract.
protected_oracles:
  - path: cmd/kbcheck/skill_eval_epistemic_test.go
    role: Compatibility and epistemic scoring behavior oracle.
    sha256: filled by kb-work after RED/protection
    update_policy: Protect before implementation; later semantic changes require explicit plan review.
  - path: evals/skill-eval/epistemic/oracles/
    role: Hidden decision and evidence labels.
    sha256: filled by kb-work from a corpus manifest
    update_policy: Freeze before baseline; any later change invalidates the comparison.
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: Recheck live worktree/lease overlap, write protected tests, prove RED, then extend existing owners.
human_action: ""
can_continue_other_slices: false
---

# Existing Skill-Eval Extension

## Deliverable

One end-to-end epistemic result path through the existing result schema,
`computeSkillEval`, live adapter, baseline rows, and regression report.

## Acceptance Criteria

- All current route, claim, quality, manifest, and baseline fixtures pass
  unchanged.
- Epistemic fields are optional for legacy results and strict when present.
- The actor chooses `proceed` or `investigate`; after investigation it reports
  `supported`, `contradicted`, or `no-justified-conclusion`.
- The scorer reports the detection confusion matrix, correct evidence
  resolution, correct revision, user questions, observed commands, and
  available runtime telemetry separately.
- The actor result carries the actual generated user-facing response text;
  supported `proceed` fixtures reject investigation announcements, proof
  checklists, or other visible verification ceremony.
- Missing runtime telemetry is `unknown`, never invented or treated as zero.
- The adapter materializes a disposable mini-repository containing visible
  task/evidence and required instruction surfaces but no oracle files.
- The adapter uses a supported isolated runtime profile when available and
  records the hashes of repo, global, user, and cached instruction surfaces
  actually capable of influencing the actor. Unprovable instruction loading
  fails the runtime comparison closed.
- Every scored oracle names a deterministic construction rule or
  machine-checkable source truth; unsupported hand-authored labels fail corpus
  readiness.
- Subjective judgment fixtures, if retained, are marked process-only and never
  contribute to conclusion-correctness metrics.
- The parent run manifest hash-binds visible fixtures, hidden oracle, scorer,
  result schema, adapter, instruction surfaces, and runtime-reported identity.
- Existing `skill-eval`, `skill-eval-regression`, `eval-run-*`, and
  `skill-eval-wrap` commands own the path; no parallel command family is added.
- When epistemic comparison data is present, `skill-eval-regression` records
  `promote`, `reject`, or `inconclusive`; only `promote` exits successfully for
  the final protected comparison. The other states preserve their report but
  cannot satisfy the objective gate.

## Test Scenarios

- Supported premise proceeds without investigation.
- Unsupported premise proceeds and scores as missed investigation.
- Supported premise investigates and scores as unnecessary investigation.
- Correct evidence supports, contradicts, or leaves the premise unresolved.
- Wrong evidence plus a guessed answer fails resolution.
- Oracle field/path enters the actor prompt or workspace and fails closed.
- An unexpected global skill, user instruction, or reused session can influence
  the actor and fails instruction-isolation readiness.
- Oracle label disagrees with its machine-checkable fixture truth and fails
  corpus readiness.
- Existing version-one result fixtures remain valid and score identically.

## Scope Boundary

No live model call, instruction treatment, global sync, or delivery.
