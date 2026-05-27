---
name: kb-handoff
description: Create or update repo-local KB handoff restart packets. Use when the user asks for a handoff, restart packet, fresh-session note, session transfer, pause/resume note, or wants the next agent/session to continue from local project memory.
argument-hint: "[what the next session should pick up]"
---

# KB Handoff

Create a compact repo-local restart packet so a fresh session can continue without relying on chat history.

## Root Rule

Resolve the active project root first:

```powershell
git rev-parse --show-toplevel
```

If no valid repo root exists, ask the user to change into the project directory or provide the project path. Do not write handoffs to home folders, drive roots, global skill folders, or `C:\Users\marowe\.copilot\handoffs`.

## Output Path

Write or update:

```text
docs/handoffs/active/YYYY-MM-DD-<short-topic>.md
```

Create `docs/handoffs/active/`, `docs/handoffs/parked/`, and `docs/handoffs/done/` if missing.

## What To Include

Keep the handoff under 1200 words unless the user explicitly asks for more.

1. **Purpose** — what the next session should do.
2. **Current state** — branch, repo root, relevant manifest/plan/todo status.
3. **Files to read first** — exact repo-local paths.
4. **Work completed** — compact facts only; link to commits, plans, or docs instead of duplicating.
5. **Next action** — the recommended `kb-start <task or handoff>` prompt.
6. **Blockers / HITL** — exact missing input, access, decision, or failing command.
7. **Verification state** — tests, QA, snapshots, or proof already run.

## Todo Pointer

Update `todo.md` only when the handoff represents active or blocked work:

- add or update a compact row pointing to `docs/handoffs/active/<file>.md`;
- use `🔒 blocked` for dependency/tool/access waits;
- use `🛑 human-required` for decisions or inputs only the user can provide;
- do not paste the whole handoff into `todo.md`.

## Rules

- A handoff is a restart packet, not an executable plan.
- If the handoff only has phases or broad next steps, set `Suggested route: kb-plan`.
- If it links a valid `docs/plans/*-kb-*-manifest.md`, set `Suggested route: kb-work`.
- If intent is unclear, set `Suggested route: kb-brainstorm`.
- Redact secrets, tokens, PII, and private credentials.
- Prefer exact file paths, commands, errors, and proof artifacts over prose.
