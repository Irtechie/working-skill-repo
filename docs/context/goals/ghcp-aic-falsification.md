# GHCP AIC Falsification

Status: blocked
Created: 2026-07-12
Last updated: 2026-07-12

## Objective

Complete the GHCP AIC/context falsification harness through
deterministic no-paid readiness, then present the attended-run contract and
pause before any AIC spend.

## Done Criteria

- All four manifest slices are done or explicitly quarantined with proof.
- No-paid conformance passes with exact accounting, qualified fixtures,
  isolation, context contracts, and paired grading.
- Review, follow-up resolution, memory refresh, and cleanup complete.
- The user receives the cohort matrix, task fixtures, exact AIC/token boundary,
  maximum credit budget, local-model/DS4 availability, and exact commands.
- No paid/model call runs without a later explicit attended approval.

## Terminal Proof

- `complete-to-ship` passes for
  `docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md`.
- The no-paid done check exits 0.
- The attended-run preview is recorded before the AIC approval pause.

## Done Check

- Type: command_exit
- Check: `go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json`
- Expected result: exit code 0
- Why sufficient: proves deterministic readiness without loading a live runner,
  provider profile, secret, or spending AIC.

## Current State

- Current artifact: `docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md`
- Next allowed action: implement a trusted human-approval verifier, restore
  provider response-contract compatibility, regenerate the paired experiment,
  then request approval.
- Last proof: exploratory Qwen Small to Sol Large cost run completed. Mean Sol
  billed cost was 13.4155 AIC; Qwen plus Sol fallback was 13.61725 AIC with
  106.4% token overhead. No sample passed correctness, so no winner exists.

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|
| Exact GHCP accounting | `kb-work` | slice-001 | done | focused/package tests and snapshot |
| Isolation and fixture authority | `kb-work` | slice-002 | done | protected oracle; no-paid Podman qualification |
| Context-diet contract | `kb-work` | slice-003 | done | context proof and frozen contracts |
| Paired falsification grader | `kb-work` | slice-004 | done | conformance/paired proof |
| Post-work completion | `kb-complete` | manifest | done | `complete-to-ship` passed |
| Attended approval gate | human | preview | blocked | trusted verifier and fleet readiness required |
| Qwen partial canary | attended run | `docs/results/2026-07-13-qwen-canary-budget-route-failure.md` | blocked | GHCP route and credit-cap enforcement |
| Qwen/Sol exploratory cost | attended run | `docs/results/2026-07-13-qwen-sol-exploratory-cost.md` | done | cost measured; correctness zero |

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|
| Trusted human approval verifier absent | safety | harness owner | implement independent approval verification; non-dry run remains disabled |
| Qwen local route unavailable | fleet/profile | fleet owner | uncordon/reserve Plato, create profile, and bind a fresh tier/availability probe |
| DS4 route unavailable | fleet | fleet owner | restore a second Ready GB10 node and free/reserve both Spark GPUs |
| GHCP route selection violated | provider/runtime | GHCP owner | requested GPT-5.4 must execute and OTel must report GPT-5.4 |
| Provider response contract fails | adapter/prompt | harness owner | Sol and Qwen must return a non-empty exact JSON file envelope and pass proof |
| Valid paired sample absent | evaluation | harness owner | obtain one correct direct and one correct AMR pair before collecting 20+ samples |

## Notes

- Temporary Podman installation and exact cleanup paths are recorded in
  `docs/handoffs/active/2026-07-11-ghcp-aic-no-paid-readiness.md`.
- Qwen/DS4 fleet reservations are deferred until immediately before an approved
  attended run.
