# KB Factory Skill Tree Requirements

Date: 2026-05-23
Status: draft

## Purpose

Design a durable skill tree that can take work from project startup through release while keeping context local, reusable, reviewable, and resumable. The system should support long-running product work without requiring multi-day chat sessions or repeated research.

The eventual target is a "dark factory" workflow: the user can queue brainstorms, turn them into plans, load the resulting slices, and let the agent work unattended until the queue is complete, blocked, or explicitly escalated. The agent should stay busy only while real work remains, move as fast as responsible verification allows, and avoid both fake completion and artificial keep-alive loops.

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
- Run unattended queues overnight with automated testing, durable blocker recording, and clear completion criteria.

## Non-Goals

- Do not create one giant mega-skill that tries to do everything.
- Do not force every task through the full brainstorm -> plan -> work pipeline.
- Do not research every ask by default.
- Do not rely on a specific model name as part of the workflow contract.
- Do not treat `todo.md`, `todo-done.md`, or handoff files as skills; they are project memory files.
- Do not keep agents busy for its own sake. The loop exists to finish queued work, not to simulate activity.

## Dark Factory Target

The workflow should support an unattended execution mode where the user can prepare work during the day and let the agent run through it overnight.

Dark factory does not mean reckless autonomy. It means the queue, instructions, verification, blockers, and stop conditions are explicit enough that the agent can work without repeatedly asking the user to steer.

Required properties:

- Queueable brainstorms and plans.
- Slice manifests with dependency order and runnable/blocked/manual status.
- Fresh-context execution per slice when useful.
- Agent-owned tests and browser verification.
- Repair loops with progress-based ceilings.
- Durable parked-slice records for human-only blockers.
- Automatic continuation to unrelated runnable slices when one slice parks.
- `kb-complete` runs after all runnable work finishes.
- Final state is unambiguous: complete, parked, blocked, or failed with evidence.

The unattended executor should stop only when:

- All runnable slices are complete and post-work gates passed.
- All remaining slices are parked/manual/blocked and recorded durably.
- Verification cannot proceed after bounded repair/escalation.
- A destructive, external, credentialed, or subjective decision is genuinely human-only.
- The user configured a time, cost, or risk budget and the run reached it.

The unattended executor should not stop when:

- A browser test could be automated.
- Inputs are missing but safe fixtures or previously captured `test_inputs` exist.
- One slice is parked but independent slices remain runnable.
- A test fails once and `kb-repair` has not run.
- The agent merely wants the user to manually QA a normal web/app behavior.

Completion contract:

```markdown
## Dark Factory Run Summary

### Completed Slices

### Parked / Manual / Blocked Slices

| Slice | Status | Reason | Resume When | Next Agent Action | Human Action |
|---|---|---|---|---|---|

### Verification Run

- Tests:
- Browser:
- Lint/typecheck:
- Review:
- Cleanup:

### Failures / Escalations

### Next Recommended Run
```

Dark factory mode should optimize for finishing correctly, not for staying on. If the queue is done, stop. If real work remains, keep moving.

## Scheduling and Swarm Policy

After the workflow can produce good slices, the harness needs to decide how to keep work moving: serial execution, limited parallel execution, or a true swarm. Parallelism is an optimization, not a default virtue.

Default stance:

- Run in series when slices share files, subsystems, migrations, schemas, generated artifacts, or architectural decisions.
- Run in parallel only when slices have clear file ownership and low integration risk.
- Use research/review agents in parallel more freely than code-editing agents because read-only work has much lower merge risk.
- Prefer slower correct execution over fast conflicting execution that creates rework.

### Swarm Readiness Checklist

Before dispatching multiple implementation agents, the orchestrator must classify each slice:

```yaml
slice_id:
  expected_files:
    - "src/playbooks/..."
  expected_subsystems:
    - playbooks
  writes_generated_files: false
  touches_schema_or_migration: false
  touches_shared_types: false
  touches_shared_test_fixtures: false
  depends_on:
    - other-slice-id
  parallel_group: null
  parallel_safety: low|medium|high
```

Parallel execution is allowed only when:

- Expected file sets do not overlap.
- Subsystems are independent or have a stable interface boundary.
- No slice changes shared schemas, migrations, package config, generated code, global styles, central routing, shared prompts, or shared test fixtures.
- No slice depends on another currently running slice.
- Each slice has its own verification command.
- The orchestrator can merge results deterministically.

Parallel execution is not allowed when:

- Multiple slices touch the same feature area, such as three slices all touching playbooks.
- File ownership is unknown or likely to converge on the same files.
- The work changes architectural direction, public contracts, data models, auth, streaming/chat protocols, or tool routing.
- Slices need shared refactors to land first.
- The repo has many unstaged changes and the orchestrator cannot distinguish user work from agent work.

### Parallel Work Modes

