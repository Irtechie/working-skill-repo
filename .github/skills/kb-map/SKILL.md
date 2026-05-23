---
name: kb-map
description: Project-memory setup, lookup, and refresh skill for KB workflows. Use when starting in a repo, when another KB skill needs the right architecture docs without broad repo crawling, or when the user says "map", "project context", "where is this", "what should I read", "setup memory", or "bootstrap context". If memory is missing or badly stale, invokes kb-map-bootstrap.
argument-hint: "[lookup|refresh] [optional task or subsystem]"
---

# KB Map

Use local project memory so fresh sessions do not need the user to reteach the app.

This skill owns the project-memory preflight. `kb-start` and other skills should call `kb-map` instead of checking bootstrap rules themselves.

Keep normal lookup cheap. Deep indexing belongs in `kb-map-bootstrap`.

## Project Root Rule

Anchor every lookup to the active project root before reading memory.

1. Determine the project root:
   - Prefer the current working directory's Git root: `git rev-parse --show-toplevel`.
   - If Git is unavailable, use the current working directory only when it is clearly a project directory.
   - Treat drive roots such as `E:\`, home directories such as `~`/`%USERPROFILE%`, `.copilot`, `.codex`, and `.agents` as invalid project roots unless the user explicitly chose them.
   - If the resolved root is invalid or not the user's intended project, ask for the project path before searching.
2. Read memory only from that root:
   - `<repo>/todo.md`
   - `<repo>/docs/context/PROJECT.md`
   - `<repo>/docs/handoffs/**`
3. Do not search `~`, `%USERPROFILE%`, `.copilot/handoffs`, the whole drive, or sibling repos for KB memory unless the user explicitly asks for cross-repo/global lookup.
4. If the project root has no KB memory, invoke `kb-map-bootstrap` in that project root. Do not silently substitute a global or unrelated handoff.

This prevents the agent from picking up stale personal handoffs when the user is working inside a specific repo.

Forbidden fallback: do not use glob/search to find `todo.md`, `PROJECT.md`, or handoffs when the project root is unresolved. Resolve the root first or ask the user.

## Contract Check

Before lookup or refresh, check the standard layout.

- Standard memory files must be checked by exact path under the project root, not by grep/glob:
  - `<repo>/todo.md`
  - `<repo>/docs/context/PROJECT.md`
  - `<repo>/docs/handoffs/active/`
  - `<repo>/docs/handoffs/parked/`
  - `<repo>/docs/handoffs/done/`
- If `todo.md` or `docs/context/PROJECT.md` is missing, invoke `kb-map-bootstrap`.
- If only directories are missing, create them during `refresh`.
- Never overwrite non-empty user docs without reading and merging.
- After bootstrap or refresh, continue the original lookup so the caller receives route-ready context.

Exact-path example on Windows:

```powershell
$root = git rev-parse --show-toplevel
if (-not $root -or $root -match '^[A-Za-z]:\\?$') { throw "Project root required" }
Test-Path (Join-Path $root 'todo.md')
Test-Path (Join-Path $root 'docs/context/PROJECT.md')
Get-ChildItem (Join-Path $root 'docs/handoffs') -Recurse -File -ErrorAction SilentlyContinue
```

## Modes

| Mode | Use When | Cost |
|---|---|---|
| `preflight` | Another skill needs memory verified before routing | low to high only if bootstrap is needed |
| `lookup` | Memory exists; find the right context for the current request | low |
| `refresh` | Recent work changed project memory or route pointers | medium |
| `setup` | User explicitly wants memory initialized | high; delegates to `kb-map-bootstrap` |

Default to `lookup`.

## Standard Layout

```text
todo.md
todo-done.md
docs/context/
  PROJECT.md
  architecture/
    README.md
    <major-subsystem>.md
  research/
    README.md
    <topic>.md
  decisions/
    README.md
  operations/
    README.md
    testing.md
docs/handoffs/
  active/
  parked/
  done/
```

## Lookup Mode

Read in order:

1. `todo.md`.
2. `docs/context/PROJECT.md`.
3. Active handoff files linked from `todo.md`.
4. Only the subsystem, research, decision, operation, brainstorm, or plan files needed for the request.

Stop reading once you can answer:

- What app/repo is this?
- What is active, blocked, parked, or queued?
- Which subsystem is relevant?
- Which files or commands are likely involved?
- What was already tried or researched?
- Which KB lane should handle the request?

Report route, docs loaded, and any stale-work refresh needed. Do not bulk-load all context docs.

Do not use `rg`, glob, or whole-repo search to find the standard memory files. Use search only after the exact project-root memory files are loaded and only for task-specific context.

## Missing Memory and Setup

If `todo.md` or `docs/context/PROJECT.md` is missing, invoke `kb-map-bootstrap`.

If handoff directories are missing but the project map exists, create or recommend the missing directories during `refresh`; do not deep-crawl the repo.

Use `setup` when the user explicitly wants to initialize KB memory. It always delegates to `kb-map-bootstrap` unless the standard layout already exists, in which case run `refresh`.

`kb-map-bootstrap` is the expensive first-pass mapper. `kb-map` is the durable entry point that decides whether bootstrap is needed.

## Refresh Mode

Use after meaningful architecture, workflow, or project-memory changes.

Refresh is required when work changes:

- User-visible behavior, feature boundaries, or major workflows.
- API contracts, data models, storage, auth, permissions, routing, streaming, tools, actions, jobs, or integrations.
- Build, run, test, deploy, or QA commands.
- Subsystem ownership, entry points, or first files a fresh session should read.
- Known sharp edges, rejected approaches, or "do not repeat" lessons.

Refresh is usually not required for:

- Pure styling, copy, formatting, lint-only changes, dependency lockfile churn, or isolated tests that do not change behavior.

When unsure, write a one-line manifest or `todo.md` note explaining why refresh was skipped or required.

Workflow:

1. Read `docs/context/PROJECT.md`.
2. Inspect changed files, recent manifests, active handoffs, and `todo.md`.
3. Update only affected subsystem docs and indexes.
4. Add child docs when a parent doc grows too large.
5. Update `todo.md` if active state, blockers, or pointers changed.
6. Update active handoff files if restart instructions changed.
7. Run `document-review` when changes are substantial.

Do not re-bootstrap the whole repo here.

## Contracts

`PROJECT.md` is a route map, not an encyclopedia. Subsystem docs carry durable app truth. `todo.md` carries current operational truth. Handoff files carry resumable work packets.
