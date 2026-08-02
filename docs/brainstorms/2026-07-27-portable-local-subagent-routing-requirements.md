---
date: 2026-07-27
topic: portable-local-subagent-routing
brainstorm_style: kb-brainstorm
---

# Portable Local Subagent Routing

## Problem Frame

KB plans correctly record portable capability tiers instead of concrete models,
but optional local routes are not yet automatically usable as qualified
subagents. A user can register an OpenAI-compatible/LiteLLM endpoint and probe
its model list, yet that discovery remains a declaration rather than capability
and dispatch proof.

The requested outcome is a skeptic-ready demonstration that bounded work can
move from a hosted orchestrator to qualified home routes without weakening
proof, followed by a separate measured test of incremental operating-cost
savings.
The workflow therefore needs a portable, optional path where teams can document
how to connect to their own fleet without installing an MCP provider, copying
private endpoints, or using the project owner's infrastructure.

The first release is deliberately narrower than a general local coding agent:
it dispatches a bounded, tool-free child task through an OpenAI-compatible
inference endpoint. The orchestrator owns context assembly, output validation,
artifact application, and deterministic proof.

## Research Summary

**Findings that shaped requirements:**

- `kb-plan` already forbids model, route, provider, endpoint, adapter, and
  transport pins. The executing CLI owns route selection. Affected: R1-R3.
  Evidence: `.github/skills/kb-plan/SKILL.md`.
- `kbrouter models add` stores connection state under the operating-system
  user's `~/.kb`; repository policy may narrow stable aliases but cannot store
  connection details. Affected: R4-R8. Evidence:
  `.github/skills/kb-models/SKILL.md` and `cmd/kbrouter/catalog.go`.
- OpenAI-compatible model discovery proves presence only. Discovered child
  routes remain `EvidenceDeclared`, `DispatchQualified: false`, and cannot enter
  automatic selection. Affected: R9-R16. Evidence:
  `cmd/kbrouter/catalog.go` and `internal/modelrouting/selector.go`.
- Existing KB routing already has route fingerprints, capability-envelope
  hashes, `RoutingReceipt`, observation-only dispatch receipts, and
  dispatcher-owned HMAC attestations. The minimum change is to extend these
  surfaces rather than create a parallel receipt system. Affected: R12-R16,
  R19.
  Evidence: `internal/modelrouting/receipt.go` and
  `internal/modelrouting/dispatch_attestation.go`.
- `kbrouter dispatch` currently rejects direct-provider dispatch and requires a
  trusted Codex profile. OpenAI compatibility is therefore only the inference
  wire, not an existing subagent runtime. Affected: R9-R16. Evidence:
  `cmd/kbrouter/dispatch.go`.
- LLMCommune is a reservation/controller surface, not an inference proxy or
  chain executor. It returns direct call targets for OpenAI-family runtimes but
  supplies no inference aliases, child receipts, token persistence, energy
  telemetry, or structured-output guarantee. Fleet-specific reservation,
  readiness, and concurrency facts belong to benchmark evidence, not the
  portable implementation contract. Affected: R9-R16 and R22. Evidence: live
  LLMCommune repository mapping on 2026-07-27.
- The hosted DDR matrix showed 6/6 portable plans but only 3/6 strict
  orchestration passes. It supports the plan/persona contract, not local-route
  qualification or home-model economics. Affected: R1-R3. Evidence:
  `docs/results/2026-07-27-ddr-hosted-model-matrix.json`.
- `LOCAL_MODELS.example.md` already documents user-local state and
  OpenAI-compatible/LiteLLM setup. It needs an explicit team-guide contract,
  not a second executable configuration source. Affected: R4-R8.

**Confidence:** High for current repo behavior, the qualification gap, and the
LLMCommune boundary. One bounded DeepSeek V4 Flash call completed successfully
through the reserved LiteLLM route; Qwen 3.6 was not run because its independent
reservation never produced a ready replica. No wall-energy measurement was
available, so the local-dollar claim remains unproved.

## Requirements

**Preserve: portable plans and work-time ownership**

