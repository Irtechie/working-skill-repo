---
name: kb-fix
description: Small bug-fix lane with bounded diagnosis, agent-owned verification, and escalation. Use when the user reports a bug, failing test, broken behavior, or asks for a narrow fix that should not require a full brainstorm or plan unless it grows.
argument-hint: "[bug description, failing command, or issue path]"
---

# KB Fix

Fix narrowly, verify yourself, and escalate when the bug is not actually small.

## Preflight

1. Read `todo.md`, `docs/context/PROJECT.md`, and relevant handoff/subsystem docs.
2. Check `docs/context/research/` and `docs/solutions/` for related failures.
3. If the request is older than 72 hours or comes from a parked handoff, run a stale-work refresh first.

## Loop

1. **Orient** - classify as regression, known bug, flaky behavior, environment issue, or missing requirement.
2. **Reproduce** - build an agent-runnable signal: failing test, command, browser flow, curl, fixture, or logs.
3. **Hypothesize** - write 1-3 falsifiable causes before editing.
4. **Fix narrowly** - edit the smallest responsible area.
5. **Verify** - rerun the reproduction and relevant tests. For UI bugs, use browser verification when available.
6. **Record** - update `todo.md`, handoff files if needed, and learning docs when the fix teaches something reusable.
7. **Refresh memory if durable** - run `kb-map refresh` when the fix changes behavior, architecture, run/test commands, integrations, sharp edges, or "do not repeat" knowledge. Skip with a note when the fix is cosmetic or isolated.

Do not tell the user to test normal app behavior if the agent can test it.

## Ceilings

Default ceiling:

- 3 reproduction strategies.
- 5 fix/verify iterations.
- 2 root-cause hypothesis resets after confirmed reproduction.
- 1 escalation before asking for human help.

Progress can extend the ceiling. Progress means the failure narrows, reproduction improves, evidence points to a subsystem, or one verification layer now passes.

No progress means the same failure persists after unrelated edits, hypotheses change without evidence, or the next attempt cannot say what it proves.

## Escalation

Escalate to:

- `kb-research` when external/framework behavior is unclear.
- `kb-brainstorm` when product behavior or scope is unclear.
- `kb-plan` when the fix becomes multi-slice.
- `kb-map refresh` when architecture docs changed.

When stuck, produce:

```markdown
## Bug Escalation Report

### Symptom
### Expected Behavior
### Reproduction
### Attempts
| Attempt | Hypothesis | Change/Test | Result |
|---|---|---|---|
### Evidence
### Current Best Theory
### Why I Am Stuck
### Recommended Route
```

Park only the blocked slice or handoff; continue unrelated runnable work when possible.
