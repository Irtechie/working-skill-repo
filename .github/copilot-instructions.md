# Copilot Instructions

Use the KB workflow in this repo.

For ambiguous KB/workflow requests, start with `kb-start`. Skills live under `.github/skills/`.

Fresh-session preflight:

- Run `kb-map lookup <request>` before routing work.
- `kb-map` must resolve the active project root first and read memory from that repo only.
- If `todo.md` or `docs/context/PROJECT.md` is missing, `kb-map` invokes `kb-map-bootstrap`.
- If context or handoff folders are partial, `kb-map` refreshes or creates the missing structure.
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

If local memory is missing or stale, use `kb-map`; it decides whether lookup, refresh, or bootstrap is required. For normal startup, use `kb-start`.