| Mode | Use When | Code Writes |
|---|---|---|
| Serial | Dependencies or overlapping files exist | One slice at a time |
| Read-only swarm | Research, review, repo scan, test discovery | None |
| Limited parallel | Independent slices, disjoint files, clear merge plan | 2-3 agents |
| Full swarm | Large backlog with many isolated areas and mature harness | Many agents, still grouped by ownership |

The workflow should bias toward read-only swarms and limited parallel coding before full implementation swarms.

### Conflict Prevention

Each coding agent should receive:

- Its slice plan path.
- Explicit allowed file areas.
- Files it must not edit.
- The dependency and verification contract.
- Instruction to stop and report if it needs to edit outside its allowed area.

If an agent discovers it must touch a shared file, it should stop and return:

```markdown
## Parallel Safety Escalation

Slice:
Needed shared file:
Why this file is needed:
Risk if edited in parallel:
Recommended action:
```

The orchestrator then either:

- Converts the slice to serial execution.
- Creates an enabling shared-refactor slice.
- Re-groups the parallel batch.
- Parks the slice if the conflict needs human direction.

### Merge and Integration Gate

After any parallel batch:

1. Review each agent's changed files before combining.
2. Reject changes outside declared ownership unless justified.
3. Run targeted verification for each slice.
4. Run an integration verification for the whole batch.
5. Run `kb-qa` if user-visible behavior changed.
6. Update `todo.md`, slice manifests, and any affected active handoff files.

The batch is not complete until the combined tree passes verification. Individual agent success does not imply batch success.

### Human Involvement Without Blocking The Factory

The harness should keep the user involved in high-leverage decisions without requiring constant babysitting.

Good human checkpoints:

- Before choosing between architecture directions.
- Before accepting a shortcut with meaningful revisit risk.
- Before running a high-conflict swarm.
- After a batch completes with tradeoff findings.
- When a slice parks for a true human-only blocker.

Bad human checkpoints:

- Asking the user to manually test ordinary app behavior.
- Asking the user to arbitrate conflicts that the orchestrator can detect from file ownership.
- Blocking the whole queue because one non-critical slice is parked.
- Re-asking decisions already captured in `todo.md`, active handoff files, or architecture docs.

The harness should preserve human control over direction while removing human labor from routine execution and QA.

## Durability Over Shortcuts

The workflow should bias toward durable decisions over cheap exits. The agent should not recommend a shortcut just because it is faster in the current session.

Principle:

- Simple because it is the right abstraction is good.
- Simple because it avoids the hard part is technical debt.
- Cheap now is expensive if it sends the project down the wrong architecture path.

When comparing approaches, the agent must separate:

| Question | Meaning |
|---|---|
| Time to first patch | How quickly can this be made to appear fixed? |
| Time to durable fix | How long until the issue is actually solved without likely rework? |
| Revisit risk | How likely is this to force another pass later? |
| Path dependency | Does this choice make the correct future architecture harder? |
| Reversibility | If wrong, how expensive is it to undo? |

Shortcuts are acceptable only when:

- The work is explicitly disposable or exploratory.
- The shortcut is isolated behind a seam and easy to replace.
- The user deliberately chooses the shortcut after seeing the rework risk.
- The shortcut does not lock in an architecture decision.
- Verification can still prove the actual user-facing behavior.

Shortcuts are not acceptable when:

- The agent has not researched known failure modes.
- The shortcut changes infrastructure direction.
- The shortcut is being chosen because the correct design feels large or unfamiliar.
- The shortcut creates a known future migration.
- The problem is likely to compound across many slices.

The agent should prefer "measure three times, cut once" for architecture, protocol, data model, auth, chat/streaming, persistence, tool routing, and other choices that create path dependency.

Example: choosing SSE as a middle layer for LLM chat may look simpler than WebSocket/Socket.io initially, but if bidirectional control, retries, backpressure, reconnect semantics, cancellation, or multi-client coordination matter, the shortcut can force a later rewrite. That class of decision belongs in brainstorm/research before planning hardens around it.

## Project Bootstrap And Parity

The workflow must create the same project-memory shape whether it starts in a brand-new app or is installed into an app that has been running for months. Existing apps need backfill/bootstrap to reach parity with new apps.

### New App Bootstrap

When starting a new app with this workflow, create:

```text
todo.md
todo-done.md
docs/context/
  PROJECT.md
  architecture/
    README.md
  research/
    README.md
  decisions/
    README.md
  operations/
    README.md
    testing.md
docs/brainstorms/
docs/plans/
docs/handoffs/
  active/
  done/
  parked/
```

Optional when applicable:

```text
docs/context/decisions/starter-kit-deltas.md
docs/context/epics/
docs/context/history/
```

### Existing App Bootstrap

When bringing the skills into an existing app, run an expensive `kb-map-bootstrap` pass once.

Bootstrap responsibilities:

