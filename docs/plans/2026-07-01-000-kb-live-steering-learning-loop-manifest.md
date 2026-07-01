---
type: kb-manifest
kb_id: kb-2026-07-01-live-steering-learning-loop
scope-verified-files:
  - .github/skills/kb-goal/SKILL.md
  - .github/skills/kb-plan/SKILL.md
  - .github/skills/kb-complete/SKILL.md
  - .github/skills/learn/SKILL.md
  - AGENTS.md
  - README.md
  - .atv/instincts/project.yaml
  - .atv/kb-completions.txt
  - docs/context/architecture/kb-workflow.md
  - docs/context/goals/live-steering-learning-loop.md
  - docs/context/memory-maintenance.md
  - docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md
  - docs/plans/2026-07-01-001-skill-live-steering-goal-plan.md
  - docs/plans/2026-07-01-002-skill-feedback-learning-plan.md
  - docs/plans/2026-07-01-003-docs-proof-sync-plan.md
  - docs/solutions/workflow-issues/live-steering-learning-loop-2026-07-01.md
  - todo.md
  - todo-done.md
brainstorm_path: docs/context/goals/live-steering-learning-loop.md
created: 2026-07-01
status: reviewed
workflow_shape: "skill-bundle-change"
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-goal
    status: passed
    required_evidence:
      - "goal ledger exists"
      - "source comparison identified adopted and rejected mechanics"
      - "no unresolved ask-now or research-first items remain"
      - "safe assumptions and scope boundaries are recorded"
    proof:
      - docs/context/goals/live-steering-learning-loop.md
      - "HumanLayer design-control-loop clone at 39fb32786ae7a7cd864cf2c237148c38b1e4db07 was compared against local KB learning skills"
    blockers: []
    passed_at: "2026-07-01T00:00:00-04:00"
    allowed_next_action: "kb-plan docs/context/goals/live-steering-learning-loop.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest path exists"
      - "all slice plan paths exist"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk"
      - "HITL classification recorded"
    proof:
      - docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md
      - docs/plans/2026-07-01-001-skill-live-steering-goal-plan.md
      - docs/plans/2026-07-01-002-skill-feedback-learning-plan.md
      - docs/plans/2026-07-01-003-docs-proof-sync-plan.md
      - "DAG validated: slice-002 depends on slice-001, slice-003 depends on slice-002; no missing blockers or cycles; all slices include acceptance criteria and verification metadata"
    blockers: []
    passed_at: "2026-07-01T00:00:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "deterministic check ran"
      - "memory impact classified"
    proof:
      - "Edited .github/skills/kb-goal/SKILL.md"
      - "Edited .github/skills/kb-plan/SKILL.md"
      - "command passed: kbcheck skill-lint, 0 errors and known size warnings"
      - "scope-check: forecast=2 changed=2 discovered=0 unexplained=0"
      - "memory-impact: durable; areas=goal planning workflow; docs=docs/context/architecture/kb-workflow.md; refresh=pending"
    blockers: []
    passed_at: "2026-07-01T00:00:00-04:00"
    allowed_next_action: "slice-002"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "deterministic check ran"
      - "memory impact classified"
    proof:
      - "Edited .github/skills/kb-complete/SKILL.md"
      - "Edited .github/skills/learn/SKILL.md"
      - "command passed: kbcheck skill-lint, 0 errors and known size warnings"
      - "scope-check: forecast=2 changed=2 discovered=0 unexplained=0"
      - "memory-impact: durable; areas=completion learning workflow; docs=docs/context/architecture/kb-workflow.md; refresh=pending"
    blockers: []
    passed_at: "2026-07-01T00:00:00-04:00"
    allowed_next_action: "slice-003"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "deterministic checks ran"
      - "memory impact classified"
    proof:
      - "Edited README.md"
      - "Edited docs/context/architecture/kb-workflow.md"
      - "Updated docs/context/goals/live-steering-learning-loop.md"
      - "Updated todo.md"
      - "command passed: kbcheck core, checks=26"
      - "command passed: git diff --check, line-ending warnings only"
      - "scope-check: forecast=4 changed=4 discovered=0 unexplained=0"
      - "memory-impact: durable; areas=visible workflow docs; docs=README.md, docs/context/architecture/kb-workflow.md; refresh=done"
    blockers: []
    passed_at: "2026-07-01T00:00:00-04:00"
    allowed_next_action: "work-to-complete"
  - gate_id: work-to-complete
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "every non-skipped slice has passing slice-to-done gate"
      - "final verification command/result is recorded"
      - "no unresolved P0/P1 exists before completion review"
      - "board and manifest are synced"
      - "scope-verified-files is populated"
    proof:
      - "slice-slice-001-to-done passed"
      - "slice-slice-002-to-done passed"
      - "slice-slice-003-to-done passed"
      - "command passed: kbcheck core, checks=26"
      - "command passed: git diff --check"
      - "todo.md slice statuses synced"
      - "scope-verified-files populated with 12 paths"
    blockers: []
    passed_at: "2026-07-01T00:00:00-04:00"
    allowed_next_action: "kb-complete docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md"
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
      - docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md
      - .github/skills/kb-goal/SKILL.md
      - .github/skills/kb-plan/SKILL.md
      - .github/skills/kb-complete/SKILL.md
      - .github/skills/learn/SKILL.md
      - README.md
      - AGENTS.md
      - docs/context/architecture/kb-workflow.md
      - docs/context/goals/live-steering-learning-loop.md
      - docs/context/memory-maintenance.md
      - docs/solutions/workflow-issues/live-steering-learning-loop-2026-07-01.md
      - .atv/instincts/project.yaml
      - .atv/kb-completions.txt
      - todo.md
      - todo-done.md
      - "review-mode: local-fallback; P0=0 P1=0 P2=1(resolved duplicate observation logging) P3=0"
      - "functional-test: skipped - docs and portable skill instruction changes only, no runtime UI, API, or CLI behavior"
      - "follow-up-resolution: resolved 1, logged 0, blocked 0"
      - "steering-feedback: current=0 memory=0 observations=0 landmine-candidates=0 instinct-evidence=0"
      - "compound: solution note written"
      - "learn: 1 new instinct, 1 updated, 1 decayed; evolve skipped at completion count 11"
      - "project-memory refresh: done - README and workflow architecture updated"
      - "cleanup: todo feature section archived to todo-done.md; observations log absent"
      - "alerts: repo-local change not propagated to global or ATV skill roots"
    proof_commands:
      - "go run ./cmd/kbcheck core"
      - "git diff --check"
      - "python .github/skills/kb-gate/scripts/check_gate_ledger.py docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md --gate work-to-complete --allowed-next \"kb-complete docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md\""
    blockers: []
    passed_at: "2026-07-01T00:00:00-04:00"
    allowed_next_action: "kb-ship docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md"
