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

✅ done warning-quality cleanup completed and reviewed.

Audit note: `docs/context/research/2026-05-29-skill-repo-gap-audit.md`
Requirements: `docs/brainstorms/2026-05-29-cross-runtime-skill-quality-requirements.md`
Manifest: `docs/plans/2026-05-29-000-kb-cross-runtime-skill-quality-manifest.md`
Live eval requirements: `docs/brainstorms/2026-05-30-live-cross-runtime-skill-eval-harness-requirements.md`
Live eval manifest: `docs/plans/2026-05-30-000-kb-live-cross-runtime-skill-eval-harness-manifest.md`
Skill minimalism epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`
Proof/pipeline manifest: `docs/plans/2026-05-31-000-kb-proof-pipeline-spike-manifest.md`
Learning/landmines manifest: `docs/plans/2026-05-31-010-kb-learning-landmines-manifest.md`
Routing/trim manifest: `docs/plans/2026-05-31-020-kb-routing-trim-manifest.md`
Lazy-lane manifest: `docs/plans/2026-05-31-030-kb-lazy-lane-consolidation-manifest.md`

## Current Truth

- This repo is the working source for portable skills under `.github/skills/`.
- Personal/global installs currently match this repo for KB skills.
- ATV scaffold/plugin copies are intentionally optional thinner bundles by default.
- Deterministic skill lint, route-complexity fixtures, captured-result scoring, trace rule scoring, claim verification, output-quality rubric checks, regression reporting, sync drift checks, and Codex/GHCP live adapters exist. Live model calls remain explicit and outside `kb-check -All`.

## Active Work

None.

## Handoff Queue

| Handoff | Status | Route | Created | Stale Check | Link |
|---|---|---|---|---|---|

## Human Required

- None.

## Parked / Cold Storage

- Build cross-model benchmark prompts for route selection, complexity, and proof discipline.
- Add path-specific `.github/instructions/*.instructions.md` for Copilot if the workflow starts editing multiple file classes with different rules.

## Blocked

None.

## Work Log

- 2026-05-29: Completed cross-runtime skill quality manifest. `kb-check -All` now runs skill lint, route-complexity evals, and read-only sync drift report for Codex and GHCP.
- 2026-05-30: Completed `kb-eval-map` manifest. Bootstrap now invokes repo-native eval mapping; required Codex/Copilot/agents/ATV skill copies are synced; proof: `kb-check -All` and `git diff --check` passed.
- 2026-05-30: Added deterministic `skill-eval` scorer for captured skill result JSON. `kb-check -All` now self-tests route/proof/claim failures before sync drift.
- 2026-05-30: Added Codex live skill eval adapter. `scripts/skill-eval-run-codex.ps1` runs route fixtures through `codex exec`, captures schema JSON, and scores it with `skill-eval`; dry-run is included in `kb-check -All`.
- 2026-05-30: Planned the remaining live cross-runtime eval harness: GHCP adapter, live corpus runner, trace/claim scoring, output quality, cost regression, and eval-map negative validation.
- 2026-05-30: Completed the live cross-runtime eval harness. GHCP adapter, corpus runner, trace/claim scoring, output-quality selftests, regression reports, and eval-map scaffold negative-validation are implemented and documented.
- 2026-05-31: Planned the skill minimalism/proof harness epic into four manifests: proof pipeline spike, learning landmines, routing trim, and lazy lane consolidation.
- 2026-05-31: Completed and reviewed the skill minimalism/proof harness epic. Proof: `kb-check -All`, `git diff --check`, and required sync report passed; one baseline-regression review finding was fixed.
- 2026-05-31: Completed warning-quality cleanup. Missing `argument-hint` warnings were removed, review local fallback was codified, and optional sync drift output was compacted by default.
