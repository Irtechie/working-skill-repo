# Dexhorthy / HumanLayer Agent Harness Research

Checked: 2026-07-05
Budget mode: standard

## Question

What can KB borrow from Dex Horthy / HumanLayer without replacing KB, and which ideas should become future KB skill-bundle work?

## Findings

### 1. Merge, do not replace KB

HumanLayer is strongest at context/harness discipline and recurring automation. KB is already stronger at project-local memory, vertical-slice manifests, scoped learning, gate ledgers, deterministic proof, cross-runtime sync, and model-tiered decomposition.

Adopt HumanLayer ideas as narrower surfaces:

- recurring loop generator / template;
- `/iterate` feedback ingestion;
- context-efficient check wrappers;
- conditional instruction compression;
- context-fork/replay harness notes;
- optional durable runtime adapter inspired by ACP/Shannon, not a KB replacement.

### 2. RPI and frequent intentional compaction map to KB but sharpen review economics

HumanLayer's RPI pattern uses research -> plan -> implement, with compaction into docs and human review focused on research/plan because a bad research or plan line can create many bad code lines.

KB already has `kb-research`, `kb-plan`, `kb-work`, and `kb-complete`. Gap: KB should make "review the research/plan before code" cheaper and more explicit for high-risk work, maybe by adding a plan-quality/risk checklist or `kbcheck plan-audit`.

### 3. Slice context packets are the highest-value planning improvement

Dex's 12-factor context work says the model input should be owned and shaped for the job, not left as a generic chat transcript. The planning analogue for KB is: each slice should carry a compact, deliberate context packet that a child agent or lower-cost model can execute from without rediscovering the world.

KB already has `expected_files`, acceptance criteria, verification, functional risk, and model tier. Gap: the slice plan does not yet explicitly say "this is the context payload the worker should receive." That leaves too much context assembly to the executor and makes tiny/small model delegation weaker than it needs to be.

Adopt a `context_packet` or `context_prefetch` section per slice:

- relevant repo-memory files already checked;
- source files/interfaces already read and why they matter;
- deterministic prefetch results, such as `rg` inventories, API/schema summaries, existing failing checks, route maps, or dependency edges;
- constraints, active landmines, protected oracles, and out-of-scope boundaries;
- acceptance/proof target in one compact block;
- `model_tier` plus `model_tier_reason`;
- escalation triggers for when a tiny/small worker must hand back to medium/large.

This combines Factor 3 "own your context window", Factor 13 prefetch, and Factor 10 small focused agents. It should improve slice handoffs, reduce repeated repo crawling, and make cheap-model work safer.

### 4. Control-loop automation is the biggest steal

HumanLayer's `design-control-loop` frames recurring work as set point, sensor, controller, actuator, disturbances, dampener, scope gate, batch size, memory, and WIP bound.

KB now records these fields in `kb-plan` and `kb-goal`, but does not yet generate a runnable recurring workflow or local sensor/controller scaffold.

Adopt next: `kb-control-loop` or a `kb-goal` extension that creates:

- local sensor command script;
- controller script/policy file selecting one reviewable increment;
- actuator prompt/skill pointer;
- `.github/agent-memory/<loop>.md` or KB equivalent steering file;
- PR bound of 1 open PR;
- optional `/iterate` marker that routes reviewer comments into steering memory;
- optional dampener check on PRs.

### 5. `/iterate` is better than raw learning for recurring PR loops

HumanLayer separates durable loop feedback from one-off PR comments. Maintainers can comment `/iterate`; workflow reads PR body/comments/memory and updates the same PR plus the memory file.

KB already has feedback classification: current-only, steering-memory, observation, scoped-instinct, landmine, and instinct-evidence. Gap is runtime ingestion from PR comments. Adopt the routing shape, but write feedback into KB steering memory first, not straight to instincts.

### 6. Context-efficient backpressure should become a `kbcheck` output mode

HumanLayer's backpressure wrapper prints compact success and only dumps output on failure. Fail-fast and output filtering reduce context damage.

KB has deterministic gates, but many commands still return noisy output through the agent. Adopt:

- `kbcheck run --summary` or a check wrapper: success emits one compact line with counts/timestamps; failure emits command, exit code, focused stderr/stdout, and artifact path;
- fail-fast mode when available (`go test -failfast`, `pytest -x`, etc.);
- verbose logs stored as artifacts and cited by path.

### 7. Instruction compression via `<important if>` is useful but runtime-specific

HumanLayer's `improve-claude-md` likely helps Claude Code by matching its system-prompt relevance semantics. KB is cross-runtime, so do not copy the tag as a global convention.

Adopt the principle:

- root `AGENTS.md`/README contains only universal context and command map;
- conditional/task-specific rules move into skills/references;
- if generating Claude-specific files, use narrow `<important if>` wrappers.

### 8. ACP validates Go for durable orchestrator mechanics

ACP is Go/Kubernetes and models agent execution as CRDs: Agent, Task, ToolCall, MCPServer, LLM, and ContactChannel. Task status owns phase, context window, output, error, spans, and related execution state.

For KB, this supports keeping deterministic local machinery in Go (`cmd/kbcheck`). Rust is not required for orchestration.

Rust only becomes attractive if we need:

