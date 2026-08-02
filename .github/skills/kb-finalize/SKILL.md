---
name: kb-finalize
description: "Internal post-work quality phase. Runs aggregate proof, one proportional semantic review, final exact-tree proof, signal-driven learning/memory work, and cleanup."
argument-hint: "[path to completed KB manifest]"
---

# KB Finalize - One Integrated Quality Boundary

Run once after all slices are integrated. This phase proves the complete tree,
selects zero or one semantic reviewer, resolves findings, reruns invalidated
proof, and performs learning or memory work only when evidence warrants it.

## Terminal Contract

Continue until the manifest has a passing `complete-to-ship` gate or a real
blocker with an exact resume condition. Passing slice tests or finishing review
alone is not terminal.

## Preflight

1. Confirm the current session owns the shared work claim.
2. Require manifest `status: completed`, all slices done/skipped, and a passing
   `work-to-complete` gate that allows this exact finalization command.
3. Run `manifest-contract` when the repo provides it.
4. Collect authoritative requirements, requirements hash, accepted slice
   commits, scope-verified paths, proof receipts, and memory-impact notes.
5. Resolve the immutable baseline and integrated tree.
6. Reuse valid slice receipts and run one aggregate proof for the integrated
   tree. Use `kb-check proof-plan` to avoid replaying unchanged checks.
7. Require matching functional proof for UI, API, CLI, persistence, auth,
   streaming, or integration behavior.

Do not enter semantic review with failing or missing integrated proof.

## Step 1: Review Classification

Invoke `kb-review` once unless the entire diff qualifies for its conservative
skip classification:

- docs-only with no executable contract change;
- generated-only from a proven generator; or
- mechanically constrained changes fully covered by deterministic proof.

Runtime, behavior, contract, configuration, trust-boundary, persistence, API,
CLI, or UI changes require review. Unknown classification requires review.

Pass:

- authoritative intent and requirements hash;
- base and integrated tree;
- aggregate proof receipt and hash;
- exact scope-verified paths;
- manifest path.

`kb-review` chooses one broad or replacement specialist profile. Never invoke
another reviewer at this boundary. Record the review/skip receipt fingerprint,
profile, mode, counts, and residual risk in the manifest.

## Step 2: Findings and Fixes

| Severity | Completion rule |
|---|---|
| P0/P1 | Resolve before completion |
| P2/P3 | Fix when cheap and clearly correct; otherwise record with owner |

Apply only deterministic safe fixes automatically. Route behavior, contract,
permission, migration, or product decisions through `kb-gate`.

A code-affecting fix invalidates:

- the aggregate proof receipt;
- affected functional proof;
- the semantic review receipt.

After the last fix, rerun only affected deterministic/functional checks, then
run one bounded confirmation review with the same profile. This confirmation
is the same review boundary, not permission to add another profile.

Record `follow-up-resolution: resolved N, logged M, blocked K`.

## Step 3: Final Exact-Tree Proof

Run or reuse one aggregate receipt bound to the final tree after all fixes.
Every slice must retain machine-verifiable proof. Screenshots and prose may
support proof but cannot replace executable assertions.

Record:

- commands or receipts reused/run;
- tree and relevant-input fingerprints;
- functional routes/workflows exercised;
- artifact paths;
- top-level `done_check` and each slice `proof_check`.

If exact-tree proof fails, return to bounded repair. Do not declare completion.

## Step 4: Signal-Driven Knowledge Work

Do not run compound, learn, evolve, memory refresh, memory review, or compact
merely because finalization occurred.

| Signal | Action |
|---|---|
| Novel solved problem, non-obvious decision, or reusable failure mode | Invoke `ce-compound` |
| Repeated correction, resolved P0/P1 pattern, or evidenced landmine | Invoke `learn` |
| Mature instincts meet the configured cadence/threshold | Invoke `evolve` |
| Durable behavior, architecture, commands, integration, or sharp edge changed | Invoke `kb-map refresh` |
| Recorded contradiction, overlap, stale-doc, rediscovery, or bloat threshold | Update maintenance signals; invoke `kb-memory-review` only when due |
| A specific startup document is too large | Invoke `kb-compact` for that path |

Otherwise record a specific skip reason. Never create lessons from routine
success, style preferences, or generic advice. Keep workflow/domain lessons at
the narrowest owning scope.

## Step 5: Cleanup and Gate

1. Keep manifests and slice plans.
2. Remove only exact, run-owned ephemeral artifacts. If the tree still holds
   uncommitted work that the run does not intend to deliver, preserve it with
   `kbcheck session-preserve --action apply --session-id <session-id>` before
   any cleanup step. A preserved WIP commit is durability only: it is never
   pushed and never counts as proof, review, or completion evidence.
3. Preserve stable build caches and use repository cleanup receipts where
   required.
4. Move the completed feature summary from `todo.md` to `todo-done.md`; retain
   active, blocked, human-required, parked, and handoff-pointer work.
5. Register completion state and evidence for `kb-complete`, including the exact
   tree, proof receipt, manifest/gate, session, branch, and declared semantic
   resources. Finalization never deletes the current worktree and does not
   release its ownership claim; registration and endpoint selection happen
   before a later controller performs physical cleanup.
6. Write `complete-to-ship` with:
   - final exact-tree proof and functional proof;
   - review or valid skip receipt;
   - P0/P1 resolution;
   - follow-up summary;
   - signal-driven knowledge/memory outcomes or skip reasons;
   - cleanup result and alerts.
7. Run the gate-ledger checker. Set manifest `status: reviewed` only when the
   gate passes or explicitly quarantines unrelated work.
8. Require an advanceable `passed|quarantined` gate before leaving
   finalization. Preserve quarantine evidence and forbidden claims.
9. Invoke `kb-complete <manifest>` automatically.

`kb-finalize` does not commit, push, open a PR, merge, or integrate the resolved
remote default branch. It returns reviewed durable state to the delivery
orchestrator; it does not acquire delivery authority.

## Failure Policy

- Before reporting any blocked or human-required item, rerun its named sensor or
  the cheapest owning probe.
- Proof or unresolved P0/P1 is an implementation blocker.
- Reviewer unavailability permits one honest local fallback, not a swarm.
- Optional knowledge or compaction failure is non-blocking but recorded.
- Delivery, signing, deployment, or optional-platform failure does not erase
  proven implementation completion.
- Report proven code as `Implementation: complete` and the affected downstream
  state as `Delivery: blocked` when only delivery failed.
- Never ask to skip a mandatory proof or review requirement.

## Build Storage Cleanup

When Cargo ran, execute `cargo-storage --action finalize` with the run receipt.
Record `retained_bytes` and `removed_bytes`. In portable fallback mode,
temporary targets and deletion were prohibited; retain the stable target and
record `removed_bytes: 0`. Record `not-applicable` only when Cargo did not run.

## Done

Report:

```text
KB <name> finalized.
- Proof: <exact-tree receipt>
- Review: <profile or conservative skip>; P0=N P1=N P2=N P3=N
- Follow-up: resolved N, logged M, blocked K
- Knowledge/memory: <actions or skip reasons>
- Cleanup: done

Continuing automatically to kb-complete <manifest>.
```

## Integration

- Input: `kb-work` or `kb-complete`
- Proof: `kb-check`, `kb-functional-test`, regression snapshots
- Semantic review: `kb-review`
- Decisions: `kb-gate`
- Conditional knowledge: `ce-compound`, `learn`, `evolve`
- Conditional memory: `kb-map`, `kb-memory-review`, `kb-compact`
- Delivery owner: `kb-complete`
