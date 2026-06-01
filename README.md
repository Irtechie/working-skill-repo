# Working Skill Repo

Voice-friendly KB workflow skills and required reviewer agents for GitHub
Copilot and Codex.

Status: actively used, pre-1.0. The top-level gate is native Go and has Windows
parity smoke proof against the prior PowerShell wrappers; expect churn while the
marketplace, eval, and pipeline pieces settle.

Most of this repo is an augmentation layer on top of the original
All-The-Vibes ATV StarterKit and CE review/learning workflow. KB adds the
voice-friendly routing, project-memory map, fresh-session handoff loop,
proportional planning, and execution gates; it still depends on selected ATV
skills and reviewer agents. Original ATV `upstream/main` is a source to mine
for useful ATV-native changes, while this repo remains the source of truth for
the KB overlay and any KB replacements.

This repo is the portable skill bundle I use when I want an agent to walk into a
project, recover local project memory, choose the right workflow, execute work in
vertical slices, test its own changes, review the result, and leave durable
handoff/context files behind.

## Token-Minimizing Design

The core purpose is to reduce wasted context without lowering the engineering
bar.

- **Fresh sessions by default.** Handoffs, `todo.md`,
  `docs/context/PROJECT.md`, plans, and architecture notes let a new session
  recover without carrying days of chat history.
- **Map once, then load narrowly.** `kb-map` builds or refreshes project memory,
  then future sessions follow exact pointers instead of crawling the repo.
- **Choose the smallest correct lane.** `kb-start` routes by task shape:
  `kb-fix` for small known bugs, `kb-troubleshoot` for unclear broken behavior,
  `kb-brainstorm` for unclear ideas, `kb-plan` for slicing known work,
  `kb-work` for execution, and `kb-epic` only when the work is truly large.
- **Spend tokens where they prevent rework.** Slicing, checks, and review cost
  tokens up front, but they are cheaper than redoing broken or under-tested
  work later.

KB means **Kanban-Based**. The workflow still uses boards, manifests, vertical
slices, and done archives, but user-facing commands use the shorter `kb-`
prefix because it works better with voice input.

## What Is Installed

This is not the full ATV StarterKit. It is the smaller KB overlay that should be
safe to copy into active projects without dragging in every experiment or
historical workflow.

The installed runtime surface is intentionally smaller than the repository:
about 37 skills plus 52 reviewer/specialist agents.

Installed/runtime surface:

- `.github/skills/*/SKILL.md` - portable skills
- `.github/agents/*.agent.md` - reviewer and specialist agents
- `AGENTS.md` - Codex/agent repo contract
- `.github/copilot-instructions.md` and `.github/instructions/*.instructions.md`
  - Copilot guidance
- `cmd/kbcheck` - Go quality/release gate entrypoint

Development scaffolding that is usually not copied into consuming projects:

- `docs/` - this bundle's own memory and reference docs
- `evals/` - route, quality, live-adapter, and benchmark fixtures
- `scripts/` - deterministic helper/scorer/sync scripts
- `config/` - skill quality, marketplace, and pipeline config

Consuming projects get their own `todo.md`, `docs/context/`,
`docs/handoffs/`, eval map, and project-local memories.

## Quick Start

Install or sync the bundle globally, then start work inside a target repo:

```powershell
cd E:\path\to\your\project
kb-start "what I want done"
```

Normal flow:

```text
kb-start -> kb-map -> chosen lane
```

For a fully hands-off feature flow:

```text
klfg: kb-brainstorm -> kb-plan -> kb-work -> kb-complete
```

`kb-work` now owns the loop until the work is terminal: all runnable slices are
done or intentionally skipped, then `kb-complete` runs review, follow-up
resolution, proof, learning, memory refresh, and cleanup. "All slices passed" is
progress; `kb-complete` is the done gate.

## Common Commands

| Command | Use When |
| --- | --- |
| `kb-start` | Fresh session, ambiguous ask, or "figure out the right workflow" |
| `kb-task` | First-principles task runner that continues until verified or blocked |
| `kb-map` | Setup, lookup, or refresh project memory |
| `kb-eval-map` | Map repo-native eval surfaces and proof commands |
| `kb-fix` | Narrow bug, failing test, or small contained change |
| `kb-troubleshoot` | Broken behavior needs logs/browser/test investigation |
| `kb-brainstorm` | Product or technical framing is still unclear |
| `kb-plan` | Requirements exist and need vertical slices |
| `kb-work` | A manifest exists and should be executed |
| `kb-review` | KB-specific code review with structural quality review |
| `kb-complete` | Work needs review, proof, learning, memory, cleanup |
| `kb-memory-review` | High-cost pass for stale, bloated, or contradictory memory |
| `kb-ship` | Release, PR, deploy, or final readiness check |
| `kb-epic` | Large migration, rewrite, or multi-brainstorm initiative |
| `klfg` | Fully hands-off route from brainstorm through completion |
| `repo-critic` | Claims-vs-code evidence review before a claim ships |

## Project Memory

The workflow keeps memory in files so sessions can stay short.

Required consuming-project memory:

