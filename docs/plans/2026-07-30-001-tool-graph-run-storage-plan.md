---
kb_id: kb-2026-07-30-bounded-graph-run-provenance
slice_id: slice-001
title: "Bound graph-run storage and retention"
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Deletion safety crosses path containment, ownership markers, concurrent state, byte accounting, and active-run preservation."
model_requirements: ["Go filesystem safety", "cross-platform locking", "atomic receipt persistence", "adversarial cleanup testing"]
escalation_triggers: ["cleanup can reach an unowned path", "active or pinned state lacks an authoritative sensor", "symlink containment cannot be proven", "retention requires a daemon"]
workspace_mode: shared-serial
conflict_domains: ["file:cmd/kbcheck/graph_run_storage.go", "file:cmd/kbcheck/main.go", "namespace:.kb/runs"]
shared_resources: ["git:integration-owner", "storage:.kb/runs"]
context_packet_path: docs/plans/2026-07-30-bounded-graph-run-provenance-context/slice-001.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'GraphRunStorage' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/graph_run_storage.go
    op: create
    scope: "Implement marker-owned graph-run accounting, dry-run retention planning, and guarded cleanup apply."
  - path: cmd/kbcheck/graph_run_storage_test.go
    op: create
    scope: "Protect ownership, containment, active/pinned/corrupt preservation, byte accounting, and interrupted-cleanup behavior."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Register the bounded graph-run storage CLI surface and flags."
protected_oracles:
  - path: cmd/kbcheck/graph_run_storage_test.go
    role: "cleanup safety and accounting oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Write protected cleanup-safety tests, then implement marker-owned accounting, dry-run retention, and guarded apply."
human_action: ""
can_continue_other_slices: true
---

# Bound Graph-Run Storage and Retention

## What to Build

Add a cross-platform graph-run storage lifecycle under repo-local `.kb/runs`
that reports per-run and aggregate age/bytes, computes a bounded retention plan,
and applies cleanup only to terminal marker-owned runs.

## Acceptance Criteria

- Every managed run directory has an ownership marker and stable run identity.
- Inspection reports age, bytes, lifecycle status, pin state, ownership state,
  corruption state, and cleanup eligibility.
- Dry-run is the default and reports exact retained/removed byte projections.
- Apply persists deletion intent before mutation and reconciles interruption.
- Active, pinned, corrupt, unowned, escaped, and symlinked paths are retained
  with an explicit reason.
- Locking reloads state after acquisition and concurrent updates do not lose
  accounting.

## Test Scenarios

- Terminal owned run is selected in dry-run and removed in apply.
- Active and pinned runs remain.
- Missing or forged marker remains and reports unowned.
- Corrupt receipt remains and reports corrupt.
- Escaped path and symlink are rejected.
- Interrupted deletion reconciles on retry.
- Concurrent registrations preserve both runs.

## Scope Boundary

No global cleanup, background sweeper, daemon, raw trace storage, or deletion of
legacy/unowned `.kb` data.
