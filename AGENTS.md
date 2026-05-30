# Agent Instructions

For KB workflow requests, start with `kb-start`, unless the user explicitly invokes `kb-task` or asks for a first-principles task runner that should continue until done.

On every fresh session or ambiguous work request, let `kb-map` perform the KB memory preflight:

- Run `kb-map lookup <request>` before routing work.
- `kb-map` must resolve the active project root first and read memory from that repo only.
- If `todo.md` or `docs/context/PROJECT.md` is missing, `kb-map` invokes `kb-map-bootstrap`.
- If context or handoff folders are partial, `kb-map` refreshes or creates the missing structure.
- Do not ask the user to confirm bootstrap or refresh unless the operation would overwrite non-empty user files.

This repo is the portable skill bundle. Do not bootstrap consuming-project memory or create project-work handoffs here by accident. If the user is trying to hand off work from another project, switch to that project root or ask for its path. Only create `todo.md`, `docs/context/PROJECT.md`, or `docs/handoffs/*` in this repo when the work is explicitly about maintaining this skill bundle.

## Skill Sync Workflow

When changing skills in this repo, treat `E:\working-skill-repo` as the working bundle source, but check for newer drift before overwriting anything.

1. Compare the target skill across:
   - `E:\working-skill-repo\.github\skills\<skill>\`
   - `E:\all-the-vibes\.github\skills\<skill>\`
   - `E:\all-the-vibes\pkg\scaffold\templates\skills\<skill>\`
   - `E:\all-the-vibes\plugins\atv-everything\skills\<skill>\`
   - `C:\Users\marowe\.copilot\skills\<skill>\`
   - `C:\Users\marowe\.agents\skills\<skill>\`
   - `C:\Users\marowe\.codex\skills\<skill>\`
2. If a global or ATV copy differs, review the diff before copying over it. Newer useful work found only in a global install must be merged back into this repo first, not discarded.
3. After editing this repo, sync the final approved copy to:
   - Codex global: `C:\Users\marowe\.codex\skills\<skill>\`
   - Copilot global: `C:\Users\marowe\.copilot\skills\<skill>\`
   - shared agents global: `C:\Users\marowe\.agents\skills\<skill>\`
   - ATV fork: `E:\all-the-vibes\.github\skills\<skill>\`
   - ATV scaffold/plugin copies when that skill is shipped there.
4. Update `README.md` in this repo when the visible workflow, installed-skill list, install commands, or repo hygiene contract changes.
5. Update `E:\all-the-vibes\README.md` when the ATV-facing workflow, bundled skills, or scaffold/plugin behavior changes.
6. Verify with hashes for copied `SKILL.md` files and `git diff --check` in every touched repo.
7. Commit and push both repos when requested or when the user asks for the full propagation flow.

Before syncing or propagating skills, run the canonical skill-repo quality gate:

```powershell
.\.github\skills\kb-check\scripts\kb-check.ps1 -All
```

This gate is cross-runtime: it validates the shared skill contract for Codex and GitHub Copilot/GHCP using `config/skill-quality.json`, deterministic skill lint, route-complexity fixtures, and a read-only sync drift report. The sync report may warn about optional ATV scaffold/plugin differences, but required targets must stay clean before propagation.

Do not remove `kb-review`, `ce-review`, `ce-compound`, or `ce-compound-refresh` from this bundle unless the skills that invoke them are rewritten first. KB completion uses `kb-review`; `ce-review` remains the generalized CE review skill.

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
- `docs/context/eval-map.md` for repo-native eval surfaces and canonical proof commands.
- `docs/handoffs/active/`, `docs/handoffs/parked/`, and `docs/handoffs/done/` for handoff lifecycle.

Do not treat these files as skills. Skills live under `.github/skills/`.

When local memory is missing or badly stale, use `kb-map`; it decides whether lookup, refresh, or bootstrap is required. For normal startup, use `kb-start`.

## Agent-Owned Verification

Do not ask the user to test normal application behavior when the agent can test it.

For apps with a UI frontend, if a change touches frontend code or user-visible UI behavior, verify it through the rendered UI with Playwright, CDP, or the repo's browser transport. Use real navigation, clicks, inputs, and programmatic DOM assertions. Do not substitute backend calls, source inspection, screenshots alone, or prose claims.

Use unit/integration tests, CLI/API probes, browser automation, screenshots, traces, logs, and DOM assertions as needed. Screenshots are evidence, not the pass/fail oracle.

Only ask the user to test when verification requires something the agent truly cannot access: credentials or MFA/session access not already available, subjective product/design judgment, external hardware or production-only systems, destructive/risky real-world action, or missing test input that cannot be safely generated.

If blocked, state exactly what was attempted, what command/tool failed, and what specific human input is needed.
