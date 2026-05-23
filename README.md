# working-skill-repo

Working repository for the minimal shared KB skill set.

## Included Skills

The repo intentionally carries only the skills needed for the KB pipeline and
its direct skill-to-skill dependencies.

Core KB workflow:

- `kb-brainstorm`
- `kb-plan`
- `kb-work`
- `kb-complete`
- `kb-qa`
- `kb-repair`
- `kb-first-principles`
- `klfg`

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

- `deepen-brainstorm` and `deepen-plan` are optional enhancement passes.
- `ce-ideate` is an upstream input option for brainstorming.
- `land` is a separate deliberate shipping step.
- `todo-resolve` is a follow-up implementation workflow after todo triage.
- `agent-browser` is a CLI/browser tool option referenced by `kb-qa`, not a skill dependency.

## Layout

Skills live under `.github/skills/` so a repo-local agent can discover them
using the standard project skill location.