1. Detect app type, frameworks, package managers, test commands, dev server commands, and deploy/release shape.
2. Create the standard file layout if missing.
3. Fill `docs/context/PROJECT.md` with a compact route map, not a full repo dump.
4. Create subsystem docs for major architecture areas.
5. Build `docs/context/architecture/README.md` as an index.
6. Build `docs/context/research/README.md` as an index.
7. Capture known test/run commands in `docs/context/operations/testing.md`.
8. Detect existing brainstorms, plans, TODO files, tickets, docs, ADRs, or handoffs and index them.
9. Detect stale or completed active work and move/archive it instead of leaving it in the active board.
10. If the repo is based on a starter kit, create or update `docs/context/decisions/starter-kit-deltas.md`.

Bootstrap output should answer:

- What is this app?
- How do I run it?
- How do I test it?
- What are the major subsystems?
- Where do active work, done work, plans, brainstorms, and handoffs live?
- What old work is stale, done, parked, or unsafe to resume without refresh?

### Parity Rule

After bootstrap, a fresh session should not care whether the app is new or old. It should use the same read order and file layout.

### Root Instructions Versus Skills

`AGENTS.md` is a root instruction block, not a skill. It is loaded broadly by compatible agents and should stay tiny: point agents to `kb-route`, name the memory files, and avoid duplicating workflow details.

Skills live under `.github/skills/<name>/SKILL.md`. They are loaded when invoked or selected and can contain the procedural workflow.

Use both:

- `AGENTS.md` says: "For KB work, start with `kb-route`."
- `kb-route` decides which workflow lane to use.
- Lane skills (`kb-fix`, `kb-plan`, `kb-work`, etc.) do the specialized work.

## Project Memory Files

Every project using this workflow should have these repo-root files:

- `todo.md` - live execution board for active work, current truth, parked work, blockers, handoff queue pointers, and work log.
- `todo-done.md` - archive of completed work with validation notes and links to completed plans, brainstorms, and handoffs.

Every project should also have durable app context:

```text
docs/context/
  PROJECT.md
  architecture/
    <subsystem>.md
  research/
    <topic>.md
```

Standard project memory layout:

```text
todo.md
todo-done.md
docs/context/
  PROJECT.md
  architecture/
    README.md
    <major-subsystem>.md
    <major-subsystem>/
      <child-area>.md
  research/
    README.md
    <topic>.md
  decisions/
    README.md
    <YYYY-MM-DD>-<decision>.md
  operations/
    README.md
    runbooks.md
    testing.md
docs/handoffs/
  active/
  done/
  parked/
```

Naming rules:

- Use lowercase kebab-case for context docs: `playbooks.md`, `mcp-capabilities.md`, `tool-routing.md`.
- Use `PROJECT.md` only for the top-level route map.
- Use `README.md` only as a folder index inside `docs/context/*`.
- Put major subsystem docs directly under `docs/context/architecture/`.
- Put child docs in a folder named after the parent subsystem when the parent grows too large.
- Do not invent project-specific top-level memory names unless the user explicitly asks.
- Do not collide with app source docs; this system owns `docs/context/`, `docs/handoffs/`, and repo-root `todo*.md` files.
- New active-work docs should use `todo.md`, not `todo.md`.

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

Comparative research is captured in `docs/context/research/2026-05-23-skill-system-comparative-research.md`.
Use that note when refining skill boundaries or checking whether a proposed KB pattern should borrow from ATV/KB, Matt Pocock's skills, or G-Stack.

Key research decisions:

- Keep the KB backbone: brainstorm -> plan -> work -> complete.
- Add a tiny router instead of turning every skill into a mega-skill.
- Put shared policy in route/lane contracts and lazy reference docs, not repeated in every skill.
- Borrow Matt Pocock's small-skill discipline, G-Stack's operating-system ideas, and ATV/KB's review/learning loop.
- Treat "dumb questions" as a measurable tuning problem: ask only when the answer can change scope, behavior, priority, acceptance criteria, risk, or verification.

### Routing and Memory

`kb-route`

- Default entrypoint when the user says "KB" or gives an ambiguous work request.
- Classifies work as small, medium, large, research-only, review-only, or ship.
- Reads `todo.md`, active handoff pointers, and `docs/context/PROJECT.md` before routing.
- Chooses the next skill and explains the route briefly.
- Escalates when uncertainty or repeated failure exceeds the current lane.

`kb-map`

- Creates and refreshes project memory files.
- Maintains `docs/context/PROJECT.md` and subsystem architecture docs.
- Records durable facts, rejected approaches, known failure modes, and reusable research pointers.
- Runs document review on major project-memory changes.
- Has two modes:
  - `lookup` - cheap mode for every new session; reads the pointer map and follows only relevant docs.

`kb-map-bootstrap`

- Expensive setup skill for projects without project memory.
- Scans the repo deeply and creates the standard memory structure.
- Should be lazy-loaded only when `kb-map` sees missing or badly stale project memory.
- Keeps `kb-map` small so normal session startup stays cheap.

### Small Work

`kb-fix`

