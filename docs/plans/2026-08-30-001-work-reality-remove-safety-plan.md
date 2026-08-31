---
slice_id: closure-001
title: Preserve non-removable todo rows during work-reality removal
status: planned
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: destructive
execution_class: cli
model_tier: medium
model_tier_reason: "This changes a user-authored todo mutation path whose wrong result invents lifecycle state. The current unsafe branch and exact desired behavior are known, but data-integrity reasoning and fixture coverage are required."
model_requirements:
  - "Go edits and tests within the existing kbcheck command"
  - "line-oriented Markdown mutation safety"
  - "distinguishing terminal proof from preservation"
escalation_triggers:
  - "a nonterminal or uncontained pairing would still rewrite a todo row"
  - "removal requires changing the parser shape or weakening containment proof"
  - "a test cannot prove byte-for-byte row preservation"
token_budget: 5000
cost_tier: 2
cheaper_option_ruled_out: "Documenting the trap alone leaves a command that corrupts truthful pending and in-progress states; the existing mutation path and fixture harness are the sufficient reuse surface."
owning_component: kbcheck
expected_files:
  - cmd/kbcheck/work_reality.go
  - cmd/kbcheck/work_reality_test.go
conflict_domains: [go:kbcheck-commands, cli:kbcheck, docs:todo]
shared_resources: [todo.md mutation grammar, work-reality report schema]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'WorkReality|RehabRemoval' -count=1"
  expect: 0
hitl: none
---

# Preserve Non-Removable Todo Rows During Removal

## Outcome

`go run ./cmd/kbcheck work-reality --root . --action remove --json` deletes
only an eligible terminal row. Every row that is active, pending, blocked,
uncontained, missing a resolving artifact, or otherwise ineligible remains
byte-for-byte unchanged.

## Ordered Work

1. Read `applyWorkRealityMarks` and the existing removal fixtures. Identify the
   branch that creates `rehabMarkerBlocked` or `rehabMarkerSkipped` under
   `removalRequested`.
   - Pass criterion: each non-removable pairing increases/report-preserves
     without appending a planned marker write.
2. Change only removal-mode handling: keep existing terminal marking behavior
   for `--action mark`; retain removal for terminal, contained rows with a
   resolving artifact.
   - Pass criterion: removal mode has no planned rewrite for an ineligible row.
3. Replace the existing test that expects a blocked re-mark with tests covering
   pending and in-progress Markdown rows unchanged byte-for-byte, while the
   eligible removal test remains green.
   - Pass criterion: the before/after file equality assertion catches status
     and link-cell corruption.

## Acceptance Criteria

- An uncontained `pending` row is preserved, not marked blocked.
- An in-progress row with no resolving artifact is preserved, not marked
  skipped.
- A contained superseded row with a landed resolving artifact is removed.
- A fail-closed report still writes nothing.
- JSON reports distinguish `preserved` from a real write.

## Scope Boundary

Do not change terminal classifications, remote-authority proof, marker grammar,
or any `todo.md` row in the repository outside fixtures.
