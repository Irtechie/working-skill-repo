---
date: 2026-07-11
topic: ghcp-aic-amr-falsification
brainstorm_style: kb-brainstorm
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-brainstorm
    status: passed
    required_evidence:
      - docs/brainstorms/2026-07-11-ghcp-aic-amr-falsification-requirements.md
      - Question Gate classification completed
      - Resolve Before Planning is empty
      - claims are separated into verified, provisional, and parked
      - document review has no unresolved P0/P1 findings
    proof:
      - "local GHCP 1.0.70-0 OTel schema probe and official GitHub Copilot CLI OTel documentation"
      - "invalid live canary: requested claude-haiku-4.5, observed claude-sonnet-5, 7.21096 AIU, proof failed"
      - "repo-critic: consolidate incoming amrbench; protected-oracle P0 and grader/accounting P1 findings incorporated"
      - "document-review pass 1: security P0 plus evidence P1s and coherence experimental-design P1s resolved"
      - "document-review pass 2: security and coherence clear; no unresolved P0/P1"
    blockers: []
    passed_at: "2026-07-11T06:37:11-04:00"
    allowed_next_action: "kb-plan docs/brainstorms/2026-07-11-ghcp-aic-amr-falsification-requirements.md"
---

# GHCP AIC and AMR Falsification

## Problem Frame

AMR is valuable only when a bounded worker path reaches near-driver correctness
at lower all-inclusive cost. GitHub Copilot CLI now provides enough per-call
OpenTelemetry evidence to measure that claim, but the benchmark must not confuse
requested model names, aggregate agent spans, prompt bloat, or weakened tests
with real savings.

The current production AMR runtime remains conservative: passing bounded
attempts may be retained, but automatic correction dispatch is fail-closed
until an isolated workspace, host-owned proof, and compare-and-swap apply path
exist. This benchmark may evaluate an isolated correction prototype; it must
not advertise that prototype as shipped KB behavior.

## Evidence Checked

Verified from the local GHCP 1.0.70-0 probe and official GitHub Copilot CLI
OpenTelemetry documentation:

- one `chat` span represents one LLM request and includes requested/resolved
  model, token counts, AI units, trace/session/turn identifiers, duration, and
  parent lineage;
- `invoke_agent` contains aggregate usage, so summing it with child `chat`
  spans double-counts cost;
- the legacy probe emits `github.copilot.nano_aiu`; dividing the integer by
  1,000,000,000 yields AI units. Current documentation also exposes
  `github.copilot.aiu`, so the adapter must be versioned and support both;
- the original probe used 2.223925 AI units for one small reply and 17,768 input
  tokens, showing a serious context-cost signal.

Verified from the first incoming `amrbench` live canary:

- the harness requested `claude-haiku-4.5`, but all three leaf calls resolved to
  `claude-sonnet-5`;
- the invalid route sample spent 7.21096 AI units, 69,665 input tokens, and 447
  output tokens, made no accepted edit, and failed deterministic proof;
- requested alias is therefore not model provenance, and no further paid matrix
  run is allowed until exact mismatch handling and test protection pass.

Provisional:

- loaded skills/tools are likely a major cause of the high input count, but the
  trace does not isolate causality. A controlled same-model, same-task context
  A/B is required;
- load-aware routing can use observed call latency, but queue/load is unknown
  unless a route exposes trusted health evidence.

Parked:

- DSpark-style logits, token-level acceptance, KV-cache sharing, Markov or
  rejection-sampling guarantees, and inference-latency claims;
- production partial-reuse correction until the isolated CAS runner exists.

## Terms

- **AIU/AIC:** versioned GHCP accounting unit. Raw integer nano-AIU or raw
  documented AIU is authoritative; display conversion is derived.
- **Leaf call:** a unique `chat` span, deduplicated by `(trace_id, span_id)`.
- **Span identity:** `(trace_id, span_id)`; a bare span ID is never globally
  unique enough for accounting.
- **Arm total:** every unique leaf call causally bound to one direct or AMR arm,
  including driver, worker, review, repair, fallback, proof-driven follow-up,
  and compaction calls.
- **Route match:** every observed leaf call resolves to an allowed exact model
  identity for the requested observable route tuple. Provider-reported model
  names are observations, not immutable model attestation.
- **Planned-tier edit churn:** trusted phase-baseline delta attributed to the
  planned-tier phase. It replaces the unsupported phrase "frontier-authored
  lines" and is a secondary diagnostic.

## Requirements

### Accounting adapter

- R1. Parse bounded OTel JSONL strictly from one exclusively created export file
  per process/phase. Reject malformed rows, oversized input,
  duplicate span identities with conflicting content, counter overflow, symlinks,
  hardlinks, and schema conflicts instead of silently skipping them.
- R2. Identify spans by `(trace_id, span_id)` and sum unique leaf `chat` spans
  only. Use `invoke_agent` solely for hierarchy and reconciliation; never add
  its aggregate usage to leaf totals. Reject missing identities, conflicting
  duplicates, cycles, ambiguous/orphan parents, cross-phase contamination, and
  any leaf not bound to exactly one phase and arm.