- For known, bounded fixes where the agent can likely solve the issue directly.
- Requires a failing test or clear reproducible verification before editing.
- Keeps edits narrow.
- Runs targeted tests.
- Updates `todo.md` and affected active handoff files.
- Captures failed attempts and surprising discoveries when they are likely to matter later.
- Runs learning capture after meaningful fixes, not only feature work.
- Escalates to `kb-plan` or `kb-brainstorm` after repeated failure.

Escalation triggers:

- Two failed fix/verify loops.
- The suspected root cause changes twice.
- The fix touches unexpected architecture boundaries.
- The task requires product judgment.
- The agent cannot produce a meaningful verification path.

### Bug Fix Lane

`kb-fix` should support a formal bug-fix lifecycle, not just quick code edits.

Bug-fix flow:

1. **Orient**
   - Read `todo.md`, relevant active handoff files, and relevant `docs/context/` pointers.
   - Identify whether this is a known bug, regression, flaky behavior, environment issue, or missing requirement.
2. **Reproduce**
   - Prefer an automated failing test.
   - For UI bugs, use browser automation and screenshot/console evidence.
   - If reproduction requires human-only access, record the blocker precisely.
3. **Hypothesize**
   - Write down 1-3 likely causes before editing.
   - Use code search and local memory to avoid repeating failed approaches.
4. **Fix narrowly**
   - Edit the smallest responsible area.
   - Avoid opportunistic refactors unless they are required for the fix.
5. **Verify**
   - Re-run the failing test or reproduction.
   - Run targeted adjacent checks.
   - If a web app is involved, run browser verification when the behavior is visible.
6. **Record**
   - Update `todo.md` work log.
   - Create or update an active handoff file if follow-up remains.
   - Capture reusable learning through `ce-compound` / `learn` when the fix teaches something.

### Bug Testing Ceiling

Bug investigation and verification must have an end. The agent should not loop forever, and it should not silently hand vague testing back to the user.

Default ceiling:

- 3 reproduction strategies.
- 5 fix/verify iterations.
- 2 distinct root-cause hypotheses after the first confirmed reproduction.
- 1 escalation to broader research or architecture review before asking for human help.

The ceiling can be raised for high-value work, but only deliberately and with a short reason.

Progress can extend the ceiling. The point is to stop dead loops, not to stop a productive investigation.

Progress signals:

- A previously unreproduced bug becomes reproducible.
- The failure surface narrows.
- A test moves from broad failure to a more specific failure.
- One class of errors is fixed while another remains.
- Logs, stack traces, screenshots, or console output now point to a more specific subsystem.
- The root-cause hypothesis becomes more precise and better supported.
- A fix passes one verification layer but fails a later, more realistic layer.

Non-progress signals:

- Same failure after multiple unrelated edits.
- Root-cause theory changes randomly without new evidence.
- Edits are speculative and not tied to a hypothesis.
- The agent reruns the same command without learning anything new.
- Failures move around because of side effects, not because the target bug is narrowing.
- The agent cannot explain what the next attempt is expected to prove.

If progress is real, continue in bounded increments and record why the ceiling was extended:

```text
ceiling-extension: continuing because reproduction narrowed from page crash to null config in playbook step loader
```

If progress is absent, stop and escalate rather than spending more attempts.

When the ceiling is hit, the agent must produce a structured escalation report:

```markdown
## Bug Escalation Report

### Symptom

### Expected Behavior

### Confirmed / Unconfirmed Reproduction

### What I Tried

| Attempt | Hypothesis | Change/Test | Result |
|---|---|---|---|

### Evidence

- Test output:
- Screenshots:
- Logs:
- Console/network:

### Current Best Theory

### Why I Am Stuck

### Options

1. Bring in human debugging help with the evidence above.
2. Park this bug and continue unrelated runnable work.
3. Convert this into a larger `kb-brainstorm` / `kb-plan` if the bug exposes architecture ambiguity.
4. Drop or defer the task if the value no longer justifies the cost.
```

Human-in-the-loop at this point is not "please test this." It is "I exhausted the bounded investigation; here is the evidence and the decision needed."

If unrelated slices are runnable, park the bug and continue them unless the bug blocks the dependency chain.

Bug fixes should feed the same learning system as feature work:

- `ce-compound` for non-obvious root causes, gotchas, or reusable patterns.
- `learn` for recurring project instincts.
- `evolve` when repeated fixes prove a durable skill is needed.

### Tried / Failed Knowledge

Failed attempts are useful only when they prevent future wasted work.

Keep a failed attempt when:

- The failure was non-obvious.
- The same approach is tempting and likely to be retried.
- It explains why the current design is shaped a certain way.
- It identifies an environment, auth, framework, or platform limitation.
- It distinguishes this repo from a starter kit or upstream template.

Do not keep a failed attempt when:

- It was a typo, transient command failure, or simple mistake.
- It is fully explained by the final fix.
- It has no future decision value.
- It would bloat the map without changing future behavior.

