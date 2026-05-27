# Agent Instructions

For KB workflow requests, start with `kb-start`.

On every fresh session or ambiguous work request, let `kb-map` perform the KB memory preflight:

- Run `kb-map lookup <request>` before routing work.
- `kb-map` must resolve the active project root first and read memory from that repo only.
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

When local memory is missing or badly stale, use `kb-map`; it decides whether lookup, refresh, or bootstrap is required. For normal startup, use `kb-start`.

## Agent-Owned Verification

Do not ask the user to test normal application behavior when the agent can test it.

For apps with a UI frontend, if a change touches frontend code or user-visible UI behavior, verify it through the rendered UI with Playwright, CDP, or the repo's browser transport. Use real navigation, clicks, inputs, and programmatic DOM assertions. Do not substitute backend calls, source inspection, screenshots alone, or prose claims.

Use unit/integration tests, CLI/API probes, browser automation, screenshots, traces, logs, and DOM assertions as needed. Screenshots are evidence, not the pass/fail oracle.

Only ask the user to test when verification requires something the agent truly cannot access: credentials or MFA/session access not already available, subjective product/design judgment, external hardware or production-only systems, destructive/risky real-world action, or missing test input that cannot be safely generated.

If blocked, state exactly what was attempted, what command/tool failed, and what specific human input is needed.
