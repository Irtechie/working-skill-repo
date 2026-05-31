# Base Layer Contract Requirements

Status: draft
Created: 2026-05-31
Epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`

## Problem

The skill bundle needs a small always-available base that keeps sessions
grounded without loading broad workflow text by default. Deleting too much loses
the behavior the user values; keeping too much burns context and creates noise.

## Decision

The base layer should contain mechanisms and behavioral enforcement, not
project-specific facts.

Keep in the base:

- `kb-map`: resolve the active project, load repo-local memory, and prevent
  stale/global-context leakage.
- `kb-first-principles`: force evidence-grounded conversation when the model is
  guessing, overclaiming, reversing too easily, or skipping code/research.
- `kb-check`: route verification through deterministic commands and proof
  artifacts.
- `kb-start`: keep a thin router that maps natural user requests to lazy lanes.

Everything else should be lazy-loaded by route or stored as repo-local evidence.

`kb-first-principles` remains provisionally standalone while direct-chat vs
vibe-coding behavior is still being clarified. The user needs an explicit
brake for moments where the model starts guessing, overclaiming, or skipping
code/research. If that behavior is later embedded elsewhere, the embedded form
must preserve the stop-before-edit and verify-checkable-claims behavior.

## Requirements

- Base skills must work in a fresh session with minimal context.
- Base skills must not carry project-specific facts globally.
- Base skills must prefer repo-local files over chat memory.
- Base skills must fail closed when deterministic proof is required.
- Base skills must not duplicate each other's long rules.
- Base skills must expose enough route information for the model to choose the
  next lane without loading every workflow skill.

## Landmine vs Generic Rule

A base-layer instruction earns its place only if it prevents a behavior the
model commonly gets wrong:

- wrong repo root;
- stale/global memory leakage;
- fabricated certainty;
- unverified factual claims;
- skipping tests;
- skipping required workflow phases;
- sycophantic reversal under pushback.

Generic engineering advice should not live in the base layer.

## Candidate Trims

| Skill | Keep | Trim Candidate |
|---|---|---|
| `kb-map` | project-root resolution, local memory lookup, bootstrap/refresh triggers | long explanatory prose and duplicate bootstrap detail if reference files can carry it |
| `kb-first-principles` | pushback classification, verify facts, stop-before-edit, no wholesale reversal | repeated examples and anti-patterns once rules are crisp |
| `kb-check` | command discovery, deterministic proof fields, functional-check escalation | generic testing advice if already covered by repo instructions |
| `kb-start` | map-first rule, route table, stale-work rule | repeated ceremony/context guidance |

## Resolve Before Planning

- Decide whether direct user chat should continue to rely on
  `kb-first-principles` as a standalone brake, or whether `kb-start` should
  carry a short embedded trigger that invokes the same behavior.
- Decide whether `kb-check` belongs in the base layer or is loaded only when work
  reaches verification.

## Slice Candidates

- Measure current base-layer line/token surface.
- Trim `kb-first-principles` without deleting enforcement rules.
- Trim `kb-start` route text after route fixtures cover the expected behavior.
- Extract bulky `kb-map` reference detail into lazy docs if evals still pass.
