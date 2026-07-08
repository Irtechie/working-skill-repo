# HumanLayer Pinned Repos Planner Economy Research

Checked: 2026-07-05
Budget mode: deep

## Question

How should KB use the ideas in HumanLayer's three pinned repos to make decomposition
and planning reliable enough that expensive models stay mostly in planner/orchestrator
roles while smaller models execute bounded slices?

## Snapshot

Repos checked locally on 2026-07-05:

- `humanlayer/humanlayer` at `99abe673498cf8bdcd5f989aebe9406a27185b3b`
- `humanlayer/agentcontrolplane` at `eaa2a7ed1d9cb4e13dc53defaf420e36f481dcad`
- `humanlayer/12-factor-agents` at `d20c728368bf9c189d6d7aab704744decb6ec0cc`

`humanlayer/humanlayer` is useful as design archaeology, not as a direct code
source. Its README says the public code is mostly deprecated and points to the
rebuilt HumanLayer product.

## Fable Correction Incorporated

Fable's review changed the decision posture. The earlier bottom line, "keep KB
as core and use HumanLayer as reference architecture," was directionally useful
but too confident.

Corrected position:

- HumanLayer/CodeLayer already has durable sessions, approvals as state,
  HITL/state-machine mechanics, lineage, status, cost/token fields, and event
  history. Do not describe those as missing from HumanLayer.
- The public HumanLayer repo is not a clean fork target because its README says
  the code is mostly deprecated.
- A reported HumanLayer issue shows a session stuck in Needs Approval /
  INTERRUPTING and unable to recover. This does not invalidate durable state; it
  means recovery invariants must be deterministic and tested.
- The next decision point is a bounded absorption spike: add a real repo-local
  task-state store and context-packet object to KB. If it composes cleanly, KB
  can remain core. If it fights markdown-as-state, KB should become payload on a
  smaller runtime/state engine.

## Findings

### 1. Keep KB as payload now; prove core through an absorption spike

HumanLayer is stronger at harness thinking: context isolation, explicit control
flow, session state, approvals, subagent delegation, and task/event history. KB
is already stronger at repo-local project memory, vertical-slice manifests,
deterministic proof, sync discipline, scoped learning, and portable skills.

The best immediate merge is not replacing KB. It is hardening KB's planning
surface so the planner creates self-contained worker packets and the worker path
becomes cheap, boring, and measurable. Whether KB should also own durable runtime
state is an open question until the absorption spike proves it.

### 2. The newer RPI lesson is "split planning," not "write a bigger plan"

Dex's 2026 RPI correction is that broad research-plan-implement prompts fail
because they overload the instruction budget, rely on magic words, and make
humans review giant tactical plans instead of the leverage points.

Adopt this stage split in KB:

1. Questions: identify unknowns and research probes.
2. Objective research: facts only; hide the intended implementation when possible.
3. Design concept: current state, desired state, patterns to follow, resolved
   decisions, open questions.
4. Structure outline: header-file-like order of changes, interfaces, checkpoints,
   and tests.
5. Tactical plan: vertical slices with dependencies and proof.
6. Work: execute one bounded slice at a time.
7. Completion: review code/proof, learn, sync.

The expensive planner should own stages 1, 3, 4, risky stage 5 synthesis, and
final review. Research probes and many slice executions can be delegated when
they have narrow packets.

### 3. 12-factor gives the context contract KB is missing

The 12-factor repo's strongest ideas for KB are:

- own the prompt and context window as first-class code;
- prefetch deterministic context before asking a model to reason;
- unify execution state as a serializable event/history stream;
- use deterministic control flow around the model;
- compact errors before retrying;
- keep agents small and focused.

KB already has `expected_files`, `verification`, `functional_risk`, and
`model_tier`. The gap is that a slice does not yet carry the exact context
payload a cheap worker should receive.

Add a required `context_packet` for new non-trivial slices:

- repo-memory files checked;
- source files/interfaces already read and why they matter;
- deterministic prefetch outputs such as `rg` inventories, routes, schema/API
  summaries, failing checks, dependency edges, or active landmines;
