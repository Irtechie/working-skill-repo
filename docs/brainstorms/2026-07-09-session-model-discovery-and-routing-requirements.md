---
date: 2026-07-09
topic: session-model-discovery-and-routing
brainstorm_style: kb-brainstorm
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-brainstorm
    status: passed
    required_evidence:
      - docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md
      - Question Gate classification completed
      - Outstanding Questions / Resolve Before Planning is empty
      - No unresolved ask-now or research-first items remain
      - Safe assumptions, deferred planning questions, and parked items are recorded
      - Two document-review passes have no unresolved P0/P1 findings
    proof:
      - docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md
      - docs/context/research/2026-07-09-project-model-routing-surfaces.md
      - Requirements section contains stable behavior IDs and success criteria
      - Outstanding Questions section records Resolve Before Planning as None
      - Assumptions, Deferred to Planning, and Parked sections are populated
      - Document review pass 1 and refinement pass 2 findings were resolved in this artifact
    blockers: []
    passed_at: "2026-07-10T03:39:22Z"
    allowed_next_action: "kb-plan docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md"
---

# Session Model Discovery and Routing

## Problem Frame

KB should route plan slices to suitable subagents without making users classify
their models as Small, Medium, Large, and Planner. Native hosts already know
some available models; the useful missing surface is discovery plus an optional
catalog for local, custom, or cross-provider routes the orchestration surface
cannot know.

Selection happens when `kb-work` is ready to dispatch. Plans remain
provider-neutral. The router should bias upward when capability is uncertain,
escalate without throwing away useful context, show its choices, and let the
user override or disable routing. Work correctness remains determined by proof,
not by which model produced it.

## Research Summary

**Findings that shaped requirements:**

- This Codex App surface exposes Sol, Terra, Luna, and GPT-5.5. Codex CLI
  0.143.0-alpha.9 separately exposes `codex debug models`; its refreshed visible
  catalog here contains GPT-5.5, GPT-5.4, GPT-5.4-Mini, and
  GPT-5.3-Codex-Spark. Codex custom agents may set their own model, provider,
  effort, tools, and instructions. Native discovery must therefore be
  product-surface-specific rather than treating a CLI catalog as all of Codex.
  See the official Codex manual and
  `docs/context/research/2026-07-09-project-model-routing-surfaces.md`.
- GHCP 1.0.70 exposes model selection, per-subagent configuration, and
  OpenAI-compatible providers. Exact catalog enumeration varies by surface and must
  be proven by adapter fixtures rather than assumed. See
  `docs/context/research/2026-07-09-project-model-routing-surfaces.md`.
- LLMCommune's supported app paths are the TinyBoss controller and LiteLLM.
  Fleet MCP currently discovers/runs fleet capabilities; LiteLLM is the current
  OpenAI-compatible LLM inference surface. This shaped the split between
  discovery/control and inference routes. See the fleet connection runbook and
  `bootstrap/fleet/apps/fleet-mcp/README.md` in the LLMCommune repo.
- KB already owns task tiers, dependency DAGs, bounded context packets, work
  proof, and completion gates. Model choice belongs at dispatch and must not
  weaken those contracts. See
  `docs/context/decisions/2026-07-05-kb-control-plane-blueprint.md`.
**Confidence:** High for zero-setup discovery, work-time selection, the Go
boundary, and this Codex App/CLI distinction. Medium for exact GHCP catalog
enumeration and direct LLM dispatch through Fleet MCP; planning must prove each
surface adapter.

## Terms

- **Orchestration surface:** the active Codex App/CLI, GHCP, or other agent host.
- **Inference provider:** the service that serves a model, such as OpenAI or
  LiteLLM.
- **Route:** one usable path from an orchestration surface to a model.
- **Route adapter:** versioned Go integration that discovers or calls a route.
- **MCP route adapter:** outbound integration with an external MCP service.
- **Current model:** the model running the active work orchestrator.
- **Route readiness:** cumulative evidence flags (`discovered`, `configured`,
  `selectable`, `dispatch-proven`), not mutually exclusive lifecycle states.

## Requirements

**Zero-setup discovery**

