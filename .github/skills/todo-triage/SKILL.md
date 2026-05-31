---
name: todo-triage
description: Use when reviewing pending todos for approval, prioritizing code review findings, or interactively categorizing work items
argument-hint: "[findings list or source type]"
disable-model-invocation: true
---

# Todo Triage

Review unresolved durable work and decide what belongs on root `todo.md`.

Do not write implementation code during triage. Triage only classifies,
prioritizes, merges, parks, or removes work items.

## Sources

Read in this order:

1. Root `todo.md` active board.
2. Linked manifests, handoffs, or review artifacts.
3. Legacy `.context/compound-engineering/todos/` only when a CE skill explicitly
   generated file-per-todo findings.

## Decisions

For each item, choose one:

| Decision | Action |
|---|---|
| ready | keep on `todo.md` with owner/status/next action |
| blocked | record blocker and resume condition |
| parked | move under parked/cold-storage section |
| duplicate | merge into the canonical existing item |
| delete | remove only if stale, invalid, or already completed |

Ask the user only when priority, product intent, risk acceptance, or deletion is
a real human decision. Otherwise update the board directly.

## Output

```markdown
## Todo Triage

| Item | Decision | Reason | Next action |
|---|---|---|---|
```

If legacy CE todo files were read, report whether they were converted, left in
legacy storage, or no longer relevant.