Where to record:

- Subsystem doc `What We Tried And Rejected` for architecture-level failures.
- Research note `Rejected Approaches` for external/prior-art failures.
- `docs/solutions/` via `ce-compound` for reusable bug/root-cause knowledge.
- `todo.md` work log for short current-session notes.
- `todo-done.md` for completed feature/fix summaries.

Retention rule:

- Keep durable, decision-shaping failures indefinitely until the subsystem changes enough to make them stale.
- Age out noisy session observations through the existing 90-day observation decay.
- Prefer summaries and pointers over raw logs.

### Starter Kit Delta Awareness

Projects may diverge from the ATV starter kit or the user's Irtechie starter kit. The workflow should preserve meaningful deltas so future agents do not accidentally "restore" starter-kit assumptions.

Track deltas when:

- The project intentionally changed a starter-kit convention.
- A default skill, workflow, command, route, or architecture pattern was removed or renamed.
- A local workaround exists because the starter kit did not fit this project.
- A bug fix depends on understanding how this repo differs from the starter kit.

Suggested file:

```text
docs/context/decisions/starter-kit-deltas.md
```

Suggested sections:

```markdown
# Starter Kit Deltas

## Source Baseline

## Intentional Differences

| Area | Baseline | This Project | Reason | Last Verified |
|---|---|---|---|---|

## Do Not Revert

## Needs Reconciliation
```

`kb-map-bootstrap` should create this file only when it detects or is told that the repo is based on a starter kit. `kb-map refresh` should update it when meaningful deltas change.

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
- Updates `todo.md` and affected active handoff files.

`kb-work`

- Executes slices in dependency order.
- Uses scope gates, diff checks, destructive command guards, tests, QA, and repair.
- Updates `todo.md`, `todo-done.md`, and manifests as the source of resumability.

`kb-complete`

- Runs post-work review, compound knowledge capture, learn, evolve, and cleanup.
- Uses `ce-review`, `ce-compound`, `learn`, and `evolve`.

### Large Work

`kb-epic`

- Coordinates large projects that require many brainstorms, plans, and manifests.
- Maintains an epic map in `todo.md`.
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

1. Record it in `todo.md` under `Human Required`.
2. Record the exact blocked check in the relevant active handoff file when a restart pointer is needed.
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
3. Update `todo.md`, manifest notes, and affected active handoff files.
4. Ask for the human input with a minimal, structured prompt.
5. Resume that slice when the input arrives.

### Parked Slice Pattern

When a slice cannot continue because it needs human input, access, or debugging help, it becomes a parked slice. A parked slice is durable project state, not session memory.

Use a manifest status that cannot be confused with done:

- `manual` - waiting on a human action or decision.
- `parked` - intentionally paused and safe to skip while other slices run.
- `blocked` - cannot proceed because a dependency or failure blocks it.

Every parked/manual/blocked slice must be recorded in all relevant places:

1. Slice plan frontmatter.
2. KB manifest.
3. `todo.md` under `Human Required`, `Blocked`, or active feature notes.
4. `docs/handoffs/active/<handoff>.md` if a new session needs to resume it.

Required parked slice fields:

```yaml
status: manual|parked|blocked
owner: human|agent|mixed
blocked_reason: "<specific reason>"
resume_when: "<condition that makes this runnable again>"
next_agent_action: "<what the agent should do after resume>"
human_action: "<smallest requested human action, if any>"
can_continue_other_slices: true|false
parked_at: "YYYY-MM-DDTHH:MM:SS"
```

Rules:

- Do not rely on the current chat to remember a parked slice.
- Do not mark a parked slice `done`.
- Do not let a parked slice freeze unrelated runnable slices.
- Do not ask for broad manual testing; ask for the smallest missing action or decision.
- When the human action is completed, update the slice status and resume automated verification.

`kb-work` startup must scan the manifest and `todo.md` for parked/manual/blocked slices before choosing the next runnable slice.

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
- Log cleanup in the manifest notes and `todo.md` work log.

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

## Prompt Cache Policy

Prompt caching changes the token trade-off. Saving 1,000 tokens is not automatically good if the savings come from changing early prompt/context shape and reducing cache hits.

Cache-aware principles:

- Keep stable instructions, skill metadata, and shared workflow contracts as early and stable as possible.
- Put variable task context, file excerpts, research notes, and user-specific details later.
- Prefer stable pointer maps over injecting changing architecture summaries into the early prompt.
- Lazy-load large docs only when they are likely to change the decision.
- Avoid constantly changing the set or order of loaded skills if the platform cache depends on the shared prefix.
- Measure cached-token rate when the platform exposes it; otherwise approximate by repeated-run latency and input-token deltas.

Soft loading wins when:

- The loaded document is large.
- The document is rarely needed.
- Loading it changes later context but not the stable prefix.
- The cost of missing it is low because the route can follow a pointer when needed.

