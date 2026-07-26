---
kb_id: kb-2026-07-26-change-aware-proof-governor
slice_id: slice-001
title: "Define sealed proof coverage and working-tree-aware receipts"
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: broad
execution_class: cli
model_tier: large
model_tier_reason: "Incorrect proof identity or invalidation can silently skip a changed behavior, so the contract needs conservative dependency and dirty-working-tree semantics."
model_requirements: ["Go schema and CLI contract design", "content-addressed evidence", "working-tree and untracked-file hashing", "negative-test discipline"]
escalation_triggers: ["coverage can be self-asserted without a sealed check registry", "unknown inputs can produce reuse", "dirty relevant files are omitted from fingerprints", "a suite exit code claims unenumerated child coverage"]
context_packet_path: docs/plans/2026-07-26-proof-governor-context/slice-001.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'ProofCoverage|ProofReceipt|RelevantInputFingerprint' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/proof_governor.go
    op: create
    scope: "Define sealed check identities, relevant-input fingerprints, environment compatibility, and proof receipts."
  - path: cmd/kbcheck/proof_governor_test.go
    op: create
    scope: "Protect false-reuse rejection, dirty input hashing, and enumerated suite coverage."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Expose proof-governor contract commands and bounded arguments."
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "Validate receipt contents, lineage, hashes, and freshness instead of accepting a proof path by existence alone."
  - path: config/proof-governor.schema.json
    op: create
    scope: "Document the portable proof spec, receipt, and decision shapes."
protected_oracles:
  - path: cmd/kbcheck/proof_governor_test.go
    role: "false-reuse rejection and working-tree fingerprint oracle"
    sha256: "5aeef634af709f68ccfda6d27556a287eef7c9a79240f5753c25b7f0597811a4"
    oracle_update_reason: "Explicit plan amendment after first green to add the missing public CLI-dispatch acceptance oracle."
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: false
---

# Slice 001: Sealed Proof Coverage and Receipts

## What to Build

Create the smallest deterministic contract that can answer whether an earlier
passing result actually proves a requested check against the current working
tree. A suite may cover child checks only when the registry enumerates those
children. The receipt records the individual check IDs and relevant-input
fingerprints actually evaluated.

## Acceptance Criteria

1. Reuse requires a passing receipt, requested coverage as a subset, exact
   relevant-input hashes, compatible environment fields, and valid policy age.
2. Tracked edits, unstaged edits, staged edits, and relevant untracked files all
   affect fingerprints; unrelated files do not invalidate narrowly declared
   checks.
3. Missing files, broad/ambiguous globs, unknown generated inputs, registry
   drift, or absent environment evidence produce `RUN`, never `REUSE`.
4. A suite exit code cannot claim unnamed child checks, and a failed/partial
   receipt cannot satisfy any later request.
5. Receipts are written atomically under `.kb/proof/` and carry schema version,
   command identity, check IDs, relevant hashes, environment class, timestamps,
   duration, and result without secrets.
6. Identity binds goal/slice/run namespace, oracle/verifier inputs, command and
   argv, working directory, timeout, expected result, environment contract, and
   relevant external evidence; any semantic change changes the digest.
7. Gate-ledger proof that names a receipt validates its schema, lineage, hashes,
   current inputs, and freshness rather than merely checking that a path exists.

## Test Scenarios

- Passing suite `{rust, cli, checksum, browser}` satisfies a later `rust`
  request when fingerprints match.
- One changed Rust input invalidates `rust` and its declared dependents but not
  an unrelated documentation check.
- Dirty and relevant untracked input changes invalidate the prior receipt.
- Timeout, cwd, expected output, verifier script, oracle file, environment, and
  run namespace changes each invalidate the prior receipt.
- Unknown dependency metadata, changed registry hash, failed suite, and
  unenumerated coverage all reject reuse.

## Scope Boundary

This slice defines identity and evidence. It does not yet select a run set,
launch a process, modify snapshot policy, or grant GUI approval.

## Proof

`go test ./cmd/kbcheck -run 'ProofCoverage|ProofReceipt|RelevantInputFingerprint' -count=1`
