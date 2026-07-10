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

When changing skills in this repo, treat `<working-skill-repo>` as the working bundle source, but check for newer drift before overwriting anything.

1. Compare the target skill across:
   - `<working-skill-repo>/.github/skills/<skill>/`
   - `<atv-repo>/.github/skills/<skill>/`
   - `<atv-repo>/pkg/scaffold/templates/skills/<skill>/`
   - `<atv-repo>/plugins/atv-everything/skills/<skill>/`
   - `~/.copilot/skills/<skill>/`
   - `~/.agents/skills/<skill>/`
   - `~/.codex/skills/<skill>/`
2. If a global or ATV copy differs, review the diff before copying over it. Newer useful work found only in a global install must be merged back into this repo first, not discarded.
3. After editing this repo, sync the final approved copy to:
   - Codex global: `~/.codex/skills/<skill>/`
   - Copilot global: `~/.copilot/skills/<skill>/`
   - shared agents global: `~/.agents/skills/<skill>/`
   - ATV fork: `<atv-repo>/.github/skills/<skill>/`
   - ATV scaffold/plugin copies only when that skill is intentionally shipped there.
4. Update `README.md` in this repo when the visible workflow, installed-skill list, install commands, or repo hygiene contract changes.
5. Update `<atv-repo>\README.md` when the ATV-facing workflow, bundled skills, or scaffold/plugin behavior changes.
6. Verify with hashes for copied `SKILL.md` files and `git diff --check` in every touched repo.
7. Commit and push both repos when requested or when the user asks for the full propagation flow.

For repo-local contributor quality, run:

```shell
go run ./cmd/kbcheck core
```

Before syncing or propagating skills, run the release/sync gate:

```shell
go run ./cmd/kbcheck local-release
```

`core` is contributor-safe on a fresh clone and validates repo-local deterministic checks. `local-release` composes `core`, `git diff --check`, static reports, and blocking read-only sync drift reports using `config/skill-quality.json`. Required targets are Codex global, Copilot global, shared agents global, and `<atv-repo>/.github/skills`. ATV scaffold/plugin targets are optional thin bundles; warnings there are acceptable unless the current change explicitly ships that skill surface.

Do not remove `kb-review`, `ce-review`, `ce-compound`, or `ce-compound-refresh` from this bundle unless the skills that invoke them are rewritten first. KB completion uses `kb-review`; `ce-review` remains the generalized CE review skill.

Every token must pay rent. Be concise by default:

- No preamble or closing filler.
- Do not restate the user's request.
- Lead with the answer, route, command, or code.
- Keep exact paths, commands, errors, decisions, risks, and safety warnings.
- Use longer explanations only when they change the decision or reduce rework.
- Keep stable policy in ambient instructions and volatile task state in
  `todo.md`, plans, or handoffs so prompt prefixes stay reusable.
- Move deterministic data gathering outside the reasoning loop when practical:
  prefetch with repo-native CLI/search commands, then pass only needed paths,
  fields, or compact output to the agent.
- Do not register broad MCP/tool catalogs in repo config. Prefer built-in
  file/search/CLI tools and enable optional tools only when a task needs them.

Use these project memory files:

- `todo.md` for active work, blockers, parked work, and handoff pointers.
- `todo-done.md` for completed-work summaries.
- `docs/context/PROJECT.md` for the project route map.
- `docs/context/eval-map.md` for repo-native eval surfaces and canonical proof commands.
- `docs/solutions/` for documented solutions to past workflow, tooling, and implementation problems; entries use searchable frontmatter and are relevant when implementing or debugging in documented areas.
- `docs/handoffs/active/`, `docs/handoffs/parked/`, and `docs/handoffs/done/` for handoff lifecycle.

Do not treat these files as skills. Skills live under `.github/skills/`.

## Learning Model

Instincts are stored at the narrowest scope that owns them. Durable instinct files live in `docs/context/kb/` (git-tracked); ephemeral artifacts live in `.kb/` (git-ignored).

Key paths:

- `docs/context/kb/instincts/project.yaml` — project-tier and global-tier instincts (tagged by `scope: project` or `scope: global`)
- `docs/context/kb/instincts/scoped/<scope-path>.yaml` — workflow/domain instincts (default home for new lessons)
- `docs/context/kb/kb-completions.txt` — kb-complete counter
- `.kb/observations.jsonl` — optional passive observation feed (ephemeral)

Scope hierarchy: `workflow/domain → project → global`. Default = narrowest owning scope. Pull = active scope + all ancestors; never siblings. Promotion = nearest common ancestor when the same lesson recurs independently across sibling scopes. Landmines = instant one-shot at owning scope.

**X pipeline's lessons are not visible to Y pipeline unless promoted to a shared ancestor.**

When running `learn` or recording an instinct, target the workflow/domain scope unless the lesson is demonstrably cross-workflow. Do not default to project or global.

When local memory is missing or badly stale, use `kb-map`; it decides whether lookup, refresh, or bootstrap is required. For normal startup, use `kb-start`.

## Agent-Owned Verification

Do not ask the user to test normal application behavior when the agent can test it.

For apps with a UI frontend, if a change touches frontend code or user-visible UI behavior, verify it through the rendered UI with Playwright, CDP, or the repo's browser transport. Use real navigation, clicks, inputs, and programmatic DOM assertions. Do not substitute backend calls, source inspection, screenshots alone, or prose claims.

Use unit/integration tests, CLI/API probes, browser automation, screenshots, traces, logs, and DOM assertions as needed. Screenshots are evidence, not the pass/fail oracle.

Only ask the user to test when verification requires something the agent truly cannot access: credentials or MFA/session access not already available, subjective product/design judgment, external hardware or production-only systems, destructive/risky real-world action, or missing test input that cannot be safely generated.

If blocked, state exactly what was attempted, what command/tool failed, and what specific human input is needed.

## Optional Context Providers

CCE is an owned, supported optional context adapter. MCP search, vector indexes,
and similar tools are optional adapters too. Do not commit or auto-start their
hooks/configs, and do not require a daemon/app for skills, install, sync, or
checks. The file-native `rg`/`kb-map`/`kbcheck` path must keep working.

Phoenix is credited prior art whose useful proof/routing mechanics have been
absorbed into KB. Keep research and attribution, but do not add a Phoenix
runtime, MCP server, daemon, or required install.
