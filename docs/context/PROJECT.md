# Project Map

Bootstrap: 2026-05-29
Bootstrap confidence: mixed

## What This Is

Portable standalone skill bundle for KB workflow skills, reviewer agents, root `AGENTS.md`, Copilot instructions, and native Go validation tooling. This repo is the source of truth and syncs only into Codex, Copilot, and shared-agent global installs.

## How To Run

There is no app runtime. Optional context providers may accelerate lookup later,
but they are not part of the required runtime, install path, or release gate.

Use repo inspection and deterministic scripts:

```shell
git status --short
ls .github/skills
go run ./cmd/kbcheck core --list
go run ./cmd/kbcheck core
go run ./cmd/kbcheck local-release
git diff --check
```

On macOS/Linux, use the same Go entrypoint. CI runs the Go package tests and
core gate on Windows, macOS, and Linux.

## How To Test

Primary testing today is structural:

- `git diff --check`
- `go test ./...` when the Go wrapper is touched
- skill inventory and frontmatter checks
- hash drift comparison across install targets
- targeted route/eval prompts run manually against scratch repos

The canonical quality command is now:

```shell
go run ./cmd/kbcheck core
```

It runs skill lint, route-complexity fixture validation, and read-only sync
drift reporting from the shared Codex/GHCP contract in
`config/skill-quality.json`.

## Current Architecture

See `docs/context/architecture/README.md`.

## Subsystem Index

| Area | Read This | Use When | Confidence |
|---|---|---|---|
| Skill bundle layout | `docs/context/architecture/README.md` | Locating skills, agents, scripts, and install targets | verified |
| Eval map | `docs/context/eval-map.md` | Understanding repo-native eval surfaces, canonical commands, and open eval gaps | verified |
| Testing and verification | `docs/context/operations/testing.md` | Finding current checks, eval-map ownership, and harness gaps | verified |
| Private skill marketplace | `docs/context/architecture/private-skill-marketplace.md`; `config/skill-marketplace.json` | Deciding when project-local learned skills or pipelines can be promoted into the private reusable catalog | verified |
| Landmines | `docs/context/landmines.md` | Checking active repo-specific traps with owner/fix/proof fields | verified |
| Epics | `docs/context/epics/` | Coordinating multi-workstream skill-bundle initiatives | verified |
| 2026 quality audit | `docs/context/research/2026-05-29-skill-repo-gap-audit.md` | Understanding current gaps and recommended priorities | verified |
| Agent Skills Git distribution | `docs/context/research/2026-05-30-agent-skills-git-distribution.md` | Deciding canonical source, global installs, ATV policy, and deterministic sync scripts | verified |
| Drift and propagation | `AGENTS.md`; `README.md`; `docs/context/memory-maintenance.md` | Syncing skills across global installs | verified |
| ATV upstream resync | `docs/context/epics/atv-upstream-resync.md`; `docs/context/research/2026-05-31-atv-upstream-skill-delta.md` | Understanding original ATV imports, rejected KB deletions, and workflow skill handling | verified |

## Current Work Pointers

- Active board: `todo.md`
- Active landmines: `docs/context/landmines.md`
- Audit note: `docs/context/research/2026-05-29-skill-repo-gap-audit.md`
- Maintenance signals: `docs/context/memory-maintenance.md`

## Known Sharp Edges

- Portable repo hygiene conflicts with the normal KB bootstrap requirement unless this repo's own memory is treated as skill-bundle maintenance.
- `kb-check` finds skill-repo checks through `config/skill-quality.json`; Codex, Copilot, and shared-agent global skill roots must stay hash-synced.
- `kb-eval-map` is now the bootstrap-owned setup skill for repo-native eval surfaces; consuming repos still need their own `docs/context/eval-map.md`.
- Some skills are long enough to make route-start context expensive; lazy references are used inconsistently.
- `kb-work` is now bounded-swarm oriented: independent ready slices may run in
  isolated contexts, but shared-checkout mutation and observed write overlap
  serialize. `expected_files` remains a forecast, not proof of disjointness.
