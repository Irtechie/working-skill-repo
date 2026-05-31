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
   - If the resolved root is invalid or not the user's intended project, ask the user to change into the project directory or provide the project path before searching.
2. Read memory only from that root:
   - `<repo>/todo.md`
   - `<repo>/docs/context/PROJECT.md`
   - `<repo>/docs/context/landmines.md` when it exists
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
if (-not $root -or -not (Test-Path $root) -or $root -match '^[A-Za-z]:\\?$') {
  throw "Project root required"
}
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
3. `docs/context/landmines.md` if it exists. Read only `Active Landmines`; do
   not load resolved/archive entries during startup.
4. Active handoff files linked from `todo.md`.
5. Only the subsystem, research, decision, operation, brainstorm, or plan files needed for the request.

Stop reading once you can answer:

- What app/repo is this?
- What is active, blocked, parked, or queued?
- Which subsystem is relevant?
- Which files or commands are likely involved?
- What was already tried or researched?
- Which KB lane should handle the request?
- Are there active repo-specific landmines that apply to this request?

Report route, docs loaded, and any stale-work refresh needed. Do not bulk-load all context docs.

Do not use `rg`, glob, or whole-repo search to find the standard memory files. Use search only after the exact project-root memory files are loaded and only for task-specific context.

## Coverage Gap Rule

If lookup cannot get a fresh session meaningfully up to speed on the requested
subsystem from `PROJECT.md` plus pointed docs, treat that as a project-memory
coverage gap.

Coverage is insufficient when the agent must rediscover basics by broad search,
for example:

- the relevant subsystem doc is missing or too generic;
- a broad parent doc exists but does not point to the child workflow;
- source-of-truth files, scripts, CI workflows, generated artifacts, or release
  assets are not named;
- current mode is unclear, such as bundled runtime vs download-on-demand;
- known failure modes or "do not assume" notes are missing;
- the user says the session has no clue about the subsystem after `kb-map`.

When coverage is insufficient:

1. Stop normal routing long enough to run a targeted `refresh` for that
   subsystem.
2. Search/read only the files needed to understand the missing workflow.
3. Update `docs/context/PROJECT.md` and/or the smallest relevant
   `docs/context/architecture/<subsystem>.md` child doc.
4. Add a `docs/context/memory-maintenance.md` signal:
   `stale-doc` or `repeated-rediscovery`, with the missing subsystem and source
   paths.
5. Re-run `kb-map lookup <same request>` and report the exact docs a fresh
   session should read next.

Example: an installer or release workflow that spans build config, packaging
scripts, CI workflows, release assets, and runtime startup checks needs its own
architecture pointer or child doc. A generic runtime doc is not enough if it
cannot explain how the workflow is built, shipped, and updated.

## Coverage Audit Mode

Use this when the user says the initial map missed a whole class of things, or
when one invisible subsystem suggests there may be more.

Do a bounded repo-wide coverage audit:

1. Inventory major surfaces:
   - routes/screens/UI shells;
   - commands/CLIs/tools/actions/workflows/service capabilities;
   - jobs/workflows/integrations;
   - data/auth/session/storage;
   - build/test/package/installer/updater/release/deploy flows.
2. Compare the inventory to:
   - `docs/context/PROJECT.md`;
   - `docs/context/architecture/README.md`;
   - `docs/context/architecture/*.md`;
   - `docs/context/operations/*.md`;
   - `docs/context/research/*.md` and `docs/context/decisions/*.md` when relevant.
3. Produce a gap table:

   ```markdown
   | Area | Evidence Files | Indexed? | Doc Exists? | Good Enough? | Gap | Priority |
   |---|---|---:|---:|---:|---|---|
   ```

4. Fix P0/P1 coverage gaps immediately when they block fresh-session routing.
   Record lower-priority gaps in `docs/context/memory-maintenance.md`.

This is not a full re-bootstrap unless `PROJECT.md` or the architecture index is
so wrong that targeted repair would be more expensive than rebuilding.

## Shallow Map Detection

During lookup, warn and recommend `refresh` or `kb-map-bootstrap` when the KB
appears much thinner than the repo:

- architecture docs are generic top-level names while substantial child
  directories, route/page folders, commands, tools, actions, or workflows are
  missing from the subsystem index;
- a major source area has a rough source-file to architecture-doc ratio above
  about 25:1 without child pointers or explicit skip reasons;
- root `README.md`, `AGENTS.md`, `.github/copilot-instructions.md`, memories, or
  project notes mention architecture topics absent from `docs/context`;
- file prefix clusters, large files, large directories, or cross-process flows
  are visible but undocumented;
- must-cover concerns such as auth, storage, IPC, browser/HTTP, telemetry,
  settings, build/install, LLM/model, workers, or integrations exist but have no
  doc, parent pointer, or known-unknown entry.

If the requested work touches one of these gaps, stop normal routing and run a
targeted refresh. If many gaps exist, recommend re-bootstrap with coverage
discovery instead of one-off patching.

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
- A lookup exposed a coverage gap: the map could not explain a named subsystem
  without broad rediscovery.

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
7. If `docs/context/landmines.md` exists, archive resolved landmines whose
   owner surface was fixed and verification passed. Leave unfixed stale entries
   active or mark `stale-review`; do not silently delete high-severity traps.
8. If `todo_rules.md`, `todo-rules.md`, or another separate todo rules file exists, inline the rules into the top `## Rules` section of `todo.md`, move any unique durable content to `docs/context/*`, then delete the separate rules file.
9. Run `document-review` when changes are substantial.

Do not re-bootstrap the whole repo here.

## Contracts

`PROJECT.md` is a route map, not an encyclopedia. Subsystem docs carry durable app truth. `todo.md` carries current operational truth and its own board rules. `todo-done.md` carries completed-work summaries. Handoff files carry resumable work packets.
