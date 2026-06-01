# Working Skill Repo

Voice-friendly KB workflow skills and required reviewer agents for GitHub
Copilot and Codex.

Status: actively used, pre-1.0. The top-level gate is now native Go and has
Windows parity proof against the prior PowerShell wrappers; expect churn while
the marketplace, eval, and pipeline pieces settle.

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

The core purpose of this skill set is to reduce wasted context without lowering
the engineering bar.

It does that in four ways:

- **Fresh sessions by default.** Work is designed to survive session restarts.
  Handoffs, `todo.md`, `docs/context/PROJECT.md`, plans, and architecture notes
  let a new session recover the project without carrying three or four days of
  chat history.
- **Map once, then load narrowly.** `kb-map` builds or refreshes the project
  memory, then future sessions use exact pointers instead of crawling the repo
  or making the user reteach the app.
- **Choose the smallest correct lane.** `kb-start` routes by actual task shape:
  `kb-fix` for small known bugs, `kb-troubleshoot` for unclear broken behavior,
  `kb-brainstorm` for unclear ideas, `kb-plan` for slicing known work,
  `kb-work` for execution, and `kb-epic` only when the work is truly large. The
  goal is not to turn every request into a ceremony.
- **Spend tokens where they prevent rework.** Vertical slicing and functional
  verification cost tokens up front, but they are cheaper than redoing broken or
  under-tested work later. The target is not the fewest tokens per turn; it is
  the fewest wasted tokens per finished, verified change.

The tradeoff is intentional: slicing, checks, and review may add overhead, but
they keep the agent from taking shortcuts, losing context, or making the user
become QA.

KB means **Kanban-Based**. The underlying workflow still uses boards, manifests,
vertical slices, and done archives, but user-facing commands use the shorter
`kb-` prefix because it works better with voice input.

## Fresh Session Loop

This workflow is meant to make every new task safe to start in a fresh session:

1. Finish or pause the current task with a handoff.
2. Close the old session.
3. Start a new session in the project repo.
4. Run `kb-start <next task or handoff>`.

`kb-start` calls `kb-map`, which reads the local project memory and points the
new session to the specific files it needs. The handoff tells the model what
work is being resumed; `docs/context/PROJECT.md` tells it what the app is and
where the relevant architecture docs live. The goal is to stop carrying days of
chat history just so the model remembers what the app is.

## What This Repo Is

This is not the full ATV StarterKit. It is the smaller KB overlay that should be
safe to copy into active projects without dragging in every experiment or
historical workflow. Original ATV skills stay in the ATV repo; selected shared
skills are mirrored here only when KB depends on them. The reviewer agents are
still part of the required runtime surface; `document-review`, `kb-review`,
`ce-review`, `kb-complete`, and related gates fail or degrade without them.

It includes:

- repo-local GitHub Copilot skills in `.github/skills/`
- repo-local Copilot agents in `.github/agents/`
- root agent guidance in `AGENTS.md`
- Copilot guidance in `.github/copilot-instructions.md`
- a deterministic Go check helper at `cmd/kbcheck`

The installed runtime surface is intentionally smaller than the repository:
about 36 skills plus 52 reviewer/specialist agents. The repo-local `docs/`,
`evals/`, `scripts/`, and `config/` folders are development scaffolding for this
portable bundle. They are not normally copied into consuming projects when you
install the skills. Consuming projects get their own `todo.md`,
`docs/context/`, `docs/handoffs/`, eval map, and project-local memories.

## Platform Reality

This repo is Go-first for top-level gate orchestration, with PowerShell still
used by several individual validator scripts.

- The canonical quality gate is `go run .\cmd\kbcheck core`.
- `cmd/kbcheck` natively orchestrates `core`, `local-release`, and
  `live-release`; it no longer delegates to the old `kb-check.ps1` or
  `kb-release-gate.ps1` wrappers.
- Install examples use Windows paths because the active development machine is
  Windows.
- The live eval adapters shell to local Codex and GitHub Copilot CLIs and assume
  those CLIs are installed and authenticated.

Codex and Copilot are both supported runtimes for the skill instructions.
Windows parity for the Go gate is recorded in
`docs/reports/go-gate-parity-2026-06-01.md`. macOS/Linux should be able to use
the same Go entrypoint, but full OS proof still requires running it on those
machines. PowerShell 7 remains required for individual validator scripts that
have not been ported yet.

The default entry point is `kb-start`. In normal use, ask for work in plain
language and let `kb-start` choose the ceremony.