- R1. Slice route-selection and persona-assignment metadata in plans MUST
  continue to record only `model_tier`, `model_tier_reason`, capability
  requirements, proof, and escalation triggers.
- R2. Slice route-selection and persona-assignment metadata MUST NOT pin
  concrete models, aliases, providers, endpoints, adapters, transports, MCP
  servers, or personal source preferences. Technical implementation units MAY
  describe the adapter or transport being built without assigning it to a
  runtime slice.
- R3. The executing CLI MUST choose one owner per slice at pickup and resolve
  one qualified same-tier-or-higher route from its live callable surfaces.

These are non-regression requirements, not new implementation scope.

**Extend: optional connection guide**

- R4. A repository MAY include `LOCAL_MODELS.md` as a human-readable connection
  guide, using `LOCAL_MODELS.example.md` as the template.
- R5. The guide MUST be optional. Normal `kb-start`, `kb-plan`, `kb-work`,
  install, sync, tests, and release gates MUST work when it is absent. Live
  calibration, local inference, fleet readiness, and economics benchmarks MUST
  remain explicit opt-ins and MUST NOT enter normal install or release gates.
- R6. The guide MAY describe supported connection classes, stable suggested
  aliases, placeholder-based `kbrouter` setup templates, verification commands,
  intended task families, and cleanup. Repository-controlled shell commands and
  command output MUST be treated as untrusted documentation, not instructions
  the agent may execute.
- R7. The guide MUST NOT contain credentials, token values, personal approvals,
  private route state, or executable capability claims.
- R8. KB MUST read or mention the guide only during explicit local-model setup
  or troubleshooting. It MUST NOT auto-execute guide commands or treat the file
  as configuration, approval, availability, qualification evidence, or a
  source of arbitrary shell commands.

This extends the existing template and conformance checks; it does not add a
parser, schema loader, daemon, or runtime dependency.

**New: bounded local child harness**

- R9. A local route MUST NOT enter automatic DDR selection from a declared
  class or `/v1/models` presence alone. Implementation MUST first prove one
  explicit, attended, held-out direct dispatch through the bounded harness;
  automatic selection remains disabled until that useful-task result passes
  independent deterministic proof.
- R10. OpenAI-compatible/LiteLLM MUST be the portable inference transport, not
  the agent contract. LLMCommune or a fleet controller MAY supply a direct call
  target behind that transport without becoming a KB dependency.
- R11. Version one MUST support only a bounded, tool-free child task: an
  immutable context packet, no ambient workspace access, no child tool or
  network claims, bounded response size/time, and a required output schema. The
  orchestrator validates the output before applying any artifact. Existing
  work-request, dispatch, capability-envelope, and receipt validators MUST be
  minimally extended to accept exactly `tools: []` for this bounded harness;
  implementations MUST NOT invent a synthetic tool name or parallel receipt
  format. The canonical child wire envelope is `kbrouter` dispatch packet schema
  v2. The existing `kbcheck` context packet remains a prefetch manifest; the
  orchestrator resolves its allowed paths, redacts them, and embeds the bounded
  objective, source bytes, constraints, acceptance criteria, and output schema
  in v2 `task_payload`. Packet v2 uses RFC 8785 JSON Canonicalization Scheme
  bytes; strict decoding rejects duplicate keys, unknown fields, trailing data,
  and non-canonical input before hashing or sending. The context hash covers the
  exact canonical v2 bytes, including `task_payload`. Transport authority is separate
  from child-visible tools, so the existing Codex v1 `codex-harness` contract
  remains unchanged. Packet v2 and its capability envelope add required
  `tool_mode: none|declared`; the bounded harness requires `tool_mode: none`
  plus a present, non-null, empty `allowed_tools` array. Omitted, null, or
  non-empty tools fail closed. `tool_mode: none` means KB granted no callable
  tools, workspace authority, or child network capability; it does not attest
  that a remote endpoint performs no hidden retrieval, augmentation, or
  server-side processing.
