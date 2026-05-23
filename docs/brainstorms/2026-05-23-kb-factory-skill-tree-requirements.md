# KB Factory Skill Tree Requirements

Date: 2026-05-23
Status: draft

## Purpose

Design a durable skill tree that can take work from project startup through release while keeping context local, reusable, reviewable, and resumable. The system should support long-running product work without requiring multi-day chat sessions or repeated research.

## Problem

Current agent workflows rely too heavily on live conversation context. When a session runs for days, tokens are spent remembering what the repo is, what was tried, what failed, what research found, and where the work currently stands. Starting a new session forces the user to reteach the app and the agent often repeats research or makes confident guesses without enough local context.

The desired outcome is a repeatable workflow where the agent can:

- Understand the project from local files.
- Route small, medium, and large tasks into the right workflow.
- Escalate when a "small fix" turns out not to be small.
- Maintain a large backlog safely.
- Execute vertical slices without scope drift.
- Run adversarial review and QA gates.
- Capture reusable research and implementation learnings.
- Resume after interruption from local memory, not chat history.

## Non-Goals

- Do not create one giant mega-skill that tries to do everything.
- Do not force every task through the full brainstorm -> plan -> work pipeline.
- Do not research every ask by default.
- Do not rely on a specific model name as part of the workflow contract.
- Do not treat `kb.md`, `kb-done.md`, or `kb-handoff.md` as skills; they are project memory files.

## Project Memory Files

Every project using this workflow should have these repo-root files:

- `kb.md` - live execution board for active KB work, current truth, parked work, blockers, and work log.
- `kb-done.md` - archive of completed KB work with validation notes.
- `kb-handoff.md` - compact restart document for new sessions.

Every project should also have durable app context:

```text
docs/context/
  PROJECT.md
  architecture/
    <subsystem>.md
  research/
    <topic>.md
```

`docs/context/PROJECT.md` is the new-session entry point. It explains what the app is, how it runs, which subsystems exist, and which context docs to read next. Subsystem docs should include:

- What it is
- Current shape
- Entry points
- What works
- What was tried and rejected
- What does not work
- Research worth keeping
- Open questions
- Files to read first

## Skill Tree

### Routing and Memory

`kb-route`

- Default entrypoint when the user says "KB" or gives an ambiguous work request.
- Classifies work as small, medium, large, research-only, review-only, or ship.
- Reads `kb-handoff.md`, `kb.md`, and `docs/context/PROJECT.md` before routing.
- Chooses the next skill and explains the route briefly.
- Escalates when uncertainty or repeated failure exceeds the current lane.

`kb-map`

- Creates and refreshes project memory files.
- Maintains `docs/context/PROJECT.md` and subsystem architecture docs.
- Records durable facts, rejected approaches, known failure modes, and reusable research pointers.
- Runs document review on major project-memory changes.

### Small Work

`kb-fix`

- For known, bounded fixes where the agent can likely solve the issue directly.
- Requires a failing test or clear reproducible verification before editing.
- Keeps edits narrow.
- Runs targeted tests.
- Updates `kb.md` and `kb-handoff.md`.
- Escalates to `kb-plan` or `kb-brainstorm` after repeated failure.

Escalation triggers:

- Two failed fix/verify loops.
- The suspected root cause changes twice.
- The fix touches unexpected architecture boundaries.
- The task requires product judgment.
- The agent cannot produce a meaningful verification path.

### Medium Work

`kb-brainstorm`

- For fuzzy requirements, product choices, or work where prior art can change framing.
- Produces requirements docs.
- Should not always perform heavy external research; research depth must be routed.

`kb-research`

- Captures reusable research in `docs/context/research/`.
- Runs only when evidence is stale, missing, high-stakes, or likely to affect decisions.
- Prevents repeating the same web/API/library research across sessions.

`kb-plan`

- Converts requirements or clear feature descriptions into vertical slices.
- Produces `docs/plans/*-kb-*-manifest.md` and slice plans.
- Updates `kb.md` and `kb-handoff.md`.

`kb-work`

- Executes slices in dependency order.
- Uses scope gates, diff checks, destructive command guards, tests, QA, and repair.
- Updates `kb.md`, `kb-done.md`, and manifests as the source of resumability.

