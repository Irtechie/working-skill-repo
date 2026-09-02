# Local Model Setup

Use this tracked guide to configure models that exist on one person's machine
or private network. Do not fill personal values into this repository.

The configured state belongs to the operating-system user:

| State | User-local path | Commit it? |
|---|---|---|
| Model routes | `~/.kb/models.json` | No |
| Endpoint approvals | `~/.kb/trust.json` | No |
| Per-project source priority | `~/.kb/project-priorities.json` | No |

Use `kbrouter models` to manage these files. Do not hand-edit them. Runtime reads
canonical `~/.kb` state, never the repository import file.

Temporarily bypass every configured user-local route without deleting endpoints,
aliases, or approvals:

```powershell
kbrouter models local-routing --enabled false
kbrouter models local-routing --enabled true
```

The command atomically changes top-level `enabled` in `~/.kb/models.json`.
Missing `enabled` remains backward-compatible and means `true`.

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

## Import A Route

Copy the checked-in placeholder to the ignored operator file, fill it, and
import it:

```powershell
Copy-Item config\kbrouter-routes.example.json kbrouter-routes.local.json
kbrouter models import --file kbrouter-routes.local.json
```

The import is strict, bounded, non-symlinked, and atomic. It rejects unknown
fields, placeholders, endpoint credentials, and auth values. `auth_env` stores
only an environment-variable name. Import never grants or renews trust.

## Register One Route Directly

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
Authenticated HTTP is limited to endpoints whose resolved addresses are all
loopback. Use HTTPS for authenticated private-LAN routes. Unauthenticated
private-LAN HTTP remains supported when the route uses the private boundary.

## Choose Approval Mode And Check

Bounded local routing defaults to no approval prompt. Persist that preference
explicitly with:

```powershell
kbrouter models approval-mode --mode disabled
```

Users who want a project-bound endpoint/auth gate can opt in:

```powershell
kbrouter models approval-mode --mode required
kbrouter models approve --alias <alias> --project-root <project-root>
```

Only `models approve` is attended. Disabling approval does not disable explicit
route denials, endpoint safety checks, data policy, bounded-attempt receipts, or
deterministic proof.

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
4. An eligible local route receives exactly one attempt for a canonical
   project/run/slice. Endpoint/model unavailability, bounded probe failure,
   timeout, 5xx, dispatch failure, or deterministic proof failure emits a
   `parent-return` receipt (exit `10`) and returns immediately to the active
   parent.
5. A returned local result is `awaiting-proof`. The parent runs the named
   deterministic proof and records `pass|fail` with `kbrouter ddr resolve`.
6. The parent continues with its active model or host-native selection logic.
   It does not wait for setup, ask for model approval, select a second local
   route, or use a hardcoded provider fallback.
7. `require <alias>` remains the explicit hard pin and blocks that slice when
   the route cannot complete.
8. Deterministic proof remains authoritative regardless of which route ran.

Write the bounded request under ignored project state (for example,
`.kb\ddr\request.json`) using `config\kbrouter-ddr-request.example.json` and its
schema. Then invoke:

```powershell
kbrouter ddr attempt `
  --project-root <project-root> `
  --run-id <run-id> `
  --slice-id <slice-id> `
  --alias <alias> `
  --request <project-root>\.kb\ddr\request.json `
  --tier <small|medium|large> `
  --tier-reason "<reason>" `
  --task-family <family> `
  --tool <tool> `
  --context-size <tokens> `
  --risk normal `
  --sensitive-data=<true|false> `
  --json
```

On `awaiting-proof`, consume the response exactly once, run the named
deterministic proof, then resolve:

```powershell
kbrouter ddr resolve `
  --project-root <project-root> `
  --run-id <run-id> `
  --slice-id <slice-id> `
  --proof-result <pass|fail> `
  --proof-command "<exact command>" `
  --proof-artifact-hash sha256:<64-hex-digest> `
  --json
```

Exit `0` means a response awaits proof or proof completed, `10` means structured
parent return, `3` means a required alias blocked, and `2` means invalid usage.

The orchestrator must not assume model names from memory. Host-native and CLI
catalogs are separate callable surfaces unless an adapter proves otherwise.

Normal DDR does not require a lower-tier trial or use unverified benchmark
results as production routing policy.

Run-scoped controls remain available:

- `use <alias>` selects that eligible route for delegated execution.
- `require <alias>` hard-pins that delegated route and pauses when unavailable.
- `ignore model routing` explicitly chooses current execution, subject to
  current-route capability validation.

Trust, destination, retention, tools, context, and proof boundaries still apply
to every override.
