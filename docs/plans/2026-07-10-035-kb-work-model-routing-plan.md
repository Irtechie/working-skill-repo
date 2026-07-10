---
kb_id: kb-2026-07-10-session-model-routing
slice_id: slice-005
title: "Route kb-work slices by live difficulty and preserve proof"
blockers: [slice-001, slice-004]
verification: integration
test_level: functional-cli
functional_risk: broad
model_tier: medium
context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-005.json
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck skill-lint"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-goal/SKILL.md
    op: edit
    scope: "initialize/reuse one ephemeral session catalog and keep model details out of durable goals"
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "discover, preview, select, dispatch, fallback, and record receipts at work time"
  - path: .github/skills/kb-work/references/execution-prompt.md
    op: edit
    scope: "carry route request, bounded authority, receipt, and escalation handoff contract"
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "review routing evidence without choosing models or invalidating proven work"
  - path: evals/skill-eval/selftest/pass-session-model-routing.json
    op: create
    scope: "route-time behavior fixture for selection, overrides, fallback, and provenance"
protected_oracles:
  - path: evals/skill-eval/selftest/pass-session-model-routing.json
    role: "KB workflow routing behavior oracle"
    sha256: "342ba9e2bf109433e145453f0102ea814acc7c3fd41ec52d31660214201ef669"
    update_policy: "requires explicit plan update"
status: done
owner: agent
can_continue_other_slices: true
---

# KB Work-Time Model Routing

## What To Build

Make `kb-work` call the live router for each ready slice, show compact non-empty route groups, dispatch bounded subagents, and preserve every existing proof/scope/HITL gate.

## Acceptance Criteria

- Goal/work owns one run catalog and refreshes only on fingerprint change.
- Plan tier is a capability floor; selection occurs immediately before dispatch and never appears as a plan commitment.
- `use`, `require`, `prefer local/hosted`, different-family review, and `ignore model routing` have exact run-scoped semantics.
- Missing router or unavailable routes use the current model when policy permits and do not block ordinary proof.
- Pre-existing correct work with unknown provenance is investigated once and never redone for telemetry.

## Test Scenarios

- Medium route absent -> same-class alternative -> stronger qualified route -> current model.
- Exact required route absent pauses only that slice; unrelated ready work continues.
- Generic spawn with no selector cannot claim a requested model.
- Completed external work lands with `unknown` provenance and passing independent proof.

## Tier Rationale

Medium: workflow integration follows the now-deterministic router contract; escalate if host capability or safety policy changes.

## Scope Boundary

No new task taxonomy, proof weakening, direct endpoint secrets, GHCP support claim, or model rerun during completion.

## Completion Evidence

- Medium worker: Codex CLI `gpt-5.4`, thread `019f4de1-f9ce-7e53-88db-722752518e10`.
- RED: the new selftest failed until its required sync-proof evidence was represented.
- GREEN: focused fixture, full skill-eval selftest (14 files), skill lint, and slice-005 context-packet validation passed.
- The fixture records model-reported trace as self-reported, not externally observed; ordinary deterministic checks remain authoritative.
- Post-review remediation added the public non-mutating `kbrouter models select`
  surface, per-slice dispatch schema names, exact discover/select/dispatch skill
  commands, and hardened hashed/contained Go sandbox paths. Focused router tests
  and skill lint pass; the dispatch oracle hash remains unchanged.
