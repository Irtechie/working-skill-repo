# `kbcheck` Subsystem

## Purpose

`cmd\kbcheck` is the native Go gate for this repo. It owns deterministic proof,
release gating, skill/eval scoring, sync drift inspection, marketplace
firebreaks, workflow-state validation, graph-routing checks, and model-routing
release validation.

## Read First

1. `cmd\kbcheck\main.go`
2. `cmd\kbcheck\checks.go`
3. `config\skill-quality.json`

Then branch by task:

- graph packets/evals: `cmd\kbcheck\graph_route.go`,
  `cmd\kbcheck\graph_routing_lifecycle.go`,
  `cmd\kbcheck\graph_routing_eval.go`
- proof spine: `cmd\kbcheck\proof_spine.go`
- manifests/run state: `cmd\kbcheck\manifest_contract.go`,
  `cmd\kbcheck\run_state.go`
- release/model routing: `cmd\kbcheck\model_routing_release.go`,
  `cmd\kbcheck\model_routing_ghcp_release.go`
- provider/install drift: `cmd\kbcheck\provider_hygiene.go`,
  `cmd\kbcheck\review_reference_guard.go`

## Command Groups

| Group | Commands | Use When |
|---|---|---|
| Core/release | `core`, `local-release`, `live-release`, `release-selftest` | Running contributor-safe or release-blocking proof |
| Workflow state | `ready-set`, `manifest-contract`, `gate-ledger`, `run-state`, `workflow-governor-selftest` | Validating manifests, route history, and phase advancement |
| Proof spine | `sense`, `trace-verify`, `accept`, `learning-adoption` | Proving RED→GREEN repair claims and adoption evidence |
| Context/telemetry | `context-packet`, `execution-telemetry` | Validating bounded worker inputs and measured result artifacts |
| Graph routing | `graph-route`, `graph-routing-lifecycle-selftest`, `graph-routing-eval` | Validating packets, lifecycle behavior, and deterministic corpus readiness |
| Model routing | `model-routing-release`, `provider-hygiene` | Checking route evidence boundaries and optional-provider hygiene |
| Skill quality / eval | `skill-lint`, `route-eval`, `skill-eval*`, `eval-run-*`, `surface-report`, `minimality` | Scoring skill docs, route fixtures, captured results, and surface size |
| Sync / marketplace | `skill-sync-report`, `doctor`, `review-reference-guard`, `marketplace-firebreak`, `marketplace-promote` | Inspecting or repairing install drift and enforcing reusable-skill policy |
| Concurrency / isolation | `scope-lease`, `slice-lease`, `worktree`, `terminal-cleanup`, `cargo-storage` | Local leases, worktree coordination, terminal retirement, and shared Cargo build storage |
| Reconciler conformance | `go test ./internal/reconcile ./cmd/kbreconcile ./cmd/kbcheck -run 'SemanticClaim|Fence|Idempot|DeliveryChain'` | Proving repo-native lifecycle and queue surfaces preserve global claim/fencing rules |

`manifest-contract` also validates opt-in `pre_slice_review_contract` receipts:
schema-v2 triggered plans must bind the review to the current requirements
SHA-256, use allowlisted persona-to-reason evidence, and have no failed personas
or unresolved P0/P1 findings. Proportional plans may instead record a specific
`not_required_reason`; manifests without a schema remain legacy-compatible.

## Canonical Commands

```powershell
go run ./cmd/kbcheck core --list
go run ./cmd/kbcheck core
go run ./cmd/kbcheck local-release
go run ./cmd/kbcheck graph-routing-eval --require-ready
go test ./cmd/kbcheck -run TerminalCleanup -count=1
go run ./cmd/kbreconcile claim-capability --json
go run ./cmd/kbreconcile claim-conformance --json
go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence evals/model-routing/initial-pilot-release-evidence.json
```

## Key Inputs

- `.github\skills\**\SKILL.md`
- `.github\agents\*.agent.md`
- `config\skill-quality.json`
- `config\skill-marketplace.json`
- `evals\**`
- `docs\plans\**`, `docs\context\**`, `.kb\**` depending command

## Sharp Edges

- `core` is intentionally repo-local. User-global inspection moves behind
  explicit commands like `provider-hygiene --include-user`.
- On Windows, the `go-test` check keeps ordinary packages in one contained
  aggregate but runs `cmd/kbcheck` and `cmd/kbrouter` separately without the
  outer job. Those packages own child-process containment themselves; nesting
  them can stall command fixtures and cleanup. The isolated runner gives test
  binaries an earlier Go timeout, bounds inherited output pipes with
  `WaitDelay`, and reports timeout/output termination as exit 124/125.
- `local-release` is the sync gate. A green `core` does not prove global roots
  are current.
- `graph-routing-eval --require-ready` is deterministic fixture proof only. It
  does not certify live graph providers.
- `model-routing-release` validates evidence honesty, not live AMR promotion.
- `terminal-cleanup` removes only registered Git worktrees from a different
  session after queue and delivery proof. Receipts bind a Git-admin generation
  marker, real path, and observed remote ref SHAs; sweep requires monotonic
  remote evolution and exact-SHA local-ref deletion. It refreshes exact remote
  authority immediately before each destructive action and fails closed when no
  remote default can be resolved. Sweep captures the primary checkout as its
  stable Git command context before removal, so `--root` may identify the target
  worktree itself. A retry may reconcile only an exact empty residual directory
  after a prior non-force removal, and only while the saved admin identity is
  gone and the receipt, queue claim, branch SHA, and delivery evidence still
  match; any residual data or identity drift remains blocked. Its shared queue lock preserves
  repository-inherited permissions/ACLs while using the same fail-closed OS
  file lock as the PowerShell queue helper; private model-routing locks retain
  strict private ACL hardening. Squash/rebase merge proof, host UI records, and
  remote feature refs remain host/provider-owned.
- `cargo-storage` resolves one absolute Cargo target from canonical repository
  identity across linked worktrees and unrelated repositories. Versioned,
  collision-resistant receipts live under the Git common directory; updates
  are serialized. `validate-ready` guards execution, `not-applicable` records
  no-Cargo runs, and final validation proves cleanup invariants. Temporary
  targets require an existing real approved root, ownership marker, and
  exact-path finalization.
- `eval-run-codex` and `eval-run-ghcp` dry-run surfaces are safe defaults; live
  runs are explicit and require host auth.

## Standalone Reconciler Contract

`kbreconcile` is a standalone global-capable CLI, not a `kbcheck` subcommand.
Its deterministic in-memory fixtures define canonical semantic resources,
authoritative CAS takeover, monotonic controller epochs, scoped authorization,
gateway high-water fencing, durable idempotency, and direct/alternate bypass
denial. They certify the reference contract only. Git-common-directory leases
remain local coordination. Without a live adapter proving atomic conditional
commit and gateway-only production credentials, protected mutation is
unavailable.
