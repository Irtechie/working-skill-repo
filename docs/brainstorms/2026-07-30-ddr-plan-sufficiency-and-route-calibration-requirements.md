---
date: 2026-07-30
topic: ddr-plan-sufficiency-and-route-calibration
brainstorm_style: kb-brainstorm
---

# DDR Plan Sufficiency and Route Calibration

## Problem Frame

KB production DDR already delegates bounded slices by portable capability tier,
but its planning contract does not yet prove that the planner reduced residual
work enough for the selected tier. The current benchmark can observe a failed
medium execution and a successful large fallback, but it cannot distinguish an
insufficient plan from an underestimated tier. Concrete routes such as Terra
and DeepSeek 4 also must not be treated as equivalent merely because a static
table labels both `medium`.

The durable fix must combine skill judgment with deterministic validation and
controlled benchmark evidence. Prose-only guidance is insufficient.

## Research Summary

**Findings that shaped requirements:**

- Current `kb-plan` records `model_tier`, requirements, and escalation triggers,
  but defaults to `medium`, says to choose higher when unsure, and does not
  require a lower-tier feasibility pass after the execution contract is written
  - affects R1-R4 - `.github/skills/kb-plan/SKILL.md`.
- A parked `kb-ddr-plan` design already defines minimum-sufficient execution
  contracts, residual-work tiering, leakage limits, and controlled failure
  diagnosis, but it is absent from the repo-local skill bundle - affects R1-R3
  and R5 - `docs/handoffs/parked/2026-07-27-eddr-experimental-state.md`.
- `kbcheck` already validates manifests and context packets structurally, while
  its existing DDR test protects production ownership and route-announcement
  policy rather than plan sufficiency - affects R4 -
  `cmd/kbcheck/manifest_contract.go`, `cmd/kbcheck/context_packet.go`, and
  `cmd/kbcheck/ddr_contract_test.go`.
- `kbrouter models calibrate` is currently attended descriptive guidance and
  does not independently attest a route's tier by task family - affects R6 -
  `docs/context/architecture/kbrouter.md` and `cmd/kbrouter/main.go`.
- Latest controlled DDR evidence proves portable planning output but does not
  prove savings or distinguish plan insufficiency from tier underestimation;
  one accepted local DeepSeek planning response is not medium code-execution
  qualification - affects R5-R8 - latest TokenZoom DDR results summarized in
  the 2026-07-30 session.

External research skipped: local implementation and benchmark evidence directly
own the decision.

**Confidence:** High for the contract and harness gaps; low for any claim about
DeepSeek 4's execution tier until it runs controlled task-family fixtures.

## Requirements

**Plan quality and tier selection**

- R1. Add `kb-ddr-plan` to the repo-local bundle and require `kb-plan` to invoke
  its decision process for every non-trivial runnable slice.
- R2. Every DDR-planned slice must carry a minimum-sufficient execution
  contract covering objective, repository route, behavior, invariants and edge
  cases, implementation boundary, exact proof, and observable stop conditions.
- R3. The planner must choose the minimum capable tier from work remaining
  after the execution contract is written, not from the task title or a static
  model table. Security, data authority, irreversible risk, unresolved product
  decisions, and weak proof remain hard lower bounds.
- R4. Planning must not make a lower tier appear successful by embedding a
  patch, function body, fixture answer, hidden-oracle deduction, or
  line-by-line implementation recipe. Ordinary local implementation choices
  remain executor-owned.

**Deterministic enforcement**

- R5. Extend `kbcheck` so opted-in DDR manifests and context packets fail closed
  when required execution-contract fields, residual-work tier reasoning,
  bounded plan-token policy, leakage declaration, or preregistration evidence
  are absent or malformed. Deterministic checks may validate structure and
  forbidden content, but must not claim to prove semantic plan quality.

**Failure diagnosis and model qualification**

- R6. Extend the benchmark contract to freeze task inputs, plan, selected tier,
  and protected proof before execution, then classify controlled follow-ups as
  `plan-insufficient`, `tier-underestimated`, or `task-or-oracle-defect`.
  Stronger-model fallback must never rewrite a failed DDR attempt as success.
- R7. Qualify concrete routes independently by task family and evidence
  freshness. Terra and DeepSeek 4 must not inherit each other's `medium`
  classification, and an unqualified route must not be presented as proven for
  that tier.
- R8. Whole-route reporting must include planner and executor cost, failed
  attempts, assigned-tier acceptance, plan-token share, and a same-task direct
  large-tier baseline before claiming savings.

**Portable bundle and operator clarity**

- R9. Update the skill inventory, architecture, eval map, and testing guidance
  so maintainers can distinguish DDR planning, runtime route selection, route
  qualification, and experimental AMR.
- R10. Keep plans provider-neutral. Concrete model names, aliases, endpoints,
  and provider state remain in runtime catalogs and receipts, never in portable
  manifests.

