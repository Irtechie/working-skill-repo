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

Skill bundle hardening plus the markdown/runtime contract extraction is locally complete; GitHub workflow changes are intentionally omitted from this push. The canonical contributor and release gates are
`go run ./cmd/kbcheck core` and `go run ./cmd/kbcheck local-release`.

Audit note: `docs/context/research/2026-05-29-skill-repo-gap-audit.md`
Requirements: `docs/brainstorms/2026-05-29-cross-runtime-skill-quality-requirements.md`
Manifest: `docs/plans/archive/2026-05/2026-05-29-000-kb-cross-runtime-skill-quality-manifest.md`
Live eval requirements: `docs/brainstorms/2026-05-30-live-cross-runtime-skill-eval-harness-requirements.md`
Live eval manifest: `docs/plans/archive/2026-05/2026-05-30-000-kb-live-cross-runtime-skill-eval-harness-manifest.md`
Skill minimalism epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`
Proof/pipeline manifest: `docs/plans/archive/2026-05/2026-05-31-000-kb-proof-pipeline-spike-manifest.md`
Learning/landmines manifest: `docs/plans/archive/2026-05/2026-05-31-010-kb-learning-landmines-manifest.md`
Routing/trim manifest: `docs/plans/archive/2026-05/2026-05-31-020-kb-routing-trim-manifest.md`
Lazy-lane manifest: `docs/plans/archive/2026-05/2026-05-31-030-kb-lazy-lane-consolidation-manifest.md`
ATV resync epic: `docs/context/epics/atv-upstream-resync.md`
ATV resync manifest: `docs/plans/archive/2026-05/2026-05-31-070-kb-atv-upstream-resync-manifest.md`
Claude remaining hardening epic: `docs/context/epics/claude-remaining-hardening.md`
Claude remaining hardening manifest: `docs/plans/archive/2026-06/2026-06-01-080-kb-claude-remaining-hardening-manifest.md`
Go validator replacement epic: `docs/context/epics/go-native-validator-port.md`
Go validator full replacement manifest: `docs/plans/archive/2026-06/2026-06-01-130-kb-go-validator-full-replacement-manifest.md`

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
| Session model routing advisory pilot | 🔧 in_progress | P0 | `docs/plans/2026-07-10-030-kb-session-model-routing-manifest.md` |
| Plan-to-PR finish lane | 🔧 in_progress | P0 | `docs/plans/2026-07-09-020-kb-plan-to-pr-finish-manifest.md` |
| One-loop AMR cleanup, completion review, and image refresh | 🔧 in_progress | P0 | `docs/plans/2026-07-10-030-kb-session-model-routing-manifest.md` |
| GHCP AIC/context falsification harness | 🔧 in_progress | P1 | `docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md` — session-routing baseline landed via PR #1; deterministic no-paid work authorized |
| Invisible workflow UX: retire KLFG, preserve consent, report route/proof/AIC | 🔒 blocked | P1 | `docs/plans/2026-07-11-045-kb-invisible-workflow-ux-manifest.md` — waits for overlapping AMR/AIC work to stabilize and explicit plan-to-work consent |

## Queued Improvements

- ⬜ Turn Dex/HumanLayer control-loop research into a KB recurring-loop plan —
  design a repo-local sensor/controller/actuator generator with PR-bound
  `/iterate` steering memory, context-efficient check output, and safe
  worktree/session adapter criteria. Research:
  `docs/context/research/2026-07-05-dexhorthy-humanlayer-agent-harness-research.md`.
- ⬜ Add runtime hook enforcement for the workflow governor — implement Codex
  and/or Claude hook files that mirror the Question Gate and gate-ledger checks,
  block stop/phase advancement when the artifact says blocked, and prove the
  hooks with deterministic selftests instead of claiming hook enforcement from
  skill text alone.
- ⬜ Continue markdown-to-runtime extraction — move remaining deterministic
  hot-path skill rules into `kbcheck` checks; keep `SKILL.md` for judgment,
  scope, escalation, and tradeoffs.
- ⬜ Add command-aware `kbcheck` failure summarizers after the compact core
  profile proves stable; preserve `--verbose` for raw output.
- ⬜ Complete graphify/TokenMasterX brainstorm and plan — finish the large-repo
  map/bootstrap graph-routing work already started in `kb-map` and
  `kb-map-bootstrap`; decide thresholds, proof, install prerequisites,
  TokenMasterX vs raw graphify boundaries, and sync/update path before shipping.

## Handoff Queue

| Handoff | Status | Route | Created | Stale Check | Link |
|---|---|---|---|---|---|
| One-loop AMR cleanup and image gate | active | `kb-work` | 2026-07-10 | fresh | `docs/handoffs/active/2026-07-10-sol-amr-ddr-completion-review.md` |

## Human Required

None.

## Parked / Cold Storage

- H2 controlled KB workflow experiment draft: `docs/brainstorms/2026-06-10-h2-controlled-kb-experiment.md` stays parked for human review; no harness changes authorized.
- Deletion/trim decisions for remaining cold-storage candidates stay parked
  until a dedicated trim/deletion pass reviews the new evidence classes.
- Live cross-model benchmark execution is parked; fixtures and deterministic
  schema validation now exist, but live model calls remain explicit.

## Blocked

- Invisible workflow UX execution is blocked on overlapping writes in the
  session-routing and GHCP AIC workstreams. Recheck the actual diff after those
  settle, then ask before starting `kb-work`.

## Work Log

Completed work is archived in `todo-done.md`.
