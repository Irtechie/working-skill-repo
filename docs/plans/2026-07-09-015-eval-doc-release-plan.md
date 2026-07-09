---
kb_id: kb-2026-07-09-phoenix-routing-slicing-absorption
slice_id: slice-005
title: "Add measured eval result docs and release sync"
blockers: [slice-003, slice-004, slice-006]
verification: verification-only
test_level: full
functional_risk: none
model_tier: medium
model_route: hosted-sonnet
hitl: false
expected_files:
  - path: "evals/"
    op: edit
    scope: "Add fixtures that prove false-done, vacuous done_check, and missing proof_check are rejected."
  - path: "docs/results/"
    op: add
    scope: "Record RESULT-style measured KB workflow outcomes with commands, dates, corpus, and limitations."
  - path: "README.md"
    op: edit
    scope: "Update visible Phoenix absorption status after implementation."
  - path: "docs/context/eval-map.md"
    op: edit
    scope: "List the new eval surfaces and commands."
  - path: "docs/context/operations/testing.md"
    op: edit
    scope: "Document canonical verification commands."
  - path: "todo.md"
    op: edit
    scope: "Move this work to done or update blockers after verification."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "None; slice proof passed and is recorded in the parent manifest."
human_action: ""
can_continue_other_slices: true
---

# Slice 005: Add Measured Eval Result Docs And Release Sync

## What This Delivers

The new Phoenix-inspired routing/slicing contract is backed by deterministic
fixtures, visible README/testing docs, and a small public measurement record,
not just skill prose.

## Acceptance Criteria

- Eval fixtures reject:
  - a goal whose `done_check` is already green/vacuous;
  - a slice marked done without required `proof_check` evidence;
  - invalid `model_route`;
  - route-history oscillation without progress.
- README describes what is already implemented and what remains optional.
- A `docs/results/` RESULT-style artifact records:
  - corpus size and fixture IDs;
  - exact commands run;
  - date and environment;
  - pass/fail/skip counts;
  - honest limitations;
  - a clear statement that Phoenix's metrics are not KB's metrics.
- `docs/context/eval-map.md` and `docs/context/operations/testing.md` list the
  canonical commands.
- `go run ./cmd/kbcheck local-release` passes or any blocker is recorded with
  exact command/output.
- Required global/ATV skill roots are synced only after the release gate passes.

## Test Scenarios

- Run the new fixture selftests.
- Generate or update the RESULT-style artifact from the actual run output.
- Run `git diff --check`.
- Run `go run ./cmd/kbcheck local-release` after the gate-health issue from
  slice 001 is resolved or explicitly documented.

## Scope Boundary

Do not claim Phoenix's published metrics for KB. Do not invent a baseline. This
slice may only claim the fixtures and commands it actually runs.

## Result

Done on 2026-07-09. Proof:

- `go run ./cmd/kbcheck dishonest-completion-selftest` passed, rejecting 4/4
  deterministic false-completion fixtures.
- `go run ./cmd/kbcheck skill-sync-report` passed with 234 comparisons and 0
  required issues.
- `go run ./cmd/kbcheck doctor --fix` passed with 0 required issues.
- `go run ./cmd/kbcheck local-release --json` passed with 0 required failures
  and 0 optional failures.
- RESULT artifact:
  `docs/results/2026-07-09-kb-phoenix-routing-slicing-result.md`.
