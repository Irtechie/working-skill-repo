# Copilot Instructions

Use the KB workflow in this repo.

For ambiguous KB/workflow requests, start with `kb-route`. Skills live under `.github/skills/`.

Fresh-session preflight:

- If `todo.md` or `docs/context/PROJECT.md` is missing, run `kb-map-bootstrap` immediately.
- If the context folders or handoff folders are partially missing, run `kb-map refresh`.
- Do not ask for confirmation unless a non-empty user file would be overwritten.

Every token must pay rent:

- No preamble. No closing filler.
- Do not restate the request.
- Lead with the answer, route, command, or code.
- Keep exact paths, commands, error messages, decisions, risks, and safety warnings.
- Prefer short bullets over paragraphs.
- Expand only when detail changes the decision, prevents rework, or preserves safety.

Project memory:

- `todo.md` holds active work, blockers, parked work, and handoff pointers.
- `todo-done.md` holds completed-work summaries.
- `docs/context/PROJECT.md` is the project route map.
- `docs/handoffs/active/`, `docs/handoffs/parked/`, and `docs/handoffs/done/` hold resumable handoffs.

If local memory is missing or stale, use `kb-map-bootstrap`. For normal startup, use `kb-map` or `kb-route`.
