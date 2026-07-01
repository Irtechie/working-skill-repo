---
title: Live Steering Complements Post-Work Learning
date: 2026-07-01
category: docs/solutions/workflow-issues
module: KB workflow
problem_type: workflow_issue
component: development_workflow
severity: medium
applies_when:
  - A recurring or long-running KB goal needs future runs to improve from durable feedback.
tags: [kb-goal, learning-loop, steering-memory, control-loop, feedback-classification]
---

# Live Steering Complements Post-Work Learning

## Context

The HumanLayer `design-control-loop` skill exposed a useful gap in KB: KB was
strong after work completed (`kb-complete`, `learn`, `evolve`, memory refresh),
but it had weaker in-flight steering for repeated work. A scheduled or repeated
loop needs durable feedback before the next run, not only post-hoc instincts
after completion.

## Guidance

Adopt the loop shape, not the runner:

- Use `kb-goal` to record an optional live-steering block for recurring,
  scheduled, or trend-improvement goals.
- Name the set point, sensor, controller, actuator, disturbances, optional
  dampener, scope gate, batch size, WIP bound, and steering-memory path.
- Keep steering memory concise and curated. It is for permanent scope
  exclusions, known false positives, reviewer preferences, and target-selection
  guidance.
- Classify feedback before learning: `current-only`, `steering-memory`,
  `observation`, `landmine-candidate`, or `instinct-evidence`.
- Keep `kb-complete` as the terminal learning and proof gate. Live steering
  changes future run selection and prompting; it does not replace review,
  proof, compounding, scored instincts, evolution, or map refresh.

## Why This Matters

Without a live steering layer, repeated agent work has two bad choices: forget
reviewer feedback until completion, or over-promote one-off comments into
project-wide instincts. Steering memory provides the middle layer. It changes
the next run while preserving the stronger evidence gates for instincts,
landmines, and generated skills.

## When to Apply

- A `kb-goal` objective spans multiple runs or days.
- The work is directional, such as reducing a class of issues over time.
- The next target should be selected from a measurement, lint report, review
  signal, or backlog.
- Maintainer feedback should affect future selections without becoming a
  permanent project convention yet.

Do not apply this to ordinary one-shot fixes or feature slices.

## Examples

Use the goal ledger for simple loops:

```markdown
## Live Steering (optional)

- Set point: no new cross-boundary imports in package A
- Sensor: `go run ./cmd/kbcheck boundary-report`
- Controller: select one high-confidence violation per run
- Actuator: `kb-plan` -> `kb-work` -> `kb-complete`
- Scope gate: `pkg/a/**`, read-only outside package A
- Batch size: 1 violation
- WIP bound: 1 active manifest
- Steering memory: this goal ledger
```

Create `docs/context/operations/steering/<slug>.md` only when the guidance would
bloat the goal ledger or should outlive the active goal.

## Related

- `docs/context/goals/live-steering-learning-loop.md`
- `docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md`
- `docs/context/architecture/kb-workflow.md`
