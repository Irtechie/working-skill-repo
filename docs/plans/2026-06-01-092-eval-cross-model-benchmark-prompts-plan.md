---
kb_id: kb-2026-06-01-cold-storage-follow-through
slice_id: slice-092
title: "Add cross-model benchmark prompt pack"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: evals/cross-model-benchmarks/README.md
    op: create
    scope: "Document benchmark purpose, prompt categories, scoring expectations, and live-run boundaries."
  - path: evals/cross-model-benchmarks/route-selection.json
    op: create
    scope: "Fixture prompts for route choice and complexity discipline."
  - path: evals/cross-model-benchmarks/proof-discipline.json
    op: create
    scope: "Fixture prompts that test whether agents demand machine-verifiable proof."
  - path: evals/cross-model-benchmarks/minimalism.json
    op: create
    scope: "Fixture prompts that test token-diet and deletion/trim judgment."
  - path: scripts/cross-model-benchmark-validate.ps1
    op: create
    scope: "Validate benchmark fixture schema without running live models."
  - path: .github/skills/kb-check/scripts/kb-check.ps1
    op: edit
    scope: "Discover the benchmark fixture validator when present."
protected_oracles: []
status: done
---

# Slice 092: Add Cross-Model Benchmark Prompt Pack

## What To Build

Create a benchmark prompt pack that can later be run across Codex, GHCP, Claude,
or other hosts, but make the default proof deterministic: fixture schema and
expected scoring rules only.

## Acceptance Criteria

- Benchmark fixtures cover route selection, complexity, proof discipline, and
  minimalism.
- Fixtures specify expected route/behavior, forbidden failure modes, and scoring
  dimensions.
- Validator checks schema and fails on missing expected outputs.
- `kb-check -All` runs the validator.
- README states live model execution is explicit and not part of local gates.

## Test Scenarios

- Run `pwsh -NoProfile -File scripts/cross-model-benchmark-validate.ps1`.
- Intentionally malformed temp fixture fails in selftest or validator negative
  case if a selftest is added.
- Run `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`.

## Scope Boundary

- Do not run live model calls in this slice.
- Do not add a scoring claim that requires subjective model judgment as proof.
- Do not replace the existing route-complexity eval; this is a cross-model
  prompt pack for later comparison.

## Completion Proof

- `scripts/cross-model-benchmark-validate.ps1` exited 0 with 3 files and 7
  cases.
- `go run .\cmd\kbcheck core --list` includes
  `cross-model-benchmark-validate`.
