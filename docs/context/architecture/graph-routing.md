# Graph Routing Surface

## Purpose

This repo treats graph routing as optional structural evidence layered on top of
file-native KB lookup. The source of truth is the provider-neutral impact packet
contract plus deterministic lifecycle/eval fixtures.

## Read First

1. `.github\skills\kb-map\references\graph-routing.md`
2. `config\graph-route.schema.json`
3. `cmd\kbcheck\graph_route.go`
4. `cmd\kbcheck\graph_routing_lifecycle.go`
5. `cmd\kbcheck\graph_routing_eval.go`
6. `internal\graphrouting\*.go`
7. `evals\graph-routing\README.md`

## Preflight Result

`graphify-size-check: 2026-07-21 code_files=1797 project_md_bytes=9276 decision=use`

Reason:

- repo size is far above the `>=200` code-file threshold
- `graphify.exe` is installed locally
- normal `kb-map lookup` should still stay doc-first; use graph-assisted refresh
  only when a task needs structural traversal or blast-radius work

## Main Surfaces

| Surface | Files |
|---|---|
| Packet contract | `config\graph-route.schema.json`; `cmd\kbcheck\graph_route.go` |
| Lifecycle validation | `cmd\kbcheck\graph_routing_lifecycle.go`; `cmd\kbcheck\graph_routing_lifecycle_test.go` |
| Deterministic eval corpus | `cmd\kbcheck\graph_routing_eval.go`; `evals\graph-routing\expected-results.json`; `evals\graph-routing\fixtures\*.json` |
| Optional exact-symbol adapter | `internal\graphrouting\scip.go`; `internal\graphrouting\scip_test.go`; `evals\graph-routing\exact-symbol-index.json` |
| Optional graphify adapter | `internal\graphrouting\graphify.go`; `internal\graphrouting\graphify_test.go` |
| Traversal recipes | `evals\graph-routing\traversal-recipes.json` |

## Evidence Rules

- Validate packets with `go run ./cmd/kbcheck graph-route --packet <packet.json>`.
- File-native fallback is valid proof; optional providers must downgrade
  explicitly when missing, stale, unsupported, or fingerprint-mismatched.
- Exact or observed edges need source spans.
- LLM-inferred edges are never exact evidence.
- Dense provider output should not be copied into `PROJECT.md`; route via named
  docs or graph packet references instead.

## Canonical Commands

```powershell
go run ./cmd/kbcheck graph-route --packet <packet.json>
go run ./cmd/kbcheck graph-routing-lifecycle-selftest
go run ./cmd/kbcheck graph-routing-eval --require-ready
```

## Known Limits

- Current checked-in proof is fixture-first; no live provider packet is treated
  as authoritative.
- Local slice leases/worktree receipts prove clone-family coordination only, not
  cross-machine locking.
- `uv` is not installed locally, but raw graphify use remains possible because
  `graphify.exe` is present.