`kb-start` delegates project-memory setup to `kb-map`. Route decides what the
user's idea/request needs; map decides whether local memory needs lookup,
refresh, or first-time bootstrap.

`kb-start` replaces the older route name. The behavior is the same front
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

## Why KB Start Exists

`kb-start` is the workflow router. Its job is to choose the right lane for the
actual work, not blindly run the ceremony implied by the user's wording.

Every request starts by calling `kb-map lookup <request>` so the session has the
current project memory before it decides what to do. Then `kb-start` classifies
the work by task shape, risk, and available artifacts:

- Use `kb-fix` for small, bounded bugs or narrow changes where the likely fix is
  obvious and verification can prove it.
- Use `kb-troubleshoot` when broken behavior needs evidence gathering and
  self-correction. It must inspect local logs/tests/browser behavior and, when
  framework/tool/dependency behavior may matter, research current external docs,
  issues, changelogs, or known fixes before editing.
- Use `kb-brainstorm` when product behavior, technical framing, success
  criteria, or tradeoffs are still unclear.
- Use `kb-plan` when requirements or a handoff already explain the work and the
  next useful output is vertical slices.
- Use `kb-work` when a valid manifest and slice plans already exist.
- Use `kb-epic` when the request is too large for one brainstorm or manifest:
  migrations, framework rewrites, multi-subsystem initiatives, or long backlogs.
- Use `kb-research` only when external docs, prior art, framework behavior, or
  known failure modes could change the decision.

The goal is proportional ceremony. A typo fix should not become a brainstorm. A
framework migration should not become a quick fix. A clear handoff should not
rerun discovery just because the user said "brainstorm" casually. The user's
words are input; the route should be based on the actual task, the repo memory,
and the cost of being wrong.

## Why KB Map Exists

`kb-map` is the context router for fresh sessions.

The workflow assumes sessions are disposable. Instead of keeping one expensive
chat open for days, a new session should be able to enter a repo, ask "what am I
working on?", and get pointed to the exact local memory needed for the current
handoff, bug, feature, or plan.

`kb-map` makes that possible by resolving the active project root, checking the
standard memory files, and loading only the relevant pointers:

- `todo.md` for current work, blockers, parked items, and handoff links
- `docs/context/PROJECT.md` for the app map and subsystem index
- `docs/context/architecture/*` for the subsystem involved in the current task
- `docs/context/operations/*` for run, test, QA, and deploy commands
- `docs/handoffs/*` for resumable work packets

`docs/context/PROJECT.md` is the entry map. It explains what the app is, how to
run and test it, what major subsystems exist, and which subsystem documents to
read next. `docs/context/architecture/*.md` files are the deeper subsystem
notes. `kb-map` should read `PROJECT.md` first, then follow its pointers to the
smallest relevant architecture file for the current task.

Coverage matters. If a fresh session asks about a named high-risk workflow such
as installer, release, auth, workflows, actions, tools, runtime, or deployment,
`kb-map` must be able to point to the exact subsystem doc and source-of-truth
files without broad rediscovery. If it cannot, that is a map coverage gap:
targeted refresh should create or refine the missing child architecture doc and
record a memory-maintenance signal.

Bootstrap is responsible for the first coverage pass. It must inventory the
repo, reconcile discovered systems against `PROJECT.md` and
`docs/context/architecture/README.md`, and route-test every mapped major area.
One invisible subsystem is evidence to run a coverage audit, not to keep fixing
one doc at a time.

Bootstrap must also validate chains, not just describe files. For high-risk
systems like installers, releases, auth, data, integrations, and embedded
runtimes, it should connect build config, shipped artifacts, install locations,
first-launch downloads, runtime consumers, version pins, architecture-specific
paths, and smoke tests. A subsystem doc is not good enough if a smaller fresh
session still has to rediscover what must exist on disk or what gets used at
runtime.

Bootstrap must discover concepts, not just folders. It descends into substantial
child directories, clusters cross-cutting concerns, mines repo memories and
AGENTS/README files for subsystem hints, checks route/page and filename-prefix
patterns, and records known-unknowns. `kb-map` also warns when lookup sees a
thin map compared with the actual repo shape.

The point is not to load every architecture file or crawl the whole repo. The
point is to guide the model directly to the slice of project truth that matters
now, so each token pays for useful orientation instead of rediscovery.