## Success Criteria

- A representative runnable slice cannot pass `manifest-contract` when its DDR
  execution contract or residual-tier justification is missing.
- A valid slice can pass deterministic DDR contract checks without naming a
  provider or concrete model.
- The benchmark mechanically preserves the frozen original result and reports
  which controlled arm supports each failure classification.
- Route qualification evidence can distinguish Terra from DeepSeek 4 and can
  represent `unqualified` without silently promoting either route.
- No documentation or output claims DDR savings without an accepted assigned-
  tier result and an identical-task direct-large baseline.
- Repo-local contributor and release gates include the new deterministic
  contracts and pass without making live model calls.

## Scope Boundaries

- No automatic downward routing during ordinary `kb-work`.
- No automatic stronger-model fallback counted as DDR success; AMR remains a
  separate experiment.
- No provider or model pin in `kb-plan` artifacts.
- No paid or live-model cohort runs without separate attended authorization.
- No claim that deterministic validators can judge semantic plan sufficiency.
- No broad model leaderboard or universal tier; qualification is task-family
  specific and evidence-bound.
- No automatic process termination, fleet mutation, or route reservation.

## Key Decisions

- Extend `kb-plan` through a focused `kb-ddr-plan` skill rather than creating a
  parallel planning lane: DDR changes how runnable slices are prepared, not the
  user-facing workflow. Evidence: parked DDR planning design and current
  `kb-plan` integration contract.
- Add deterministic `kbcheck` enforcement rather than relying on prose:
  structural omissions and forbidden leakage are machine-checkable, while
  semantic quality remains a benchmark claim. Evidence:
  `cmd/kbcheck/manifest_contract.go` and `cmd/kbcheck/context_packet.go`.
- Separate plan diagnosis from route calibration: a planner decides residual
  capability; runtime evidence decides whether a concrete route satisfies that
  capability for a task family. Evidence:
  `docs/context/architecture/kbrouter.md`.
- Treat DeepSeek 4 as unqualified until controlled execution evidence exists:
  one accepted bounded planning response does not prove medium implementation
  capability. Evidence: latest local-route benchmark result.

## Dependencies / Assumptions

- [safe-assumption] The parked `kb-ddr-plan` is reusable prior art, not an
  authoritative final copy - Reversible because: implementation will diff and
  reconcile it against current `kb-plan` before adoption - Evidence/proof:
  skill contract tests and document review.
- [safe-assumption] Existing `manifest-contract` and `context-packet` commands
  are the narrowest enforcement surfaces - Reversible because: planning may
  introduce a dedicated DDR subcommand if extending those commands creates an
  ambiguous contract - Evidence/proof: focused Go tests must demonstrate the
  final command boundary.
- [defer-to-planning] Decide whether route qualification extends
  `models calibrate` or consumes a separate evidence artifact, preserving
  user-local runtime state and no-live-call defaults.

## Alternatives Considered

- Keep DDR guidance only inside `kb-plan`: rejected because the decision
  discipline is substantial, lazy-loadable, and already has a focused parked
  skill design.
- Automatically force every slice down one tier: rejected because authority,
  security, risk, and proof constraints are not planning verbosity problems.
- Label DeepSeek 4 medium from its model family or one planning response:
  rejected because route tiers must be earned per task family.
- Treat successful large fallback as DDR success: rejected because it prevents
  diagnosis and hides failed assigned-tier economics.

## Slice Candidates (advisory for /kb-plan)

- DDR-ready planning contract - planners produce bounded execution contracts
  and residual-work tier decisions for runnable slices.
- Deterministic plan guard - invalid or solution-leaking DDR artifacts fail
  before work dispatch.
- Controlled failure diagnosis - benchmark results distinguish inadequate plans
  from underestimated tiers without rewriting original outcomes.
- Evidence-bound route qualification - concrete routes expose task-family
  qualification or an honest unqualified state.
- Portable docs and release proof - the installed bundle explains and verifies
  the complete DDR path without live inference.

## Outstanding Questions

### Resolve Before Planning

None.

### Deferred to Planning

- [defer-to-planning][Affects R5][Technical] Extend the existing manifest and
  packet schemas or add one focused DDR artifact validator based on the
  smallest coherent Go boundary.
- [defer-to-planning][Affects R7][Technical] Reuse `models calibrate` or add a
  read-only qualification-evidence command after inspecting catalog ownership
  and current result schemas.

### Parked / Out of Scope

- [parked][Affects R8] Live paid cohort execution - Forbidden claim: deterministic
  harness completion proves provider savings.
- [parked][Affects R7] Universal model tier labels - Forbidden claim: one task
  family qualifies a route for all repository work.

## Next Steps

-> `/kb-plan docs/brainstorms/2026-07-30-ddr-plan-sufficiency-and-route-calibration-requirements.md`
