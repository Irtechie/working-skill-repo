---
kb_id: kb-2026-07-19-evidence-graph-routing
slice_id: slice-002
title: "Atomically claim and release slices across local sessions"
blockers: [slice-001]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Cross-process ownership and stale recovery must be correct on Windows and Unix or the workflow can duplicate and overwrite work."
model_requirements: ["cross-process race tests", "Windows and Unix atomic filesystem semantics", "Git common-dir identity", "fail-closed ownership tokens"]
escalation_triggers: ["same-slice double claims succeed", "recovery can steal a live lease", "coordination escapes the Git common directory", "a daemon becomes required"]
context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-002.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'SliceLease|ScopeLease'"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/slice_lease.go
    op: create
    scope: "Atomic acquire, status, renew, release, and guarded stale recovery under Git common-dir state."
  - path: cmd/kbcheck/slice_lease_test.go
    op: create
    scope: "Two-process races, owner tokens, expiry, same-slice duplication, and path safety."
  - path: cmd/kbcheck/swarm.go
    op: edit
    scope: "Keep passive scope-collision reporting but bind it to canonical active claims."
  - path: cmd/kbcheck/swarm_test.go
    op: edit
    scope: "Compatibility and path-prefix conflict cases."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Expose slice-lease acquire/status/renew/release/recover commands."
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "Register deterministic concurrency selftests in core."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Require atomic claim before board projection or mutation."
  - path: .github/skills/kb-work/references/execution-prompt.md
    op: edit
    scope: "Pass and release the owner token in slice execution."
protected_oracles:
  - path: cmd/kbcheck/slice_lease_test.go
    role: "cross-process claim and recovery oracle"
    sha256: "9371669B79F75CF9F43D5A5CB0949933EEB25458E9B63BCC5E6939154B64C8A1"
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Proceed to slice 003 worktree isolation and serialized integration."
human_action: ""
can_continue_other_slices: false
---

# Slice 002: Atomic Local Slice Claims

## What To Build

Turn the passive scope-lease checker into a real local coordination primitive.
Use a canonical Git common-directory state root so sibling worktrees share the
same claims. Record an opaque owner token, run/slice IDs, canonical repo
identity, base SHA, worktree/branch when known, conflict domains, timestamps,
heartbeat, and status.

## Why This Slice Exists

`todo.md` claiming is a TOCTOU race: two sessions may both read `pending` and
both begin. The existing checker only detects collisions in a ledger someone
already wrote, and treats two sessions with the same slice ID as the same owner.

## Acceptance Criteria

- Exactly one of two simultaneous same-slice acquisition processes succeeds.
- Only the matching owner token may renew or release a live lease.
- Stale recovery requires expiry plus a compare-and-swap generation/token; it
  never silently steals a live lease.
- Claims coordinate worktrees sharing one Git common directory and explicitly
  do not claim coordination across separate clones or machines.
- Conflict domains support exact files plus path prefixes, generated outputs,
  browser/port/database/index resources, and canonical path normalization.
- Board/manifest status is projected only after acquisition; acquisition
  failure leaves them unchanged.
- No daemon, network service, broad home scan, or global state is introduced.
- Pre-edit `kb-work` drift is reviewed across required skill roots.

## Test Scenarios

- Two processes race for one slice: one winner, one deterministic collision.
- Two sessions using the same slice ID still conflict through distinct owner tokens.
- Disjoint claims succeed; parent/child path claims conflict.
- Wrong-token release fails; expired generation mismatch fails.
- Two worktrees share claims; an unrelated clone does not.
- Windows case/slash variants and symlink/path escapes fail safely.

## Proof Check

`go test ./cmd/kbcheck -run 'SliceLease|ScopeLease'`

Passed 2026-07-19 with:

- `go test ./cmd/kbcheck -run 'SliceLease|ScopeLease' -count=1 -v`
- `go run ./cmd/kbcheck slice-lease-selftest`

Protected oracle:

- `cmd/kbcheck/slice_lease_test.go` SHA256 `9371669B79F75CF9F43D5A5CB0949933EEB25458E9B63BCC5E6939154B64C8A1`

## Scope Boundary

No worktree creation or mergeback yet. The board remains the human view, but the
atomic lease becomes execution authority.

## Dependencies

Slice 001 supplies canonical repo/freshness identity reused by claim receipts.

## Concurrency

This slice itself is serial-only. Its tests spawn competing processes in
temporary repositories; they do not race on the live checkout.
