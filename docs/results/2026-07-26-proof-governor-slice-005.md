# Proof Governor Slice 005 Result

Date: 2026-07-26
Slice: `slice-005`
Route: current orchestrator (`router-unavailable: binary-not-found`,
`no-qualified-route`)
State: done; slice-004 release blocker unchanged

## Outcome

Removed the repo-owned GUI approval ceremony. `proof-run` now unconditionally
blocks `visible-browser` and `native-gui` checks before calling the process
runner and reports
`automatic-gui-execution-disabled:<check-id>`.

Deleted `proof-approve`, `--approval`, `--approval-ttl`, approval JSON parsing,
unkeyed integrity hashes, expiry/identity validation, and single-use marker
files. Manifests now use `execution_class` alone. Headless/CLI execution,
working-tree-aware receipts, changed-input invalidation, passing-proof reuse,
identical-failure blocking, timeouts, and audit rows remain intact.

An explicitly requested attended GUI session, if needed, is a bounded user/host
action outside the portable proof runner. This repo records that blocker and
does not implement another HITL/token workflow.

## TDD and Focused Proof

- Protected RED:
  `go test ./cmd/kbcheck -run '^TestProofExecutionBlocksAutomaticGUIClassesBeforeSpawn$' -count=1 -timeout 30s`
  failed for both GUI classes because the old implementation returned
  `attended-approval-required:desktop`; runner count remained zero.
- GREEN:
  `go test ./cmd/kbcheck -run 'ProofExecution|ProofGovernorCLI|ManifestContract' -count=1 -timeout 45s`
  passed.
- End-to-end no-GUI proof:
  `go run ./cmd/kbcheck proof-governor-selftest` passed
  `scenarios=12 gui_launches=0`.
- Manifest contract passed after removing `approval_required`.
- `skill-lint` passed with 0 errors and 12 existing warnings.
- `git diff --check` passed with existing line-ending warnings only.
- Protected oracle:
  `cmd/kbcheck/proof_execution_test.go`
  SHA-256 `80672702ed84bf6c7ffffdf706af8966df11c7b6d2031c966bdd6779673dc03e`.

No visible browser, WebView, Tauri, WDIO, or native desktop process launched.

## Skill Sync

Reviewed `kb-plan`, `kb-work`, `kb-functional-test`, and `kb-qa` against Codex,
Copilot, and shared Agents installs. All three installed roots were identical
before sync; the repo delta contained only this intended policy correction.
Only those four `SKILL.md` files were copied. Each skill is now hash-identical
across all four roots.

## Release Boundary

The full `core` and composing `local-release` gates were not replayed. Their
known inputs did not change: twelve existing `cmd/kbrouter`
canonical-project-path failures and unrelated `kb-gate` /
`safe-shell-quoting` sync drift still block slice-004. After those inputs
change, run one fresh `core` and one `local-release`.

No files were staged, committed, or pushed.
