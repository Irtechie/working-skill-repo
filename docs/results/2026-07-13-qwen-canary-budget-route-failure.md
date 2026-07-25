# Qwen Canary Safety Stop

Recorded: 2026-07-13

## Decision

The approved partial canary stopped after the hosted direct arm. The AMR/Qwen
arm was not launched.

The user subsequently clarified that the credit value was a measurement target,
not a hard stop. The 15.37-AIC call remains excluded because GHCP replaced the
requested GPT-5.4 route with Sol. Follow-on exploratory measurements are
recorded in `docs/results/2026-07-13-qwen-sol-exploratory-cost.md`.

## Approved Contract

- Task: `retry-after-parser`
- Direct route: hosted `gpt-5.4`
- AMR route: Qwen Small attempt plus hosted `gpt-5.4` Medium fallback
- Direct maximum: 5 AI credits
- AMR maximum: 10 AI credits
- Experiment maximum: 15 AI credits for this approved canary
- Policy-hook gate: 10/10 serialized probes passed
- Deterministic conformance: `ready=true`, `runner=disabled`, `paid_calls=0`

Serialized hook latencies:

| Probe | Latency |
|---:|---:|
| 1 | 155 ms |
| 2 | 151 ms |
| 3 | 118 ms |
| 4 | 40 ms |
| 5 | 101 ms |
| 6 | 131 ms |
| 7 | 163 ms |
| 8 | 149 ms |
| 9 | 116 ms |
| 10 | 53 ms |

## Observed Direct Arm

```text
run_id: 20260713T141426.442229600Z-retry-after-parser-direct
process_exit: 0
duration_ms: 53155
outcome: invalid-direct
proof: not run
changed_files: none
```

The GHCP command was invoked with:

```text
--model gpt-5.4
--max-ai-credits 5
```

Bounded OTel evidence:

```text
span_name: chat gpt-5.6-sol
requested_model: gpt-5.6-sol
actual_model: gpt-5.6-sol
input_tokens: 20767
output_tokens: 797
cache_read_tokens: unavailable
cache_write_tokens: unavailable
nano_aiu: 15370000000.0
AIU: 15.37
```

Two independent invalidity conditions occurred:

1. GHCP ignored or silently replaced the requested `gpt-5.4` route with
   `gpt-5.6-sol`.
2. GHCP consumed 15.37 AIU despite `--max-ai-credits 5`, exceeding both the
   per-call limit and the approved 15-AIC total canary ceiling before AMR.

The exporter also encoded legacy `nano_aiu` as an integral-valued decimal
(`15370000000.0`). The strict parser rejected that lexical form before route
matching; this is a schema-compatibility issue, but accepting the numeric form
would not make the run valid because the route and budget were already wrong.

## Safety Response

- No correction/fallback launched.
- No AMR/Qwen call launched.
- Qwen reservation `rsv-b1200b23e4984b13b73eae4a9cecf352` was released.
- `qwen36-27b-fp8-vllm` was scaled back to `0`.
- DS4 was never started; its controller lifecycle remains unsupported and its
  Spark nodes were unavailable.

## Artifact Integrity

| Artifact | SHA-256 |
|---|---|
| `result.json` | `5bace385686b0f303198639d186ecbf4fc4390e3cb3c9ca12d024e2793006734` |
| `direct/otel.jsonl` | `1aaeb6a3e6e71855b6674a98f33fec43da70ee5a7b83b7115b996db6ac79de70` |
| `direct/stdout.txt` | `5d537ccd385c750754b4f428b233cbd282ba63cd0ea4954b384b6466cb99644a` |
| `direct/stderr.txt` | `9704ad051617a9aea7af81d501deaf4e78d6b855c9e3dcecb8f81a15535d98a6` |

Raw artifacts remain under:

`E:\working-skill-repo\.kb\amr-model-benchmark\qwen-canary-20260713\`

## Resume Condition

Do not spend further AIC until GHCP proves both:

1. requested model identity is honored and reflected in OTel;
2. `--max-ai-credits` is a hard pre-dispatch ceiling rather than a soft or
   post-hoc limit.

A fresh canary must repeat the 10/10 serialized policy probes, deterministic
conformance, exact route preview, and attended approval.
