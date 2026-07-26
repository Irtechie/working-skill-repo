# Proof Governor Slice 001 Result

Date: 2026-07-26
Slice: `slice-001`
Route: current orchestrator (`router-unavailable: binary-not-found`,
`no-qualified-route`)

## Outcome

Implemented sealed proof coverage, content-addressed receipts, exact semantic
identity, working-tree-aware input hashing, strict receipt validation, atomic
receipt writes, manifest-gate receipt validation, a public CLI validator, and
the portable JSON schema.

## TDD Proof

- RED: focused Go test failed to compile because the proof-governor contract
  types and functions did not exist.
- GREEN:
  `go test ./cmd/kbcheck -run Proof -count=1 -timeout 45s`
  passed.
- Protected oracle:
  `cmd/kbcheck/proof_governor_test.go`
  first reached green at SHA-256
  `6598e1a9cf5cc3d8dcb0e607e5ed940ed26993d47e544fe1c713887f0bdc0bdd`.
  The plan was then explicitly amended to add the missing public CLI-dispatch
  oracle; the accepted final SHA-256 is
  `5aeef634af709f68ccfda6d27556a287eef7c9a79240f5753c25b7f0597811a4`.

## Coverage

- Passing enumerated superset receipt reuses a requested subset.
- Tracked, dirty, and relevant untracked content changes invalidate reuse.
- Missing input, registry drift, expiry, failed/partial evidence, unknown
  coverage, namespace drift, command/argv, cwd, timeout, expected result, and
  environment changes reject reuse.
- Tampered receipts and stale `.proof.json` gate evidence fail closed.
- `proof-receipt-validate --receipt <path> --json` is exercised through the
  public `kbcheck` command dispatcher.

## Scope

Forecast implementation files changed: 5.

Discovered workflow-state/evidence files:

- `docs/results/2026-07-26-proof-governor-plan-to-work.md`
- `docs/results/2026-07-26-proof-governor-slice-001.md`
- `docs/plans/2026-07-26-020-kb-change-aware-proof-governor-manifest.md`
- `docs/plans/2026-07-26-021-tool-proof-coverage-contract-plan.md`
- `todo.md`

Unexplained implementation files: 0.

## QA

- Go formatting applied to the exact slice files.
- Backend/CLI only; browser and native GUI checks are not applicable.
- The legacy all-snapshot replay was not invoked because it is the unsafe
  behavior this manifest replaces. Focused deterministic proof is recorded
  above; a compact governed snapshot is deferred until the selector exists.
