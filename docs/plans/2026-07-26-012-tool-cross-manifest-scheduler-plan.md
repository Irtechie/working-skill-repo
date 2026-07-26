---
kb_id: kb-2026-07-26-plan-run-worktree-isolation
slice_id: slice-002
title: "Block cross-manifest conflicts before mutation"
blockers: [slice-001]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Cross-process scheduling and shared-resource claims must be atomic and fail closed or multiple valid manifests will still race."
model_requirements:
  - "Cross-process lease design."
  - "Git common-directory identity."
  - "Path, prefix, conflict-domain, and resource normalization."
  - "Deterministic contention and stale-recovery tests."
escalation_triggers:
  - "Two conflicting manifests can both acquire mutation authority."
  - "Separate clones are falsely reported as coordinated."
  - "A stale run can be stolen without token and generation proof."
  - "Shared resources cannot be represented without a mandatory daemon."
context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-002.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'PlanRunLease|CrossManifestScheduler|ScopeLease' -count=1"
  expect: 0
hitl: false
workspace_mode: shared-serial
conflict_domains: [file:cmd/kbcheck/plan_run_scheduler.go, file:cmd/kbcheck/slice_lease.go, skill:kb-work]
shared_resources: [git:integration-owner, git:plan-run-lease]
expected_files:
  - path: cmd/kbcheck/plan_run_scheduler.go
    op: create
    scope: "Acquire, renew, release, inspect, and recover manifest-level path/domain/resource claims."
  - path: cmd/kbcheck/cross_manifest_scheduler_test.go
    op: create
    scope: "Protect conflicting-run exclusion, disjoint-run admission, stale recovery, and separate-clone limitations."
  - path: cmd/kbcheck/slice_lease.go
    op: edit
    scope: "Reuse canonical claim normalization and generation/token semantics without creating a second incompatible lock model."
  - path: cmd/kbcheck/slice_lease_test.go
    op: edit
    scope: "Prove run and slice claims compose under the same Git common directory."
  - path: cmd/kbcheck/swarm.go
    op: edit
    scope: "Compute the repo-wide safe ready set across active manifests before dispatch."
  - path: cmd/kbcheck/swarm_test.go
    op: edit
    scope: "Reject cross-manifest path/resource collisions while preserving disjoint concurrency."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Expose plan-run claim status and deterministic selftest surfaces."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Acquire manifest-level authority before board projection, slice claims, or mutation."
protected_oracles:
  - path: cmd/kbcheck/cross_manifest_scheduler_test.go
    role: "cross-manifest path and shared-resource exclusion oracle"
    sha256: "3816b2eb97f511390a04d43f94a3df18a419b88cc3a43e977f19860ba31b848a"
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Proceed to slice-003 and protect serialized slice-commit advancement on the owning plan-run branch."
human_action: ""
can_continue_other_slices: false
---

# Block Cross-Manifest Conflicts Before Mutation

## What to Build

Add a repo-wide local scheduler that sees all active plan runs sharing one Git
common directory. It must claim forecast paths, prefixes, conflict domains, and
shared resources before a manifest becomes mutating-ready.

The scheduler admits disjoint plan runs concurrently and serializes or requeues
overlapping runs. Observed writes/resources expand the live claim; they override
the original forecast.

## Acceptance Criteria

- Four or five manifests no longer operate as isolated DAG namespaces.
- Conflicting path, prefix, skill, generated-output, port/database, global-sync,
  or integration-owner claims block before mutation.
- Disjoint plan runs may remain active concurrently.
- A manifest-level claim and its slice claims compose under one owner/run
  lineage; they cannot contradict each other.
- Claim acquisition, renewal, release, and recovery use owner token plus
  generation and fail closed.
- The board and manifest remain unchanged when acquisition fails.
- The scheduler reports exactly which run/domain/resource caused serialization.
- Separate clones/machines are explicitly uncoordinated locally and rely on
  branch/PR protections instead of a false distributed-lock claim.

## Test Scenarios

1. Two manifests forecast the same file/prefix; only one acquires mutation
   authority.
2. Two manifests claim the same browser port or global skill sync; only one is
   admitted even with disjoint files.
3. Two disjoint manifests are admitted together.
4. A discovered write collides with another active run; the discovering run is
   requeued before that write.
5. Wrong-owner and stale-generation renew/release/recover fail.
6. Sibling worktrees share state; a separate clone does not.

## Proof Check

```powershell
go test ./cmd/kbcheck -run 'PlanRunLease|CrossManifestScheduler|ScopeLease' -count=1
```

## Scope Boundary

- No remote distributed lock or mandatory daemon.
- No automatic conflict resolution.
- No default-branch delivery.
- No claim that worktrees make ports, databases, generated outputs, or global
  installs safe without an explicit resource claim.

## Dependencies

Requires slice-001 plan-run identity and common-dir receipt ownership.
