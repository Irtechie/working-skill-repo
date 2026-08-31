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
- PR #36 mixes a useful complexity-based DDR direction with an unproven
  epistemic-evaluation framework, unrelated Windows ACL changes, and a
  documentation contradiction that fails the repository gate.

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
- The same PR's README change correctly moves selection toward slice
  complexity, but omits the existing invariant that local failure does not
  trigger a second local route. `ddr_contract_test.go` detects that conflict.

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

**Complexity-aware DDR without roulette**

- R4. DDR chooses an execution owner from currently qualified routes using the
  slice's required capability and complexity, never provider prestige, a model
  name, or a durable provider list.
- R5. A failed local route returns to the parent for a new explicit ownership
  decision; it must not silently try a second local route or downgrade.
- R6. `README.md`, `.github/skills/kb-work/SKILL.md`,
  `docs/context/architecture/kb-workflow.md`, and
  `cmd/kbcheck/ddr_contract_test.go` state and test one consistent policy.

**Cut unproven machinery and close delivery**

- R7. Do not land the PR #36 epistemic evaluator, its sealed corpus, or its
  runtime-identity requirement. They remain a separately recoverable commit
  only if a future owner supplies a bounded use case and a passing baseline.
- R8. Do not land the unrelated Windows ACL change from PR #36 as part of DDR
  or rehabilitation work.
- R9. Extract only the candidate DDR behavior that passes its focused tests,
  place it on this objective's branch, and close PR #36 as superseded after
  the replacement is merged.
- R10. The stale session TODOs for a new remote-authority subsystem are closed
  as already satisfied; no replacement code is written.

## Success Criteria

- A non-removable `pending` or `in_progress` fixture row is byte-for-byte
  identical before and after `--action remove`.
- A removable terminal fixture row is still removed.
- Targeted DDR tests and `go run ./cmd/kbcheck core` pass on the delivery
  branch.
- The replacement PR contains neither epistemic evaluator/corpus paths nor
  `internal/modelrouting/storage_acl_windows.go`.
- PR #36 is closed as superseded only after a merged replacement preserves the
  approved DDR behavior.
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

- Keep complexity-aware DDR; keep the no-second-local-route invariant:
  complexity decides the first qualified owner, while failure recovery stays
  explicit and bounded. Evidence: current DDR contract test and PR #36 diff.
- Cut the evaluator from this delivery: it is a large new gate with no
  established operational benefit and a failing self-test. Evidence:
  `cmd/kbcheck/skill_eval_epistemic_test.go` on PR #36.
- Change removal-mode preservation from mutation to reporting: a tool cannot
  truthfully convert missing proof into a lifecycle status. Evidence: real
  `todo.md` corruption reproduced by the rehab run.

## Dependencies / Assumptions

- [safe-assumption] The PR #36 dispatcher changes can be separated from the
  evaluator changes. Reversible because the original PR remains intact until
  the replacement merges. Evidence/proof: use a fresh topic branch, explicit
  file selection, targeted `cmd/kbrouter` tests, and `core`.
- [safe-assumption] Closing the six remote-authority session TODOs requires no
  repository edit. Reversible because they are session-local task metadata;
  `ResolveRemoteAuthority` plus the opt-in inventory and CLI surfaces are
  already tested.

## Alternatives Considered

- Repair and land the entire epistemic evaluator: rejected. Its baseline and
  promotion claims remain unproven, while repair would expand the router and
  gate surface solely to validate a premise not needed for rehabilitation.
- Permit a second local route after failure: rejected. It creates provider
  roulette and contradicts the existing parent-owned recovery contract.
- Keep remarking non-removable rows as blocked: rejected. Absence of removal
  proof is preservation evidence, not a truthful statement that the work is
  blocked.

## Slice Candidates (advisory for /kb-plan)

- Make removal-mode preservation non-mutating - an attempted cleanup cannot
  corrupt active work state.
- Extract and prove complexity-aware DDR - retain useful routing behavior with
  bounded failure recovery.
- Deliver the reduced replacement and retire the mixed salvage PR - no stranded
  branch or open PR remains after a clean replacement.

## Outstanding Questions

### Resolve Before Planning

None.

### Deferred to Planning

- [defer-to-planning][Affects R9][Technical] Identify the minimum PR #36
  dispatcher and test files that prove R4-R5 without carrying evaluator code.

### Parked / Out of Scope

- [parked][Affects R7] A future epistemic-evaluation experiment. Forbidden
  claim: its existence or a historical model complaint proves a release gate
  should block normal planning.
- [parked][Affects R10] The 32 unpaired declarations. Forbidden claim: lack
  of a paired ref means they are dead or blocked.

## Next Steps

-> /kb-plan