Bootstrap also runs `kb-eval-map` to decide how the repo should be evaluated and
to write the repo's eval map. That setup is native to the app: browser workflows
for websites, command/output goldens for CLIs, contract checks for APIs,
prompt/trace/claim eval plans for skill or agent repos, and optional dashboard
export only when it helps. This maps and scaffolds safe proof; it is not the full
live skill-eval suite. If the primary workflow is unclear, `kb-eval-map` asks
what the repo is supposed to prove instead of creating fake tests.

When memory is missing, `kb-map` invokes `kb-map-bootstrap` to build the project
map once. After that, normal startup should be cheap: `kb-start` calls
`kb-map lookup <request>`, `kb-map` returns the relevant docs and likely route,
and the next skill can work without the user reteaching the app.

## Quick Use

Use these when you know the route:

| Command | Use When |
| --- | --- |
| `kb-start` | Fresh session, ambiguous ask, or "figure out the right workflow" |
| `kb-task` | First-principles task runner that chooses the KB route and continues until verified or blocked |
| `kb-map` | Setup, lookup, or refresh project memory before other work |
| `kb-eval-map` | Bootstrap-owned setup for repo-native eval surfaces and proof commands |
| `kb-fix` | Narrow bug, failing test, or small contained change |
| `kb-troubleshoot` | Broken behavior needs autonomous logs/browser/test investigation and self-correction |
| `kb-brainstorm` | Product or technical framing is still unclear |
| `kb-plan` | Requirements exist and need vertical slices |
| `kb-work` | A manifest exists and should be executed |
| `kb-review` | KB-specific code review with thermonuclear structural quality review |
| `kb-complete` | Work is done and needs review, learning, cleanup |
| `kb-memory-review` | High-cost maintenance pass for stale, contradictory, bloated, or overlapping project memory |
| `kb-ship` | Release, PR, deploy, or final readiness check |
| `kb-epic` | Large migration, rewrite, or multi-brainstorm initiative |
| `klfg` | Fully hands-off route: brainstorm -> plan -> work -> complete |

Phase boundaries:

- `kb-start` performs a startup-only session hygiene check. When exact context
  telemetry is available, it uses that signal; otherwise it falls back to
  evidence such as heavy tool output, likely compaction, task switching, or
  reliance on chat history. It recommends handoff/restart only when durable local
  memory can replace the live chat at lower total context cost or lower drift
  risk. It does not interrupt active work just because a session is long.
- `kb-brainstorm` writes/reviews requirements, resolves safe/actionable P0-P4
  findings, and invokes `kb-plan` when the brainstorm is gate-clean. It pauses
  only for unresolved blockers, required human decisions, required research, or
  an explicit user stop.
- "Don't ask many questions", "go straight to work", and similar requests reduce
  Q&A but do not skip planning. The route is requirements/assumptions ->
  `kb-plan` -> `kb-work` -> `kb-complete`; execution intent carries forward
  after the manifest is written.
- `kb-brainstorm` multiple-choice questions must always include an escape hatch
  such as `Other / let me explain` or `None of these`. If the answer may need an
  image, screenshot, file, pasted output, diagram, or longer explanation, the
  skill should ask in normal chat instead of the blocking question UI.
- `kb-plan` writes the manifest and slice plans. If the user or upstream
  orchestrator asked to execute, it immediately invokes `kb-work <manifest>`.
- `kb-work` executes all runnable slices and calls `kb-complete` only when every
  slice is done or intentionally skipped.
- `kb-work` must not run from raw brainstorm notes, phase lists, or a free-form
  feature request. If no valid manifest exists, it routes back to `kb-plan`
  first so `kb-complete` has slice scope and verification evidence. The
  manifest's `expected_files` are a forecast, not a hard allowlist; files
  discovered during implementation are allowed when required by the slice and
  recorded in the scope ledger.
- `kb-complete` records memory-maintenance signals in
  `docs/context/memory-maintenance.md`: contradictions, overlaps, stale docs,
  bloat, repeated rediscovery, durable refreshes, and closed handoffs. It stores
  pointers and the actual issue so a future deep memory pass knows what to
  inspect instead of starting from a blind full scan.
- `kb-memory-review` is the explicit high-cost pass that consumes those signals,
  reconciles/compacts/consolidates targeted memory docs, invokes narrower helper
  skills when useful, and updates the maintenance index. It is recommended by
  thresholds but does not run automatically.
