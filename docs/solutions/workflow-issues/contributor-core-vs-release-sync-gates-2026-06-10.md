---
title: Keep Contributor Core Gates Separate From Release Sync Gates
date: 2026-06-10
category: docs/solutions/workflow-issues
module: skill-repo-quality-gates
problem_type: workflow_issue
component: development_workflow
severity: medium
applies_when:
  - A repo has both fresh-clone quality checks and environment-dependent deployment or sync checks.
  - A documented default gate should be usable by contributors without local global installs.
tags: [core-gate, release-gate, sync-drift, installer-profile, contributor-onramp]
---

# Keep Contributor Core Gates Separate From Release Sync Gates

## Context

The skill bundle used `go run ./cmd/kbcheck core` as its front-door quality
gate, but `core` also ran `skill-sync-report`. That made a valid fresh or
partially installed environment fail on missing global skill roots before the
agent could distinguish repo-local quality from deployment drift.

The same audit found that the installer's `core` profile was not a dependency
closure: it installed only six startup skills even though those skills route
into the wider KB runtime.

## Guidance

Keep two gates with different contracts:

- `core`: contributor-safe, repo-local deterministic proof. It should not
  require personal global installs or adjacent sibling repos.
- `local-release`: release/sync proof. It composes `core`, `git diff --check`,
  and blocking sync drift reports.

Install profiles should match the runtime contract. If the "core" runtime tells
agents to invoke downstream KB skills, install that dependency closure rather
than leaving missing-skill fallbacks for normal paths.

## Why This Matters

Fresh contributors and agents need a command that answers "is this checkout
healthy?" without requiring the maintainer's local deployment state. Release
automation still needs the stricter answer: "do deployed skill roots match the
source of truth?"

Conflating those checks makes failure output ambiguous and encourages agents to
either ignore real drift or treat missing local install state as a code defect.

## When to Apply

- When adding a new required check to `core`.
- When a check depends on global installs, sibling repos, authenticated CLIs, or
  local deployment state.
- When changing installer profiles or documented startup commands.
- When sync drift is a release blocker but should not block fresh-clone quality.

## Examples

Contributor gate:

```powershell
go run ./cmd/kbcheck core
```

Release/sync gate:

```powershell
go run ./cmd/kbcheck local-release
go run ./cmd/kbcheck skill-sync-report
```

Installer contract:

```text
core = all runtime skills + baseline review/document agents
full = all runtime skills + every reviewer/specialist agent
```

## Related

- `docs/context/research/2026-06-10-skill-bundle-cleanup-audit-refresh.md`
- `docs/context/operations/testing.md`
- `bin/kb-install.mjs`
- `cmd/kbcheck/checks.go`

