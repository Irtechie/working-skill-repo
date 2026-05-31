# Core Workflow Trim Requirements

Status: draft
Created: 2026-05-31
Epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`

## Problem

Core workflow skills encode valuable gates, but several are long and partially
overlapping. Trimming them carelessly risks losing the workflow that makes the
bundle useful.

## Decision

Treat core workflow skills as lazy lanes. Keep the gates and artifact contracts;
trim generic explanation and duplicated routing/ceremony.

## Scope

Core workflow candidates:

- `kb-brainstorm`
- `kb-plan`
- `kb-work`
- `kb-complete`
- `kb-epic`
- `klfg`

## Requirements

- Preserve brainstorm -> plan -> work -> complete phase boundaries.
- Preserve HITL rules and planning-complete behavior for epics.
- Preserve vertical-slice manifests and per-slice verification fields.
- Preserve `kb-work` scope ledger, repair loop, and completion handoff.
- Preserve `kb-complete` review, proof, learning, memory refresh, and cleanup
  gates.
- Reduce repeated route tables and repeated generic instruction text.

## Resolve Before Planning

- Decide whether `kb-complete` should still always run `learn` and conditional
  `evolve`, or whether this is gated by evidence/changed surface.
- Decide whether `klfg` should remain a full orchestrator or become a thin alias
  over `kb-epic`/`kb-start`.

## Slice Candidates

- Add loaded-surface report for each core workflow route.
- Fix known `kb-brainstorm` missing `/ce-ideate` reference.
- Extract repeated templates from `kb-plan` and `kb-work` into lazy references.
- Trim `kb-complete` learning/evolve prose after preserving exact gates.
- Update route fixtures when any route behavior changes.
