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
      - "2026-07-11 document-review pass 1: coherence, feasibility, product, security, scope, and adversarial findings resolved"
      - "2026-07-11 document-review pass 2/final: feasibility and adversarial clear; final coherence drift auto-fixed"
      - "Resolve Before Planning remains empty; safe assumptions and later cohorts are explicit"
    blockers: []
    passed_at: "2026-07-11T04:46:57Z"
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
route-neutral: they record task difficulty, constraints, and proof, never a
hosted model version, extra-route alias, or transport choice.
The router should bias upward when capability is uncertain,
escalate without throwing away useful context, show its choices, and let the
user override or disable routing. Work correctness remains determined by proof,
not by which model produced it.

## Research Summary

**Findings that shaped requirements:**

- Codex App and Codex CLI expose separate, changing catalogs; custom agents may
  set their own model, provider, effort, tools, and instructions. Native
  discovery must therefore be product-surface-specific rather than treating a
  captured model list from one surface or date as all of Codex.
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
- **Host-native route:** a route whose immutable management origin is the active
  host ecosystem and requires no `kb-models` configuration. Live discovery is a
  separate fact: rediscovering a KB-managed profile does not make it native.
- **Extra route:** an optional user-configured route with immutable `extra`
  management origin, identified by a stable user-local alias even when the host
  later discovers or invokes its generated profile.
- **Tracked project policy:** shareable repository-controlled narrowing
  constraints. It never grants trust, destination approval, or personal route
  priority.
- **User-local project state:** personal preference and approval keyed by the
  canonical project identity, stored outside the repository.
- **Effective routing policy:** the safe merge of tracked narrowing constraints,
  user-local preference/approval, and run-scoped instructions. Only this merged
  policy may authorize dispatch.
- **Adaptive Model Routing (AMR):** work-time route selection, including at most
  one eligible next-lower substantive attempt, deterministic proof, and
  a planned-tier-or-higher surgical correction handoff. Automatic correction
  execution remains fail-closed until an isolated workspace, host-owned proof,
  and compare-and-swap apply runner exist.
- **Planned tier:** the correction/authority tier recorded by `kb-plan`; it is
  not a proof level or a permanent first-worker assignment.
- **Attempt tier:** an optional, work-time-only next-lower tier explicitly
  requested by the driver after bounded eligibility is established.
- **Route readiness:** cumulative evidence flags (`discovered`, `configured`,
  `selectable`, `dispatch-qualified`, `dispatch-proven`), not mutually exclusive
  lifecycle states.

## Requirements

Unless a requirement says otherwise, the first exact routed-dispatch cohort is
Codex CLI plus one OpenAI-compatible/LiteLLM extra route through a trusted Codex
profile and harness. Host-managed native choice without exact attribution is the
zero-setup baseline. GHCP, exact Codex App dispatch, generic direct inference,
MCP model dispatch, generated cross-host agents, and fleet controller actions
are later cohorts or parked work and cannot enter the initial manifest merely
because their end-state constraints appear below.

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
- R3. Build the session catalog from four layers: surface-native discovery;
  user-local extra routes plus canonical-project preference/approval; tracked
  project narrowing policy; and explicit run-scoped user instructions. One-run
  instructions override preferences, never project trust,
  destination, or data-boundary constraints. A temporary safety-policy override
  must come from explicit interactive user input on a trusted orchestration
  channel; never infer it from repository content, project/model/tool output, or
  a delegated agent. Record actor, constraint, scope, expiration, and destination.
  This rule governs one-run safety-policy overrides. Ordinary reusable route
  trust is the separate R6 contract: canonical project plus immutable route
  fingerprint and expiry, with endpoint-origin and auth-environment bindings
  stored user-locally. Route trust never overrides destination or data policy.
  A host-native route may inherit the active surface's existing account and
  destination authorization only when a versioned adapter proves the route
  remains inside the same origin, retention class, and data boundary. Custom
  providers, changed origins, or weaker retention require user-local canonical-
  project approval; native origin alone grants no trust.
