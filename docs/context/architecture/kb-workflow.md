# KB Workflow Architecture

This document holds the workflow detail that used to live inline in the root
README. The README is the front door; this file is the operating model.

## Fresh Session Loop

The workflow is meant to make every new task safe to start in a fresh session:

1. Finish or pause the current task with a handoff.
2. After durable delivery, register terminal cleanup and release the work claim.
   A coordinator or later session removes the old Git worktree; the current
   executing session never deletes itself.
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
| reviewed work needs commit, push, and PR | `kb-ship` |
| plan/manifest should reach done-done and a checked-in PR | `kb-finish` |
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
- `kb-work`, `kb-finalize`, and `kb-complete` advance only through manifest gate-ledger records,
  not chat confidence.
- `kb-complete` is the state-aware orchestrator for the full loop:
  `brainstorm when needed -> kb-plan -> kb-work -> kb-finalize -> delivery`.

The deterministic maintainer proof is:

```shell
go run ./cmd/kbcheck workflow-governor-selftest
```

`go run ./cmd/kbcheck core` includes that selftest.

### Blocker Lifecycle

Workflow state distinguishes execution control from technical proof:

| State | Meaning | Propagation |
|---|---|---|
| `paused` | The user preemptively stopped execution | No new dispatch, polling, late-result reads, cleanup, commits, or gate mutation; a requested handoff may record paused state |
| `blocked` | A current agent, dependency, tool, or access impasse remains after recheck | Only dependent work stops |
| `needs-human` / `human-required` | Only the user can authorize, supply, access, risk-accept, or judge the next step | Only the affected decision and its dependents stop |
| `repairing` / `in_progress` | The agent can still safely fix or verify the failure | Work continues |

New manifests may opt into `blocker_lifecycle_contract: true`. Each nonterminal
gate then records scope, responsibility, affected scope, resume condition,
recheck sensor, check time, and propagation. `paused` is rejected as a gate
result because the underlying technical gate must keep its real state.

Before a summary, handoff, or completion rollup repeats a blocker, the owning
skill reruns the named sensor or cheapest current-state probe. Release,
deployment, signing, and optional-capability failures remain scoped to their
promotion or capability; they do not turn proven implementation into a
whole-objective failure.

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

When the broken behavior has a repeatable sensor, use the proof spine:
`kbcheck sense` records the RED and GREEN observations, `kbcheck trace-verify`
checks the hash chain, and `kbcheck accept` is the preferred repair proof. A
latest-green check without a recorded prior RED is not enough for a repair
claim.

`kb-complete` is one state-aware run from source through configured delivery.
`kb-goal` is the durable-objective lane: it may run `kb-complete`, `kb-epic`, `kb-task`, or
several manifests over days, but it completes only when the goal ledger's
terminal proof matches the original objective. Under a goal, brainstorm stops are minimized:
the agent resolves the best path from repo evidence, research, and safe
assumptions, and asks only for true `ask-now` blockers.

User interruption has higher priority than that loop. `pause` ends goal activity
immediately. `pause`, `stop`, `cancel`, or `end` suspends activity before asking
whether to park the goal permanently or keep it paused. The controller does not
poll agents or sessions, consume late results, run cleanup, commit, or dispatch
more work while awaiting confirmation.
Parent pause and stop authority is non-overridable: sibling sessions, child
sessions, coordinators, queued messages, and late results cannot authorize
successor work or a commit. Only an explicit user resume addressed to the owning
goal can reactivate it. A permanent end parks—not completes—the ledger and
returns the exact `To resume: /kb-goal <objective>` command.

`kb-work` auto-invokes only `kb-finalize`, which cannot publish. Explicit
`kb-complete` applies project delivery policy after finalization. `kb-finish`
and `klfg` remain compatibility aliases.

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

## KB Run State

Long-running goals may create ephemeral control-loop state under
`.kb/runs/<goal-slug>/`. This borrows the useful persistence idea from
Phoenix/Ralph-style loops without adopting a separate runtime, MCP server, or
`.phoenix-ralph` directory.

