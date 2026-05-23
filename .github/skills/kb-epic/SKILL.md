---
name: kb-epic
description: Large-initiative coordinator for KB workflows. Use for app migrations, framework rewrites, major architecture changes, multi-subsystem initiatives, multiple brainstorms/manifests, long backlogs, or dark-factory execution across related workstreams.
argument-hint: "[initiative description or epic path]"
---

# KB Epic

Coordinate large work without turning it into one huge plan. This is the lane for "migrate this app to a different framework", "rewrite the architecture", or "run a whole program of related work."

## Goal

Create an epic map that points to multiple brainstorms, plans, manifests, handoffs, research notes, and subsystem docs. Keep each execution unit small enough for `kb-plan` and `kb-work`.

## When To Use

Use `kb-epic` when:

- The request is bigger than one brainstorm or one manifest.
- Example: Python to .NET migration, Electron rewrite, auth overhaul, full chat architecture replacement.
- One plan would be too large to execute or review.
- Work spans multiple subsystems.
- Architecture direction affects many later slices.
- Several brainstorms should feed one release.
- The user wants a dark-factory queue.

## Epic Location

Create:

```text
docs/context/epics/<epic-name>.md
```

If `docs/context/epics/` does not exist, create it and link it from `docs/context/PROJECT.md`.

## Epic Template

```markdown
# <Epic Name>

Status: draft|active|parked|complete
Created: YYYY-MM-DD
Last refreshed: YYYY-MM-DD

## Intent

## Success Criteria

## Architecture Decisions

## Research

## Workstreams

| Workstream | Brainstorm | Manifest | Status | Notes |
|---|---|---|---|---|

## Dependency Map

## Dark Factory Queue

## Human Checkpoints

## Parked / Blocked

## Completion Criteria
```

## Scheduling

Default to serial execution when workstreams share files, schemas, prompts, auth, routing, generated artifacts, or architecture decisions.

Use read-only swarms for research/review. Use coding swarms only when file ownership is declared and non-overlapping.

## Flow

1. Read `todo.md`, `docs/context/PROJECT.md`, and relevant subsystem docs.
2. Create or refresh the epic map.
3. Split the initiative into workstreams small enough for normal KB flow.
4. Route each workstream to `kb-brainstorm` or `kb-plan`.
5. Queue manifests in `todo.md`.
6. Use `kb-work` for runnable manifests.
7. Use `kb-complete` after each manifest or at epic milestones.
8. Use `kb-map refresh` after durable architecture changes.
9. Use `kb-ship` when release readiness matters.

Refresh cold epic work older than 72 hours before execution.
