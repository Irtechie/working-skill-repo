---
name: kb-configure
description: "Configure portable per-project delivery policy and an explicit experimental AMR benchmark opt-in."
argument-hint: "[show|amr-experiment-on|amr-experiment-off|delivery-local|delivery-pr|delivery-direct|reset]"
---

# KB Configure

Configure optional project execution policy without making ordinary KB startup
interactive.

Orchestrator-directed DDR is the normal execution path and needs no project
setup. This skill owns delivery policy and an explicit AMR benchmark opt-in.
The AMR flag is for experimental harnesses only; normal `kb-work` never reads it
to choose an execution owner or route. Personal source preference belongs to
user-local `kb-models` state keyed by project identity.

## Config Path

`docs/context/operations/kb-routing.yaml`

The file contains portable project policy only. Never write model endpoints,
auth environment-variable names, trust approvals, commands, or credentials
there. `kbrouter` continues to own host and user-local model configuration.

## Behavior

1. If the file exists, read it and show a compact experiment/delivery summary. Ask only
   which setting the user wants to change when their request is ambiguous.
2. If the file is absent and an argument was supplied:
   - `amr-experiment-on` permits a separate controlled AMR benchmark.
   - `amr-experiment-off` disables that benchmark.
   - Legacy `attempts-on` and `attempts-off` are readable aliases for those
     experiment flags only and must not alter normal `kb-work`.
   - `delivery-local` keeps reviewed work local.
   - `delivery-pr` commits, pushes a topic/fork branch, and opens/updates a PR.
   - `delivery-direct` permits verified direct-default integration; protection
     or policy rejection falls back to PR or blocks.
   - `show` reports AMR experiment disabled and local delivery without creating
     a file.
   - `reset` removes only this project policy after explicit confirmation.
3. If the file is absent and no mode was supplied, show the defaults and the
   exact commands above. Do not start a setup questionnaire.
4. Do not ask model-by-model questions. Normal DDR reads the active host schema
   and user-local catalog at work time; `kb-models` configures optional
   user-local extras only.
5. Preserve unrelated project policy when updating an existing file.

## Canonical Schema

```yaml
schema_version: 1

experimental_amr:
  enabled: false
  affects_normal_work: false

delivery:
  mode: pr
  merge: manual
  post_merge_sync: false
```

These safety rules are fixed rather than configurable:

- `model_tier` is the minimum execution capability, not the validator.
- Normal work uses one explicit `current` or `delegated` owner decision.
- Delegated work selects exactly one qualified same-tier-or-higher route.
- Normal work never passes `attempt_tier`, routes downward automatically, or
  silently falls back across owners.
- AMR experiments do not affect normal manifests or work execution.
- Ordinary proof remains authoritative. Routing receipts are telemetry.
- Repository ownership/write access never selects direct delivery by itself.
- Direct delivery, automatic merge, and post-merge sync require explicit policy.

## Defaults

When no config exists:

- AMR experiment: disabled.
- Normal DDR: orchestrator-owned and unaffected by the AMR experiment flag.
- Delivery: local.
- `kb-start`, `kb-plan`, and `kb-work` do not ask configuration questions.

## Output

After writing, report the path and a one-line summary:

```text
KB configured: AMR experiment disabled; normal DDR unchanged; delivery PR/manual.
```
