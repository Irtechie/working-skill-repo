# KB Memory Templates

Use these templates only when creating missing project-memory files.

## `todo.md`

```markdown
# Todo

## Rules

**Conventions:** these match the KB skill spec. Keep them inline here; do not split into `todo_rules.md`, `todo-rules.md`, or any separate rules file.

**This file is the single source of truth for active work** — not chat history, not session SQL, and not stale manifests. Any agent should be able to claim a row from here cold.

**Status markers** (applied to individual rows):

| Marker | Meaning |
|--------|---------|
| ⬜ pending | Ready when blockers clear |
| 🔧 in_progress | Agent claimed and actively working |
| ✅ done | Complete and verified — move summary to `todo-done.md` promptly |
| 🔒 blocked | Cannot proceed — explain in `## Blocked` with `Depends on:` |
| ⊘ skipped | Intentionally skipped with reason |
| 🛑 human-required | Needs human action (HITL) — also surface under `## Human Required` |

**Section icons** (section headers, not row markers):

- 💡 Feature Ideas — not yet brainstormed; a human promotes to active
- 📋 Queued Improvements — approved but not yet planned
- 🧊 Parked / Cold Storage — intentionally out of bounds today; never auto-runs, human-promote only
- 🛑 Human Required — items only a person can complete (logins, approvals, decisions)
- 📝 Work Log — short dated entries for cross-session visibility

**Task metadata** lines under a row when relevant: `Task ID:`, `Ready: yes|no`, `Depends on:`, `Discovered from:`, `Validation:`.

**Promotion rules:**
- Newly discovered work goes to 🧊 Parked / Cold Storage first. Never auto-execute from there.
- Items stalled because another agent, dependency, tool failure, or missing input must finish first go to 🔒 Blocked, not Parked.
- Human-required work must not be silently folded into generic blocked notes.
- Detailed handoffs live under `docs/handoffs/`; link them here instead of pasting content.
- Refresh cold or parked work older than 72 hours before execution.
- Keep this file current and small. When all active todos are done, check the handoff queue.

## Objective
## Current Focus
## Current Truth
## Active Work
## Handoff Queue
| Handoff | Status | Route | Created | Stale Check | Link |
|---|---|---|---|---|---|
## Human Required
## Parked / Cold Storage
## Blocked
## Work Log
```

## `todo-done.md`

```markdown
# Completed Work

> Archive of completed items from `todo.md`. Most recent at top.

## YYYY-MM-DD
- <feature or slice group> — <compact outcome, important proof, commit/link if available>
```

## `PROJECT.md`

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
