# Proof Governor Slice 002 Result

Date: 2026-07-26
Slice: `slice-002`
Route: current orchestrator (`router-unavailable: binary-not-found`,
`no-qualified-route`)

## Outcome

Implemented the read-only `proof-plan` selector and strict check registry. It
returns `RUN`, `REUSE`, or `BLOCK`, reuses passing superset receipts, invalidates
changed dependencies with exact path reasons, collapses duplicate/composite
requests, blocks unknown checks, and refuses pre-integration namespace receipts.

The local release profile no longer executes `skill-surface-minimality` twice:
`core` remains its single child owner for that immutable release state.

## TDD Proof

- RED: focused tests failed to compile before the registry and selector existed.
- GREEN:
  - `go test ./cmd/kbcheck -run "ProofSelection|CoverageSubsumption|ImpactInvalidation|ReleaseProfile" -count=1 -timeout 45s`
  - `go test ./cmd/kbcheck -run ProofPlan -count=1 -timeout 45s`
- Protected oracle:
  `cmd/kbcheck/proof_selection_test.go`
  SHA-256 `79fe938cefa3070f9252e7e29573ac24a30e0023e0abfbb01ea2435d888ad439`.

## Coverage

- A full passing receipt satisfies later Rust and CLI subset requests.
- A Rust-only edit reruns Rust while retaining browser reuse.
- A shared dependency invalidates every declared dependent.
- A requested composite profile collapses its covered child checks to one run.
- Unknown checks block before process execution.
- Worker/pre-integration namespace evidence cannot replace integrated proof.
- `proof-plan` is exercised through the public command dispatcher and is
  verified not to mutate the fixture tree.

## Scope and QA

Forecast files changed: 4.

Discovered files:

- `cmd/kbcheck/proof_selection.go`
- `cmd/kbcheck/proof_selection_test.go`
- `docs/results/2026-07-26-proof-governor-slice-002.md`
- workflow state in the manifest, slice plan, and `todo.md`

`cmd/kbcheck/proof_governor.go` was a justified adjacent change to normalize
Windows path casing. Unexplained implementation files: 0.

Go formatting and `git diff --check` passed for the exact slice code. This is a
backend/CLI slice; browser/native GUI proof is not applicable. Legacy
all-snapshot replay remains deferred until slice 003 replaces its execution
path.