- constraints and out-of-scope boundaries;
- acceptance/proof target;
- `model_tier` and `model_tier_reason`;
- escalation triggers for returning to the planner.

This is the main cost-control lever. Small models are cheap when the planner has
already removed ambiguity.

### 4. ACP gives KB the durable task shape

Agent Control Plane models work as `Agent`, `Task`, and `ToolCall` objects with
phases, status, context window, output, errors, spans, tool calls, and delegated
child tasks.

Useful KB imports:

- represent a slice as a durable task record, not just markdown prose;
- keep task phase/status separate from final output;
- make subagent delegation a single structured message to a child task;
- keep parent/child lineage for forks and delegated work;
- record context window/message count or a lightweight equivalent;
- treat approvals and human input as task states, not informal chat.

Do not import Kubernetes/CRDs now. The pattern can live in repo-local files and
Go validators first.

### 5. HumanLayer/HLD gives KB runtime evidence, not a blank slate

The HumanLayer daemon code persists sessions, conversation events, tool calls,
approvals, file snapshots, parent session IDs, token usage, cost, turns, status,
and result/error state.

KB should borrow the state/telemetry ideas, but first prove whether it can do so
without turning markdown into a brittle database:

- predicted model tier from the plan;
- actual model/tier used;
- token/cost/turn estimate when available;
- files touched/read;
- proof commands and outcomes;
- rework count;
- whether the slice was under-decomposed or over-decomposed;
- whether the worker had to escalate because the packet was insufficient.

This telemetry becomes the feedback loop that teaches KB where Haiku-class,
mini-class, Sonnet-class, and top-tier planner models actually succeed.

### 5a. State machines need recovery invariants

HumanLayer's stuck-state issue is a useful warning for KB's self-healing work. A
state machine is not safer by virtue of being explicit. It is safer only when:

- every blocked/waiting/interrupted state has a deterministic next action;
- stale states can be detected without asking a model to infer them;
- resume/fork rules are tested;
- invalid transitions fail with repair hints;
- human input is represented as state, not lost in chat prose.

The absorption spike should include a stale waiting/interrupted fixture and a
kill/resume proof before claiming self-healing.

### 6. Model tiering should be capability based, not vendor-name based

Keep the manifest field as `tiny/small/medium/large` and map those tiers to
runtime models outside the plan. Vendor model names change faster than the skill
contract.

Suggested tier contract:

| Tier | Best work | Planner packet requirement | Escalate when |
|---|---|---|---|
| `tiny` | grep inventories, file lists, route classification, schema/frontmatter fill, simple doc edits | exact files/commands/output shape | any ambiguity, conflicting facts, code behavior change |
| `small` | narrow mechanical edits, fixture updates, simple tests, localized refactors | expected files, interfaces, before/after shape, proof command | API design, cross-file uncertainty, failing proof not obvious |
| `medium` | normal vertical slices, moderate debugging, integration tests | complete context packet and bounded slice goal | architecture/security/data model decisions |
| `large` | decomposition, design, hard debugging, security/architecture, final synthesis | may gather/shape context | when the plan itself is wrong or requirements shift |

Practical mapping: Haiku/mini-class models belong mostly to `tiny` and some
`small`; Sonnet-class engines belong to `medium`; Opus/top Codex-class models
stay on `large` planner/orchestrator/reviewer work. The repo contract should
not hardcode those names.

### 7. Vertical slices need a size budget

HumanLayer's small-agent guidance maps cleanly to KB slice sizing:

- target 3-10 meaningful worker steps;
- hard-review anything likely to exceed 20 steps;
- each slice must include one user-visible or behavior-visible proof when the
  work is application behavior;
- if a slice requires broad rediscovery, split or enrich the context packet;
- if a worker needs new design decisions, stop and return to planner.

This is how KB avoids spending large-model tokens on implementation churn.

### 8. What KB is already better at

KB should keep these advantages:

- vertical-slice manifests and dependency DAGs;
- deterministic `kbcheck` gates;
- repo-local memory and handoff structure;
- scoped learning that prevents sibling-workflow contamination;
- sync drift checks across working/global/ATV skill roots;
- proof-first completion and learning adoption;
- cross-runtime portability.

