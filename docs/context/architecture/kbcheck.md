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
| Concurrency / isolation | `scope-lease`, `slice-lease`, `worktree` | Local lease and isolated worktree coordination |

## Canonical Commands

```powershell
go run ./cmd/kbcheck core --list
go run ./cmd/kbcheck core
go run ./cmd/kbcheck local-release
go run ./cmd/kbcheck graph-routing-eval --require-ready
go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json
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
- `local-release` is the sync gate. A green `core` does not prove global roots
  are current.
- `graph-routing-eval --require-ready` is deterministic fixture proof only. It
  does not certify live graph providers.
- `model-routing-release` validates evidence honesty, not live AMR promotion.
- `eval-run-codex` and `eval-run-ghcp` dry-run surfaces are safe defaults; live
  runs are explicit and require host auth.
