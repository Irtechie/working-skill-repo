# working-skill-repo

Working repository for the minimal shared KB skill set.

## Included Skills

The repo intentionally carries only the skills needed for the KB pipeline and
its direct skill-to-skill dependencies.

Core KB workflow:

- `kb-route`
- `kb-map`
- `kb-map-bootstrap`
- `kb-compact`
- `kb-check`
- `kb-functional-test`
- `kb-gate`
- `kb-fix`
- `kb-research`
- `kb-brainstorm`
- `kb-plan`
- `kb-work`
- `kb-complete`
- `kb-qa`
- `kb-repair`
- `kb-first-principles`
- `kb-epic`
- `kb-ship`
- `klfg`

You should not have to pick these manually. In normal use, ask for the work in
plain language and let `kb-route` choose the lane:

- Small bug or narrow fix -> `kb-fix`
- Bounded feature/refactor -> `kb-brainstorm` or `kb-plan`
- Large initiative, migration, or rewrite -> `kb-epic`
- Existing manifest -> `kb-work`
- Finished work -> `kb-complete`
- Release/PR/deploy readiness -> `kb-ship`

Standalone phase skills stop at their artifact boundary:

- `kb-brainstorm` writes/reviews requirements, then recommends `kb-plan`.
- `kb-plan` writes the manifest and slice plans, then recommends `kb-work`.
- `kb-work` executes the manifest, then automatically runs `kb-complete` after all slices are done or intentionally skipped.
- `klfg` is the only default auto-chain from idea to done.

Required dependencies:

- `document-review` - called by `kb-brainstorm`
- `tdd` - used by `kb-plan` / `kb-work` verification modes
- `ce-review` - called by `kb-complete`
- `ce-compound` - called by `kb-complete`
- `learn` - called by `kb-complete`
- `evolve` - called by `kb-complete`
- `todo-create` - called by `ce-review` when residual review work is externalized
- `todo-triage` - called by `todo-create` for interactive approval
- `ce-compound-refresh` - conditionally called by `ce-compound` when new learnings make older docs stale

## Intentionally Not Bundled

These are mentioned by the copied docs but are not required for the normal KB
pipeline:

- Upstream `deepen-*` enhancement passes are not bundled; this repo uses `kb-research` and proportional planning instead.
- `ce-ideate` is an upstream input option for brainstorming.
- `land` is a separate deliberate shipping step.
- `todo-resolve` is a follow-up implementation workflow after todo triage.
- `agent-browser` is a CLI/browser tool option referenced by `kb-qa`, not a skill dependency.

## Repository Instructions

This repo includes both `AGENTS.md` and `.github/copilot-instructions.md`.
They are intentionally short: route KB work to `kb-route`, point fresh sessions at
local memory files, and enforce the "every token must pay rent" rule without
installing a bulky persona.

## Layout

Skills live under `.github/skills/` so a repo-local agent can discover them
using the standard project skill location.

Agent roles live under `.github/agents/`. They are bundled because the KB and CE
flows depend on parallel research, security, adversarial, and review personas.
They are lazy-loaded by the host agent system; normal startup should still begin
with `kb-route`, not with every agent file.

Reusable scripts live inside their owning skills. `kb-check` uses
`.github/skills/kb-check/scripts/kb-check.ps1` to discover common
lint/test/build commands so agents can run deterministic checks instead of
spending tokens on manual inspection.

## Skill Quality Standard

KB skills should be structured, not brain dumps:

- Frontmatter says when to use the skill.
- The body starts with the job and the non-goals.
- Workflows are split into sections with clear gates.
- Required outputs and file locations are explicit.
- Questions are blocking-decision driven, not quota driven.
- Tool names stay generic unless the tool is bundled or guaranteed.
- Long shared doctrine lives in one skill or instruction file, then other skills reference it.

## Modern Model Assumptions

These skills assume current frontier models are competent. Do not spend tokens on
basic programming explanations, generic motivational text, repeated reminders to
be careful, or long examples for obvious patterns. Keep:

- trigger rules;
- file contracts;
- hard gates;
- escalation thresholds;
- exact commands and paths;
- output formats;
- failure handling;
- review/test criteria that prevent real mistakes.

Everything else should either move to a lazily loaded agent/reference/script or
be cut by `kb-compact`.

## Credits

This repo is primarily based on the ATV / All The Vibes skill set and its
Compound Engineering workflow. It also borrows useful ideas from
[Matt Pocock's skills](https://github.com/mattpocock/skills), especially small
composable skills and vertical slicing, and from
[G-Stack](https://github.com/garrytan/gstack), especially persistent workflow
memory, QA ownership, and operating-system-style orchestration. The compactness
rules are also informed by
[kevin-copilot](https://github.com/shyamsridhar123/kevin-copilot), especially
Copilot-first instruction surfaces and measured terse-output modes.

## KB Project Memory Files

The KB workflow uses repo-root markdown files for local memory instead of
keeping long-running chat sessions alive:

- `todo.md` - live execution board for active work, blockers, parked work, and handoff pointers
- `todo-done.md` - compact archive of completed work with links to details
- `docs/handoffs/active/` - resumable handoff files
- `docs/handoffs/parked/` - valuable but not currently runnable handoffs
- `docs/handoffs/done/` - completed or superseded handoffs
- `docs/context/PROJECT.md` - project route map for fresh sessions

This naming replaces older `docs/kanban.md`, `docs/kanban-done.md`,
`kb.md`, `kb-done.md`, and ad-hoc `*handoff.md` usage in the KB workflow.

## Memory Lifecycle

Fresh sessions should not need a project tour.

`kb-route`, `AGENTS.md`, and `.github/copilot-instructions.md` all enforce the
same preflight:

- Missing `todo.md` or `docs/context/PROJECT.md` means run `kb-map-bootstrap`.
- Missing context or handoff directories means run `kb-map refresh`.
- Do not ask first unless a non-empty user file would be overwritten.

During execution, `kb-work` classifies each slice as `memory-impact: none`,
`operational`, or `durable`. Durable changes refresh `docs/context/*` before
the work is treated as complete. `kb-complete` has the same refresh gate as a
backstop after review fixes.
