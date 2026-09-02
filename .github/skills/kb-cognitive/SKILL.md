---
name: kb-cognitive
description: "Lower the cognitive burden of KB memory, docs, handoffs, skill drafts, user-facing completion/status returns, or PR first screens while preserving technical truth. Select the smallest useful presentation and escalate from prose through static visuals to a bounded interactive walkthrough only when the information shape earns it. Use when the user says 'cognitive', 'cognitive burden', 'compact', 'fewer words', 'make this terse', 'organize the response', 'talk to me like a person', 'limited time', 'low cognitive burden', 'make this easy to understand', 'easy to read', 'show me visually', 'token diet', 'every token pays rent', or when KB docs are getting too large for routine session startup."
argument-hint: "[file path, response, or doc area to make easier to read]"
---

# KB Cognitive

Reduce comprehension and decision effort without deleting meaning. Shorter is
useful only when it makes the artifact easier to act on. This is not a style
bit; it is a preservation pass.

## Protect

Do not change:

- File paths, commands, flags, env vars, URLs, branch names, IDs, error text, dates, numbers.
- Requirements, acceptance criteria, blockers, HITL reasons, stale thresholds, and safety warnings.
- Code blocks unless the user explicitly asks to edit code.
- Links between `todo.md`, handoffs, plans, brainstorms, research notes, and architecture docs.

## Cut

- Preamble, recap, motivation, thanks, sign-off, and "let me know" text.
- Duplicate rules already present in `AGENTS.md`, `.github/copilot-instructions.md`, or the relevant skill.
- Chatty explanation that does not change execution.
- Historical detail that belongs in `todo-done.md`, `docs/context/history/`, or a linked research note.

## Return Boundary

Apply this pass when control returns to the user with a completion, meaningful
status, blocker, decision, or pull-request first screen. Internal reasoning,
tool calls, and subagent exchanges are not presentation boundaries; keep their
working detail when it helps finish the task.

Put practical meaning first and the exact technical term second. A reader should
understand what a cache, queue, framework, protocol, model route, or version is
doing before they must recognize its product name.

When ownership of the next action could be unclear, lead with exactly one:

| State | Meaning |
|---|---|
| **Done** | The requested endpoint is reached; no response is needed |
| **Agent continues** | Safe work remains and the agent owns the next action |
| **You need to decide** | Only the user can authorize, supply, or choose the disposition after safe agent-owned repair is exhausted |

Ordinary direct answers do not need a decorative state label. Completion,
meaningful status, blockers, and decisions do.

## Modes

- **Lite**: tighten prose; preserve headings and most bullets.
- **Full**: convert prose to short bullets; remove repeated rationale.
- **Surgical**: compact only the requested section.

Default to Lite for durable docs. For chat/status output, choose the mode that
minimizes cognitive burden; do not force Full when a short explanation prevents
confusion or rework.

## Response Shape

For chat or status output:

1. Lead with the outcome or next action, never an announcement.
2. Put the decision, blocker, or required action before supporting detail.
3. Use **Done**, **Agent continues**, or **You need to decide** when control
   ownership is not otherwise obvious. Do not restate an unchanged state on
   routine progress turns.
4. Keep the primary surface to five ranked items. Put optional depth under
   `Details` or `Later`; never hide protected facts to satisfy the cap.
5. Give time estimates only when grounded in observed work or a known wait.
6. End when complete. Do not manufacture a user action or closing recap.

## Format Selection

Choose the smallest format that prevents the reader from reconstructing the
relationships themselves. Do not default to bullets, tables, or diagrams.

| Information shape | Best default | Use it when |
|---|---|---|
| One answer, outcome, or action | One to three plain sentences | There is no meaningful comparison or sequence |
| Several independent points | Ranked bullets | Order matters, but the items do not share repeated fields |
| Options, mappings, states, or repeated fields | Table | Side-by-side comparison removes repeated prose |
| Visible UI change | One real screenshot with a short caption | Rendered evidence shows the change faster than prose; never synthesize one for a nonvisual change |
| Sequence, dependency, branching, ownership, or state change | Mermaid or a compact text flow | At least three meaningful nodes and two relationships are easier to scan than describe |
| One bounded medium-complexity path | `interactive-workflow-workbench-light` | Selecting a few nodes or steps reduces effort and Mermaid is insufficient |
| Epic, multi-path, deep-evidence, or client/showcase walkthrough | `interactive-workflow-workbench` | The reader needs richer exploration than one bounded view can provide |
| One human-owned decision | Decision block | The exact ask, why it matters, what blocks, and the recommendation must stay together |
| Supporting proof or nuance | `Details` section or linked artifact | It matters, but not before the outcome or decision |

A visual must earn its space. Skip it when it merely redraws a short list, when
the reader must study a legend, or when prose is faster. Prefer one useful
visual over several partial ones.

Use `pr-review-workbench` for a source-backed pull-request review that needs
commit-pinned impact, source anchors, evidence gaps, or reviewer drill-down. It
is not the generic full presentation workbench.

Interactive HTML never triggers merely because a PR exists. If an optional
visual skill is unavailable, fall back to the best static format and continue;
do not block completion, review, or delivery.

Read [references/response-patterns.md](references/response-patterns.md) when the
format choice is unclear, the user asks for examples, or a reusable response
contract is being written. Do not load it for routine one-line answers.

## Response Responsibility

Before asking the user anything, classify it:

| Class | Meaning | Behavior |
|---|---|---|
| `hard response required` | Only the user can authorize, supply, or decide it, and dependent work cannot safely continue | Ask plainly; state why the user must answer, what blocks, and the recommended option |
| `soft preference` | The user may care, but the agent has a safe, reversible default | State the default and continue unless overridden |
| `no response needed` | Status, proof, completion, or an agent-owned decision | Inform without asking |

Do not turn an agent-owned decision into review work for the user. If several
options are reasonable but one is evidence-backed, recommend it and handle it.
Use plain human language; define unavoidable jargon once.

## Workflow

1. Identify whether this is a user-facing return boundary and, if so, who owns
   the next action.
2. Identify the artifact and its purpose: startup memory, active task, handoff,
   research, architecture, skill text, completion, status, or PR first screen.
3. Choose mode based on risk and requested scope.
4. Identify the information shape and choose the smallest useful format.
5. Rewrite for low cognitive burden: practical meaning, visible state, then optional
   depth.
6. Verify protected atoms survived exactly and the visual, if any, expresses
   real relationships.
7. Report what was compacted and anything intentionally moved or left unchanged.

## Output Contract

For file edits:

- Edit the file directly.
- Summarize token-saving shape, not fake exact token math.
- Name any protected content you checked.

For chat-only compression:

- Return only the compressed text unless the user asked for explanation.
- On a completion, meaningful status, blocker, or decision, verify the reader
  can identify **Done**, **Agent continues**, or **You need to decide** within
  five seconds.
- Do not manufacture a follow-up action after **Done**.
