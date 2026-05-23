---
name: kb-ship
description: Final release readiness lane for KB work. Use when slices are complete and the user wants PR, release, deploy, or final shipping checks after kb-complete, or asks "ship this", "release this", "open PR", or "final readiness".
argument-hint: "[manifest path, branch, or release target]"
---

# KB Ship

Ship deliberately after implementation, QA, review, and learning gates have run.

## Preconditions

Before shipping:

1. Read `todo.md`, `todo-done.md`, and the relevant manifest.
2. Confirm runnable slices are complete or explicitly parked/blocked.
3. Confirm `kb-complete` has run, or run it first.
4. Confirm no unrelated user changes will be staged.

Do not ship if the tree contains unexplained failures or unrecorded human-only blockers.

## Checks

Run `kb-check` first. Shipping requires deterministic verification before prose review.

Run the strongest practical verification:

- Targeted tests for touched areas.
- Full test suite when practical.
- Lint/typecheck/build.
- Browser QA for user-visible web/app behavior.
- Review that the PR/release matches the source brainstorm/plan.
- Cleanup temporary QA artifacts that are no longer needed.

## Release Notes / PR Summary

Produce:

```markdown
## Summary

## Verification

## Slices Completed

## Parked / Deferred

## Risks

## Follow-Up
```

## Git Rules

- Stage only intentional files.
- Do not force push.
- Do not include unrelated dirty files.
- Commit or push only when the user asked for it or the local workflow explicitly requires it.

## Memory Updates

After shipping:

- Update `todo.md`.
- Update `todo-done.md`.
- Move completed handoffs to `docs/handoffs/done/`.
- Update `docs/context/PROJECT.md` or subsystem docs if shipped work changes how future sessions should understand the app.

End with a clear status: shipped, ready to ship, blocked, or parked.