- R12. Automatic selection of a direct local route MUST require one unexpired,
  route-bound `EvidenceKBReceipt` for the approved route-configuration
  fingerprint, task
  family, tier, context ceiling, `tool_mode: none`, `tools: []`, and low-risk
  boundary. Attended calibration verifies the installed bounded-harness-adapter
  revision, release-owned fixture hashes, policy version, and repeated-pass
  threshold inline before issuing that receipt. Consuming repositories and
  `LOCAL_MODELS.md` cannot replace or relax those installed release artifacts.
  The fixture set and policy digest are embedded in the authenticated `kbrouter`
  binary so version one adds no separate release trust root. Exact calibration
  instances MUST NOT be enumerable from those release artifacts: the trusted
  router derives held-out challenge instances after enrollment from
  release-owned generators, a user-local secret, and a unique nonce.
  Qualification also requires a preregistered held-out generalization check
  distinct from public examples and calibration echoes.
- R13. The first execution eligible to contribute qualification evidence MUST be
  an attended, fixture-only calibration with no project content or mutation. A
  single successful production task MUST NOT self-promote a route; the versioned
  calibration policy defines the repeated independent passes required to set
  `DispatchQualified: true` for the exact measured envelope. Independence
  requires unique challenge inputs and nonces, distinct dispatch attempts,
  duplicate response/receipt rejection, cache bypass, and minimum task diversity;
  each fact is bound into the qualifying receipt. After route approval, the
  repeated calibration attempts run behind one attended command; the report
  records elapsed time and every additional manual intervention.
- R14. Calibration evidence MUST be ingested into user-local state, merged into
  the run catalog, expired, and revoked when the route fingerprint, requested or
  provider-reported model, bounded-harness-adapter revision, endpoint approval,
  or measured capability envelope changes. The final `EvidenceKBReceipt` binds
  the installed adapter revision, release fixture hashes, every qualifying
  calibration receipt hash, measured envelope, policy version, and expiry.
  Direct local routes MUST NOT enter automatic selection from
  `EvidenceAdapterPrior` alone, and this work MUST NOT add a persisted
  multi-source evidence graph. One deterministic right-to-wrong production
  failure immediately demotes that exact task-family/envelope pending attended
  recalibration. Demotion is durably persisted under the dispatch/proof lock
  before the failing proof result is published; persistence failure blocks the
  slice, and every overlapping qualification record that could select the failed
  route for that request is revoked. The route-configuration fingerprint binds
  the approved destination, requested and provider-reported model strings,
  bounded-harness-adapter revision, sampling defaults, context configuration,
  and optional trusted-controller mapping revision. Without a trusted controller
  attestation, this is short-lived behavioral evidence for an approved route,
  not proof of exact weights or deployment identity.
  `LocalQualificationV1` is the authenticated persistence envelope whose
  authoritative payload is the final `EvidenceKBReceipt`. Discovery projects
  only its route fingerprint, task family, tier, context ceiling, `tool_mode`,
  empty tools, risk boundary, issue/expiry, and demotion state into run-catalog
  `CapabilityEvidence`; the projection cannot add authority.
  `kbrouter models calibrate --live` is the sole issuer of new qualification
  records. The trusted proof-credit path MAY only demote or revoke an existing
  record through the same private-state writer transaction; no other path may
  create or upgrade qualification. The `LocalQualificationV1` record lives under
  `~/.kb/qualifications/<route-hash>/<task-family>.json`. The HMAC-authenticated
  record contains schema version, route-configuration fingerprint, router and
  embedded-fixture digests, policy version, unique challenge/dispatch/receipt
  hashes, measured envelope, issue/expiry times, and demotion status. Discovery
  verifies that record against the running router and user-local HMAC before
  projecting its bounded fields into run-catalog `CapabilityEvidence`; user and
  native catalogs remain declaration-only.