- R4. Native discovery is surface-specific. Codex adapters use the catalog and
  custom-agent surfaces the current Codex product exposes; later-cohort GHCP adapters use its
  available model/subagent/provider surfaces; OpenAI-compatible adapters may use
  `/v1/models`; MCP route adapters use only versioned discovery tools they
  actually expose. If a surface can launch an unpinned host-chosen subagent but
  cannot enumerate or attribute the selected model, expose one synthetic
  `host-auto` route with the real executable dispatch method and explicitly
  unknown model attribution. Do not invent concrete models. Otherwise report
  only the current model and explicitly configured named agents as usable.
- R4a. Normalize each candidate with cumulative `discovered`, `configured`,
  `selectable`, `dispatch-qualified`, and `dispatch-proven` evidence flags, and
  name its executable dispatch method. A versioned adapter prior may establish
  `dispatch-qualified` for advisory pilot selection. Reserve
  `dispatch-proven` and capability credit for an exact route-bound receipt
  linked to deterministic work proof. A visible but unselectable model remains
  informative and cannot appear as the promised worker in a routing preview.
- R4b. Run discovery concurrently with cancellation, per-adapter deadlines, and
  one whole-session startup budget. A slow/dead adapter becomes temporarily
  unavailable while current-model work continues; planning owns exact timing
  targets and slow/dead-adapter fixtures.

**Optional extra-model catalog**

- R5. `kb-models` is optional. It manages routes with immutable `extra`
  management origin, including profiles the active host later discovers or
  invokes, plus explicit routing preferences. No model or
  project policy file is created merely because a project starts using KB.
- R6. Use `~/.kb/models.json` as the user-local catalog for reusable private,
  self-hosted, local-GPU, cross-provider, or custom routes. Store a user's saved
  project preference in user-local state keyed by canonical project identity.
  A tracked project `kb-models.json` is optional and explicit; it may express
  shareable constraints or alias references, never a person's endpoint,
  credential binding, or default priority. Machine/account-specific connection
  details stay user-local and are never written to the skills repo or copied
  into every project. Treat tracked project policy as untrusted input: it may
  narrow routes but cannot activate or prefer a private alias without user-local
  approval for this canonical project identity. `auto-use` or source priority
  changes ordering only; it never grants route trust, destination approval,
  credential access, or action authority. Store approval only in user-local state.
- R7. An extra route records a stable alias, display/model ID, built-in adapter
  kind, provider or MCP reference, optional endpoint, auth environment-variable
  name, trust/data boundary, declared capability hint (`small`, `medium`,
  `large`, or `planner`), immutable management origin (`native` or `extra`),
  one or more live discovery sources, hosting class
  (`self-hosted`, `provider-hosted`, or `unknown`), model/provider family,
  normalized retention,
  training-use and residency claims with provenance, and a concise usage
  description. User-supplied capability and hosting classes are `declared`, not
  measured evidence. A private trust boundary alone must not be relabeled
  `self-hosted`; native adapters establish origin, and extra-route setup records
  hosting class explicitly. The route contains no credential values or arbitrary
  executable commands. Literal private endpoints default to the user-local
  catalog; tracked project files use stable aliases and non-secret narrowing
  constraints only. Endpoint, provider, profile, and auth-environment references
  remain user-local.
  `self-hosted` means inference compute controlled by the user or their
  organization, whether on the same machine or an explicitly declared
  LAN/private server. Managed or unknown hosting does not match
  `self-hosted-first` merely because its endpoint is private.
  A KB-managed profile rediscovered by the host merges by immutable route
  fingerprint into the original extra-route record; rediscovery never converts
  its management origin to native or splits capability evidence.
