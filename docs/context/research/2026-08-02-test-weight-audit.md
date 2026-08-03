# Test weight audit

`date: 2026-08-02`

Per-test coverage and cost measurement across the four packages that account
for 97% of suite runtime. Produced to answer one question with evidence rather
than judgment: **what can be reduced without losing functionality?**

## Method

For each package: build one instrumented test binary
(`go test -c -covermode=count`), then run every test *alone* with its own
coverprofile, executed from the package directory. Attribute each covered
block to the set of tests that cover it. A block is **unique** to a test when
no other test in the package covers it.

The safety property this buys: a test whose covered blocks are all covered by
someone else cannot, by deletion alone, reduce statement coverage.

### What this measurement cannot tell you

Coverage identity is not behavioural identity. Two tests can execute the same
`return err` line while proving that two *different* inputs are rejected.
Statement coverage is also blind to tests that drive PowerShell scripts,
external binaries, or OS process semantics -- those report near-zero covered
statements while proving something no Go-level test does.

So the measurement **nominates**; it does not decide. Every verdict below that
removes or restructures a test carries a stated reason beyond its numbers.

### Reproduce

```text
go test -c -covermode=count -o <bin> <pkg>
<bin> -test.run "^<Name>$" -test.coverprofile <p>   # run from the package dir
```

Running the compiled binary from the repo root instead of the package
directory breaks every test that resolves the repo root via `../..`. That
produced 18 phantom failures on the first pass of this audit.

## Headline numbers

| Measure | Value |
|---|---|
| Tests measured | 447 |
| Failing when run alone | 0 (no hidden inter-test dependencies) |
| Contended wall time | 563.1s |
| Tests with zero unique statements | 161 |

Cost is extremely concentrated:

| Slice | Time | Share |
|---|---|---|
| Top 10 tests | 193.9s | 34% |
| Top 25 tests | 318.3s | 57% |
| Top 50 tests | 428.4s | 76% |
| Top 100 tests | 504.6s | 90% |
| All 279 tests under 200ms | 32.8s | 6% |

> Wall times are measured under contention and inflate roughly 3-4x versus
> serial execution. Use them to rank, not as absolute budgets.

## The trap: deletion is the wrong lever

The obvious move is to delete the 161 zero-unique-coverage tests. The numbers
say do not.

| If we deleted... | Tests lost | Time saved |
|---|---|---|
| all zero-unique tests | 161 | 133.5s |
| ...of which cost under 1s | 137 | 23.7s |
| ...of which cost over 1s | 24 | 109.8s |

Deleting 137 tests to save 23.7s is a bad trade: it surrenders 137 behavioural
assertions for roughly the cost of a single expensive test. The remaining
24 expensive zero-unique tests are worth 109.8s -- but they are expensive because
of **fixture setup and deliberate sleeps**, not because they assert nothing.

The lever is therefore *setup cost*, not test count.

## Ranked reduction targets

### 1. Share the git fixture across `TestTerminalCleanup*` (largest win)

28 tests, 183.1s contended, 68 unique statements between them.
Each covers 178-464 statements while contributing 0-10 unique ones: they walk
the same path with different inputs, which is exactly what a well-factored
edge-case suite looks like.

Every one of them calls `createTerminalCleanupWorktree`, which spawns roughly
eight git processes before a single assertion runs: `init --bare`,
`remote add`, `push -u`, `symbolic-ref`, `worktree add -b`, `add`, `commit`,
`rev-parse`. The file makes 92 git invocations across 28 tests.

**Action:** build the fixture once and clone/copy it per test, or table-drive
the family. Keeps all 28 edge cases and every assertion; pays git setup once.
Do not delete these tests.

### 2. Retune sleep-bound process-kill tests

`TestRunProcessCheckTimeoutKillsGrandchild` (18.2s) proves that a killed parent
does not orphan its grandchild. Auditing it surfaced a **logic flaw, not just a
cost problem**: the timeout fired at t=15s, the survival check ran at t=18s, but
the grandchild did not write its sentinel until t=25s. A surviving grandchild
was still sleeping at check time, so the assertion passed whether or not
containment worked.

Verified experimentally before changing anything: extending the settle window to
t=30s (past the 25s write) still passed, which proves containment genuinely
works. The production behaviour was correct; only the test was blind.

Rewritten around an explicit timing contract -- the sentinel write must land
*after* the timeout but *before* the check -- with named constants and a comment
so it cannot silently regress to blindness. Verified by mutation: making the
grandchild write before the kill now fails the test.

**Result: 18.2s -> 11.6s, and the assertion is now load-bearing.**

`TestC1WindowsJobObjectKillsGrandchild` (6.2s, `cmd/kbrouter`) was checked for
the same flaw and **is correctly designed**: timeout 120ms, grandchild writes at
700ms, check at ~1020ms, so a survivor is observable. Its six attempts are
deliberate race-detection. **No change.**

### 3. Investigate the single most expensive test

`TestModelTierEvalDeterministicCorpus` -- 42.9s, 500 covered, 42 unique. It is
the most expensive test in the repo by a wide margin. Determine whether the
corpus size is load-bearing for determinism or merely inherited.

### 4. Harness-level wins (outside the tests themselves)

Recorded here because they dominate the gate and require no test changes:

| Change | Saving |
|---|---|
| Collapse 3 sequential `go test` invocations into 1 | ~120s (see caveat) |
| Drop `kbrouter-catalog-tests` (re-runs what `go test ./...` already ran) | 4.9s (measured) |
| Run sub-second native checks concurrently in `runCore` | ~20s |
| Record `DurationMS` per check | 0s, but prevents cost going dark again |

Caveat on the first row: the two isolated packages (`cmd/kbcheck`,
`cmd/kbrouter`) deliberately run outside the outer job object because they own
child-process containment. They cannot simply be merged into the single
`go test ./...` invocation. They *can* run concurrently with each other and with
the regular batch, but the memory-derived parallelism budget must then be
divided between them or the commit-charge exhaustion this repo already hit
returns. `cmd/kbcheck` at 133s is the long pole either way, so reducing it is
the prerequisite for that win.

## Applied so far

| Change | Before | After | Verified by |
|---|---|---|---|
| `TestRunProcessCheckTimeoutKillsGrandchild` timing contract | 18.2s, blind assertion | 11.6s, mutation-proven | mutation test fails as required |
| Removed duplicate `kbrouter-catalog-tests` check | 4.9s | 0s | 19 re-run tests all covered by `go test ./...` |

## Verdict legend

| Verdict | Meaning |
|---|---|
| `keep` | No action. Cheap, or carries coverage/behaviour nothing else does. |
| `share-fixture` | Keep every assertion; move setup cost out of the per-test path. |
| `retune` | Keep the test; reduce a deliberate duration without weakening the proof. |
| `investigate` | Cost unexplained by this measurement; needs a look before acting. |

## Per-test data

Sorted by cost. `unique` = statements covered by this test and no other in its
package.

### Expensive (>= 1s) -- 94 tests

