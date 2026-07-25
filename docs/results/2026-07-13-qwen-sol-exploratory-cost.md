# Qwen Small to Sol Large Exploratory Cost Run

Recorded: 2026-07-13

## Scope

Exploratory, non-promotional comparison on `retry-after-parser`:

- hosted direct: GPT-5.6 Sol;
- local attempt: Qwen Small;
- full fallback proxy: local Qwen attempt plus a fresh Sol direct-equivalent
  fallback.

The user explicitly removed the experiment credit ceiling so cost could be
observed. Deterministic no-paid conformance and 10/10 serialized policy probes
passed before live work.

## Route Attempts

| Route | Result |
|---|---|
| Qwen3.6-27B-FP8, 10,976 context | No inference: GHCP static context used 191% of available input tokens |
| Qwen3.6-35B-A3B, 32K, 5090 | No inference: controller selected an incompatible Fast member for a Quant-only deployment; pod remained unschedulable and was released |
| Qwen3-Coder-Next-NVFP4, 262K, Spark | Inference succeeded twice through BYOK; route identity matched; responses did not satisfy the JSON file contract |
| GPT-5.6 Sol | Inference and exact AIC telemetry succeeded; responses returned `{"files":[]}` and failed proof |

DS4 did not run. Its distributed lifecycle remains controller-unsupported.

## Raw Completed Measurements

### Sol direct samples

| Sample | Billed AIC | Input | Output | Model latency | End-to-end | Result |
|---|---:|---:|---:|---:|---:|---|
| Direct 1 | 13.21375 | 20,773 | 77 | 116.258 s | 193.625 s | invalid response: `files` empty |
| Fallback-equivalent | 13.61725 | 20,761 | 214 | 29.504 s | 70.033 s | invalid response: `files` empty |

Mean Sol billed AIC: **13.4155**.

### Qwen Coder Small samples

| Sample | Billed AIC | Input | Output | Model latency | Result |
|---|---:|---:|---:|---:|---|
| Qwen 1 | 0 / unavailable by design | 22,191 | 35 | 54.912 s | asked for task context; no JSON |
| Qwen 2 | 0 / unavailable by design | 22,193 | 85 | 34.102 s | prose response; no JSON |

Mean Qwen input/output: **22,192 / 60 tokens**.
Mean Qwen model latency: **44.507 s**.

## Cost Comparison

Because local Qwen has no GHCP-billed AIC, a full-fallback arm bills only the
hosted Sol fallback.

- Direct Sol mean: **13.4155 AIC**
- Qwen plus observed Sol fallback: **13.61725 AIC**
- Difference versus direct mean: **+0.20175 AIC (+1.50%)**
- Same-sequence direct/fallback samples: **+3.05%**

This single-sample difference is within obvious run variance. It does not prove
a billed-AIC saving or regression.

Token workload proxy:

- Direct mean: **20,912.5 total tokens**
- Qwen plus Sol fallback: **43,164.5 total tokens**
- Token overhead: **+106.4%**

The local attempt roughly doubles processed tokens before fallback.

## Correctness

No completed sample produced a valid implementation:

- Sol returned an empty `files` array.
- Qwen returned prose and no JSON object.
- The fixture remained at `panic("TODO")`; proof failed.

Therefore:

- cost per successful result is undefined;
- no AMR/direct winner exists;
- no route may be qualified, suspended, or promoted from this run;
- more repetitions would waste compute until prompt/response compatibility is
  fixed.

## Excluded / Invalid Samples

- Requested GPT-5.4 but GHCP executed Sol: 15.37 AIC, excluded for route
  mismatch.
- First explicit Sol retry timed out before priced telemetry: excluded; cost
  unavailable.
- First Qwen27 attempt never reached inference due context capacity.
- Qwen35 deployment never scheduled.
- One Sol correction attempt failed before inference because the Windows
  command line was too long.

Total measured hosted spend across all priced calls in this investigation:
**42.201 AIC**. One timed-out hosted attempt has unavailable cost.

## Artifact Integrity

| Artifact | SHA-256 |
|---|---|
| Sol direct result | `a9dc8c970acd4ff0fb1b6c4d8083e97d59ebb840c14fdccdd38c8b3343f3d2f9` |
| Sol direct OTel | `68fdac6bf1a4f5cb50869570b8c84e9b3f7c12db13e20f0c2b467cace3a4db8c` |
| Sol fallback result | `6f249aa68b36aa4f80dec95ba0da55e737c41057d9be825fe86c5c0d63d3f43d` |
| Sol fallback OTel | `18059597b523e99d4cfe738b3470810cfcc86abfac983f00c8d92eda7f4c936d` |
| Qwen27 result | `c8bf9c785e60628aec25785895f4eca2daf99bb22b55a66a8e0944a252c80bfb` |
| Qwen27 OTel | `2660ab5c7f4e4ea07ad603ee99ef9630476d18f707196db4f0a74f41ad6cce77` |
| Qwen Coder result | `aa1464cfa2c2a4603dc172296cfa2f27aa0ef6d7eef6ff90987140f45e9b6feb` |
| Qwen Coder OTel | `c237d81695c9abb151b76ba98b48601dd52578f439594794c55d8fd3c5d452fb` |

Raw artifacts:

`E:\working-skill-repo\.kb\amr-model-benchmark\qwen-sol-cost-20260713\`

## Cleanup

- Qwen27, Qwen35, and Spark Qwen deployments returned to replicas `0`.
- Reservation `rsv-e01eeacc7cd04b5db71f81d7ee36757e` released.
- SSH tunnel stopped.
- Temporary BYOK profile removed.
- Session-only attended runner re-locked.

## Next Falsifiable Step

Before another cost matrix:

1. add a tiny provider-contract fixture that proves each route returns the exact
   JSON file envelope before using the coding fixture;
2. reduce/parameterize static prompt context for smaller local windows;
3. pass correction prompts via a file/stdin-safe transport rather than Windows
   command-line arguments;
4. run one valid direct and one valid AMR pair;
5. only then collect the preregistered 20+ paired samples.
