# Landmines

Last reviewed: 2026-05-31

This file records active repo-specific traps the model is likely to miss without
explicit warning. Do not use it for generic engineering advice.

## Active Landmines

None.

## Entry Schema

```yaml
- id: YYYY-MM-DD-short-name
  status: active|resolved|stale-review
  severity: low|medium|high|critical
  owner_surface: "<skill, script, doc, generator, fixture, or workflow that owns the durable fix>"
  trigger: "<specific situation where the trap applies>"
  failure_mode: "<what the model will likely get wrong>"
  evidence:
    - "<file, command, review finding, failing test, or observed mistake>"
  fix_condition: "<what change retires the landmine>"
  verification: "<command, eval, test, or review check proving the fix>"
  created: YYYY-MM-DD
  last_seen: YYYY-MM-DD
  archive_reason: ""
```

## Lifecycle

- Add a landmine only with concrete evidence and an owner surface.
- Reject entries that only say "write better code" or repeat generic workflow
  advice.
- Keep active entries concise enough for `kb-map` startup context.
- When the owner surface is fixed and verification passes, move the entry to
  `Resolved Landmines` immediately.
- Use stale review only for unfixed entries that have not been seen recently;
  do not silently delete unresolved high-severity traps.

## Resolved Landmines

None.
