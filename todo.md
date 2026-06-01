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

No active work claimed. Cold storage follow-through is complete. The canonical
quality and release gates are now `go run .\cmd\kbcheck core` and
`go run .\cmd\kbcheck local-release`.

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
ATV resync epic: `docs/context/epics/atv-upstream-resync.md`
ATV resync manifest: `docs/plans/2026-05-31-070-kb-atv-upstream-resync-manifest.md`
Claude remaining hardening epic: `docs/context/epics/claude-remaining-hardening.md`
Claude remaining hardening manifest: `docs/plans/2026-06-01-080-kb-claude-remaining-hardening-manifest.md`

## Current Truth

- This repo is the working source for portable skills under `.github/skills/`.
- Personal/global installs and tracked ATV skill roots are expected to match this repo for KB skills.
- ATV scaffold/plugin copies are no longer intentionally thin for tracked skills.
- Deterministic skill lint, route-complexity fixtures, captured-result scoring, observed trace scoring, claim verification, computed output-quality checks, regression reporting, sync drift checks, and Codex/GHCP live adapters exist. Live model calls remain explicit and outside `kb-check -All`.
- `cmd/kbcheck` provides the native Go top-level gate for `core`,
  `local-release`, and `live-release`. Several individual validators remain
  PowerShell scripts until separately ported.
- `skill-surface-minimality.ps1` has a protected classification so repo-policy
  dependencies such as `ce-review`, `ce-compound`, `ce-compound-refresh`, and
  `document-review` do not become deletion candidates from static analysis
  alone.
- ATV upstream resync must be category-merged from original `All-The-Vibes/ATV-StarterKit` `upstream/main`, not the fork. KB is preserved as this repo's overlay; original ATV is a source to mine, not a mirror target. Superseded workflow skills (`lfg`, `slfg`, `workflows-*`, CE workflow aliases replaced by KB lanes) stay out unless the app uses them or a focused porting plan proves value. The useful upstream `ce-review` mechanics are already present in local references.

## Active Work

| Workstream | Status | Priority | Link |
|---|---|---|---|
| Claude remaining hardening | ✅ done | P1 | `docs/plans/2026-06-01-080-kb-claude-remaining-hardening-manifest.md` |

## Queued Improvements

## Handoff Queue

| Handoff | Status | Route | Created | Stale Check | Link |
|---|---|---|---|---|---|

## Human Required

None.

## Parked / Cold Storage

- Deletion/trim decisions for remaining cold-storage candidates stay parked
  until a dedicated trim/deletion pass reviews the new evidence classes.
- Live cross-model benchmark execution is parked; fixtures and deterministic
  schema validation now exist, but live model calls remain explicit.
- macOS/Linux proof for the Go gate is parked until those environments are
  available. Windows parity is recorded in
  `docs/reports/go-gate-parity-2026-06-01.md`.

## Blocked

None.

## Work Log

- 2026-05-31: Hardened future-work gaps: computed output-quality scoring from raw result JSON, collapsed `kb-start` to one ranked routing list, and made harness subprocesses prefer PowerShell 7 (`pwsh`) with Windows PowerShell fallback. Proof: `kb-check -All`, `skill-sync-report`, and `git diff --check` passed.
- 2026-05-31: Planned ATV upstream resync as a six-slice epic. Original ATV `upstream/main` is authoritative for inspecting ATV-native changes; KB-owned skills are preserved as this repo's overlay, shared overlap skills get three-way review, and superseded workflow skills stay out unless currently used.
- 2026-05-31: Completed ATV upstream resync correction. Backed out transient `lfg`, `slfg`, and `workflows-*` imports, rejected upstream KB deletions and the `atv-security` OSV regression, kept globals clean, and confirmed useful upstream `ce-review` mechanics are already present locally. Proof: `kb-check -All`, `skill-sync-report`, working/ATV/marketplace `git diff --check`, focused review-skill diff, and direct `Test-Path` checks for removed workflow dirs.
- 2026-05-29: Completed cross-runtime skill quality manifest. `kb-check -All` now runs skill lint, route-complexity evals, and read-only sync drift report for Codex and GHCP.
- 2026-05-30: Completed `kb-eval-map` manifest. Bootstrap now invokes repo-native eval mapping; required Codex/Copilot/agents/ATV skill copies are synced; proof: `kb-check -All` and `git diff --check` passed.
- 2026-05-30: Added deterministic `skill-eval` scorer for captured skill result JSON. `kb-check -All` now self-tests route/proof/claim failures before sync drift.
- 2026-05-30: Added Codex live skill eval adapter. `scripts/skill-eval-run-codex.ps1` runs route fixtures through `codex exec`, captures schema JSON, and scores it with `skill-eval`; dry-run is included in `kb-check -All`.
- 2026-05-30: Planned the remaining live cross-runtime eval harness: GHCP adapter, live corpus runner, trace/claim scoring, output quality, cost regression, and eval-map negative validation.
- 2026-05-30: Completed the live cross-runtime eval harness. GHCP adapter, corpus runner, trace/claim scoring, output-quality selftests, regression reports, and eval-map scaffold negative-validation are implemented and documented.
- 2026-05-31: Planned the skill minimalism/proof harness epic into four manifests: proof pipeline spike, learning landmines, routing trim, and lazy lane consolidation.
- 2026-05-31: Completed and reviewed the skill minimalism/proof harness epic. Proof: `kb-check -All`, `git diff --check`, and required sync report passed; one baseline-regression review finding was fixed.
- 2026-05-31: Completed warning-quality cleanup. Missing `argument-hint` warnings were removed, review local fallback was codified, and optional sync drift output was compacted by default.
- 2026-06-01: Planned Claude remaining hardening into release confidence, skill-surface minimality classification, read-only upstream delta report, and parked Go wrapper plus trim/deletion queues. Manifest: `docs/plans/2026-06-01-080-kb-claude-remaining-hardening-manifest.md`.
- 2026-06-01: Completed Claude remaining hardening. Added release confidence gate profiles, skill/agent minimality classification, read-only ATV upstream delta reporting, and fixed a concurrent baseline-selftest temp-directory collision found during completion. Proof: `kb-check -All`, `kb-release-gate.ps1 -Profile local-release`, `skill-sync-report`, protected selftests, and `git diff --check` passed.