HumanLayer's pinned repos show stronger harness/runtime ideas, but they do not
replace KB's skill-bundle governance.

### 9. Segment custom instructions, agents, subagents, skills, commands, and tools by ownership

KB currently has 38 skills and 51 reviewer/specialist agents. That is useful
surface area, but it needs a stricter contract so planning does not blur
"workflow," "persona," "runtime child task," and "entrypoint."

Use this segmentation:

| Surface | Owns | Should contain | Should not contain |
|---|---|---|---|
| Custom instruction | Ambient host/project policy | durable repo invariants, safety rules, source precedence, canonical commands, always-loaded constraints | live task state, long workflow bodies, persona prompts, volatile research, broad decomposition |
| Command | User or host entrypoint | name, args, target skill or deterministic command, local/global source | workflow logic, persona prompts, broad instructions |
| Skill | Workflow policy | route, gates, artifacts, escalation, proof contract, lazy references | repo facts that belong in project memory, persona-specific review heuristics |
| Agent | Capability/persona | narrow expertise, evidence rules, output schema, read/write permissions | orchestration, cross-slice planning, global workflow state |
| Subagent | Ephemeral execution | one packet, one context window, one result, parent/child lineage, model tier | independent planning authority, broad rediscovery, unbounded tool use |
| Tool | Deterministic actuator | typed input/output, side-effect policy, approval policy, compact result | hidden reasoning or policy that should live in skills/control flow |

HumanLayer's HLD treats slash commands as discoverable objects with `name` and
`source` (`local` or `global`). ACP treats subagents as tools that create child
tasks. 12-factor treats tools as structured outputs that deterministic code
executes. Custom instructions are also valid, but they are the ambient policy
layer: `AGENTS.md`, `CLAUDE.md`, Copilot instructions, and similar host-loaded
files should set stable rules and point into skills, not absorb the full
workflow. Combined for KB, the shape should be:

```text
custom instructions -> command -> skill -> plan/slice task -> optional subagent(agent + packet) -> tools/checks
```

Practical rules:

- Custom instructions stay concise and stable: repo invariants, source
  precedence, safety rules, canonical commands, and "start with kb-start" style
  routing defaults.
- Commands stay thin aliases such as `kb-start`, `kb-plan`, or `kbcheck core`.
- Skills decide stage, artifacts, gates, and when delegation is allowed.
- Agents are reusable specialist definitions, not workflow routers.
- Subagents are runtime invocations of an agent plus a `context_packet`, not new
  durable skill types.
- Tools and `kbcheck` commands perform deterministic work and return compact,
  parseable evidence.
- Local commands/skills override global ones only by explicit source precedence
  and drift checks.
- A cheap worker can only run as a subagent when the parent supplies the packet,
  proof target, allowed files/tools, and escalation triggers.

This segmentation should become a `kbcheck` inventory/lint surface so the bundle
can catch bloated custom instructions, bloated commands, orchestration leaking
into agents, and skills that load too much surface by default.

## Recommended Implementation Plan

1. Run one bounded absorption spike before any fork/replacement decision.
2. Add a minimal repo-local task-state store and context-packet object, validated
   by `kbcheck`.
3. Include stale waiting/interrupted recovery fixtures inspired by the
   HumanLayer stuck-state issue.
4. Wire `kb-plan` and `kb-work` to produce/consume packet data only after the
   state object proves clean.
5. Add telemetry for predicted tier, actual tier/model, proof, rework,
   escalation, and packet sufficiency.
6. Add a custom-instructions/agents/subagents/skills/commands/tools
   segmentation contract and one daily-runtime adapter boundary.
7. Write a decision report: KB remains core, KB becomes payload on a small
   runtime, or a named replacement/fork is justified.

## Rejected Approaches

- Replace KB with Phoenix/HumanLayer/ACP wholesale: loses KB memory, scoped
  learning, sync governance, and deterministic proof.
- Claim KB can absorb runtime state cleanly without an absorption spike.
- Use a bakeoff as the first decision point; comparing a mature runtime to a
  week-old prototype mostly measures maturity, not architecture fit.