- high-performance AST/indexing across huge repos;
- strong sandbox boundaries;
- native cross-platform daemon with tight memory/control guarantees;
- parser-heavy refactoring engine.

Otherwise Go remains the pragmatic answer.

### 9. Shannon and multiclaude provide runtime adapter ideas

Shannon drives real interactive Claude in tmux, tails transcript JSONL, emits SDK-like stream JSON, and builds parity fixtures from native `claude -p` / Agent SDK behavior using Haiku.

Multiclaude creates worktrees, tmux windows, and plan-file prompts for parallel agents.

KB should not import these wholesale. Adopt:

- transcript/session adapter idea for replay/fork/resume evidence;
- native fixture parity testing for any harness adapter;
- safer worktree launch with KB scope leases and non-destructive cleanup.

### 10. Model-tier decomposition is already better in KB

HumanLayer talks about using Opus for parent/orchestration and Sonnet/Haiku for bounded subagents. KB already encodes `tiny/small/medium/large` in `kb-plan` and `kb-work`, including escalation boundaries and a constant proof bar.

Refinement: add examples in the model-tier contract:

- tiny: grep inventories, schema/frontmatter fill, route classification, docs copy;
- small: narrow mechanical edits and simple tests;
- medium: ordinary vertical slice;
- large: decomposition, hard debugging, security/architecture, final synthesis.

Also consider a `model_tier_reason` field in slice plans if reviews show assignments are opaque.

## Sources

- GitHub profile: https://github.com/dexhorthy
- HumanLayer skills: https://github.com/humanlayer/skills
- `design-control-loop`: https://raw.githubusercontent.com/humanlayer/skills/main/plugins/design-control-loop/skills/design-control-loop/SKILL.md
- Control-loop taxonomy: https://raw.githubusercontent.com/humanlayer/skills/main/plugins/design-control-loop/skills/design-control-loop/references/control-loop-taxonomy.md
- `build-iterated-agentic-loop`: https://github.com/humanlayer/skills/blob/main/plugins/build-iterated-agentic-loop/skills/build-iterated-agentic-loop/SKILL.md
- 12-factor agents: https://github.com/humanlayer/12-factor-agents
- Factor 3, Own your context window: https://github.com/humanlayer/12-factor-agents/blob/main/content/factor-03-own-your-context-window.md
- Factor 10, Small focused agents: https://github.com/humanlayer/12-factor-agents/blob/main/content/factor-10-small-focused-agents.md
- Factor 13, Pre-fetch context: https://github.com/humanlayer/12-factor-agents/blob/main/content/appendix-13-pre-fetch.md
- Advanced Context Engineering: https://www.humanlayer.dev/blog/advanced-context-engineering
- Harness Engineering: https://www.humanlayer.dev/blog/skill-issue-harness-engineering-for-coding-agents
- Context-Efficient Backpressure: https://www.humanlayer.dev/blog/context-efficient-backpressure
- Context Forking: https://www.humanlayer.dev/blog/context-forking-to-save-time-trouble-and-tokens
- Long-Context Isn't the Answer: https://www.humanlayer.dev/blog/long-context-isnt-the-answer
- Getting Claude to Read CLAUDE.md: https://www.humanlayer.dev/blog/stop-claude-from-ignoring-your-claude-md
- Agent Control Plane: https://github.com/humanlayer/agentcontrolplane
- Shannon: https://github.com/dexhorthy/shannon
- Multiclaude: https://github.com/dexhorthy/multiclaude

## Applies When

- Designing KB self-healing, recurring improvement loops, or scheduled agent work.
- Deciding whether KB should become HumanLayer/Phoenix/ACP versus absorb pieces.
- Improving model-tier decomposition and subagent cost routing.
- Reducing context blowups from tests, logs, MCP tools, and broad instructions.

## Stale When

- HumanLayer changes the public skills/workflows materially.
- Codex gains first-class hooks, context forking, PR comment routing, or stable subagent runtime APIs that make custom workflow glue unnecessary.
- KB implements a recurring loop generator and validates it against real repos.

## Rejected Approaches

- Replace KB with HumanLayer skills wholesale: loses KB gate ledger, scoped learning, sync discipline, and deterministic proof.
- Adopt Claude-only `<important if>` everywhere: useful for Claude, not guaranteed for Codex/GHCP.
- Build a Kubernetes ACP clone now: too much runtime surface for the portable skill repo.
- Switch `cmd/kbcheck` to Rust now: no evidence current Go tooling is the bottleneck.
- Turn all lessons into global memory: contradicts KB scoped-learning model and causes sibling contamination.

## Impact On Current Project

Recommended next slices:

1. Add `context_packet` / `context_prefetch` to KB slice plans and make `kb-work` load it before local or delegated execution.
2. Add `kbcheck run` or `kbcheck check-summary` for context-efficient command output.
3. Add a KB recurring-loop generator or `kb-goal` extension for local sensor/controller/actuator scaffolds plus PR-bound workflow.
4. Add `/iterate`-style steering-memory ingestion for PR comments or local review notes.
5. Add optional Claude-targeted AGENTS/CLAUDE conditionalization helper, gated as runtime-specific.
6. Add harness-adapter parity fixture pattern from Shannon before any tmux/session adapter is trusted.
