---
name: kb-compact
description: Compress KB memory, docs, handoffs, skill drafts, or responses while preserving technical truth. Use when the user says "compact", "fewer words", "make this terse", "token diet", "every token pays rent", or when KB docs are getting too large for routine session startup.
argument-hint: "[file path, response, or doc area to compact]"
---

# KB Compact

Reduce token load without deleting meaning. This is not a style bit; it is a preservation pass.

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

Default to Lite for durable docs and Full for chat/status output.

## Workflow

1. Identify the artifact and its purpose: startup memory, active task, handoff, research, architecture, or skill text.
2. Choose mode based on risk and requested scope.
3. Rewrite for density.
4. Verify protected atoms survived exactly.
5. Report what was compacted and anything intentionally moved or left unchanged.

## Output Contract

For file edits:

- Edit the file directly.
- Summarize token-saving shape, not fake exact token math.
- Name any protected content you checked.

For chat-only compression:

- Return only the compressed text unless the user asked for explanation.
