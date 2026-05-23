# Agent Instructions

For KB workflow requests, start with `kb-route`.

On every fresh session or ambiguous work request, perform the KB memory preflight:

- If `todo.md` or `docs/context/PROJECT.md` is missing, run `kb-map-bootstrap` immediately.
- If `docs/context/architecture/`, `docs/handoffs/active/`, `docs/handoffs/parked/`, or `docs/handoffs/done/` is missing, run `kb-map refresh`.
- Do not ask the user to confirm bootstrap or refresh unless the operation would overwrite non-empty user files.

Every token must pay rent. Be concise by default:

- No preamble or closing filler.
- Do not restate the user's request.
- Lead with the answer, route, command, or code.
- Keep exact paths, commands, errors, decisions, risks, and safety warnings.
- Use longer explanations only when they change the decision or reduce rework.

Use these project memory files:

- `todo.md` for active work, blockers, parked work, and handoff pointers.
- `todo-done.md` for completed-work summaries.
- `docs/context/PROJECT.md` for the project route map.
- `docs/handoffs/active/`, `docs/handoffs/parked/`, and `docs/handoffs/done/` for handoff lifecycle.

Do not treat these files as skills. Skills live under `.github/skills/`.

When local memory is missing or badly stale, use `kb-map-bootstrap`. For normal startup, use `kb-map` or `kb-route`.
