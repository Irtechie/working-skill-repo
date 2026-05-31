# Workflow Shape Routing Requirements

Status: draft
Created: 2026-05-31
Epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`

## Problem

The model needs to decide whether a request is just a skill edit, a normal KB
task, or a large pipeline with multiple streams. Without an explicit routing
rule, it can under-scope multi-stream work as a single file edit, or over-scope
small work into unnecessary ceremony.

## Decision

Add workflow-shape routing as a base-router behavior first, not a standalone
skill by default. The landmine is route selection, so the first home should be
`kb-start` and `kb-epic` routing criteria, backed by route fixtures.

A standalone skill is justified only if the decision tree becomes large enough
that keeping it in `kb-start` would bloat the base layer.

## Routing Categories

| Shape | Use When | Likely Route |
|---|---|---|
| Single skill edit | One skill file, no sync/pipeline consequences | `kb-task` or direct edit with `kb-check` |
| Skill-bundle change | Skill edit plus sync, drift checks, docs, or evals | `kb-plan` |
| Pipeline change | Multiple skills/scripts/evals/docs must change together | `kb-epic` |
| Multi-stream initiative | Independent streams with blockers, brainstorms, and manifests | `kb-epic` through all planning |
| Ambiguous idea | User is exploring scope or tradeoffs | `kb-brainstorm` before planning |

## Pipeline-Worthiness Signals

A request is probably pipeline-sized when it includes any of:

- multiple owning surfaces: skills plus scripts plus evals plus docs;
- cross-runtime behavior: Codex, Copilot, agents, ATV scaffold/plugin;
- proof harness changes;
- propagation/sync rules;
- several independent workstreams that can be planned separately;
- human checkpoints that block different streams;
- loaded-surface or token-budget measurement;
- route/eval fixtures needed before deletion or merge.

## Skill-Worthiness Signals

A new standalone skill earns its place only when:

- the behavior is triggered often enough to need discovery;
- the model would make a specific high-cost mistake without it;
- the instruction is too large or specialized for `kb-start`;
- route fixtures can distinguish it from existing lanes;
- it has a clear owner and proof gate.

If those are not true, prefer a short `kb-start` rule, a `kb-epic` section, or a
repo-local memory note.

## Requirements

- `kb-start` should ask: is this a file edit, skill-bundle change, pipeline
  change, or multi-stream initiative?
- `kb-epic` should front-load blockers and questions, then continue through all
  brainstorms and plans.
- Route fixtures should include examples that prove small skill edits do not
  become epics and multi-stream pipeline changes do not become single tasks.
- Planning manifests should record the chosen shape and why.

## Resolve Before Planning

- Decide whether workflow-shape routing stays inside `kb-start`/`kb-epic` or
  becomes a tiny lazy skill after measurement.
- Define the minimum route fixtures for skill-edit vs pipeline vs multi-stream
  work.
- Decide whether manifests need a required `workflow_shape` field.

## Slice Candidates

- Add route-complexity fixtures for skill edit, skill-bundle change, and
  multi-stream pipeline work.
- Add a compact workflow-shape check to `kb-start`.
- Add a `workflow_shape` field to kb-plan/kb-epic manifests if useful.
- Update `kb-epic` to continue through all planning for multi-stream initiatives.