- R8. `kb-models` supports `show`, `add`, `remove`, `approve`, `revoke`,
  `prefer`, `ignore-routing`, `doctor`, and `calibrate`. Ask for connection details only when the user
  explicitly adds an extra route. A missing route or router degrades to the safe
  current-model path unless the user explicitly declines that fallback. Never
  ask a model-tier questionnaire. If configured extra routes exist and this
  canonical project has no saved preference, ordinary work silently uses
  `automatic` source choice. Offer `automatic`, `self-hosted-first`, or
  `native-first` only during explicit project setup or an explicit
  `kb-models prefer/configure` action; never pause ordinary `kb-work` for this
  choice. Ordinary `kb-map` lookup, silent bootstrap, and planning ask nothing.
  Persist an explicit choice user-locally so it is not repeated. `use <model>`
  means try it first with normal fallback; `require <model>` means exact for
  that run and pauses if unavailable. Preferences and `ignore-routing` are
  run-scoped unless the user explicitly saves them with a `user-global` or
  `user-local canonical-project` scope; project scope describes the matching
  key, not tracked repository storage. Clear/reset uses the same scope.
  User-local project preferences use a versioned collection keyed by canonical
  project identity, retain an explicit saved `automatic` marker, and clear/reset
  only the matching key unless the user requests a user-global reset.
  `native-first` ranks immutable native management origin, not every route the
  host happens to rediscover through a KB-managed profile.
- R8a. Quick-add for the initial OpenAI-compatible/LiteLLM adapter asks only for
  alias, endpoint, model ID, and optional auth environment-variable name.
  Hosting class is one optional explicit choice and defaults to `unknown`.
  Capability, retention, training-use, residency, and family metadata use
  conservative `unknown`/unqualified defaults and move to advanced flags or
  attended calibration. Quick-add never grants project trust or automatic
  eligibility; approval and evidence remain separate steps.
- R8b. The initial cohort must prove `show`, quick `add`, attended approval,
  non-mutating `doctor`, `select`, run-scoped `use`/`require`/`ignore`, explicit
  project source priority, matching clear/reset, and `remove`. User-global saved
  priority, generic calibration, composite controller setup, and additional
  adapter conveniences are later-cohort commands unless required by a protected
  conformance fixture.
- R8c. Attended `approve` binds the immutable route fingerprint, canonical
  project identity, destination origin, retention class, auth binding, and
  expiry after showing a redacted preview. `revoke` invalidates that approval.
  Neither `add`, `doctor`, repository content, nor route self-report may perform
  this transition.
- R8d. Quick-add creates a configured but non-executable record. Attended
  approval resolves dispatch-blocking destination and retention declarations.
  A versioned profile-binding step then either binds an existing trusted
  user-local Codex profile or creates a KB-owned profile with deterministic
  naming, an ownership marker, collision refusal, typed serialization, atomic
  writes, secret references only, and matching cleanup. Exact executable,
  profile-revision, model, destination, and harness checks establish
  `dispatch-qualified`; `doctor` reports the next missing transition without
  silently performing it.

**Planning and conservative selection**

- R9. `kb-plan` records each slice's `small`, `medium`, or `large` planned
  correction/authority tier, rationale, risk, tools, context bounds, and proof.
  It never writes a concrete model, route alias, source preference, provider,
  profile revision, endpoint, transport, subagent, or `attempt_tier`. Legacy
  `tiny` uses the Small selection lane while retaining `planned_tier=tiny` in
  telemetry.
- R9a. During migration, new manifests omit `model_route`; legacy values remain
  readable as advisory hints. `kb-work` and `kbcheck` accept both shapes, require
  no manifest rewrite, and record the actual work-time route only in the run
  receipt. Templates, validators, and fixtures change atomically.
- R9b. DDR is retired as a separate workflow, configuration, manifest, and proof
  concept. New artifacts omit DDR fields. Legacy DDR metadata remains readable
  only as inert telemetry and never creates a DDR-specific proof gate.
