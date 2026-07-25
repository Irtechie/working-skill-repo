# GHCP AIC Falsification Harness

## Purpose

`cmd/amrbench` proves deterministic readiness for a future attended GHCP/local
AMR experiment without spending AI credits or claiming promotion.

## Main Surfaces

- `internal/ghcpotel/` - bounded, exact leaf-call OTel accounting.
- `cmd/amrbench/` - no-paid conformance, fixture admission, context contracts,
  dry-run approval preview, artifact-derived paired grading, and disabled live
  runner boundary.
- `evals/amr-model-benchmark/` - fixed fixtures, proof hashes, known solutions,
  context artifacts/contracts, budget, and operator documentation.
- `cmd/kbcheck/model_routing_ghcp_release.go` - follow-on evidence validator.
  Deterministic evidence is not promoted; attended evidence also fails closed
  until an independent verifier exists.
- `docs/context/goals/ghcp-aic-falsification.md` - objective and approval gate.
- `docs/handoffs/active/2026-07-11-ghcp-aic-no-paid-readiness.md` - temporary
  Podman inventory and exact cleanup commands.

## Trust Boundaries

- `conformance --no-paid` constructs only `DisabledRunner`.
- `run --dry-run` validates the fixed config, fixture admission, context
  artifacts, route evidence, tiers, and budget, then emits an approval template
  with `paid_calls: 0` and `profiles_loaded: false`.
- Every non-dry `run` exits before config, profile, secret, or provider access.
  Paid execution is intentionally unavailable until a trusted human-approval
  verifier exists.
- Fixture proof files and context artifacts are path-bound and SHA-256 verified.
- Model output is limited to declared mutable files and cannot write Git
  controls, proof files, modules, new paths, or links.
- Paired grading consumes hash-bound direct/AMR run artifacts sharing task,
  seed, experiment, route-catalog, context, proof, and planned-route identity.

## Canonical Proof

```powershell
go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json
go test ./...
go run ./cmd/kbcheck core
```

Expected conformance fields:

```text
ready=true
no_paid=true
runner=disabled
paid_calls=0
release_decision=not-promoted
```

## Temporary Runtime

The current Windows proof uses rootless Podman with a digest-pinned Go image.
Podman is not a permanent project dependency. Keep its install/machine/image
paths in the active handoff and remove them only after no further proof or
approved experiment needs the runtime.
