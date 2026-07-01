# Live Steering Learning Loop

Status: complete
Created: 2026-07-01
Last updated: 2026-07-01

## Objective

Integrate the useful HumanLayer `design-control-loop` mechanics into KB's learning workflow without replacing KB completion, review, proof, or memory gates.

## Done Criteria

- KB long-running work can declare a set point, sensor, controller, actuator, batch size, scope gate, dampener, and WIP bound when that framing is useful.
- KB workflows have a durable steering-memory contract for feedback that should influence future runs.
- Feedback from review, iteration, or completion is classified as current-only, steering memory, observation, landmine, or instinct evidence.
- The implementation avoids importing HumanLayer-specific CI, Bun, CodeLayer, or GitHub Actions assumptions as KB defaults.
- Skill docs, project docs, and proof commands are updated in this repo only.

## Terminal Proof

- `go run ./cmd/kbcheck core`
- `git diff --check`
- Focused review of changed KB learning/control-loop skills and docs

## Current State

- Current artifact: `docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md`
- Next allowed action: none
- Last proof: `go run ./cmd/kbcheck core` passed checks=26; `git diff --check` passed; `complete-to-ship` gate passed

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|
| Mine HumanLayer control-loop mechanics | kb-start/kb-research | temp clone + local notes | done | Compared `design-control-loop` references against KB `learn`, `kb-complete`, and `kb-goal` |
| Plan KB live steering changes | kb-plan | `docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md` | done | `plan-to-work` gate passed |
| Implement steering memory and feedback classification | kb-work | `docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md` | done | `go run ./cmd/kbcheck core`; `git diff --check` |
| Verify and refresh memory | kb-complete | `docs/plans/2026-07-01-000-kb-live-steering-learning-loop-manifest.md` | done | `complete-to-ship` gate passed; solution note and learning ledger updated |

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|

## Notes

- Adopt the shape, not the toolchain: durable steering memory, `/iterate`-style feedback routing, control-loop vocabulary, WIP bounds, and optional dampeners.
- Keep KB's post-work pipeline: `kb-review`, proof gates, `ce-compound`, `learn`, `evolve`, `kb-map refresh`, and cleanup remain the terminal learning path.
- Local review fixed one P2 follow-up: avoid double-logging resolved P0/P1 review findings into `.atv/observations.jsonl`.