Soft loading loses when:

- The sub-skill is needed in most runs.
- Its absence causes wrong routing or repeated clarification.
- Loading it later changes earlier instruction/tool context in a way that reduces cache effectiveness.
- The saved tokens are smaller than the extra reasoning, routing, or retry cost.

Benchmark both modes before deciding. Do not assume smaller prompt equals cheaper workflow.

## Benchmarking and Evaluation

Benchmark this system by scenario class, not by pretending all tasks have equal complexity.

Create a small eval suite of replayable scenarios:

| Scenario | Purpose | Expected Route |
|---|---|---|
| Known one-file bug | Measures `kb-fix` speed and escalation discipline | `kb-route -> kb-fix` |
| UI bug with browser evidence | Measures agent-owned QA and repair | `kb-fix -> kb-qa -> kb-repair` |
| Clear medium feature | Measures skipping unnecessary brainstorm | `kb-route -> kb-plan -> kb-work` |
| Fuzzy product idea | Measures brainstorm question quality and intent capture | `kb-brainstorm -> kb-plan` |
| Large architecture change | Measures epic decomposition and review gates | `kb-epic -> multiple KB loops` |
| Existing repo with no memory | Measures bootstrap cost and map quality | `kb-map-bootstrap` |
| New session resume | Measures local memory usefulness | `kb-map -> route` |

Metrics:

- Input tokens.
- Output tokens.
- Cached input tokens when available.
- Number of files read.
- Number of docs loaded.
- Number of user questions asked.
- Number of dumb or avoidable questions.
- Number of human QA handoffs.
- Number of automated verification steps completed.
- Whether the route chosen was correct.
- Whether the work completed without scope drift.
- Whether future sessions can resume from local memory.

Quality scoring:

| Score | Meaning |
|---|---|
| 0 | Failed or unsafe |
| 1 | Worked only with heavy user correction |
| 2 | Completed but wasted tokens/questions |
| 3 | Completed with acceptable cost and verification |
| 4 | Completed cleanly, low waste, good memory updates |
| 5 | Completed cleanly and improved future runs |

Cache benchmark:

Run each scenario twice:

1. Cold run - no prior cache/context assumptions.
2. Warm run - same stable setup, same skill tree, similar request.

Compare:

- cached token ratio
- total input tokens
- latency
- number of extra route/clarification turns
- final quality score

Use this to decide whether a sub-skill should be:

- always present in the stable prefix,
- lazy-loaded by `kb-map` or `kb-route`,
- converted into a local doc contract,
- removed or merged.

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

`todo.md`, active handoff files, and requirements docs should preserve intent in a compact form so future agents do not turn the work into a generic implementation exercise.

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

- `docs/handoffs/active/` - current restart files and queued handoffs.
- `todo.md` - live work board and current truth.
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

1. Active handoff files under `docs/handoffs/active/`
   - Current restart point.
   - Active and deferred handoffs.
   - Human-required blockers.
   - Pointers to relevant project memory.
2. `todo.md`
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
   - Load only docs pointed to by `PROJECT.md`, active handoff files, `todo.md`, or the user's request.
5. Relevant research notes from `docs/context/research/`
   - Load only when the task depends on prior research or the note is specifically linked.

If none of the first three files exist, `kb-route` should propose running `kb-map-bootstrap` to initialize project memory before doing major work. For a truly small fix, it may proceed after a quick repo scan, but it should still recommend bootstrapping memory afterward.

### Entry Skill Behavior

The entry skill should behave like a map reader:

1. Read `todo.md`, active handoff pointers, and `docs/context/PROJECT.md`.
2. Classify the user's idea or request.
3. Identify likely subsystems from the route map.
4. Follow subsystem pointers only as needed.
5. Return the chosen route and the minimal context it loaded.

Example:

```text
User idea mentions playbooks.
Read PROJECT.md -> Subsystem Index says playbooks live at docs/context/architecture/playbooks.md.
Read playbooks.md -> Child Docs points to actions.md and mcp-capabilities.md.
If the task touches MCP capability selection, read mcp-capabilities.md.
If not, stop at playbooks.md.
```

The agent should not require the user to name `playbooks.md`. The route map should make that discoverable.

### Deep Bootstrap Mode

When project memory does not exist or is clearly stale, run `kb-map-bootstrap`.

This is intentionally token-expensive. It pays for itself by creating local memory that future sessions can read cheaply.

Bootstrap workflow:

1. Inventory the repo:
   - top-level files and folders
   - app entry points
   - package/build/test config
   - routes or screens
   - services/actions/tools
   - tests
   - docs
2. Identify major subsystems:
   - user-facing workflows
   - backend domains
   - tool/action layers
   - data/storage layers
   - integrations
   - runtime shells such as Electron or browser automation