- Go-native validator migration is complete for the skill-repo harness:
  `cmd/kbcheck` owns the default quality/release/eval/marketplace gates.
  `core` is contributor-safe and repo-local; `local-release` is the pre-sync
  gate that blocks on required sync drift. Remaining `.ps1` files are narrow
  installer or skill helpers, not top-level gate dependencies. See
  `docs/context/epics/go-native-validator-port.md`.
- Required global skill roots should hash-match before release.
- Planner economy uses vendor-neutral context packets validated by
  `kbcheck context-packet`; existing manifests, goal ledgers, run state, and
  proof traces remain the lifecycle state spine.
- `kbcheck provider-hygiene` rejects Phoenix activation and treats CCE as an
  optional adapter. `surface-report` reports base and conditional skill costs
  separately.
- `kb-finish` is the explicit plan-to-PR lane: it recovers missing
  plan/work/complete phases, then `kb-ship` commits intentional files, pushes a
  topic branch, and creates or updates a PR without merging.
- Original ATV `upstream/main` is authoritative for ATV-native changes to
  inspect, not a mirror target. Upstream KB deletions are rejected because KB is
  this repo's overlay; superseded workflow skills such as `lfg`, `slfg`, and
  `workflows-*` stay out unless the app uses them.
- Focused review-skill merge check found the useful upstream `ce-review`
  mechanics already present in local references. Keep KB caller names rather
  than reviving old CE/LFG entry points.
- `<agent-marketplace>` is a private approved catalog, not a global install.
  Project-local learned skills and pipelines must prove reuse value before
  promotion; public imports land in quarantine first.
- `atv-security` is approved as the single trusted ATV security skill installed
  globally and mirrored in `<agent-marketplace>\skills\atv-security`. Its
  pinned SHA256 is recorded in
  `<agent-marketplace>\catalog\approved-skills.json`.
- ATV security A06 dependency checks prefer OSV Scanner machine evidence when
  `osv-scanner` is installed. If unavailable, record `skipped-unavailable`
  rather than inventing vulnerability findings from version age alone.

## Research Index

- `docs/context/research/2026-05-29-skill-repo-gap-audit.md`
- `docs/context/research/2026-05-30-agent-skills-git-distribution.md`
- `docs/context/research/2026-05-31-atv-upstream-skill-delta.md`
- `docs/context/eval-map.md`

## Do Not Repeat

- Do not bootstrap consuming-project memory inside this repo.
- Do not sync over global copies without reviewing drift first.
- Do not remove reviewer agents until an eval proves no workflow dispatches them.

## Maintenance Notes

Use `docs/context/memory-maintenance.md` for stale map, drift, and skill-quality gaps.

## Learning Model

Learning is kb-native. Durable instincts are git-tracked under `docs/context/kb/`; ephemeral run artifacts are git-ignored under `.kb/`.

| Path | Tier | Tracked |
|---|---|---|
| `docs/context/kb/instincts/project.yaml` | project + global instincts | yes |
| `docs/context/kb/instincts/scoped/<scope>.yaml` | workflow/domain instincts | yes |
| `docs/context/kb/instincts/archive/` | decayed instincts | yes |
| `docs/context/kb/kb-completions.txt` | kb-complete counter | yes |
| `.kb/observations.jsonl` | passive observation feed | no |
| `.kb/snapshots/` | regression snapshots | no |

Scope hierarchy: `workflow/domain → project → global`. Default = narrowest owning scope. Pull = active scope + all ancestors, never siblings. Promotion = nearest common ancestor when the same lesson recurs across sibling scopes. Landmines = instant one-shot at owning scope.

**X pipeline's lessons are not visible to Y pipeline unless promoted to a shared ancestor.**

Canonical reference: `docs/context/architecture/kb-learning-model.md`.
