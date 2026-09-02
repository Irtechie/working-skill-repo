# Agent Instructions

For KB workflow requests, start with `kb-start`, unless the user explicitly invokes `kb-task` or asks for a first-principles task runner that should continue until done.

On every fresh session or ambiguous work request, let `kb-map` perform the KB memory preflight:

- Run `kb-map lookup <request>` before routing work.
- `kb-map` must resolve the active project root first and read memory from that repo only.
- If `todo.md` or `docs/context/PROJECT.md` is missing, `kb-map` invokes `kb-map-bootstrap`.
- If context or handoff folders are partial, `kb-map` refreshes or creates the missing structure.
- Do not ask the user to confirm bootstrap or refresh unless the operation would overwrite non-empty user files.

This repo is the portable skill bundle. Do not bootstrap consuming-project
memory or create project-work handoffs here by accident. If the user is trying
to hand off work from another project, switch to that project root or ask for
its path. Maintenance work may create the normal KB memory files here, but
they are local, Git-ignored state and must not be added to the public bundle.

## Skill Sync Workflow

When changing skills in this repo, treat `<working-skill-repo>` as the working bundle source, but check for newer drift before overwriting anything.

1. Compare the target skill across:
   - `<working-skill-repo>/.github/skills/<skill>/`
   - `~/.copilot/skills/<skill>/`
   - `~/.agents/skills/<skill>/`
   - `~/.codex/skills/<skill>/`
2. If a global copy differs, review the diff before copying over it. Newer useful work found only in a global install must be merged back into this repo first, not discarded.
3. After editing this repo, sync the final approved copy to:
   - Codex global: `~/.codex/skills/<skill>/`
   - Copilot global: `~/.copilot/skills/<skill>/`
   - shared agents global: `~/.agents/skills/<skill>/`
4. Update `README.md` in this repo when the visible workflow, installed-skill list, install commands, or repo hygiene contract changes.
5. Verify copied `SKILL.md` files with hashes and run `git diff --check` here.
6. Commit and push this repo when requested.

ATV repositories are not sync, release, or delivery targets. Do not inspect,
modify, commit, push, or gate on a neighboring ATV checkout unless the user
explicitly starts a separate ATV task.

For repo-local contributor quality, run:

```shell
go run ./cmd/kbcheck core
```

Before syncing or propagating skills, run the release/sync gate:

```shell
go run ./cmd/kbcheck local-release
```

`core` is contributor-safe on a fresh clone and validates repo-local deterministic checks. `local-release` composes `core`, `git diff --check`, static reports, and blocking read-only sync drift reports using `config/skill-quality.json`. Required targets are Codex global, Copilot global, and shared agents global.

Do not remove `kb-review`, `ce-compound`, or `ce-compound-refresh` from this
bundle unless the skills that invoke them are rewritten first. `kb-review` is
the single code-review owner for KB completion and standalone bundle review.

Optimize for comprehension and decision effort, not the fewest words. Every
token must pay rent:

- Apply this policy at user-facing return boundaries: completion, meaningful
  status, a blocker, a decision, and a pull-request first screen. Internal
  reasoning, tool calls, and subagent exchanges may stay detailed enough to
  finish the work.
- Put practical meaning first and the exact technical name second. For example,
  say what a cache is doing before naming its implementation or version.
- When control ownership could be unclear, label one state:
  - **Done**: the requested endpoint is reached; no response is needed.
  - **Agent continues**: safe work remains and the agent owns the next action.
  - **You need to decide**: only the user can unblock the named work or choose
    its disposition after safe agent-owned repair is exhausted.
- No preamble or closing filler.
- Do not restate the user's request.
- Lead with the answer, route, command, or code.
- For active work, lead with **Agent continues** plus only the changed outcome,
  next action, or blocker when that improves orientation; do not repeat
  unchanged state every turn.
- Keep the primary response surface to five ranked items. Put optional depth
  under a named `Details` or `Later` section instead of hiding it.
- Give time estimates only when grounded in observed work or a known wait.
  Never invent an estimate, and do not force a user action when work is done.
- Keep exact paths, commands, errors, decisions, risks, and safety warnings.
- Use longer explanations only when they change the decision or reduce rework.
- Match the format to the information: use plain sentences for a simple answer,
  ranked bullets for independent points, a table for repeated-field comparison,
  a real screenshot for a visible UI change, and one Mermaid workflow diagram
  for sequence, dependency, branching, ownership, or state change. When those
  are insufficient, use `interactive-workflow-workbench-light` for one bounded
  clickable path and the full interactive workbench only for epic, multi-path,
  deep-evidence, or client/showcase work. A visual must remove mental
  reconstruction, not decorate.
