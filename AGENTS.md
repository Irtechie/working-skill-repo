# Agent Instructions

For KB workflow requests, start with `kb-route`.

On every fresh session or ambiguous work request, let `kb-map` perform the KB memory preflight:

- Run `kb-map lookup <request>` before routing work.
- If `todo.md` or `docs/context/PROJECT.md` is missing, `kb-map` invokes `kb-map-bootstrap`.
- If context or handoff folders are partial, `kb-map` refreshes or creates the missing structure.
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

When local memory is missing or badly stale, use `kb-map`; it decides whether lookup, refresh, or bootstrap is required. For normal startup, use `kb-route`.
