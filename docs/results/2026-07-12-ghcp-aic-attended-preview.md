# GHCP AIC Attended Preview

Recorded: 2026-07-12

No model call or AIC spend is authorized by this artifact.

## 1. Cohort Matrix

| Cohort | Direct arm | AMR arm | Families | Current state |
|---|---|---|---|---|
| Hosted control | planned Medium | one Small attempt, then full Medium fallback | `go-local-logic` | model IDs/tier evidence not independently attested |
| Hosted cross-file | planned Large | one Medium attempt, then full Large fallback | `go-cross-file` | model IDs/tier evidence not independently attested |
| Qwen 5090 candidate | tier determined by a fresh bound probe | eligible only if an exact next-lower route and planned fallback are both proven | eligible Go family only | deployment exists; no profile; Plato is Ready but cordoned; not reserved |
| DS4 Spark candidate | expected planned Large pending probe | eligible only with proven Medium attempt and full DS4 fallback | eligible Go family only | requires two GB10 nodes; only spark-f147 is Ready and it is occupied; unavailable |

No route enters a matrix from its name alone. A strict route catalog must bind
model ID, runner/profile, observed tier, availability, timestamp, expiry, and
the hash of the probe artifact.

## 2. Task Fixtures

| Task | Family | Mutable files | Protected proof |
|---|---|---|---|
| `retry-after-parser` | `go-local-logic` | `retry/retry.go` | `retry/retry_test.go`, `go.mod` |
| `canonical-cache-key` | `go-cross-file` | `key/query.go`, `request/cache.go` | both test files and `go.mod` |

Each fixture mechanically proves baseline RED, known solution GREEN, and a
negative mutation RED in rootless Podman with no network.

HTML and subjective-policy fixtures are ineligible.

## 3. Exact AIC and Token Measurement Boundary

- Cost identity: unique leaf `(trace_id, span_id)` from the phase-local GHCP
  OTel export.
- Parent `invoke_agent` and non-leaf chat spans contribute zero.
- Cost authority: exact integer `github.copilot.nano_aiu`, or decimal
  `github.copilot.aiu` only when exactly representable in nano units.
- Token fields: input, output, cache-read, and cache-write tokens from the same
  credited leaf spans.
- Arm total: every valid direct phase, or every valid AMR attempt plus full
  fallback phase.
- Invalid rather than loss/failure: missing or partial telemetry, zero calls,
  route mismatch, process error, output overflow, oracle drift, isolation
  failure, application failure, or proof-infrastructure failure.
- Paired identity: task, seed, family, experiment ID, route-catalog hash,
  context-contract hash, proof-closure hash, and planned-route identity.

## 4. Maximum Credit Budget

From `evals/amr-model-benchmark/config.json`:

| Boundary | Maximum |
|---|---:|
| Per call | 5 AI credits |
| Per direct/AMR arm | 10 AI credits |
| One experiment ID across commands | 80 AI credits |

These are worst-case reservations, not measured spend. The durable user-local
ledger refuses duplicate approved matrices and cannot be reset by changing the
result output directory.

## 5. Local Model and DS4 Availability

Tinyboss read-only evidence:

- `qwen36-27b-fp8-vllm` and other Qwen deployments exist at replicas `0`.
- Qwen 5090 fast route selects `gpu-pool=5090`, `coder-role=fast`.
- Plato is the only Ready matching 5090; it is `SchedulingDisabled`.
- Socrates and Aristotle are NotReady/SchedulingDisabled.
- `deepseek-v4-sparkrun-head` and `worker` exist at replicas `0` and require two
  118 GiB / one-GPU GB10 placements.
- spark-f147 is Ready but currently runs `sfx-stable-audio-3`.
- gx10-b041 is NotReady/SchedulingDisabled.
- `~/.kb/amr-bench-models.json` is missing, so no local route is executable by
  the harness.

Result:

- Qwen: configured and potentially reservable after uncordoning Plato, but not
  profile-ready, not tier-attested, and not reserved.
- DS4: unavailable; the required second Ready GB10 node is absent and the sole
  Ready Spark is occupied.

## 6. Commands

Deterministic commands that have run:

```powershell
go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json
go test ./...
go run ./cmd/kbcheck core
.github\skills\kb-regression-snapshot\scripts\kb-regression-snapshot.ps1 verify
```

Zero-spend preview template:

```powershell
go run ./cmd/amrbench run `
  --dry-run `
  --experiment-id <approved-experiment-id> `
  --routes <hash-bound-routes.json> `
  --context baseline `
  --mode direct `
  --task retry-after-parser `
  --model <attested-medium-model-id> `
  --model-runner ghcp `
  --repeat 1 `
  --max-ai-credits 5 `
  --config evals/amr-model-benchmark/config.json
```

Paid commands that will run: **none**. Every non-dry `run` currently exits
before config, profile, secret, budget reservation, or provider access:

```text
attended execution is disabled until a trusted human-approval verifier is implemented; use --dry-run
```

## Approval Gate

Before any future spend:

1. implement and verify a trusted human-approval adapter;
2. restore two Ready Spark nodes for DS4 or remove DS4 from the matrix;
3. create user-local Qwen/DS4 profiles without committing endpoints/secrets;
4. run fresh route probes and bind their artifacts into the route catalog;
5. regenerate this preview with exact model IDs, tier evidence, reservations,
   experiment ID, and commands;
6. obtain explicit attended approval.
