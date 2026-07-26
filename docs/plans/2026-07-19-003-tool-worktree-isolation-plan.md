---
kb_id: kb-2026-07-19-evidence-graph-routing
slice_id: slice-003
title: "Isolate mutating slices in worktrees and serialize integration"
superseded_by: docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md
blockers: [slice-002]
verification: integration
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Git lifecycle and cleanup operate around dirty user state; incorrect integration can lose or duplicate work."
model_requirements: ["Git worktree lifecycle", "non-destructive Windows paths", "serialized integration", "post-integration verification"]
escalation_triggers: ["force removal/reset is needed", "worker lifecycle-file writes cannot be separated", "base drift cannot be detected", "cleanup cannot prove integrated and clean state"]
context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-003.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'Worktree|IntegrationOwner'"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/worktree_isolation.go
    op: create
    scope: "Resolve workspace policy, create/reuse slice worktrees, write receipts, integrate, and release safely."
  - path: cmd/kbcheck/worktree_isolation_test.go
    op: create
    scope: "Temporary-repo dirty checkout, branch uniqueness, conflict, proof, and cleanup fixtures."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Expose workspace prepare/status/integrate/release commands."
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "Validate isolation intent/conflict domains when concurrency is declared."
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: "Compatibility and invalid concurrent-plan fixtures."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Plan isolation intent and conflict domains without live paths."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Own live worktree creation, coordinator/worker split, serialized integration, and cleanup."
  - path: .github/skills/kb-work/references/worktree-isolation.md
    op: create
    scope: "Lazy-loaded runtime worktree and integration checklist."
  - path: .github/skills/kb-regression-snapshot/SKILL.md
    op: edit
    scope: "Define snapshot visibility and replay across worktrees/integration."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Durable multi-session isolation architecture."
protected_oracles:
  - path: cmd/kbcheck/worktree_isolation_test.go
    role: "dirty-state preservation and integration oracle"
    sha256: "F52A314AC9E21D6EEAB78728189C7953FFB6E89FA79351B4E798F8DF30542D30"
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Proceed to slices 004 and 005 through the new lease/worktree path when safe."
human_action: ""
can_continue_other_slices: false
---

# Slice 003: Worktree Isolation And Integration

> Superseded on 2026-07-26. Active plan runs now use one manifest-owned
> worktree and branch with shared-serial slices; they do not create per-slice
> worktrees. See
> `docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md` and
> `.github/skills/kb-work/references/worktree-isolation.md`.

## What To Build

Add a runtime workspace adapter. `kb-plan` declares `workspace_mode`, conflict
domains, shared resources, and integration dependencies; `kb-work` resolves
live state, atomically claims the slice, creates/reuses a unique
`codex/<kb-id>/<slice-id>-<short-id>` branch/worktree when required, and writes
a receipt. One coordinator integrates results and updates canonical lifecycle
files.

## Why This Slice Exists

Worktrees prevent concurrent filesystem mutation, but without atomic claims and
serialized mergeback they only move the race into `todo.md`, manifests, graph
indexes, and Git conflicts. Plan-time paths also drift, so mechanics belong at
work time.

## Acceptance Criteria

- Plan fields describe isolation intent/conflict domains without absolute paths.
- Runtime chooses `shared-serial` for one safe mutator and
  `worktree-required` for concurrent mutators, dirty shared checkout risk, or an
  explicit isolation requirement.
- Workers do not directly integrate or independently update canonical
  `todo.md`, manifest, or handoffs; they return commit/diff, proof receipt,
  observed writes, and conflict discoveries.
- Integration validates owner token and base SHA, occurs one result at a time,
  rechecks diff scope, and reruns proof on the integration branch.
- Browser/ports/databases/generated/index namespaces serialize unless explicitly isolated.
- Cleanup never uses force and requires integrated status, clean worktree, and
  matching lease release.
- Existing dirty/untracked files in the source checkout remain byte-for-byte unchanged.
- Regression snapshots needed across slices are exported/replayed through
  receipts rather than assumed to exist in another worktree.
- Pre-edit drift for `kb-plan`, `kb-work`, and `kb-regression-snapshot` is reviewed.

## Test Scenarios

- Dirty main + isolated slice leaves main unchanged.
- Two disjoint slice worktrees run and integrate serially.
- Same path/conflict domain causes requeue before editing.
- Base revision changes before integration: fail and require rebase/re-plan.
- Merge conflict preserves both worktrees and leases for recovery.
- Integrated clean worktree releases; dirty/unintegrated cleanup refuses.

## Proof Check

`go test ./cmd/kbcheck -run 'Worktree|IntegrationOwner'`

Passed 2026-07-19 with:

- `go test ./cmd/kbcheck -run 'Worktree|IntegrationOwner' -count=1 -v`
- `go run ./cmd/kbcheck manifest-contract --manifest docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .github\skills\kb-regression-snapshot\scripts\kb-regression-snapshot.ps1 capture -SliceId evidence-graph-routing-slice-003 -SpecPath .kb\snapshots\evidence-graph-routing-slice-003-spec.json`

Protected oracle:

- `cmd/kbcheck/worktree_isolation_test.go` SHA256 `F52A314AC9E21D6EEAB78728189C7953FFB6E89FA79351B4E798F8DF30542D30`

Snapshot:

- `.kb/snapshots/evidence-graph-routing-slice-003.json`

## Scope Boundary

No daemon, tmux/session manager, remote clone coordination, force cleanup, or
automatic PR delivery.

## Dependencies

Slice 002 provides atomic ownership. Worktree creation without that ownership
would still permit duplicate work.

## Concurrency

Serial while implemented. After integration, it unlocks the parallel pair of
slices 004 and 005.
