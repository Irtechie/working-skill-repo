# `kbrouter` / Model Routing Subsystem

## Purpose

`cmd\kbrouter` discovers callable routes and applies delegation-first DDR.
Configured local routes are imported into canonical user-local state, approved
separately, and receive at most one eligible bounded attempt before control
returns to the active parent.

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
| `models import` | Strictly import an operator-filled route file into canonical `~/.kb/models.json` without changing trust |
| `models add` | Add an optional route definition at user or project scope |
| `models remove` | Remove a route alias |
| `models prefer` | Mark a route preferred |
| `models clear` / `reset` | Clear preference or stored state |
| `models approve` / `revoke` / `deny` | Manage attended-approval state |
| `models ignore-routing` | Disable optional routing for a scope |
| `models doctor` | Inspect routing state and configuration health |
| `models calibrate` | Calibrate a route alias |
| `ddr attempt` | Reserve and run one bounded trusted local attempt, returning a result once or a parent-return / required-pin-block receipt |
| `ddr resolve` | Finalize an awaiting local result from the parent's deterministic proof receipt |

## Source-of-Truth Rules

- Planning stores minimum capability tiers (`small`, `medium`, `large`), not
  concrete models.
- The active runner resolves that portable tier at plan pickup. Any compatible
  CLI or host may legitimately select a different concrete model for the same
  slice; the chosen model belongs in the runtime receipt.
- The orchestrator owns planning, minimum-tier judgment, selection, supervision,
  proof, and synthesis. A qualified subagent normally owns bounded execution.
- Ownership is singular per slice, not per plan. The KB coordinator may select
  and dispatch many isolated ready slices in parallel; `kbrouter` returns one
  route for each individual selection request.
- Current execution requires one recognized exception reason:
  `reasoning-required`, `context-required`, `tool-required`,
  `authority-required`, `trust-required`, `user-required`, or
  `no-qualified-route`.
- `no-qualified-route` must be proved against the active host surface and the
  CLI/user-local catalog. `kbrouter` can disprove it from eligible CLI routes;
  the caller remains responsible for checking host-native targets.
- `kbrouter` selects Codex CLI and optional user-local routes only. Native App
  targets execute through the active host's exact callable-agent tool and do
  not enter this catalog. App and CLI identities remain distinct unless an
  explicit adapter proves a route callable.
- Delegated selection returns one qualified same-tier-or-higher route. It does
  not automatically route downward or fall back to current.
- `self-hosted-first` applies only after capability, approval, tools, context,
  and normal-risk checks. Broad risk stays with automatic parent/host selection.
- Local DDR records one receipt per canonical project/run/slice under `~/.kb`.
  The reservation is durable before network I/O, so crash recovery cannot
  redispatch an uncertain attempt. Replays never contact the endpoint again.
  Because successful response content is not persisted, replay before proof returns
  `result-not-retained` to the parent instead of claiming an unavailable result.
- Exit `10` means structured `parent-return`; the parent continues with its
  active model or host-native selector. `require <alias>` instead blocks that
  slice. No fallback model/provider name is embedded in policy.
- User-local and project-local route state is runtime configuration, not a repo
  source-of-truth replacement for manifests or proof.

Related evidence and docs:

- `docs/context/goals/session-model-routing.md`
- `docs/results/2026-07-10-session-model-routing-initial-pilot.json`
- `cmd/kbcheck/model_routing_release.go`

## Canonical Commands

```powershell
go run ./cmd/kbrouter --help
go test ./cmd/kbrouter -run 'Import|DDRAttempt|Catalog|Doctor|Policy'
go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json
```

## Sharp Edges

- Attended approval requires an interactive console; redirected approval is
  refused.
- Optional self-hosted/private routes are kept in user/project runtime state,
  not committed into KB manifests.
- Import reads one bounded regular non-symlink JSON file, rejects unknown fields
  and secret values, and atomically writes only canonical route state. It never
  creates, renews, or changes `trust.json`.
- Discovery can probe trusted OpenAI-compatible endpoints, but only when the
  explicit probe path is chosen.
- DDR attempt receipts record latency. The response body is returned once to the
  parent but omitted from the persisted receipt. `ddr resolve` records the
  parent's later deterministic proof verdict.
- Windows junctions and canonical paths resolve through the same filesystem
  object identity; missing roots fail closed.
- AMR remains an experimental benchmark and is not part of normal selection.
