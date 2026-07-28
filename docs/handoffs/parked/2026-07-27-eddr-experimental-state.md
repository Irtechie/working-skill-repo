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

Checked at: `2026-07-28T00:10:16-04:00`
Worktree: `deaderestpool-ubiquitous-system`
Branch: `deaderestpool-model-routing-benchmark`
HEAD: `831e794fe2273131c75d34d797393d6aa890e92d`
Owner: `Model routing benchmark` session
State: owner-controlled DDR cohort followed by same-fixture AMR

Path-only status: 415 dirty paths (15 modified, 400 untracked).
Checkpoint: `81f4351e72eb85e2d886a757bf2f77321fee7be2`.

This is an owner-local Copilot checkpoint, not a delivered or remote
preservation ref. It contains generated evidence and host-local runtime data,
so it must never be pushed, cherry-picked, or merged. It is cited only to make
the path inventory reproducible in the preserved worktree.

The checkpoint is an exact path snapshot relative to the branch HEAD:

```powershell
git diff --name-only 831e794fe2273131c75d34d797393d6aa890e92d 81f4351e72eb85e2d886a757bf2f77321fee7be2
```

It contains:

| Path group | Count |
|---|---:|
| `evals/ddr-model-benchmark/**` | 391 |
| `cmd/amrbench/**` | 5 |
| `docs/context/**` | 5 |
| `docs/results/**` | 5 |
| `evals/amr-model-benchmark/**` | 3 |
| `.github/skills/**` | 2 |
| `cmd/kbcheck/**` | 1 |
| `docs/brainstorms/**` | 1 |
| `evals/cross-model-benchmarks/**` | 1 |
| `README.md` | 1 |

The 15 modified paths are:

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

Resume condition for future benchmark promotion work: the owner supplies a new
explicit objective and claim boundary. Then run:

`/kb-goal Reconcile all remaining valuable Irtechie/working-skill-repo work into clean, reviewable check-ins tonight.`

Do not commit generated evidence, host-local catalogs, credentials, private
endpoints, or model runtime state.
