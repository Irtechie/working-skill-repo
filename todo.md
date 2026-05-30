# Todo

## Rules

**This file is the single source of truth for active skill-bundle work.** Keep it small; move finished summaries to `todo-done.md`.

Status markers:

| Marker | Meaning |
|---|---|
| ⬜ pending | Ready when blockers clear |
| 🔧 in_progress | Agent claimed and actively working |
| ✅ done | Complete and verified |
| 🔒 blocked | Cannot proceed; explain under Blocked |
| ⊘ skipped | Intentionally skipped with reason |
| 🛑 human-required | Needs human input or approval |

Promotion rules:

- Newly discovered work goes to Parked / Cold Storage unless explicitly in scope.
- Refresh parked work older than 72 hours before execution.
- Use `docs/context/PROJECT.md` as the fresh-session map.
- Use `docs/context/memory-maintenance.md` for map/doc quality issues.

## Objective

Make this the highest-reliability portable skill bundle for the user's workflow: low ceremony, complexity-aware routing, autonomous verification, fresh-session recovery, and drift-safe propagation.

## Current Focus

⬜ pending decide ATV scaffold/plugin shipping contract for optional KB skills.

Audit note: `docs/context/research/2026-05-29-skill-repo-gap-audit.md`
Requirements: `docs/brainstorms/2026-05-29-cross-runtime-skill-quality-requirements.md`
Manifest: `docs/plans/2026-05-29-000-kb-cross-runtime-skill-quality-manifest.md`

## Current Truth

- This repo is the working source for portable skills under `.github/skills/`.
- Personal/global installs currently match this repo for KB skills.
- ATV scaffold/plugin copies are not fully aligned with the KB skill surface.
- Deterministic skill lint, route-complexity fixtures, sync drift checks, and captured-result scoring exist; Codex has a live route-fixture adapter, while GHCP and broader live corpora remain planned.

## Active Work

### Skill Repo Quality Follow-Up

Task ID: skill-repo-audit-2026-05-29
Ready: yes
Validation: `git diff --check`; hash/drift probe; web-source scan; repo inventory.

| Gap | Status | Priority | Link |
|---|---|---:|---|
| Add deterministic skill lint/eval harness | ✅ done | P0 | `docs/context/research/2026-05-29-skill-repo-gap-audit.md` |
| Add route-complexity rubric/evals to prevent over/under-planning | ✅ done | P0 | `docs/context/research/2026-05-29-skill-repo-gap-audit.md` |
| Resolve portable-bundle memory exception vs bootstrap requirement | ✅ done | P1 | `docs/context/memory-maintenance.md` |
| Decide ATV scaffold/plugin propagation contract for KB skills | ⬜ pending | P1 | `docs/context/memory-maintenance.md` |

## Handoff Queue

| Handoff | Status | Route | Created | Stale Check | Link |
|---|---|---|---|---|---|

## Human Required

- Decide whether ATV scaffold/plugin should ship the full KB surface or explicitly remain a thinner compatibility bundle. The new sync report treats those targets as optional until this is decided.

## Parked / Cold Storage

- Build cross-model benchmark prompts for route selection, complexity, and proof discipline.
- Build GHCP adapter and broader live Codex/GHCP corpora that run route fixtures and emit `evals/skill-eval` result JSON.
- Add path-specific `.github/instructions/*.instructions.md` for Copilot if the workflow starts editing multiple file classes with different rules.

## Blocked

None.

## Work Log

- 2026-05-29: Completed cross-runtime skill quality manifest. `kb-check -All` now runs skill lint, route-complexity evals, and read-only sync drift report for Codex and GHCP.
- 2026-05-30: Completed `kb-eval-map` manifest. Bootstrap now invokes repo-native eval mapping; required Codex/Copilot/agents/ATV skill copies are synced; proof: `kb-check -All` and `git diff --check` passed.
- 2026-05-30: Added deterministic `skill-eval` scorer for captured skill result JSON. `kb-check -All` now self-tests route/proof/claim failures before sync drift.
- 2026-05-30: Added Codex live skill eval adapter. `scripts/skill-eval-run-codex.ps1` runs route fixtures through `codex exec`, captures schema JSON, and scores it with `skill-eval`; dry-run is included in `kb-check -All`.
