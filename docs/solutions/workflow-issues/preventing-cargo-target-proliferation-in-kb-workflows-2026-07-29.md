---
title: "Preventing Cargo Target Proliferation in KB Workflows"
date: 2026-07-29
category: workflow-issues
module: KB workflow
problem_type: workflow_issue
component: development_workflow
severity: high
applies_when:
  - "KB workflows invoke Cargo across phases, retries, workers, or worktrees"
  - "CARGO_TARGET_DIR must remain stable across repository work"
  - "Temporary Cargo targets require guarded cleanup"
tags:
  - cargo-storage
  - cargo-target-dir
  - build-storage
  - worktrees
  - kbcheck
  - cleanup-safety
---

# Preventing Cargo Target Proliferation in KB Workflows

## Context

KB agents created phase- and run-specific `CARGO_TARGET_DIR` directories such
as `target-check`, `target-repair`, `target-repro`, and probe targets. Every
fresh directory forced Cargo to rebuild the Rust/Tauri dependency graph. One
run accumulated 19.05 GB across several targets.

Prose-only prohibitions and token-presence tests were insufficient. They did
not define deterministic target selection, portable execution, concurrent
receipt updates, or safe cleanup.

## Guidance

Probe `cmd/kbcheck help` for the `cargo-storage` capability. When present,
resolve and validate one project-keyed absolute target:

```powershell
go run ./cmd/kbcheck cargo-storage --action resolve --run-id <run-id> --root <project-root> --json
go run ./cmd/kbcheck cargo-storage --action validate-ready --run-id <run-id> --root <project-root> --json
```

The resolver treats an external absolute `CARGO_TARGET_DIR` as a cache root,
appends a key derived from canonical Git common-directory identity, and reuses
the authoritative receipt across retries and worktrees. Receipt filenames are
collision-resistant, updates are serialized per run, and Cargo configuration
fingerprints must remain current.

When the capability is unavailable, use the fail-closed portable fallback:
derive the same project key beneath an existing external absolute cache root,
reuse an already-keyed path, prohibit temporary targets, and perform no
automated deletion.

Temporary targets are exceptional. Native mode requires a direct child of an
existing real approved root, a documented technical reason, and a random
ownership marker:

```powershell
go run ./cmd/kbcheck cargo-storage --action register-temp --run-id <run-id> `
  --target <absolute-target> --temp-root <approved-root> --reason "<reason>" `
  --root <project-root> --json
```

Finalization persists deletion intent before removing a target, allowing a
retry to reconcile an interrupted receipt write. Final validation requires the
exact registered removal set and nonnegative byte accounting:

```powershell
go run ./cmd/kbcheck cargo-storage --action finalize --run-id <run-id> --root <project-root> --json
go run ./cmd/kbcheck cargo-storage --action validate --run-id <run-id> --root <project-root> --json
```

Record runs with no Cargo command explicitly:

```powershell
go run ./cmd/kbcheck cargo-storage --action not-applicable --run-id <run-id> `
  --reason "no Cargo command selected" --root <project-root> --json
```

## Why This Matters

Stable project-scoped targets reuse Cargo artifacts without mixing unrelated
repositories. Receipt-backed validation makes temporary cleanup owned,
auditable, idempotent, and crash-safe. Capability probing keeps portable skills
usable in consuming repositories without requiring this bundle's Go command.

## When to Apply

- Any KB lane runs Cargo, Rust, or Tauri commands.
- Multiple workers, sessions, or linked worktrees share repository work.
- A repair or reproduction loop is tempted to create a fresh target.
- Temporary build isolation has a proven technical requirement.
- Global skill changes must propagate without exposing partial generations.

## Examples

Bad:

```powershell
$env:CARGO_TARGET_DIR = "$env:TEMP\agent-run\target-repair"
cargo test
```

Good:

```powershell
$receipt = go run ./cmd/kbcheck cargo-storage --action resolve `
  --run-id $runId --root $repo --json | ConvertFrom-Json
$env:CARGO_TARGET_DIR = $receipt.receipt.applied_environment.CARGO_TARGET_DIR
cargo test
```

## Related

- [Contributor core vs release sync gates](contributor-core-vs-release-sync-gates-2026-06-10.md)
- [Optional provider hygiene](optional-provider-hygiene-2026-07-09.md)
- [Plan-run worktree isolation](plan-run-worktree-isolation-2026-07-26.md)
- `docs/plans/2026-07-29-000-kb-cargo-build-storage-contract-manifest.md`
