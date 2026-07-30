---
kb_id: kb-2026-07-30-model-tier-qualification
slice_id: slice-002
title: "Classify model tier evidence deterministically"
blockers: [slice-001]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "The scorer makes reusable capability claims from hostile cross-repository evidence and must fail closed across trust, state, statistics, and replay boundaries."
model_requirements: ["Go strict-input security", "receipt and signature verification", "deterministic state-machine design", "statistical threshold reasoning", "CLI contract testing"]
escalation_triggers: ["the scorer needs network or inference access", "private keys or endpoints enter repository state", "unsupported attestations can produce qualified", "routing promotion is coupled to classification"]
workspace_mode: shared-serial
conflict_domains: ["file:cmd/kbcheck/main.go", "namespace:model-tier-eval", "path:evals/model-tier-qualification"]
shared_resources: ["git:integration-owner"]
context_packet_path: docs/plans/2026-07-30-model-tier-qualification-context/slice-002.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'ModelTierEval' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/model_tier_eval.go
    op: create
    scope: "Implement strict offline evidence validation, admission state machine, scoped Medium decision, and compact JSON/text output."
  - path: cmd/kbcheck/model_tier_eval_test.go
    op: create
    scope: "Protect every fatal, exclusion, failure, replay, trust, statistical, freshness, and qualification branch."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Register model-tier-eval and its evidence flag without enabling live calls."
  - path: evals/model-tier-qualification/fixtures.json
    op: create
    scope: "Provide fixed inputs and deterministic expected decisions consumed by the scorer selftest."
  - path: evals/model-tier-qualification/README.md
    op: create
    scope: "Document evidence authorities, Medium policy, exclusions, redaction, and experimental limitations."
  - path: docs/context/eval-map.md
    op: edit
    scope: "Add the deterministic model-tier qualification workflow and exact proof command."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document focused and canonical classifier checks."
  - path: README.md
    op: edit
    scope: "Describe experimental evidence-based tier qualification without a static model roster or automatic routing promotion."
protected_oracles:
  - path: cmd/kbcheck/model_tier_eval_test.go
    role: "model-tier evidence and decision oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan update"
  - path: evals/model-tier-qualification/fixtures.json
    role: "fixed deterministic classifier corpus"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Freeze RED fixtures for every decision and exclusion path, then implement the offline experimental scorer and docs."
human_action: ""
can_continue_other_slices: true
---

# Classify Model Tier Evidence Deterministically

## What to Build

Add an experimental offline `kbcheck model-tier-eval` command that consumes one
bounded strict evidence file, authenticates or reproduces its receipts, applies
the admission state machine and frozen Medium policy, and returns
`qualified`, `not-qualified`, or `inconclusive` without network or inference.

## Acceptance Criteria

- Input is bounded, strict, repo-contained, non-symlinked, and secret-safe.
- Cohort preregistration, completeness, family independence, attempt identity,
  execution fingerprint, plan/oracle/proof receipts, trust authority, and
  freshness are mechanically checked.
- Fatal document errors are distinct from exclusions and admitted model
  failures.
- Output-schema and deterministic-proof failures after valid execution count as
  model failures.
- The Medium policy requires 30 admitted unique fixtures, five independent
  families, max 20% per family, one holdout family, and zero model failures.
- Stale, unsupported, underpowered, or incomplete evidence is inconclusive.
- Any admitted model failure is not-qualified.
- The result records tested scope and cannot promote routing automatically.
- Default contributor checks remain no-paid and deterministic.

## Test Scenarios

- Complete passing cohort qualifies.
- One admitted output-schema or proof failure is not-qualified.
- Zero/29 samples, four families, missing holdout, or family concentration is
  inconclusive.
- Plan, oracle, preflight, route, no-response, indeterminate, and proof
  infrastructure outcomes are excluded and visible.
- Omitted/replayed/conflicting attempts, artifact mismatch, mixed fingerprint,
  unknown fields, traversal, symlink, oversize, sensitive data, and forged or
  unsupported authority fail or become inconclusive according to contract.
- Stale evidence and unobservable provider revision beyond 30 days are
  inconclusive.
- A second target tier is rejected as unsupported.

## Scope Boundary

No live calls, route discovery, credential access, private-key storage, model
roster, automatic production promotion, or inference benchmark runner.