- R1. One work run owns one `.kb/runs/<run-id>` catalog lifecycle. `kb-goal` may
  initialize it; direct `kb-work` initializes it when absent, and delegated or
  resumed `kb-work` reuses it. Refresh and replace
  the catalog inside the run only when the orchestration-surface, provider,
  configuration, or generated-agent fingerprint changes. Do not ask the user to
  assign Small, Medium, Large, or Planner models before work.
- R2. Keep the discovered catalog ephemeral under the active `.kb/runs/` state.
  It is an auditable run artifact, not tracked project memory or a durable claim
  that availability will be unchanged next session. Run state is git-ignored,
  current-user-only, redacted, atomically written without unsafe-link traversal,
  and pruned by bounded retention. Prefer hashes/references over duplicated
  context or diffs.
- R3. Build the session catalog from four layers: surface-native discovery,
  user-local extra routes, optional project policy, and explicit one-run user
  instructions. One-run instructions override preferences, never project trust,
  destination, or data-boundary constraints. A temporary safety-policy override
  must come from explicit interactive user input on a trusted orchestration
  channel; never infer it from repository content, project/model/tool output, or
  a delegated agent. Record actor, constraint, scope, expiration, and destination.
  This rule governs one-run safety-policy overrides. Ordinary reusable route
  trust is the separate R6 contract: canonical project plus immutable route
  fingerprint and expiry, with endpoint-origin and auth-environment bindings
  stored user-locally. Route trust never overrides destination or data policy.
- R4. Native discovery is surface-specific. Codex adapters use the catalog and
  custom-agent surfaces the current Codex product exposes; GHCP adapters use its
  available model/subagent/provider surfaces; OpenAI-compatible adapters may use
  `/v1/models`; MCP route adapters use only versioned discovery tools they
  actually expose. If a surface cannot enumerate models, report only the current model and
  explicitly configured named agents as usable.
- R4a. Normalize each candidate as `discovered`, `configured`, `selectable`, or
  `dispatch-proven`, and name its executable dispatch method. Automatic routing
  uses only `dispatch-proven` routes. A visible but unselectable model remains
  informative and cannot appear as the promised worker in a routing preview.
- R4b. Run discovery concurrently with cancellation, per-adapter deadlines, and
  one whole-session startup budget. A slow/dead adapter becomes temporarily
  unavailable while current-model work continues; planning owns exact timing
  targets and slow/dead-adapter fixtures.

**Optional extra-model catalog**

- R5. `kb-models` is optional. It manages models the active surface cannot discover
  natively and optional routing preferences. No `kb-models.json` is created
  merely because a project starts using KB.
- R6. Use `~/.kb/models.json` as the user-local catalog for reusable private
  routes such as TinyBoss, LiteLLM, local GPUs, or custom providers. An optional
  project-root `kb-models.json` may reference those aliases or express project
  preferences and trust policy. Machine/account-specific connection details
  stay user-local and are never written to the skills repo or copied into every
  project. Treat tracked project policy as untrusted input: it may narrow routes
  but cannot activate or prefer a private alias unless that alias permits
  project auto-use or the user approves it for this canonical project identity.
  Store that approval only in user-local state.
- R7. An extra route records a stable alias, display/model ID, built-in adapter
  kind, provider or MCP reference, optional endpoint, auth environment-variable
  name, trust/data boundary, declared capability hint (`small`, `medium`,
  `large`, or `planner`), model/provider family, normalized retention,
  training-use and residency claims with provenance, and a concise usage
  description. User-supplied class is `declared`, not measured capability. The
  route contains no credential values or arbitrary
  executable commands. Literal private endpoints default to the user-local
  catalog; tracked project files use aliases or environment-backed references.
- R7a. A route may separate control from inference. For TinyBoss, a controller or
  MCP route adapter may discover, reserve, wake, or start capacity while a
  LiteLLM/OpenAI-compatible route performs inference. Both parts share one
  user-facing alias and are checked independently by `doctor`. Discovery and
  state-changing controller permissions are separate. Reserve/wake/start needs
  user-local action authorization plus project approval, bounded idempotent
  operations, retry/concurrency limits, resumable failure handling, and a
  redacted receipt; `doctor` remains non-mutating. State-changing calls use an
  idempotency key and expiring lease/reservation ID. Cancellation, inference
  failure, or route fallback releases the lease in a recorded cleanup step;
  retries reuse or close the prior lease before another start.
