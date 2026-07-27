---
name: kb-compact
description: "Lower the cognitive burden of KB memory, docs, handoffs, skill drafts, or responses while preserving technical truth. Select the smallest useful presentation: plain language, ranked bullets, a comparison table, a decision block, or a workflow diagram when relationships matter. Use when the user says 'compact', 'fewer words', 'make this terse', 'organize the response', 'talk to me like a person', 'limited time', 'low cognitive burden', 'make this easy to understand', 'show me visually', 'token diet', 'every token pays rent', or when KB docs are getting too large for routine session startup."
argument-hint: "[file path, response, or doc area to compact]"
---

# KB Compact

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
3. For active work, show `Done | Next | Blocked` only when it improves
   orientation; do not restate unchanged state every turn.
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
| Sequence, dependency, branching, ownership, or state change | Mermaid or a compact text flow | At least three meaningful nodes and two relationships are easier to scan than describe |
| One human-owned decision | Decision block | The exact ask, why it matters, what blocks, and the recommendation must stay together |
| Supporting proof or nuance | `Details` section or linked artifact | It matters, but not before the outcome or decision |

A visual must earn its space. Skip it when it merely redraws a short list, when
the reader must study a legend, or when prose is faster. Prefer one useful
visual over several partial ones.

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

1. Identify the artifact and its purpose: startup memory, active task, handoff, research, architecture, or skill text.
2. Choose mode based on risk and requested scope.
3. Identify the information shape and choose the smallest useful format.
4. Rewrite for low cognitive burden: clear action, visible state, then optional
   depth.
5. Verify protected atoms survived exactly and the visual, if any, expresses
   real relationships.
6. Report what was compacted and anything intentionally moved or left unchanged.

## Output Contract

For file edits:

- Edit the file directly.
- Summarize token-saving shape, not fake exact token math.
- Name any protected content you checked.

For chat-only compression:

- Return only the compressed text unless the user asked for explanation.
