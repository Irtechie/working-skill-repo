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
| durable objective across sessions | `kb-goal` |
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

Not every route produces a planned slice. Planned slices are for manifest work
owned by `kb-plan` and executed by `kb-work`. `kb-fix` and `kb-troubleshoot`
still plan before editing, but their plan is a compact reproduction/diagnostic
plan with lane-local proof, not a manifest, unless the bug expands into
multi-slice work.

Every phase handoff must be explicit for hosts that do not auto-chain skills.
After a gate-clean brainstorm, ask whether to continue with
`kb-plan <requirements-doc>` unless execution intent or an orchestrator already
authorized continuing. After planning, ask whether to continue with
`kb-work <manifest-path>` unless execution intent is already present. If the
host cannot invoke the next skill, print the exact `Next command:` and stop.

## Workflow Governor

The KB workflow governor is the contract that keeps an agent from assuming,
skipping phases, or claiming done without proof.

Enforced by skills and artifacts today:

- `kb-brainstorm` owns the Question Gate before planning. Material unknowns are
  classified as `ask-now`, `research-first`, `safe-assumption`,
  `defer-to-planning`, or `parked`.
- `ask-now` and unresolved `research-first` items block planning.
- `safe-assumption` items may pass only when they name evidence,
  reversibility, and the later proof that would catch a wrong assumption.
- `kb-plan` refuses to slice source material that still contains unresolved
  brainstorm blockers.
- `kb-work` and `kb-complete` advance only through manifest gate-ledger records,
  not chat confidence.
- `klfg` is the strict orchestrator for the full loop:
  `kb-brainstorm -> kb-plan -> kb-work -> kb-complete -> DONE`.

The deterministic maintainer proof is:

```shell
go run ./cmd/kbcheck workflow-governor-selftest
```

`go run ./cmd/kbcheck core` includes that selftest.

Not shipped yet: platform hook enforcement that blocks a Codex/Claude stop or
prompt transition at runtime. The hook layer should mirror the same gate
classes and ledger checks once the target runtime hook files are implemented.
Until then, do not claim hook-enforced phase blocking.

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
- **Long-running objective:** use `kb-goal` to keep the durable goal ledger,
  then route each unit through the smallest valid KB lane.
- **Medium feature:** use `kb-brainstorm` -> `kb-plan` -> `kb-work`.
- **Large initiative:** use `kb-epic`.

`kb-fix` and `kb-troubleshoot` both require agent-run verification. The proof is
not just "the edit looks right"; rerun the reproduction plus the relevant tests,
browser checks, CLI/API probes, or logs that prove the broken behavior changed.
They also require a compact pre-edit plan that freezes the reproduced signal,
likely target, protected oracle/test files, and verification command. That plan
is deliberately smaller than a `kb-plan` manifest; route to `kb-plan` only when
the fix becomes multi-slice or needs dependency ordering.

`klfg` is one strict pipeline run from brainstorm through completion. `kb-goal`
is the durable-objective lane: it may run `klfg`, `kb-epic`, `kb-task`, or
several manifests over days, but it completes only when the goal ledger's
terminal proof matches the original objective. Under a goal, brainstorm stops are minimized:
the agent resolves the best path from repo evidence, research, and safe
assumptions, and asks only for true `ask-now` blockers.

For recurring or trend-improvement goals, `kb-goal` may add a live-steering
block to the goal ledger. That block names the set point, sensor, controller,
actuator, disturbances, optional dampener, scope gate, batch size, WIP bound,
and steering-memory path. This is a control-loop framing for repeated work, not
a requirement for one-shot goals. If one repo tool or agent prompt naturally
fuses sensor, controller, and actuator, the ledger records the fused component
instead of inventing fake artifacts.

Steering memory is the middle layer between a one-off PR comment and a promoted
project instinct. It stores concise durable feedback that should influence
future runs: permanent scope exclusions, known false positives, reviewer
preferences, and target-selection guidance. It lives either in the goal ledger
or in `docs/context/operations/steering/<slug>.md` when the guidance would bloat
the ledger. Raw transcripts, single-run logs, and current-PR-only instructions
do not belong there.

`kb-plan` produces vertical slices with expected files, verification,
dependencies, test level, functional risk, and HITL flags.

`kb-work` executes the safe ready set from the slice dependency DAG. Once
execution starts, it does not ask before each slice. The default WIP is every
ready slice that can run in an isolated context without a serial-only gate.
Shared-checkout mutation still runs one slice at a time, and observed write
overlap serializes or requeues one of the colliding slices. `expected_files` is
a forecast, not the safety oracle. `kb-work` pauses only for real gates: HITL,
destructive approval, blocked/human-required work, scope failures, QA/repair
exhaustion, dependency deadlock, observed overlap that cannot be safely
serialized, or explicit user stop.

`kb-work` is not done when slices pass. It must invoke `kb-complete` after all
runnable slices are done or intentionally skipped.

`kb-complete` owns the terminal half of the loop:

- deterministic final checks
- `kb-review`
- P0/P1 resolution
- follow-up resolution
- proof/demo evidence
- steering feedback classification
- compound/learn/evolve
- project memory refresh
- memory maintenance signals
- cleanup

The steering step classifies review, iteration, and maintainer feedback as
current-only, steering memory, observation, landmine candidate, or instinct
evidence. `learn` still owns scored instincts and `evolve` still owns skill
promotion; live steering only changes how future runs are selected and prompted.

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
