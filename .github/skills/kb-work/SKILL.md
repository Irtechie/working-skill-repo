---
name: kb-work
description: "Execute ready vertical slices from a validated KB manifest with bounded ownership, deterministic proof, scope control, and automatic finalization."
argument-hint: "[manifest path]"
---

# KB Work - Bounded Slice Execution

Execute the safe ready set until every slice is terminal, then invoke
`kb-finalize`. Do not review code per slice.

## Input

<input> #$ARGUMENTS </input>

Require a KB manifest. If input is a feature or unsliced handoff, invoke
`kb-plan` first. If blank, use the one active manifest or ask only when several
are plausible.

## Preflight

1. Require the current session's live `kb-start` work claim.
2. Read the manifest and validate the acyclic DAG, slice files, statuses, and
   expected-file forecasts.
3. Require `plan-to-work: passed` for this exact manifest.
4. Validate `pre_slice_review`: one completed requirements-wide reviewer or a
   specific `not_required_reason`. Document-review personas are not slice
   implementation owners. Do not rerun plan-wide specialist personas per slice.
   A material new product, flow, architecture, scope, or trust decision must
   return to `kb-plan` for one new requirements-wide review.
5. Run `manifest-contract` and validate required context/impact packets.
6. Note dirty source work. Prepare/resume the manifest-owned plan-run worktree;
   never copy, stash, reset, or overwrite unrelated work. Give a new worktree a
   short, task-specific codename with irreverent, self-aware humor that remains
   safe to show at work; reject fandom themes and generic random names.
7. Acquire the plan-run lease, then a slice lease before board projection or
   mutation. Expand claims before touching discovered paths.
8. Read relevant active landmines and proof receipts.

Manifest worktrees and leases coordinate sibling worktrees in one local Git
common directory. Separate clones or machines are outside that guarantee:
local leases are not team locks.

## Continuous Loop

An explicit pause is not technical terminal proof. Stop mutation immediately
when the user pauses; write state only when requested. Otherwise continue until:

- all slices are done/skipped and finalization completes;
- only real blocked, human-required, or parked work remains with exact resume
  criteria; or
- the user stops.

Continue unrelated runnable slices. A blocked dependency stops only its real
dependents.

## Ready Set

A pending slice is ready when every blocker is done/skipped and no active
claim conflicts. Dispatch independent ready slices in parallel only with proven
write/resource isolation. A shared checkout runs one mutating slice at a time.
Observed overlap overrides forecasts.

## Proof Batch Cadence

- Run narrow slice-local proof after implementation stabilizes.
- Run protected-oracle and safety checks immediately when touched.
- Do not run the manifest aggregate after each slice.
- Run aggregate proof at coherent integrated batch boundaries.
- Bind the final aggregate to the exact final tree.
- Reuse fresh receipts whose tree and relevant-input fingerprints match.

Before a later batch, verify affected regression snapshots. A failure blocks
new mutation until repaired or explicitly parked/skipped.

## Build Storage

Before Cargo work, require a native receipt accepted by
`cargo-storage --action validate-ready` or the fail-closed portable fallback.
Apply the returned exact `CARGO_TARGET_DIR` across every slice, worker, repair,
proof batch, session, and worktree. A temporary target must be created first in
native mode with an approved parent and technical reason. Never invent
phase-, worker-, slice-, or run-named targets.

## Live Route Selection

By default, normal work uses delegation-first DDR. The orchestrator owns decomposition,
tier judgment, route selection, supervision, proof, and synthesis.

1. **Current:** retain execution only through a recognized exception:
   `reasoning-required`, `context-required`, `tool-required`,
   `authority-required`, `trust-required`, `user-required`, or
   `no-qualified-route`.
2. **Native host delegation:** inspect the active host's exact callable-agent
   schema and choose one qualified target.
3. **CLI or user-local delegation:** use the live `kbrouter` catalog once per
   unchanged host/config fingerprint. Do not merge App-only aliases with CLI-only aliases.

Default `execution_owner` to `delegated`.
The current-owner exception gate accepts only the recognized reasons above.
`no-qualified-route` is valid only after inspecting both host-native and CLI/user-local surfaces.
Resolve that portable tier when the plan is picked up.
"Exactly one" is per slice, not per plan.
For a disjoint ready set, dispatch those subagents in parallel.

