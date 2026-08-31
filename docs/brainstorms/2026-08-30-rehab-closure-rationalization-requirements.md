---
date: 2026-08-30
topic: rehab-closure-rationalization
brainstorm_style: kb-brainstorm
---

# Rehab Closure and DDR Rationalization

## Problem Frame

`kb-rehab` exposed three forms of false work:

- a stale queue claim protected 163 hours of abandoned changes;
- `work-reality --action remove` rewrote active and pending rows as `blocked`
  when it could not prove they were removable;
- PR #36 mixes an unproven epistemic-evaluation framework, unrelated Windows
  ACL changes, and a documentation-only DDR claim that fails the repository
  gate.

The outcome is not more enforcement. It is a smaller, truthful workflow that
does not preserve abandoned work forever, invent blocked status, or turn
unproven model behavior into a release gate.

## Research Summary

**Findings that shaped requirements:**

- `internal/reconcile/authority.go` already owns `ResolveRemoteAuthority`;
  `InventoryOptions.RefreshRemoteAuthority`, `kbreconcile --refresh-authority`,
  and `cmd/kbcheck/work_reality.go` already consume it. The six session TODOs
  proposing another shared authority path duplicate shipped behavior.
- `applyWorkRealityMarks` writes `rehabMarkerBlocked` for every uncontained or
  non-terminal row when removal is requested. A real run converted pending and
  in-progress P0 rows to blocked and appended tool noise to their links.
- PR #36's evaluator adds 1,000+ lines across the dispatch surface and
  evaluation corpus, but its new runtime-identity test fails. It has no
  captured baseline or demonstrated promotion result.
- The same PR's README change claims complexity-based selection, but its
  `cmd/kbrouter/dispatch.go` diff adds only evaluator flags, actor-workspace
  handling, sealed-input checks, and captured evaluator output. It does not
  implement route selection. `ddr_contract_test.go` detects that the claim
  also omits the existing no-second-local-route invariant.

**Confidence:** High. Each finding comes from executed tests or the exact
runtime and mutation paths in this repository.

## Requirements

**Truthful rehabilitation**

- R1. `work-reality --action remove` removes only a `dead` or `superseded`
  `todo.md` row with containment proof and a resolving artifact in the
  authoritative default tree.
- R2. When a row is not removable, `--action remove` reports it as preserved
  without changing any byte of that row. It must not change its status marker,
  append an explanation, or turn pending/in-progress work into blocked.
- R3. `--action mark` retains its existing behavior for terminal rows. The
  change is limited to removal-mode nonterminal or uncontained paths.

**No unsupported DDR delivery**

- R4. Do not land PR #36's documentation-only complexity-selection claim.
  Preserve the existing, already-consistent DDR policy and its
  no-second-local-route invariant.

**Cut unproven machinery and close delivery**

- R7. Do not land the PR #36 epistemic evaluator, its sealed corpus, or its
  runtime-identity requirement. They remain a separately recoverable commit
  only if a future owner supplies a bounded use case and a passing baseline.
- R8. Do not land the unrelated Windows ACL change from PR #36 as part of DDR
  or rehabilitation work.
- R8. Close PR #36 as rejected after the work-reality replacement merges. Its
  commits remain recoverable by SHA, but it supplies no independently proven
  behavior to the replacement.
- R9. The stale session TODOs for a new remote-authority subsystem are closed
  as already satisfied; no replacement code is written.

## Success Criteria

- A non-removable `pending` or `in_progress` fixture row is byte-for-byte
  identical before and after `--action remove`.
- A removable terminal fixture row is still removed.
- `go run ./cmd/kbcheck core` passes on the delivery branch.
- The replacement PR contains no code or documentation copied from PR #36.
- PR #36 is closed as rejected only after the work-reality repair merges.
- The repository finishes on `main`, clean and level with `origin/main`.

## Scope Boundaries

- No Opus-, Claude-, or provider-specific block. A historical strict-schema
  failure is not evidence of a current universal runtime defect.
- No model-quality or reasoning gate, live paid run, benchmark, or
  runtime-identity framework.
- No new remote-authority abstraction: the existing shared implementation is
  retained and its stale task list is corrected.
- No silent deletion of the 32 unpaired work declarations. They remain
  preserved until specific terminal or replacement evidence exists.
- No unrelated Windows storage-ACL changes.

## Key Decisions

- Keep the existing DDR contract unchanged. The candidate does not implement
  the claimed first-owner selection, so retaining its README wording would
  turn documentation into a false behavior claim. Evidence: the exact
  `cmd/kbrouter/dispatch.go` PR #36 diff.
- Cut the entire PR #36 evaluator delivery: it is a large new gate with no
  established operational benefit, a failing self-test, and no separable DDR
  behavior.
- Change removal-mode preservation from mutation to reporting: a tool cannot
  truthfully convert missing proof into a lifecycle status. Evidence: real
  `todo.md` corruption reproduced by the rehab run.

## Dependencies / Assumptions

- [safe-assumption] Closing the six remote-authority session TODOs requires no
  repository edit. Reversible because they are session-local task metadata;
  `ResolveRemoteAuthority` plus the opt-in inventory and CLI surfaces are
  already tested.

## Alternatives Considered

- Repair and land the entire epistemic evaluator: rejected. Its baseline and
  promotion claims remain unproven, while repair would expand the router and
  gate surface solely to validate a premise not needed for rehabilitation.
- Extract the claimed DDR change: rejected. The candidate's router diff does
  not contain a route-selection change to extract.
- Keep remarking non-removable rows as blocked: rejected. Absence of removal
  proof is preservation evidence, not a truthful statement that the work is
  blocked.

## Slice Candidates (advisory for /kb-plan)

- Make removal-mode preservation non-mutating - an attempted cleanup cannot
  corrupt active work state.
- Deliver the removal repair and retire the mixed salvage PR - no stranded
  branch or open PR remains after a clean replacement.

## Outstanding Questions

### Resolve Before Planning

None.

### Parked / Out of Scope

- [parked][Affects R7] A future epistemic-evaluation experiment. Forbidden
  claim: its existence or a historical model complaint proves a release gate
  should block normal planning.
- [parked][Affects R10] The 32 unpaired declarations. Forbidden claim: lack
  of a paired ref means they are dead or blocked.

## Next Steps

-> /kb-plan
