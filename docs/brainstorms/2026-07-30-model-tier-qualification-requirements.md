# Model Tier Qualification Requirements

Status: ready-for-planning
Created: 2026-07-30

## Problem Frame

DDR currently records a requested tier and whether one routed attempt passed
proof, but it does not provide a reusable, deterministic way to decide whether
a concrete model route has earned a tier classification. A failed run may be a
model failure, or it may come from an insufficient plan, invalid oracle, schema
violation, unavailable route, or infrastructure failure. Treating all failures
as model evidence misclassifies models and hides planning defects.

Deepseek4 is the first subject. Existing evidence says it passed five bounded
Medium asks, while several apparent cohort failures were later attributed to
infrastructure, an unstated protected-test requirement, or insufficient
planning. The classifier must re-evaluate that evidence without hardcoding a
Deepseek4 verdict.

## Requirements

**Evidence admission**

- **R1:** Accept one bounded, versioned, strict evidence file. Reject unknown
  fields, external references, links, absolute paths, path traversal,
  symlinks, oversized/deep structures, duplicate keys/IDs, and sensitive
  values. The scorer performs no network access or inference.
- **R2:** Bind the file to a preregistered cohort manifest frozen before the
  first attempt. The manifest defines the target tier, independently selected
  fixtures/families, threshold revision, execution envelope, and complete run
  ledger. Missing eligible attempts, undeclared retries, or a changed cohort
  manifest invalidate the evidence document.
- **R3:** Use an immutable execution fingerprint covering the privacy-preserving
  route fingerprint, provider model revision when observable, route
  configuration revision, system instructions, tools, context/risk policy,
  planner/scorer/oracle versions, and evidence date. Mixed fingerprints cannot
  share one qualification result.
- **R4:** Give every attempt a stable identity bound to the cohort, fixture and
  fixture revision, frozen plan and sufficiency receipt, target tier, route
  fingerprint, request hash, response hash, and proof hash. Identical imports
  deduplicate; conflicting or overlapping identities invalidate the evidence.
- **R5:** Domain owners emit normalized, hash-bound receipts:
  `kb-plan`/`kbcheck` own plan sufficiency; the route adapter owns readiness and
  execution identity; the executor wrapper owns one-attempt/schema evidence;
  and the proof runner owns oracle qualification and proof outcome. The
  classifier verifies these receipts, counts attempts, and applies thresholds;
  it does not reimplement their workflows. Unsupported self-authored claims
  make the result `inconclusive`.
- **R5a:** Each receipt is either independently reproducible from included
  redacted artifacts or signed with an issuer key trusted by user-local
  classifier policy. The trust policy maps issuer IDs to allowed receipt
  classes and key-validity windows; unknown, expired-at-signing, or revoked
  issuers are unsupported. Private keys and trust policy are never committed.
  The result lists which facts were reproduced, signature-verified, or
  unsupported.
- **R6:** Artifact hashes cover the cohort manifest, fixture, frozen plan,
  plan-sufficiency receipt, public executor package, redacted route/execution
  receipt, oracle qualification receipt, and proof result. Any missing,
  unreadable, mismatched, or out-of-root artifact is a fatal evidence-document
  error, not an attempt exclusion.

**Admission state machine**

- **R7:** Evaluate gates in this order: document integrity and cohort
  completeness; fixture-family independence; frozen plan sufficiency; oracle
  qualification/isolation; request/preflight schema; route readiness; dispatch
  identity; model-output schema; proof availability/binding; proof outcome.
  Report the first failing gate plus any independently established secondary
  causes.
- **R8:** Admit an attempt after the model produced a response under a valid
  one-attempt execution identity and a qualified, unchanged, executable proof
  is bound to it. Proof success is not an admission prerequisite.
- **R9:** Count `model-failure` when an admitted response violates the declared
  model-output schema or fails unchanged deterministic proof. A malformed
  request/evidence package detected before dispatch is not model evidence.
- **R10:** Exclude non-model outcomes as `plan-insufficient`,
  `oracle-invalid`, `preflight-schema`, `route-infrastructure-not-run`,
  `execution-failed-no-response`, `execution-indeterminate`, or
  `proof-infrastructure`. Exclusions never count as model passes or failures.
  An uncertain dispatch is not retried under the same attempt/fixture identity.
- **R11:** Oracle qualification is independent of the evaluated model and
  frozen before execution. Its receipt includes version/hash, declared
  invariant coverage, positive/negative controls, isolation attestation,
  execution command, and unchanged post-run hash. Isolation evidence is emitted
  by the sandbox owner, not the model/executor, and proves the executor package
  could not read protected-oracle paths.
