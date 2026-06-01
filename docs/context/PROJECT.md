# Project Map

Bootstrap: 2026-05-29
Bootstrap confidence: mixed

## What This Is

Portable skill bundle for KB workflow skills, reviewer agents, root `AGENTS.md`, Copilot instructions, and helper scripts. The repo is the working source that syncs into personal/global installs and selected ATV copies. Original ATV `upstream/main` is a source to mine for useful ATV-native changes; this repo remains the source of truth for the KB overlay and its replacements.

## How To Run

There is no app runtime.

Use repo inspection and deterministic scripts:

```powershell
git status --short
Get-ChildItem .\.github\skills -Directory
go run .\cmd\kbcheck core --list
go run .\cmd\kbcheck core
go run .\cmd\kbcheck local-release
git diff --check
```

On macOS/Linux, install PowerShell 7 and run the same gates with `pwsh
-NoProfile -File <script>.ps1`. Child harness calls prefer `pwsh` and fall back
to Windows PowerShell only when needed. `cmd/kbcheck` is the native Go gate for
top-level orchestration; some individual validators still require PowerShell.

## How To Test

Primary testing today is structural:

- `git diff --check`
- `go test ./...` when the Go wrapper is touched
- skill inventory and frontmatter checks
- hash drift comparison across install targets
- targeted route/eval prompts run manually against scratch repos

The canonical quality command is now:

```powershell
go run .\cmd\kbcheck core
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
| Drift and propagation | `AGENTS.md`; `README.md`; `docs/context/memory-maintenance.md` | Syncing skills across global installs and ATV copies | mixed |
| ATV upstream resync | `docs/context/epics/atv-upstream-resync.md`; `docs/context/research/2026-05-31-atv-upstream-skill-delta.md` | Understanding original ATV imports, rejected KB deletions, and workflow skill handling | verified |

## Current Work Pointers

- Active board: `todo.md`
- Active landmines: `docs/context/landmines.md`
- Audit note: `docs/context/research/2026-05-29-skill-repo-gap-audit.md`
- Maintenance signals: `docs/context/memory-maintenance.md`

## Known Sharp Edges

- Portable repo hygiene conflicts with the normal KB bootstrap requirement unless this repo's own memory is treated as skill-bundle maintenance.
- `kb-check` now finds skill-repo checks through `config/skill-quality.json`; working, global, ATV `.github`, ATV scaffold, and ATV plugin skill roots are expected to stay hash-synced unless a deliberate packaging exception is recorded.
- `kb-eval-map` is now the bootstrap-owned setup skill for repo-native eval surfaces; consuming repos still need their own `docs/context/eval-map.md`.
- Some skills are long enough to make route-start context expensive; lazy references are used inconsistently.
- ATV scaffold/plugin copies are no longer intentionally thin for the tracked KB/CE skill set; `skill-sync-report` should show matches across all tracked roots.
- Original ATV `upstream/main` is authoritative for ATV-native changes to
  inspect, not a mirror target. Upstream KB deletions are rejected because KB is
  this repo's overlay; superseded workflow skills such as `lfg`, `slfg`, and
  `workflows-*` stay out unless the app uses them.
- Focused review-skill merge check found the useful upstream `ce-review`
  mechanics already present in local references. Keep KB caller names rather
  than reviving old CE/LFG entry points.
- `E:\agent-marketplace` is a private approved catalog, not a global install.
  Project-local learned skills and pipelines must prove reuse value before
  promotion; public imports land in quarantine first.
- `atv-security` is approved as the single trusted ATV security skill installed
  globally and mirrored in `E:\agent-marketplace\skills\atv-security`. Its
  pinned SHA256 is recorded in
  `E:\agent-marketplace\catalog\approved-skills.json`.
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
- Do not sync over global or ATV copies without reviewing drift first.
- Do not remove reviewer agents until an eval proves no workflow dispatches them.

## Maintenance Notes

Use `docs/context/memory-maintenance.md` for stale map, drift, and skill-quality gaps.