slices:
  - id: slice-001
    title: "Add long-goal control-loop steering contract"
    path: docs/plans/2026-07-01-001-skill-live-steering-goal-plan.md
    blockers: []
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Edit kb-goal and kb-plan to name optional live-steering fields for long-running goals and recurring improvement work."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-forecast: loaded 2 expected files; scope-check: forecast=2 changed=2 discovered=0 unexplained=0; proof: go run ./cmd/kbcheck skill-lint passed with known size warnings; memory-impact: durable; areas=goal planning workflow; docs=docs/context/architecture/kb-workflow.md; refresh=pending"
    protected_oracles: []
  - id: slice-002
    title: "Add durable feedback classification and steering memory"
    path: docs/plans/2026-07-01-002-skill-feedback-learning-plan.md
    blockers: [slice-001]
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Edit kb-complete and learn so review/iteration feedback can become steering memory, observations, landmines, or instincts without replacing post-work learning."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-forecast: loaded 2 expected files; scope-check: forecast=2 changed=2 discovered=0 unexplained=0; proof: go run ./cmd/kbcheck skill-lint passed with known size warnings; memory-impact: durable; areas=completion learning workflow; docs=docs/context/architecture/kb-workflow.md; refresh=pending"
    protected_oracles: []
  - id: slice-003
    title: "Refresh docs and proof surface"
    path: docs/plans/2026-07-01-003-docs-proof-sync-plan.md
    blockers: [slice-002]
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Update README, workflow architecture docs, goal ledger, and todo with proof. Run core and diff checks."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-forecast: loaded 4 expected files; scope-check: forecast=4 changed=4 discovered=0 unexplained=0; proof: go run ./cmd/kbcheck core passed checks=26; proof: git diff --check passed with line-ending warnings only; memory-impact: durable; areas=visible workflow docs; docs=README.md, docs/context/architecture/kb-workflow.md; refresh=done"
    protected_oracles: []
---

# KB: Live Steering Learning Loop

## Origin

Goal: `docs/context/goals/live-steering-learning-loop.md`

## Workflow Shape

`skill-bundle-change` - the change touches several portable KB skills plus
visible workflow docs. It does not need a new app runtime, Go harness feature,
or HumanLayer-specific CI import.

## Adopted Mechanics

- Control-loop framing for long-running improvement work: set point, sensor,
  controller, actuator, disturbance, dampener, scope gate, batch size, and WIP
  bound.
- Durable steering memory loaded before future runs.
- `/iterate`-style feedback routing as a portable concept: PR/session feedback
  may update the active branch and future steering, but only when durable.
- One-open-work-item flow control for scheduled or repeated loops.

## Rejected Mechanics

- Do not import HumanLayer's Bun, CodeLayer, or GitHub Actions templates as KB
  defaults.
- Do not replace `kb-complete`, `learn`, `evolve`, or repo memory refresh with
  a scheduled-agent loop.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Add long-goal control-loop steering contract | - | verification-only | no | done |
| 2 | Add durable feedback classification and steering memory | slice-001 | verification-only | no | done |
| 3 | Refresh docs and proof surface | slice-002 | verification-only | no | done |

## Goal Link

Update `docs/context/goals/live-steering-learning-loop.md` as each slice
finishes. The goal remains active until deterministic proof is recorded.
