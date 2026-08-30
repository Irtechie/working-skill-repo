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
3. After the project root is resolved, reap eligible terminal work from a
   different session. Prefer the globally installed `kbreconcile` when its
   capability probe succeeds: run a cutoff-bound `plan`, then checked `apply`
   and `verify`. This is also the scheduled/on-demand portfolio sweep path.
   When it is absent, unavailable, or incompatible, use the repo-native guard
   as a safe fallback:

   ```powershell
   go run ./cmd/kbcheck terminal-cleanup --action sweep `
     --session-id <current-project-session-id> --root <project-root>
   ```

   The guard holds the shared lock across the final claim reread and refreshes
   authoritative remote evidence immediately before each destructive removal;
   direct delivery refreshes again before local-ref compare-and-swap. Repos with
   no authoritative remote default fail closed. It must preserve the current executing session by both
   session ID and worktree path, primary checkout, tracked/untracked/ignored
   dirt, locked/moved/recreated worktrees, active queue claims, rewritten or
   uncontained commits, and unresolved paths. A cleanup-only blocker does not
   block unrelated startup work; report it when it overlaps the requested
   branch/worktree or requires host-owned session-record deletion.
   A reconciler outage is an agent-owned retry state and does not create a human
   packet. Only unresolved irreversible authority ambiguity may do that.
4. After `kb-map` returns project context, check or claim the shared work queue,
   then classify the user request and route it.
5. If `kb-map` reports stale work or missing memory, honor that before executing work.

## Session-End Durability

Stranded uncommitted work is the most common form of lost session output. When
this session ends with a dirty tree, preserve it before exiting:

```shell
go run ./cmd/kbcheck session-preserve --action apply --session-id <session-id> --json
```

One WIP commit on the session's own branch. Never pushed, never merged, never a
completion claim. Refuses on the default branch, detached HEAD, and in-progress
merge/rebase. Excludes build artifacts and oversized files, reporting them in
`excluded[]`.

Preservation is durability, not delivery. It never substitutes for `kb-complete`
and never satisfies a proof gate.

## Shared Work Queue Gate

Before any mutating route, successor session, plan-run worktree, test wave, or
delegated worker starts, publish an early repository-wide work claim.

`<skill-dir>` below is the directory containing the `kb-start` SKILL.md you
loaded. Resolve it from that path. Do not assume a fixed install root:
`.github/skills/`, `~/.copilot/skills/`, `~/.codex/skills/`, and
`~/.agents/skills/` are all valid locations for this bundle.

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File "<skill-dir>/scripts/work_queue.ps1" `
  -Action claim `
  -WorkId <stable-kebab-case-objective> `
  -SessionId <project-session-id> `
  -Branch <branch> `
  -Summary "<one-line outcome>" `
  -Scope "<paths or subsystem>" `
  -SemanticResources <code:module,publisher:product,release-manifest:product,deploy:environment> `
  -Status in_progress
```

The queue lives under the Git common directory, so every worktree sees the same
claims. A mutating session without a session ID must stop before work. Read-only
answers do not need a claim.

Every mutating route registers lifecycle identity and declared semantic resources
before mutation. The helper normalizes declarations, counts distinct active
owners for repository WIP, and fails closed on a conflicting active writer.
Its `global_authority: false` result is deliberate: Git-common-directory state
coordinates only linked local worktrees. A protected writer remains unavailable
unless `kbreconcile claim-capability` reports a real authoritative adapter,
scoped authorization verifier, atomic fenced gateway, durable idempotency, and
sole-path production controls. A local lease, generation number, or queue entry
never substitutes for authoritative adapter capability.

- Default WIP is three active sessions per repository and one active session per
  `work_id`.
- On conflict, do not open a successor or repeat discovery/tests. Report the
  owning `session_id`, branch, worktree, status, and heartbeat; coordinate with
  or inspect that session.
- Heartbeat with `-Action update` at every route/phase boundary and before a
  long test wave.
- Use active statuses `queued`, `in_progress`, and `active`; lifecycle release
  statuses include `paused`, `awaiting-review`, `local-durable`,
  `delivery-integrated`, `blocked`, `quarantined`, `retired`, `done`, and
  `superseded`.
- Mark `blocked` with the exact resume condition. Mark `done` or `superseded`
  before cleanup. Never silently abandon an active claim.
- `-Action list -StaleMinutes 60` identifies claims whose session may have died;
  inspect that exact session ID before takeover. Stale is a review signal, not
  automatic permission to overwrite work.

Every startup response for mutating work must include:

```text
Work: <work_id>
Status: <queued|in_progress|blocked|done|superseded>
Session: <project-session-id>
Branch: <branch>
```

