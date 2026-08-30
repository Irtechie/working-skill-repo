# Bounded Graph-Run Provenance

Status: retired
Created: 2026-07-30
Last updated: 2026-08-28

## Retirement

Retired on 2026-08-28 without delivery. Planning completed on 2026-07-30
(manifest, five slice plans, five context packets, and a research note), but
execution never advanced past the slice-001 red test. That test lived alone on
`codex/kb-2026-07-30-bounded-graph-run-provenance` with no implementation
behind it, so it could never merge without breaking the build, and the branch
was dropped when this goal was retired.

Nothing in the shipped tool depends on graph-run storage, so no consumer
regresses. The slice plans and research note are kept as design records. Reopen
by setting this status back to `active` and re-planning from the existing
slice plans rather than resurrecting the dropped branch.

## Objective

Deliver bounded graph-run storage, immutable node-attempt receipts, gate-linked
provenance, compact failure/completion diagnostics, and fault-injection proof,
then commit, push, open, and merge the reviewed PR.

## Done Criteria

- Marker-owned graph-run storage has inspectable age and byte accounting,
  dry-run cleanup, bounded retention, and safe apply behavior that preserves
  active, pinned, corrupt, and unowned paths.
- A versioned, OpenTelemetry-compatible node-attempt receipt records only
  bounded metadata and hashes; it stores no prompts, model outputs, diffs, raw
  transcripts, or screenshots.
- Node receipts link existing dependency, lease, revision, gate, and proof
  evidence without creating a second orchestration or telemetry platform.
- `graph-run inspect --failed` identifies the first causal failure.
- `graph-run inspect --why-not-done` explains missing terminal nodes, proof,
  fan-in gates, or unresolved blocking edges.
- Deterministic fault-injection tests prove diagnosis, bounded retry without
  replaying accepted nodes, and completion accounting.
- Relevant architecture, testing, and user-facing command documentation is
  current.
- The reviewed change is committed, pushed, delivered through a PR, merged, and
  confirmed on the authoritative remote default branch.

## Terminal Proof

- Targeted graph-run contract, retention, diagnostics, and fault-injection tests pass.
- `go run ./cmd/kbcheck core` passes for the delivery tree.
- `go run ./cmd/kbcheck local-release` passes for the delivery tree.
- The PR is merged and `origin/main` contains the delivered commit.

## Done Check

- Type: command_exit
- Check: `go run ./cmd/kbcheck local-release`
- Expected result: exit code 0
- Why sufficient: proves the repo-local release contract for the exact delivery tree; PR merge and remote-default containment remain separate terminal delivery proof.

## Current State

- Current artifact: `docs/plans/2026-07-30-000-kb-bounded-graph-run-provenance-manifest.md`
- Next allowed action: `kb-work docs/plans/2026-07-30-000-kb-bounded-graph-run-provenance-manifest.md`
- Last proof: no open pull requests existed in `Irtechie/working-skill-repo` on 2026-07-30

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|
| Storage accounting and retention | `kb-work` | `docs/plans/2026-07-30-001-tool-graph-run-storage-plan.md` | planned | manifest validation |
| Node-attempt receipt contract | `kb-work` | `docs/plans/2026-07-30-002-tool-node-attempt-receipt-plan.md` | planned | manifest validation |
| Receipt emission and gate linkage | `kb-work` | `docs/plans/2026-07-30-003-tool-gate-linked-provenance-plan.md` | planned | manifest validation |
| Failure/completion projections | `kb-work` | `docs/plans/2026-07-30-004-tool-graph-run-diagnostics-plan.md` | planned | manifest validation |
| Fault-injection proof | `kb-work` | `docs/plans/2026-07-30-005-eval-graph-run-fault-injection-plan.md` | planned | manifest validation |
| Review and PR delivery | `kb-complete` | pending PR | pending | pending |

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|
| Existing Windows dispatch tests can fail with temp `.cmd` access denied inside `local-release` | delivery-only | harness-validation recovery workstream | Named release sensor passes on the exact delivery tree; do not expand this goal into unrelated harness repair |

## Notes

- Scope corresponds to approved items 2–6 only.
- No OpenTelemetry Collector, observability backend, raw trace capture,
  persistent agent organization, runtime hooks, or universal tool tracing.
- Static context-economics measurement and tool-enforcement work remain deferred.