3. Create the standard memory layout.
4. Write `docs/context/PROJECT.md` as the route map.
5. Write one architecture doc per major subsystem.
6. Add child docs only when a subsystem is too large for one concise file.
7. Create `todo.md`, `todo-done.md`, and the `docs/handoffs/active/`, `docs/handoffs/parked/`, and `docs/handoffs/done/` directories if missing.
8. Run `document-review` on `PROJECT.md` and any large subsystem docs.
9. Record bootstrap date and confidence in `PROJECT.md`.

Bootstrap output should favor pointers over exhaustive prose. It should not paste large code excerpts unless a specific snippet is essential for future routing.

Bootstrap should mark uncertain claims:

- `verified` - confirmed by reading source.
- `inferred` - likely from naming or structure, but not fully traced.
- `unknown` - needs follow-up.

### Refresh Mode

Run `kb-map refresh` after meaningful architecture changes.

Refresh should:

- Read current `PROJECT.md`.
- Inspect changed files and recent plans/manifests.
- Update only affected subsystem docs.
- Add or update child pointers when a doc is getting too large.
- Run document review when changes are substantial.

Refresh should not re-bootstrap the whole repo unless the map is missing or badly stale. If a deep pass is required, route to `kb-map-bootstrap`.

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

### Todo Board Contract

`todo.md` is the current execution board. It should preserve current truth and active work state, not every historical detail.

The top of `todo.md` should include compact rules so any agent that opens the file knows how to maintain it. If the rules grow too long, keep only a short pointer at the top and move the full rules to `docs/context/operations/todo-rules.md`.

Suggested top-of-file rules:

```markdown
# Todo

## Rules

- Keep this file current and small.
- Active, blocked, parked, and human-required work belongs here.
- Completed work moves to `todo-done.md`.
- Detailed handoffs live under `docs/handoffs/active/`, `docs/handoffs/parked/`, or `docs/handoffs/done/`; link them here instead of pasting full content.
- Before running a cold or parked item older than 72 hours, refresh it against recent code changes.
- When all active todos are done, check the handoff queue for unfinished work.
```

Required sections:

```markdown
## Objective

## Current Focus

## Current Truth

## Active Work

## Handoff Queue

| Handoff | Status | Route | Created | Stale Check | Link |
|---|---|---|---|---|---|

## Human Required

## Parked / Cold Storage

## Blocked

## Work Log
```

Completed feature sections move to `todo-done.md`.

### Done Board Contract

`todo-done.md` is the archive summary, not an infinite history dump. It should record enough to know what shipped and where details live, then age older entries into dated archive files when the file becomes too large.

Suggested sections:

```markdown
# Todo Done

## Recently Completed

| Date | Work | Verification | Links |
|---|---|---|---|

## Completed Handoffs

## Archived Ranges
```

Archive policy:

- Keep recent completed work in `todo-done.md`.
- Move older or high-volume history to `docs/context/history/todo-done-YYYY-QN.md` or similar dated archive.
- Leave a pointer in `todo-done.md`, not the full old detail.

### Handoff File Contract

Handoffs are individual files, not one giant handoff document.

Directories:

```text
docs/handoffs/
  active/
  parked/
  done/
```

Filename:

```text
YYYY-MM-DD-<short-topic>.md
```

Required handoff sections:

```markdown
# <Handoff Title>

Created: YYYY-MM-DD
Last refreshed: YYYY-MM-DD
Status: active|parked|done|superseded
Suggested route: kb-fix|kb-brainstorm|kb-plan|kb-epic|kb-research

## Intent

## Current State

## Next Agent Action

## Human Required

## Pointers

- Project map:
- Subsystem docs:
- Brainstorm:
- Plan:
- Research:
- Code:

## Staleness Check

## Completion Criteria
```

Handoff lifecycle:

1. Create handoff in `docs/handoffs/active/`.
2. Add a compact pointer to `todo.md` under `Handoff Queue`.
3. Before execution, run a staleness check if the handoff is older than 72 hours or the touched subsystem changed since creation.
4. If resumed and completed, move file to `docs/handoffs/done/` and update `todo-done.md`.
5. If still valuable but not runnable, move file to `docs/handoffs/parked/` and keep a pointer in `todo.md`.
6. If superseded, mark status `superseded`, explain why, and move to `docs/handoffs/done/`.

Handoffs should be compact. If a handoff needs more detail, put the detail in `docs/context/`, `docs/brainstorms/`, or `docs/plans/`, then link it.

### Brainstorm And Plan Lifecycle

Brainstorms and plans are durable artifacts, but they need status metadata so old work does not look active forever.

Brainstorm frontmatter:

```yaml
status: draft|ready-for-plan|planned|superseded|parked
created: YYYY-MM-DD
last_refreshed: YYYY-MM-DD
routes_to:
  - docs/plans/<plan-or-manifest>.md
```

Plan or manifest frontmatter:

```yaml
status: draft|ready-for-work|in-progress|complete|parked|superseded
created: YYYY-MM-DD
last_refreshed: YYYY-MM-DD
source_brainstorm: docs/brainstorms/<brainstorm>.md
todo_pointer: todo.md
```