- Once `kb-work` starts execution, runnable slices continue without per-slice
  confirmation. It pauses only for HITL, blocked/human-required work, destructive
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
- `kb-task`
- `kb-map`
- `kb-map-bootstrap`
- `kb-eval-map`
- `kb-compact`
- `kb-check`
- `kb-functional-test`
- `kb-regression-snapshot`
- `kb-gate`
- `kb-fix`
- `kb-troubleshoot`
- `kb-handoff`
- `kb-research`
- `kb-architecture-deepening`
- `kb-brainstorm`
- `kb-plan`
- `kb-work`
- `kb-review`
- `kb-complete`
- `kb-memory-review`
- `kb-qa`
- `kb-repair`
- `kb-first-principles`
- `kb-epic`
- `kb-ship`
- `klfg`

Direct skill dependencies carried forward from ATV/CE:

- `document-review`
- `tdd` (lazy compatibility lane; KB protected-oracle behavior lives in plan/work/check)
- `ce-review` (generalized CE review; KB completion uses `kb-review`)
- `ce-compound`
- `ce-compound-refresh`
- `learn`
- `evolve`
- `todo-create`
- `todo-triage`

Required ATV agent dependencies:

- Document review agents: `coherence-reviewer`, `feasibility-reviewer`,
  `product-lens-reviewer`, `design-lens-reviewer`, `security-lens-reviewer`,
  `scope-guardian-reviewer`, and `adversarial-document-reviewer`.
- KB/CE code review agents: `correctness-reviewer`, `testing-reviewer`,
  `thermo-nuclear-code-quality-reviewer`, `maintainability-reviewer`,
  `project-standards-reviewer`,
  `security-reviewer`, `performance-reviewer`, `api-contract-reviewer`,
  `data-migrations-reviewer`, `reliability-reviewer`,
  `adversarial-reviewer`, `cli-readiness-reviewer`,
  `previous-comments-reviewer`, `dhh-rails-reviewer`,
  `kieran-rails-reviewer`, `kieran-python-reviewer`,
  `kieran-typescript-reviewer`, `julik-frontend-races-reviewer`,
  `schema-drift-detector`, `deployment-verification-agent`,
  `agent-native-reviewer`, and `learnings-researcher`.
- Supporting specialist agents from ATV, including repo research, design,
  browser/QA, debugging, pattern, and documentation agents. The full restored
  runtime surface is `.github/agents/*.agent.md`.

There are 52 agent files in `.github/agents/`. They are grouped by load-bearing
path:

- `document-review` needs the document-review lens agents:
  `coherence-reviewer`, `feasibility-reviewer`, `product-lens-reviewer`,
  `design-lens-reviewer`, `security-lens-reviewer`, `scope-guardian-reviewer`,
  and `adversarial-document-reviewer`.
- `kb-review`, `ce-review`, and `kb-complete` need the correctness, testing,
  structural quality, standards, security, performance, API, migration,
  reliability, frontend, Rails, Python, TypeScript, schema, deployment,
  agent-native, prior-comment, and learnings reviewers listed above.
- Planning and troubleshooting lanes use the research, architecture, debugging,
  design, and pattern agents opportunistically.

Some agents are inherited ATV specialists rather than hot-path requirements.
They stay for now because removing agents already caused review dispatch
degradation; the intended deletion gate is an eval/dispatch benchmark, not a
README claim.

## Token Diet and Lazy References

Heavy inherited ATV/CE skills keep their routing and safety rules in `SKILL.md`,
but detailed phase mechanics live under `references/` and are loaded only when
that phase is actually running.

2026-05-24 token-diet pass:

- `ce-review` was reduced to 235 lines by moving full review execution and
  post-review behavior into lazy references.
- `kb-review` is the KB-specific fork of the review orchestrator. It keeps the
  review pipeline but replaces the standard maintainability persona with
  `thermo-nuclear-code-quality-reviewer`.
- `ce-compound-refresh` was reduced to 289 lines by moving detailed maintenance
  mechanics into lazy references.
- Project-generated brainstorm and research artifacts were removed from this
  portable repo; they belong in the project that created them or in the broader
  ATV starter-kit history.
- The restored ATV agent surface stayed installed because runtime testing
  showed missing agents break review dispatch.

Current cuts:

- `ce-review` keeps mode detection, reviewer selection, severity/routing, and
  quality gates in the skill body. Full scope detection, dispatch, merge/dedup,
  headless output, fixer, artifact, todo, push, and PR flows moved to lazy
  references.