- R8. `kb-models` supports `show`, `add`, `remove`, `prefer`, `ignore-routing`,
  `doctor`, and `calibrate`. Ask for route details only when the user adds an
  unknown external model or when no safe executable route exists. Never ask a
  general model-tier questionnaire. `use <model>` means try it first with normal
  fallback; `require <model>` means exact for that run and pauses if unavailable.
  Preferences and `ignore-routing` are run-scoped unless the user explicitly
  saves them with a global or project scope; their matching clear/reset command
  uses the same scope.

**Planning and conservative selection**

- R9. `kb-plan` records each slice's `tiny`, `small`, `medium`, or `large` task
  tier, rationale, risk, tools, context bounds, and proof. It never freezes a
  model, provider, profile revision, endpoint, or subagent. `tiny` uses the Small
  selection lane while retaining `planned_tier=tiny` in telemetry.
- R9a. During migration, new manifests omit `model_route`; legacy values remain
  readable as advisory hints. `kb-work` and `kbcheck` accept both shapes, require
  no manifest rewrite, and record the actual work-time route only in the run
  receipt. Templates, validators, and fixtures change atomically.
- R10. When routing is enabled, `kb-work` selects the model-backed subagent
  immediately before dispatch from the live session catalog. The current orchestrator chooses among
  eligible models using task fit, tool support, context capacity, trust policy,
  latency/cost preference, prior evidence, and desired provider/family diversity.
  External routes are deny-by-default for sensitive content unless project
  policy authorizes the destination and retention class. Send only the bounded,
  secret-redacted packet and reject routes with missing or insufficient trust
  metadata.
- R10a. Bootstrap automatic eligibility from versioned surface-adapter priors,
  surface/provider capability metadata, or prior KB-owned proof in the matching
  capability envelope. An unknown route remains visible and may be used by
  explicit user direction or attended calibration; it is not selected
  automatically for high-risk/unattended work from branding or a declared class
  alone.
- R11. Treat the plan tier as a minimum capability floor, not a model ceiling.
  Any stronger model may perform simpler work, including Sol or Sonnet handling
  a Small fix. When task or model fit is uncertain, prefer a route with stronger
  matching evidence, regardless of nominal class. High
  risk, broad ambiguity, security/auth, migrations, cross-system behavior, or
  weak context packets default toward the route with the strongest matching
  evidence rather than gambling on an unproven smaller model.
- R12. A user may say `use <model>`, `prefer local`, `prefer hosted`, `use a
  different family for review`, `require <model>`, or `ignore model routing`.
  `use` is a preferred first attempt; `require` is an exact pin. These
  instructions apply to the current run unless the user explicitly saves a
  global or project preference.
- R13. Try the selected model, then other eligible options in the same task
  class, then routes in higher classes only when they have stronger evidence for
  the slice's task family, tools, context, and risk. Class alone is not monotonic
  capability. Never fall downward automatically. Tiny/Small may consider Medium
  and Large; Medium may consider Large; Large tries remaining evidence-qualified
  routes. After exhaustion, use the current model as a degraded fallback only
  when policy permits; otherwise pause that slice with a resumable diagnostic
  and continue unrelated ready work. Never assume a model visible on
  Codex is visible on GHCP or a local provider. A planner model may do worker
  tasks when it is independently eligible or the user asks. Keep a finite
  per-slice attempt ledger and never retry the same route in one dispatch cycle.
  After exhaustion, use the current model when policy permits and mark routing
  degraded; otherwise pause only that slice and continue unrelated ready work.
- R14. Before dispatch, show only non-empty groups with the intended model and
  fallback path, for example `Medium - Terra -> Sonnet -> DeepSeek 4;
  then reselect eligible Large routes: slice-004, slice-007`. Concrete arrows
  name only dispatch-proven routes; show unresolved class movement separately.
  This is a work-time preview, not a plan commitment.