Lifecycle:

1. `kb-brainstorm` creates or updates a brainstorm under `docs/brainstorms/`.
2. When it is ready for planning, set `status: ready-for-plan`.
3. `kb-plan` creates a manifest and slice plans under `docs/plans/`, then updates the brainstorm to `status: planned`.
4. `todo.md` links only active or queued manifests, not every historical plan.
5. `kb-work` updates manifest/slice status as work runs.
6. `kb-complete` moves completed work summaries to `todo-done.md` and sets completed manifests/slices to `complete`.
7. If a brainstorm or plan becomes invalid, mark it `superseded` and link the successor.

Cold artifact rule:

- A brainstorm or plan older than 72 hours must run stale-work refresh before execution unless it has been refreshed more recently.
- A completed or superseded brainstorm/plan should not appear as active work in `todo.md`.
- Keep historical artifacts in place for traceability, but route from current indexes and statuses.

### Stale Work Refresh

Before running a brainstorm, plan, handoff, or parked todo that is older than 72 hours, the agent must check what changed since it was created or last refreshed.

Refresh inputs:

- `git log --since=<created-or-refreshed-date>`
- `git diff --name-only <baseline>...HEAD` when a baseline exists
- touched subsystem docs under `docs/context/architecture/`
- related research notes and decisions
- active and done todo history

Refresh output:

```markdown
## Refresh Check

Item:
Age:
Relevant changes since creation:
Still valid:
Needs update before execution:
Route:
```

If the shape of the work changed, update the brainstorm/plan/handoff before execution. Do not run stale work blindly.

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

- active handoff files when the next session needs a restart pointer.
- `todo.md` when active state, blockers, or current truth changes.
- `docs/context/PROJECT.md` when a new subsystem, command, or important pointer appears.
- A subsystem doc when architecture, entry points, known failures, or first-read files change.
- A research note when new reusable external findings were discovered.

Do not update memory for every trivial file change. Update when future sessions would otherwise need the chat to understand what happened.

## Handoff Queue

The handoff queue is not only for the current interruption. It can also hold deferred work that is not ready for brainstorm or execution.

Recommended sections:

- `todo.md` includes a compact `Handoff Queue` table.
- `docs/handoffs/active/` holds runnable or soon-runnable handoff files.
- `docs/handoffs/parked/` holds handoffs that remain valuable but are not runnable.
- `docs/handoffs/done/` holds completed or superseded handoff files.

A handoff item should include:

- Short title
- Why it exists
- Current status
- Next recommended action
- Pointers to `PROJECT.md`, subsystem docs, research notes, plans, or code files
- Whether it should route to `kb-fix`, `kb-brainstorm`, `kb-plan`, or `kb-epic`

Completed or superseded handoffs should move to `docs/handoffs/done/` and get a compact summary row in `todo-done.md` so the active board stays small.

## Wrap-Up Comparative Review

Before turning this brainstorm into final skills, run a comparative review:

1. Compare the proposed KB workflow against the current KB skills in this repo.
2. Compare against the original KB skills in `E:\all-the-vibes\.github\skills`.
3. Compare against ATV packaged/base skills in `E:\all-the-vibes\plugins\` and scaffold templates.
4. Check the upstream/original ATV fork if available to see whether newer skill ideas should be accounted for.
5. Summarize:
   - what improved,
   - what has not improved yet,
   - what new skills were added,
   - what old behavior must be preserved,
   - what compensating controls are needed because the new design moved or removed a behavior.

This review should happen before implementation, not after, so the final skill edits do not accidentally regress useful ATV behavior.

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

- `todo.md` tracks active workstreams, blockers, parked work, and human-required steps.
- `kb-epic` groups many manifests under one initiative.
- `kb-work` only runs slices whose blockers are complete.
- Human-required tasks remain explicit and cannot be silently skipped.
- Parked work never auto-executes.
- Completion archives move to `todo-done.md`.
- `todo.md` and active handoff pointers always state where a new session should resume.

## Open Decisions

- Is `kb-route` the default entrypoint for all user asks that mention KB?
- Should `kb-fix` commit changes or leave commits to `kb-ship`?
- Should `kb-research` be standalone or mostly called from `kb-route`, `kb-brainstorm`, and `kb-plan`?
- How much of `ce-compound` should be folded into `kb-map` versus kept as a separate post-work learning step?
- Should `kb-epic` create a separate `docs/context/epics/<name>.md` file, or should epic tracking live only in `todo.md`?

## Initial Build Order

1. Create `kb-route`.
2. Create `kb-map`.
3. Create `kb-fix`.
4. Create `kb-research`.
5. Tighten existing `kb-brainstorm`, `kb-plan`, `kb-work`, `kb-complete` to use the router and memory files.
6. Create `kb-epic`.
7. Create `kb-ship`.
