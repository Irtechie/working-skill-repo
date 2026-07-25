# AMR Model Benchmark

This benchmark compares direct model execution with one-loop AMR:

```text
lower-tier attempt -> deterministic proof
  -> pass: keep work
  -> fail: one planned-tier full fallback -> full proof
```

It is separate from production `kbrouter` state. Every attended run uses a
known-answer fixture in a disposable workspace and records exact leaf-call AIC
from GHCP OTel.

## Safety

- Fixtures contain no secrets or private source.
- Model writes are limited to declared mutable files.
- Tests, modules, schemas, and proof dependencies are protected oracles.
- Generated code is proved in a read-only, network-disabled Podman/Docker
  container with a minimal environment.
- Proof images use immutable digests.
- Proof requires structured pass events for every declared test.
- AMR gets one lower-tier attempt and at most one full planned-tier fallback.
- Results and raw OTel remain under `.kb/amr-model-benchmark/`.

## Deterministic Readiness

```powershell
go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json
```

This terminal no-paid gate constructs only `DisabledRunner`, loads no provider
profile or secret, launches no model process, and requires:

- rootless Podman/Docker proof with no network;
- an immutable Go image digest;
- fixture baseline RED, known solution GREEN, and negative mutation RED;
- protected proof-closure hashes;
- disjoint, parity-controlled baseline/minimal context contracts;
- positive per-call, per-arm, and experiment credit ceilings.

Current maximum reservations from `config.json` are 5 credits per call, 10 per
direct/AMR arm, and 80 for one attended experiment process. They are ceilings,
not measured spend.

## Attended Models

Concrete model IDs and routes are selected from live host/fleet evidence
immediately before an attended run; they are not stored in `config.json`.
The exact command matrix requires user approval after route availability is
proven.

Every attended `run` requires:

- `--routes <catalog.json>` with each model's observed availability/tier and a
  hash-bound probe artifact;
- `--experiment-id <id>` shared by direct and AMR arms;
- `--approval <receipt.json>` bound to the exact config, route catalog, context,
  task, model/profile choices, repeats, and budget.

For this milestone, provider execution remains hard-disabled even when those
artifacts are present because no trusted human-approval verifier exists yet.
Only `run --dry-run` is executable; it is the approval-preview surface.

First run the same command with `--dry-run` and without `--approval`. It performs
fixture/context admission, loads no profile, reserves no credits, starts no
provider, and emits the exact approval template. A human fills only
`approved_at` and `expires_at` after reviewing that template.

Durable worst-case reservations live under the result root at
`experiments/<experiment-id>/budget.json`. A crash conservatively retains the
reservation; it never resets the experiment ceiling.

Local endpoints are configured outside the repository in
`~/.kb/amr-bench-models.json`. Profiles supply endpoint/model settings without
putting credentials or private routes in the repository. Profile presence is
availability evidence only; fleet reservation and a bounded probe remain
required before an attended run.

## Paired Grading

```powershell
go run ./cmd/amrbench grade-paired --results <paired-results.jsonl>
```

The grader pairs direct and AMR by task and seed, rejects incomplete telemetry,
route mismatch, oracle mutation, missing proof, and isolation failure, and
reports correctness, aggregate/median AIU, latency p90, fallback count, repair
tail, and interventions per family.

`promotion_eligible` requires at least 20 samples, zero right-to-wrong outcomes,
non-inferior correctness, lower aggregate AIU, at least 20% paired median
reduction with a confidence lower bound above zero, and no intervention
regression. Deterministic conformance and simulated rows always retain
`release_decision: not-promoted`.

## Claims This Can Test

- Direct correctness and complete AIC by task family.
- Whether one lower-tier attempt plus full fallback costs less than direct.
- Whether AMR preserves final correctness.
- Fallback frequency, repair tail, latency, and interventions.
- Which family/model combinations have enough attended evidence to enter a
  separate follow-on promotion validator.

It cannot prove speculative-decoding equivalence, partial-reuse economics, or
production suitability from deterministic fixtures or a handful of samples.