- R3. Accept a documented direct `github.copilot.aiu` value or legacy
  non-negative integer `github.copilot.nano_aiu`. Preserve exact raw values and
  adapter/runtime revision. Reject conflicting simultaneous fields.
- R4. Normalize model, input/output/cache token counts, trace, span, parent,
  conversation, turn, interaction, duration, status, and error fields. Missing
  evidence is `unavailable`, never zero or inferred.
- R5. Every credited leaf must contain a recognized cost field. After process
  exit, require export flush/stability, expected-root binding, and
  adapter-versioned aggregate reconciliation. Missing or unreconciled spans make
  phase cost `unavailable` and the sample promotion-ineligible; observed cost is
  retained only as a lower-bound waste record.
- R6. A sample with missing or mismatched actual model identity is invalid for
  model qualification and correctness comparison. Its spent cost remains in an
  operational waste report keyed by observed route/model, not the requested
  alias; it cannot become another model's credited sample or a pass/fail sample.
- R7. Bind qualification to a versioned observable route tuple: runner, adapter
  and GHCP revision, hosting/provider origin, redacted profile fingerprint,
  requested model, response model, request/response identifiers, and model
  artifact digest when the provider exposes one. Without immutable artifact or
  revision evidence, report `observed-route-match`, never exact-model proof.

### Trust and isolation

- R8. Raw traces, prompts, stdout/stderr, workspaces, and local profiles remain
  under ignored `.kb/` or user-local `~/.kb/` storage with content capture off.
  Create owner-only directories/files before capture. Launch children from a
  minimal allowlisted environment, inject only the selected route secret, clear
  inherited provider/exporter/content/plugin/custom-instruction variables, and
  reject content-bearing telemetry such as prompts, messages, tool arguments,
  tool results, authorization data, usernames, or absolute host paths. Committed
  evidence is redacted, normalized, bounded, and hash-linked only.
- R9. Each arm uses a fresh fixture workspace and immutable baseline.
  Fixture paths resolve under the benchmark fixture root; output resolves under
  `.kb/`; symlink/special-file escapes fail closed.
- R10. Directory separation alone is not isolation. The model phase uses
  mediated writes limited to an explicit mutable allowlist, read-only proof
  inputs, provider-only network access, a minimal environment, and process/time/
  memory limits. The proof phase runs generated code with no network or host
  filesystem access, a minimal environment, immutable proof inputs, and process/
  time/memory containment. If either sandbox is unsupported or fails, record
  `invalid-isolation`, prohibit the paid phase, and never qualify the route.
- R11. Protect the complete proof dependency closure before the model: tests,
  specifications, goldens, fixtures, scripts, helper packages, module/package
  manifests, lockfiles, harness revision, verification argv, controlled
  environment contract, and toolchain identity. Use an explicit mutable and
  new-file allowlist; reject added test/oracle files, `TestMain`, dependency
  redirects, or any proof-input mutation even when tests exit zero.
- R12. The harness, not the model, runs focused and final regression proof.
  Repair never skips final regression proof.
- R13. Phase roles are bound through a trusted run envelope and parent/trace
  lineage. Do not infer worker/review/repair role from model name.
- R14. Route/profile details and credentials remain user-local. Tracked config
  may define abstract cohorts and known-answer tasks, not a volatile hosted
  model ladder or private endpoint.
- R15. Every fixture has a hash-linked qualification preflight: immutable
  baseline fails the intended criterion, a separately stored trusted solution
  passes, declared negative mutations fail, and the proof exercises the intended
  behavior. Requalify whenever fixture, oracle, harness, verification argv,
  environment, or toolchain revision changes. HTML/UI requires rendered DOM,
  interaction, focus/keyboard, accessibility, and console assertions; substring
  matching is insufficient.

### Experimental design

- R16. Run a context-diet A/B before routing claims on a context-development
  corpus disjoint from the held-out routing corpus: same observed route, task,
  fixture, proof, order control, and route revision; vary only ambient
  skill/tool/context payload. Predeclare and hash both contracts and the winner
  rule. Reduced context wins only with non-inferior correctness and lower paired
  AIU; otherwise retain the baseline. Freeze the result before routing trials.
- R17. Paired direct and AMR arms use fresh workspaces/sessions, randomized or
  crossover order, the same known-answer task, and the same protected proof.
  Both receive an identical hashed base task packet plus preregistered hashed
  role overlays. Freeze tools, proof, timeout, retry, stopping, and terminal
  criteria. The direct arm may self-repair to the same terminal criterion; every
  orchestration call counts.
- R18. Planner assigns task family and difficulty before route identities or
  outcomes are visible. Direct uses the exact planned-tier observed route;
  experimental AMR gets exactly one next-lower attempt and planned-tier-or-higher
  full fallback. Freeze cohort routes; disable load-aware switching, online
  calibration, and post-hoc tier changes during held-out trials.
