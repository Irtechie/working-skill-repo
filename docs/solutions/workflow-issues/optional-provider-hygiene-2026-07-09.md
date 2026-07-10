---
title: Keep Optional Provider State Out of Repo-Local Gates
date: 2026-07-09
category: workflow-issues
module: kbcheck
problem_type: workflow_issue
component: tooling
severity: high
applies_when:
  - Adding MCP servers, hooks, context engines, or provider-specific adapters
  - Auditing a machine without making its global state part of contributor proof
tags: [provider-hygiene, mcp, cce, phoenix, repo-local-gates, portability]
---

# Keep Optional Provider State Out of Repo-Local Gates

## Context

Phoenix had been removed from the repo, but a stale user-level Copilot MCP entry
and globally installed Phoenix skills could still activate it. CCE is an owned
context engine that should remain available as an opt-in adapter. The first
hygiene check caught provider configuration but accidentally made `core`
dependent on user-global state and treated any Phoenix text as activation.

## Guidance

- Keep `core` repo-local. It may inspect provider files inside the checkout, but
  not user-global config by default.
- Make machine inspection explicit:

  ```powershell
  go run ./cmd/kbcheck provider-hygiene --include-user
  ```

- Parse provider configuration semantically:
  - decode JSON before checking server names or commands;
  - inspect only MCP/hook/provider sections;
  - ignore entries with `enabled: false` or `enabled = false`;
  - ignore comments and unrelated attribution text;
  - fail on unreadable or malformed files instead of treating them as clean.
- Treat CCE entries as optional configuration, not an error.
- Reject active Phoenix provider entries without deleting Phoenix research or
  attribution.
- Keep context packets immutable execution input. Store measured runtime usage
  in a separate telemetry artifact linked by `packet_id`.

## Why This Matters

Repo-local gates must return the same result on a clean clone and on a
maintainer workstation. Global machine state is useful diagnostic input, but it
is not repository correctness. Semantic parsing also avoids both escaped-name
bypasses and false failures from disabled providers.

## When to Apply

- Adding optional MCP/context providers.
- Writing contributor versus release/environment checks.
- Removing a runtime while preserving its research history.
- Adding worker context or usage telemetry contracts.

## Examples

```powershell
# Contributor-safe and repo-local
go run ./cmd/kbcheck core

# Explicit machine/provider audit
go run ./cmd/kbcheck provider-hygiene --include-user
```

```json
{
  "mcpServers": {
    "phoenix": {
      "enabled": false,
      "command": "phoenix-mcp"
    },
    "context-engine": {
      "command": "cce",
      "args": ["serve"]
    }
  }
}
```

The disabled Phoenix entry is ignored; CCE is reported as optional. An enabled
Phoenix entry fails.

## Related

- [Contributor Core vs Release Sync Gates](contributor-core-vs-release-sync-gates-2026-06-10.md)
- `docs/context/research/2026-07-09-cross-runtime-token-efficiency.md`
- `docs/context/kb/context-packet-schema.md`
