# Skill System Comparative Research

Date: 2026-05-23
Status: draft

## Research Question

How should the KB workflow borrow from ATV/KB skills, Matt Pocock's skills, and G-Stack without turning into a token-heavy mega-skill system?

Sources:

- ATV local skills: `E:\all-the-vibes\.github\skills` at `b3e523add14458eb97f4c5a915502a7967b38e66`
- Working KB skill repo: [Irtechie/working-skill-repo](https://github.com/Irtechie/working-skill-repo) at `cff93a34491becb28510451c8a37400073b5a297` before this research note
- Matt Pocock skills: [mattpocock/skills](https://github.com/mattpocock/skills) at `b8be62ffacb0118fa3eaa29a0923c87c8c11985c`
- G-Stack: [garrytan/gstack](https://github.com/garrytan/gstack) at `61c9a20bd2e3a579c3d6184ed2fc95b51a528f7c`
- Kevin Copilot: [shyamsridhar123/kevin-copilot](https://github.com/shyamsridhar123/kevin-copilot) at `39eae5248819a6ae2bdf5ea850f460570d26e49d`

## Executive Takeaway

The KB system should be a small operating model, not one huge skill.

Matt Pocock's strongest pattern is small, composable skills with sharp engineering doctrine: domain glossary, ADR awareness, TDD, diagnosis, vertical slices, and issue triage. G-Stack's strongest pattern is a full engineering operating system: skill routing, persistent context, browser QA, plan review roles, release flow, benchmarks, and operational learning. ATV/KB's strongest pattern is the vertical-slice pipeline with review, compound learning, and evolution.

The right KB design is:

- A tiny `kb-route` entry skill.
- A durable local memory contract.
- A few workflow lanes: fix, research, brainstorm, plan, work, complete, ship.
- Shared policy referenced by path, not pasted into every skill.
- Escalation rules that move work up or sideways without blocking unrelated slices.
- A durability bias: choose the path that avoids predictable rework, not the path that merely minimizes the current turn.
- A dark-factory target: queueable slices, unattended execution, agent-owned verification, durable parked blockers, and clear stop conditions.
- A scheduler policy: swarm only when file ownership, dependency shape, and merge risk make parallel code edits safe.
- A bootstrap/parity contract: new apps and long-running existing apps must converge on the same memory layout before unattended execution.
- A compactness contract: every token must pay rent, but exact paths, commands, errors, requirements, blockers, and safety warnings are protected.

## What Each System Does Right

### ATV / KB

Strengths:

- Strong phase separation: brainstorm answers what, plan answers how, work executes, complete reviews and compounds.
- Vertical-slice planning is already the right core abstraction for long backlogs.
- `kb-complete` correctly treats review, learning, and cleanup as part of done, not optional polish.
- `ce-review`, `document-review`, `ce-compound`, `learn`, and `evolve` form a learning loop instead of one-off coding.
- `kb-first-principles` already contains the anti-sycophancy logic the user wants: classify pushback, accept user-owned context, verify factual claims, defend recommendations only when evidence supports them, and avoid wholesale reversal.

Risks:

- The current KB skills are accumulating shared policies in multiple places.
- `kb-brainstorm` may over-research small asks unless research is explicitly proportional to decision value.
- Some QA ownership and HITL policies belong in a shared contract consumed by `kb-work`, `kb-qa`, and `kb-complete`, not repeated everywhere.

Design implication:

- Keep ATV/KB as the backbone, but extract shared policy into slim reference docs or a small `kb-protocol` section that orchestrators load only when needed.

### Matt Pocock Skills

Strengths:

- Small skills with clear jobs. They are easy to adapt and do not try to own the whole process.
- Setup distinguishes hard dependencies from soft dependencies. Hard dependency skills require repo config; soft dependency skills degrade gracefully.
- `grill-with-docs` improves alignment by updating domain vocabulary and ADRs as decisions crystallize.
- `diagnose` centers debugging on an agent-runnable feedback loop before hypotheses or fixes.
- `tdd` insists on one vertical test-and-implementation loop at a time.
- `to-issues` converts plans into independently grabbable vertical slices.
- `handoff` avoids duplicating already-captured artifacts and references existing docs instead.

Risks:

- `grill-me` style interrogation can become annoying if questions are not ambiguity-driven.
- The issue-tracker orientation may not fit a file-backed KB board unless adapted.
- Some skills expect the user to approve planning/test scope more than this KB workflow should require for solo/AFK execution.

Design implication:

- Borrow the doctrine and smallness, not necessarily the GitHub issue flow.
- Add a KB setup/bootstrap skill that creates local context docs and board files, but distinguish hard dependencies from nice-to-have context.
- Keep question rules strict: ask only if the answer changes scope, behavior, priority, acceptance criteria, risk, or ability to verify.

### G-Stack

Strengths:

- Treats skills as an engineering operating system: product review, engineering review, design review, QA, ship, deploy, context save/restore, learning, health, benchmark.
- Shared preamble pattern gives skills consistent routing, context recovery, question format, operational learning, and safety behavior.
- Uses a persistent browser daemon for fast, stateful QA, reducing the "human please test this" failure mode.
- Has explicit `context-save` and `context-restore` flows for cross-session continuity.
- Has `benchmark-models` and E2E skill tests, which is useful for measuring skill quality instead of arguing vibes.
- `plan-tune` is a useful answer to dumb repeated questions: tune question sensitivity instead of deleting all questions.
- `autoplan` shows how to auto-decide mechanical review questions while surfacing taste/user-challenge decisions at a final gate.
- Its completeness framing separates complete lake-sized work from ocean-sized overreach.

Risks:

- G-Stack is intentionally large. Copying its preamble and operating model directly would defeat the KB token-budget goal.
- Generated shared preambles are powerful, but too much universal startup content can hurt prompt cache stability and make every skill expensive.
- G-Stack's completeness bias can conflict with the user's need for small, cheap quick-fix lanes.
- Heavy automation, telemetry, sync, browser state, and cross-model review are valuable only when they solve a real workflow bottleneck.

Design implication:

- Borrow the system ideas, not the bulk.
- Keep G-Stack-style context save/restore, QA ownership, question tuning, and benchmarks, but implement them as KB contracts and optional lanes.
- Do not put every policy in every skill. Put route-critical policy in `kb-route`; put lane-specific policy in the lane skill; put detailed references behind lazy docs.
- Keep G-Stack's completeness instinct, but bound it with KB's task-sizing lanes so "do it right" does not become "rewrite everything."

### Kevin Copilot

Strengths:

- It targets GitHub Copilot surfaces directly: `AGENTS.md`, `.github/copilot-instructions.md`, VS Code agents, and VS Code prompts.
- It treats brevity as an installable instruction layer instead of relying on "be concise" as a vague preference.
- It has explicit modes with word targets and an eval suite that compares against a generic terse baseline.
- Its rules protect correctness: exact commands, paths, errors, and warnings survive compression.
- Its conflict behavior is practical: merge with sentinel blocks instead of overwriting existing instructions.

Risks:

- The persona and token-savings footer are wrong for KB; they add noise and could fight the user's desired assistant tone.
- Output-token savings do not automatically reduce reasoning tokens or tool-call cost.
- Always-on voice rules can conflict with high-stakes explanations if "brevity" outranks clarity.

Design implication:

- Borrow the surfaces and measurement mindset, not the persona.
- Add `.github/copilot-instructions.md` alongside `AGENTS.md` so GHCP has a repo-wide instruction path.
- Add `kb-compact` for durable memory and skill-text compression.
- Make "every token must pay rent" a shared rule, with protected atoms that cannot be compacted away.

## Shortcut Avoidance Finding

LLMs often optimize for the easiest path they can justify in the current turn: "the complete fix takes a week; the simple fix takes 20 minutes." That is not engineering judgment unless it accounts for rework, path dependency, and whether the shortcut blocks the correct future architecture.

The KB workflow should require every approach comparison to distinguish:

- Fastest visible patch.
- Durable fix.
- Expected revisit risk.
- Architecture path dependency.
- Reversibility.

This also reinforces research timing. If research happens only during planning, the plan may already be anchored to an easy path. Research that can change architecture direction belongs in brainstorm or a dedicated `kb-research` pass before plan hardens the slices.

Rule of thumb:

- Use quick fixes for isolated bugs with low path dependency.
- Use deeper research/brainstorm for protocols, architecture, data models, auth, streaming, persistence, and tool-routing decisions.
- A shortcut that creates a known later migration is not cheap; it is deferred cost.

## Anti-Sycophancy / First Principles Finding

The current `kb-first-principles` and `kb-brainstorm` direction is correct. The missing piece is not "more contrarian." The missing piece is concise classification and proportional response.

Needed behavior:

- If the user corrects intent/context, accept that exact correction.
- If the user makes a factual claim, verify if it matters.
- If the user challenges a recommendation, concede only the premise that changed.
- If the issue is preference/taste, name the tradeoff and let the user decide.
- Never pendulum-swing from one extreme to the other.
- Never argue for sport.

Skill text should prefer this response shape:

```text
I still think X because Y.
You're right about Z, which changes A.
So I would revise to B, not the opposite extreme.
```

## Kitchen Sink Verdict

Yes, the current direction can defeat token management if every skill embeds every rule.

The fix is not fewer rules. The fix is better placement:

| Policy | Belongs In |
|---|---|
| Task routing and lane choice | `kb-route` |
| Local memory read order | `kb-route` and `kb-map` |
| Deep repo indexing | `kb-map-bootstrap` only |
| Brainstorm question quality and pushback | `kb-brainstorm` plus shared first-principles reference |
| Research reuse | `kb-research` and research docs |
| Vertical slicing and HITL flags | `kb-plan` |
| Slice execution and escalation | `kb-work` |
| Automated verification | `kb-qa` / `kb-repair` |
| Review, compound, learn, cleanup | `kb-complete` |
| Release readiness | `kb-ship` |
| Context compaction/resume | `kb-handoff` / `kb-route` |
| Token compaction | `kb-compact` |
| Benchmarking the skill system | separate eval docs/scripts, not production workflow skills |

## Recommended KB Skill Tree

Minimum viable durable system:

- `kb-route` - start here, read minimal memory, choose lane.
- `kb-map` - cheap lookup/update of project memory.
- `kb-map-bootstrap` - expensive first-time repo indexing.
- `kb-compact` - compress memory, handoffs, skill drafts, and responses without dropping protected facts.
- `kb-fix` - small known bug/fix lane with bounded diagnosis.
- `kb-research` - reusable research note creation/refresh.
- `kb-brainstorm` - proportional product/requirements discovery.
- `kb-plan` - vertical-slice decomposition.
- `kb-work` - slice execution.
- `kb-qa` and `kb-repair` - automated verification and bounded repair.
- `kb-complete` - review, compound, learn, evolve, cleanup.
- `kb-ship` - release/PR/deploy readiness.
- `klfg` - full orchestrator for large hands-off runs.

Project memory naming decision:

- Use `todo.md` for active work, not `kb.md`.
- Use `todo-done.md` for completed-work summaries, not a giant backlog file.
- Use individual handoff files under `docs/handoffs/active/`, `docs/handoffs/parked/`, and `docs/handoffs/done/`.
- Keep `todo.md` small by linking to handoffs, brainstorms, plans, research notes, and subsystem docs.
- Refresh cold or parked work before execution when it is older than roughly 72 hours or touched subsystems changed.

Dark-factory requirement:

- `kb-plan` must produce slice metadata that an unattended executor can consume.
- `kb-work` must continue independent runnable slices when one slice parks.
- `kb-qa` and `kb-repair` must run before human QA is requested for normal app behavior.
- `kb-complete` must be the default terminal gate after all runnable slices finish.
- The final report must distinguish complete, parked, blocked, failed, and human-only states.

Scheduling requirement:

- Read-only research/review swarms are usually safe.
- Coding swarms require declared file ownership and non-overlap.
- Shared files, schemas, generated artifacts, routing, auth, chat/streaming, prompts, and central types force serial execution unless an enabling slice lands first.
- A parallel batch is not complete until the combined tree passes integration verification.
- If a worker needs a shared file unexpectedly, it should stop and escalate instead of editing through the conflict.

Optional later:

- `kb-question-tune` - stores "never ask this kind of thing again" preferences.
- `kb-benchmark` - scenario suite for cost, route correctness, QA quality, and dumb-question rate.
- `kb-architecture-review` - periodic Matt-style deep-module review for AI navigability.

## Research-To-Requirements Decisions

1. Keep the KB backbone. It already matches the user's desired factory: brainstorm -> plan -> work -> complete.
2. Add a default router, but keep it tiny. A new session should not load the entire factory.
3. Treat local memory as the product. `docs/context/PROJECT.md`, subsystem docs, research notes, `todo.md`, and active handoff files are how token use drops.
4. Separate setup/bootstrap from lookup. Deep crawling is allowed once, not on every ask.
5. Make dumb questions measurable. Track whether a question changed a decision; if not, tune it out.
6. Agent owns QA. Human-in-the-loop is for unavailable credentials, external systems, subjective approval, or genuinely blocked investigation.
7. Use escalation ceilings. Quick fixes should escalate to diagnose/brainstorm after repeated non-progress.
8. Benchmark the workflow with scenario tasks before declaring it mature.

## Open Questions

- Should `kb-route` be a real skill, a repo instruction block, or both?
- Should `kb-question-tune` exist now, or should question tuning be part of `kb-route` first?
- Should project memory live only in each repo, or should there also be a global user memory under the skill repo?
- How much of G-Stack-style browser persistence should KB assume versus defer to available host tools?
