---
name: kb-route
description: Default KB entrypoint and lightweight project-memory router. Use when the user says "kb", gives an ambiguous work request, starts a fresh session, asks what to do next, or wants the workflow to choose the right lane without making the user pick ceremony.
argument-hint: "[user request or blank for session startup]"
---

# KB Route

Pick the right KB lane with the least context needed. The user should be able to ask normally; do not make them choose ceremony.

## Required Memory Contract

A KB-enabled repo must have:

```text
todo.md
todo-done.md
docs/context/PROJECT.md
docs/context/architecture/
docs/context/research/
docs/context/decisions/
docs/context/operations/
docs/handoffs/active/
docs/handoffs/parked/
docs/handoffs/done/
```

On every fresh session or ambiguous work request:

1. Check whether `todo.md` and `docs/context/PROJECT.md` exist.
2. If either is missing, immediately invoke `kb-map-bootstrap`. Do not ask first unless a non-empty user file would be overwritten.
3. If the core files exist but context or handoff directories are missing, invoke `kb-map refresh`.
4. After bootstrap or refresh, return to this router and continue with the user's request.

## Read Order

Read only what is needed:

1. `AGENTS.md` if present.
2. `todo.md`.
3. `docs/context/PROJECT.md`.
4. Relevant active handoff files linked from `todo.md`.
5. Specific subsystem, research, brainstorm, or plan files pointed to by the above.

If the required memory contract is missing, follow the automatic preflight above before routing.

## Current Truth

`todo.md` may hold short-lived operational truth: current focus, active manifest, parked slices, blockers, and handoff pointers.

Durable app truth belongs in `docs/context/architecture/*`. If an operational fact becomes durable architecture knowledge, route to `kb-map refresh`.

## Stale Work Rule

Before running a handoff, brainstorm, plan, or parked todo older than 72 hours, perform a refresh check:

- What changed since it was created or last refreshed?
- Did touched files/subsystems change?
- Does the route still make sense?
- Does the artifact need updating before execution?

Do not run stale work blindly.

## Route Table

Use plain task classes first, then map to skills:

| Request Shape | Route |
|---|---|
| `todo.md` or `docs/context/PROJECT.md` missing | `kb-map-bootstrap` immediately |
| Context or handoff directories missing | `kb-map refresh` |
| Project memory badly stale | `kb-map-bootstrap` |
| Memory/docs/responses are too verbose | `kb-compact` |
| Need to find app/subsystem context | `kb-map lookup` |
| Recent work changed project memory | `kb-map refresh` |
| Small known bug or narrow fix | `kb-fix` |
| External/prior-art research needed | `kb-research` |
| Fuzzy idea, product direction, high path dependency | `kb-brainstorm` |
| Clear feature needs slices | `kb-plan` |
| Manifest exists and work should run | `kb-work` |
| All runnable slices done, need review/learning/cleanup | `kb-complete` |
| Large initiative with many brainstorms/plans | `kb-epic` |
| Release, PR, deploy, final readiness | `kb-ship` |
| User wants everything from idea to done | `klfg` |

## Task Sizing

- **Small fix**: one bug, obvious scope, low path dependency. Use `kb-fix`.
- **Feature/refactor**: one bounded feature or refactor that can become one manifest. Use `kb-brainstorm` if behavior is unclear; otherwise `kb-plan`.
- **Large initiative**: multi-manifest work such as framework migration, major architecture replacement, cross-subsystem rewrite, or a backlog that needs multiple brainstorms/plans. Use `kb-epic`.
- **Release**: packaging, PR, deploy, or final readiness. Use `kb-ship`.

When in doubt, prefer the lane that prevents rework. Do not pick a 20-minute shortcut when the decision creates path dependency.

## Ceremony Rule

Minimize visible ceremony:

- Do not ask "which KB skill should I use?"
- State the chosen lane in one line, then proceed when safe.
- Ask only when the choice changes risk, cost, or user intent.
- If the wrong lane becomes obvious, switch lanes and record why.

## Token Budget

Every token must pay rent. Keep startup output short and load only pointed-to files.

Route to `kb-compact` when:

- `todo.md`, handoffs, research notes, or architecture docs carry repeated history instead of current signal.
- A skill draft repeats rules already in `AGENTS.md` or `.github/copilot-instructions.md`.
- The user asks for fewer words, terser output, or token reduction.

Do not compact away exact commands, paths, dates, IDs, acceptance criteria, blockers, HITL reasons, or safety warnings.

## Output

Report briefly:

- Route chosen.
- Why that route fits.
- Any stale-work refresh needed.
- Next action.

If the route is obvious and safe, proceed into the chosen skill workflow.