- R10. When routing is enabled, `kb-work` selects the model-backed subagent
  immediately before dispatch from the live session catalog. The current orchestrator chooses among
  eligible models using task fit, tool support, context capacity, trust policy,
  latency/cost preference, prior evidence, and desired provider/family diversity.
  Any route crossing an unapproved destination or data boundary is
  deny-by-default for sensitive content. Host-native versus extra describes
  immutable management origin, not live discovery source or trust. Only the effective routing policy, including any
  required user-local approval, may authorize the destination and retention
  class. Send only the bounded, secret-redacted packet and reject routes with
  missing or insufficient trust metadata.
- R10a. Bootstrap advisory automatic eligibility from versioned surface-adapter priors,
  surface/provider capability metadata, or prior KB-owned proof in the matching
  capability envelope. Adapter priors establish `dispatch-qualified`, not
  `dispatch-proven` or capability success. An unknown route remains visible and may be used by
  explicit user direction or attended calibration; it is not selected
  automatically for high-risk/unattended work from branding or a declared class
  alone.
- R10b. The current driver, not the selector, decides whether one next-lower
  attempt is eligible. Eligibility requires settled intent, bounded files,
  interfaces and authority, objective proof that can reject a bad result,
  acceptable destination/trust risk, and exact escalation triggers. “This is
  code,” a file extension, model branding, or lower price is insufficient.
  Missing proof, sensitive/high-risk ambiguity, authority expansion, or the
  project opt-out starts at the planned tier.
- R11. Treat the plan tier as correction/authority required if an initial
  attempt fails, not as the validator or a mandatory first-worker floor. Any
  stronger model may perform simpler work. AMR may explicitly try only the next
  lower tier under R10b, then deterministic proof either preserves the result
  or triggers planned-tier correction. High risk, broad ambiguity,
  security/auth, migrations, cross-system behavior, subjective philosophy or
  product meaning, weak packets, and weak proof begin at the planned tier or
  require HITL.
- R12. A user may say `use <model>`, `prefer self-hosted` (including the natural
  shorthand `prefer local`), `prefer native`, `use a different family for
  review`, `require <model>`, or `ignore model routing`.
  `use` is a preferred first route inside R10/R10b constraints; it does not
  create lower-tier eligibility. `require` is an exact pin that bypasses the
  automatic attempt and pauses if unavailable. `ignore model routing` also
  bypasses the attempt. These instructions apply to the current run unless the
  user explicitly saves a user-global or user-local canonical-project-scoped
  preference. `use` overrides the
  automatically selected first route for the run; only `require` hard-pins.
  `require` bypasses only automatic route and attempt selection; it never
  bypasses R3, R10, R15, R16, destination/retention authorization, credential
  binding, or tool and filesystem authority.
- R13. If R10b passes, an explicit pilot may try one requested next-lower
  attempt tier using an eligible dispatch-qualified route. Public-default
  next-lower attempts require the exact model/route revision cohort that passed
  R22a/R22b or matching exact route-bound capability evidence; adapter priors
  alone remain planned-tier/advisory. Otherwise begin at the planned tier.
  Before a substantive work result exists, availability or dispatch failure may
  try another qualified same-tier route, the planned tier, higher
  evidence-qualified routes, or the current model as a degraded fallback when
  policy permits. Once a lower-tier result fails deterministic proof, do not try
  another lower-tier worker: invoke R15 planned-tier-or-higher surgical
  correction. This explicit attempt is not an inferred downward fallback;
  the selector must reject an unrequested `attempt_tier`. Never assume a model
  visible on Codex is visible on GHCP or a local provider. A planner model may
  do worker tasks when independently eligible or requested. Keep a finite
  per-slice attempt ledger and never retry the same route in one dispatch cycle.
  If policy forbids fallback, pause only that slice and continue unrelated work.
