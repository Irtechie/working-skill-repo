# DDR Cost Routing Benchmark

Checked: 2026-07-27
Budget mode: deep

## Question

Does Difficulty-Driven Routing produce portable plans that classify small work
as small, leave concrete model selection to the executing CLI, and reduce cost
per accepted result without weakening proof?

## Findings

1. **The plan contract is model-agnostic.** `kb-plan` records only the minimum
   `small|medium|large` capability, a falsifiable reason, requirements, proof,
   and escalation triggers. It explicitly forbids model names, route aliases,
   providers, adapters, endpoints, transports, and production `attempt_tier`.
   `kb-work` chooses `current` or `delegated` at pickup from live host evidence.
2. **All six hosted parents produced the expected portable plan.** Sol, Terra,
   Luna, Opus, Sonnet, and Haiku each assigned `small, small, medium, large`,
   used delegated ownership for every slice, preserved proof, and stored no
   concrete route. Planner-contract pass rate: **6/6**.
3. **Runtime delegation is not yet equally reliable.** Only Luna, Opus, and
   Sonnet passed the strict end-to-end contract. Sol and Terra stopped after the
   requested `rubber-duck` child type proved unavailable. Haiku completed two
   children but violated the machine-output contract and invented placeholder
   child IDs. Strict pass rate: **3/6**.
4. **Measured cost contradicts "largest is safest."** The six-parent matrix used
   **392.432422 AI credits = $3.924324**. Sonnet passed strictly at **$0.312787
   parent cost**, 45.88% below Opus at **$0.577971**. Sol spent **$1.928412**
   parent cost and still failed strict delegation. Terra produced the correct
   plan for **$0.186677** but failed strict delegation, so it has no valid
   cost-per-accepted-result claim.
5. **Cheaper tokens alone are not enough.** The Luna and Haiku model families
   accumulated child work as well as parent work. Haiku's failed arm required
   many retries/calls, making its pooled family cost higher than Luna's. The
   economic objective must remain **dollars per accepted result**, with tokens
   used only to diagnose why cost moved.
6. **The local route map exposed a readiness split.** LLMCommune is
   a reservation/controller surface that returns direct OpenAI-family call
   targets; it is not an inference proxy, chain executor, receipt issuer, or
   energy telemetry source. FleetController later proved both intended reservations
   coexist on disjoint hardware, but readiness diverged: DeepSeek V4 Flash was
   active with 1/1 ready/available replicas after 536.578 seconds; Qwen 3.6
   remained `starting` with zero observed/ready/available replicas and zero pods
   at the 2,100-second warm-up deadline. Qwen inference therefore did not run.
   This is an infrastructure-readiness result, not a Qwen model failure, and it
   is excluded from model pass/fail and cost-per-accepted-result denominators.
   Reservation coexistence is not route readiness.
7. **DeepSeek then passed the bounded local contract exactly.** `Deepseek4`
   returned HTTP 200 in **7.626 seconds** with **452 total tokens** and the exact
   `small, small, medium, large` tiers, delegated ownership, checksum 7, and no
   model pin. One retry was counted because the first successful inference used
   an unsupported post-call touch event before a receipt could be emitted. The
   rerun used the controller-supported `touch` event. Qwen did not audit the
   response because infrastructure never reached readiness; the same audit must
   be rerun after a green readiness preflight. The provider-billed inference
   charge is **$0**. No energy sensor was available, so energy cost and total
   operating dollars per accepted result remain unavailable.
8. **Authoritative cleanup caught a success-shaped release failure.** Qwen
   released normally. DeepSeek's first FleetClient release timed out and returned
   a success-shaped fallback while the controller record remained active. An
   authoritative state read rejected that result; an idempotent bounded
   controller retry released DeepSeek at 2026-07-27T05:14:22Z. Both records were
   verified released. The final health snapshot showed zero jobs and zero
   reservations with all controller surfaces green. Release command output alone
   is not cleanup proof.
9. **The first real-code pilot failed DDR and passed only the fallback path.**
   Sol produced a strict tier-only plan for one small and one medium slice, but
   neither assigned tier completed its preregistered plan unaided. Luna and
   Terra artifacts failed unchanged protected tests; Sol later repaired both.
   Eleven planner/child calls, including rejected and setup-defect attempts,
   cost **107.4937525 AI credits = $1.074937525**, or **$0.5374687625 per
   fallback-accepted slice**. This is not DDR success: the plans were too thin
   to isolate plan insufficiency from tier underestimation, and no matched
   all-large baseline ran.
10. **The guide-compliant fail-fast pair improved one slice and stopped on the
    next.** Sol produced frozen, code-free execution contracts after the
    `kb-ddr-plan` change. Luna passed the unchanged Retry-After oracle at the
    selected small tier. Terra then failed the unchanged cache-key oracle at the
    selected medium tier because repeated query values remained unsorted,
    despite the plan explicitly requiring value sorting. The hard-stop policy
    prevented every later model call. Six planner/executor calls cost
    **35.6673775 AI credits = $0.356673775**; planning represented **35.25%** of
    that spend. The final planner response also exceeded the combined declared
    plan budget by **308 output tokens**, proving that plan-token ceilings need
    mechanical enforcement rather than self-report.

