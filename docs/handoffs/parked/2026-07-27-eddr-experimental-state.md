# EDDR Experimental State

Status: parked-exception
Created: 2026-07-27
Repository: `Irtechie/working-skill-repo`

## Durable Planning Snapshot

Remote ref: `refs/heads/parked/eddr-planning-2026-07-27`
Commit: `ff733b6d7065c272c69e4457a8d212ff3e5f808e`

Exact paths:

- `.github/skills/kb-ddr-plan/SKILL.md`
- `.github/skills/kb-plan/SKILL.md`
- `docs/context/architecture/skills.md`

The ref is preservation only. It has no PR and is not approved for merge or
global propagation. A sensitive-pattern scan found no credential, private
endpoint, or host-path match in these three files.

## Active Benchmark Cohort

Checked at: `2026-07-28T05:10:46Z`
Worktree: `deaderestpool-ubiquitous-system`
Branch: `deaderestpool-model-routing-benchmark`
HEAD: `831e794fe2273131c75d34d797393d6aa890e92d`
Owner: `Model routing benchmark` session
State: owner-controlled DDR cohort followed by same-fixture AMR

Path-only status: 15 modified files and 10 untracked status entries. Untracked
directory entries may contain many files, so this is a status-entry count, not
an expanded file count.

The prior 415-path local checkpoint inventory is historical. Active state has
advanced beyond it, so it must not be used as the current convergence sensor.

The exact modified files are:

- `.github/skills/kb-plan/SKILL.md`
- `README.md`
- `cmd/amrbench/approval.go`
- `cmd/amrbench/approval_test.go`
- `cmd/amrbench/main.go`
- `cmd/amrbench/main_test.go`
- `cmd/amrbench/runner_test.go`
- `cmd/kbcheck/ddr_contract_test.go`
- `docs/context/architecture/skills.md`
- `docs/context/eval-map.md`
- `docs/context/operations/testing.md`
- `docs/context/research/README.md`
- `evals/amr-model-benchmark/README.md`
- `evals/amr-model-benchmark/config.json`
- `evals/amr-model-benchmark/fixtures/retry-after-parser/SPEC.md`

The exact untracked status entries are:

- `.github/skills/kb-ddr-plan/`
- `docs/brainstorms/2026-07-27-portable-local-subagent-routing-requirements.md`
- `docs/context/research/2026-07-27-ddr-cost-routing-benchmark.md`
- `docs/results/2026-07-27-ddr-cost-routing-evidence.pptx`
- `docs/results/2026-07-27-ddr-guide-conformance-pair.json`
- `docs/results/2026-07-27-ddr-hosted-model-matrix.json`
- `docs/results/2026-07-27-ddr-local-route-readiness.json`
- `docs/results/2026-07-27-ddr-real-execution.json`
- `evals/cross-model-benchmarks/ddr-portable-plans.json`
- `evals/ddr-model-benchmark/`

The owner issued a hard stop against convergence, cleanup, delivery, or
reservation release while the cohort runs. This handoff records the parent
goal's authorized exception; it does not transfer mutation authority or park
the owner's live execution.

PR #2 may merge while this exception remains active. Merging the handoff does
not merge, push, approve, release, or otherwise mutate the checkpoint or cohort.

## Qwen Infrastructure Disposition

Qwen is `INFRASTRUCTURE NOT-RUN`. Its single prescribed acquire ended
`descriptive_only` after an `allocation-launcher-v2` `AttributeError`. It was
not retried after reboot. Owner `ddr-medium-qwen-20260727-c5047831` has no
remaining reservation, cleanup target, or endpoint. This is permitted DDR
outlier evidence, not unfinished non-DDR work or a model-result claim.

## Luna Proof Disposition

The resumed Luna executor found all four durable todos/plan items already done.
It inspected the benchmark fixture and passed both:

```powershell
go test ./...
go vet ./...
```

in
`evals/ddr-model-benchmark/evidence/layered-access-policy-medium/luna-workspace`.
No further Luna changes or DDR proof gate remain for PR #2 convergence.

Resume condition for convergence: the owner supplies the final cohort safe-stop
package after five DDR asks and planned same-fixture AMR, or the user explicitly
terminates or parks the campaign. Then run:

`/kb-goal Reconcile all remaining valuable Irtechie/working-skill-repo work into clean, reviewable check-ins tonight.`

Do not commit generated evidence, host-local catalogs, credentials, private
endpoints, or model runtime state.
