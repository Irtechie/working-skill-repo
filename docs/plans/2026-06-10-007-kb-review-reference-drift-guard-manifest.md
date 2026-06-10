---
type: kb-manifest
kb_id: kb-2026-06-10-review-reference-drift-guard
brainstorm_path: direct-chat
created: 2026-06-10
status: reviewed
workflow_shape: "skill-bundle-change"
scope-verified-files:
  - cmd/kbcheck/checks.go
  - cmd/kbcheck/checks_test.go
  - cmd/kbcheck/main.go
  - cmd/kbcheck/review_reference_guard.go
  - cmd/kbcheck/review_reference_guard_test.go
  - cmd/kbcheck/skill_validators.go
  - config/skill-quality.json
  - docs/context/memory-maintenance.md
  - docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md
  - docs/plans/2026-06-10-008-review-reference-drift-guard-plan.md
  - todo.md
  - todo-done.md
  - .atv/kb-completions.txt
  - docs/solutions/workflow-issues/review-reference-drift-guard-2026-06-10.md
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-start
    status: passed
    required_evidence:
      - "user supplied verification findings"
      - "kb-map preflight found repo-local memory"
      - "no unresolved ask-now or research-first blockers remain"
      - "safe assumptions are recorded"
    proof:
      - docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-plan direct-chat-review-reference-drift-guard"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest path exists"
      - "all slice plan paths exist"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk"
      - "HITL classification is explicit"
    proof:
      - docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md
      - docs/plans/2026-06-10-008-review-reference-drift-guard-plan.md
      - todo.md
      - docs/context/PROJECT.md
      - config/skill-quality.json
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-work docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "deterministic checks ran"
      - "functional CLI proof ran"
      - "memory impact classified"
      - "protected oracles absent by design"
    proof:
      - cmd/kbcheck/review_reference_guard.go
      - cmd/kbcheck/review_reference_guard_test.go
      - config/skill-quality.json
      - docs/context/memory-maintenance.md
      - docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md
      - docs/plans/2026-06-10-008-review-reference-drift-guard-plan.md
    proof_commands:
      - "go test ./cmd/kbcheck"
      - "go run ./cmd/kbcheck review-reference-guard"
      - "go run ./cmd/kbcheck core"
      - "go run ./cmd/kbcheck local-release"
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-work docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md"
  - gate_id: work-to-complete
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "every non-skipped slice has passing slice-to-done gate"
      - "final verification command/result recorded"
      - "no unresolved P0/P1 exists"
      - "board/manifest are synced"
      - "scope-verified-files is populated"
      - "memory refresh completed"
    proof:
      - docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md
      - docs/plans/2026-06-10-008-review-reference-drift-guard-plan.md
      - todo.md
      - todo-done.md
      - config/skill-quality.json
      - docs/context/memory-maintenance.md
    proof_commands:
      - "go test ./cmd/kbcheck"
      - "go run ./cmd/kbcheck review-reference-guard"
      - "go run ./cmd/kbcheck core"
      - "go run ./cmd/kbcheck local-release"
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-complete docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md"
  - gate_id: complete-to-ship
    owner_skill: kb-complete
    status: passed
    required_evidence:
      - "kb-check final command/result recorded"
      - "functional-test skip reason recorded"
      - "kb-review mode and finding counts recorded"
      - "P0/P1 resolved or no P0/P1"
      - "follow-up-resolution summary recorded"
      - "proof/demo evidence recorded"
      - "compound/learn/evolve result recorded"
      - "project-memory refresh proof recorded"
      - "memory-maintenance update recorded"
      - "cleanup result recorded"
      - "alerts list recorded"
    proof:
      - docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md
      - docs/plans/2026-06-10-008-review-reference-drift-guard-plan.md
      - cmd/kbcheck/review_reference_guard.go
      - cmd/kbcheck/review_reference_guard_test.go
      - config/skill-quality.json
      - docs/context/memory-maintenance.md
      - docs/solutions/workflow-issues/review-reference-drift-guard-2026-06-10.md
      - .atv/kb-completions.txt
      - todo.md
      - todo-done.md
      - cmd/kbcheck/checks.go
    proof_commands:
      - "go test ./cmd/kbcheck"
      - "go run ./cmd/kbcheck review-reference-guard"
      - "go run ./cmd/kbcheck core"
      - "go run ./cmd/kbcheck local-release"
      - "git diff --check"
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-ship docs/plans/2026-06-10-007-kb-review-reference-drift-guard-manifest.md"
slices:
  - id: slice-001
    title: "Protect shared review-skill references from silent drift"
    path: docs/plans/2026-06-10-008-review-reference-drift-guard-plan.md
    blockers: []
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add the review-reference drift contract, wire it into the Go gate, document intentional forks, and run the focused CLI proof."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=7 changed=12 discovered=3 unexplained=0; test-level=functional-cli; proof=go test ./cmd/kbcheck, go run ./cmd/kbcheck review-reference-guard, go run ./cmd/kbcheck core, go run ./cmd/kbcheck local-release; memory-impact=durable docs=context/memory-maintenance.md"
    protected_oracles: []