- R15. If a worker fails because of likely model capability, escalate with the
  original context packet, worktree, diff, failing proof, and concise diagnosis.
  Do not restart from zero. Provider, auth, quota, tool, flaky-test, weak-plan,
  and weak-packet failures are not automatically blamed on the model.
  Every fallback is a fresh dispatch decision: rebuild and re-redact diagnostics,
  recheck destination/retention authorization, never carry provider credentials
  or raw prior-provider responses, and require approval before crossing to a
  less-trusted destination.

**Subagent execution and proof**

- R16. Every ready implementation slice is a bounded subagent job. DAG
  readiness, write isolation, concurrency limits, and HITL policy still govern
  whether jobs can run together. Derive least-privilege tools, filesystem roots,
  network destinations, and credentials from the slice; deny undeclared tools
  and require HITL for any privilege or trust-boundary expansion. Model choice
  never broadens authority.
- R17. Use explicit per-call model selection when the surface supports it. Otherwise
  use project-scoped named agents generated for any eligible discovered or
  configured route whose surface supports named-agent model binding. A generic
  spawn API with no model/agent selector proves only that a subagent was spawned;
  it does not prove which model ran. Generate agents with typed serialization,
  bounded identifiers, canonical project containment, KB ownership markers,
  collision refusal for user files, preview, and atomic writes. Generated agents
  are redacted, deterministic, untracked, registered in user-local run state,
  and removed when the run ends or catalog fingerprint changes unless the user
  explicitly preserves them. If a surface cannot load agents from a safe
  untracked location, that dispatch method is unavailable.
- R18. Go owns discovery normalization, selection guards, route adapters, and
  routing receipts. Direct CLI/file operation is the universal baseline;
  outbound MCP route adapters call configured services through the same core.
  Ordinary users install prebuilt artifacts and do not need a Go toolchain. If the matching binary is
  absent or fails to start, only model routing degrades: `kb-work` uses the
  current model, records `router-unavailable`, and preserves every ordinary
  proof gate. Skill-only file-copy installs therefore remain functional with
  routing disabled.
- R19. Record the chosen route, fallback/escalation, adapter, KB `run_id`,
  orchestration/provider `session_id`, and provider-reported model when
  available. Routing evidence is separate from work
  proof. Missing or mismatched routing evidence changes routing status but never
  invalidates otherwise proven work.
- R20. Existing work with no routing receipt is preserved and verified normally.
  Perform a bounded provenance inquiry using available run/session/repository
  evidence and record `explained-external` or `unknown`; never redo correct work
  only to improve routing telemetry.
- R21. `kb-complete` does not select models or rerun proven work for routing
  compliance. It reviews and proves the result, records routing observations for
  future selection, and advances passing work to the shipping gate.

**Capability evidence**

- R22. Each new orchestration-surface session revalidates the catalog fingerprint;
  it reuses the run catalog only while that fingerprint is unchanged. Capability evidence
  may be cached locally with a TTL and must be keyed by orchestration surface,
  inference provider, exact model/adapter revision, task family, tools, context
  bound, and risk. Unknown or stale evidence
  is conservative guidance, not a reason to block user-directed execution.
  Only KB-owned receipts linked to deterministic work proof may establish
  capability success. Catalog/evidence files have strict schemas and size limits,
  trusted writers, permission-preserving atomic writes, and no symlink/path
  escape; repository claims and model self-report are not capability evidence.
  Credit a route only when its receipt proves exact route-bound dispatch with no
  model mismatch; stronger surface/provider evidence may raise confidence. Missing,
  unknown, or mismatched attribution records an observation but cannot credit or
  degrade a model.
- R22a. Automatic routing starts in advisory/pilot mode. Promote it to the
  default only after a representative current-model baseline comparison shows
  no right-to-wrong proof regressions, no increase in repeat-work or user
  interventions, and at least one material benefit such as lower cost/latency or
  useful parallel/offloaded throughput. Report unavailable provider cost/usage
  honestly rather than inventing it.
**Trust and public distribution**

- R25. Only built-in, versioned adapters may execute. Discovery probes are
  bounded, non-mutating, credential-safe, and send no project content. A cloned
  repo cannot introduce executable adapter commands or forward credentials to a
  new destination. Validate normalized endpoints; require TLS off-loopback except
  for explicitly approved private-network HTTP; reject link-local metadata
  targets, DNS rebinding, and cross-origin credential forwarding; bind each auth
  environment variable to its approved adapter and origin.
