---
title: Proof spine digests must include all check semantics
date: 2026-07-05
last_updated: 2026-07-12
category: logic-errors
module: cmd/kbcheck proof spine
problem_type: logic_error
component: tooling
symptoms:
  - A RED trace could be recorded with one timeout and later accepted against the same command with a different timeout.
root_cause: logic_error
resolution_type: code_fix
severity: high
tags: [proof-spine, trace-integrity, kbcheck, verification]
---

# Proof spine digests must include all check semantics

## Problem

`kbcheck accept` relies on a check digest to decide whether prior RED/GREEN trace
events apply to the current check. If the digest omits a field that changes the
check's behavior, an agent can accidentally reuse proof from a different check.

## Symptoms

- A trace history looks intact, but it was produced under different check
  semantics.
- A repair can appear to satisfy RED-before-GREEN even though the current sensor
  is easier, looser, or otherwise different from the sensor that recorded RED.

## What Didn't Work

- Hashing only `kind`, `target`, `expect`, and the target file content was not
  enough. `timeout_ms` changes command behavior too.
- Relying on the current GREEN run alone would reintroduce the latest-green proof
  gap the proof spine is meant to close.

## Solution

Include every behavior-changing field in the canonical check digest. For the
current proof spine that means positive `timeout_ms` values are part of the
digest, and tests prove the digest changes when the timeout changes.

Also reject fractional expected exit codes instead of silently truncating JSON
numbers such as `1.5` to `1`.

## Why This Works

The trace is only useful when each event is bound to the exact same sensor. By
including timeout semantics in the digest, the RED event and the final
acceptance check must agree on the command, expected result, target file content
when applicable, and timeout behavior.

## Prevention

- When adding a new `ProofCheck` field, decide whether it changes sensing
  semantics. If yes, add it to `proofCheckDigest`.
- Add a digest-change regression test for each new semantic field.
- Keep `accept` strict: RED-before-GREEN proof should fail closed when the check
  definition changes.

### Apply the same rule to evaluation harnesses

The GHCP AIC falsification harness exposed the same failure mode at a larger
boundary: labels such as `route_match: true`, `oracle_intact: true`, or a
`sha256:` prefix are not evidence unless the harness loads the referenced
artifact and recomputes the claim.

For paid or promotion-sensitive evaluations:

- bind proof files, context packets, role overlays, route probes, and run
  results to real artifact paths plus exact SHA-256 values;
- derive paired grades from those bound run artifacts instead of caller-authored
  booleans;
- use the same admission function for no-paid conformance and any future live
  execution;
- reject changed, missing, extra, linked, or out-of-root proof inputs;
- count exact nano-AIU only from complete leaf-call telemetry;
- keep deterministic evidence `not-promoted`;
- hard-disable paid execution when trusted human approval or independent live
  verification does not yet exist.

This turns the digest rule into a general principle: every field that can change
readiness, cost, correctness, pairing, or promotion must either be recomputed
from protected evidence or remain unable to authorize the next phase.

## Related Issues

- `cmd/kbcheck/proof_spine.go`
- `cmd/kbcheck/proof_spine_test.go`
- `cmd/amrbench/`
- `internal/ghcpotel/`
- `evals/amr-model-benchmark/`
