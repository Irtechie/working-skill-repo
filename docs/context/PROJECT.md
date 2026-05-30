# Project Map

Bootstrap: 2026-05-29
Bootstrap confidence: mixed

## What This Is

Portable skill bundle for KB workflow skills, reviewer agents, root `AGENTS.md`, Copilot instructions, and helper scripts. The repo is the working source that syncs into personal/global installs and selected ATV copies.

## How To Run

There is no app runtime.

Use repo inspection and deterministic scripts:

```powershell
git status --short
Get-ChildItem .\.github\skills -Directory
.\.github\skills\kb-check\scripts\kb-check.ps1 -List
.\.github\skills\kb-check\scripts\kb-check.ps1 -All
git diff --check
```

## How To Test

Primary testing today is structural:

- `git diff --check`
- skill inventory and frontmatter checks
- hash drift comparison across install targets
- targeted route/eval prompts run manually against scratch repos

The canonical quality command is now:

```powershell
.\.github\skills\kb-check\scripts\kb-check.ps1 -All
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
| 2026 quality audit | `docs/context/research/2026-05-29-skill-repo-gap-audit.md` | Understanding current gaps and recommended priorities | verified |
| Drift and propagation | `AGENTS.md`; `README.md`; `docs/context/memory-maintenance.md` | Syncing skills across global installs and ATV copies | mixed |

## Current Work Pointers

- Active board: `todo.md`
- Audit note: `docs/context/research/2026-05-29-skill-repo-gap-audit.md`
- Maintenance signals: `docs/context/memory-maintenance.md`

## Known Sharp Edges

- Portable repo hygiene conflicts with the normal KB bootstrap requirement unless this repo's own memory is treated as skill-bundle maintenance.
- `kb-check` now finds skill-repo checks through `config/skill-quality.json`; optional ATV scaffold/plugin differences remain warnings until their distribution contract is decided.
- `kb-eval-map` is now the bootstrap-owned setup skill for repo-native eval surfaces; consuming repos still need their own `docs/context/eval-map.md`.
- Some skills are long enough to make route-start context expensive; lazy references are used inconsistently.
- ATV scaffold/plugin copies are partially missing KB skills or contain older inherited skill variants.

## Research Index

- `docs/context/research/2026-05-29-skill-repo-gap-audit.md`
- `docs/context/eval-map.md`

## Do Not Repeat

- Do not bootstrap consuming-project memory inside this repo.
- Do not sync over global or ATV copies without reviewing drift first.
- Do not remove reviewer agents until an eval proves no workflow dispatches them.

## Maintenance Notes

Use `docs/context/memory-maintenance.md` for stale map, drift, and skill-quality gaps.