| sec | covered | unique | verdict | test | why |
|---:|---:|---:|---|---|---|
| 42.87 | 500 | 42 | investigate | TestModelTierEvalDeterministicCorpus<br>cmd/kbcheck | single most expensive test at 42.9s; 500 covered / 42 unique. Determine if corpus size is load-bearing. |
| 26.7 | 138 | 25 | keep | TestGoTestsIsolateChildProcessOwnersAndPropagateFailure<br>cmd/kbcheck | expensive but carries unique coverage. |
| 20.5 | 482 | 3 | share-fixture | TestPlanRunAdvanceRejectsWrongIdentityDirtyStateAndStaleIntegrationHead<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 18.16 | 68 | 0 | retune | TestRunProcessCheckTimeoutKillsGrandchild<br>cmd/kbcheck | proves OS process-tree kill; 0 stmt-coverage is a metric blind spot, not absence of value. Cost is deliberate time.Sleep(25s/30s). Shrink timeout+sleeps proportionally. |
| 16.76 | 464 | 9 | share-fixture | TestTerminalCleanupRequiresRemoteContainmentAndRetainsPRRefs<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 16.64 | 621 | 8 | share-fixture | TestApplyFreshEvidencePreservesDriftAndDirt<br>internal/reconcile | expensive, low unique yield; inspect setup cost before assertions. |
| 14.71 | 426 | 1 | share-fixture | TestTerminalCleanupRefreshesRemoteEvidenceBetweenReceipts<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 12.71 | 1013 | 147 | keep | TestPlanWorktreeSelftestExercisesDisposableLifecycle<br>cmd/kbcheck | 147 unique stmts, highest unique yield in the repo. |
| 12.42 | 743 | 15 | keep | TestApplyRefGuards<br>internal/reconcile | expensive but carries unique coverage. |
| 12.4 | 411 | 0 | share-fixture | TestTerminalCleanupSweepUsesStableRootWhenRootEqualsTarget<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 11.84 | 694 | 5 | share-fixture | TestPlanRunAdvanceRejectsMissingOrFalseClaimEvidence<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 10.2 | 382 | 0 | share-fixture | TestTerminalCleanupRejectsRewrittenRemoteDefaultEvidence<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 9.99 | 943 | 122 | keep | TestPlanRunAdvanceAcceptsSequentialSliceCommitsWithIntegrationHeadCAS<br>cmd/kbcheck | high unique coverage. |
| 9.44 | 324 | 3 | share-fixture | TestTerminalCleanupRejectsLockedAndMissingWorktrees<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 8.99 | 369 | 1 | share-fixture | TestTerminalCleanupResumePacketDigestDriftBlocksAwaitingReviewRetirement<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 8.69 | 395 | 0 | share-fixture | TestTerminalCleanupKeepsBlockedReceiptAssociation<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 8.54 | 205 | 1 | share-fixture | TestTerminalCleanupUsesAuthoritativeRemoteDefault<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 8.09 | 405 | 7 | share-fixture | TestTerminalCleanupReconcilesEmptyPartialRemoval<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 7.89 | 411 | 0 | share-fixture | TestTerminalCleanupDirectIntegrationDeletesOnlyMergedLocalRef<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 7.65 | 398 | 2 | share-fixture | TestTerminalCleanupDefersCurrentActiveAndDirtyWorktree<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 7.18 | 290 | 10 | share-fixture | TestTerminalCleanupSerializesAgainstQueueClaims<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 7.14 | 383 | 4 | share-fixture | TestTerminalCleanupRejectsNonEmptyPartialResidual<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 6.45 | 755 | 30 | keep | TestApplyIntegratedRef<br>internal/reconcile | expensive but carries unique coverage. |
| 6.24 | 94 | 0 | retune | TestC1WindowsJobObjectKillsGrandchild<br>cmd/kbrouter | same class as above: job-object kill proof, sleep-bound cost. |
| 6.1 | 354 | 1 | share-fixture | TestTerminalCleanupRejectsPartialBranchIdentityMismatch<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 5.93 | 0 | 0 | keep | TestNoKBRepoBinaryDryRun<br>cmd/kbreconcile | exercises external binary; Go coverage blind. Verify cost source before touching. |
| 5.75 | 336 | 1 | share-fixture | TestTerminalCleanupRejectsPartialReceiptIdentityMismatch<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 5.66 | 354 | 1 | share-fixture | TestTerminalCleanupPreservesIgnoredFiles<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 5.62 | 623 | 20 | keep | TestApplyRepairsOnlyExactEmptyResidual<br>internal/reconcile | expensive but carries unique coverage. |
| 5.58 | 354 | 1 | share-fixture | TestTerminalCleanupRejectsReplacedWorktreeGeneration<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 5.5 | 705 | 29 | keep | TestVerifyKeepsOutcomeDimensionsSeparate<br>internal/reconcile | expensive but carries unique coverage. |
| 5.31 | 178 | 0 | share-fixture | TestTerminalCleanupChecksEveryAuthoritativeRemoteDefault<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 5.27 | 678 | 3 | share-fixture | TestApplyReceiptRequiresExactPlanCoverageAndVerification<br>internal/reconcile | expensive, low unique yield; inspect setup cost before assertions. |
| 5.06 | 327 | 3 | share-fixture | TestTerminalCleanupRejectsMovedWorktree<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 5.01 | 178 | 0 | share-fixture | TestTerminalCleanupRejectsNewAuthoritativeRemoteDefaultTarget<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 4.88 | 127 | 0 | share-fixture | TestDefaultBranchBoundaryRejectsLocalAndRemoteDefaultInternalTargets<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 4.82 | 342 | 8 | share-fixture | TestTerminalCleanupRejectsBrokenAdminRoundTrip<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 4.38 | 310 | 1 | share-fixture | TestTerminalCleanupPreservesCurrentSessionByIdentity<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 4.16 | 345 | 29 | keep | TestWorktreeIntegrateAndReleaseRequiresCleanIntegratedWorktree<br>cmd/kbcheck | expensive but carries unique coverage. |
| 4.08 | 306 | 3 | share-fixture | TestDirtyBaseAuthorityBlocksRelevantWIPAndPreservesUnrelatedDirt<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 3.78 | 622 | 7 | share-fixture | TestSliceCommitRequiresCoordinatorProofReplayAndAggregateSuccess<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 3.78 | 153 | 2 | share-fixture | TestCargoStorageRejectsPhaseAndRunSpecificTemporaryTargets<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 3.6 | 0 | 0 | keep | TestWorkQueueUpdateMigratesOwnedWorktreeIdentity<br>cmd/kbcheck | exercises a PowerShell script; Go coverage cannot see it. 0 covered_stmts is meaningless here. |
| 3.46 | 553 | 6 | share-fixture | TestApplyRejectsExpiredOrMismatchedPlanAndProtectedAction<br>internal/reconcile | expensive, low unique yield; inspect setup cost before assertions. |
| 3.32 | 555 | 3 | share-fixture | TestApplySerializesOnCompatibleRepositoryLock<br>internal/reconcile | expensive, low unique yield; inspect setup cost before assertions. |
| 3.32 | 496 | 1 | share-fixture | TestApplyEnforcesPlanExecutionAuthorization<br>internal/reconcile | expensive, low unique yield; inspect setup cost before assertions. |
| 3.18 | 674 | 3 | share-fixture | TestApplyRetiresCleanWorktreeAndIsIdempotent<br>internal/reconcile | expensive, low unique yield; inspect setup cost before assertions. |
| 3.18 | 295 | 2 | share-fixture | TestWorktreeIntegrateRequiresOwnerAndStableBase<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 2.92 | 168 | 2 | share-fixture | TestTerminalCleanupFailsClosedWhenRemoteDefaultIsUnresolved<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 2.57 | 145 | 3 | share-fixture | TestPlanRunWorkspaceAdoptRejectsPrimaryDefaultAndDirtyCheckouts<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 2.54 | 204 | 3 | share-fixture | TestPlanRunWorkspaceRejectsDefaultBranchOwnerMismatchAndUnsafeRelease<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 2.54 | 203 | 1 | share-fixture | TestPlanRunWorkspaceReleaseKeepsHarnessOwnedWorktree<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 2.47 | 242 | 8 | share-fixture | TestInventoryNoKBRepo<br>internal/reconcile | expensive, low unique yield; inspect setup cost before assertions. |
| 2.46 | 97 | 0 | share-fixture | TestCargoStorageResolveIsStableAcrossWorktrees<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 2.3 | 107 | 1 | share-fixture | TestCargoStorageRemapsAbsoluteTargetInsideAnyLinkedWorktree<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 2.17 | 281 | 4 | share-fixture | TestDeliveryOwnerDefaultsPRReadyAndKeepsPolicyOutsidePlanIntegration<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 2.1 | 275 | 0 | share-fixture | TestPlanRunWorkspacePrepareIsIdempotentAndPreservesDirtySource<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 2.07 | 179 | 2 | share-fixture | TestPlanRunWorkspaceAdoptsHarnessWorktreeWithoutNesting<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 2.02 | 130 | 2 | share-fixture | TestCargoStorageMigratesLegacyReceiptIdentities<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.98 | 265 | 0 | share-fixture | TestWorktreePreparePreservesDirtySourceAndWritesReceipt<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.9 | 66 | 0 | share-fixture | TestPlanWritesDurableStableJSON<br>cmd/kbreconcile | expensive, low unique yield; inspect setup cost before assertions. |
| 1.89 | 575 | 26 | keep | TestDDRAttemptFunctionalOutcomesAreBoundedAndNeverRetry<br>cmd/kbrouter | expensive but carries unique coverage. |
| 1.89 | 154 | 3 | share-fixture | TestTerminalCleanupFailsClosedWithoutAuthoritativeDefault<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 1.7 | 461 | 12 | keep | TestWorktreeCommandStatusJSON<br>cmd/kbcheck | expensive but carries unique coverage. |
| 1.68 | 146 | 39 | keep | TestApplyVerifyStableJSONAndFailClosedInput<br>cmd/kbreconcile | expensive but carries unique coverage. |
| 1.66 | 249 | 17 | keep | TestCargoStorageFinalizesOnlyOwnedContainedTemporaryTarget<br>cmd/kbcheck | expensive but carries unique coverage. |
| 1.64 | 108 | 2 | share-fixture | TestPlanRunPrepareBlocksWhenRemoteDefaultAuthorityIsUnresolved<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.64 | 94 | 0 | share-fixture | TestSessionPreserveCommitsOnSessionBranchWithoutPushing<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.62 | 4 | 0 | keep | TestTerminalCleanupLockIsCompatibleWithPowerShell<br>cmd/kbcheck | cross-runtime lock-format proof; only 4 stmts by design. |
| 1.61 | 89 | 0 | share-fixture | TestSessionPreserveIgnoresGitignoredFiles<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.55 | 33 | 2 | share-fixture | TestSessionPreserveRefusesMidMerge<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.54 | 103 | 0 | share-fixture | TestCargoStorageSeparatesRepositoriesUnderExternalCacheRoot<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.51 | 595 | 0 | share-fixture | TestCatalogDoesNotMintTrustOrSendCatalogNamedSecret<br>cmd/kbrouter | expensive, low unique yield; inspect setup cost before assertions. |
| 1.51 | 110 | 3 | share-fixture | TestApplyRejectsEmptyCurrentSessionForMutationPlan<br>cmd/kbreconcile | expensive, low unique yield; inspect setup cost before assertions. |
| 1.46 | 45 | 6 | share-fixture | TestUncontainedRunnerBoundsTimeoutOverflowAndInheritedPipes<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.42 | 17 | 0 | share-fixture | TestPlanRunLeaseStateRootSharesWorktreesButNotSeparateClone<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.41 | 98 | 1 | share-fixture | TestSessionPreserveExcludesBuildArtifactsButKeepsSource<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.4 | 458 | 4 | share-fixture | TestModelTierEvalTextOutputIsScoped<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.36 | 54 | 0 | share-fixture | TestNoKBRepoDryRun<br>cmd/kbreconcile | expensive, low unique yield; inspect setup cost before assertions. |
| 1.32 | 11 | 0 | share-fixture | TestSliceLeaseGitCommonDirCoordinatesWorktreesButNotClones<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.29 | 93 | 1 | share-fixture | TestSessionPreserveExcludesOversizedFiles<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.27 | 92 | 1 | share-fixture | TestSessionPreserveHandlesDeletions<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.26 | 148 | 2 | share-fixture | TestCargoStorageFinalValidationRejectsFalseAccounting<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.25 | 376 | 4 | share-fixture | TestCargoStoragePublicCLIFlow<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.2 | 92 | 1 | share-fixture | TestPlanRunWorkspaceRequiresExplicitLocalCommitAuthorization<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.17 | 143 | 2 | share-fixture | TestCargoStorageResolveResumesAuthoritativeReceipt<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.17 | 79 | 2 | share-fixture | TestSessionPreservePlanDoesNotMutate<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.17 | 750 | 29 | keep | TestCatalogOpenAICompatibleDiscoveryUsesTrustedRouteAndBoundedResponse<br>cmd/kbrouter | expensive but carries unique coverage. |
| 1.17 | 215 | 4 | share-fixture | TestCargoStorageReconcilesDeletionIntentAfterTargetDisappears<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1.13 | 84 | 2 | share-fixture | TestTerminalCleanupRefusesPrimaryAndDefaultTargets<br>cmd/kbcheck | cost is per-test git fixture setup (~8 process spawns), not assertions. Share/cache fixture; keep every assertion. |
| 1.08 | 561 | 0 | share-fixture | TestC1ConcurrentDispatchLockPreservesFailedHandoff<br>cmd/kbrouter | expensive, low unique yield; inspect setup cost before assertions. |
| 1.03 | 149 | 2 | share-fixture | TestCargoStorageFinalValidationRejectsTamperedCleanup<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1 | 213 | 5 | share-fixture | TestCargoStorageBlocksForgedOwnershipMarker<br>cmd/kbcheck | expensive, low unique yield; inspect setup cost before assertions. |
| 1 | 59 | 2 | keep | TestSessionPreserveIsNoOpOnCleanTree<br>cmd/kbcheck | moderate cost, no strong reduction signal. |

