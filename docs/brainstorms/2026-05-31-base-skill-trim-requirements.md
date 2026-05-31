# Base Skill Trim Requirements

Status: draft
Created: 2026-05-31
Epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`

## Problem

The base layer must stay strong enough to prevent common agent failures while
shrinking loaded surface. The base should not become a pile of generic advice.

## Decision

Trim base skills only after loaded-surface measurement exists. Preserve behavior
that prevents real failures:

- wrong project root;
- stale/global memory leakage;
- unverified claims;
- sycophantic reversal;
- skipped verification;
- over-routing or under-routing work.

## Scope

Base candidates:

- `kb-start`
- `kb-map`
- `kb-first-principles`
- `kb-check`

## Requirements

- Keep `kb-map` fast-session recovery intact.
- Keep `kb-first-principles` available outside startup; it is a mid-conversation
  correction tool.
- Keep `kb-check` as the proof contract unless the base-layer decision moves it
  to verification-only loading.
- Keep `kb-start` as a thin router, not a knowledge store.
- Remove repeated examples, generic advice, and duplicated phase rules only when
  route/eval coverage protects behavior.

## Resolve Before Planning

- Decide the base-layer status of `kb-check`.
- Decide whether `kb-first-principles` remains standalone with a tiny pointer in
  `kb-start` or is partially embedded.

## Slice Candidates

- Add loaded-surface measurement for base-route startup.
- Trim `kb-first-principles` examples while preserving enforcement rules.
- Add a concise `kb-start` pointer to `kb-first-principles`.
- Extract bulky `kb-map` bootstrap/reference details into lazy docs if evals
  still pass.
