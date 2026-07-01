# KB-Native, Component-Scoped Learning (de-atv)

Status: active
Created: 2026-07-01
Last updated: 2026-07-01

## Objective

Make the KB skill bundle self-contained (no ATV dependency) and add
component-scoped learning so knowledge attaches to the exact component that needs
it, not only a global project bucket.

## Done Criteria

- No skill or harness hard-depends on `.atv/`; durable knowledge lives under
  `docs/context/kb/`, ephemeral run state under `.kb/`.
- Instinct schema supports a `scope` field + a component-local tier; `/learn`
  writes scoped lessons and `/evolve` can promote them into component-owned skills.
- Installed and source skills agree (no evolve path drift).
- `go test ./cmd/kbcheck/...` and `kbcheck core` pass; docs updated.

## Terminal Proof

- `go test ./cmd/kbcheck/...` green after harness path rename.
- `kbcheck core` green; `git diff --check` clean.
- grep shows zero remaining hard `.atv/` skill references (except historical
  provenance notes).
- A scoped instinct round-trips: written by learn to
  `docs/context/kb/instincts/scoped/<scope>.yaml`, read back when that scope runs.

## Current State

- Current artifact: docs/plans/2026-07-01-010-kb-native-scoped-learning-manifest.md
- Next allowed action: kb-work docs/plans/2026-07-01-010-kb-native-scoped-learning-manifest.md
- Last proof: schema inspection (no component scope today); atv coupling quantified
  (37 skill refs + ~14 Go touchpoints).

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|
| Plan the refactor | kb-plan | manifest 010 + slices 011-016 | done | this ledger + manifest gate_ledger plan-to-work passed |
| Execute slices | kb-work | manifest 010 | pending | per-slice gates |

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|

## Notes

- Origin incident: fleet-eval face-ID enrolled the wrong Colin reference; the lesson
  could not attach to the image-comparer because learning is global-only. That
  motivated component-scoped learning.
- Downstream consumer (separate repo, not this plan): llmcommune fleet-eval image
  comparer gets its own `calibration.yaml` + labeled fixtures using the new
  component-scope tier.
- ATV observer hook not reintroduced; passive capture optional.