### Hosted matrix

| Parent | Plan | Strict run | Parent/family cost | Main failure |
|---|---:|---:|---:|---|
| GPT-5.6 Sol | pass | fail | $1.928412 parent | unavailable child type was not replaced |
| GPT-5.6 Terra | pass | fail | $0.186677 parent | unavailable child type was not replaced |
| GPT-5.6 Luna | pass | pass | $0.245429 pooled family | child IDs not exposed |
| Claude Opus 5 | pass | pass | $0.577971 parent | child IDs not exposed |
| Claude Sonnet 5 | pass | pass | $0.312787 parent | child IDs not exposed |
| Claude Haiku 4.5 | pass | fail | $0.673049 pooled family | malformed envelope and invented IDs |

The full matrix cost per strict pass was **$1.308108**. This is not a production
forecast; it is the honest cost of the controlled experiment, including failed
arms.

### First real-code DDR failure / fallback pilot

| Slice | Planned tier | DDR result | Fallback result |
|---|---:|---|---|
| Retry-After parser | small | failed: assigned tier did not pass | Terra also failed; Sol accepted |
| Canonical cache key | medium | failed: assigned tier did not pass | Sol accepted |

The Sol plan stored tiers, reasons, proof, escalation triggers, and delegated
ownership without naming a model, but it did not provide an execution-grade
contract. Opus chose the same tiers but failed the strict schema by adding an
unknown field, so that planner call remains in the numerator. Protected Go
tests and module files retained their recorded SHA-256 values. The rerun must
freeze guide-compliant plans, count planner tokens/cost, and diagnose same-tier
plan insufficiency separately from next-tier capability.

Machine-readable evidence:
`docs/results/2026-07-27-ddr-real-execution.json`.

### Guide-compliant conformance pair

| Slice | Assigned tier | Result | Interpretation |
|---|---:|---|---|
| Retry-After parser | small | pass | The richer execution contract was sufficient for the selected tier. |
| Canonical cache key | medium | fail | Terra omitted an explicitly required value sort; tier versus stochastic failure remains unresolved. |

This was intentionally not repaired or escalated. No direct baseline, Claude
arm, diagnostic arm, or remaining cohort call ran after the failure. The
result therefore supports neither a savings claim nor a tier-policy change.
It does establish that planner cost is material and must remain in the
numerator.

Machine-readable evidence:
`docs/results/2026-07-27-ddr-guide-conformance-pair.json`.

### What is actually proved

- Static contract: `cmd/kbcheck/ddr_contract_test.go` mechanically rejects model
  pins in plans, automatic lower-tier AMR, missing delegated-default ownership,
  duplicate route announcements, and lower-tier fallback.
- Live planner behavior: all six parents followed the portable tier contract on
  the same four-slice prompt.
- Live orchestration behavior: three parents recovered from the unavailable
  child persona and completed two bounded small-model children; three did not.
- Cost: exact AI credits were read from
  `assistant_usage_events.total_nano_aiu`, not estimated from token counts.
- Real work: two code slices proved deterministic rejection and higher-tier
  fallback, while failing the DDR requirement that the assigned tier execute
  the preregistered plan unaided.

### Local DeepSeek and Qwen economics

DeepSeek and Qwen are home models. Their provider-billed inference charge is
currently **$0** and their token usage is recorded when returned. That is not
the same as zero operating cost. Use:

```text
local_operating_cost =
  measured_wall_energy_kwh * electricity_rate_usd_per_kwh
  + allocated_idle_cooling_and_controller_overhead
  + hosted_orchestrator_cost

hardware_amortization = separate sensitivity view
```

The user's current operating estimate is roughly **$1 per day of inference**.
That implies about **$0.0417 per running hour** and **$0.0035 for five minutes**.
This is a scenario, not measured evidence. The local arm recorded accepted
DeepSeek output and **$0 provider-billed cost**, but no wall energy. Energy cost,
total operating cost, and total cost per accepted result remain unavailable.

### Local route-readiness result

| Question | Evidence-backed answer |
|---|---|
| Does LLMCommune proxy inference? | No; the caller invokes `call_target` directly. |
| Does it execute hosted-main -> DeepSeek -> Qwen? | No; orchestration and receipts must be caller-owned. |
| Can intended DeepSeek and Qwen 3.6 reservations coexist? | Yes; FleetController admitted DeepSeek on a two-member GB10 gang and Qwen 3.6 on a separate RTX 5090. |
| Were both routes inference-ready? | No; DeepSeek was 1/1 ready, while Qwen had zero replicas and pods at its 2,100-second deadline. |
| Is LLMCommune's Qwen235 entry relevant? | No, unless FleetController live inventory explicitly exposes it for this run. |
| Where did live proof run? | In the FleetController session against the reserved LiteLLM route; no endpoint or credential was checked in. |
| Did DeepSeek produce an accepted result? | Yes; exact JSON, HTTP 200, 7.626s, 452 tokens, deterministic oracle pass. |
| Did Qwen audit it? | No; test status `not-run` because infrastructure was not ready. No model conclusion is allowed. |
| Is local cost measured? | Provider-billed inference cost is $0; runtime/tokens exist; energy and total operating cost are unavailable. |
| Were reservations cleaned up? | Yes; both controller records were verified released. |