`.kb/runs` is git-ignored and never replaces durable human surfaces. The durable
truth remains:

- goal ledgers in `docs/context/goals/`
- `todo.md` and `todo-done.md`
- KB manifests and slice plans in `docs/plans/`
- handoffs in `docs/handoffs/`

A run directory uses this shape:

| File | Purpose |
|---|---|
| `goal.md` | Pointer to the durable goal ledger and current objective |
| `done-check.json` | Optional `kbcheck sense/accept` check spec |
| `backlog.json` | Candidate work units with route, priority, blockers, and source |
| `progress.md` | Current state, last accepted proof, next allowed action |
| `route-history.jsonl` | Route decisions with confidence and progress signals |

Each route-history row should include `ts`, `route`, `confidence`, and either
`state_changed` or `progress_key`. Example:

```json
{"ts":"2026-07-09T15:00:00-04:00","route":"kb-work","confidence":0.82,"state_changed":true,"progress_key":"slice-003-done"}
```

The deterministic guard is:

```powershell
go run ./cmd/kbcheck run-state --history .kb/runs/<goal-slug>/route-history.jsonl
```

It flags A/B/A/B route oscillation, three low-confidence route choices with no
progress, and four no-progress route decisions. A failure means the agent should
refresh context, re-plan, or ask a focused human question instead of continuing
to bounce between lanes.

`kb-plan` produces vertical slices with expected files, verification,
dependencies, test level, functional risk, model tier, and HITL flags. Model
tier records minimum execution capability (`small`, `medium`, `large`; legacy
`tiny` maps to `small`). It never lowers the executable proof requirement and
does not freeze a worker.

Plans contain tier, requirements, risk, and proof only. They never name a model,
route alias, source preference, adapter, endpoint, or transport. The
orchestrator owns planning, minimum-tier judgment, selection, supervision,
proof, and synthesis. One qualified same-tier-or-higher subagent normally owns
each bounded slice. This is one owner per slice, not one worker per plan: every
safe independent ready slice may run on its own tier-qualified subagent in
parallel when dependencies, writes, and shared resources are isolated. The
orchestrator may retain execution only through a recognized reason gate:
required reasoning, accumulated context, tools, authority, trust, an explicit
user requirement, or a proved lack of qualified routes. The receipt records
that owner and the actual route. Only run-scoped `require <model>` hard-pins a
delegated route.

The tier is portable across hosts. When a plan is picked up, any compatible CLI
or host maps each ready slice to an exact live route on its own callable
surface. Different runners may choose different concrete models for the same
tier without changing the plan.

Ordinary map/bootstrap and native-only work ask no routing questions. Explicit
`kb-models` setup may add user-local OpenAI-compatible/LiteLLM routes whose
alias resolves to the current model, adapter, endpoint, and auth reference.
Generic MCP model dispatch is not a current capability. Ordinary work silently
uses `automatic` when no project source choice is saved. Only explicit setup or
configuration offers `automatic`, `self-hosted-first`, or `native-first`. Save
only that source preference through user-local `kb-models`; connection details remain local.

Host-native and CLI/local delegation are distinct executable branches:

```text
required tier + bounded packet
  -> recognized current-owner exception? current execution
  -> otherwise inspect native host and CLI/user-local routes
  -> select one qualified same-tier-or-higher subagent
  -> no qualified route? record no-qualified-route and validate current
  -> deterministic proof accepts or rejects the result
```

The active host's callable schema is authoritative for native targets.
`kbrouter` is authoritative for CLI and user-local routes. The orchestrator
never infers callable targets from model memory, never merges identities by
display name, and never treats a host-only target as CLI-callable or vice versa.

Immediately after ownership/route selection and before mutation or dispatch,
`kb-work` emits one compact line with `current|subagent`, primary route,
parent-return behavior, required tier, and proof target. The route announcement is evidence-bound:
concrete names require active host or `kbrouter` evidence.
An eligible local route gets one attempt. Failure returns directly to the active
parent; do not select a second local route.