`kb-complete`

- Runs post-work review, compound knowledge capture, learn, evolve, and cleanup.
- Uses `ce-review`, `ce-compound`, `learn`, and `evolve`.

### Large Work

`kb-epic`

- Coordinates large projects that require many brainstorms, plans, and manifests.
- Maintains an epic map in `kb.md`.
- Splits a major initiative into bounded KB workstreams.
- Ensures each workstream can run through the normal medium workflow.
- Prevents one giant plan from becoming impossible to execute or review.

Examples:

- Rewriting a web app into Electron.
- Major architecture migration.
- Multi-subsystem product launch.
- Large enterprise connector or automation platform.

### Shipping

`kb-ship`

- Deliberate final release step after `kb-complete`.
- Prepares commit, push, PR, release notes, or deployment checklist.
- Should not be bundled into `kb-complete`; shipping is a human-owned decision.

## Task Levels

### Small

Examples:

- Fix a known bug.
- Add a focused unit test.
- Repair a broken selector.
- Update a small config or typo with verification.

Path:

```text
kb-route -> kb-fix -> tests -> update memory
```

Escalate if the fix loops or expands.

### Medium

Examples:

- Add functionality to one or more pages.
- Introduce a bounded backend capability.
- Improve a known workflow.

Path when requirements are clear:

```text
kb-route -> kb-plan -> kb-work -> kb-complete -> kb-ship
```

Path when requirements are fuzzy:

```text
kb-route -> kb-brainstorm -> kb-plan -> kb-work -> kb-complete -> kb-ship
```

### Large

Examples:

- App rewrite.
- Multi-week migration.
- New product surface with several subsystems.

Path:

```text
kb-route -> kb-epic -> multiple kb-brainstorm/kb-plan/kb-work/kb-complete loops -> kb-ship
```

## Research Rules

Research should run when:

- The topic is unfamiliar.
- External API, library, platform, security, privacy, or compliance behavior matters.
- Existing local research is stale.
- Prior art can materially change the product or implementation direction.
- A small fix failed twice and the agent may be guessing.

Research should not run when:

- The repo has a strong local pattern.
- The task is mechanical or local.
- A targeted test can answer the question faster.
- The user explicitly wants a quick local fix and risk is low.

Research output belongs in `docs/context/research/<topic>.md` with:

- Checked date
- Sources
- Findings
- Applies when
- Stale when
- Rejected approaches
- Impact on current decision

## Review and Tire-Kicking

The system must keep adversarial review as a first-class gate:

- `document-review` reviews requirements, plans, project maps, and major architecture docs.
- `ce-review` reviews code changes after work.
- `kb-qa` verifies rendered behavior and lint.
- `kb-repair` fixes QA/lint failures with bounded retries.

Review should be confidence-gated and scoped. It should not spawn an unbounded committee for small local fixes, but large plans and architecture docs should get multi-perspective pressure testing.

## QA Ownership Policy

The agent owns verification. The user should not become the default QA tester after the agent finishes coding.

After `kb-plan`, every slice must include a verification strategy that the agent can run or a clearly named reason it cannot be automated. `kb-work` is not complete until those checks have run, passed, been repaired, or been explicitly blocked by a human-only dependency.

Required verification layers:

- Unit or integration tests for logic and data behavior.
- Lint/type checks for touched files when available.
- Browser verification for web UI changes.
- Console/network checks for frontend flows.
- Screenshot evidence for visual or interaction changes.
- Diff-scope verification to ensure the slice changed only intended files.
- Post-fix re-run of the failed check after any repair.

The final response must not say "you should test this" unless the workflow hit a human-only boundary. It should say what was tested, what passed, what could not be tested, and why.

Valid human-only QA blockers:

- MFA, SSO, or login that only the user can complete.
- External account approval or permissions.
- Email/SMS/push receipt on a real user-owned device.
- Payment or purchase flow requiring user-owned credentials.
- Physical hardware, camera, microphone, location, or OS permission that cannot be automated in the current environment.
- Corporate/internal site where no authenticated CDP session is available.
- User judgment on taste, copy, or business correctness after the agent verifies mechanics.