- `todo.md` - active work, blockers, parked work, handoff pointers
- `todo-done.md` - compact archive of completed work
- `docs/context/PROJECT.md` - fresh-session route map
- `docs/context/eval-map.md` - repo-native eval surfaces and proof commands
- `docs/context/architecture/` - architecture notes by domain
- `docs/context/operations/` - run/test/deploy/QA commands
- `docs/handoffs/active/` - resumable work
- `docs/handoffs/parked/` - valuable work that is not runnable today
- `docs/handoffs/done/` - completed or superseded handoffs

`kb-map` resolves the active project root first and reads memory only from that
repo. It must not search `~`, `.copilot/handoffs`, the whole drive, or sibling
repos unless explicitly asked for cross-repo lookup.

Deep dive: [KB workflow architecture](docs/context/architecture/kb-workflow.md).

## Review Agents

The reviewer agents are runtime dependencies, not optional docs. Removing them
causes `document-review`, `kb-review`, `ce-review`, `kb-complete`, and related
gates to fail or degrade.

Always-on KB code review personas:

- `correctness-reviewer`
- `testing-reviewer`
- `thermo-nuclear-code-quality-reviewer`
- `project-standards-reviewer`

Conditional reviewers include security, performance, API contracts, migrations,
reliability, frontend races, schema drift, deployment, prior comments,
language-specific reviewers, and adversarial review.

Document-review uses its own lens agents: coherence, feasibility, product,
design, security, scope, and adversarial document review.

Deep dive: [KB workflow architecture](docs/context/architecture/kb-workflow.md)
and [kb-review persona catalog](.github/skills/kb-review/references/persona-catalog.md).

## Quality Gates

Run before propagating skill changes:

```powershell
go run .\cmd\kbcheck core
```

Run before releasing or syncing globals:

```powershell
go run .\cmd\kbcheck local-release
```

`local-release` composes deterministic local proof: native `core`, sync drift,
line-ending checks, static reports, and the available local eval surfaces.
`live-release` is explicit:

```powershell
go run .\cmd\kbcheck live-release
```

Live mode may call authenticated Codex/GHCP CLIs. A local green gate is not a
claim that live model evals ran.

The current top-level gate is Go. Several individual validators are still
PowerShell scripts, so PowerShell 7 remains required for the full local suite.

Deep dive: [testing operations](docs/context/operations/testing.md) and
[eval map](docs/context/eval-map.md).

## Install

Default to personal/global installs. They keep active project repos clean and
avoid skill drift between copies.

GitHub Copilot personal install:

```powershell
$src = 'E:\working-skill-repo'
Copy-Item "$src\.github\skills\*" "$env:USERPROFILE\.copilot\skills" -Recurse -Force
Copy-Item "$src\.github\agents\*" "$env:USERPROFILE\.copilot\agents" -Force
```

Codex personal install:

```powershell
$src = 'E:\working-skill-repo'
Copy-Item "$src\.github\skills\*" "$env:USERPROFILE\.codex\skills" -Recurse -Force
Copy-Item "$src\.github\agents\*" "$env:USERPROFILE\.codex\agents" -Force
```

Shared agent-skills install:

```powershell
$src = 'E:\working-skill-repo\.github\skills'
Copy-Item "$src\*" "$env:USERPROFILE\.agents\skills" -Recurse -Force
```

Use repo-local installs only when a project needs pinned/project-specific
overrides or when the skills should be versioned with that codebase.

Deep dive: [skill bundle maintenance](docs/context/operations/skill-bundle-maintenance.md).

## Platform Reality

This repo supports Codex and GitHub Copilot/GHCP instruction surfaces. The
development machine is Windows, so examples use Windows paths.

Current state:

- Go owns top-level gate orchestration.
- PowerShell 7 is still required for individual validators.
- Windows parity smoke proof is recorded in `docs/reports/go-gate-parity-2026-06-01.md`.
- macOS/Linux should use the same Go entrypoint plus `pwsh`, but full OS proof
  is still parked until those machines are available.

## Marketplace And Security

`E:\agent-marketplace` is a private approved catalog, not a global install. New
skills and pipelines should prove themselves project-local first, then move into
the marketplace only after evidence, review, hash pinning, and human approval.

Public imports go to quarantine first. Quarantine is an enforced firebreak:
active and approved skill roots must not resolve into quarantine.

`atv-security` is the current approved ATV security skill. Dependency
vulnerability proof prefers OSV Scanner machine evidence when `osv-scanner` is
installed.

Deep dive:

- [private skill marketplace](docs/context/architecture/private-skill-marketplace.md)
- [skill bundle maintenance](docs/context/operations/skill-bundle-maintenance.md)

## What Is Not Bundled

These are intentionally left out of the minimal working bundle:

- upstream `deepen-*` passes; use `kb-research` and proportional research
- one-shot LFG/SLFG style workflows; use `klfg` only when you want the full
  pipeline
- upstream `workflows-*` aliases; use KB lanes directly unless a current app
  explicitly needs an ATV alias
- `land`; shipping remains a deliberate separate decision
- browser tools such as `agent-browser`; skills can call them when installed,
  but this repo does not vendor them

The useful LFG finish pattern is preserved inside `kb-complete`: resolve
follow-up review/TODO work, rerun proof on the final diff, capture demo evidence
when useful, then compound, learn, evolve, refresh memory, compact, clean up, and
alert.

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
