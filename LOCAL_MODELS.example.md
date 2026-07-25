# Local Model Setup

Use this tracked guide to configure models that exist on one person's machine
or private network. Do not fill personal values into this repository.

The configured state belongs to the operating-system user:

| State | User-local path | Commit it? |
|---|---|---|
| Model routes | `~/.kb/models.json` | No |
| Endpoint approvals | `~/.kb/trust.json` | No |
| Per-project source priority | `~/.kb/project-priorities.json` | No |

Use `kbrouter models` to manage these files. Do not hand-edit them. This lets
different users and CLIs expose different models without repository changes.

## What To Collect

Configure one route at a time. Ask the user only for missing values:

```yaml
alias: "<stable local name, such as local.coder>"
model_id: "<model name exposed by the endpoint>"
endpoint: "<OpenAI-compatible or LiteLLM /v1 URL>"
auth_env: "<environment-variable name, or blank when no token is required>"
hosting: "self-hosted | provider-hosted | unknown"
retention: "none | session | limited | unknown"
training_use: "no | yes | unknown"
trust_provenance: "<how the user knows and controls this route>"
declared_class: "small | medium | large | unknown"
project_root: "<project that may use the route>"
source_priority: "automatic | self-hosted-first | native-first"
```

`declared_class` is a starting claim, not capability proof. Delegated routing
uses a route only when the router has dispatch-qualified evidence covering the
required tier, task family, tools, context size, and risk.

Generic MCP model dispatch and raw GGUF/model-file execution are not implemented
by `kbrouter`. A file path or MCP address is not enough to configure a route.
Expose the model through an OpenAI-compatible or LiteLLM `/v1` endpoint and
configure the endpoint's model ID.

## Register A Route

No-auth local route:

```powershell
kbrouter models add --scope user `
  --alias <alias> `
  --model <model-id> `
  --endpoint <endpoint> `
  --hosting <self-hosted|provider-hosted|unknown> `
  --retention <none|session|limited|unknown> `
  --training-use <no|yes|unknown> `
  --trust-provenance "<provenance>" `
  --class <small|medium|large|unknown>
```

Authenticated route:

```powershell
kbrouter models add --scope user `
  --alias <alias> `
  --model <model-id> `
  --endpoint <endpoint> `
  --auth-env <ENVIRONMENT_VARIABLE_NAME> `
  --hosting <self-hosted|provider-hosted|unknown> `
  --retention <none|session|limited|unknown> `
  --training-use <no|yes|unknown> `
  --trust-provenance "<provenance>" `
  --class <small|medium|large|unknown>
```

Store only the environment-variable name in route configuration. Never put a
token value in this file, a command, a plan, a handoff, or Git.

## Approve And Check

Private endpoints require an attended approval tied to the canonical project
path. The user must run and confirm this command in their own interactive
console:

```powershell
kbrouter models approve --alias <alias> --project-root <project-root>
```

Inspect configuration without contacting the endpoint:

```powershell
kbrouter models doctor --project-root <project-root>
```

Run a bounded endpoint/model presence check only when explicitly requested:

```powershell
kbrouter models doctor --project-root <project-root> --probe
```

Optionally prefer eligible self-hosted routes for this project:

```powershell
kbrouter models priority --project-root <project-root> --mode self-hosted-first
```

Use `automatic` when the user has no strong source preference.

## Orchestrator-Directed DDR

Plans remain portable. Every runnable slice records:

```yaml
model_tier: small | medium | large
model_tier_reason: "<why this minimum execution capability is required>"
```

The plan never stores a person's model ID, endpoint, adapter, auth environment
variable, or source preference.

Difficulty-Driven Routing (DDR) is the normal path:

1. `kb-plan` records the minimum execution capability and its reason.
2. Immediately before work, the current orchestrator decides whether its own
   reasoning, context, tools, trust, or authority require `current` execution.
3. Otherwise it chooses `delegated` and selects exactly one qualified
   same-tier-or-higher route. Native App targets execute through the App's exact
   callable-agent tool; CLI and user-local targets execute through `kbrouter`.
4. Neither owner silently falls back to the other. A failed or unavailable route
   requires an explicit new decision, re-plan, or block.
5. Deterministic proof remains authoritative regardless of which route ran.

The orchestrator must not assume model names from memory. For example, a Codex
App agent schema may expose Sol and Terra while a CLI catalog exposes
`gpt-5.4-mini`; those are separate callable surfaces unless an adapter proves
otherwise.

Adaptive Model Routing (AMR) is not promoted. Normal work does not configure or
pass `attempt_tier`, does not require a lower-tier trial, and does not use AMR
results as production routing policy.

Run-scoped controls remain available:

- `use <alias>` selects that eligible route for delegated execution.
- `require <alias>` hard-pins that delegated route and pauses when unavailable.
- `ignore model routing` explicitly chooses current execution, subject to
  current-route capability validation.

Trust, destination, retention, tools, context, and proof boundaries still apply
to every override.