- `ce-compound-refresh` keeps the maintenance model, core rules, scope
  selection, investigation flow, and phase map in the skill body. Document-set
  analysis, action boundaries, decision prompts, execution flows, reporting,
  commits, and discoverability checks moved to lazy references.

Do not move a rule out of `SKILL.md` if missing it would make the skill choose
the wrong lane, mutate files unsafely, or skip a required gate. Move details out
when they are only needed after the lane/phase is already chosen.

## Token-Minimizing Interaction

Blocking question pickers save tokens only when the answer is truly one short
choice. They waste tokens when the user needs voice dictation, paste, images,
screenshots, files, or a long correction, because the picker is a menu rather
than a text editor.

Use closed choices for simple decisions such as proceed/pause/continue. Do not
use labels like `Suggest changes` as a closed-choice option when the expected
answer is free-form feedback. Ask in normal chat instead, or include
`Other / let me explain` and return to chat before collecting the details.

This keeps the workflow voice-friendly and reduces rework: one good dictated
answer in chat is cheaper than several picker turns trying to express nuance.

## Agent Runtime Tiers

The agent files are split conceptually, even though they are all shipped today:

- **Required dispatch agents** are called by `document-review`, `kb-review`,
  `ce-review`, `kb-complete`, and related gates. Removing these causes failed
  agent dispatch or degraded review.
- **Conditional specialist agents** are loaded only when the diff/task warrants
  the lens, such as security, performance, API contracts, migrations, design,
  browser/QA, or repo research.
- **Optional direct-use agents** can be removed later only after a benchmark run
  proves no skill dispatches them and no documented workflow depends on them.

Until that benchmark exists, keep `.github/agents/*.agent.md` intact. The prior
test failure proved that deleting ATV agents blindly is worse than carrying the
small file surface.

## Benchmarking Skills

Skill quality is measured by behavior, not by line count alone.

Use repeatable prompts against a scratch repo and record:

- correct route selected by `kb-start` (`kb-fix`, `kb-troubleshoot`,
  `kb-brainstorm`, `kb-plan`, `kb-work`, `kb-epic`, or `kb-research`)
- whether `kb-map` resolves the active repo root and exact memory files without
  searching the drive
- number of avoidable user questions
- whether planned slices include `expected_files`, verification, blockers, and
  HITL flags
- whether runnable slices continue without per-slice confirmation
- whether review agents dispatch successfully
- deterministic checks run by command instead of model judgment
- token/line load at skill start versus lazy phase load
- final artifacts moved to the right lifecycle file or folder

Optimize for fewer loaded tokens only when the same prompt still routes,
executes, verifies, and records state correctly. A shorter skill that drops a
gate is not an improvement.

## Canonical Quality Gate

Run this before propagating skill changes:

```powershell
go run .\cmd\kbcheck core
```

Run this before releasing or syncing globals:

```powershell
go run .\cmd\kbcheck local-release
```

`local-release` composes deterministic local proof: native `core`, sync drift,
line-ending checks, and available static reports. `live-release` is explicit:

```powershell
go run .\cmd\kbcheck live-release
```

It attempts the live Codex/GHCP corpus only when selected and reports unavailable
live surfaces as `skipped-explicit`; a local green gate is not a claim that live
model evals ran.

Current expected result: 0 errors and all tracked skill roots in sync. Known
warnings are long hot-path skills.

For this repo, `kb-check` now discovers the cross-runtime skill quality suite:

- `scripts/skill-lint.ps1` validates `SKILL.md` frontmatter, lazy references,
  conflict markers, and hot-path size budgets.
- `scripts/route-complexity-eval.ps1` validates deterministic route-complexity
  fixtures for Codex and GitHub Copilot/GHCP applicability.
- `scripts/skill-eval.ps1` scores captured skill result JSON and self-tests
  route/proof/claim failures. Its `trace` checks are model-reported intent
  unless an observed wrapper adds `observed_trace`.
- `scripts/skill-eval-run-codex.ps1` can run route fixtures through `codex exec`
  and emits scorer-compatible result JSON; `kb-check -All` runs its dry-run.
- `scripts/skill-eval-run-ghcp.ps1` can run route fixtures through GitHub
  Copilot CLI and emits scorer-compatible result JSON; GHCP uses strict parsing
  rather than schema enforcement because the observed CLI has no
  `--output-schema` flag. `kb-check -All` runs its dry-run.
- `scripts/skill-eval-wrap.ps1` wraps an adapter with PATH shims and git-status
  diffing, then adds `observed_trace` so forbidden-command and no-write safety
  checks are externally captured instead of only self-reported.
