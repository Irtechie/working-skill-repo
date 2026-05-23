---
name: kb-map-bootstrap
description: Token-expensive bootstrap skill that deeply indexes a new or existing project and creates the standard KB memory layout. Use when kb-map reports missing or badly stale memory, when entering an existing project without todo.md/docs/context/PROJECT.md, or when the user says "bootstrap this project", "deep map this repo", "build project memory", or "index this app".
argument-hint: "[optional project focus or subsystem hints]"
---

# KB Map Bootstrap

Build parity: after this runs, a fresh session should use the same files whether the app is new or years old.

Use `kb-map` for normal startup. Use this only for missing or badly stale memory.

## Automatic Invocation

When `kb-map`, `AGENTS.md`, or `.github/copilot-instructions.md` detects missing `todo.md` or `docs/context/PROJECT.md`, run this skill immediately. Do not ask the user first unless a non-empty user file would be overwritten or moved.

Run bootstrap in the active project root only. Prefer `git rev-parse --show-toplevel`; otherwise use the current working directory only if it is clearly a project directory. Never bootstrap a drive root such as `E:\`, `~`, `%USERPROFILE%`, `.copilot`, `.codex`, `.agents`, the whole drive, or a sibling repo unless the user explicitly chose that path.

## Create Layout

```text
todo.md
todo-done.md
docs/context/
  PROJECT.md
  architecture/
    README.md
  research/
    README.md
  decisions/
    README.md
  operations/
    README.md
    testing.md
docs/brainstorms/
docs/plans/
docs/handoffs/
  active/
  parked/
  done/
```

Optional:

```text
docs/context/decisions/starter-kit-deltas.md
docs/context/epics/
docs/context/history/
```

## Workflow

1. **Inventory the repo**
   - Top-level structure, entry points, frameworks, package managers.
   - Build/test/dev commands.
   - Routes, screens, commands, tools, actions, jobs, integrations.
   - Tests, docs, existing TODOs, brainstorms, plans, ADRs, and handoffs.

2. **Identify subsystems**
   - User-facing workflows.
   - Backend domains.
   - Tool/action layers.
   - Data/storage layers.
   - External integrations.
   - Runtime shells such as Electron, browser, mobile, or CLI.

3. **Create or merge memory files**
   - Preserve existing user docs.
   - Do not overwrite non-empty files without reading and merging.
   - Move stale or completed active work out of the active board.
   - Use lowercase kebab-case except `PROJECT.md` and folder `README.md`.

4. **Write `docs/context/PROJECT.md`**
   - Keep it short.
   - Include run/test commands.
   - Include subsystem, research, operation, and active-work pointers.
   - Mark confidence as verified, inferred, or unknown.

5. **Write testing operations**
   - Create/update `docs/context/operations/testing.md`.
   - Include deterministic commands discovered by `.github/skills/kb-check/scripts/kb-check.ps1 -List` or equivalent manifest inspection.
   - Note which checks are fast, broad, flaky, external-service dependent, or CI-only.

6. **Write subsystem docs**
   - One concise doc per major subsystem.
   - Parent docs summarize and point to child docs.
   - Include known sharp edges, rejected approaches, and first files to read.

7. **Write board and handoff structure**
   - `todo.md` for active work and handoff queue pointers.
   - `todo-done.md` for compact completion summaries.
   - `docs/handoffs/active/`, `parked/`, and `done/`.
   - If `todo-rules.md`, `docs/todo-rules.md`, or another separate todo rules file exists, inline current board rules at the top of `todo.md`. Delete the separate rules file only after moving any unique project content into `todo.md` or `docs/context/*`.

8. **Starter-kit deltas**
   - If the app is based on ATV, another starter kit, or a fork, create/update `docs/context/decisions/starter-kit-deltas.md`.

9. **Review**
   - Run `document-review` on `PROJECT.md` and large architecture docs when available.
   - Record unresolved findings in `todo.md` or an active handoff.

10. **Route test**
   - Run a cheap `kb-map lookup` against the new memory.
   - Confirm a fresh session can answer: what this app is, how to run it, how to test it, what work is active, and which subsystem docs to read first.

## Templates

### `todo.md`

```markdown
# Todo

## Rules
- Keep this file current and small.
- Active, blocked, parked, and human-required work belongs here.
- Completed work does not stay here. When a feature, slice group, handoff, or fix is complete, move the compact completion summary to `todo-done.md` and remove the completed entry from this file.
- Do not keep routine "slice complete" or verification-success logs here after completion.
- Handoffs live under `docs/handoffs/`; link them here.
- Refresh cold or parked work older than 72 hours before execution.
- These rules live at the top of `todo.md`; do not rely on a separate `todo-rules.md`.

## Objective
## Current Focus
## Current Truth
## Active Work
## Handoff Queue
## Human Required
## Parked / Cold Storage
## Blocked
## Work Log
```

### `PROJECT.md`

```markdown
# Project Map

Bootstrap: YYYY-MM-DD
Bootstrap confidence: verified|mixed|rough

## What This Is
## How To Run
## How To Test
## Current Architecture
## Subsystem Index
| Area | Read This | Use When | Confidence |
|---|---|---|---|
## Current Work Pointers
## Known Sharp Edges
## Research Index
## Do Not Repeat
## Maintenance Notes
```

## Output

Finish with:

- Files created or updated.
- Major subsystems discovered.
- Uncertain areas.
- Stale or completed work moved.
- Result of the `kb-map lookup` route test.
