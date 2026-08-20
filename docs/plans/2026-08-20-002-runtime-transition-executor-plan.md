---
kb_id: kb-2026-08-20-runtime-state-contract
slice_id: runtime-002
title: Execute transitions through structured tool actions and boundary projections
blockers: [runtime-001]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: Integrates phase owners with a new deterministic runtime surface.
model_requirements: [CLI design, lifecycle integration, proof reuse]
escalation_triggers: [tool action can mutate outside declared phase, projection becomes source of truth]
token_budget: 7000
workspace_mode: shared-serial
conflict_domains: [namespace:runtime-transition, skill:kb-work, skill:kb-finalize, skill:kb-complete]
proof_check: {kind: command_exit, command: "go test ./cmd/kbcheck -run RuntimeTransition -count=1", expect: 0}
hitl: false
status: pending
---

# Execute Transitions Through Structured Tool Actions

## Acceptance Criteria

- Phase adapters request deterministic actions using machine-readable inputs.
- The executor emits machine-readable outputs and a compact Markdown projection
  only at user, reviewer, or PR boundaries.
- Reusable exact-tree proof is not rerun merely to generate new prose.
- A repository-local one-candidate gate blocks creation of a second plan-run
  branch/worktree/PR while the current delivery train is locally durable,
  awaiting review, or otherwise unsettled.
- Default harness output provides one decision-first human projection: current
  state, recommendation/next action, proof status, and blocker if any. JSON and
  raw logs remain machine/detail surfaces, never the required human interface.

## Expected Files

- `cmd/kbcheck/*runtime_transition*.go` and tests.
- `.github/skills/kb-work/SKILL.md`, `kb-finalize/SKILL.md`, `kb-complete/SKILL.md`.
- `docs/context/operations/testing.md`.

## Scope Boundary

No change to delivery policy or model-routing selection logic.
