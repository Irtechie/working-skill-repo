# Proof Governor Slice 003 Result

Date: 2026-07-26
Slice: `slice-003`
Route: current orchestrator (`router-unavailable: binary-not-found`,
`no-qualified-route`)

## Outcome

Implemented governed proof execution through `proof-run`. Passing proof is
reused without process creation, an identical failed fingerprint blocks an
immediate replay, and changed relevant input permits the affected check to run
again. Every RUN, REUSE, and BLOCK decision appends a bounded audit row.

Visible-browser and native-GUI checks fail before spawning unless an attended
approval matches the project, exact command hash, execution class, one-launch
budget, timeout, approver, and expiry. Approval artifacts are atomically
consumed on launch. `proof-approve` refuses non-interactive invocation.

Snapshot verification now requires selected IDs or a milestone. Milestones
fingerprint snapshot definitions and return `REUSE` unchanged. CLI checks have
a timeout and tree kill; DOM checks share one headless Playwright lifecycle.

## TDD Proof

- RED: focused tests failed to compile because execution and approval symbols
  did not exist.
- GREEN:
  - `go test ./cmd/kbcheck -run "ProofExecutionBudget|AttendedGUI|SnapshotSelection|RepairPolicy|ManifestContract" -count=1 -timeout 45s`
  - empty milestone smoke: first invocation `PASS 0/0`, second invocation
    `REUSE` with the same fingerprint.
- Protected oracle:
  `cmd/kbcheck/proof_execution_test.go`
  SHA-256 `de181281c6cc6f50e5c4b1dc085d14f2eb8ba40ad74ea185064a361c2ee95cea`.

The oracle was completed before its final hash with single-use approval and
decision-audit assertions; those directly cover acceptance criteria 5 and 6.

## Workflow Contract

- `kb-repair` reruns failed plus impacted checks and stops on identical failure.
- `kb-work` selects affected snapshots and reserves a full sweep for milestones.
- `kb-check`, `kb-qa`, `kb-functional-test`, `kb-finalize`, and `kb-ship` share
  receipt reuse, headless defaults, one aggregate release, and GUI approval.
- `kb-plan` and the manifest contract recognize `functional-native-gui`,
  execution class, and approval requirements.

## Scope and QA

All implementation and workflow files were within the slice acceptance scope.
Adjacent changes to `cmd/kbcheck/main.go` and `cmd/kbcheck/swarm.go` expose the
public commands and parse the new manifest fields. Plan/manifest state and this
result are workflow bookkeeping. Unexplained implementation files: 0.

No visible browser, WebView, Tauri, WDIO, or desktop process was launched.
Focused Go proof and the headless-free empty snapshot milestone smoke passed.
The full repository/release aggregate was intentionally not run in this slice.