- If an optional visual capability is unavailable, fall back to the best static
  format and continue. Presentation depth never blocks proven core work.
- Before asking a question, classify it:
  - **hard response required**: only the user can authorize, supply, or decide
    it and dependent work cannot safely continue;
  - **soft preference**: the agent has a safe, reversible default and continues
    unless the user overrides it;
  - **no response needed**: status, proof, completion, or an agent-owned choice.
- For a hard question, state the exact ask, why the user must answer, what is blocked, and the recommended option.
- For a soft preference, state the default and handle it without blocking. Never turn an agent-owned decision into review work for the user.
- Treat `paused` as execution control, not a technical gate result. A plain
  pause stops mutation without writing a handoff or converting the task to
  blocked; `pause and handoff` permits only that bounded state write.
- Before repeating a blocker, rerun its named sensor or the cheapest owning
  probe. Keep test, code, controller, UI, and reproducibility failures
  agent-owned while safe repair remains. Reserve `human-required` for authority,
  access, private input, irreversible risk, or subjective judgment.
- Propagate a blocker only through its real dependents. Release, deployment,
  signing, and optional-capability failures do not erase proven implementation
  completion.
- Use plain human language. Define unavoidable jargon once and keep the
  important decision on the primary response surface.
- Keep stable policy in ambient instructions and volatile task state in
  `todo.md`, plans, or handoffs so prompt prefixes stay reusable.
- Move deterministic data gathering outside the reasoning loop when practical:
  prefetch with repo-native CLI/search commands, then pass only needed paths,
  fields, or compact output to the agent.
- Do not register broad MCP/tool catalogs in repo config. Prefer built-in
  file/search/CLI tools and enable optional tools only when a task needs them.

Use these project memory files. Consuming projects decide whether to track
them; this skill-bundle source repo keeps them local and Git-ignored:

- `todo.md` for active work, blockers, parked work, and handoff pointers.
- `todo-done.md` for completed-work summaries.
- `docs/context/PROJECT.md` for the project route map.
- `docs/context/eval-map.md` for repo-native eval surfaces and canonical proof commands.
- `docs/solutions/` for documented solutions to past workflow, tooling, and implementation problems; entries use searchable frontmatter and are relevant when implementing or debugging in documented areas.
- `docs/handoffs/active/`, `docs/handoffs/parked/`, and `docs/handoffs/done/` for handoff lifecycle.

Do not treat these files as skills. Skills live under `.github/skills/`.

## Learning Model

Instincts are stored at the narrowest scope that owns them. Durable instinct
files live in `docs/context/kb/` and are normally git-tracked by consuming
projects; this bundle source keeps its own maintenance instincts ignored.
Ephemeral artifacts live in `.kb/` (git-ignored).

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

## Session-End Durability

Durability is not delivery. Conflating them is why sessions end with work
stranded on disk, invisible to every downstream tool.

Delivery is a separate, policy-owned step. Under `kb-complete`, reviewed work
defaults to `mode: pr`/`merge: manual`: it is committed, pushed, and opened as
a PR, and `kb-complete` then asks once whether to sync or wait for review.
Merging a PR, direct-default integration, and post-merge sync always require
explicit authorization. Ad-hoc session end is not `kb-complete` and never
delivers.

Durability is agent-owned. Before a session ends with a dirty tree, run:

```shell
go run ./cmd/kbcheck session-preserve --action apply --session-id <session-id> --json
```

This makes one WIP commit on the session's own branch. It never pushes, never
merges, never claims completion, and is reversible with `git reset`.

The gate fails closed and refuses on: the resolved default branch, a detached
HEAD, a branch that does not match `--branch`, an in-progress
merge/rebase/cherry-pick/revert/bisect, or a missing `--session-id`. A clean
tree is a no-op, never an empty commit.

Compiled build outputs and files over 5 MB are excluded and reported in
`excluded[]` rather than committed or silently dropped. Gitignored paths are
never preserved. Use `--action plan` to forecast without mutating.

A preserved commit is not evidence of completion. Do not cite it as proof, and
do not let it satisfy a proof gate.

## Optional Context Providers

MCP search, vector indexes, and similar tools are optional adapters. Do not
commit or auto-start their
hooks/configs, and do not require a daemon/app for skills, install, sync, or
checks. The file-native `rg`/`kb-map`/`kbcheck` path must keep working.

Phoenix is credited prior art whose useful proof/routing mechanics have been
absorbed into KB. Keep research and attribution, but do not add a Phoenix
runtime, MCP server, daemon, or required install.