- R15. Direct local dispatch MUST extend the existing route fingerprint,
  capability-envelope hash, `RoutingReceipt`, observation-credit, and HMAC
  attestation path. Dispatcher-attested receipt bytes remain observation-only;
  independent deterministic proof later credits or rejects the observation
  without mutating the attested receipt. Version one adds exactly one linked
  acceptance-credit record under the existing telemetry boundary. It references
  the attested receipt hash plus proof artifact hash/result, is HMAC-authenticated
  by the existing user-local attestation authority, and MUST be recorded before
  the dispatch attestation expires. It is not a `RoutingReceipt`, cannot qualify
  a route by itself, and does not create a second receipt framework.
  Direct-HTTP receipts label attribution as `endpoint-self-reported` unless a
  trusted controller reservation snapshot independently binds the call target
  and route configuration. Dispatcher HMAC and deterministic output proof do
  not upgrade endpoint-reported model identity.
  Packet v2 uses `DirectHTTPAttributionV1`: a client-generated request ID and
  nonce, required normalized provider-reported model, optional provider response
  ID, approved origin hash, response hash, and attribution class. It does not
  populate a synthetic Codex `session_id`. Missing/mismatched model, request ID,
  nonce, origin, or response hash fails closed; a missing provider response ID
  remains explicitly endpoint-self-reported. The request ID and nonce MUST be
  present in the authenticated request and schema-valid child response and MUST
  be covered by the response hash; production duplicate responses fail closed.
  An optional trusted-controller snapshot may add issuer, snapshot hash,
  call-target hash, mapping revision, freshness, and expiry but is not required
  for portable behavioral evidence. A `self-hosted` or home-model attribution
  claim requires an attended, user-approved controller issuer, an authenticated
  snapshot, verified issuer-key lifecycle and revocation state, and a fresh
  call-target binding. If that support is absent or verification fails, the
  attribution remains `endpoint-self-reported` and MUST NOT be described as
  self-hosted.
  `AcceptanceCreditV1` is written only by the trusted `kbrouter proof-credit`
  path under
  `~/.kb/acceptance-credits/<project>/<run>/<slice>/<attempt>.json`. Its unique
  key is project/run/slice/attempt/receipt hash; an identical retry is
  idempotent, a conflicting write fails closed, and a write after dispatch
  attestation expiry cannot credit or apply the candidate. Telemetry and apply
  gates consume this record instead of mutating or re-evaluating
  `RoutingReceipt.WorkProof`.
- R16. The end-to-end dispatch-and-acceptance workflow MUST fail closed on
  unavailable, timed-out, oversized, empty, or malformed output; missing or
  mismatched required model/execution attribution; stale, malformed, or
  unverified qualification evidence; failed receipt persistence or attestation;
  and failed deterministic proof. Immediately before transmitting any context
  bytes, the dispatcher MUST reread authoritative approval,
  revocation, route-fingerprint, and calibration state rather than trust a
  cached catalog. Missing state, changed state, or revocation invalidates cached
  eligibility and fails closed. A bounded live readiness probe MUST also pass
  immediately before send; endpoint discovery alone is insufficient. Approval,
  revocation, calibration, and dispatch share one user-local private-state lock
  and monotonic trust generation; run-specific dispatch locks are separate and
  MUST have a documented acquisition order. Dispatch binds the checked
  generation into the packet and receipt, then uses a bounded, cancellable
  transmission without blocking trust mutation. Revocation commits immediately
  and cancels matching in-flight transmission. The workflow revalidates
  generation, approval, qualification, route fingerprint, and expiry before
  staging, proof credit, and artifact application; any change rejects the
  in-flight result. It MUST NOT silently switch owner or route.

**Authoritative dispatch lifecycle**

1. **Selected:** the exact qualified route and immutable context packet are
   bound; no child output or project artifact exists.
2. **Observed:** the bounded child returns a schema-valid candidate; the
   dispatcher persists the candidate by hash/path, writes an observation-only
   `RoutingReceipt`, and HMAC-attests the exact receipt bytes. Nothing is yet
   accepted or applied to the owned project state.
3. **Staged:** the orchestrator materializes the candidate only in its owned
   staging workspace and verifies scope, attribution, schema, and integrity.
4. **Proved:** deterministic proof runs against the staged candidate. A separate
   immutable acceptance-credit record references the dispatch receipt hash and
   proof artifact hash; it credits or rejects that dispatch observation without
   rewriting the attested receipt. A missing or late record leaves the
   observation uncredited.