Invalid human QA blockers:

- "Run this and see if it works" when the agent can run it.
- "Try the playbook manually" when the agent can run the playbook with provided inputs.
- "Click through the page" when browser automation can do it.
- "Verify the UI" when screenshots, DOM checks, console checks, and interaction checks are available.
- "Test later" as a reason to stop the whole backlog.

When a human-only blocker exists, the agent must:

1. Record it in `kb.md` under `Human Required`.
2. Record the exact blocked check in `kb-handoff.md`.
3. Explain what the agent already verified before the blocker.
4. Provide the smallest possible human action needed.
5. Resume automated verification after the human step is complete.

## HITL Scheduling Policy

Human-in-the-loop must not stop a 50-slice backlog unless the blocked decision is truly on the critical path.

Classify HITL items:

| HITL Type | Meaning | Scheduling |
|---|---|---|
| `critical-path` | Later slices depend on this decision/access/input | Stop only the dependent path |
| `parallel-blocker` | This slice is blocked, but unrelated slices can continue | Park this slice and continue runnable slices |
| `final-validation` | Human judgment is useful before release, but not needed for development | Defer to `kb-ship` or final review |
| `agent-runnable-with-inputs` | Human only needs to provide values; agent can run the check | Ask for inputs, then agent runs it |

`kb-work` should continue executing any slice whose blockers are satisfied. A blocked slice should not freeze unrelated work.

When a slice hits HITL:

1. Determine whether the human step blocks only this slice or the whole dependency chain.
2. If unrelated slices are runnable, mark this slice `blocked` or `manual` and continue.
3. Update `kb.md`, manifest notes, and `kb-handoff.md`.
4. Ask for the human input with a minimal, structured prompt.
5. Resume that slice when the input arrives.

## Test Input Capture Policy

If verification needs realistic input values, capture them before execution whenever possible.

During `kb-brainstorm`, ask for test inputs only when they affect acceptance criteria, demo paths, or external-system verification. During `kb-plan`, each slice should declare any needed test inputs.

Examples:

- Playbook name and sample parameter values.
- Test account, role, or tenant.
- Safe customer/deal/project record to use.
- Feature flag or environment toggle.
- Expected output for a known input.
- External service sandbox identifiers.

Slice plans should include:

```yaml
test_inputs:
  - name: "<input name>"
    source: user|fixture|env|generated
    required_for: "<acceptance criterion or QA step>"
    value: "<literal value, fixture path, env var name, or TODO-human>"
```

If an input is missing but the slice is otherwise runnable:

- Ask for the specific missing value.
- If the missing value is not critical-path, park the check and continue unrelated work.
- If a safe generated or fixture value is acceptable, create it instead of asking.
- Never convert missing inputs into "user should manually test this" unless the agent truly cannot run the check after receiving inputs.

Bad final handoff:

```text
Please test the app and let me know if it works.
```

Good final handoff:

```text
Automated checks passed: unit tests, lint, browser smoke on /playbooks, console clean.
Human required: complete Microsoft SSO in the already-open browser so I can verify the authenticated Teams embed flow.
```

## QA Artifact Cleanup

Testing artifacts should not pile up indefinitely.

Rules:

- Store temporary QA screenshots under `.atv/qa-screenshots/<feature-or-run-id>/`.
- Store other temporary browser artifacts under `.atv/qa-artifacts/<feature-or-run-id>/` or an equivalent ignored scratch directory.
- Keep failure evidence while a slice is failing or under review.
- After `kb-complete` marks the manifest reviewed, prune passing screenshots and temporary artifacts for that feature.
- Preserve artifacts that are linked in a PR, bug report, or durable learning doc.
- Log cleanup in the manifest notes and `kb.md` work log.

The cleanup step should be automatic. The user should not have to manually delete 60 screenshots after a successful run.

## Research Placement Policy

Research belongs where it helps the decision.

Idea creation and product direction often need research before planning. Implementation planning sometimes needs research, but only when local patterns are insufficient or the decision is high risk.

Research lanes:

- **Idea research** - runs during `kb-brainstorm` or `kb-research`; helps decide what to build and what not to build.
- **Plan research** - runs during `kb-plan`; helps choose slice boundaries, implementation patterns, and risk controls.
- **Failure research** - runs after repeated `kb-fix` or `kb-repair` failures; helps escape confident wrong guesses.
- **Refresh research** - updates stale `docs/context/research/` notes when their stale condition is met.

Do not repeat idea research during planning if the brainstorm already captured reusable findings. Planning should read the research note and only deepen the parts that affect implementation.

## Model and Agent Policy

The workflow should describe task complexity and evidence needs, not hard-code model names.

Guidance:

- Use stronger reasoning for routing, architecture decisions, plans, and review synthesis.
- Use smaller or cheaper workers for bounded read-only research, file scans, and independent reviewer passes when the platform supports it.
- Do not require model switching for correctness.
- If the platform cannot spawn subagents, run the same gates sequentially.

## Token Budget Policy

The workflow must support changing token economics. Sometimes the user is token-rich and wants deeper review. Sometimes the user is using a constrained or expensive environment and needs lean execution. The skill tree should make this explicit instead of assuming one permanent budget mode.

Core mindset: every token must earn its keep. A token is worth spending when it reduces risk, improves judgment, prevents rework, preserves reusable knowledge, or lowers user cognitive load. A token is waste when it repeats local memory, asks performative questions, produces oversized docs, or explores paths that cannot change the decision.

Budget modes:

| Mode | Use When | Behavior |
|---|---|---|
| `lean` | Paid or constrained usage, routine work | Read only entrypoint memory, follow pointers on demand, avoid parallel review unless risk is high |
| `standard` | Default daily development | Use local memory, targeted research, scoped review, and normal QA |
| `deep` | High-risk architecture, product strategy, expensive mistakes | Run fuller research, adversarial document review, stronger code review, and broader QA |

Spend tokens on:

- Better route classification before work starts.
- Parallel adversarial review when the task is high risk or architecture-shaping.
- Research when it can change product or implementation direction.
- Clear handholding and explicit next-step options.
- Verification summaries that help the user trust the result.

Do not spend tokens on:

- Re-reading the same context every session when it can live in local memory files.
- Repeating old research when a current research note already answers the question.
- Running full ceremony for small local fixes.
- Loading every subsystem doc when the task touches one component.
- Generating giant documents no one will maintain.

Principle: use tokens for judgment and review, not for carrying memory that should be local.

## Intent Continuity

The user's intent must carry across the whole workflow: brainstorm, research, planning, execution, review, shipping, and future restart sessions.

Every major artifact should preserve:

- What the user is trying to accomplish.
- Who the work is for.
- Why it matters.
- What constraints or preferences the user stated.
- What trade-offs the user chose.
- What the system should avoid doing.

This matters when work is handed to another session, another agent, or someone the user is mentoring. The agent should not reinterpret the project from scratch every time.

Artifacts should distinguish:

- **User intent** - authoritative when it describes goals, preferences, constraints, lived context, or desired outcome.
- **Factual claims** - must be verified when important.
- **Agent recommendations** - must include reasoning and can be challenged.
- **Open ambiguity** - must be resolved or explicitly carried forward.

`kb-handoff.md`, `kb.md`, and requirements docs should preserve intent in a compact form so future agents do not turn the work into a generic implementation exercise.

## Question Quality Policy

Questions must be ambiguity-driven, not quota-driven.

Do not ask questions just because a workflow phase says to "ask questions." Ask only when the answer can change one of:

- Scope
- User-facing behavior
- Priority
- Risk tolerance
- Technical direction
- Acceptance criteria
- Human-required dependency
- Whether to proceed, pause, or escalate

Bad questions:

- Ask for information already present in local memory, code, requirements, or prior user statements.
- Ask the user to decide implementation details the repo can answer.
- Ask broad preference questions when the trade-off is not explained.
- Ask a fixed number of questions to satisfy ceremony.
- Ask questions whose answers cannot affect the next artifact.

Good questions:

- Name the specific ambiguity.
- Explain why it matters.
- Offer a recommended default when possible.
- Prefer one decision at a time.
- Let the user skip when the default is acceptable.

Question template:

```text
I need one decision because <ambiguity> affects <artifact/behavior>.
Recommendation: <default> because <reason>.
Choose:
1. <option>
2. <option>
3. <defer / use default>
```

If the model is confused, it should ask. If the model is merely following a question quota, it should stop.

## Hierarchical Local Memory

Project memory should work like a pointer tree. A new session reads a small entrypoint, then follows only the pointers relevant to the task.

Top-level files:

- `kb-handoff.md` - current restart point and queued handoffs.
- `kb.md` - live work board and current truth.
- `docs/context/PROJECT.md` - project map and pointer index.

`PROJECT.md` should not become a huge architecture book. It should contain:

- What the app is.
- How to run it.
- Current high-level architecture.
- Subsystem index with links.
- Current active work pointers.
- Known sharp edges.
- Where to look first for common task types.

Subsystem docs should be short pointer hubs. If a subsystem doc gets large, split it:

```text
docs/context/architecture/playbooks.md
docs/context/architecture/playbooks/actions.md
docs/context/architecture/playbooks/steps.md
docs/context/architecture/playbooks/tool-routing.md
```

Each file should answer one layer of the architecture. Parent docs summarize and point to children. Child docs can point to exact files, tests, routes, schemas, or examples.

Rule: load the smallest document that can answer the current question. Follow child pointers only when the current task needs that layer.

## Route Path Documentation Protocol

`kb-route` must help a new session understand the project without requiring the user to re-explain it or point directly to files. It does this through a deterministic documentation loading path.

### Required Startup Read Order

When a session begins work in a repo, read these in order if they exist:

1. `kb-handoff.md`
   - Current restart point.
   - Active and deferred handoffs.
   - Human-required blockers.
   - Pointers to relevant project memory.
2. `kb.md`
   - Current board.
   - Current truth.
   - Active workstreams.
   - Parked work and blockers.
3. `docs/context/PROJECT.md`
   - Project overview.
   - Subsystem index.
   - Run/test commands.
   - Pointer map for where to go next.
4. Relevant subsystem docs from `docs/context/architecture/`
   - Load only docs pointed to by `PROJECT.md`, `kb-handoff.md`, `kb.md`, or the user's request.
5. Relevant research notes from `docs/context/research/`
   - Load only when the task depends on prior research or the note is specifically linked.

If none of the first three files exist, `kb-route` should propose running `kb-map` to initialize project memory before doing major work. For a truly small fix, it may proceed after a quick repo scan, but it should still recommend creating memory afterward.

### Stop Reading Rule

Stop loading context once the agent can answer these questions with evidence:

- What app/repo is this?
- What is the user trying to accomplish?
- What is currently active or blocked?
- Which subsystem is relevant?
- Which files or commands are likely involved?
- What should not be repeated because it was already tried or researched?
- Which workflow lane should handle this request?

If the agent cannot answer one of those questions, follow the next pointer. If there is no pointer, do a targeted repo search and update the memory files after discovering the answer.

### `docs/context/PROJECT.md` Contract

`PROJECT.md` is the route map. It should be short enough to read at startup.

Required sections:

```markdown
# Project Map

## What This Is

## How To Run

## How To Test

## Current Architecture

## Subsystem Index

| Area | Read This | Use When |
|---|---|---|
| Playbooks | docs/context/architecture/playbooks.md | Work touches playbook creation, execution, import/export |

## Current Work Pointers

## Known Sharp Edges

## Research Index

| Topic | Note | Stale When |
|---|---|---|

## Do Not Repeat

## Maintenance Notes
```

`PROJECT.md` should not include every detail. It should point to detail.

### Subsystem Doc Contract

Subsystem docs are second-level route maps.

Required sections:

```markdown
# <Subsystem>

## What It Is

## Current Shape

## Entry Points

## Files To Read First

## Common Tasks

| Task | Read | Commands / Checks |
|---|---|---|

## What Works

## What We Tried And Rejected

## What Does Not Work

## Research Worth Keeping

## Child Docs

## Open Questions
```

When a section grows too large, split it into a child doc and replace the section with a pointer.

### Handoff Contract

`kb-handoff.md` should be the fastest restart path.

Required sections:

