# AMR Model Benchmark

This benchmark compares direct model execution with one-loop AMR:

```text
lower-tier attempt -> deterministic proof
  -> pass: keep work
  -> fail: one planned-tier surgical correction -> full proof
```

It is intentionally separate from production `kbrouter` state. Every run copies
a known-answer fixture into `.kb/`, initializes a disposable Git repository,
runs GitHub Copilot CLI with bounded write and verification permissions, and
records exact leaf-call AIC from OTel.

## Safety

- Fixtures contain no secrets or private source.
- Models can write only inside the disposable workspace.
- Only `write`, `go`, and `node` tools are pre-approved.
- Tests are protected oracles. Trusted code rejects test edits case-insensitively.
- Generated code is proved only inside a read-only, network-disabled Docker or
  Podman container with a minimal environment. No host-execution fallback exists.
- Proof images must be supplied as immutable digests in `AMRBENCH_GO_IMAGE` and
  `AMRBENCH_NODE_IMAGE`.
- Proof requires structured pass events for every declared oracle, not merely a
  zero process exit.
- A model gets one attempt. AMR gets at most one driver correction.
- Results and raw OTel live under `.kb/amr-model-benchmark/`.

## Hosted Models

```powershell
go run ./cmd/amrbench run --mode direct --task retry-after-parser --model terra
go run ./cmd/amrbench run --mode amr --task retry-after-parser --attempt haiku --driver sol
```

Run a model/task matrix by invoking the command for each desired pair. Use
`--repeat` for independent samples; five samples are required before a model can
leave probation.

## Local Models

Local endpoints are configured outside the repository:

`~/.kb/amr-bench-models.json`

The `profile` value in `config.json` must exactly match one profile `alias` in
this user-local file. The profile supplies endpoint/model settings without
putting credentials or private routes in the repository.

```json
{
  "profiles": [
    {
      "alias": "qwen-local",
      "base_url": "http://local-gateway.example/v1",
      "provider_type": "openai",
      "model_id": "gpt-5.4",
      "wire_model": "your-qwen-model",
      "wire_api": "completions",
      "api_key_env": "LOCAL_MODEL_API_KEY",
      "max_prompt_tokens": 32768,
      "max_output_tokens": 4096
    }
  ]
}
```

Use a recognized `model_id` only when its tool/prompt behavior is genuinely
compatible. The `wire_model` is what the endpoint receives. Credentials remain
in environment variables.

## Qualification

```powershell
go run ./cmd/amrbench grade
```

Grades are per model and task family:

- `probation`: fewer than five samples or middling evidence;
- `qualified`: at least 80% deterministic pass rate;
- `suspended`: below 60% pass rate after the minimum sample count.

Suspension is scoped. A model may be qualified for HTML while suspended for
cross-file Go work. Requalification requires fresh fixed-corpus runs; no prompt
"learning" is injected into later attempts.

## Claims This Can Test

- Direct-driver correctness and AIC by model/task family.
- Whether a lower-tier attempt plus correction costs less than direct driver.
- Whether AMR preserves final correctness.
- How much code the attempt contributes before correction.
- Which model/tier combinations should be qualified, left on probation, or
  suspended.

It cannot prove speculative-decoding equivalence or production suitability from
a handful of samples.
