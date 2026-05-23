---
name: kb-start
description: Default KB start/router. Use when the user says "kb", gives an idea or ambiguous work request, starts a fresh session, asks what to do next, or wants the workflow to choose the right lane without making the user pick ceremony. Delegates project-memory setup and lookup to kb-map before choosing a lane.
argument-hint: "[user request or blank for session startup]"
---

# KB Start

Pick the right KB lane for the user's idea/request. The user should be able to ask normally; do not make them choose ceremony.

`kb-start` is not the memory bootstrapper. `kb-map` owns project-memory setup, lookup, and refresh.

## Map First

On every fresh session or ambiguous work request:

1. Invoke `kb-map lookup <user request>`.
2. Let `kb-map` decide whether lookup, refresh, or bootstrap is required.
3. After `kb-map` returns project context, classify the user request and route it.
4. If `kb-map` reports stale work or missing memory, honor that before executing work.

If `kb-map` cannot identify a valid active project root, ask for the project path before routing. Drive roots such as `E:\`, home directories, and global skill/config folders are not valid project roots unless the user explicitly chose them. Do not route from global handoffs or home-directory memory when the user is working in a repo.

## Read Order

Read only what `kb-map` points to, then only what is needed to route:

1. `kb-map` result.
2. Relevant active handoff files or manifest paths named by `kb-map`.
3. Specific subsystem, research, brainstorm, or plan files pointed to by `kb-map`.

## Current Truth

`todo.md` may hold short-lived operational truth: current focus, active manifest, parked slices, blockers, and handoff pointers.

Durable app truth belongs in `docs/context/architecture/*`. If an operational fact becomes durable architecture knowledge, ask `kb-map refresh` to update it.

## Stale Work Rule

Before running a handoff, brainstorm, plan, or parked todo older than 72 hours, perform a refresh check:

- What changed since it was created or last refreshed?
- Did touched files/subsystems change?
- Does the route still make sense?
- Does the artifact need updating before execution?

Do not run stale work blindly.

## Handoff Routing Rule

Handoffs are restart packets, not automatically executable plans.

Before resuming any `docs/handoffs/active/*` file, classify it:

| Handoff Shape | Route |
|---|---|
| Contains or links a `docs/plans/*-kb-*-manifest.md` with slice plans | `kb-work <manifest>` |
| Contains vertical slices with `expected_files`, verification, blockers, and status | `kb-plan` to normalize into a manifest, then `kb-work` |
| Contains phases, workstreams, bullets, open decisions, or broad next steps | `kb-plan` |
| Contains unclear product/technical intent | `kb-brainstorm` |
| Contains multiple child initiatives or a migration/rewrite scale objective | `kb-epic` |

Do not route a phase-shaped handoff directly to `kb-work`. `kb-work` requires a manifest and per-slice plans with `expected_files`.

## Route Table

Use plain task classes first, then map to skills:

| Request Shape | Route |
|---|---|
| Project memory missing, partial, or stale | `kb-map` decides lookup/refresh/bootstrap |
| Memory/docs/responses are too verbose | `kb-compact` |
| Need to find app/subsystem context | `kb-map lookup` |
| Active handoff has only phases/workstreams/next steps | `kb-plan` |
| Active handoff links a valid KB manifest | `kb-work` |
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

- Map status.
- Route chosen.
- Why that route fits.
- Any stale-work refresh needed.
- Next action.

If the route is obvious and safe, proceed into the chosen skill workflow.