5. **Accepted or rejected:** only a passing acceptance record may retain/apply
   the staged artifact to the slice's owned worktree. Failure deletes the
   candidate and output bodies but retains the immutable rejection
   `AcceptanceCreditV1`, proof hash/result, and bounded receipt tombstone. It
   leaves the observation uncredited and requires an explicit new route/owner
   decision or block. Application is an exactly-once transaction keyed by the
   acceptance-credit and receipt hashes, with durable terminal state and
   pre/post project-state hashes; conflicting replay fails closed.

Candidate persistence, receipt/attestation writes, staging, proof, and
application use atomic state transitions with startup reconciliation. Incomplete
transactions are rejected and scavenged after a bounded retention period;
crash recovery cannot retain orphan candidate bodies or silently replay project
mutation.

**Extend: trust and data boundaries**

- R17. Route enrollment and endpoint approval MUST remain attended and
  user-local, scoped to the canonical project and capability envelope, and
  revocable. Configuration stores only an authentication environment-variable
  name; credential values MUST NOT enter route state, logs, receipts, guides, or
  artifacts. Credentials MUST be least-scope, environment-separated, rotatable,
  and revocable; compromise revokes the route until attended reapproval.
- R18. Local dispatch MUST use an allowlist of exact source bytes assembled for
  the slice, not a denylist of known secrets. The packet builder MUST reject
  detected credentials and unknown/high-sensitivity content from automatic local
  dispatch; version one has no disclosure override for those classes. A
  release-owned, versioned classifier is authoritative for each embedded source
  and derived payload. Unsupported encoding/type, detector failure, absent
  classification, or unclassifiable derived content is `unknown` and fails
  closed. Any
  non-loopback destination requires authenticated encryption plus the existing
  attended endpoint/project approval; there is no plaintext private-network
  exception and no new disclosure-approval subsystem. Transport enforces the
  exact approved origin, verified peer identity and hostname, disabled redirects,
  explicit proxy policy, and destination-address revalidation; any certificate,
  origin, proxy, redirect, or DNS mismatch fails closed.
- R19. Receipts and attestations MUST remain user-local, bounded, and
  access-restricted. They store redacted route identity plus hashes/paths for
  retained accepted/proof artifacts, not endpoint URLs, credential values,
  usernames, prompt bodies, repository contents, or output bodies. A rejected
  candidate body is deleted; its receipt retains only bounded hash, size,
  rejection status, and tombstone metadata. Version one reuses the existing
  access-restricted, short-lived HMAC key and attestation expiry; generalized
  signing-key lifecycle management is separate hardening work. Candidate,
  staging, proof, and transport-diagnostic storage MUST use the same private
  permissions, body-free logs, atomic creation, bounded retention, and
  crash-cleanup policy.

**Threat boundary**

- In scope: malicious or mistaken repository documentation, substituted
  calibration files, stale approvals, endpoint interception, hostile responses,
  and receipt modification by actors without the user-local signing key.
- Out of scope: arbitrary code already executing with the operating-system
  user's full authority. Such code can compromise environment credentials,
  user-local trust state, and exportable HMAC material; KB MUST state that limit
  and MUST NOT present its HMAC as protection from same-user compromise.
- Within the stated no-same-user-compromise boundary, the HMAC verifies
  possession of the user-local key and integrity of the recorded receipt bytes
  and packet/run metadata. It does not independently prove which process wrote
  the record, remote execution, model identity, or truthful provider
  attribution.

**Preserve: selection and proof; evaluate economics separately**

- R20. Selection MUST prefer the exact minimum qualified tier before considering
  a higher tier. `self-hosted-first` remains a user-local preference and applies
  only after eligibility and exact-tier filtering; within that eligible
  same-tier set, it may reorder self-hosted routes ahead of native routes but
  MUST NOT relax the evidence floor. In this document, `local route` means a
  user-registered route; `self-hosted route` is the subset running on
  user-controlled infrastructure.
- R21. Deterministic proof requirements MUST remain identical regardless of
  route cost, hosting, or model tier.