The route map still changes the implementation design:

1. OpenAI-compatible/LiteLLM is only the inference wire.
2. KB needs a bounded tool-free child harness plus its existing route
   fingerprint, receipt, deterministic proof, and dispatcher-attestation path.
3. Automatic qualification must require one route-bound `EvidenceKBReceipt`
   issued after attended calibration verifies the installed harness revision,
   release fixtures, and repeated behavioral evidence; `/v1/models` is presence
   only.
4. FleetController reservation state and route readiness are separate gates.
   Coexisting reservations do not authorize inference until ready replicas and
   served-model presence are observed for each route.
5. Infrastructure preflight must pass before a model enters the benchmark
   denominator. Setup/readiness failures are tracked separately and rerun after
   repair; they are not correctness failures.

Machine-readable evidence and exact response:
`docs/results/2026-07-27-ddr-local-route-readiness.json`.

## Sources

- GitHub Copilot model pricing and AI credit conversion:
  https://docs.github.com/en/copilot/reference/copilot-billing/models-and-pricing
- Planner contract: `.github/skills/kb-plan/SKILL.md`
- Work-time owner and route contract: `.github/skills/kb-work/SKILL.md`
- Mechanical DDR contract tests: `cmd/kbcheck/ddr_contract_test.go`
- Hosted raw result:
  `docs/results/2026-07-27-ddr-hosted-model-matrix.json`
- DDR failure / fallback pilot result:
  `docs/results/2026-07-27-ddr-real-execution.json`
- Guide-compliant fail-fast pair:
  `docs/results/2026-07-27-ddr-guide-conformance-pair.json`
- Local route-readiness result:
  `docs/results/2026-07-27-ddr-local-route-readiness.json`
- Prior failed Qwen-to-Sol cost experiment:
  `docs/results/2026-07-13-qwen-sol-exploratory-cost.md`

## Applies When

- Deciding whether plans should name models or only capability tiers.
- Evaluating planner/orchestrator models for small-model delegation.
- Comparing hosted and home inference economics.
- Presenting DDR claims to a skeptical technical or financial audience.

## Stale When

- The host exposes stable child-agent receipts for foreground task calls.
- The child-agent persona schema changes.
- Copilot AI credit conversion or model pricing changes.
- Repeated DeepSeek accepted implementation results include measured wall power;
  Qwen 3.6 reaches ready replicas and served-model presence in a later
  reservation.

## Rejected Approaches

- **Tokens saved as the primary KPI:** rejected because home tokens have near-zero
  marginal price and failed cheap attempts can increase total cost.
- **Cheapest parent output wins:** rejected because Terra's inexpensive correct
  plan did not complete the required delegation contract.
- **Dry-run routing as live proof:** rejected because it cannot prove dispatch,
  output shape, correctness, or billed cost.
- **Putting model names in plans:** rejected because host catalogs and local
  availability change; the executing CLI owns route selection.
- **Ignoring failed arms:** rejected because cost per accepted result must include
  the spend of failures and retries.
- **Calling controller discovery a local-model pass:** rejected because a route
  map does not prove inference, attribution, output acceptance, or energy.
- **Using stale Qwen235 metadata for the intended route:** rejected because the
  requested Qwen 3.6 runs on a separate RTX 5090; FleetController live inventory
  is the only authority for current concurrent availability.
- **Treating reservation admission as readiness:** rejected because the live
  Qwen reservation coexisted with DeepSeek but had zero replicas and pods at its
  warm-up deadline.
- **Scoring infrastructure failure as model failure:** rejected because Qwen
  never received the test. Its model outcome is null until a readiness-green
  rerun.
- **Trusting a success-shaped release fallback:** rejected because the first
  DeepSeek helper result contradicted authoritative controller state.

## Impact On Current Project

- Keep `kb-plan` model-agnostic; no skill change is required for concrete route
  portability.
- Add `evals/cross-model-benchmarks/ddr-portable-plans.json` to preserve the
  plan-tier and no-model-pin contract.
- Repair child-agent persona discovery so planners select from the live callable
  schema instead of assuming `rubber-duck`.
- Keep local connections optional and user-local. A tracked `LOCAL_MODELS.md`
  may explain setup but cannot provide trust or capability evidence.
- Add a bounded direct inference harness only by extending existing
  `RoutingReceipt` and dispatcher-attestation semantics; do not treat
  LLMCommune as an agent runtime.
- Treat the hosted result as an initial controlled sample, not a promotion
  decision. Repeat with multiple implementation fixtures and exact child route
  receipts before claiming production savings.