- R14. Before dispatch, show only non-empty groups. For a bounded lower-tier
  attempt, show at most one beginner-readable line, for example `Trying Small
  for a bounded, objectively proved change; Medium correction fallback.`
  Otherwise show the intended
  planned-tier route and bounded fallback. Concrete arrows name only
  dispatch-qualified routes and label whether exact route-bound proof already
  exists. This is a work-time preview, not a plan commitment.
- R15. If an attempt fails proof, send an eligible planned-tier-or-higher
  correction route the original
  packet, accepted result, exact failed criterion/location, smallest allowed
  change, preserved invariants, relevant interfaces, current file and compact
  diff, failing proof/artifact, attempt ledger, and focused plus regression
  checks. It returns a corrective diff and must preserve independently accepted
  hunks; do not
  restart, rewrite the whole file, or spend stronger-model tokens redoing proven
  work. Before pilot promotion, a defect that cannot be localized, crosses an
  interface or authority boundary, invalidates the plan, or defeats focused
  correction aborts the surgical path and starts separately measured ordinary
  planned-tier execution.
  Provider, auth, quota, tool, flaky-test, weak-plan, and weak-packet failures
  are not automatically blamed on the model.
  In the initial pilot, the handoff is the smallest versioned correction packet
  linked to the attempt receipt that can measure preserved work. It records the
  attempt baseline, accepted diff/hunk hashes plus their machine-verifiable
  hunk-local acceptance oracle, exact failed
  criterion and location, permitted files/spans, preserved invariants, focused
  proof, and regression proof. A deterministic post-correction diff-scope check
  rejects changes to accepted or unrelated regions. Formatters, generated files,
  unstable anchors, or a defect that cannot be localized are ineligible for the
  surgical pilot and count against its benefit result. If proof cannot
  independently establish that a hunk is correct, record zero accepted hunks
  and declare the attempt ineligible for the surgical pilot. These failures do
  not trigger a generalized broadened-authority protocol before R22a passes.
  Generalize the packet and authority-expansion machinery only after the pilot
  earns promotion.
  The trusted Go driver constructs and validates all authority fields. Worker
  output, diffs, files, logs, and diagnostics are typed, delimited, size-bounded,
  redacted untrusted data referenced by contained path plus verified hash/size;
  the correction worker must be able to consume the bound artifacts without
  rediscovery. They cannot request credentials, new tools, new files, weaker
  proof, or expanded authority. Post-correction hunk and proof observations are
  computed by the trusted driver from disk, never accepted from worker self-report.
  The changed set is derived from a correction-only diff against a driver-owned
  pre-correction baseline; callers cannot omit unrelated edits from that set.
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
  Router-unavailable fallback still uses a router-independent bounded
  current-model subagent when the host supports it. If the host cannot spawn one,
  degraded direct orchestrator execution is explicit and preserves the same
  packet, authority, write-scope, proof, and receipt controls.
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
  Generated cross-host agents beyond the initial Codex profile/harness route are
  a later cohort.
- R18. Go owns discovery normalization, selection guards, route adapters, and
  routing receipts. Direct CLI/file operation is the universal baseline. The
  initial extra-model route is an OpenAI-compatible or LiteLLM endpoint invoked
  through a trusted user-local Codex profile and the bounded Codex coding-agent
  harness. Direct chat-completions dispatch is unavailable until a versioned
  agent runtime supplies bounded filesystem tools, proof execution,
  containment, receipts, and conformance fixtures. Generic MCP model dispatch
  likewise remains unavailable until a versioned adapter and fixtures exist.
  Ordinary users install prebuilt artifacts and do not need a Go toolchain. If the matching binary is
  absent or fails to start, only model routing degrades: `kb-work` uses the
  current model, records `router-unavailable`, and preserves every ordinary
  proof gate. Skill-only file-copy installs therefore remain functional with
  routing disabled.
  Failure to verify or establish owner-only permissions, ownership, safe-link
  containment, or atomic-write guarantees disables routed dispatch before
  sensitive state is written or reused. Record a redacted security-state failure
  and continue through the current-model proof path when policy permits.