- R19. Direct cost is the all-leaf arm total. AMR cost includes every leaf call
  across attempt, driver/review, repair, fallback, and related orchestration.
  Role subtotals must reconcile to the arm total.
- R20. Primary promotion metrics are correctness and all-inclusive AIU. Latency
  is secondary unless end-to-end comparable. Planned-tier edit churn and
  independently accepted worker hunks surviving final output are diagnostics,
  not correctness substitutes.
- R21. Promotion remains bound to the exact adapter/runtime/policy/proof harness
  and exact observable route-tuple revisions. Missing telemetry, model mismatch, insufficient
  samples, or incomparable context produces `not-promoted`.
- R22. Use the existing promotion boundary per eligible task family: zero
  observed right-to-wrong
  regressions; a one-sided 95% bound excluding more than 2 percentage points of
  correctness regression; no protected-oracle or human-intervention increase;
  at least 20% paired median AIU improvement; and lower aggregate AIU than direct
  execution on a power-justified held-out corpus. Report mean, median, total,
  p90, fallback rate, and repair tail. Seeded faults and invalid routes do not
  count as efficiency wins.
- R23. Use outcome classes `right`, `wrong`, `invalid`, and `ungradable`.
  Human-intervention events are attended clarification, manual edit, safety
  approval beyond preregistered execution, route repair, or manual proof
  interpretation; report them per arm and family.
- R24. Task-family admission is versioned and preregistered. Eligibility requires
  bounded authority, stable specification, independent deterministic proof, and
  localizable failure. Speculative/philosophical/subjective work is unrunnable
  in both direct and AMR scoring. HTML/UI is eligible only from approved design
  inputs with deterministic browser/DOM/accessibility proof. Promotion and
  disablement are per family, never pooled across unlike work.

### Product boundary

- R25. GHCP remains a later conformance cohort until the accounting adapter,
  trusted collection, exact route match, and attended canary pass. Do not add it
  to the initial Codex-first supported cohort retroactively.
- R26. This milestone measures GHCP accounting and full-fallback economics, not
  production partial reuse. The benchmark correction arm is labeled experimental. It may
  measure `attempt + full planned-tier repair`, but may not claim production
  partial-reuse or preserved-hunk execution. An isolated-CAS partial-reuse arm is
  a separately planned future milestone.
- R27. Implement this as a separate follow-on GHCP conformance manifest that
  depends on the shipped slice-007 baseline and reuses the existing execution
  telemetry and promotion validator. It does not change or block active
  Codex-first slice 006.
- R28. No second paid live run starts until R1-R15 pass deterministic tests and
  the first canary is classified `invalid-route`, not `failed-direct`.
- R29. Paid execution is an attended state machine with a hash-linked conformance
  artifact; explicit per-call, per-arm, and experiment-wide call/credit ceilings;
  budget/gate recheck before every phase; and immediate stop on route mismatch,
  missing/incomplete telemetry, containment failure, or oracle mutation. An
  invalid attempt cannot launch correction/fallback. New approval is required
  before canary, context A/B, and paired-matrix stages.
- R30. Current `qualified`/`suspended` labels remain disabled until paired
  correctness/cost math and family-specific promotion gates pass.
- R31. Adaptive scope, queue-aware switching, and online policy updates remain
  parked during controlled trials. Any later calibration is keyed by family,
  observed route revision, proof harness, and context contract.

## Keep / Consolidate / Delete for Incoming `amrbench`

Keep, after proof:

- isolated known-answer fixtures, explicit subjective-task ineligibility,
  process-tree containment, leaf-chat deduplication, requested/actual model
  capture, deterministic harness-run proof, and user-local provider profiles.

Consolidate or rebuild:

- OTel normalization into the existing execution-telemetry contract;
- exact integer/decimal accounting and route mismatch as a first-class outcome;
- phase totals, context A/B, paired-arm grading, and existing AMR promotion
  thresholds;
- fixture/test protection and workspace containment with existing KB storage
  primitives where possible.

Delete or forbid:

- float-only AIC as authoritative data;
- prompt-only "do not edit tests" protection;
- current-model-name qualification after a mismatch;
- tracked volatile hosted model ladders;
- README claims that tools or protected oracles are enforced when code does not
  enforce them;
- any automatic paid model matrix before conformance gates pass.

## Question Gate

- `ask-now`: none.
- `research-first`: none; official docs, local probe, live invalid canary, and
  current code were inspected.
- `safe-assumption`: benchmark work remains experimental and spends no further
  AI credits until deterministic gates, attended approval, and sandbox support
  pass.
- `defer-to-planning`: exact file split and whether the adapter is a new
  internal package or an extension of `execution-telemetry`.
- `parked`: production isolated correction/CAS execution, a true partial-reuse
  benchmark arm, adaptive/queue-aware trial routing, and DSpark-only inference
  mechanics.

## Resolve Before Planning

None.

## Acceptance

The plan is ready only when independent document review finds no unresolved
P0/P1 issue and each slice has a deterministic proof target, risk, tier, bounded
context packet, and explicit paid-run gate.