- R22. Cost MUST NOT control route admission or weaken qualification. Economic
  promotion is a follow-on evidence gate, not a functional version-one release
  gate. DDR economics extends the existing `cmd/kbcheck benchmark-validate` and
  `evals/cross-model-benchmarks` validation spine with a paired-cost result
  schema; `cmd/amrbench` remains AMR-specific and may contribute shared
  statistical helpers without owning DDR semantics. The DDR report presents
  incremental operating dollars per accepted result: hosted billed credits;
  measured local energy when available; and all retries/failures in the
  numerator. The operating-cost view includes hosted
  orchestration, calibration amortized over
  its qualification window, failed pre-promotion work, idle/cooling allocation,
  and wall power; a marginal dispatch-only view may be shown separately.
  Hardware amortization is reported as a separate all-in ownership sensitivity
  view. If
  measured wall energy and allocation inputs are
  unavailable, local dollar cost is `unavailable`; the user's $1/day estimate
  may appear only as a labeled scenario, not measured savings. A functional
  routing pass with unavailable cost does not pass the economic-promotion gate.
  Any savings comparison MUST bind hosted and local arms to identical canonical
  task-packet bytes, deterministic proof, retry/timeout budgets, repetition
  counts, acceptance denominator, and end-to-end boundary including hosted
  orchestration plus every child attempt. Before execution, the benchmark MUST
  preregister electricity tariff, wall-power measurement, idle/cooling and
  controller-overhead allocation, calibration window, recurring operator-labor
  treatment, hardware life/residual value/utilization for the sensitivity view,
  paired analysis, and decision rule. The benchmark MUST preregister immutable
  task IDs, assignment, order, and attempt limits. Promotion requires a fixed
  cohort of at least ten unique paired tasks per route for each representative
  task family; every assigned task must reach an accepted result in both arms
  within those limits. Every attempt and failure remains in the cost numerator.
  Promotion also requires zero right-to-wrong regressions, a mean incremental
  operating-cost reduction of at least 20%, and a 95%
  bootstrap confidence interval for paired cost difference
  (`local - hosted`) wholly below zero. Hardware amortization remains a separate
  sensitivity view. The unqualified word “cheaper” requires that all-in view to
  agree; otherwise the claim is limited to measured incremental operating cost.
  Infrastructure preflight MUST pass before a route enters model correctness or
  cost-per-accepted-result denominators. Startup, capacity, credential, endpoint,
  or readiness failures are recorded as `test_status: not-run` with a separate
  infrastructure outcome and setup cost/time. They remain visible in the
  intention-to-test cohort and availability report, MUST be repaired and rerun
  against the same preregistered task, and leave economic promotion unavailable
  if unresolved rather than being replaced or silently excluded.

## Success Criteria

- A fresh user can clone a consuming repo without `LOCAL_MODELS.md`, MCP, fleet
  access, or local routes and run the normal KB workflow unchanged.
- A team member can follow a checked-in `LOCAL_MODELS.md`, create only
  user-local route/trust state, run attended fixture-only calibration, and reach
  one automatically selected bounded child whose output passes deterministic
  acceptance proof. The proof run sets `self-hosted-first`, provides an eligible
  same-tier hosted comparator, and asserts that the selected route is the
  qualified user-registered route. It may assert self-hosted execution only when
  the R15 independent infrastructure-attribution contract passes.
- The same user journey works from a direct OpenAI-compatible/LiteLLM call target
  without MCP, LLMCommune, FleetController, or a provider-specific KB adapter.
- Within the stated no-same-user-compromise boundary, the resulting HMAC
  attestation verifies possession of the user-local key and integrity of the
  exact receipt bytes and packet/run metadata; it does not independently attest
  the writer process, endpoint execution, or model identity. Independent proof
  may credit only that route-bound observation without mutating attested bytes.
- Removing the guide or local route returns behavior to live host-native
  selection without modifying the plan.
- A negative fixture proves that a guide, declared class, `/v1/models` response,
  repository alias policy, single production receipt, stale calibration,
  mismatched model, malformed output, or tampered receipt cannot promote or
  credit a route or credit a dispatch observation. It also proves omitted/null
  tool fields, changed route-configuration fingerprints, mismatched installed
  fixture/adapter/calibration hashes, and one right-to-wrong production result
  demote or reject the exact envelope.
