# Questionable Global Skill Trim Requirements

Status: draft
Created: 2026-05-31
Epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`

## Problem

Some skills may be useful in the abstract but costly as global or portable-bundle
surface. The right question is not whether the skill is well written; it is
whether the model would get something important wrong without it.

## Decision

Evaluate questionable global skills by landmine value and call graph impact.
Do not delete skills that are still invoked by core workflows until callers are
rewritten and evals pass.

Treat protected test-first oracles as pipeline behavior in
`kb-plan`/`kb-work`/`kb-complete`, not as a reason to keep a standalone global
`tdd` skill. The landmine is not generic red/green/refactor advice. The
landmine is anti-cheat: if expected behavior is captured in a test, fixture,
scorer, snapshot, or contract before implementation, and that proof file is
protected by SHA/checksum, the model cannot quietly rewrite the oracle after
seeing the implementation fail.

The useful `tdd` content should be absorbed into the pipeline: behavior through
public interfaces, one vertical test cycle at a time when practical, avoiding
horizontal "write all tests first" bulk planning, proving RED when possible,
and protecting the test oracle before implementation.

Treat `todo-*` as lazy queue-management behavior. It should not be loaded unless
the user is creating, triaging, or maintaining durable work items.

Current landmine: KB planning already writes the active execution board to
root `todo.md`, but `todo-create` still describes an older
`.context/compound-engineering/todos/` file-per-todo system. The repo should not
add `backlog.md` until this split is resolved. For KB workflows, `todo.md`
should remain the live backlog/board unless volume proves a separate backlog
file is needed.

## Scope

Questionable/global-value candidates:

- `tdd`
- `todo-triage`
- `learn`
- `evolve`
- future `learned-*` skills
- any generic workflow skill that mostly repeats model defaults

## Requirements

- Preserve the actual `learn -> instincts -> evolve -> learned-*` mechanism
  unless measurement proves it is too costly.
- Add explicit approval before this portable bundle promotes or syncs generated
  `learned-*` skills.
- Keep `evolve` numeric gates as minimum maturity checks: confidence greater
  than 0.85, more than five observations, and last seen within 90 days. Human
  approval is still required before portable-bundle promotion or sync.
- Move essential `tdd` landmines into planning/work verification guidance:
  define the behavior oracle before implementation when practical, prove RED,
  protect the oracle with SHA/manifest, then implement.
- Treat standalone `tdd` as a compatibility/deletion question after callers and
  route fixtures are updated. Do not keep it globally loaded for generic advice.
- Merge or collapse `todo-triage` into a single lazy todo lane if reference
  scans and callers can be updated without losing behavior.
- Align todo skills with KB's root `todo.md` model, or explicitly mark the
  legacy file-per-todo system as non-KB/CE-only. Do not introduce `backlog.md`
  as a third queue without evidence that `todo.md` is failing.
- Use reference scanner warnings as deletion blockers.

## Resolve Before Planning

- Define the exact approval prompt for `evolve` generated skills.
- Decide how `kb-plan`/`kb-work` represent protected test-first oracles.
- Decide whether standalone `tdd` can be deleted after its anti-cheat behavior
  moves into the pipeline and reference scans are clean.
- Decide the merged shape for `todo-create`/`todo-triage`, including whether
  the root `todo.md` board fully replaces the legacy file-per-todo system for
  KB workflows.

## Slice Candidates

- Add reference integrity tests for deleted/merged skill names.
- Add an `evolve` promotion guard for portable-bundle sync.
- Audit `tdd` call sites and move anti-cheat behavior into `kb-plan`/`kb-work`;
  delete or park standalone `tdd` only after route fixtures and references are
  clean.
- Audit `todo-triage`/`todo-create` call sites and merge around the root
  `todo.md` active-board model.
