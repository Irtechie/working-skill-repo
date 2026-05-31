---
name: todo-create
description: Use when creating durable work items, managing todo lifecycle, or tracking findings across sessions in the file-based todo system
argument-hint: "[work item, finding, or board update]"
disable-model-invocation: true
---

# Todo Create

Create durable KB work items in the repo root `todo.md`.

Do not introduce `backlog.md` for KB work. Do not create
`.context/compound-engineering/todos/` unless a legacy CE review flow explicitly
requires file-per-todo compatibility.

## KB Todo Model

Use root `todo.md` as the active board:

- active manifests and slices;
- blockers and human-required items;
- parked work that should survive sessions;
- handoff pointers.

Use `todo-done.md` for completed summaries. Move finished routine history out
of `todo.md` so the active board stays small.

## Create A KB Item

1. Read the existing `todo.md` board contract at the top of the file.
2. Add the item to the closest existing section.
3. Include owner/status, manifest or handoff link when available, blocker, and
   next action.
4. Prefer one concise row or bullet. Create a new section only when the board has
   no natural home.

Create a durable todo when work survives this turn, has dependencies, requires
prioritization, or must be visible to the next fresh session. Use the platform
task list for temporary in-turn steps.

## Legacy CE Compatibility

Older CE review flows may still ask for file-per-todo output under
`.context/compound-engineering/todos/`. Treat that as non-KB compatibility:

- read legacy files when a CE skill explicitly points to them;
- do not make the legacy directory canonical for KB;
- prefer converting unresolved CE findings into root `todo.md` entries when
  continuing through KB workflows.