- Revoking approval removes the route from eligibility; known secret-bearing
  and unknown/high-sensitivity input is rejected; an unauthenticated or plaintext non-loopback
  endpoint fails with no exception; and receipt/attestation fixtures prove endpoint,
  credential, username, prompt, repository-content, and output-body redaction.
- A consuming repository cannot substitute calibration policy/fixtures, and a
  revocation made after catalog discovery but before send prevents dispatch.
- The local benchmark either reports measured energy and accepted-result cost or
  explicitly reports cost as unavailable; it never presents the $1/day scenario
  as observed evidence.
- Infrastructure-not-ready arms are excluded from model pass/fail denominators,
  retain a null model outcome, and carry an explicit rerun requirement after a
  green preflight.
- Functional-route success and economic-promotion success are reported
  separately. “Cheaper” requires the paired R22 protocol; unavailable energy or
  fewer than ten unique paired accepted tasks per family leaves economic
  promotion unproved.
  The economically evaluated child route MUST pass R15 independent
  self-hosted-attribution checks; provider-hosted or hosting-unverified routes
  registered in user-local state do not satisfy the home-model claim.
- After calibration, at least one useful bounded non-sensitive task fixture
  distinct from the calibration prompts must produce an artifact or decision
  consumed by the orchestrator and accepted by an independent deterministic
  oracle. Calibration echoes do not count as user-value proof. Representative
  task families are: portable DDR plan classification/audit with an exact JSON
  oracle; pure-function implementation such as `ParsePort` with closed unit
  tests; and bounded repository-policy extraction with expected-fact checks.
  Each family uses preregistered held-out tasks with explicit family boundaries;
  at least one result must supply a downstream artifact or decision that would
  otherwise require hosted execution. Economic promotion covers all three and
  includes at least one implementation artifact.

## Scope Boundaries

- Generic MCP model dispatch is not required and remains out of scope.
- The repository does not distribute model weights, fleet credentials, private
  endpoints, or personal trust approvals.
- The connection guide is not a package manager, daemon bootstrapper, or
  executable policy file.
- Tool-capable local coding agents, direct workspace mutation, and arbitrary
  provider adapters are not part of version one.
- Automatic lower-tier AMR attempts remain separate from production DDR.
- This work does not hard-code DeepSeek, Qwen, or any hosted model into a plan.

## Key Decisions

- **Keep plans model-agnostic:** live catalogs and local availability change;
  durable plans should survive different hosts. Evidence:
  `.github/skills/kb-plan/SKILL.md`.
- **Use a human guide, not project connection config:** this preserves
  portability and prevents repository content from granting trust. Evidence:
  `.github/skills/kb-models/SKILL.md`.
- **OpenAI-compatible/LiteLLM is the baseline:** it avoids an MCP requirement
  while supporting direct home-fleet call targets. It is only the inference
  wire; the bounded-harness adapter supplies context/output/receipt semantics.
  Evidence: `LOCAL_MODELS.example.md` and the 2026-07-27 LLMCommune mapping.
- **Presence is not qualification:** `/v1/models` proves that a name exists, not
  that the route can safely execute a slice. Evidence:
  `cmd/kbrouter/catalog.go`.
- **Extend existing evidence machinery:** route fingerprints, capability hashes,
  routing receipts, and dispatcher attestations own observation integrity;
  linked `AcceptanceCreditV1` owns post-proof credit without rewriting those
  bytes. Evidence: `internal/modelrouting/receipt.go` and
  `internal/modelrouting/dispatch_attestation.go`. The separate credit is needed
  because independent proof completes after the receipt bytes are attested;
  mutating `RoutingReceipt.WorkProof` would invalidate that immutable
  observation.
- **Keep economics in the benchmark:** cost per accepted result evaluates DDR;
  it does not authorize a route or replace capability evidence. Evidence:
  `docs/context/research/2026-07-27-ddr-cost-routing-benchmark.md`.

## Dependencies / Assumptions

- [safe-assumption] `LOCAL_MODELS.md` is the conventional consuming-repo guide
  name. Reversible because it is informational and can be renamed without
  changing runtime state. Evidence/proof: `LOCAL_MODELS.example.md` remains the
  canonical template and normal gates ignore both files.