If `kb-map` cannot identify a valid active project root, ask the user to change into the project directory or provide the project path before routing. Drive roots such as `E:\`, home directories, and global skill/config folders are not valid project roots unless the user explicitly chose them. Do not route from global handoffs or home-directory memory when the user is working in a repo.

## Run-State Guard

When a durable goal or autonomous loop names `.kb/runs/<goal-slug>/`, validate
its route history before choosing another lane:

```powershell
go run ./cmd/kbcheck run-state --history .kb/runs/<goal-slug>/route-history.jsonl
```

If the guard reports `route-oscillation`, `low-confidence-no-progress`, or
`no-progress-loop`, do not keep routing. Stop and choose the smallest repair:
refresh stale context, re-plan the work unit, or ask the one human question that
would break the loop.

After choosing a lane for an active run-state goal, append one route-history row
with at least `ts`, `route`, `confidence`, and either `state_changed` or
`progress_key`. Use confidence as a routing-confidence signal, not a completion
claim.

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

## Ranked Routing Decision List

Choose the first matching route. This list is the only routing taxonomy; do not
reconcile it with a second shape or complexity table.

| Rank | Request Signal | Route | Proof/Gate |
|---|---|---|---|
| 1 | Project memory missing, partial, stale, or root invalid | `kb-map` | `kb-map` decides lookup/refresh/bootstrap |
| 2 | User explicitly says `kb-goal`, sets a durable goal, wants work to run for days, asks for vDone, or needs cross-session objective tracking | `kb-goal` | goal ledger plus delegated KB gates |
| 3 | User explicitly says `kb-task`, asks for first-principles execution, or wants one bounded task carried until verified/blocked | `kb-task` | task runner owns verification |
| 3a | User asks what outstanding work exists across the project, wants dead or superseded plans marked, or wants branches reconciled against declared work before new work starts | `kb-rehab` | `kbcheck work-reality` containment proof plus a bounded decision packet |
| 4 | Direct explanation, tradeoff discussion, or pushback with no file changes requested | answer directly; use `kb-first-principles` behavior when challenged | no work gate |
| 5 | User wants a feature/plan/manifest taken from its current state through configured local, PR, or direct delivery | `kb-complete` | plan/work/finalize gates plus delivery policy |
| 5a | User says `w2d`/"work to done"/"land it" | `w2d` | invocation is run-scoped merge intent; resolves a durable manifest or plans proportionally to recover; merge still needs permissions and checks |
| 5b | User says `p2d`/"plan to done", or planning is the expected main event | `p2d` | `kb-plan` gate, then `w2d` delivery |
| 6 | Existing valid non-plan-run manifest should be executed without check-in intent | `kb-work` | manifest plus slice verification |
| 6a | Existing plan-run manifest has explicit local check-in authority | `kb-work` | manifest, commit authority, and slice verification |
| 7 | All runnable slices are done and only internal review/learning/cleanup is needed | `kb-finalize` | `kb-check`, `kb-review`, learning gates |
| 8 | Already reviewed work needs configured delivery | `kb-complete` | delivery policy plus release/ship/land gates |
| 9 | Broken behavior needs logs, browser checks, test iteration, or self-correction | `kb-troubleshoot` | reproduce and regression proof |
| 10 | Architecture/module-depth exploration before implementation | `kb-architecture-deepening` | source/docs evidence and tradeoff table |
| 11 | External docs, prior art, framework behavior, or market/product research materially affects the answer | `kb-research` | cited source notes |
| 12 | Multiple independent streams, many blockers, deletion policy, migration scale, or several brainstorms/plans needed | `kb-epic` | brainstorms and plans complete before work |
| 13 | Scripts/evals/proof harness plus skills/docs must change together, or cross-runtime propagation is part of the change | `kb-epic` or coded pipeline manifest | eval/proof/sync gate |
| 14 | Clear feature/refactor needs slices, or user wants execution but no valid manifest exists | `kb-plan` | vertical-slice manifest |
| 15 | Skill-bundle change with sync/docs/eval/standard gate implications | `kb-plan` | `kb-check -All` and sync report |
| 16 | Fuzzy idea, product direction, or high path dependency | `kb-brainstorm` | answered questions before planning |
| 17 | Small known bug, typo, narrow cleanup, or one skill/doc edit with no sync/eval/proof-harness implications | `kb-fix` or bounded direct edit | targeted proof plus `kb-check` when relevant |
| 18 | Memory/docs/responses are too verbose | `kb-cognitive` | preserve commands, paths, dates, blockers |
| 19 | Legacy full-pipeline or finish wording | `kb-complete` | full state-aware pipeline |

A commit-required plan-run manifest without explicit local check-in authority
stops before mutation. Do not route it through row 6 or infer authority from
repository ownership, a prior planning request, or delivery policy.

Pipeline-worthy changes have at least one of these signals: multiple owning
surfaces, cross-runtime behavior, scorer/fixture/baseline changes,
propagation/sync rules, several independent workstreams, or deletion/loaded
surface measurement.

If none of those signals are present, keep the route small. Do not build a
pipeline just because the request mentions a skill.

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

Route by complexity, not by file count or guessed duration. The useful signals are uncertainty, blast radius, coupling, reversibility, verification burden, and user/product path dependency.

Record `workflow_shape` in generated manifests when planning follows the ranked
list. Use the closest rank label, such as `single-skill-edit`,
`skill-bundle-change`, `pipeline-change`, or `multi-stream-epic`.

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

Route to `kb-cognitive` when:

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
- Exact next action, including the skill command and artifact path when known.

If the route is obvious and safe, proceed into the chosen skill workflow. Also
name the exact command before or as you invoke it, so users on hosts that do not
auto-chain skills can continue manually. If the host cannot invoke the target
skill, stop with: `Next command: <skill> <artifact-or-request>`.

## Tooling Availability

`cmd/kbcheck` belongs to this bundle's source repo and does not ship with an
installed skill. Treat every `go run ./cmd/kbcheck ...` command in this skill as
conditional: run it when the repo provides it, otherwise substitute the
project's own equivalent check and record which command produced the proof.
A missing harness changes which command you run, never whether you prove the work.