- `scripts/skill-eval-run-live-corpus.ps1` explicitly runs selected fixtures
  across Codex and GHCP adapters and writes local summary artifacts; it is not
  part of `kb-check -All` because it may call live models.
- `scripts/skill-eval-claims.ps1` checks transcript-derived claim artifacts
  against deterministic file/command/read evidence and reports ambiguous claims
  without turning them into proof.
- `scripts/skill-eval-quality.ps1` computes deterministic output-quality scores
  from raw captured result JSON for completeness, maintainability-shape,
  relevance, proof quality, and right-sized ceremony. It is independent for
  code-checkable dimensions, not a subjective LLM style judge.
- `scripts/skill-eval-regression-report.ps1` summarizes local live-run artifacts
  and compares pass/non-pass and size/time proxies against a selected baseline.
- `cmd/kbcheck` is the Go CLI for `core`, `local-release`, and `live-release`.
  It is tested with `go test ./...` and owns top-level orchestration. The parity
  report under `docs/reports/` records the proof used to retire the old
  PowerShell wrappers.
- `scripts/skill-surface-minimality.ps1` statically classifies skills and
  reviewer agents as `protected`, `required`, `conditional`, `unproven`,
  `unused-candidate`, or `trim-candidate`. Protected skills such as `ce-review`
  and `document-review` are excluded from cold-storage deletion candidates. The
  output is a candidate list, not deletion approval.
- `scripts/atv-upstream-delta.ps1` reads original ATV upstream changes and
  classifies them as "kb-owned-reject", "shared-overlap-review",
  "atv-native-candidate", "superseded-workflow-reject", or "unknown-review".
  It has no apply mode.
- `scripts/skill-sync-report.ps1` compares working, global, and ATV skill
  targets without copying or overwriting anything.

The unusual part is the live model-in-the-loop eval path. Dry-run adapters are
part of `kb-check -All`; live mode is explicit because it shells to authenticated
CLIs:

```powershell
.\scripts\skill-eval-run-codex.ps1 -FixtureId tiny-typo-fix
.\scripts\skill-eval-run-ghcp.ps1 -FixtureId tiny-typo-fix
.\scripts\skill-eval-run-live-corpus.ps1 -All -Runtime codex,ghcp
```

Those commands generate result JSON under `.atv/eval-runs/`, then
`scripts/skill-eval.ps1` scores route choice, model-reported trace intent, proof
strings, and claim artifacts. For safety proof, wrap a run:

```powershell
.\scripts\skill-eval-wrap.ps1 -Runner scripts\skill-eval-run-ghcp.ps1 -FixtureId tiny-typo-fix -DryRun -Sealed
```

The wrapper adds `observed_trace` using PATH-shim command capture and git-status
write/delete checks. A fixture can fail for the right reason: wrong route,
missing planned proof command, externally observed forbidden command, observed
write/delete during a routing eval, bad claim evidence, or regression against a
baseline. Live mode requires installed and authenticated Codex/GHCP CLIs.

The live cross-runtime eval harness work is tracked in
`docs/plans/2026-05-30-000-kb-live-cross-runtime-skill-eval-harness-manifest.md`.
The core planned harness is implemented; the remaining growth path is expanding
the fixture corpus beyond the current route set and adding optional exporters.

The shared contract lives in `config/skill-quality.json`. Working, global, ATV
`.github`, ATV scaffold, and ATV plugin skill roots are expected to match unless
a deliberate packaging exception is recorded.

## Portable Repo Hygiene

This repo should contain skills, agents, scripts, templates, and durable
references needed by the workflow. It should not carry project-generated
brainstorms, plans, handoffs, research notes, or context maps. Those artifacts
belong in the consuming project or in the larger ATV starter kit history.

`kb-handoff` follows that boundary. It writes repo-local handoffs under the
validated active work repo so `kb-map` and `kb-start` can find them later. It
must not create project-work handoffs in this portable skill repo unless the
handoff is explicitly about maintaining the skill bundle itself.

Skill changes are propagated from this working bundle to the personal/global
installs and the ATV fork after diff review. Before overwriting a global copy,
compare it against this repo and merge any newer useful drift back here first.
Then sync the approved copy to Codex, Copilot, shared agents,
`E:\all-the-vibes\.github\skills`, and the ATV scaffold/plugin copies. Keep the
repo README and ATV README current when the visible workflow or shipped skill
surface changes.