- **R11a:** The executor and proof runner use disposable bounded sandboxes with
  explicit filesystem allowlists, no inherited route credentials, denied-
  by-default network access, process/time/memory/output limits, and cleanup of
  model-created artifacts. Protected tests/solutions are absent from the
  executor sandbox and available only to the separately authorized proof
  runner. Missing verifiable containment evidence yields `oracle-invalid`.

**Qualification-plan sufficiency**

- **R12:** Plans intended as model-qualification evidence must set an explicit
  qualification-evidence contract. Ordinary KB plans do not become invalid
  merely because they omit this stricter evidence record.
- **R13:** The qualification record enumerates stable invariant IDs. Every
  nontrivial invariant must include either:
  a repository mechanism/hazard hint that makes the invariant actionable, or
  explicit uncertainty plus a target-tier raise that prevents the attempt from
  entering the lower-tier cohort.
- **R14:** Each mechanism/hazard entry names the invariant, repository
  file/symbol/interface or observed hazard, expected executor action, and proof
  target. Restating an acceptance criterion or adding generic hazard prose is
  insufficient. Adversarial paraphrase fixtures and known weak plans must fail
  the validator.
- **R15:** `document-review` remains the judgment-oriented plan-quality
  reviewer. `kb-plan` emits the structured record, and `kbcheck` is the
  mechanical authority that validates completeness and hashes it before
  execution. The classifier only consumes the resulting receipt. Do not add a
  separate DDR planner.

**Tier decision**

- **R16:** The first implementation supports a generic evidence schema but
  ships only the Medium policy. Higher-tier policy and qualification remain
  parked and cannot reuse a Medium result.
- **R17:** Define Medium independently of Deepseek4: ordinary bounded
  implementation with settled intent, clear files/interfaces, normal risk,
  deterministic acceptance/proof, and no architecture, security/destructive,
  broad-authority, or unresolved product judgment. Fixtures must require
  repository-aware implementation and failure diagnosis beyond mechanical
  transformation.
- **R18:** Freeze the initial Medium threshold before consuming Deepseek4
  outcome evidence: at least 30 admitted unique fixtures, at least five
  materially independent fixture families, no family contributing more than
  20% of the denominator, zero admitted model failures, and at least one
  holdout family not used to author or tune the planner, scorer, or threshold.
  Thirty independent zero-failure observations put the one-sided 95% binomial
  upper bound below 10%; the result reports this assumption and does not claim
  independence where the family record shows correlation.
- **R18a:** A preregistered family record gives every fixture one family and
  normalized fingerprints for task structure, repository mechanism, oracle
  type/coverage, and primary failure mode. Two families count independently
  only when they differ materially in at least three dimensions. The cohort
  authority signs this record before execution; holdout designation is part of
  the same signed record.
- **R19:** Return one bounded result: `qualified`, `not-qualified`, or
  `inconclusive`, plus the exact admitted denominator, pass/failure counts,
  exclusions, fixture families, and decision reasons.
- **R20:** Decision table: malformed/incomplete evidence is a command error;
  any admitted model failure is `not-qualified`; zero or insufficient admitted
  samples/families/holdout coverage is `inconclusive`; a complete cohort that
  meets R18 with only admitted passes is `qualified`. Exclusions cannot improve
  a result and must remain visible.
- **R21:** A qualified result is scoped to the tested task families, tools,
  context/risk envelope, route revision, and evidence date. It is not a claim
  that the model can perform every task at that tier.
- **R22:** The result is an evidence artifact, not automatic routing
  promotion. Any future consumer must match the current task family, tools,
  context/risk envelope, execution fingerprint, and freshness policy exactly;
  otherwise the qualification cannot satisfy route eligibility.
- **R23:** Evidence becomes stale when a bound model/route, system instruction,
  tool, context/risk policy, fixture/oracle, planner, scorer, or threshold
  revision changes. Stale evidence may be historical input but cannot produce a
  current `qualified` result and returns `inconclusive`. When the provider model
  revision is unobservable, qualification expires after 30 days and requires a
  newly preregistered sentinel cohort before current route eligibility could be
  considered.

**Ownership and safety**

- **R24:** The portable skill bundle owns the classifier schema, command,
  deterministic fixtures, and planning contract. TokenZoom owns live DDR/AMR
  execution evidence.
- **R25:** The portable scorer accepts only a redacted allowlist: hashes,
  stable opaque IDs, bounded enums/counts/timestamps, repository revisions, and
  normalized receipts. It rejects endpoints, credentials, secrets, response
  bodies, prompts containing private data, host paths/state, protected tests,
  oracle contents, known solutions, and reconstructable derivatives in normal
  fields, logs, errors, temporary files, or output.