- R19. Record the chosen route, fallback/escalation, adapter, KB `run_id`,
  orchestration/provider `session_id`, and provider-reported model when
  available. Attempt and correction receipts record their phase, parent/child
  linkage, proof artifact, and correction-boundary result. Routing evidence is separate from work
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
  Only an exact route-bound KB receipt linked to deterministic work proof may
  upgrade `dispatch-qualified` to `dispatch-proven` and establish capability
  success. Catalog/evidence files have strict schemas and size limits,
  trusted writers, permission-preserving atomic writes, and no symlink/path
  escape; repository claims and model self-report are not capability evidence.
  Credit a route only when its receipt proves exact route-bound dispatch with no
  model mismatch; stronger surface/provider evidence may raise confidence. Missing,
  unknown, or mismatched attribution records an observation but cannot credit or
  degrade a model.
- R22a. Evidence-qualified planned-tier host-native selection is the zero-setup
  baseline. Substantive next-lower AMR attempts are disabled by default and
  pilot-only until promotion; they require an explicit pilot/opt-in. An exploratory
  run uses at least 20 representative, objectively proved slices across at least
  three straightforward task families, but cannot by itself certify public
  non-regression. Promotion requires a preregistered independent known-answer
  corpus, a power-justified sample size, zero observed right-to-wrong outcomes,
  and a one-sided 95% confidence bound excluding more than a 2 percentage-point
  correctness regression. The measurement boundary includes discovery, driver,
  attempt, proof, correction, ordinary fallback, review, retries, and human
  intervention. It also requires zero accepted-hunk violations or added human
  interventions and at least 20% median improvement in one predeclared
  all-inclusive primary metric: total billed tokens/cost, wall-clock time, or
  proven offloaded throughput. Positive results below 20% remain advisory.
  Neutral/worse all-inclusive benefit or increased rework/collateral diff
  disables next-lower attempts. Unavailable cost/usage cannot justify promotion.
  The corpus uses held-out known-answer cases, contamination controls, protected
  pre-attempt oracles, and a proof-quality audit that rejects mocked or
  implementation-shaped theater.
- R22a1. Correction safety is a separate mandatory promotion stratum with a
  preregistered, power-justified minimum of localized failure cases across
  multiple eligible task families. Every case has independently established
  accepted hunks, an exact failed criterion, planned-tier-or-higher correction,
  focused plus regression proof, and a passing preservation/scope check.
  Protected seeded-fault fixtures may fill a shortfall in naturally observed
  failures but never contribute to the primary efficiency-benefit metric.
- R22b. Promotion is a release-level gate for one adapter/runtime/selection-policy
  and proof-harness cohort. User-local route evidence informs that user's
  selection but cannot promote public defaults. Adapter/runtime/schema,
  selection-policy, protected-oracle, or proof-harness changes invalidate the
  release gate. Public next-lower promotion is also bound to exact tested
  model/route revisions; a new revision returns to pilot-only until matching
  route-bound evidence exists. Record owner, version, expiry, reset reason, and
  benchmark artifact.
- R22c. Unattributed `host-auto` evidence is route-level only, keyed by
  orchestration surface, host policy/catalog fingerprint, adapter revision, task
  envelope, and proof. It may support bounded planned-tier host selection but
  never model-level capability credit or next-lower AMR qualification without
  exact model attribution.
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
- R26a. Host-native automatic choice is the zero-setup baseline, including
  surfaces such as Codex App where the host selects unpinned subagents but does
  not expose exact model attribution. The initial exact routed-dispatch cohort
  is Codex CLI plus one
  OpenAI-compatible/LiteLLM extra route invoked through a trusted Codex profile
  and the Codex harness. Codex App native choice remains usable but
  model-unattributed until an App-specific catalog, explicit selector, route-bound receipt,
  and conformance fixtures exist. Add GHCP to the supported label only after its
  independent conformance fixtures pass. A passive LiteLLM endpoint may serve as
  a private proving route; TinyBoss reserve/wake/start/controller actions are
  separately parked and never a machine-specific public default.
