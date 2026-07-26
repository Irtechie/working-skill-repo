---
kb_id: kb-2026-07-26-change-aware-proof-governor
slice_id: slice-004
title: "Prove, document, and release the proof governor"
blockers: [slice-003]
verification: integration
test_level: full
functional_risk: full
execution_class: cli
model_tier: large
model_tier_reason: "The completion claim spans CLI behavior, snapshot execution, workflow skills, browser/native safety, documentation, and global skill propagation."
model_requirements: ["end-to-end deterministic fixtures", "Windows process checks", "skill contract review", "release and sync verification"]
escalation_triggers: ["the selftest launches a real GUI", "core or local-release remains unresponsive", "global skill drift contains useful target-only changes", "a fixture can pass after skipping changed coverage"]
context_packet_path: docs/plans/2026-07-26-proof-governor-context/slice-004.json
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck proof-governor-selftest"
  expect: 0
hitl: false
expected_files:
  - path: evals/proof-governor/fixtures.json
    op: create
    scope: "Cover unchanged superset reuse, changed-input invalidation, unknown-impact rerun, replay ceiling, and GUI refusal."
  - path: cmd/kbcheck/proof_governor_selftest.go
    op: create
    scope: "Run the deterministic end-to-end proof-governor matrix without real GUI execution."
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "Register the proof-governor selftest in core."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document focused, manifest, and release proof timing plus decision output."
  - path: docs/context/eval-map.md
    op: edit
    scope: "Add the proof-governor fixture and coverage/invalidation evaluation surface."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Document proof receipt ownership, conservative invalidation, and attended GUI boundary."
  - path: README.md
    op: edit
    scope: "Expose the visible change-aware verification workflow and command."
  - path: config/skill-quality.json
    op: edit
    scope: "Include new workflow-contract and proof-governor checks in release quality."
protected_oracles:
  - path: evals/proof-governor/fixtures.json
    role: "end-to-end changed-input, superset-reuse, loop, and GUI refusal corpus"
    sha256: "7a5df23f89ec255c7d75fdd05d8750e75292a4366a466e93bdeb7c01dbdb9ece"
    update_policy: "requires explicit plan amendment"
status: in_progress
owner: agent
blocked_reason: "Core exposed twelve unchanged cmd/kbrouter canonical-project-path fixture failures; full sync report also has unrelated kb-gate/safe-shell-quoting drift. Local-release was not launched because it composes the same failing gates."
resume_when: "The cmd/kbrouter fixture failure and unrelated required global sync drift are repaired with focused proof."
next_agent_action: "After that external blocker changes, run one fresh core and then one local-release. If both pass, close the slice gate and continue to kb-finalize."
human_action: ""
can_continue_other_slices: true
---

# Slice 004: End-to-End Proof and Release

## What to Build

Add a deterministic corpus that demonstrates both sides of the policy: no
changed behavior is skipped, and no unchanged equivalent proof is needlessly
executed. Document the exact decision output, add the check to the contributor
gate, then propagate only after reviewing global drift.

## Acceptance Criteria

1. Fixtures cover passing-superset reuse, relevant input changes, shared
   dependency changes, unrelated changes, unknown impact, failed/partial
   receipts, semantic digest fields, run namespaces, registry drift, replay
   ceilings, timeout cleanup, browser batching, and GUI approval denial.
2. The selftest never launches a real browser or native window and asserts zero
   child-process launches for reuse/block cases.
3. Documentation distinguishes slice-focused proof, changed-workflow manifest
   smoke, and one full release/milestone suite.
4. `go run ./cmd/kbcheck core` includes the proof-governor selftest and runs with
   bounded diagnostics; `local-release` runs once after all focused proof passes.
5. Repo/global skill drift is reviewed before syncing. Final approved skill
   copies are hash-identical across Codex, Copilot, and shared Agents.
6. `git diff --check`, focused tests, core, local-release, and
   `proof-governor-selftest` pass without a visible/native GUI process.

## Test Scenarios

- Mutate a relevant fixture and prove the prior result is invalidated.
- Mutate an unrelated fixture and prove the covered request is reused.
- Request a subset after a full passing suite and assert no process launch.
- Request a visible/native test and assert unconditional automatic pre-spawn
  denial.
- Run a hanging fake child and assert timeout, process-tree cleanup, exit 124,
  partial evidence, and no global-pass claim.
- Run multiple fake DOM checks and assert a single bounded headless browser
  batch rather than one browser process per assertion.
- Run the final full release gate once, not after each fixture or repair.

## Scope Boundary

No live model call, production deployment, Chromium issue filing, GUI launch,
commit, push, or global sync occurs without its separately required authority.

## Proof

Focused objective: `go run ./cmd/kbcheck proof-governor-selftest`

Release once at the genuine milestone:

```powershell
go run ./cmd/kbcheck core
go run ./cmd/kbcheck local-release
git diff --check
```