### Moderate (200ms - 1s) -- 78 tests

| sec | covered | unique | verdict | test | why |
|---:|---:|---:|---|---|---|
| 0.99 | 167 | 0 | keep | TestCargoStorageConcurrentRegistrationsAreMerged<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.98 | 339 | 32 | keep | TestPreSliceReviewContractRejectsInvalidReceipts<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.92 | 148 | 0 | keep | TestCargoStorageRejectsTemporaryTargetOutsideApprovedRoot<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.91 | 415 | 0 | keep | TestDirectDispatchCannotBypassDelegatedCapabilityEnvelope<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.91 | 0 | 0 | keep | TestPRReviewWorkbenchContract<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.88 | 103 | 1 | keep | TestCargoStorageNativeKeyMatchesPortableFallbackAndIsAppliedOnce<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.84 | 50 | 1 | keep | TestSessionPreserveRefusesOnDefaultBranch<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.82 | 759 | 1 | keep | TestFallbackUsesFreshFallbackRouteModelAndNeverDownward<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.82 | 103 | 0 | keep | TestCargoStorageKeysExternalAbsoluteCacheRootByRepository<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.81 | 289 | 16 | keep | TestCargoStorageNotApplicableCLIFlow<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.81 | 36 | 1 | keep | TestSessionPreserveRefusesBranchMismatch<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.79 | 34 | 1 | keep | TestSessionPreserveRefusesDetachedHead<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.73 | 626 | 7 | keep | TestDispatchTimeoutFiniteLedgerDirectProviderAndCustomUserRoot<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.71 | 22 | 1 | keep | TestCargoStorageReceiptNamesDoNotCollideAfterSanitizing<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.7 | 19 | 16 | keep | TestUpdateApprovedCatalogPreservesMixedEvidenceAndPromotionException<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.68 | 753 | 0 | keep | TestC1FallbackUsesPerAttemptArtifacts<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.68 | 10 | 0 | keep | TestCurrentWorktreeResolvesGitTopLevelFromSubdirectory<br>cmd/kbreconcile | moderate cost, no strong reduction signal. |
| 0.63 | 471 | 12 | keep | TestQualificationPlanContractRejectsWeakOrStaleRecords<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.61 | 5 | 0 | keep | TestCargoStorageRejectsSymlinkedTemporaryRoot<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.6 | 142 | 22 | keep | TestModelRoutingReleaseRejectsDishonestOrUnsafeEvidence<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.58 | 13 | 1 | keep | TestSessionPreserveRequiresSessionIdentity<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.46 | 737 | 7 | keep | TestDispatchCodexExecArgvProfileModelAndProofUnknown<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.45 | 443 | 12 | keep | TestPlanRunLeaseCommandReportsClaimOwner<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.44 | 716 | 5 | keep | TestDispatchPreservesAttemptAndPlannedCorrectionTiers<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.41 | 600 | 0 | keep | TestDispatchForgedChildEvidenceIsObservationOnlyAndDoesNotFallbackOnExitZero<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.39 | 316 | 0 | keep | TestCodexDiscoveryPromotesOnlyExactWindowsExecutableIdentity<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.39 | 618 | 2 | keep | TestDispatchEnvKeepsOnlyOSCodexHomeAndRouteAuth<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.37 | 360 | 9 | keep | TestSliceLeaseCommandStatusJSON<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.37 | 294 | 2 | keep | TestCodexDiscoveryLeavesModelsDiscoveredOnlyOnContractFailure<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.37 | 498 | 0 | keep | TestDDRAttemptConcurrentReservationAllowsOneNetworkAttempt<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.35 | 0 | 0 | keep | TestSliceLeaseTwoProcessSameSliceRace<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.34 | 129 | 14 | keep | TestFenceGatewayAuthorizationIdempotencyAndBypass<br>internal/reconcile | moderate cost, no strong reduction signal. |
| 0.34 | 534 | 1 | keep | TestDDRAttemptPendingProofReplayReturnsToParentWithoutRetry<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.32 | 308 | 0 | keep | TestCatalogCapturedCodexFixtureAndFingerprintRefresh<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.32 | 443 | 8 | keep | TestC2DispatchRevalidatesAdapterPriorExecutableRevisionBeforeLaunch<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.3 | 354 | 1 | keep | TestDDRAttemptRequiredReservationReplayBlocksWithoutRetry<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.3 | 292 | 1 | keep | TestModelsSelectLoadsSavedProjectPriorityUnlessRunPreferenceOverrides<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.3 | 475 | 2 | keep | TestC1ContainmentUnavailableDoesNotStartOrFallback<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.29 | 319 | 1 | keep | TestC1TrustedStateRejectsExpiredState<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.29 | 262 | 3 | keep | TestC1RejectsDuplicateAndReservedDispatchArtifacts<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.29 | 383 | 0 | keep | TestApprovedPrivateRouteSurvivesRedactedRunCatalogSelection<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.29 | 417 | 2 | keep | TestDispatchRejectsForgedRouteStateAndPathEscapes<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.29 | 324 | 1 | keep | TestDDRAttemptRequirePinBlocksInsteadOfReturningToParent<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.28 | 232 | 3 | keep | TestDDRAttemptRejectsInvalidTierAndImplicitSensitivity<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.27 | 315 | 5 | keep | TestConcurrentApprovalPreservesInterveningDenial<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.27 | 0 | 0 | keep | TestWebUIProofRemainsAgentOwned<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.27 | 482 | 0 | keep | TestDDRAttemptDefaultApprovalModeDoesNotRequireTrustReceipt<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.26 | 0 | 0 | keep | TestPlanRunLeaseTwoProcessClaimRaceHasOneMutationAuthority<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.26 | 421 | 0 | keep | TestDeniedHostedRouteIsNeverProbed<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.25 | 286 | 2 | keep | TestDispatchRejectsPacketWithoutCodexHarnessAuthority<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.25 | 158 | 9 | keep | TestPlanRunManifestContractRejectsIncompleteWorkspaceIntent<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.25 | 377 | 1 | keep | TestExecutionTelemetryCommandUsesReceiptEnvelopeEvidence<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.25 | 267 | 3 | keep | TestModelsSelectRejectsVagueCurrentOwnerReason<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.25 | 329 | 5 | keep | TestCatalogDiscoverWritesOnlyRunCatalogAndBoundsSlowAdapter<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.24 | 156 | 0 | keep | TestManifestContractRejectsEachUnsafeRoutingMutation<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.24 | 254 | 0 | keep | TestImpactInvalidationFansOutAcrossSharedDependency<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.24 | 416 | 0 | keep | TestC2DispatchRejectsVisibleCurrentRouteBeforeLaunch<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.24 | 540 | 3 | keep | TestDDRRejectsConflictingAttemptAndProofReplays<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.23 | 322 | 4 | keep | TestPlanRunLeaseComposesWithSliceClaimsAndStatesLocalLimitation<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.23 | 290 | 1 | keep | TestDispatchCorrectionPacketRefusesBeforeWorkerLaunchOrReceipt<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.23 | 195 | 0 | keep | TestCodexDiscoveryLeavesCLIModelsInformativeWithoutContainmentProof<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.23 | 261 | 1 | keep | TestModelsSelectReportsExplicitAttemptAndPlannedCorrectionTiers<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.22 | 275 | 55 | keep | TestGraphRoutingEvalRequireReadyPasses<br>cmd/kbcheck | high unique coverage. |
| 0.22 | 192 | 0 | keep | TestGraphRouteCommandRejectsStalePacket<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.22 | 122 | 6 | keep | TestProofAcceptsRedThenGreenTrace<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.22 | 532 | 4 | keep | TestDDRRequiredProofFailureBlocksWithExactExit<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.22 | 296 | 3 | keep | TestProjectPolicyErrorFailsConfiguredProbeClosed<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.22 | 369 | 4 | keep | TestDispatchRequiresMarkedRunCatalogAndRejectsArbitraryExec<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.22 | 116 | 35 | keep | TestClearResetsMatchingScope<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.21 | 324 | 2 | keep | TestDispatchRejectsProjectPathCodexExecutable<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.21 | 372 | 1 | keep | TestDDRAttemptHonorsProjectAliasPolicyBeforeNetwork<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.21 | 182 | 2 | keep | TestProofReceiptRejectsFailedPartialUnknownCoverageAndSemanticDrift<br>cmd/kbcheck | moderate cost, no strong reduction signal. |
| 0.2 | 101 | 3 | keep | TestModelsImportRejectsPlaceholdersUnsupportedCapabilityAndDuplicateTools<br>cmd/kbrouter | cheap; deleting saves nothing measurable. |
| 0.2 | 353 | 0 | keep | TestDDRAttemptReservationPreventsRetryAfterUncertainDispatch<br>cmd/kbrouter | cheap; deleting saves nothing measurable. |
| 0.2 | 324 | 0 | keep | TestDDRAttemptRequiresRouteApprovalForHostedBoundary<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.2 | 307 | 1 | keep | TestDispatchRequiresPrivateRouteState<br>cmd/kbrouter | moderate cost, no strong reduction signal. |
| 0.2 | 319 | 4 | keep | TestTrustIsSeparateExplicitProjectBoundAndRevocable<br>cmd/kbrouter | cheap; deleting saves nothing measurable. |
| 0.2 | 304 | 0 | keep | TestExecutionTelemetryRejectsCrossRunAndPacketReplay<br>cmd/kbcheck | cheap; deleting saves nothing measurable. |

