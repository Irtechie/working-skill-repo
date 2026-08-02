# ATV Upstream Resync

Status: closed (superseded 2026-08-02)
Created: 2026-05-31
Last refreshed: 2026-08-02

## Closure

This epic is closed and no resync is planned. The bundle now owns its copied and
adapted skills outright: no ATV checkout, plugin, or upstream branch is an
install, sync, release, or delivery dependency, and the required sync targets
are Codex global, Copilot global, and shared agents global. Attribution for
prior art stays in README Credits. The content below is kept as historical
record of the original intent, not as pending work.

## Intent

Refresh the local ATV fork and portable skill bundle against current upstream
ATV without deleting the KB workflow system or blindly promoting upstream skills
that KB already replaced.

The resync is a merge, not an overwrite:

- KB-owned skills stay sourced from `<working-skill-repo>`.
- ATV-native skills can update from upstream after review when they are still
  part of the local app/workflow surface.
- Shared CE/learning skills require manual three-way comparison.
- Original ATV upstream is a source to mine, not a mirror target. Superseded
  workflow skills are kept out unless the app actually uses them or a concrete
  improvement is ported into the KB replacement.

## Success Criteria

- Upstream ATV changes are inspected from a clean integration surface before
  applying anything to dirty working trees.
- `kb-*`, `klfg`, `tdd`, and `todo-*` are preserved unless a dedicated KB
  replacement plan and eval proof exists.
- Shared overlap skills keep useful upstream fixes while preserving local
  token-trim and proof-gate improvements.
- Superseded upstream workflow skills such as `lfg`, `slfg`, and `workflows-*`
  remain out of globals and ATV roots unless a dedicated use case re-approves
  them. `klfg`, `kb-work`, `kb-plan`, and related KB lanes are the replacements.
- Review-related upstream improvements are not bulk-copied. The focused check
  confirmed the useful `ce-review` mechanics are already present in local
  references; keep KB caller names.
- ATV native skill refreshes are propagated only after `kb-check -All`,
  `skill-sync-report`, and `git diff --check` pass.

## Architecture Decisions

- Do not merge `upstream/main` directly into `<atv-repo>` while that
  worktree is dirty from skill propagation.
- Treat `origin/main` as the current Irtechie fork state and `upstream/main` as
  the latest ATV source to compare against.
- Treat upstream deletion of the KB family as non-authoritative for this repo.
  The portable skill bundle owns KB.
- Treat `upstream/main` as the authoritative source for ATV-native changes to
  inspect. Your fork is comparison context only, not the canonical source.
- Treat KB replacements as authoritative when they intentionally supersede ATV
  workflows (`ce-work` -> `kb-work`, `ce-plan` -> `kb-plan`, `lfg`/`slfg` ->
  `klfg`).

## Research

- Current ATV remotes:
  - `origin`: `https://github.com/Irtechie/ATV-StarterKit.git`
  - `upstream`: `https://github.com/All-The-Vibes/ATV-StarterKit.git`
- Fetched upstream head: `upstream/main` at `cbe5d07 docs: add comprehensive /atv-security user guide (#49)`.
- Current fork head: `origin/main` at `35b0925 Add OSV proof to ATV security skill`.
- Existing distribution research:
  `docs/context/research/2026-05-30-agent-skills-git-distribution.md`

## Workstreams

| Workstream | Brainstorm | Manifest | Status | Notes |
|---|---|---|---|---|
| Clean upstream integration audit | skipped-clear | `docs/plans/archive/2026-05/2026-05-31-070-kb-atv-upstream-resync-manifest.md` | planned | Create clean compare surface and machine-readable skill delta inventory. |
| Shared overlap skill merge | skipped-clear | `docs/plans/archive/2026-05/2026-05-31-070-kb-atv-upstream-resync-manifest.md` | planned | Three-way review `ce-compound`, `ce-compound-refresh`, `ce-review`, `document-review`, `evolve`, and `learn`. |
| KB preservation and propagation | skipped-clear | `docs/plans/archive/2026-05/2026-05-31-070-kb-atv-upstream-resync-manifest.md` | planned | Reject upstream KB deletions and reassert working repo as source for tracked KB skills. |
| ATV-native refresh | skipped-clear | `docs/plans/archive/2026-05/2026-05-31-070-kb-atv-upstream-resync-manifest.md` | planned | Review ATV-owned skills such as `atv-security`, git/release helpers, and docs skills; keep only used or clearly useful changes. |
| Superseded workflow cleanup | skipped-clear | `docs/plans/archive/2026-05/2026-05-31-070-kb-atv-upstream-resync-manifest.md` | planned | Remove transient `lfg`, `slfg`, and `workflows-*` imports; keep already-ported review mechanics in KB/CE skills. |
| Proof, docs, and release sync | skipped-clear | `docs/plans/archive/2026-05/2026-05-31-070-kb-atv-upstream-resync-manifest.md` | planned | Run gates, update memory/docs, and sync approved roots only. |

## Dependency Map

```text
Clean upstream integration audit
  -> Shared overlap skill merge
  -> ATV-native refresh
  -> Superseded workflow cleanup
  -> KB preservation and propagation
  -> Proof, docs, and release sync
```

Shared overlap, ATV-native refresh, and superseded workflow cleanup can be reviewed
independently after the clean delta inventory exists. Final propagation waits
for all accepted changes.

## Execution Queue

Runnable after planning:

1. `docs/plans/archive/2026-05/2026-05-31-070-kb-atv-upstream-resync-manifest.md`

## Human Checkpoints

No planning-blocking questions.

Assumptions recorded for execution:

- Preserve KB globally and in this portable repo.
- Do not auto-install upstream workflow skills globally.
- Do not keep `lfg`, `slfg`, or `workflows-*` merely because they exist on
  `upstream/main`; they are superseded unless the app uses them.
- Keep the useful upstream review mechanics that are already present in local
  `ce-review` references instead of reviving old workflow entry points.

## Parked / Blocked

- Non-upstream branch-only workflow material remains parked unless the user
  explicitly asks to revisit it.
- Superseded upstream workflow material remains parked unless a current app use
  case or focused porting plan proves it belongs.
- Any upstream skill whose purpose overlaps an existing KB skill must be reviewed
  before promotion.

## Completion Criteria

- Every accepted upstream change is traceable to a source commit/path.
- Every rejected upstream deletion or addition has a recorded rationale.
- `go run ./cmd/kbcheck core` exits 0.
- `go run ./cmd/kbcheck skill-sync-report` reports zero required issues.
- `git diff --check` exits 0 in `<working-skill-repo>` and `<atv-repo>`.
