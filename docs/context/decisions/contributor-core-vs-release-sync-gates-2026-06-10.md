# Contributor Core vs Release Sync Gates

Date: 2026-06-10

## Decision

Keep `go run ./cmd/kbcheck core` contributor-safe and repo-local. Enforce global
install and ATV propagation drift in `go run ./cmd/kbcheck local-release` and
`go run ./cmd/kbcheck skill-sync-report`.

## Reason

Fresh clones must be able to prove repository-local quality without requiring
the contributor to have already synchronized personal/global skill installs or
ATV sibling copies. Maintainer release flow still needs drift protection before
propagation.

## Consequences

- `core` is the default contributor gate.
- `local-release` is the pre-sync/release gate.
- Required sync drift blocks release, not ordinary local contribution.
- Optional ATV scaffold/plugin drift remains warning-only unless that surface is
  intentionally being shipped.

## Related

- Compound note: `docs/solutions/workflow-issues/contributor-core-vs-release-sync-gates-2026-06-10.md`
- Config: `config/skill-quality.json`
- Gate entrypoint: `cmd/kbcheck`
