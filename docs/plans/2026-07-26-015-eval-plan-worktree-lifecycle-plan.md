---
kb_id: kb-2026-07-26-plan-run-worktree-isolation
slice_id: slice-005
title: "Prove and release the multi-plan worktree lifecycle"
blockers: [slice-004]
verification: integration
test_level: full
functional_risk: full
model_tier: large
model_tier_reason: "The final claim spans Git lifecycle, concurrency, skill behavior, team delivery policy, documentation, and global skill propagation."
model_requirements:
  - "Disposable multi-repository functional harness."
  - "Release-gate and sync verification."
  - "Cross-skill documentation audit."
  - "Failure-artifact preservation."
escalation_triggers:
  - "The functional harness mutates the real repository."
  - "Core or local-release remains unresponsive."
  - "Global skill drift contains useful target-only work."
  - "Any test requires force cleanup or default-branch delivery."
context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-005.json
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck plan-worktree-selftest"
  expect: 0
hitl: false
workspace_mode: shared-serial
conflict_domains: [eval:plan-worktree-lifecycle, docs:workflow, sync:global-skills]
shared_resources: [git:integration-owner, sync:global-skills]
expected_files:
  - path: cmd/kbcheck/plan_worktree_selftest.go
    op: create
    scope: "Run a disposable multi-manifest Git lifecycle and preserve compact failure artifacts."
  - path: cmd/kbcheck/plan_worktree_selftest_test.go
    op: create
    scope: "Protect the end-to-end fixture, real-repo firebreak, and terminal assertions."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Expose the feature-level objective selftest."
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "Add the stable worktree lifecycle contract to the appropriate contributor/release gate."
  - path: cmd/kbcheck/checks_test.go
    op: edit
    scope: "Protect check registration and bounded execution."
  - path: README.md
    op: edit
    scope: "Explain one worktree per manifest group, shared-serial slices, and PR-first team delivery."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Document the two-level workspace, repo-wide scheduler, integration-head, and authority boundaries."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Add focused and full proof commands plus failure interpretation."
  - path: docs/context/eval-map.md
    op: edit
    scope: "Register the disposable multi-plan lifecycle proof surface."
  - path: docs/context/PROJECT.md
    op: edit
    scope: "Refresh the workflow route map and sharp edge after the behavior lands."
  - path: config/skill-quality.json
    op: edit
    scope: "Include changed skills in required sync/release coverage when not already covered."
  - path: todo.md
    op: edit
    scope: "Project final state, blockers, and completion pointer."
  - path: todo-done.md
    op: edit
    scope: "Archive the completed workstream after all proof and release gates pass."
  - path: ~/.codex/skills/kb-plan/
    op: edit
    scope: "Sync the final approved planning contract."
  - path: ~/.copilot/skills/kb-plan/
    op: edit
    scope: "Sync the final approved planning contract."
  - path: ~/.agents/skills/kb-plan/
    op: edit
    scope: "Sync the final approved planning contract."
  - path: ~/.codex/skills/kb-work/
    op: edit
    scope: "Sync the final approved execution and worktree contracts."
  - path: ~/.copilot/skills/kb-work/
    op: edit
    scope: "Sync the final approved execution and worktree contracts."
  - path: ~/.agents/skills/kb-work/
    op: edit
    scope: "Sync the final approved execution and worktree contracts."
protected_oracles:
  - path: cmd/kbcheck/plan_worktree_selftest_test.go
    role: "end-to-end multi-plan lifecycle oracle"
    sha256: "50943d314434e7b87f8bc3b32e408c5b0bede35d5e08cd43ef272b4fec8fa451"
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Return to the completed manifest for finalization and configured delivery."
human_action: ""
can_continue_other_slices: true
---

# Prove and Release the Multi-Plan Worktree Lifecycle

## What to Build

Add one disposable functional harness that proves the full two-level lifecycle
without touching the real repository. Then update the public workflow contract,
refresh project memory, review global skill drift, propagate the approved source,
and run contributor plus release gates.

## Acceptance Criteria

- A disposable repository starts two disjoint plan runs from one default branch;
  each gets its own plan-run branch/worktree.
- Each run advances two serialized slice commits in its one worktree.
- A cross-manifest path/resource collision blocks before mutation with owner
  evidence.
- A dirty or mismatched slice receipt preserves the plan-run worktree, branch,
  receipt, and lease.
- The source/default branch SHA and dirty state remain unchanged throughout
  internal execution.
- Completed local delivery leaves plan-run branches intact.
- PR/manual simulation produces a non-default topic candidate and never merges.
- The harness rejects any target path resolving to the real repository.
- README, architecture, testing, eval map, and project map agree on the same
  boundaries.
- Pre-sync drift is reviewed; repository source is synced outward only after
  useful target-only work is reconciled.
- `go run ./cmd/kbcheck core`, `git diff --check`, and
  `go run ./cmd/kbcheck local-release` pass before completion.

## Test Scenarios

1. Two disjoint manifests, each with two serialized slice commits, complete in
   a temporary repo.
2. Two conflicting manifests are not co-admitted.
3. A shared resource collision blocks despite disjoint files.
4. A dirty, stale-head, or wrong-worktree slice receipt remains recoverable.
5. An attempted internal default-branch integration is refused.
6. Local and PR/manual delivery paths stop before merge.
7. A real-repo target or force cleanup request is rejected by the harness.

## Proof Checks

```powershell
go run ./cmd/kbcheck plan-worktree-selftest
go run ./cmd/kbcheck core
git diff --check
go run ./cmd/kbcheck local-release
```

## Scope Boundary

- No production repository mutations from the functional harness.
- No live PR creation or merge is required by the deterministic test.
- No automatic direct delivery, force cleanup, or distributed lock.
- No external provider or daemon becomes required.

## Dependencies

Requires slices 001-004. Final release also waits for the existing
harness-validation recovery if `core` or `local-release` remains unresponsive.

## Completion Evidence

- `go run ./cmd/kbcheck plan-worktree-selftest` passed with two isolated
  manifest groups, four serialized slice commits, two blocked collisions,
  unchanged source state, and delivery stopped before merge.
- `go test ./cmd/kbcheck -count=1` passed after all review repairs.
- Required Codex, Copilot, and shared-agent skill roots report zero drift.
- `go run ./cmd/kbcheck local-release` passed.
- Independent correctness, structure, and testing review ended with no
  unresolved P0/P1/P2.
- Review-driven scope expansion added commit-authorization provenance, exact
  write/claim validation, live integration-head binding, durable release
  evidence, immutable per-slice proof archives, idempotent completion
  journaling, the public workflow solution, scoped learning, and lifecycle
  archival. Every discovered file was added to the active manifest and slice
  claims before subsequent mutation.
- This implementation worktree was a bootstrap created before plan-worktree
  receipts existed. No receipt was fabricated; the live leases are released
  directly after the final commit.
