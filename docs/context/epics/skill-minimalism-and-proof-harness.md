# Skill Minimalism and Proof Harness

Status: reviewed
Created: 2026-05-31
Last refreshed: 2026-05-31

## Intent

Reduce loaded skill surface without losing the user's core workflow:
fast repo recovery, first-principles honesty, deterministic verification, and
drift-safe propagation.

The principle is separation of concerns:

- global base skills encode behavior and enforcement;
- repo-local memory records project-specific evidence;
- lazy workflow skills load only when their lane is needed.

## Success Criteria

- `kb-map`, `kb-first-principles`, `kb-check`, and `kb-start` remain available
  as the base layer.
- Every workstream reaches a planning-complete state: manifest created, parked
  with rationale, or blocked by an explicit human decision.
- The eval harness proves freshness with `eval_run_id` and proves verifier
  integrity with protected-file SHA checks.
- Regression baselines persist and fail the gate on meaningful regressions.
- Skill deletion or merge candidates are blocked by reference integrity, route
  fixtures, sync hashes, and eval proof.
- Loaded-surface reports show route-level line/token reduction before and after
  trims.

## Architecture Decisions

- Do not move repo-specific learning into global skills. Global learning skills
  should operate on repo-local evidence files.
- Do not delete `kb-first-principles`; trim it to the behavioral brake:
  verify checkable claims, stop before edits on factual pushback, avoid fake
  certainty, and avoid wholesale reversal.
- Keep `kb-first-principles` provisionally standalone while direct-chat vs
  vibe-coding behavior is clarified; if embedded later, preserve the same brake.
- Do not delete `kb-map`; preserve project-root resolution and local memory
  lookup. Trim or lazy-load only details that do not affect fast session
  recovery.
- Treat landmines as local evidence with an owner surface and fix condition.
  Promote them into the owning skill/doc/generator instruction only after proof
  and human approval; archive them when fixed and verified.
- Keep `learn`/`evolve` numeric maturity gates, but add a human approval gate
  before generated `learned-*` skills are promoted or synced from this bundle.
- Treat `todo-*` as lazy lanes unless route fixtures prove they need to be
  globally available.
- Preserve TDD's anti-cheat landmine inside the pipeline: define the behavior
  oracle before implementation when practical, prove RED, protect the
  test/fixture with a SHA/manifest, then implement.
- Treat SHA checks as proof that the verifier ran against the intended files,
  not as a substitute for behavior checks.
- Treat root `todo.md` as the KB live backlog/execution board during planning.
  Do not add `backlog.md` unless `todo.md` fails under real load.

## Research

- Existing audit: `docs/context/research/2026-05-29-skill-repo-gap-audit.md`
- Eval map: `docs/context/eval-map.md`
- Testing map: `docs/context/operations/testing.md`
- Architecture deepening comparison:
  `docs/context/research/2026-05-31-architecture-deepening-vs-deslop-thermo.md`
- Distribution research:
  `docs/context/research/2026-05-30-agent-skills-git-distribution.md`

## Workstreams

| Workstream | Brainstorm | Manifest | Status | Notes |
|---|---|---|---|---|
| Harness proof hardening | `skipped-clear` | `docs/plans/2026-05-31-000-kb-proof-pipeline-spike-manifest.md` | reviewed | Persisted baselines, protected SHA manifests, and coded pipeline spike. |
| Base-layer contract | `docs/brainstorms/2026-05-31-base-layer-contract-requirements.md` | `docs/plans/2026-05-31-020-kb-routing-trim-manifest.md` | reviewed | Base kept narrow; first-principles trimmed; surface reporting added. |
| Repo-local learning model | `docs/brainstorms/2026-05-31-repo-local-learning-landmines-requirements.md` | `docs/plans/2026-05-31-010-kb-learning-landmines-manifest.md` | reviewed | Landmine schema, learn fields, evolve approval, and KB loading added. |
| Loaded-surface measurement | `skipped-clear` | `docs/plans/2026-05-31-020-kb-routing-trim-manifest.md` | reviewed | Route -> skills -> lines/token estimate -> SHA report implemented. |
| Reference graph cleanup | `skipped-clear` | `docs/plans/2026-05-31-020-kb-routing-trim-manifest.md` | reviewed | Legacy refs fixed; scanner catches unknown local skill references. |
| Architecture deepening lane | `docs/brainstorms/2026-05-31-architecture-deepening-lane-requirements.md` | `docs/plans/2026-05-31-030-kb-lazy-lane-consolidation-manifest.md` | reviewed | Added compact lazy lane distinct from cleanup and diff review. |
| Workflow-shape routing | `docs/brainstorms/2026-05-31-workflow-shape-routing-requirements.md` | `docs/plans/2026-05-31-020-kb-routing-trim-manifest.md` | reviewed | `kb-start` classifies direct edit, skill bundle, pipeline, and epic shapes. |
| Base skill trim | `docs/brainstorms/2026-05-31-base-skill-trim-requirements.md` | `docs/plans/2026-05-31-020-kb-routing-trim-manifest.md` | reviewed | Trimmed `kb-first-principles`; left other base skills intact behind measurements. |
| Core workflow trim | `docs/brainstorms/2026-05-31-core-workflow-trim-requirements.md` | `docs/plans/2026-05-31-020-kb-routing-trim-manifest.md` | reviewed | Preserved gates; deferred larger extraction until measured separately. |
| Narrow lane trim | `docs/brainstorms/2026-05-31-narrow-lane-trim-requirements.md` | `docs/plans/2026-05-31-030-kb-lazy-lane-consolidation-manifest.md` | reviewed | Kept narrow lanes lazy where distinct proof value remains. |
| Questionable global skill trim | `docs/brainstorms/2026-05-31-questionable-global-skill-trim-requirements.md` | `docs/plans/2026-05-31-030-kb-lazy-lane-consolidation-manifest.md` | reviewed | TDD and todo lanes trimmed into compatibility/lazy roles. |
| Propagation policy | blank | `docs/plans/2026-05-31-030-kb-lazy-lane-consolidation-manifest.md` | reviewed | Required globals/ATV `.github` sync; ATV scaffold/plugin remain optional thin bundles. |