---

# KB: Review Reference Drift Guard

## Origin

Direct user challenge after commit `1a39aeb9c8cac10024b26a5147a945869d34fefd`: slice 005 was marked done, but the planned guard for duplicated `ce-review` / `kb-review` references was not shipped.

## Workflow Shape

`skill-bundle-change` - this changes deterministic validation for portable review skills and updates durable repo memory.

## Safe Assumptions

- The existing public skill split remains: `kb-review` is the KB completion lane, `ce-review` is the generalized CE lane.
- Byte-identical shared references should stay duplicated on disk for portability unless a later approved refactor centralizes them.
- Intentional forks should be documented explicitly rather than hidden by the hash guard.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Protect shared review-skill references from silent drift | - | integration / functional-cli | no | done |

## Out of Scope

- Agent deletion or merge decisions.
- Hot-path skill line-count trimming.
- CI workflow changes for `local-release`.
- Broad duplicate-detection heuristics beyond the declared review-reference contract.

## Work Completion Evidence

- scope-check: forecast=7 changed=12 discovered=3 unexplained=0.
- discovered files: `cmd/kbcheck/main.go`, `todo.md`, `todo-done.md`.
- proof: `go test ./cmd/kbcheck`, `go run ./cmd/kbcheck review-reference-guard`, `go run ./cmd/kbcheck core`, `go run ./cmd/kbcheck local-release` passed.
- functional test: functional-cli proof through the new `review-reference-guard` command and inclusion in `core`; browser proof skipped because this is non-UI tooling.
- memory impact: durable; `docs/context/memory-maintenance.md` updated with the review-skill drift contract.

## kb-complete closure - 2026-06-10

- review: local-fallback; P0=0 P1=0 P2=0 P3=0.
- follow-up-resolution: resolved 0, logged 0, blocked 0.
- proof/demo evidence: `go test ./cmd/kbcheck`, `go run ./cmd/kbcheck review-reference-guard`, `go run ./cmd/kbcheck core`, `go run ./cmd/kbcheck local-release`, and `git diff --check` passed; browser verification skipped because this is non-UI tooling.
- compound: `docs/solutions/workflow-issues/review-reference-drift-guard-2026-06-10.md`.
- learn: no new instinct promoted; the reusable pattern is captured in the compound note without inflating project instincts from one observation.
- evolve: skipped; completion counter is 8, not divisible by 5.
- project memory: refreshed `docs/context/memory-maintenance.md`; `PROJECT.md` route map already points to maintenance signals and testing gates.
- memory-maintenance: review-skill drift contract row added.
- cleanup: no QA screenshots or observation artifacts created; todo hygiene complete.
- alerts: macOS/Linux `local-release` proof remains the existing open platform signal; cold-storage deletion candidates remain human-gated.
