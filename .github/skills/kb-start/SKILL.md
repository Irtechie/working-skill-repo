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

If `kb-map` cannot identify a valid active project root, ask the user to change into the project directory or provide the project path before routing. Drive roots such as `E:\`, home directories, and global skill/config folders are not valid project roots unless the user explicitly chose them. Do not route from global handoffs or home-directory memory when the user is working in a repo.

## Session Hygiene Check

Run this check only when `kb-start` begins a request. Do not interrupt an active brainstorm, plan, work slice, review, or test loop just to suggest a restart.

Goal: decide whether the user is better served by staying in the current session, compacting, or creating/updating a handoff and restarting fresh.

Use exact context telemetry when the platform exposes it. In GitHub Copilot CLI, `/context` shows context usage. If the agent cannot read telemetry directly, do not guess a percentage; use the evidence-based fallback below.

Context thresholds when exact telemetry is available:

| Context Used | Default |
|---|---|
| `<60%` | Stay in session. Do not mention restart unless the user asks. |
| `60-80%` | Mention restart only if the user is switching tasks or lanes. |
| `80-90%` | Recommend handoff/restart before starting substantial new work. |
| `>90%` | Strongly recommend handoff/restart, or compact if the user must continue here. |

Evidence-based fallback when telemetry is unavailable:

- Suggest restart when the session is long, tool output has been heavy, compaction likely happened, the user is switching tasks, or the agent is relying on chat history instead of local files.
- Do not suggest restart merely because the session feels long.

Before recommending restart, estimate rebuild cost:

| Rebuild Cost | Signals | Recommendation |
|---|---|---|
| Low | current handoff exists; `todo.md`, `PROJECT.md`, and manifest/plan pointers are current | Recommend fresh session when context pressure exists. |
| Medium | project memory exists but handoff needs updating | Offer to update/create a handoff, then restart. |
| High | important nuance is only in chat; mid-debug observations matter; no current handoff/map | Stay, or compact first, then write durable memory before restarting. |

Restart rule:

> Do not recommend a fresh session merely because the session is long. Recommend it only when durable local memory can replace the live chat at lower total context cost or lower drift risk.

When restart is advisable, ask once:

```text
This looks like a good reset point. I can create/update a handoff so the next session starts cleanly, or we can keep going here.

1. Create/update handoff and restart
2. Compact current context and continue
3. Continue in this session
4. Other / let me explain
```

If the user chooses handoff/restart, create or update the active handoff under `docs/handoffs/active/`, ensure `todo.md` points to it, and include the exact `kb-start <handoff/task>` prompt for the next session. Do not run the next workflow in the old session unless the user asks.

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
| User explicitly says `kb-task`, asks for first-principles execution, or wants one bounded task carried until verified/blocked | `kb-task` |
| User says "go straight to work", "don't ask many questions", "just build it", or similar, but no valid KB manifest exists | `kb-plan` with execution intent, then `kb-work` |
| User says troubleshoot/debug/self-correct, asks to inspect logs/browser behavior, or wants the agent to iterate until broken behavior is fixed | `kb-troubleshoot` |
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

Execution intent is not permission to skip planning. If the user wants fewer questions or wants the agent to continue directly to implementation, reduce Q&A and carry an `execute_after_plan` intent forward, but still create or reuse a KB manifest before `kb-work`.

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