- R26. Public distribution provides one-command install, deterministic upgrade,
  clear uninstall, Windows/macOS/Linux artifacts, secret redaction, and a
  file-copy fallback. Add only the `kb-models` skill; extend existing `kb-goal`,
  `kb-plan`, `kb-work`, and `kb-complete` surfaces.
- R26a. The initial public surface cohort is Codex, GHCP, and generic
  OpenAI-compatible extra routes, staged rather than promised simultaneously.
  Prove the adapter contract first on this Codex App plus one generic
  OpenAI-compatible route; add GHCP to the supported label after conformance
  fixtures pass. TinyBoss is the composite control-plus-LiteLLM proving route,
  not a machine-specific public default.
- R27. Public artifacts credit Phoenix and other prior art for the specific
  mechanics they informed. Comparisons stay factual, evidence-based, and free of
  personal commentary.

## Success Criteria

- Initial release: Codex plus one generic OpenAI-compatible route reaches work
  without a model-setup question and proves multi-model discovery/dispatch end
  to end; a current-model-only surface degrades cleanly.
- GHCP supported label: GHCP independently passes multi-model discovery,
  named-subagent dispatch, fallback, and receipt conformance fixtures.
- Codex App discovery includes Sol, Terra, Luna, and GPT-5.5 when that surface
  exposes them; Codex CLI discovery independently reflects its own live catalog.
- The first `kb-work` run discovers the current surface catalog, shows conservative
  slice choices and fallbacks, and dispatches model-backed subagents without
  putting model IDs in the plan.
- A user can add DeepSeek, Qwen, Gemma, TinyBoss, LiteLLM, or another custom route
  once and make it available to any project without exposing credentials.
- A Medium route unavailable on the current surface falls through same-class
  options and then Large options; the chosen route is recorded.
- A model visible in a surface catalog but not proven selectable is never promised
  as the dispatched subagent.
- An underestimated route escalates with its context and failing proof instead
  of repeating the slice from scratch.
- `ignore model routing` uses the current/user-selected model while preserving
  all ordinary work-proof requirements.
- Correct pre-existing work completes even when model provenance is unknown.
- A missing/incompatible Go router falls back to the current model and ordinary
  KB proof instead of blocking work.
- A clean user can install without a Go toolchain, start work without answering
  a questionnaire, understand one compact routing preview, and continue safely
  when discovery is unavailable.
- A representative baseline comparison proves routing does not increase
  first-pass proof failures, repeat-work, or user interventions before routing
  becomes the default, and records at least one material efficiency/throughput
  benefit.

## Scope Boundaries

- No mandatory per-project model questionnaire or role assignment.
- No hardcoded claim that every Codex, GHCP, or local surface exposes the same
  model names.
- No provider purchase, authentication, entitlement change, or secret creation.
- No weakening of tests, review, browser proof, or acceptance for smaller
  models.
- No automatic global skill promotion or deletion.
- Project-local skill anti-sprawl remains a separate follow-up initiative and is
  not a release dependency for session model routing.
- No requirement to install Phoenix's lifecycle skills. Focused MCP
  interoperability remains eligible.

## Key Decisions

- Native discovery first; `kb-models` only adds what the surface cannot know.
  Evidence: live Codex catalog inspection and surface-specific model catalogs.
- No setup questionnaire. The work orchestrator chooses and tells the user what
  it will use. Evidence: user requirement; reversible through per-run overrides.
- User-local global config owns machine-specific connection details; optional
  project policy references aliases and preferences. Evidence: user decision
  and portability/security boundaries.
- Plans freeze task complexity, not models. Evidence: model availability differs
  by surface, provider, and session.
- Bias upward under uncertainty. Evidence: the cost of one stronger dispatch is
  often lower than repeating failed work, but class is only a hint. Select by
  matching evidence, favor first-pass success, and expose the tradeoff through
  preview and `prefer local`/saved cost preferences.
- User-local routes and endpoints are separate from optional tracked project
  policy. Evidence: portability and credential/network-boundary concerns.
