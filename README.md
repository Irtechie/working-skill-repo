# Working Skill Repo

Minimal, voice-friendly KB workflow skills for GitHub Copilot and Codex.

This repo is the portable skill bundle I use when I want an agent to walk into a
project, recover local project memory, choose the right workflow, execute work in
vertical slices, test its own changes, review the result, and leave durable
handoff/context files behind.

KB means **Kanban-Based**. The underlying workflow still uses boards, manifests,
vertical slices, and done archives, but user-facing commands use the shorter
`kb-` prefix because it works better with voice input.

## What This Repo Is

This is not the full ATV StarterKit. It is the smaller working set that should be
safe to copy into active projects without dragging in every experiment or
historical workflow.

It includes:

- repo-local GitHub Copilot skills in `.github/skills/`
- repo-local Copilot agents in `.github/agents/`
- root agent guidance in `AGENTS.md`
- Copilot guidance in `.github/copilot-instructions.md`
- a deterministic check helper at `.github/skills/kb-check/scripts/kb-check.ps1`

The default entry point is `kb-start`. In normal use, ask for work in plain
language and let `kb-start` choose the ceremony.

`kb-start` delegates project-memory setup to `kb-map`. Route decides what the
user's idea/request needs; map decides whether local memory needs lookup,
refresh, or first-time bootstrap.

`kb-start` replaces the older `kb-route` name. The behavior is the same front
door pattern, but the new name matches how you use it: start a session, map the
project, then choose the right lane.

`kb-map` is project-root anchored. It must resolve the active repo first and read
memory from that repo only; it should not search `~`, `.copilot/handoffs`, the
whole drive, or sibling repos unless you explicitly ask for cross-repo lookup.
It checks standard memory files by exact path under the repo root, not by
grep/glob.