- R27. Public artifacts credit Phoenix and other prior art for the specific
  mechanics they informed. Comparisons stay factual, evidence-based, and free of
  personal commentary.

## Success Criteria

- Initial release: host-native automatic choice works without setup; Codex CLI
  plus one OpenAI-compatible/LiteLLM route through a trusted Codex profile and
  harness reaches exact routed work
  without a model-setup question and proves multi-model discovery/dispatch end
  to end; a current-model-only surface degrades cleanly.
- Later GHCP supported label: GHCP independently passes multi-model discovery,
  named-subagent dispatch, fallback, and receipt conformance fixtures.
- Codex CLI discovery reflects its live catalog without a static hosted-model
  list. Codex App may use host-managed unpinned selection but cannot claim an
  exact routed model until its separate selector and receipt fixtures pass.
- The first `kb-work` run discovers the current surface catalog, shows conservative
  slice choices and fallbacks, and dispatches model-backed subagents without
  requiring model IDs or route aliases in the plan.
- A user can register stable local/custom aliases once through supported
  OpenAI-compatible or LiteLLM adapters and make them available to any
  project without exposing credentials or transport details in plans.
- Ordinary work silently uses automatic source choice. Explicit setup may ask
  one project preference question, and the saved personal choice survives later
  sessions without modifying the repo unless the user explicitly asks for
  shared policy.
- A bounded Medium slice may explicitly try Small; passing proof keeps the
  result, while failing proof sends a surgical packet to Medium. Without attempt
  eligibility it begins at Medium. Route and proof outcomes are recorded.
- A model visible in a surface catalog but not proven selectable is never promised
  as the dispatched subagent.
- A failed attempt escalates with accepted hunks, exact failure context, and
  failing proof instead of repeating the slice from scratch or rewriting the
  whole file.
- The correction diff passes a deterministic preservation/scope check backed by
  hunk-local acceptance evidence. Ineligible broadening aborts the pilot path and
  becomes separately measured ordinary planned-tier execution; passing tests
  alone cannot prove surgical preservation.
- `ignore model routing` uses the current/user-selected model while preserving
  all ordinary work-proof requirements.
- Correct pre-existing work completes even when model provenance is unknown.
- A missing/incompatible Go router falls back to the current model and ordinary
  KB proof instead of blocking work.
- A clean user can install without a Go toolchain, start work without answering
  a questionnaire, understand one compact routing preview, and continue safely
  when discovery is unavailable.
- A representative baseline comparison proves next-lower AMR attempts do not
  increase first-pass proof failures, repeat-work, or user interventions before
  those attempts become a public default, and records at least one material
  efficiency/throughput benefit.

## Scope Boundaries

- No mandatory per-project model questionnaire or role assignment.
- No routing-priority prompt when no configured extra route is eligible.
- No dispatch-time priority prompt during ordinary work; unsaved preference
  means automatic source choice.
- No hardcoded claim that every Codex, GHCP, or local surface exposes the same
  model names.
- No provider purchase, authentication, entitlement change, or secret creation.
- No weakening of tests, review, browser proof, or acceptance for smaller
  models.
- No automatic global skill promotion or deletion.
- No GHCP supported label, exact Codex App routing claim, direct generic
  chat-completions worker, MCP model dispatch, or state-changing TinyBoss/fleet
  controller operation in the initial cohort.
- Project-local skill anti-sprawl remains a separate follow-up initiative and is
  not a release dependency for session model routing.
- No requirement to install Phoenix's lifecycle skills. Focused MCP
  interoperability remains eligible.

## Key Decisions