- Go is the single deterministic core; outbound MCP routing is an adapter, not
  another workflow.
  Evidence: existing `cmd/kbcheck` ownership and user decision.

## Dependencies / Assumptions

- [safe-assumption] `~/.kb/models.json` is the user-local extra-route catalog.
  Reversible because discovery and project references use stable aliases; a
  later path migration can preserve the schema. Proof: Windows/macOS/Linux path
  and migration fixtures.
- [safe-assumption] Native discovery runs once per work session and refreshes on
  a surface/provider/config fingerprint change. Reversible because manual `doctor` refresh
  remains available. Proof: catalog-change fixtures.
- [safe-assumption] Capability evidence has a TTL and never replaces live
  availability. Reversible because stale evidence is ignored. Proof: model,
  adapter, surface, provider, and expiry mismatch fixtures.

## Alternatives Considered

- Ask every project to define Planner/Small/Medium/Large: rejected because native
  hosts already expose usable models and the questionnaire adds ceremony.
- Hardcode Sol/Terra/Luna/GPT/Opus classes globally: rejected because catalogs and
  access differ by orchestration surface, account, and inference provider.
- Freeze model routes in plans: rejected because `kb-work` owns live dispatch and
  plans should remain portable.
- Store private endpoints in tracked project policy by default: rejected because
  local network routes are user/machine-specific.
- Use only the current model: retained as `ignore model routing`, not the default
  when safe subagent choices are discoverable.

## Slice Candidates (advisory for /kb-plan)

- Session catalog - goal/work discover native and configured subagent routes
  without asking the user to classify models.
- Extra-route catalog - users register local/custom/MCP/OpenAI-compatible models
  once and optionally constrain them per project.
- Conservative selector - work maps slice requirements to live models, previews
  choices, and escalates upward without losing context.
- Host dispatch adapters - Codex, GHCP, and local/TinyBoss routes spawn the
  selected subagent through proven orchestration surfaces.
- Routing evidence - receipts and telemetry explain actual choices without
  overriding work correctness.
- Distribution proof - Go packaging, fixtures, docs, sync, and install gates
  prove the feature across supported hosts.

## Outstanding Questions

### Resolve Before Planning

None.

### Deferred to Planning

- [defer-to-planning][Affects R4/R17][Needs research] Prove the exact GHCP model
  catalog and named-subagent discovery commands for the supported CLI version.
- [defer-to-planning][Affects R7/R25][Technical] Finalize the extra-route schema,
  endpoint/reference precedence, and secret-redaction fixtures.
- [defer-to-planning][Affects R10/R22][Technical] Define conservative capability
  priors, TTL, risk uplift, same-class choice, and escalation fixtures.
- [defer-to-planning][Affects R22a/R26a][Technical] Define the representative
  baseline corpus, observable cost/latency fields, material-benefit threshold,
  and adapter-conformance promotion gate.
- [defer-to-planning][Affects R17/R19][Technical] Define model selection and
  receipt evidence for Codex surfaces whose generic spawn call lacks a model
  parameter.
- [defer-to-planning][Affects R4/R18][Needs research] Decide whether TinyBoss
  local LLM discovery/dispatch uses LiteLLM plus Fleet MCP control or a future
  versioned MCP LLM capability. Current Fleet MCP is capability/job oriented.
- [defer-to-planning][Affects R18/R26][Technical] Define prebuilt Go packaging,
  signing, upgrade, uninstall, and file-copy fallback.

### Parked / Out of Scope

- [parked][Affects R27] Installing Phoenix's full lifecycle vocabulary. A focused
  interoperability layer remains eligible.
- [parked] Automatic global skill promotion or deletion belongs to the separate
  project-local skill-governance follow-up.
- [parked] Project-local skill inventory, consolidation, and audiobook-specific
  ownership. Preserve the requirement for a separate follow-up brainstorm.
- [parked][Affects R19] Claims that route-bound evidence proves hidden provider
  weights or serving internals.
- [parked][Affects R18/R26] An inbound KB MCP server facade. Revisit only after
  the direct Go path has a concrete external MCP client; require local-only
  defaults, explicit remote enablement, authentication, per-tool authorization,
  concurrency limits, and redacted audit logs.

## Next Steps

-> /kb-plan docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md