- [safe-assumption] Local fleets can expose an OpenAI-compatible or LiteLLM
  endpoint. Reversible because unsupported fleets remain host-native only.
  Evidence/proof: LLMCommune exposes direct OpenAI-family call targets for
  `vllm`, `trtllm`, `llama.cpp`, and `litellm`; conformance fixtures catch
  unsupported endpoint behavior.
- [safe-assumption] A tool-free child harness is sufficient for the first
  local-model proof. Reversible because tool-capable work remains ineligible and
  host-native; proof fixtures must demonstrate useful bounded tasks before the
  envelope expands.
- [safe-assumption] Arbitrary same-user code execution is outside the trust
  guarantee. Reversible because the local route can be disabled without changing
  plans. Evidence/proof: threat-model tests cover repository-controlled content
  and non-key-holder tampering, while documentation forbids a stronger claim.

## Alternatives Considered

- **Require the project owner's MCP:** rejected because it makes an optional
  private provider a runtime dependency and does not travel across users.
- **Commit endpoint/model configuration:** rejected because it leaks
  machine-specific state and conflates documentation with trust.
- **Trust declared model classes:** rejected because a file or user claim cannot
  prove tools, context, risk, containment, output shape, or correctness.
- **Put model names in plans:** rejected because plans would become stale and
  host-specific.
- **Make LLMCommune a KB adapter:** rejected because it supplies reservations and
  direct call targets, not inference proxying, agent loops, receipts, or
  telemetry.
- **Build a parallel receipt format:** rejected because KB already has route
  evidence, proof credit, and HMAC-attested dispatch receipts.

## Slice Candidates (advisory for /kb-plan)

- Portable connection guide contract - teams can document optional setup
  without creating an executable dependency.
- Bounded local child calibration - an attended fixture proves the exact
  deployment envelope without project access or mutation.
- Qualified direct child dispatch - an eligible route receives one immutable
  context packet and returns one schema-bound result through the existing
  receipt/attestation path.
- Independent acceptance credit - deterministic proof credits or rejects the
  attested dispatch observation without rewriting its receipt.
- Negative conformance suite - repository files and model-list presence cannot
  self-authorize a route, and stale/tampered/mismatched evidence fails closed.
- Accepted-result economics report - the benchmark reports hosted billed cost,
  measured local energy when available, explicit unavailability otherwise, and
  retries/failures in the numerator under an identical paired-arm protocol.

## Outstanding Questions

### Resolve Before Planning

None.

### Deferred to Planning

- [defer-to-planning][Affects R12-R14][Technical] Define the versioned
  calibration fixture set, repeated-pass threshold, evidence expiry, and
  route-configuration fingerprint fields without expanding beyond `tools: []` and
  low-risk bounded tasks; emit one final `EvidenceKBReceipt`.
- [defer-to-planning][Affects R11-R16, R19][Technical] Define the smallest direct
  HTTP harness, dispatch packet v2 `task_payload`, and schema-bound output format
  that reuse existing route fingerprint, receipt, attestation, and telemetry
  types.
- [defer-to-planning][Affects R18][Security] Reuse existing destination and
  project approval policy for non-loopback endpoints; define exact peer,
  redirect, proxy, DNS, classifier, cancellation, and generation-revalidation
  mechanics.
- [defer-to-planning][Affects R14-R16][Concurrency] Define the private-state and
  run-lock acquisition order, cancellation protocol, crash reconciliation, and
  exactly-once application transaction without holding locks across network or
  proof execution.

### Parked / Out of Scope

- [parked][Affects R11] Generic MCP dispatch adapter. Forbidden claim: the
  current router can dispatch arbitrary MCP-hosted models.
- [parked][Affects R11] Tool-capable local coding agent. Forbidden claim: an
  OpenAI-compatible endpoint alone supplies workspace, tools, containment, or an
  autonomous agent loop.
- [parked][Affects R22] Cost-optimizing selector. Forbidden claim: the router
  automatically chooses the cheapest qualified route from measured dollars.

## Next Steps

-> /kb-plan
