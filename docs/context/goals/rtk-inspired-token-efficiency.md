# RTK-Inspired Token Efficiency

Status: complete
Created: 2026-06-21
Last updated: 2026-06-21

## Objective

Implement transferable RTK-style token-efficiency improvements without replacing KB map/plan discipline.

## Done Criteria

- `kbcheck core` defaults to compact pass output.
- Failure output preserves the failing command, stdout, stderr, and exit code.
- Verbose mode preserves pre-existing full passing output when needed.
- Follow-up token-efficiency candidates are parked instead of silently lost.

## Terminal Proof

- `go test ./cmd/kbcheck`
- `go run ./cmd/kbcheck core`
- `go run ./cmd/kbcheck core --verbose`
- `git diff --check`

## Current State

- Current artifact: `cmd/kbcheck/main.go`
- Next allowed action: none
- Last proof: `go test ./cmd/kbcheck`; `go run ./cmd/kbcheck core`; `go run ./cmd/kbcheck core --verbose`; `git diff --check`

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|
| Compact `kbcheck core` output | kb-fix | `cmd/kbcheck/main.go` | done | `go test ./cmd/kbcheck`; `go run ./cmd/kbcheck core`; `go run ./cmd/kbcheck core --verbose`; `git diff --check` |
| Command-aware failure summarizers | kb-plan | none | parked | needs design |
| Optional hook/proxy integration | kb-research | none | parked | not accepted by default |

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|

## Notes

- RTK lesson adopted: deterministic tools should suppress routine success noise and preserve failure signal.
- RTK lesson rejected as default: a command proxy cannot replace good backlog entries, app maps, or sliced plans.