## Dependency Map

```text
Harness proof hardening
  -> Loaded-surface measurement
  -> Skill trim passes

Base-layer contract
  -> Skill trim passes

Repo-local learning model
  -> Skill trim passes

Reference graph cleanup
  -> Skill trim passes

Architecture deepening lane
  -> Skill trim passes

Workflow-shape routing
  -> Base-layer contract
  -> Skill trim passes

Propagation policy
  -> Final sync/ship decisions
```

## Completion Review

- Review: P0=0 P1=1(resolved locally) P2=0 P3=0.
- Resolved finding: skill-eval baseline comparison now fails negative fixtures
  that incorrectly start passing.
- Proof: `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`,
  `git diff --check`, and `scripts\skill-sync-report.ps1` passed with 0
  required sync issues.
- Review limitation: subagent review was not spawned because this session's
  delegation tool requires explicit user authorization for subagents; review was
  performed locally.

## Dark Factory Queue

Complete. Future work belongs in Parked / Blocked or a new epic.

## Human Checkpoints

Planning blockers to ask first:

1. For direct chat vs vibe-coding, should `kb-first-principles` remain the
   standalone trust brake, or should `kb-start` embed a shorter trigger?
2. Should `kb-check` stay in the always-available base layer, or load only when
   work reaches verification?
3. Should active landmines live in `docs/context/landmines.md` with owner/fix
   metadata, or in `.atv/instincts/project.yaml` with a landmine type?
4. What exact human approval prompt should `evolve` use before promoting or
   syncing `learned-*` skills out of this portable bundle?
5. Should resolved landmines archive immediately when the owning surface is
   fixed and verified, with time-decay only for unfixed stale entries?
6. Confirm architecture deepening as a compact new lazy skill once route
   fixtures prove it is distinct from deslop and thermo-nuclear review.
7. Confirm how protected test-first oracles should appear in `kb-plan`/`kb-work`,
   then decide whether standalone `tdd` can be deleted once references are clean.
8. Confirm `todo-triage`/`todo-create` should merge around root `todo.md` as
   the KB live backlog/board, with no `backlog.md` unless volume proves it.
9. Should workflow-shape routing stay inside `kb-start`/`kb-epic`, or become a
   tiny lazy skill only if measurement shows it would bloat the base router?

Planning can continue with assumptions for non-trim infrastructure:

- Harness proof hardening assumes existing eval harness ownership remains.
- Loaded-surface measurement assumes route-to-skill usage can be approximated
  from configured skill files and explicit references.
- Reference graph cleanup assumes missing/legacy refs should be corrected to
  current skill names or prose.

Resolved execution/ship decision:

- ATV scaffold/plugin remain thinner optional bundles by default; required sync
  targets are Codex global, Copilot global, shared agents global, and ATV
  `.github/skills`.

## Parked / Blocked

- Skill deletion remains blocked until loaded-surface measurement and reference
  integrity are in place.

## Completion Criteria

- Every workstream has a manifest or is explicitly parked with rationale.
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All` exits 0.
- `git diff --check` exits 0 in every touched repo.
- Required sync targets report zero required issues.
- Token/line reduction is measured, not estimated from memory.
