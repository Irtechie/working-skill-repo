# Reduce WSR Weight

Status: active
Created: 2026-08-02
Last updated: 2026-08-02

## Objective

Make the working-skill-repo bundle materially lighter to run and maintain without losing functionality.

## Done Criteria

- [user] Every test is deeply examined and its reduction verdict recorded in a shared file usable for the next plan.
- [user] Identified reductions are applied where they do not lose functionality.
- [user] The contributor loop (`kbcheck core`) is materially faster than the 467 s baseline.
- [derived] Serves criterion 2: no reduction removes a behavior that no other test proves; each deletion names the test that retains the coverage.
- [derived] Serves criterion 3: `go run ./cmd/kbcheck core` still exits 0 after every reduction batch.

## Terminal Proof

- `go run ./cmd/kbcheck core` exits 0 after reductions.
- Per-test audit file exists with a verdict for every test.
- Recorded before/after wall-clock for `core` and `go test ./...`.

## Done Check

- Type: command_exit
- Check: `go run ./cmd/kbcheck core`
- Expected result: exit 0, with recorded wall-clock materially below 467 s
- Why sufficient: proves reductions kept the gate functional while lowering its cost

## Baseline (2026-08-02)

| Measure | Value |
|---|---|
| `kbcheck core` | 467.7 s |
| `go test ./...` (parallel 7) | 164.4 s |
| `cmd/kbcheck` package alone | 160.4 s |
| All 30 native checks combined | 37.8 s |
| Go source | 59,193 lines across 193 files |
| Test source | 20,771 lines across 88 files |
| Skills | 44 |

Known structural costs, measured not assumed:

- `runGoTestsWithProcessIsolation` runs three sequential `go test` invocations (~284 s) instead of one (~164 s).
- `kbrouter-catalog-tests` re-runs kbrouter tests that `go test ./...` already ran.
- `runCore` (main.go:800) is a sequential loop that records no per-check duration, so cost has been invisible.
- 26 of 30 native checks finish under 0.5 s; the gate's reputation for slowness is misattributed.

## Current State

- Current artifact: `docs/context/research/2026-08-02-test-weight-audit.md`
- Next allowed action: decide on the `TestTerminalCleanup*` fixture refactor (largest remaining lever)
- Last proof: `go test ./...` green at 139.5 s; gate re-run in progress

## Audit Result (2026-08-02)

447 tests measured across the four packages that are 97% of suite runtime. Zero
failed when run alone, so there are no hidden inter-test dependencies.

The headline finding inverts the obvious plan: **deletion is the wrong lever.**

| If we deleted... | Tests lost | Time saved |
|---|---|---|
| all 161 zero-unique-coverage tests | 161 | 133.5 s |
| ...the 137 of them costing under 1 s | 137 | 23.7 s |
| ...the 24 costing over 1 s | 24 | 109.8 s |

279 of 447 tests cost under 200 ms and together account for 6% of runtime.
Deleting them buys ~33 s and destroys most of the behavioural assertions in the
repo. Cost is concentrated: the top 25 tests are 57% of total runtime.

The expensive tests are expensive because of **git fixture setup and deliberate
sleeps**, not because they assert nothing. The lever is setup cost, not test
count.

Two measured caveats that stopped bad deletions:

- `TestWorkQueueUpdateMigratesOwnedWorktreeIdentity` reports 0 covered
  statements because it drives a PowerShell script; Go coverage is blind to it.
- `TestRunProcessCheckTimeoutKillsGrandchild` reports 0 unique statements
  because it proves an OS process-tree kill, not a Go branch.

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|
| Memory-derived test parallelism | kb-fix | `cmd/kbcheck/checks.go`, `memory_*.go` | done | `go test ./...` green at 164.4 s |
| Doc-contract de-brittling | kb-fix | `cmd/kbcheck/doc_contract_test.go` | done | 8 contract files converted; matcher self-tests pass |
| Deep per-test audit | kb-goal | `docs/context/research/2026-08-02-test-weight-audit.md` | done | 447 tests, per-test verdicts, 0 failed in isolation |
| Fix + retune grandchild timing contract | kb-fix | `cmd/kbcheck/main_test.go` | done | 18.2 s -> 11.6 s; mutation test fails as required |
| Drop duplicate `kbrouter-catalog-tests` | kb-fix | `cmd/kbcheck/checks.go` | done | 4.9 s removed; 19 tests already run by `go test ./...` |
| Share `TestTerminalCleanup*` git fixture | tbd | `cmd/kbcheck/terminal_cleanup_test.go` | pending | largest remaining lever: 28 tests, 183 s contended |
| Concurrency + `DurationMS` in `runCore` | tbd | `cmd/kbcheck/main.go` | pending | `kbcheck core` exit 0, faster |

## Progress

| Measure | Baseline | Current |
|---|---|---|
| `go test ./...` | 164.4 s | 139.5 s |
| `cmd/kbcheck` package | 160.4 s | 137.7 s |
| `kbcheck core` | 467.7 s | 454.9 s (exit 0, 38 checks) |

The gate moved only 13 s because its dominant cost is structural, not per-test:
`go test ./...` is 139.5 s standalone but ~207 s inside the gate, because
`runGoTestsWithProcessIsolation` runs it as three sequential invocations. Per-test
reductions cannot fix that; the remaining levers are the fixture refactor and
gate concurrency.

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|

## Notes

Constraint from the user: reduce weight without losing functionality. Deletions
must be justified by evidence that coverage is retained elsewhere, not by a
judgment that a test "looks redundant".

Prefer deterministic evidence over model opinion. Unique-coverage measurement
per test is the primary signal; semantic review only adjudicates what the
measurement flags.