After the ownership decision and, when delegated, route selection, emit exactly
one compact user-visible line before mutation or worker dispatch:

```text
DDR route: <current|subagent> | primary: <current orchestrator|evidence-backed-route> | return: <none|parent-on-first-local-failure|required-alias-block> | tier: <small|medium|large> | proof: <short-proof-target>
```

The orchestrator is the sole emitter. A worker receives the populated line and
must not repeat it. Use a route name only from live evidence; otherwise use `current orchestrator`. A preferred local route gets one attempt and returns
`parent-on-first-local-failure`. Do not select a second local route. A required
alias blocks.

A grouped delegated preview may summarize the ready set.
This preview rule never suppresses the mandatory per-slice DDR route line.

AMR remains an unpromoted experimental benchmark. Ordinary work never passes
`attempt_tier` or claims AMR savings.

### Step 2.6: Orchestrator Ownership Decision (DDR)

Record minimum tier, owner, owner reason, source preference, selected route,
host evidence, proof result, and available time/token telemetry. Missing
telemetry is `unavailable`, not zero. Complexity alone is not a current-owner
reason.

## Execute One Slice

1. Re-read the board and acquire/renew the slice lease.
2. Load `references/execution-prompt.md`; for plan-run worktrees also load
   `references/worktree-isolation.md`.
3. Verify affected regression snapshots.
4. Load expected files, impact evidence, protected oracles, and test obligation.
5. Classify each discovered path:
   - directly required/generated/test-convention: claim and record it;
   - architecture, dependency, migration, auth, destructive, or product-scope
     expansion: replan or ask the required human question;
   - unrelated cleanup: do not edit.
6. Execute under the selected owner.
7. Run the narrowest proof capable of failing for the intended behavior.
8. Invoke `kb-qa`; invoke `kb-functional-test` for cross-boundary or
   user-visible behavior.
9. Compare actual diff to forecast and record discovered/unused paths.
10. Write `slice-<id>-to-done`, board/manifest status, proof receipt, and memory
    impact in the same accepted slice commit.
11. Advance the plan-worktree receipt, then release the slice lease.

### Functional Proof

Frontend or UI-reachable changes require rendered UI verification using
Playwright or the authenticated browser transport: navigate, interact, assert
observable DOM state, and retain only needed evidence. Backend checks and
screenshots alone cannot replace that proof.

### Protected Oracles

Create or update behavior-defining tests/fixtures before implementation when
practical, record their hashes, and reject unexplained later changes.

### System Effects

Trace callbacks, middleware, persistence order, failure cleanup/idempotency, and
parallel interfaces. Skip only for a proven leaf-node change.

### Destructive Guard

Never run recursive/forced deletion, hard reset, force push, database
drop/truncate, bulk deletion, or production-config overwrite without explicit
human authorization. Use exact resolved paths for authorized cleanup.

## Failure Handling

| Situation | Action |
|---|---|
| Slice execution fails with progress possible | Keep `in_progress`; run bounded repair and retry proof |
| Slice execution fails with no safe progress | Recheck sensor; record owner, scope, reason, resume condition, and propagation |
| Human-only authority/input/judgment | `human-required`; continue unrelated work |
| User pause | Preserve technical status and stop mutation |
| Claim collision | Serialize or requeue; do not race |

Do not turn an agent-owned code, test, controller, UI, or reproducibility
failure into user work while safe repair remains.

## Completion

When all slices are done/skipped:

1. Run/reuse one exact-tree aggregate.
2. Populate `scope-verified-files`.
3. Write and check `work-to-complete: passed` with
   `allowed_next_action: kb-finalize <manifest>`.
4. Refresh durable project memory when a slice recorded durable impact.
5. Archive completed board work.
6. Complete the plan-worktree receipt and release the plan-run lease.
7. Invoke `kb-finalize <manifest>` automatically.

`kb-work` never merges or pushes a resolved default branch. Missing delivery
policy is local-only. PR/manual delivery belongs to `kb-ship`; only authorized
`kb-land` integrates a remote default branch.

## Lazy References

- `references/execution-prompt.md` - worker packet and return contract.
- `references/worktree-isolation.md` - manifest-owned commit acceptance.
- `references/go-sandbox.md` - Go environment inside a worker sandbox.