- Native discovery first; `kb-models` only adds what the surface cannot know.
  Evidence: live Codex catalog inspection and surface-specific model catalogs.
- No routine setup questionnaire. The work orchestrator chooses routes
  automatically; an optional persisted project-priority question appears only
  during explicit setup/configuration. Evidence: user requirement; reversible
  through per-run overrides and clear/reset.
- User-local config owns machine-specific connection details and personal
  project priority. Optional tracked project policy contains only explicit
  shared alias constraints. Evidence: user decision and portability/security
  boundaries.
- Plans freeze task complexity, not models, route aliases, source priority, or
  transport. The master resolves native and configured extra routes at work
  time, and the receipt records the actual route. Evidence: model availability
  differs by surface, provider, account, machine, and session.
- Try smaller only when bounded evidence makes failure cheap and surgical.
  Otherwise bias upward. Evidence: the cost of one stronger dispatch is often
  lower than repeating failed work, while a proven bounded attempt can save
  tokens only when accepted work survives escalation. Compare correctness,
  total tokens/time/cost, repeated hunks, collateral diff, escalation rate, and
  user intervention against direct planned-tier execution; remove the attempt
  machinery if results are neutral or worse.
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
- [safe-assumption] Personal project source priority is a versioned user-local
  collection keyed by canonical project identity. Reversible because it changes
  selection order only, has a matching clear/reset path, and never grants trust.
  Proof: multi-project isolation, canonical path/worktree matching, no-repo-write,
  precedence, and clear/reset fixtures.
- [safe-assumption] Canonical project identity derives only from trusted local
  filesystem and VCS facts, never repository content. Worktrees of the same
  repository share preference, while route approval remains bound to an
  immutable route fingerprint plus repository identity. Unrelated clones, path
  reuse, changed origins, and unexpected identity replacement require
  reapproval; symlinked paths resolve to the same canonical identity. Proof:
  clone, rename, worktree, symlink, replacement, and origin-change fixtures.
- [safe-assumption] Self-hosted classification is explicit declared route
  metadata, independent of trust boundary and route origin. Reversible because
  `unknown` remains eligible under automatic selection and the user can update
  the route declaration. Proof: private-provider-hosted and LAN-self-hosted
  counterexample fixtures.

## Alternatives Considered

- Ask every project to define Planner/Small/Medium/Large: rejected because native
  hosts already expose usable models and the questionnaire adds ceremony.
- Hardcode hosted model/version classes globally: rejected because catalogs and
  access differ by orchestration surface, account, date, and inference provider.
- Freeze model routes in plans: rejected because `kb-work` owns live dispatch and
  plans should remain portable.
- Put personal project priority in tracked policy by default: rejected because
  route availability and cost are user/machine-specific; shared policy remains
  explicit.
- Store private endpoints in tracked project policy by default: rejected because
  local network routes are user/machine-specific.
- Use only the current model: retained as `ignore model routing`, not the default
  when safe subagent choices are discoverable.

## Slice Candidates (advisory for /kb-plan)

- Session catalog - goal/work discover native and configured subagent routes
  without asking the user to classify models.
- Extra-route catalog - users register local/custom OpenAI-compatible/LiteLLM models
  once and optionally constrain them per project.
- Conservative selector - work maps slice requirements to live models, previews
  choices, and escalates upward without losing context.
- Initial dispatch adapter - Codex CLI invokes native and one trusted-profile
  OpenAI-compatible/LiteLLM extra route through the bounded Codex harness.
- Routing evidence - receipts and telemetry explain actual choices without
  overriding work correctness.
- Distribution proof - Go packaging, fixtures, docs, sync, and install gates
  prove the feature across supported hosts.

## Outstanding Questions

### Resolve Before Planning

None.

### Deferred to Planning

- [parked][Affects R4/R17] GHCP catalog, named-subagent dispatch, fallback, and
  receipt conformance belong to a later supported cohort after the initial
  Codex CLI cohort passes.
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
