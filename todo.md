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
quality and release gates are now `go run ./cmd/kbcheck core` and
`go run ./cmd/kbcheck local-release`.

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
Go validator replacement epic: `docs/context/epics/go-native-validator-port.md`
Go validator full replacement manifest: `docs/plans/2026-06-01-130-kb-go-validator-full-replacement-manifest.md`

## Current Truth

- This repo is the working source for portable skills under `.github/skills/`.
- Personal/global installs and tracked ATV skill roots are expected to match this repo for KB skills.
- ATV scaffold/plugin copies are no longer intentionally thin for tracked skills.
- Deterministic skill lint, route-complexity fixtures, captured-result scoring,
  observed trace scoring, claim verification, computed output-quality checks,
  regression reporting, sync drift checks, marketplace firebreak/promotion
  checks, ATV delta reporting, and Codex/GHCP live adapters exist. Live model
  calls remain explicit and outside the default `core` gate.
- `cmd/kbcheck` provides the native Go gate for `core`, `local-release`,
  `live-release`, eval adapters, marketplace promotion, and drift reports.
  Remaining `.ps1` files are narrow skill helpers, not top-level gate
  dependencies.
- `kbcheck minimality` has a protected classification so repo-policy
  dependencies such as `ce-review`, `ce-compound`, `ce-compound-refresh`, and
  `document-review` do not become deletion candidates from static analysis
  alone.
- ATV upstream resync must be category-merged from original `All-The-Vibes/ATV-StarterKit` `upstream/main`, not the fork. KB is preserved as this repo's overlay; original ATV is a source to mine, not a mirror target. Superseded workflow skills (`lfg`, `slfg`, `workflows-*`, CE workflow aliases replaced by KB lanes) stay out unless the app uses them or a focused porting plan proves value. The useful upstream `ce-review` mechanics are already present in local references.

## Active Work

| Workstream | Status | Priority | Link |
|---|---|---|---|
| Claude remaining hardening | ✅ done | P1 | `docs/plans/2026-06-01-080-kb-claude-remaining-hardening-manifest.md` |
| Go validator full replacement | ✅ done | P1 | `docs/plans/2026-06-01-130-kb-go-validator-full-replacement-manifest.md` |

## Queued Improvements

- ⬜ Add runtime hook enforcement for the workflow governor — implement Codex
  and/or Claude hook files that mirror the Question Gate and gate-ledger checks,
  block stop/phase advancement when the artifact says blocked, and prove the
  hooks with deterministic selftests instead of claiming hook enforcement from
  skill text alone.
- ⬜ Complete graphify/TokenMasterX brainstorm and plan — finish the large-repo
  map/bootstrap graph-routing work already started in `kb-map` and
  `kb-map-bootstrap`; decide thresholds, proof, install prerequisites,
  TokenMasterX vs raw graphify boundaries, and sync/update path before shipping.

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

## Blocked

None.

## Work Log

- 2026-06-03: Completed cross-platform adoption on-ramp. Added `npx` installer with core/full profiles, non-destructive backups, repo-local install, and Windows/macOS/Linux CI proof. Proof: installer smoke checks, `go test ./...`, `go run ./cmd/kbcheck core`, working/ATV `git diff --check`, and required sync report passed.
- 2026-06-01: Completed Go validator full replacement. Ported all remaining skill-repo validators, eval adapters, marketplace promotion/firebreak checks, ATV delta reporting, pipeline proof, ready-set/scope-lease utilities, release selftests, surface/minimality reports, and sync drift reports into `cmd/kbcheck`; deleted all `.ps1` files. Proof: `go test ./...`, `go run ./cmd/kbcheck core`, `go run ./cmd/kbcheck local-release --json`, `go run ./cmd/kbcheck ready-set --manifest docs\plans\2026-06-01-130-kb-go-validator-full-replacement-manifest.md --json`, `rg --files -g "*.ps1"`, and `git diff --check`.

Older completed work is archived in `todo-done.md`.