The private reusable catalog lives at `E:\agent-marketplace` and is configured
in `config/skill-marketplace.json`. It is not a global install. New
project-specific skills should start in the consuming project's
`.github/skills/learned-*` path and stay there until `learn`/`evolve` evidence,
reuse, review, and human approval justify promotion into the private catalog.
Reusable pipelines follow the same rule: prove them first as project-local
`config/pipelines/*.json`, then promote approved copies to
`E:\agent-marketplace\pipelines`.

This is a hand-curated ATV-derived snapshot, pinned to the local ATV fork as of
2026-05-31. There is no automatic upstream merge bot yet. Pulling upstream ATV
fixes is a deliberate sync/review task: diff original upstream, keep only used
or clearly useful changes, port improvements into the active KB/CE replacements,
run `kb-check -All`, then propagate required targets.

Use the read-only upstream report before deciding what to port from original
ATV:

```powershell
.\scripts\atv-upstream-delta.ps1
```

The report classifies changes only. "kb-owned-reject" means upstream changed a
skill KB owns locally; do not apply it over the KB copy. `shared-overlap-review`
means review and manually port useful improvements. `atv-native-candidate`
means the upstream change may matter to an ATV-native skill such as
`atv-security`; security-sensitive rows call out OSV proof drift. Unknown rows
need human review.

`atv-security` is the current approved single-skill exception from ATV. It is
hash-pinned in `E:\agent-marketplace\catalog\approved-skills.json`, mirrored in
`E:\agent-marketplace\skills\atv-security`, and installed into the Codex,
Copilot, and shared agents global skill directories. Do not bulk-install ATV
skills globally; promote each skill through the marketplace boundary first.

Use the promotion script so the safe path is also the fast path:

```powershell
.\scripts\promote-marketplace-skill.ps1 `
  -Source <reviewed-skill-dir> `
  -SkillId <skill-id> `
  -ApprovalReason "<why this is approved>" `
  -InstallTargets codex,copilot,agents `
  -Approved