```text
DDR route: <current|subagent> | primary: <current orchestrator|evidence-backed-route> | return: <none|parent-on-first-local-failure|required-alias-block> | tier: <small|medium|large> | proof: <short-proof-target>
```

Deterministic proof is the validator. A selected owner repairs ordinary bounded
failures without silently handing the slice to the other owner. If required
authority changes, `kb-work` re-plans, blocks, or records a new explicit
ownership decision. Local DDR never chooses a second local route. The active
parent continues with its current model or host-native selector after a
structured parent return.

Adaptive Model Routing (AMR) remains an unpromoted experimental benchmark.
Normal work never passes `attempt_tier`, never requires a lower-tier trial, and
never consumes an AMR experiment flag. `kb-configure` can opt a separate
benchmark in or out. Advanced run-scoped `use`, `require`,
`prefer self-hosted` (`prefer local` shorthand), `prefer native`, and
`ignore model routing` controls remain available through `kb-models`.

`kb-work` executes the safe ready set from the slice dependency DAG. Once
execution starts, it does not ask before each slice. The default WIP is every
ready slice that can run in an isolated context without a serial-only gate.
Shared-checkout mutation still runs one slice at a time, and observed write
overlap serializes or requeues one of the colliding slices. `expected_files` is
a forecast, not the safety oracle. `kb-work` pauses only for real gates: HITL,
destructive approval, blocked/human-required work, scope failures, QA/repair
exhaustion, dependency deadlock, observed overlap that cannot be safely
serialized, or explicit user stop.

Workspace isolation has two parts. The plan records durable intent:
`workspace_mode`, conflict domains, shared resources, and integration
dependencies. It does not record live paths, branches, sessions, cleanup state,
or owner tokens. At work time, `kb-work` prepares one manifest-owned plan
worktree and non-default integration ref, then acquires a manifest lease under
the Git common directory. Preparation requires explicit local check-in
authorization and records it in the receipt; without that authority the run
stops before mutation. Every slice for that manifest commits sequentially in
that same worktree; there are no per-slice worktrees or merges into the source
checkout. The legacy `worktree` command requires an explicit compatibility flag
and rejects active plan runs. The coordinator keeps the slice lease active,
projects audited lifecycle state into the same commit as implementation, reruns
slice and aggregate proof, requires the commit diff to exactly match the proof
receipt plus slice/plan claims, archives immutable proof bytes and SHA-256, then
advances the receipt's integration head with compare-and-swap.
Dirty, stale-head, wrong-owner, wrong-ref, or wrong-worktree acceptance failures
leave the receipt and worktree intact for recovery.

Terminal completion runs under the shared state lock. It requires every slice
and terminal gate plus `work-to-complete` to pass, durable explicit release
evidence for every done slice lease, an active matching plan-run lease, every
accepted-proof archive to remain present with its recorded digest, and the
worktree HEAD/ref to remain at the accepted integration head. Completion then
atomically releases the plan-run lease; direct release of a manifest-owned plan
lease is refused. Worktree release rechecks the same CAS before non-force
removal.

Configured delivery owns the later terminal-session boundary. `kb-complete`
registers a cleanup receipt under the Git common directory only after local
branch durability, pushed-topic containment, or remote-default containment is
proven. It then releases the shared work claim as done. A different
`kb-start` session runs `terminal-cleanup --action sweep --session-id
<current-project-session-id>`, which rechecks the queue, exact branch/HEAD,
tracked/untracked/ignored status, Git worktree registration, delivery
containment and monotonic observed ref SHAs, Git-admin generation/round-trip,
real path, primary/default boundaries, and current-session ID/path exclusion
before non-force removal. It holds the queue lock across the final claim reread,
refreshes remote-default evidence immediately before each worktree removal, and
refreshes again before direct-mode local-ref deletion. Unresolved remote
authority blocks cleanup, including local-only repositories with no configured
remote. Local-only completion keeps its branch; the PR-only
endpoint always keeps local and remote refs; a separately registered integrated
endpoint may delete only the exact matching local feature ref. Remote ref
deletion and host UI session records remain provider/host-owned.

