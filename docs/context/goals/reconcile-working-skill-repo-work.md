# Reconcile Working Skill Repo Work

Status: complete
Created: 2026-07-27
Last updated: 2026-07-27

## Objective

Reconcile all remaining valuable `Irtechie/working-skill-repo` work into clean,
reviewable check-ins while separating stop-control product work from
AMR/DDR/EDDR benchmark experiments.

## Done Criteria

- Both configured `Irtechie/working-skill-repo` projects and their sessions are inventoried.
- Safe valuable source, tests, and docs are verified and delivered on non-default branches under repo policy.
- Stop-control product work is reviewed independently from benchmark experiments.
- Experimental AMR/DDR/EDDR work is either delivered as an explicit experiment or parked with exact paths and a durable handoff.
- No delivered check-in or remote preservation ref contains credentials, host-local catalogs, generated run state, or private endpoints.
- Owner-created worktrees and branches are preserved, and completed worktrees are clean.
- Final proof records exact refs, PRs, parked paths, blockers, and verification commands.

## Terminal Proof

- `git diff --check` passes on the reconciliation branch.
- Targeted Windows containment build, vet, and repeated tests pass on integrated `origin/main`.
- The reconciliation branch has a matching upstream SHA and ready PR.
- EDDR planning state has an immutable remote preservation ref with exact paths.
- All nonexperimental owner worktrees are clean; the active cohort is the sole explicit exception.

## Done Check

- Type: gate
- Check: `reconciliation-proof-v1`
- Expected result: every Work Unit is complete or the explicit EDDR exception, all Blocker resume conditions are exact, and Terminal Proof receipts are recorded
- Why sufficient: combines repository delivery, worktree preservation, targeted product proof, and the authorized experiment boundary without treating the unrelated broad harness failure as a false product regression.

## Current State

- Current artifact: `docs/context/goals/reconcile-working-skill-repo-work.md`
- Next allowed action: none
- Last proof: stop-control duplicate cleaned; targeted Windows containment proof passed; EDDR planning ref pushed; PR #2 ready with matching local/upstream SHA

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|
| Stop-control product work | kb-fix -> kb-start inventory | `docs/handoffs/done/2026-07-27-redundant-stop-control-snapshot.md` | complete | product work is on `origin/main` at `a495444`; owner worktree clean |
| Convergence stop-control branch | kb-start inventory | `deaderestpool-convergence-stop-controls` | complete | clean session, no commits, no diff |
| AMR/DDR/EDDR exception | owner-controlled cohort | `docs/handoffs/parked/2026-07-27-eddr-experimental-state.md` | parked | current 15-modified/10-untracked-entry owner snapshot and immutable planning ref |
| Duplicate project registration | kb-start inventory | duplicate registration | parked | owns no sessions; metadata removal unsafe while its path is an active worktree |
| Goal evidence and memory | kb-goal | this ledger and `todo-done.md` | complete | PR #2, refs, commands, parked paths, and blockers recorded |

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|
| None | - | - | - |

## Notes

- The canonical project owns every active session. The duplicate registration owns none but points at the active benchmark worktree; do not use or delete it until metadata-only removal is proven safe.
- Windows containment proof on `origin/main`: `go vet ./cmd/kbrouter`, `go build ./cmd/kbrouter`, two repeated `TestC1WindowsJobObjectKillsGrandchild` runs, the combined containment/configuration tests, and the full `-run Windows` package selection all exited 0.
- Harness failure is separate: `go test ./cmd/kbcheck -run '^TestPlanWorktreeSelftestExercisesDisposableLifecycle$' -v -count=1 -timeout 90s` exits 1 after timing out while blocked in `gitOutput`. It blocks `core`/`local-release`, not the independently passing containment proof.
- Required `kb-plan` global drift is part of the active EDDR exception: canonical repo SHA-256 prefix `8DD84B239838`; all three experimental global copies use prefix `D0610845BBB8`. Do not overwrite either side until the cohort owner releases the experiment.
- Delivery: PR #2, branch `deaderestpool-reconcile-skill-work`. The active cohort remains independently owner-controlled; this goal's completion parks its exact snapshot without stopping or claiming completion of that experiment.
- Merge boundary: PR #2 is ready independently of DDR. It contains only reconciliation records and the parked-exception contract; it does not contain the active benchmark state, generated evidence, experimental `kb-plan` copy, or model runtime state.
- Qwen's prescribed DDR acquire is terminal `INFRASTRUCTURE NOT-RUN` after the launcher failed before allocation. No reservation, cleanup target, endpoint, retry obligation, or non-DDR work remains from that attempt.
- Luna resumed from durable state with all four work items complete; `go test ./...` and `go vet ./...` pass in its layered-access-policy workspace. No DDR proof gate remains for PR #2 convergence.