### Cheap (< 200ms) -- 275 tests, 32s combined

**Bulk verdict: `keep`.** These are 62% of the suite by count and 6% by cost.
Deleting all of them would save under 35 seconds while destroying most of the
behavioural assertions in the repo. Listed for completeness.

| sec | covered | unique | test | pkg |
|---:|---:|---:|---|---|
| 0.19 | 49 | 1 | TestInventoryRejectsNonRepository | cmd/kbreconcile |
| 0.19 | 260 | 0 | TestModelsSelectIgnoreBypassesCorruptSavedPriority | cmd/kbrouter |
| 0.19 | 304 | 0 | TestExecutionTelemetryRejectsForgedProjectEvidenceUnderPrivateRun | cmd/kbcheck |
| 0.19 | 224 | 5 | TestManifestContractPassesValidDoneAndParkedGates | cmd/kbcheck |
| 0.19 | 185 | 5 | TestSliceLeasePathAndResourceConflictNormalization | cmd/kbcheck |
| 0.18 | 189 | 26 | TestProjectPriorityIsUserLocalCanonicalAndQuickAddIsConservative | cmd/kbrouter |
| 0.18 | 223 | 1 | TestC1ModelsApproveHostedRouteApprovalEnablesFallbackTransition | cmd/kbrouter |
| 0.18 | 110 | 0 | TestC1SessionEvidenceAcceptsLargeRolloutLogAfterEvidence | cmd/kbrouter |
| 0.18 | 267 | 2 | TestModelsApprovalModeDefaultsDisabledAndRequiredOptIn | cmd/kbrouter |
| 0.18 | 262 | 4 | TestExecutionTelemetryRejectsCallerMaxAgeExtensionByStrictSchema | cmd/kbcheck |
| 0.18 | 157 | 3 | TestPreSliceReviewContractRejectsDuplicateSectionOrOptIn | cmd/kbcheck |
| 0.18 | 119 | 3 | TestProofRejectsTamperedTrace | cmd/kbcheck |
| 0.18 | 208 | 19 | TestPlanRunLeaseOwnerGenerationLifecycleFailsClosed | cmd/kbcheck |
| 0.18 | 193 | 3 | TestContextPacketJSONReadErrorIsStructured | cmd/kbcheck |
| 0.18 | 245 | 0 | TestProofSelectionRunsOnlyInvalidatedChecksAndExplainsPaths | cmd/kbcheck |
| 0.18 | 326 | 3 | TestPlanRunLeaseRecoveryRechecksContentionAndReleaseWaitsForSlices | cmd/kbcheck |
| 0.17 | 460 | 3 | TestQualificationPlanContractRejectsEscapedOrStaleRecord | cmd/kbcheck |
| 0.17 | 117 | 1 | TestProofRejectsVacuousGreenTrace | cmd/kbcheck |
| 0.17 | 442 | 7 | TestCrossManifestSchedulerSerializesOneManifestOwnedWorktree | cmd/kbcheck |
| 0.17 | 323 | 1 | TestExecutionTelemetryRejectsProofArtifactMismatch | cmd/kbcheck |
| 0.17 | 327 | 3 | TestPlanRunLeaseAllowsRunQualifiedSiblingSliceIDsInsideOneWorkstreamEach | cmd/kbcheck |
| 0.17 | 167 | 1 | TestPreSliceReviewContractRejectsDuplicateReceiptFields | cmd/kbcheck |
| 0.17 | 367 | 6 | TestExecutionTelemetryMatchingProjectEvidenceNeedsHostAttestation | cmd/kbcheck |
| 0.17 | 179 | 9 | TestCatalogCalibrateIsAttendedOnlyNoCapabilityCredit | cmd/kbrouter |
| 0.17 | 262 | 0 | TestModelsSelectUsesValidatedRunCatalogAndRunOnlyOverride | cmd/kbrouter |
| 0.17 | 365 | 4 | TestQualificationPlanContractRejectsDuplicateJSONKeys | cmd/kbcheck |
| 0.16 | 141 | 2 | TestManifestContractRejectsQuotedOrDuplicateSchema | cmd/kbcheck |
| 0.16 | 330 | 8 | TestProofExecutionBudgetBlocksIdenticalFailedReplayUntilInputChanges | cmd/kbcheck |
| 0.16 | 452 | 22 | TestProofPlanCommandIsReadOnlyAndMachineReadable | cmd/kbcheck |
| 0.16 | 365 | 4 | TestExecutionTelemetryRejectsOmittedRunBindingWithEvidence | cmd/kbcheck |
| 0.16 | 0 | 0 | TestMustMutateContractFailsWhenTargetIsGone | cmd/kbcheck |
| 0.16 | 206 | 2 | TestManifestContractRejectsBadPassedGate | cmd/kbcheck |
| 0.16 | 30 | 6 | TestScopeLeaseCollisionAndRelease | cmd/kbcheck |
| 0.16 | 68 | 0 | TestRunProcessCheckFailsClosedOnTimeout | cmd/kbcheck |
| 0.16 | 4 | 2 | TestProofRejectsFractionalExpectedExit | cmd/kbcheck |
| 0.16 | 165 | 0 | TestModelsImportIsAtomicAcrossMultipleRoutes | cmd/kbrouter |
| 0.15 | 138 | 29 | TestFenceGatewayAmbiguousRetryFailsClosed | internal/reconcile |
| 0.15 | 0 | 0 | TestDocContractDetectsDeletedPolicyButToleratesRewrite | cmd/kbcheck |
| 0.15 | 155 | 12 | TestManifestContractValidatesOptionalImpactPacketContract | cmd/kbcheck |
| 0.15 | 181 | 0 | TestRelevantInputFingerprintRejectsChangedDirtyAndUntrackedContent | cmd/kbcheck |
| 0.15 | 304 | 0 | TestPreSliceReviewContractAcceptsRequirementsWideReceipt | cmd/kbcheck |
| 0.15 | 144 | 0 | TestManifestContractModelRouteDoesNotSubstituteForProofCheck | cmd/kbcheck |
| 0.15 | 335 | 2 | TestExecutionTelemetryCommandDowngradesMismatchedEnvelope | cmd/kbcheck |
| 0.15 | 234 | 17 | TestPlanMixedPortfolioDecisionPacket | internal/reconcile |
| 0.15 | 98 | 8 | TestSkillGuidanceRejectsOversizeMissingAuditAndDeprecatedSkill | cmd/kbcheck |
| 0.15 | 114 | 33 | TestDoctorRepairsMarkedStaleRequiredTarget | cmd/kbcheck |
| 0.15 | 114 | 0 | TestC1ReadRunChildSingleOpenRejectsReplacement | cmd/kbrouter |
| 0.15 | 83 | 3 | TestProofSelectionBlocksUnknownCheck | cmd/kbcheck |
| 0.15 | 305 | 1 | TestPreSliceReviewContractAcceptsInlineComments | cmd/kbcheck |
| 0.14 | 394 | 0 | TestRedactedRouteRejectsSubstitutedSourceID | cmd/kbrouter |
| 0.14 | 176 | 4 | TestModelsImportCanonicalizesRouteWithoutGrantingTrustOrPersistingSecrets | cmd/kbrouter |
| 0.14 | 0 | 0 | TestC1DispatchLockSerializesConcurrentOwners | cmd/kbrouter |
| 0.14 | 115 | 5 | TestModelsImportRejectsSecretValuesUnknownFieldsAndSymlinks | cmd/kbrouter |
| 0.14 | 166 | 1 | TestModelsImportRejectsAuthenticatedPrivateHTTPWithoutMutatingCatalog | cmd/kbrouter |
| 0.14 | 197 | 15 | TestModelsLocalRoutingTogglesWithoutDeletingRoutes | cmd/kbrouter |
| 0.14 | 0 | 0 | TestC1DispatchLockReusesPersistentUnlockedStateFile | cmd/kbrouter |
| 0.14 | 141 | 0 | TestManifestContractModelRouteAllowsRouteFreeObjectiveContract | cmd/kbcheck |
| 0.14 | 142 | 0 | TestManifestContractRejectsFalseProofCheck | cmd/kbcheck |
| 0.14 | 109 | 0 | TestBlockerLifecycleContractParsesInlineBlockerLists | cmd/kbcheck |
| 0.14 | 453 | 0 | TestQualificationPlanContractAcceptsSpecificGuidance | cmd/kbcheck |
| 0.14 | 324 | 7 | TestProofExecutionBudgetRunsOnceThenReuses | cmd/kbcheck |
| 0.14 | 440 | 14 | TestQualificationPlanContractAcceptsExplicitTierRaise | cmd/kbcheck |
| 0.14 | 208 | 15 | TestCrossManifestSchedulerObservedExpansionRequeuesBeforeStateMutation | cmd/kbcheck |
| 0.14 | 194 | 2 | TestGraphifyAnnotationRejectsExactStructuralProvider | cmd/kbcheck |
| 0.14 | 249 | 0 | TestBlockerLifecycleContractAcceptsScopedHumanAndExternalGates | cmd/kbcheck |
| 0.14 | 180 | 3 | TestBlockerLifecycleContractRejectsInvalidBooleanAndDuplicateGateIDs | cmd/kbcheck |
| 0.14 | 124 | 104 | TestKBNativeRootsRecognized | cmd/kbcheck |
| 0.14 | 116 | 97 | TestSkillLintPassesValidSkillAndFailsBadFrontmatter | cmd/kbcheck |
| 0.14 | 99 | 4 | TestSkillGuidanceRejectsNestedAndUnnavigableReferences | cmd/kbcheck |
| 0.14 | 204 | 0 | TestCrossManifestSchedulerSerializesNormalizedClaimsAndAdmitsDisjointRuns | cmd/kbcheck |
| 0.14 | 61 | 1 | TestModelRoutingReleaseRejectsSelfAuthoredLiveSupportWithoutExternalVerifier | cmd/kbcheck |
| 0.14 | 202 | 22 | TestSliceLeaseOwnerTokenRenewReleaseAndRecovery | cmd/kbcheck |
| 0.14 | 458 | 1 | TestQualificationPlanContractRejectsStaleReviewBinding | cmd/kbcheck |
| 0.14 | 156 | 0 | TestManifestContractRejectsForcedAMROrCrossOwnerFallback | cmd/kbcheck |
| 0.13 | 189 | 5 | TestManifestContractRequiresModelSelectionMetadataWhenEnabled | cmd/kbcheck |
| 0.13 | 54 | 0 | TestReviewReferenceGuardPassesForMatchingSharedPair | cmd/kbcheck |
| 0.13 | 139 | 1 | TestManifestContractValidatesOptInModelTiers | cmd/kbcheck |
| 0.13 | 243 | 11 | TestManifestContractValidatesProofReceiptContents | cmd/kbcheck |
| 0.13 | 8 | 0 | TestReleaseProfileDoesNotRepeatCoreChildCheck | cmd/kbcheck |
| 0.13 | 56 | 2 | TestReviewReferenceGuardRejectsUndocumentedFork | cmd/kbcheck |
| 0.13 | 130 | 0 | TestLegacyManifestWithoutSchemaDoesNotRequirePreSliceReview | cmd/kbcheck |
| 0.13 | 208 | 0 | TestProofReceiptFileValidationRejectsTampering | cmd/kbcheck |
| 0.13 | 187 | 7 | TestManifestContractValidatesPacketFileWhenEnabled | cmd/kbcheck |
| 0.13 | 187 | 1 | TestPreSliceReviewContractAcceptsNotRequiredReceipt | cmd/kbcheck |
| 0.13 | 189 | 3 | TestManifestContractAcceptsOrchestratorDirectedModelSelectionContract | cmd/kbcheck |
| 0.13 | 81 | 4 | TestRunProcessCheckFailsClosedOnOversizedOutput | cmd/kbcheck |
| 0.13 | 129 | 3 | TestModelRoutingReleaseProofRunnerIsFixedNoPaidAndFailurePropagates | cmd/kbcheck |
| 0.13 | 22 | 2 | TestSkillHashIgnoresRuntimeCachesButDetectsSourceChanges | cmd/kbcheck |
| 0.13 | 195 | 13 | TestManifestContractValidatesWorkspaceIsolationIntent | cmd/kbcheck |
| 0.13 | 213 | 0 | TestManifestContractTreatsLegacyDDRMetadataAsInertTelemetry | cmd/kbcheck |
| 0.13 | 221 | 3 | TestProofExecutionBlocksAutomaticGUIClassesBeforeSpawn | cmd/kbcheck |
| 0.13 | 81 | 0 | TestReviewReferenceGuardSweepPassesForClassifiedCommonReference | cmd/kbcheck |
| 0.13 | 241 | 3 | TestBlockerLifecycleContractRejectsStalePassAndDateOnlyCheck | cmd/kbcheck |
| 0.13 | 131 | 5 | TestModelRoutingReleaseAcceptsHonestNoPaidArtifactWithoutPromotion | cmd/kbcheck |
| 0.13 | 211 | 3 | TestModelTierEvalRejectsEscapedSymlinkAndOversizeInput | cmd/kbcheck |
| 0.13 | 216 | 0 | TestProofSelectionReusesPassingSuperset | cmd/kbcheck |
| 0.13 | 282 | 5 | TestDoctorAndDDRRejectStoredAuthenticatedPrivateHTTPBeforeNetwork | cmd/kbrouter |
| 0.12 | 103 | 10 | TestReadySetFiltersStatusesAndDetectsCycles | cmd/kbcheck |
| 0.12 | 81 | 63 | TestMarketplaceFirebreakFailsQuarantineActiveRoot | cmd/kbcheck |
| 0.12 | 123 | 2 | TestCoverageSubsumptionCollapsesCompositeAndDuplicates | cmd/kbcheck |
| 0.12 | 138 | 1 | TestManifestContractRequiresPacketForPendingSliceWhenEnabled | cmd/kbcheck |
| 0.12 | 388 | 10 | TestProofReceiptValidateCommand | cmd/kbcheck |
| 0.12 | 83 | 6 | TestDoctorRefusesUnknownRequiredDrift | cmd/kbcheck |
| 0.12 | 141 | 1 | TestManifestContractRequiresDoneCheckWhenObjectiveContractEnabled | cmd/kbcheck |
| 0.12 | 0 | 0 | TestLowCognitiveBurdenCommunicationContract | cmd/kbcheck |
| 0.12 | 0 | 0 | TestModelRoutingReleaseRejectsSymlinkFixture | cmd/kbcheck |
| 0.12 | 45 | 21 | TestDiscoverPackageChecks | cmd/kbcheck |
| 0.12 | 142 | 0 | TestManifestContractModelRouteAllowsLegacyHintOutsideTierRoutes | cmd/kbcheck |
| 0.12 | 138 | 2 | TestManifestContractRequiresDoneGate | cmd/kbcheck |
| 0.12 | 83 | 1 | TestReadySetSerialExclusionAndSingleSerial | cmd/kbcheck |
| 0.12 | 232 | 5 | TestCoreListPrintsNativeChecks | cmd/kbcheck |
| 0.12 | 304 | 14 | TestGateLedgerCommandValidatesAllowedNext | cmd/kbcheck |
| 0.12 | 140 | 0 | TestManifestContractRequiresProofCheckOrExceptionWhenObjectiveContractEnabled | cmd/kbcheck |
| 0.12 | 51 | 1 | TestReviewReferenceGuardRejectsSharedDrift | cmd/kbcheck |
| 0.12 | 56 | 0 | TestReviewReferenceGuardAllowsSingleOwnerWithoutSharedPairs | cmd/kbcheck |
| 0.12 | 250 | 21 | TestBlockerLifecycleContractRequiresQuarantineBoundary | cmd/kbcheck |
| 0.12 | 277 | 2 | TestCrossManifestSchedulerReadySetRequiresLiveManifestAuthority | cmd/kbcheck |
| 0.12 | 140 | 0 | TestModelsAddProfileRejectsProjectScope | cmd/kbrouter |
| 0.12 | 116 | 0 | TestBlockerLifecycleContractPreservesColonAndCommaListItems | cmd/kbcheck |
| 0.12 | 244 | 3 | TestBlockerLifecycleContractRejectsMisownedStaleAndOverpropagatedGates | cmd/kbcheck |
| 0.12 | 22 | 1 | TestProofSenseTimeoutFailsCleanly | cmd/kbcheck |
| 0.12 | 270 | 2 | TestProofExecutionWritesTimeoutReceiptAndNoGlobalPass | cmd/kbcheck |
| 0.12 | 175 | 0 | TestHostedExtraRouteRequiresExplicitProjectApprovalWhenEnabled | cmd/kbrouter |
| 0.12 | 187 | 10 | TestProofReceiptRejectsMissingInputsRegistryDriftAndExpiry | cmd/kbcheck |
| 0.12 | 65 | 48 | TestSurfaceReportComparisonShowsAddedAndRemovedRoutes | cmd/kbcheck |
| 0.12 | 0 | 0 | TestPlanWideSpecialistReviewContract | cmd/kbcheck |
| 0.12 | 193 | 1 | TestBlockerLifecycleContractRejectsPauseAsGateStatus | cmd/kbcheck |
| 0.12 | 121 | 1 | TestC1ProfileRejectsUnsupportedCredentialAmbiguity | cmd/kbrouter |
| 0.12 | 63 | 0 | TestDiscoverSkillRepoChecksIncludesNativeValidators | cmd/kbcheck |
| 0.12 | 39 | 0 | TestProviderHygieneDecodesEscapedPhoenix | cmd/kbcheck |
| 0.12 | 96 | 2 | TestSkillGuidanceAcceptsCompleteCompactSurface | cmd/kbcheck |
| 0.12 | 41 | 22 | TestProviderHygieneIgnoresDisabledTomlPhoenix | cmd/kbcheck |
| 0.12 | 140 | 0 | TestManifestSchemaTwoRequiresPreSliceReviewContract | cmd/kbcheck |
| 0.12 | 66 | 0 | TestSkillRepoContractForNativeCheckNames | cmd/kbcheck |
| 0.11 | 304 | 22 | TestReadySetAndScopeLeaseCommands | cmd/kbcheck |
| 0.11 | 173 | 0 | TestProofCoverageReceiptReusesPassingSuperset | cmd/kbcheck |
| 0.11 | 0 | 0 | TestProductionDDRContractExcludesAMRAndSeparatesHostSurfaces | cmd/kbcheck |
| 0.11 | 139 | 4 | TestManifestContractAllowsExplicitNoCheckException | cmd/kbcheck |
| 0.11 | 15 | 0 | TestLocalReleaseRequiresNativeModelRoutingGateWhenPilotEvidenceExists | cmd/kbcheck |
| 0.11 | 40 | 0 | TestProviderHygieneRejectsPhoenix | cmd/kbcheck |
| 0.11 | 56 | 3 | TestRunStateAllowsProgressBetweenRoutes | cmd/kbcheck |
| 0.11 | 191 | 2 | TestGraphRouteCommandValidatesImpactPacket | cmd/kbcheck |
| 0.11 | 82 | 1 | TestReviewReferenceGuardSweepRejectsUnclassifiedCommonReference | cmd/kbcheck |
| 0.11 | 14 | 1 | TestReleaseRequiresNativeSyncForSkillRepo | cmd/kbcheck |
| 0.11 | 0 | 0 | TestAutomaticDeliveryChainContract | cmd/kbcheck |
| 0.11 | 0 | 0 | TestPreSliceReviewContractRejectsSourceSymlinkOutsideRepository | cmd/kbcheck |
| 0.11 | 3 | 1 | TestGHCPFollowOnCannotRewriteInitialCohort | cmd/kbcheck |
| 0.11 | 349 | 27 | TestDoctorProbeDistinguishesReachabilityAndModelPresence | cmd/kbrouter |
| 0.11 | 66 | 4 | TestCredentialConsumingCommandsRejectCustomUserRootOutsideTestSeam | cmd/kbrouter |
| 0.11 | 176 | 2 | TestCatalogAddUsesStrictUserSchemaAndRejectsUnsafeProjectPolicy | cmd/kbrouter |
| 0.11 | 67 | 0 | TestCheckedInRouteExampleIsARejectedPlaceholder | cmd/kbrouter |
| 0.1 | 11 | 8 | TestGHCPFollowOnRefusesPromotionWithoutIndependentVerifier | cmd/kbcheck |
| 0.1 | 168 | 7 | TestPlanMissingMandatoryEvidenceFailsClosed | internal/reconcile |
| 0.1 | 111 | 0 | TestSemanticClaimDisjointConcurrencyAndSameResourceSerialization | internal/reconcile |
| 0.1 | 54 | 4 | TestCoreRunsDiscoveredCheck | cmd/kbcheck |
| 0.1 | 65 | 4 | TestCoreVerbosePreservesPassingOutput | cmd/kbcheck |
| 0.1 | 10 | 0 | TestGenericReleaseOmitsModelRoutingGateWhenFeatureIsAbsent | cmd/kbcheck |
| 0.1 | 9 | 2 | TestGraphRoutingEvalFailsStaleAuthoritativeIndex | cmd/kbcheck |
| 0.1 | 25 | 0 | TestProofDigestChangesWhenCheckerScriptChanges | cmd/kbcheck |
| 0.1 | 26 | 1 | TestProofDigestChangesWhenTimeoutChanges | cmd/kbcheck |
| 0.1 | 159 | 0 | TestGraphRouteParseRequiresPacket | cmd/kbcheck |
| 0.1 | 27 | 2 | TestLearningAdoptionRejectsRightToWrongRegression | cmd/kbcheck |
| 0.1 | 0 | 0 | TestRepairPolicyRejectsUnconditionalReplayLanguage | cmd/kbcheck |
| 0.1 | 65 | 6 | TestCoreFailurePropagates | cmd/kbcheck |
| 0.1 | 236 | 0 | TestProofSelectionRejectsPreIntegrationNamespaceReceipt | cmd/kbcheck |
| 0.1 | 25 | 2 | TestDiscoverNestedDotnetProject | cmd/kbcheck |
| 0.1 | 1 | 1 | TestReconcileTerminalCleanupPredicateContract | cmd/kbcheck |
| 0.1 | 97 | 2 | TestReadySetParallelSlices | cmd/kbcheck |
| 0.1 | 28 | 5 | TestLearningAdoptionRejectsMemorizedHoldoutString | cmd/kbcheck |
| 0.1 | 58 | 2 | TestRunStateDetectsLowConfidenceNoProgress | cmd/kbcheck |
| 0.1 | 16 | 1 | TestLiveReleaseUsesNativeLiveCorpus | cmd/kbcheck |
| 0.1 | 0 | 0 | TestDeliveryAuthorityLedger | cmd/kbcheck |
| 0.1 | 78 | 39 | TestSkillSyncReportFindsRequiredDrift | cmd/kbcheck |
| 0.1 | 0 | 0 | TestPlanRunWorktreeAndBranchShareFunnyTaskName | cmd/kbcheck |
| 0.1 | 0 | 0 | TestCargoBuildStorageContract | cmd/kbcheck |
| 0.1 | 54 | 1 | TestReleaseReportsCheckStartBeforeRunnerReturns | cmd/kbcheck |
| 0.1 | 0 | 0 | TestDeliveryOwnerSkillContracts | cmd/kbcheck |
| 0.1 | 11 | 0 | TestLiveReleaseIncludesModelRoutingGateExactlyOnce | cmd/kbcheck |
| 0.1 | 5 | 2 | TestGHCPFollowOnNoPaidEvidenceStaysNotPromoted | cmd/kbcheck |
| 0.1 | 0 | 0 | TestDeliveryAuthorityLedgerRejectsContradictoryGrants | cmd/kbcheck |
| 0.1 | 0 | 0 | TestProofSpineHelperProcess | cmd/kbcheck |
| 0.1 | 217 | 3 | TestModelRoutingFeatureMarkerRequiresGateWhenEvidenceIsMissing | cmd/kbcheck |
| 0.1 | 110 | 0 | TestC1SessionEvidenceDoesNotEnumerateHistoricalSessions | cmd/kbrouter |
| 0.1 | 280 | 3 | TestDoctorReportsSeparateDimensionsWithoutMutation | cmd/kbrouter |
| 0.1 | 0 | 0 | TestPrepareRunRootRejectsExistingRunRootSymlink | cmd/kbrouter |
| 0.1 | 73 | 7 | TestC1SessionEvidenceRejectsStaleAmbiguousAndUnsafeSessions | cmd/kbrouter |
| 0.1 | 110 | 0 | TestC1SessionEvidenceFindsDatedCodexRollout | cmd/kbrouter |
| 0.1 | 206 | 0 | TestShowRedactsPrivateConnectionDetails | cmd/kbrouter |
| 0.1 | 4 | 0 | TestConfigureProcessTreeRequestsBreakawayBeforeJobAssignment | cmd/kbrouter |
| 0.1 | 93 | 1 | TestPreparedRunRootDetectsAncestorReplacement | cmd/kbrouter |
| 0.1 | 16 | 12 | TestUserCatalogEnabledFalsePreservesConfigAndDisablesRoutes | cmd/kbrouter |
| 0.1 | 12 | 1 | TestCheckedInDDRRequestExampleIsARejectedPlaceholder | cmd/kbrouter |
| 0.1 | 101 | 4 | TestC1SessionEvidenceRequiresOptionalProfileAndRouteBindings | cmd/kbrouter |
| 0.1 | 36 | 1 | TestProviderHygieneIgnoresDisabledPhoenix | cmd/kbcheck |
| 0.1 | 17 | 0 | TestIsolatedGoTestArgsBoundThePackageTestBinary | cmd/kbcheck |
| 0.1 | 0 | 0 | TestDispatchAttestationRootRejectsSymlinkedProjectContainment | cmd/kbcheck |
| 0.1 | 47 | 3 | TestReleaseJSONReportsRequiredFailure | cmd/kbcheck |
| 0.1 | 0 | 0 | TestDocAnchorIgnoresCaseWrappingAndEmphasis | cmd/kbcheck |
| 0.1 | 0 | 0 | TestDocAnchorPreservesCommandAndKeyPrecision | cmd/kbcheck |
| 0.1 | 17 | 0 | TestReleaseSkipsSyncForGenericRepoWithoutSkillConfig | cmd/kbcheck |
| 0.1 | 25 | 1 | TestLearningAdoptionRejectsLowSampleCount | cmd/kbcheck |
| 0.1 | 0 | 0 | TestProofCadencePolicyRequiresReuseAndBatchBoundaries | cmd/kbcheck |
| 0.1 | 172 | 0 | TestParseCoreVerbose | cmd/kbcheck |
| 0.1 | 33 | 5 | TestPlanWorktreeSelftestRealRepoFirebreakCannotBeForced | cmd/kbcheck |
| 0.09 | 21 | 2 | TestGraphRoutingEvalFailsMissedImpact | cmd/kbcheck |
| 0.09 | 15 | 0 | TestGoTestParallelismTracksMemoryHeadroomNotCPUCount | cmd/kbcheck |
| 0.09 | 7 | 0 | TestAvailableProcessMemoryIsPlausibleWhenReported | cmd/kbcheck |
| 0.09 | 14 | 1 | TestGraphRoutingEvalFailsMultiWinnerRace | cmd/kbcheck |
| 0.09 | 7 | 0 | TestGoTestPackagePartitionIsolatesOnlyChildProcessOwners | cmd/kbcheck |
| 0.09 | 178 | 8 | TestParseTerminalCleanupArgs | cmd/kbcheck |
| 0.09 | 130 | 5 | TestRunStateCommandRequiresHistory | cmd/kbcheck |
| 0.09 | 22 | 1 | TestLearningAdoptionAcceptsMeasuredGain | cmd/kbcheck |
| 0.09 | 10 | 0 | TestMatchingTerminalCleanupClaimFindsExactDuplicateIdentity | cmd/kbcheck |
| 0.09 | 46 | 20 | TestSemanticClaimCanonicalKeysAndAliases | internal/reconcile |
| 0.09 | 0 | 0 | TestDocConceptRequiresEveryTerm | cmd/kbcheck |
| 0.09 | 0 | 0 | TestDocConceptTreatsProhibitionSpellingsAsEquivalent | cmd/kbcheck |
| 0.09 | 68 | 1 | TestPolicyProjectPreferencesContainNoConnectionDetails | cmd/kbrouter |
| 0.09 | 197 | 0 | TestCodexFixtureDiscoveryNeverPromotesAdapterPrior | cmd/kbrouter |
| 0.09 | 0 | 0 | TestDurableSchemasRejectMissingUnsupportedDuplicateOversizeAndSymlink | cmd/kbrouter |
| 0.09 | 57 | 0 | TestPolicyJSONHasNoUnknownFields | cmd/kbrouter |
| 0.09 | 116 | 1 | TestParseRejectsJSONForCore | cmd/kbcheck |
| 0.09 | 0 | 0 | TestSemanticClaimRepoContract | cmd/kbcheck |
| 0.09 | 38 | 0 | TestProviderHygieneAllowsUnrelatedProvider | cmd/kbcheck |
| 0.09 | 65 | 10 | TestRunStateDetectsRouteOscillation | cmd/kbcheck |
| 0.09 | 235 | 10 | TestContextPacketCommandUsesFixtures | cmd/kbcheck |
| 0.09 | 5 | 4 | TestDeliveryChainLifecycleStatesRemainSeparate | cmd/kbcheck |
| 0.09 | 8 | 1 | TestNormalizeExecutionTelemetryRequiresPacketID | cmd/kbcheck |
| 0.09 | 172 | 0 | TestParseReleaseArgs | cmd/kbcheck |
| 0.09 | 7 | 7 | TestTerminalCleanupSweepTextReportsPartialMutation | cmd/kbcheck |
| 0.09 | 36 | 0 | TestContextPacketRejectsRepoRootAndBlankMemoryPath | cmd/kbcheck |
| 0.09 | 45 | 0 | TestNormalizeExecutionTelemetryWithCreditedReceiptOverridesForgedAssertions | cmd/kbcheck |
| 0.09 | 0 | 0 | TestDocConceptSurvivesRewordingButNotDeletion | cmd/kbcheck |
| 0.09 | 304 | 1 | TestModelRoutingReleaseFailureBlocksLocalRelease | cmd/kbcheck |
| 0.09 | 37 | 0 | TestContextPacketValidation | cmd/kbcheck |
| 0.09 | 115 | 1 | TestProofGovernorCLIRemovesApprovalSurface | cmd/kbcheck |
| 0.08 | 172 | 0 | TestParseCoreList | cmd/kbcheck |
| 0.08 | 178 | 4 | TestParseContextPacketAndProviderHygiene | cmd/kbcheck |
| 0.08 | 16 | 0 | TestNormalizeExecutionTelemetryObservationOnlyForProofUnknown | cmd/kbcheck |
| 0.08 | 14 | 0 | TestReleaseChecksUseNativeCoreNotPSGate | cmd/kbcheck |
| 0.08 | 37 | 0 | TestContextPacketRequiresAuthorityBounds | cmd/kbcheck |
| 0.08 | 9 | 3 | TestResolveRepoPathExpandsHome | cmd/kbcheck |
| 0.08 | 33 | 0 | TestNormalizeExecutionTelemetryDowngradesForgedFieldsWithoutReceipt | cmd/kbcheck |
| 0.08 | 7 | 1 | TestNormalizeExecutionTelemetryRejectsInvalidCounters | cmd/kbcheck |
| 0.08 | 0 | 0 | TestSnapshotSelectionRequiresScopeAndReusesMilestone | cmd/kbcheck |
| 0.08 | 33 | 5 | TestNormalizeExecutionTelemetryKeepsExactGHCPAccountingProofNeutral | cmd/kbcheck |
| 0.08 | 37 | 0 | TestContextPacketRejectsBroadGlobWhenSearchBounded | cmd/kbcheck |
| 0.08 | 32 | 1 | TestNormalizeExecutionTelemetryDoesNotTreatLegacyModelAsActual | cmd/kbcheck |
| 0.08 | 41 | 0 | TestCodexProviderMetadataClassifiesWithoutModelNameInference | cmd/kbrouter |
| 0.08 | 9 | 9 | TestModelRoutingPromotionRequiresMeasuredNonzeroPrimaryMetric | cmd/kbcheck |
| 0.08 | 0 | 0 | TestCodexUnixDiscoveryCanonicalizesLauncherSymlinkButStaysUnproven | cmd/kbrouter |
| 0.08 | 71 | 11 | TestDDRCommandsExposeFocusedHelpAndStructuredUsageErrors | cmd/kbrouter |
| 0.08 | 57 | 0 | TestModelsShowAcceptsDisabledEmptyCatalogWithNullRoutes | cmd/kbrouter |
| 0.08 | 102 | 2 | TestAttendedApprovalFailureDoesNotMutate | cmd/kbrouter |
| 0.08 | 47 | 3 | TestDiscoverRejectsRepositoryOrAncestorRunRootWithoutChangingPermissions | cmd/kbrouter |
| 0.08 | 6 | 1 | TestDispatchPacketRequiresExplicitDelegatedOwnership | cmd/kbrouter |
| 0.08 | 195 | 0 | TestDiscoveredCurrentRouteCanRetainExplicitCurrentOwnership | cmd/kbrouter |
| 0.08 | 34 | 1 | TestModelsDiscoverDoesNotExposeFixtureFlagsInProduction | cmd/kbrouter |
| 0.08 | 32 | 0 | TestC1RejectsPreExistingDispatchArtifacts | cmd/kbrouter |
| 0.07 | 0 | 0 | TestCanonicalizeProspectivePathResolvesExistingAncestorAlias | cmd/kbrouter |
| 0.07 | 6 | 0 | TestOperatingSystemHomeIgnoresCallerHomeOverride | cmd/kbrouter |
| 0.07 | 22 | 1 | TestPrepareRunRootFailsClosedWhenProjectCannotBeCanonicalized | cmd/kbrouter |
| 0.07 | 34 | 14 | TestC1FallbackTrustTransitionRequiresApproval | cmd/kbrouter |
| 0.07 | 2 | 0 | TestDispatchSchemaNameIsSliceScoped | cmd/kbrouter |
| 0.07 | 31 | 0 | TestC1RejectsDerivedArtifactNamespaceCollision | cmd/kbrouter |
| 0.07 | 0 | 0 | TestPrepareRunRootRejectsSymlinkedProjectAncestors | cmd/kbrouter |
| 0.07 | 49 | 2 | TestModelsSelectRequiresExplicitOwnershipDecision | cmd/kbrouter |
| 0.07 | 6 | 1 | TestDispatchAttestationRequiresExactHostBuiltReceiptBytes | cmd/kbrouter |
| 0.07 | 0 | 0 | TestTrustedCodexExecutableRejectsSymlinkedProjectContainment | cmd/kbrouter |
| 0.07 | 0 | 0 | TestPrepareRunRootRejectsMixedAliasProjectAndRunPath | cmd/kbrouter |
| 0.06 | 11 | 4 | TestSafeWindowsRunIDRejectsAmbiguousNames | cmd/kbrouter |
| 0.06 | 64 | 8 | TestPolicyManifestMatchesDefault | internal/reconcile |
| 0.06 | 20 | 17 | TestSemanticClaimCapabilityAndConformanceJSON | cmd/kbreconcile |
| 0.06 | 85 | 59 | TestSemanticClaimCASTakeoverAndRollbackFence | internal/reconcile |
| 0.05 | 190 | 0 | TestPlanStableJSONAndDedupVocabulary | internal/reconcile |
| 0.05 | 28 | 2 | TestPlanRequiresOutput | cmd/kbreconcile |
