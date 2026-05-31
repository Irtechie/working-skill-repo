# Narrow Lane Trim Requirements

Status: draft
Created: 2026-05-31
Epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`

## Problem

Narrow lane skills are valuable only when their trigger is specific. If they
carry generic workflow guidance, they increase discovery and loaded-surface
noise without improving outcomes.

## Decision

Keep narrow lanes when they encode a distinct gate, proof strategy, or failure
mode. Merge or delete only after reference integrity and route fixtures show no
breakage.

## Scope

Narrow lane candidates:

- `kb-fix`
- `kb-troubleshoot`
- `kb-research`
- `kb-ship`
- `kb-functional-test`
- `kb-regression-snapshot`
- `kb-repair`
- `kb-memory-review`
- `kb-eval-map`

## Requirements

- Keep `kb-fix` for small known bugs.
- Keep `kb-troubleshoot` for unknown broken behavior and root-cause loops.
- Keep `kb-research` when reusable external or prior-art research changes
  direction.
- Keep `kb-functional-test` for proof-level classification.
- Keep `kb-regression-snapshot` if it provides real baseline replay; otherwise
  fold into the harness proof workstream.
- Keep `kb-eval-map` as repo-native eval setup/mapping.
- Avoid duplicating `kb-check` proof rules in every narrow lane.

## Resolve Before Planning

- Decide whether `kb-regression-snapshot` remains a standalone skill after
  persisted eval baselines are added.
- Decide whether `kb-functional-test` should be a standalone skill or a lazy
  reference used by `kb-work`/`kb-check`.

## Slice Candidates

- Add route fixture coverage for each retained narrow lane.
- Trim duplicated proof language from narrow lanes.
- Merge any lane that only rephrases `kb-check` or `kb-work`.
- Add deletion safety tests before removing any narrow lane.
