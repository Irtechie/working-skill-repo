# Session Model Routing

Status: active
Created: 2026-07-10
Last updated: 2026-07-10

## Objective

Ship a portable Codex-first advisory pilot that discovers callable model routes, selects workers by slice difficulty at work time, preserves proof across fallback, and degrades to the current model without blocking ordinary KB work.

## Done Criteria

- New plans freeze task difficulty and proof, never a concrete model route.
- `kbrouter` discovers and selects dispatch-proven Codex and configured OpenAI-compatible-via-Codex routes from secure user-local/project policy.
- `kb-work` previews, dispatches, falls back, and records route receipts while preserving independently proven work.
- The advisory pilot passes deterministic, live-canary, installer, sync, review, and release gates.
- Intentional changes are committed, pushed on a non-default branch, and delivered in an open PR.

## Terminal Proof

- `go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json`
- `go run ./cmd/kbcheck local-release`
- `complete-to-ship` gate passes with no unresolved P0/P1 findings.
- Topic branch and correctly based open PR exist with matching local/upstream SHA.

## Done Check

- Type: command_exit
- Check: `go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json`
- Expected result: exit code 0
- Why sufficient: validates model-neutral manifests, secure catalog/selection, route-bound dispatch receipts, graceful fallback, pilot support claims, and installer evidence.

## Current State

- Current artifact: `docs/plans/2026-07-10-030-kb-session-model-routing-manifest.md`
- Next allowed action: `kb-work docs/plans/2026-07-10-030-kb-session-model-routing-manifest.md`
- Last proof: brainstorm-to-plan gate passed in `docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md`

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|
| Initial Codex-first advisory pilot | kb-plan -> kb-work -> kb-complete -> kb-ship | `docs/plans/2026-07-10-030-kb-session-model-routing-manifest.md` | active | release command, local-release, complete-to-ship, PR |

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|
| None | - | - | - |

## Notes

- Initial supported claim is Codex CLI plus one generic OpenAI-compatible route hosted through Codex. Codex App-only names, GHCP, TinyBoss control, generated agents, and MCP dispatch remain informative/experimental until their own dispatch conformance gates pass.
- Planner: current orchestration model. Concrete worker models are selected by `kb-work`, not stored in this ledger or manifest.