Drive roots such as `E:\` are not valid project roots unless explicitly chosen.
If the repo root cannot be resolved, `kb-map` should ask for the project path
instead of searching the drive.

## Quick Use

Use these when you know the route:

| Command | Use When |
| --- | --- |
| `kb-start` | Fresh session, ambiguous ask, or "figure out the right workflow" |
| `kb-map` | Setup, lookup, or refresh project memory before other work |
| `kb-fix` | Narrow bug, failing test, or small contained change |
| `kb-brainstorm` | Product or technical framing is still unclear |
| `kb-plan` | Requirements exist and need vertical slices |
| `kb-work` | A manifest exists and should be executed |
| `kb-complete` | Work is done and needs review, learning, cleanup |
| `kb-ship` | Release, PR, deploy, or final readiness check |
| `kb-epic` | Large migration, rewrite, or multi-brainstorm initiative |
| `klfg` | Fully hands-off route: brainstorm -> plan -> work -> complete |

Phase boundaries:

- `kb-brainstorm` writes/reviews requirements, resolves safe/actionable P0-P4
  findings, and invokes `kb-plan` when the brainstorm is gate-clean. It pauses
  only for unresolved blockers, required human decisions, required research, or
  an explicit user stop.
- `kb-plan` writes the manifest and slice plans and recommends `kb-work`.
- `kb-work` executes all runnable slices and calls `kb-complete` only when every
  slice is done or intentionally skipped.
- Once `kb-work` starts execution, runnable slices continue without per-slice
  confirmation. It pauses only for HITL, blocked/manual work, destructive
  approval, scope failures, QA/repair exhaustion, dependency deadlock, or an
  explicit user stop.
- `klfg` is the full auto-chain.

Handoff routing is deliberately conservative:

- Handoffs are restart packets, not automatically executable plans.
- A handoff that links a valid `docs/plans/*-kb-*-manifest.md` can route to
  `kb-work`.
- A handoff with phases, workstreams, bullets, or broad next steps must route to
  `kb-plan` first so it becomes vertical slices with `expected_files` and
  verification.
- Before planning from a handoff, `kb-plan` checks for existing brainstorm,
  requirements, manifest, or plan files and uses the best existing source of
  truth instead of duplicating work.
- A handoff with unclear intent routes to `kb-brainstorm`; a multi-initiative
  handoff routes to `kb-epic`.

## Installed Skills

Core workflow:

- `kb-start`
- `kb-map`
- `kb-map-bootstrap`
- `kb-compact`
- `kb-check`
- `kb-functional-test`
- `kb-gate`
- `kb-fix`
- `kb-research`
- `kb-brainstorm`
- `kb-plan`
- `kb-work`
- `kb-complete`
- `kb-qa`
- `kb-repair`
- `kb-first-principles`
- `kb-epic`
- `kb-ship`
- `klfg`

Direct dependencies:

- `document-review`
- `tdd`
- `ce-review`
- `ce-compound`
- `ce-compound-refresh`
- `learn`
- `evolve`
- `todo-create`
- `todo-triage`

## Not Bundled

These are intentionally left out of the minimal working bundle:

- upstream `deepen-*` passes; use `kb-research` and proportional research
- one-shot LFG/SLFG style workflows; use `klfg` only when you actually want the
  whole pipeline
- `land`; shipping remains a deliberate separate decision
- browser tools such as `agent-browser`; skills can call them when installed, but
  this repo does not vendor them

## Project Memory

The workflow keeps memory in files so sessions can stay short.

Required project memory files:

- `todo.md` - active work, blockers, parked work, and handoff pointers
- `todo-done.md` - compact archive of completed work
- `docs/context/PROJECT.md` - project route map for fresh sessions
- `docs/context/architecture/` - concise architecture notes by domain
- `docs/context/research/` - reusable research findings
- `docs/context/decisions/` - durable decisions and tradeoffs
- `docs/handoffs/active/` - resumable work
- `docs/handoffs/parked/` - valuable work that is not currently runnable
- `docs/handoffs/done/` - completed or superseded handoffs

Fresh-session preflight:

- Start with `kb-start` for work requests.
- `kb-start` calls `kb-map lookup <request>` before choosing the lane.
- If `todo.md` or `docs/context/PROJECT.md` is missing, `kb-map` invokes
  `kb-map-bootstrap`.
- If the context or handoff folders are partially missing, `kb-map` refreshes or
  creates the missing structure.
- Do not ask before bootstrapping unless a non-empty user file would be
  overwritten.

This replaces older `docs/kanban.md`, `docs/kanban-done.md`, `kb.md`,
`kb-done.md`, and ad-hoc handoff naming for the KB workflow.

## Execution Model

The pipeline is designed around three task sizes:

- **Small:** use `kb-fix`. Write or identify a failing check, make the smallest
  fix, run deterministic verification, and stop if the fix loop stalls.
- **Medium:** use `kb-brainstorm` -> `kb-plan` -> `kb-work`. Produce vertical
  slices with expected files, verification, dependencies, and HITL flags.
- **Large:** use `kb-epic`. Break the initiative into multiple brainstorms or
  manifests before execution.

`kb-gate` owns P0/P1/P2/P3/P4 handling. Findings are expected; human-in-loop is
based on who can safely decide, not severity alone. The agent should fix
safe/actionable findings itself and ask for human help only for product
decisions, credentials/access, unsafe operations, competing reasonable paths, or
genuine ambiguity.

`kb-check` and `kb-functional-test` push verification into code whenever
possible. The model should call deterministic checks instead of spending tokens
re-inspecting behavior by hand.

## Recommended Install

Default to personal/global installs. They keep active project repos clean and
avoid skill drift between copies.

GitHub Copilot personal install:

```powershell
$src = 'E:\working-skill-repo'
$skills = "$env:USERPROFILE\.copilot\skills"
$agents = "$env:USERPROFILE\.copilot\agents"
Copy-Item "$src\.github\skills\*" $skills -Recurse -Force
Copy-Item "$src\.github\agents\*" $agents -Force
```

Shared agent-skills standard install:

```powershell
$src = 'E:\working-skill-repo\.github\skills'
$dst = "$env:USERPROFILE\.agents\skills"
Copy-Item "$src\*" $dst -Recurse -Force
```

Codex personal install:

```powershell
$src = 'E:\working-skill-repo\.github\skills'
$dst = "$env:USERPROFILE\.codex\skills"
Copy-Item "$src\*" $dst -Recurse -Force
```

Use repo-local installs only when a project needs pinned/project-specific
overrides or when the skills should be versioned with that codebase.

Repo-local GitHub Copilot install:

```powershell
$src = 'E:\working-skill-repo'
$dst = 'E:\path\to\your\project'
Copy-Item "$src\.github\skills" "$dst\.github\skills" -Recurse -Force
Copy-Item "$src\.github\agents" "$dst\.github\agents" -Recurse -Force
Copy-Item "$src\AGENTS.md" "$dst\AGENTS.md" -Force
Copy-Item "$src\.github\copilot-instructions.md" "$dst\.github\copilot-instructions.md" -Force
```

## Skill Quality Bar

KB skills should be structured, not brain dumps:

- frontmatter says exactly when to use the skill
- the body states the job, non-goals, and output contract
- workflows are split into phases with hard gates
- file paths, commands, and artifact locations are explicit
- questions are driven by blocking decisions, not a quota
- shared doctrine lives once and is referenced elsewhere
- long research, agent prompts, and scripts are lazy-loaded when needed

Every token must pay rent. Keep contracts, gates, paths, commands, error
handling, verification criteria, and escalation thresholds. Cut generic
programming advice, motivational text, repeated warnings, and long examples that
modern models do not need.

## Credits

This repo is primarily based on the ATV / All The Vibes skill set and its
Compound Engineering workflow.

It also borrows useful ideas from:

- [Matt Pocock's skills](https://github.com/mattpocock/skills), especially small
  composable workflow skills and vertical slicing
- [G-Stack](https://github.com/garrytan/gstack), especially persistent workflow
  memory, QA ownership, and operating-system-style orchestration
- [kevin-copilot](https://github.com/shyamsridhar123/kevin-copilot), especially
  terse Copilot-first instruction surfaces

The goal is not to copy any one system. The goal is to keep the pieces that make
agents more reliable, cheaper to run, easier to resume, and harder to let off
the hook.