```markdown
# KB Handoff

Last updated: <timestamp>

## Current Session Resume

## Active Handoffs

### <handoff title>

- Intent:
- Status:
- Next action:
- Suggested route:
- Pointers:
- Human required:

## Deferred Handoffs

## Context Pointers

## Recently Completed
```

Handoffs should be compact. If a handoff needs more detail, put the detail in `docs/context/`, `docs/brainstorms/`, or `docs/plans/`, then link it.

### Board Contract

`kb.md` is the current execution board. It should preserve current truth and work state, not every historical detail.

Required sections:

```markdown
# KB Board

## Purpose

## Objective

## Current Focus

## Current Truth

## Active Features

## Human Required

## Parked / Cold Storage

## Blocked

## Work Log
```

Completed feature sections move to `kb-done.md`.

### Research Note Contract

Research notes exist to avoid repeated web/API/library research.

Required sections:

```markdown
# <Research Topic>

Checked: <date>
Budget mode: lean|standard|deep

## Question

## Findings

## Sources

## Applies When

## Stale When

## Rejected Approaches

## Impact On Current Project
```

Before doing external research, `kb-route`, `kb-brainstorm`, `kb-plan`, or `kb-fix` should check whether a relevant research note exists and whether it is stale.

### Memory Update Rules

After any workflow changes the project understanding, update memory:

- `kb-handoff.md` when the next session needs a restart pointer.
- `kb.md` when active state, blockers, or current truth changes.
- `docs/context/PROJECT.md` when a new subsystem, command, or important pointer appears.
- A subsystem doc when architecture, entry points, known failures, or first-read files change.
- A research note when new reusable external findings were discovered.

Do not update memory for every trivial file change. Update when future sessions would otherwise need the chat to understand what happened.

## Handoff Queue

`kb-handoff.md` is not only for the current interruption. It can also hold deferred work that is not ready for brainstorm or execution.

Recommended sections:

- Current Session Resume
- Active Handoffs
- Deferred Handoffs
- Human Required
- Context Pointers

A handoff item should include:

- Short title
- Why it exists
- Current status
- Next recommended action
- Pointers to `PROJECT.md`, subsystem docs, research notes, plans, or code files
- Whether it should route to `kb-fix`, `kb-brainstorm`, `kb-plan`, or `kb-epic`

Completed or superseded handoffs should move to `kb-done.md` or a dated archive section so the restart file stays small.

## Post-Work Memory Update

After `kb-work` and `kb-complete`, the workflow should update local memory when work changes how the app is understood.

Update `docs/context/` when:

- A subsystem boundary changes.
- A new important file, route, tool, action, or data flow is introduced.
- A previous "what does not work" claim is disproven.
- A new rejected approach is worth preserving.
- Research produced reusable findings.
- A future session would otherwise need the completed chat to understand the change.

Do not update architecture docs for every tiny code edit. Only update memory when it changes future orientation.

## Backlog Execution

The long-term goal is a "factory" mode where the user can generate many brainstorms, plans, and slices, then let the system execute safely.

Requirements:

- `kb.md` tracks active workstreams, blockers, parked work, and human-required steps.
- `kb-epic` groups many manifests under one initiative.
- `kb-work` only runs slices whose blockers are complete.
- Human-required tasks remain explicit and cannot be silently skipped.
- Parked work never auto-executes.
- Completion archives move to `kb-done.md`.
- `kb-handoff.md` always states where a new session should resume.

## Open Decisions

- Is `kb-route` the default entrypoint for all user asks that mention KB?
- Should `kb-fix` commit changes or leave commits to `kb-ship`?
- Should `kb-research` be standalone or mostly called from `kb-route`, `kb-brainstorm`, and `kb-plan`?
- How much of `ce-compound` should be folded into `kb-map` versus kept as a separate post-work learning step?
- Should `kb-epic` create a separate `docs/context/epics/<name>.md` file, or should epic tracking live only in `kb.md`?

## Initial Build Order

1. Create `kb-route`.
2. Create `kb-map`.
3. Create `kb-fix`.
4. Create `kb-research`.
5. Tighten existing `kb-brainstorm`, `kb-plan`, `kb-work`, `kb-complete` to use the router and memory files.
6. Create `kb-epic`.
7. Create `kb-ship`.