```

The script validates `SKILL.md`, copies to the approved marketplace, pins the
SHA256 in `approved-skills.json`, syncs selected globals, verifies hash
equality, and runs the marketplace firebreak. Direct global copying is the
manual bypass path and should not be used for new promotions.

The paired `dependency-vulnerability-osv` harness uses OSV Scanner for
dependency vulnerability proof:

```powershell
osv-scanner scan source -r <repo-or-scope-path> --format json --output-file docs/security/osv-YYYY-MM-DD.json
```

`osv-scanner` is installed locally through the official Go install path and is
available on PATH from `C:\Users\marowe\go\bin`.

Quarantine is a firebreak, not a category label. `kb-check -All` runs
`scripts/skill-marketplace-firebreak.ps1`, which fails if any active or
approved skill root resolves into `E:\agent-marketplace\quarantine`, if an
approved catalog entry points into quarantine, or if a quarantine entry is
marked approved in place. Promotion means copying a reviewed, hash-pinned skill
into `E:\agent-marketplace\skills`, never loading directly from quarantine.

The runtime boundary is:

- skills do work;
- pipelines compose skills for a domain workflow;
- harnesses prove outputs;
- project lock files record pinned imports and valid local drift.

Consuming projects should keep app-specific variants under their own
`.github/skills`, `config/pipelines`, `.atv/pipeline-runs`, and
`.agent-marketplace/skill-lock.json` paths. Marketplace promotion happens only
when the local behavior proves reusable.

## Not Bundled

These are intentionally left out of the minimal working bundle:

- upstream `deepen-*` passes; use `kb-research` and proportional research
- one-shot LFG/SLFG style workflows; use `klfg` only when you actually want the
  whole pipeline
- upstream `workflows-*` aliases; use the KB lanes directly (`kb-brainstorm`,
  `kb-plan`, `kb-work`, `kb-review`, `ce-compound`) unless a current app
  explicitly needs an ATV alias
- `land`; shipping remains a deliberate separate decision
- browser tools such as `agent-browser`; skills can call them when installed, but
  this repo does not vendor them

The useful LFG finish pattern is preserved inside `kb-complete` without
hard-coding old tool choices: resolve follow-up review/TODO work, rerun proof on
the final diff, capture demo evidence when useful, then compound, learn, evolve,
refresh memory, compact, clean up, and alert.

This keeps the intent of old `/resolve_todo_parallel`, `/test-browser`, and
`/feature-video` without forcing every repo to use the same TODO storage,
browser transport, or demo-capture tool.

Do not remove `.github/agents/` from this bundle. The agent files are not
optional docs; they are the personas that the review and planning skills dispatch
at runtime.

## Project Memory

The workflow keeps memory in files so sessions can stay short.

Required project memory files:

- `todo.md` - active work, blockers, parked work, and handoff pointers
- `todo-done.md` - compact archive of completed work
- `docs/context/PROJECT.md` - project route map for fresh sessions
- `docs/context/eval-map.md` - repo-native eval surfaces and canonical proof commands
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

`todo.md` is not a history file. Keep board rules at the top of `todo.md`, not
in `todo_rules.md`, `todo-rules.md`, or any separate rules file. When a feature,
slice group, handoff, or fix is complete, move the compact summary to
`todo-done.md` and remove the completed entry plus routine completion logs from
`todo.md`.

Board row markers are part of that inline contract: `⬜ pending`,
`🔧 in_progress`, `✅ done`, `🔒 blocked`, `⊘ skipped`, and
`🛑 human-required`. Section icons are also standardized: `💡 Feature Ideas`,
`📋 Queued Improvements`, `🧊 Parked / Cold Storage`, `🛑 Human Required`, and
`📝 Work Log`. Use `🔒 blocked` for dependency/tool/another-agent waits that can
resume when the blocker clears. Use `🧊 Parked / Cold Storage` only for work
intentionally out of bounds today; only a human promotes it back to active.

## Execution Model

The pipeline is designed around three task sizes:

- **Small known bug:** use `kb-fix`. Write or identify a failing check, make the
  smallest fix, run deterministic verification, and stop if the fix loop stalls.
- **Broken behavior with unclear cause:** use `kb-troubleshoot`. It owns the
  observe -> reproduce -> localize -> fix -> verify loop, including logs,
  Playwright/CDP/Agent Browser checks, console/network inspection, and
  self-correction until fixed or honestly blocked.
- **Bounded autonomous task:** use `kb-task`. It maps the repo, reasons from
  first principles, chooses the right KB lane, and continues through verification
  or an explicit blocker.
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

`kb-functional-test` owns the test-level decision. Slice plans should record
`test_level` (`none`, `unit`, `integration`, `functional-api`,
`functional-cli`, `functional-browser`, or `full`) and `functional_risk`
(`none`, `narrow`, `broad`, or `full`). Unit tests prove local logic; functional
tests prove the user-visible/API/CLI/browser workflow did not lie. Small/mini
models may classify test level or audit mocked-theater tests when the platform
supports model-tiered agents, but executable checks remain the proof.

For UI work, `functional-browser` is automatic when `.tsx`, `.jsx`, `.vue`, or
`.svelte` files change, or when backend/state behavior is primarily reached
through the app UI. The proof must open the running app, navigate to the actual
screen, use real clicks/inputs/visible controls, assert rendered outcomes, save
screenshots as evidence, and clean up artifacts. Backend/API/unit tests can
support that result; they cannot replace it.

`kb-qa` must turn visible acceptance criteria into executable browser assertions
or the project stack equivalent. A screenshot can support the result, but it is
not the pass/fail oracle. If the behavior cannot be asserted programmatically,
the slice is human-required rather than model-verified.

Generated commands and assertions must avoid nested-quote traps. If shell
commands, file operations, JSON, SQL, HTML, config blocks, or Playwright
selectors require quotes inside quotes or escaped escapes, write the content to
a temp file, heredoc, template literal, or parameterized locator helper instead
of constructing it inline.

`kb-regression-snapshot` records deterministic state after each passed slice in
`.atv/snapshots/<slice-id>.json`, then verifies prior snapshots before the next
slice starts. The LLM writes the compact snapshot spec; the bundled runner
verifies DOM/API/CLI/file checks mechanically so later slices have machine
memory of what already worked instead of relying on prose or the model's
recollection.

`kb-complete` fails the proof gate when a slice only has prose proof. Each slice
needs command/test path, exit code, timestamp, trace/log/API artifact, or
snapshot verification evidence recorded in the manifest.

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
$src = 'E:\working-skill-repo'
$skills = "$env:USERPROFILE\.codex\skills"
$agents = "$env:USERPROFILE\.codex\agents"
Copy-Item "$src\.github\skills\*" $skills -Recurse -Force
Copy-Item "$src\.github\agents\*" $agents -Force
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
