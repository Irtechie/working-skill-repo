# `kbrouter` / Model Routing Subsystem

## Purpose

`cmd\kbrouter` discovers callable routes and applies an owner-first DDR
decision: validate current execution or select exactly one delegated worker by
required tier and risk. It stores optional user/project route preferences and
enforces attended approval for sensitive route use.

## Read First

1. `cmd\kbrouter\main.go`
2. `cmd\kbrouter\catalog.go`
3. `cmd\kbrouter\dispatch.go`
4. `cmd\kbrouter\select.go`
5. `internal\modelrouting\*.go`

## CLI Surface

| Command | Purpose |
|---|---|
| `dispatch` | Execute a route-bound worker with run/slice context |
| `models show` | Print current user catalog + project policy |
| `models discover` | Probe and save a run-scoped redacted catalog |
| `models select` | Validate current ownership or pick one delegated route for a required tier / task family / risk |
| `models priority` | Save or clear project source preference (`automatic`, `self-hosted-first`, `native-first`) |
| `models add` | Add an optional route definition at user or project scope |
| `models remove` | Remove a route alias |
| `models prefer` | Mark a route preferred |
| `models clear` / `reset` | Clear preference or stored state |
| `models approve` / `revoke` / `deny` | Manage attended-approval state |
| `models ignore-routing` | Disable optional routing for a scope |
| `models doctor` | Inspect routing state and configuration health |
| `models calibrate` | Calibrate a route alias |

## Source-of-Truth Rules

- Planning stores minimum capability tiers (`small`, `medium`, `large`), not
  concrete models.
- The orchestrator chooses `current` or `delegated` once at work time.
- `kbrouter` selects Codex CLI and optional user-local routes only. Native App
  targets execute through the active host's exact callable-agent tool and do
  not enter this catalog. App and CLI identities remain distinct unless an
  explicit adapter proves a route callable.
- Delegated selection returns one qualified same-tier-or-higher route. It does
  not automatically route downward or fall back to current.
- User-local and project-local route state is runtime configuration, not a repo
  source-of-truth replacement for manifests or proof.

Related evidence and docs:

- `docs/context/goals/session-model-routing.md`
- `docs/results/2026-07-10-session-model-routing-initial-pilot.json`
- `cmd/kbcheck/model_routing_release.go`

## Canonical Commands

```powershell
go run ./cmd/kbrouter --help
go test ./cmd/kbrouter -run Catalog|Doctor|Policy
go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json
```

## Sharp Edges

- Attended approval requires an interactive console; redirected approval is
  refused.
- Optional self-hosted/private routes are kept in user/project runtime state,
  not committed into KB manifests.
- Discovery can probe trusted OpenAI-compatible endpoints, but only when the
  explicit probe path is chosen.
- Route receipts record what ran; proof still belongs to downstream checks.
- AMR remains an experimental benchmark and is not part of normal selection.