- Score plan quality or proof quality by model judgment alone.
- Let cheap models plan from broad context: this saves money only until rework
  starts.
- Give workers broad repo-search authority by default: it burns context and
  makes results less reproducible.
- Review giant tactical plans as the main quality gate: review design concept,
  structure outline, proof, and code instead.
- Hardcode vendor model names into durable manifests: use capability tiers and
  map them at runtime.
- Build an ACP-style Kubernetes runtime now: useful architecture, wrong scale
  for this portable skill bundle today.
- Build five adapters up front; prove one daily-runtime adapter boundary first.
- Let commands, skills, and agents all become interchangeable prompt blobs:
  unclear ownership bloats context and makes cheap delegation unsafe.
- Let custom instructions become a second workflow system: they are always-loaded
  guardrails and pointers, not a place to hide live plans or specialist agents.

## Sources

- HumanLayer repo: https://github.com/humanlayer/humanlayer
- HumanLayer issue #954, session stuck in Needs Approval / INTERRUPTING:
  https://github.com/humanlayer/humanlayer/issues/954
- Agent Control Plane: https://github.com/humanlayer/agentcontrolplane
- 12-factor agents: https://github.com/humanlayer/12-factor-agents
- Factor 2, Own your prompts: https://github.com/humanlayer/12-factor-agents/blob/main/content/factor-02-own-your-prompts.md
- Factor 3, Own your context window: https://github.com/humanlayer/12-factor-agents/blob/main/content/factor-03-own-your-context-window.md
- Factor 4, Tools are structured outputs: https://github.com/humanlayer/12-factor-agents/blob/main/content/factor-04-tools-are-structured-outputs.md
- Factor 5, Unify execution state: https://github.com/humanlayer/12-factor-agents/blob/main/content/factor-05-unify-execution-state.md
- Factor 8, Own your control flow: https://github.com/humanlayer/12-factor-agents/blob/main/content/factor-08-own-your-control-flow.md
- Factor 9, Compact errors: https://github.com/humanlayer/12-factor-agents/blob/main/content/factor-09-compact-errors.md
- Factor 10, Small focused agents: https://github.com/humanlayer/12-factor-agents/blob/main/content/factor-10-small-focused-agents.md
- Factor 13, Pre-fetch context: https://github.com/humanlayer/12-factor-agents/blob/main/content/appendix-13-pre-fetch.md
- ACP subagent tool adapter: https://github.com/humanlayer/agentcontrolplane/blob/main/acp/internal/controller/task/task_controller.go
- HLD slash command model: https://github.com/humanlayer/humanlayer/blob/main/hld/sdk/typescript/src/generated/models/SlashCommand.ts
- Advanced Context Engineering: https://www.humanlayer.dev/blog/advanced-context-engineering
- Harness Engineering: https://www.humanlayer.dev/blog/skill-issue-harness-engineering-for-coding-agents
- Long-Context Isn't the Answer: https://www.humanlayer.dev/blog/long-context-isnt-the-answer
- Research-Plan-Implement talk transcript mirror: https://raw.githubusercontent.com/shanraisshan/claude-code-best-practice/main/videos/claude-dex-mlops-community-24-mar-26.md

## Applies When

- Planning KB decomposition, model-tier routing, worker delegation, or cost
  reduction.
- Deciding what tiny/small/medium/large workers are allowed to do.
- Deciding whether something belongs as a command, skill, agent, subagent
  invocation, tool, or `kbcheck` command.
- Designing context packets, task ledgers, or slice execution telemetry.

## Stale When

- HumanLayer releases a new public codebase that replaces the deprecated
  `humanlayer/humanlayer` repo.
- KB completes the absorption spike and decides whether KB remains core or
  becomes payload on a smaller runtime/state engine.
- Runtime model pricing/capabilities change enough that the tier-to-model
  mapping needs a fresh pass.

## Impact On Current Project

Promote the current "per-slice context packets" queued item into a bounded
planner-economy absorption spike. The next concrete KB work should be the
manifested spike for task state, context packets, recovery fixtures, telemetry,
custom-instruction segmentation, and one adapter boundary. Implementation
remains blocked until the user authorizes that spike.