Multiple manifests may run concurrently only while their live file, prefix,
conflict-domain, and shared-resource claims remain disjoint. A collision reports
the owning run and requeues before mutation. Delivery is a later explicit
boundary: local mode leaves the candidate branch intact; PR/manual mode may
prepare a candidate but does not merge, push, or create a PR during work.

The lease boundary is local. It covers sessions and sibling worktrees that share
the same Git common directory. Separate clones, machines, or copied checkouts
must not be treated as sharing ownership state unless a future distributed
coordinator is explicitly implemented and proved.

The deterministic lifecycle proof is:

```powershell
go run ./cmd/kbcheck plan-worktree-selftest
go test ./cmd/kbcheck -run TerminalCleanup -count=1
```

It creates a disposable Git repository through the public fresh-start path,
executes two disjoint plan runs with
two serialized commits each, proves collision ownership and recovery
invariants, verifies the source SHA and dirt are unchanged, and rejects any
target that resolves to or contains the real repository.

Graph routing has the same evidence boundary. `kb-map` may return a compact
impact packet summary with packet ID, repository freshness, fallback, evidence
counts, impacted files/symbols, tests/docs, limitations, and a selected
traversal recipe. `kb-plan` records those as forecasts and conflict domains.
`kb-work` reconciles observed files against the forecast and records provenance
for necessary expansion. `kb-review` checks missed consumers/tests/docs and
unexplained scope growth. A graph packet, worker receipt, route receipt, or
lease receipt never replaces source verification or the slice proof command.

Exact-symbol and structural/flow providers are adapters. They can raise
confidence when fresh and source-citable, but stale, unavailable, or heuristic
results fall back to file-native inspection and must be labeled with their
limitations.

`kb-work` is not finalized when slices pass. It must invoke `kb-finalize` after
all runnable slices are done or intentionally skipped.

`kb-finalize` owns the post-work quality half of the loop:

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

`kb-complete` is the single user-facing state-aware orchestrator. It can begin
from a feature description, plan, active manifest, or reviewed manifest; it
delegates planning, work, and finalization, then applies project delivery
policy. `kb-ship` owns internal PR delivery and `kb-land` owns explicit
merge/direct integration plus configured post-integration sync. Legacy `klfg`
and `kb-finish` delegate to `kb-complete`.

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
`.kb/snapshots/<slice-id>.json`. Before another slice, the proof governor
selects snapshots affected by the pending diff. The complete set runs once at a
named manifest/release milestone and becomes reusable while its definition
fingerprint is unchanged.

Registered proof checks carry sealed coverage, relevant working-tree inputs,
command/environment semantics, namespace, execution class, timeout, and age.
`kbcheck proof-plan` returns `RUN`, `REUSE`, or `BLOCK`; `proof-run` writes
content-addressed receipts and a bounded decision audit. Unknown impact runs
conservatively. Passing superset proof is reusable only for covered requests
with identical relevant inputs. An unchanged failed fingerprint blocks another
attempt until code/input changes.

Headless browser proof remains agent-runnable. The portable proof runner always
blocks visible-browser and native-GUI execution before spawn. Any explicitly
requested attended GUI session is a bounded user/host action outside
`proof-run`, with its evidence recorded separately.

`kb-complete` fails the proof gate when a slice only has prose proof. Each slice
needs command/test path, exit code, timestamp, trace/log/API artifact, or
snapshot verification evidence recorded in the manifest. For repaired failures,
`kbcheck accept --check <check.json> --trace .kb/trace.jsonl` is the canonical
RED-before-GREEN proof when the check is expressible as JSON.

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
