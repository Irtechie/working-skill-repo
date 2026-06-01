# KB Workflow Architecture

This document holds the workflow detail that used to live inline in the root
README. The README is the front door; this file is the operating model.

## Fresh Session Loop

The workflow is meant to make every new task safe to start in a fresh session:

1. Finish or pause the current task with a handoff.
2. Close the old session.
3. Start a new session in the project repo.
4. Run `kb-start <next task or handoff>`.

`kb-start` calls `kb-map`, which reads local project memory and points the new
session to the specific files it needs. The handoff tells the model what work is
being resumed; `docs/context/PROJECT.md` tells it what the app is and where the
relevant architecture docs live.

## Route Selection

`kb-start` is the workflow router. It chooses the lane for the actual work, not
the ceremony implied by the user's wording.

Every request starts by calling `kb-map lookup <request>` so the session has
current project memory before route selection.

Typical routing:

| Shape | Lane |
| --- | --- |
| small known bug, typo, narrow cleanup | `kb-fix` |
| broken behavior with unclear cause | `kb-troubleshoot` |
| unclear product or technical framing | `kb-brainstorm` |
| requirements exist and need slices | `kb-plan` |
| valid manifest exists | `kb-work` |
| all slices are done and need completion gates | `kb-complete` |
| release, PR, deploy, or final readiness | `kb-ship` |
| multi-subsystem initiative or migration | `kb-epic` |
| external docs or prior art could change the decision | `kb-research` |

The goal is proportional ceremony. A typo fix should not become a brainstorm; a
framework migration should not become a quick fix.

## Map And Bootstrap

`kb-map` is the context router for fresh sessions.

It resolves the active project root, checks standard memory files, and loads
only the relevant pointers:

- `todo.md` for current work, blockers, parked items, and handoff links
- `docs/context/PROJECT.md` for the app map and subsystem index
- `docs/context/architecture/*` for subsystem detail
- `docs/context/operations/*` for run, test, QA, and deploy commands
- `docs/handoffs/*` for resumable work packets

`docs/context/PROJECT.md` is the entry map. It explains what the app is, how to
run and test it, what major subsystems exist, and which subsystem documents to
read next.

When memory is missing, `kb-map` invokes `kb-map-bootstrap` to build the project
map once. Bootstrap inventories the repo, reconciles discovered systems against
`PROJECT.md` and `docs/context/architecture/README.md`, runs `kb-eval-map`, and
route-tests every mapped major area.

Bootstrap must discover concepts, not just folders. It descends into substantial
child directories, clusters cross-cutting concerns, mines repo memories and
AGENTS/README files for subsystem hints, checks route/page and filename-prefix
patterns, and records known-unknowns.

Bootstrap also uses `kb-map-bootstrap/scripts/code-intel.ps1` when available.
That helper samples symbols, likely entry points, largest files, extension
counts, and language-server availability. It is a precision boost, not a
mandatory LSP dependency.

## Project Memory Contract

Required memory files in consuming projects:

- `todo.md`
- `todo-done.md`
- `docs/context/PROJECT.md`
- `docs/context/eval-map.md`
- `docs/context/architecture/`
- `docs/context/research/`
- `docs/context/decisions/`
- `docs/context/operations/`
- `docs/handoffs/active/`
- `docs/handoffs/parked/`
- `docs/handoffs/done/`

`todo.md` is not a history file. Keep board rules at the top of `todo.md`. When
a feature, slice group, handoff, or fix is complete, move the compact summary to
`todo-done.md` and remove completed routine logs from `todo.md`.

## Execution Model

The pipeline is designed around task sizes:

- **Small known bug:** use `kb-fix`.
- **Broken behavior with unclear cause:** use `kb-troubleshoot`.
- **Bounded autonomous task:** use `kb-task`.
- **Medium feature:** use `kb-brainstorm` -> `kb-plan` -> `kb-work`.
- **Large initiative:** use `kb-epic`.

`kb-plan` produces vertical slices with expected files, verification,
dependencies, test level, functional risk, and HITL flags.

`kb-work` executes runnable slices in dependency order. Once execution starts, it
does not ask before each slice. The default WIP cap is one active slice per
checkout/model context; parallel exceptions require isolated contexts, disjoint
file scopes, and a manifest note before the extra slice starts. It pauses only
for real gates: HITL, destructive approval, blocked/human-required work, scope
failures, QA/repair exhaustion, dependency deadlock, or explicit user stop.

`kb-work` is not done when slices pass. It must invoke `kb-complete` after all
runnable slices are done or intentionally skipped.

`kb-complete` owns the terminal half of the loop:

- deterministic final checks
- `kb-review`
- P0/P1 resolution
- follow-up resolution
- proof/demo evidence
- compound/learn/evolve
- project memory refresh
- memory maintenance signals
- cleanup

## Verification

`kb-check` and `kb-functional-test` push verification into code whenever
possible. The model should run deterministic checks instead of spending tokens
re-inspecting behavior by hand.

`kb-functional-test` owns the test-level decision:

- `none`
- `unit`
- `integration`
- `functional-api`
- `functional-cli`
- `functional-browser`
- `full`

For UI work, `functional-browser` is automatic when `.tsx`, `.jsx`, `.vue`, or
`.svelte` files change, or when backend/state behavior is primarily reached
through the app UI. Screenshots support evidence; executable assertions are the
pass/fail oracle.

`kb-regression-snapshot` records deterministic state after each passed slice in
`.atv/snapshots/<slice-id>.json`, then verifies prior snapshots before the next
slice starts.

`kb-complete` fails the proof gate when a slice only has prose proof. Each slice
needs command/test path, exit code, timestamp, trace/log/API artifact, or
snapshot verification evidence recorded in the manifest.

## Review Agents

`kb-review` uses a layered persona model.

Always-on:

- `correctness-reviewer`
- `testing-reviewer`
- `thermo-nuclear-code-quality-reviewer`
- `project-standards-reviewer`

Conditional:

- `security-reviewer`
- `performance-reviewer`
- `api-contract-reviewer`
- `data-migrations-reviewer`
- `reliability-reviewer`
- `adversarial-reviewer`
- `cli-readiness-reviewer`
- `previous-comments-reviewer`
- language and framework reviewers
- schema/deployment/agent-native reviewers

Document review has a separate lens-agent set for coherence, feasibility,
product, design, security, scope, and adversarial review.

## Token Diet

Heavy inherited ATV/CE skills keep routing and safety rules in `SKILL.md`, but
detailed phase mechanics live under `references/` and are loaded only when that
phase is running.

Do not move a rule out of `SKILL.md` if missing it would make the skill choose
the wrong lane, mutate files unsafely, or skip a required gate. Move details out
when they are only needed after the lane or phase is already chosen.
