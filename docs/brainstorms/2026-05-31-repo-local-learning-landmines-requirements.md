# Repo-Local Learning and Landmines Requirements

Status: draft
Created: 2026-05-31
Epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`

## Problem

The existing `learn`/`evolve` pipeline is already project-scoped in mechanism:
`learn` writes `.atv/instincts/project.yaml` with confidence, observation count,
and recency decay; `evolve` promotes mature instincts into
`.github/skills/learned-*`.

The remaining risk is promotion policy and loaded surface. A project-local
instinct is cheap and appropriate. A promoted learned skill becomes instruction
surface, and in this portable skill bundle it may also be synced outward. The
user still needs the system to remember traps that the model would otherwise
miss without turning generic observations into global skill bloat.

## Decision

Learning is the capture/scoring process. Evolve is the promotion process. A
landmine is a specific class of learned item: a high-risk repo-specific trap
that should change future behavior even if it has not appeared five times.

Keep `learn` and `evolve` as separate mechanisms unless a later audit proves the
split adds unnecessary loaded surface. Tighten what is eligible for promotion.

Landmines should not become a permanent silo by default. A landmine starts as
local evidence with an owner target: the skill, repo doc, generator instruction,
test fixture, or workflow that should eventually absorb the fix. If the
landmine is fixed in that owning surface and verified, the active landmine entry
should be resolved or archived immediately.

## Definitions

| Term | Meaning | Storage |
|---|---|---|
| Learning | Capturing reusable observations as scored instincts | `learn` mechanism, `.atv/instincts/project.yaml` output |
| Evolve | Promoting mature instincts into `learned-*` skills | `evolve` mechanism, `.github/skills/learned-*` output |
| Landmine | A high-risk repo-specific trap the model is likely to miss without explicit warning | local evidence, optionally an instinct or concise context entry |
| Skill | A reusable behavior or workflow the model should apply when loaded | `.github/skills/*` |
| Local memory | Project truth that should not leak into unrelated repos | `docs/context/*`, `todo.md`, handoffs, `.atv/instincts/*` |

## Landmine Criteria

A landmine should be recorded only when it meets at least one criterion:

- The model already made this mistake.
- The repo has a non-obvious convention that conflicts with common defaults.
- A command, sync path, runtime, auth mode, or generated artifact has a specific
  failure mode.
- A workflow has a gate the model is likely to skip.
- The trap is high-cost, destructive, or hard to notice from local code alone.

Do not record generic advice as a landmine.

Landmine scoring should not rely only on repetition. A single verified
high-severity failure may deserve a landmine entry even before it qualifies for
`evolve` promotion. Promotion into a skill should still require evidence that
the behavior should be loaded as instruction, not just remembered locally.

Every landmine should identify:

- owner surface: where the durable fix belongs;
- severity: why it is worth interrupting future work;
- evidence: the specific failure, command, file, or review finding;
- fix condition: what change retires the landmine;
- verification: the command, eval, or review check proving it is fixed.

## Storage Model

Preferred local surfaces:

- `docs/context/PROJECT.md`: route map and durable current truth.
- `docs/context/eval-map.md`: proof surfaces and canonical commands.
- `docs/context/operations/testing.md`: verification commands and expectations.
- `docs/context/research/*.md`: durable investigation notes.
- `docs/handoffs/*`: resumable work state.
- `todo.md`: active queue and blockers.

Potential dedicated surface:

- `docs/context/landmines.md` if landmines become numerous enough that mixing
  them into `PROJECT.md` makes startup noisy.

Preferred model:

- Active local landmines live in repo memory with owner metadata.
- Domain-specific landmines graduate into the owning skill or instruction only
  after evidence shows the model will otherwise repeat the mistake.
- Generic landmines are rejected or kept as non-loaded notes.
- Fixed landmines are archived with proof instead of kept active until decay.

## Skill Implications

- `learn` already acts as a local-evidence writer into `.atv/instincts`; audit
  it for landmine fields rather than replacing it.
- `evolve` should remain separate, but promotion must be stricter in this repo:
  generated `learned-*` skills should not be synced globally or treated as
  portable bundle skills without explicit human approval.
- `evolve` already uses numeric maturity gates: confidence greater than 0.85,
  more than five observations, and last seen within 90 days. Keep those as the
  automatic maturity filter, but require human approval before generated
  `learned-*` skills are committed or synced from this portable bundle.
- `learn` already decays stale instincts and archives entries below confidence
  0.3. Landmine resolution should be faster than decay: once the owning surface
  is fixed and verified, mark the landmine resolved immediately.
- `kb-map` should read local landmines during startup if they exist, but it
  should load only concise active traps.
- `kb-complete` should capture new landmines only when verified by evidence,
  not because the model had a vague lesson.

## Resolve Before Planning

- Decide whether to create `docs/context/landmines.md` or keep landmines inside
  `PROJECT.md` until volume justifies a separate file.
- Decide whether `learn` should add explicit landmine fields to
  `.atv/instincts/project.yaml`, or whether `docs/context/landmines.md` should
  be the active queue with links back to instincts.
- Decide the exact human prompt and proof fields for approving `evolve`
  promotion/sync from this portable bundle.
- Decide how `kb-map refresh` should distinguish resolved landmines from stale
  but unfixed landmines.

## Slice Candidates

- Add a landmine schema/checklist to local memory or instinct docs.
- Extend or constrain `learn` so landmine candidates include severity and
  evidence, not only repetition.
- Add an explicit human approval gate before `evolve` creates or syncs
  `learned-*` skills in this bundle.
- Add landmine owner/fix-condition fields and a resolved archive flow.
- Teach `kb-map` to load local landmines only when present.
- Add lint/eval checks that reject generic landmine entries.