- **R26:** Live evaluation is explicit, local/no-paid, operator-authorized, and
  separate from default contributor gates. Credentials stay in producer-owned
  approved stores, never command arguments or evidence, and are least-privilege
  for the expected route. The portable scorer never accesses them.
- **R27:** Reuse trustworthy existing evidence before making new model calls.
  If additional evidence is necessary, use one target-tier attempt per
  fixture and hard-stop on the first model/schema/proof failure.
- **R28:** After a non-model hard stop, preserve completed identities and
  receipts. The operator may fix the owning gate and continue untouched
  preregistered fixtures, but may not rerun an attempted fixture unless the
  original attempt is formally invalidated and excluded without replacing its
  record. An admitted model failure ends qualification as `not-qualified`.
- **R29:** The portable classifier/fixtures are an independently complete
  experimental milestone. TokenZoom's Deepseek4 export and decision are a
  downstream integration milestone; unavailable or insufficient evidence
  yields `inconclusive` and does not block experimental portable delivery.
  The classifier cannot be promoted as a supported reusable capability until
  one real cohort produces a non-inconclusive result and a second model or
  consumer demonstrates the abstraction generalizes.
- **R30:** Preregistration provenance is independently verifiable: the cohort
  manifest/family record must be in a trusted signed Git commit/tag, trusted
  timestamped signature, or configured append-only authority before the first
  attempt timestamp. Hashes or producer timestamps alone are insufficient.
- **R31:** A confirmed post-dispatch provider/transport failure with no model
  response is excluded as `execution-failed-no-response`; uncertain dispatch is
  `execution-indeterminate`. Neither is model evidence, and neither may retry
  the same fixture identity.

## Scope Boundaries

- No static global roster mapping named models to tiers.
- No automatic production routing promotion from one classifier result.
- No AMR comparison requirement for basic tier qualification.
- No paid model calls.
- No new DDR-specific planning skill.
- No inference benchmark implementation in this repository.
- No claim that a hash alone proves route identity or cohort completeness.

## Success Criteria

- Deterministic fixtures prove pass, output-schema/proof model failures, every
  exclusion, fatal evidence errors, omitted/replayed attempts, correlated
  families, insufficient coverage, stale/mixed fingerprints, and scoped
  qualification.
- `kb-plan` and `kbcheck` mechanically reject qualification-evidence
  plans that lack invariant-level mechanism/hazard or tier-raise records.
- A TokenZoom evidence export can be validated and classified without trusting
  self-authored narrative conclusions; unsupported authority claims force
  `inconclusive`.
- Deepseek4 receives a bounded Medium result from the same generic command used
  for any future model, or `inconclusive` when the preregistered evidence does
  not meet the independently frozen threshold.

## Evidence and Decisions

- Existing `evals/model-routing/initial-pilot.json` tests route behavior, not
  model tier fitness.
- Existing `model-routing-release` validates release claims, but has no
  per-model tier-classification command.
- Existing `kb-plan` has plan-wide specialist review and residual constraint
  propagation. The missing rule is invariant-level actionability.
- Prior cohort evidence: Deepseek4 passed asks 1-5; its delayed fifth attempt
  was infrastructure-gated before it passed. An apparent Luna failure was
  invalidated by an unstated protected oracle. A richer plan changed another
  Luna result, proving plan detail can affect executor outcomes.
- Decision: extend generic planning and eval surfaces instead of reviving the
  parked `kb-ddr-plan` proposal.
- Initial Medium threshold: 30 admitted unique fixtures across five independent
  families, max 20% per family, one holdout family, and zero admitted model
  failures. This bounds the independent-trial one-sided 95% failure-rate upper
  limit below 10% and was frozen before importing a Deepseek4 qualification
  export.

External research skipped: local benchmark evidence and repository contracts
fully determine the decision surface.

## Question Gate

- `ask-now`: none.
- `research-first`: none.
- `safe-assumption`: a strict Medium-only policy can sit on a generic versioned
  schema; adding future tier policies is additive and deterministic tests plus
  `inconclusive` catch unsupported targets.
- `defer-to-planning`: choose the smallest representation that reuses the
  existing strict `model-routing-release` JSON/path helpers without coupling
  release promotion to tier classification.
- `parked`: automatic router promotion and higher-tier classification.

## Resolve Before Planning

None.

## Slice Candidates

1. Validate plan sufficiency at the invariant level.
2. Add the strict tier-qualification evidence schema and deterministic scorer.
3. Add classifier fixtures and repository documentation.
4. Consume the first TokenZoom Deepseek4 evidence export and record the bounded
   Medium decision.
