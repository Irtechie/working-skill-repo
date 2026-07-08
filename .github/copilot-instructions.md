# Copilot Instructions

Use the KB workflow in this repo.

For ambiguous KB/workflow requests, start with `kb-start`. Use `kb-task` when the user explicitly invokes it or asks for a first-principles task runner that should continue until done. Skills live under `.github/skills/`.

Fresh-session preflight:

- Run `kb-map lookup <request>` before routing work.
- `kb-map` must resolve the active project root first and read memory from that repo only.
- If `todo.md` or `docs/context/PROJECT.md` is missing, `kb-map` invokes `kb-map-bootstrap`.
- If context or handoff folders are partial, `kb-map` refreshes or creates the missing structure.
- Do not ask for confirmation unless a non-empty user file would be overwritten.

This repo is the portable skill bundle. Do not bootstrap consuming-project memory or create project-work handoffs here by accident. If the user is handing off work from another project, switch to that project root or ask for its path. Only create `todo.md`, `docs/context/PROJECT.md`, or `docs/handoffs/*` here when maintaining this skill bundle.

Contributor quality gate for this skill repo:

```shell
go run ./cmd/kbcheck core
```

This command is GHCP-compatible because it uses repo files and the native Go
`cmd/kbcheck` gate, not Codex-only tools. It runs repo-local deterministic
checks such as skill lint, route complexity fixture validation, eval selftests,
marketplace firebreak checks, and ATV drift reports that are available from the
checkout.

Release/sync gate:

```shell
go run ./cmd/kbcheck local-release
```

Run `local-release` before syncing or propagating skills. It composes `core`
with `git diff --check` and blocking read-only sync drift reports from
`config/skill-quality.json`.

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
- `docs/context/eval-map.md` maps repo-native eval surfaces and canonical proof commands.
- `docs/handoffs/active/`, `docs/handoffs/parked/`, and `docs/handoffs/done/` hold resumable handoffs.

If local memory is missing or stale, use `kb-map`; it decides whether lookup, refresh, or bootstrap is required. For normal startup, use `kb-start`.

## Context Engine (CCE)

CCE index is active. Prefer `context_search` over reading files directly.
Use `expand_chunk` for full source, `related_context` for call/import graph,
and `session_recall` / `record_decision` for cross-session memory.
