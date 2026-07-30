# DDR Planning Validation

Status: active
Created: 2026-07-30
Last refreshed: 2026-07-30

## Intent

Make DDR failures attributable instead of treating every failed run as evidence
that the assigned model is below its planned tier. Strengthen portable planning
so nontrivial invariants become executable guidance, then grade Deepseek4
against a bounded Medium cohort in the repository that owns the benchmark.

## Success Criteria

- A failed DDR run is classified as `model-failure`, `plan-insufficient`,
  `oracle-invalid`, `schema-gate`, or `route/infrastructure-not-run`.
- Portable planning requires each nontrivial invariant to include a
  mechanism/hazard hint or an explicit uncertainty-driven tier raise.
- `kb-plan`, `document-review`, manifest validation, and deterministic tests
  enforce one coherent plan-sufficiency contract.
- No separate `kb-ddr-plan` skill is added unless the generic planning lane
  cannot express or verify the contract.
- TokenZoom records a bounded Deepseek4 Medium qualification decision with an
  exact denominator, exclusions, artifact-integrity proof, and no paid calls.
- Model fitness is not inferred from plan, oracle, route, or infrastructure
  failures.

## Architecture Decisions

- `document-review` remains the plan-quality review owner and `kb-plan` remains
  the decomposition owner. DDR does not gain a second planner.
- The portable bundle owns planning and routing contracts. TokenZoom owns live
  DDR/AMR conformance and model qualification evidence.
- Planning quality and model quality are independent gates. A model result is
  admissible only after plan sufficiency, protected-oracle isolation, schema,
  route readiness, and deterministic proof all pass.
- Concrete model names do not become permanent portable tier assignments.

## Research

- Original benchmark evidence: Deepseek4 passed asks 1-5. Ask 5 was delayed by
  route infrastructure, then passed.
- Luna's apparent ask-5 failure was invalidated by an unstated protected-test
  requirement; unchanged candidates passed after the oracle was corrected.
- A richer frozen plan made Luna pass another fixture. Terra still failed, but
  the plan was judged insufficient because it restated the sorting requirement
  without identifying the repository mechanism or hazard.
- The plan-wide specialist-review contract landed in commit `14d41f3` after the
  benchmark failures. The remaining gap is explicit plan-sufficiency evidence,
  not another general review pipeline.
- TokenZoom PR #2 owns the portable DDR parent-failover conformance harness.

## Workstreams

| Workstream | Brainstorm | Manifest | Status | Notes |
|---|---|---|---|---|
| Portable plan-sufficiency and tier classifier | `docs/brainstorms/2026-07-30-model-tier-qualification-requirements.md` | `docs/plans/2026-07-30-010-kb-model-tier-qualification-manifest.md` | planned | Two serial slices: qualification-plan receipt, then experimental offline scorer. |
| Deepseek4 Medium qualification | skipped-clear | TokenZoom PR #2 | running | Session `bd6392b6-cc7e-4856-b511-c948ecd4df8c`; reuse trustworthy evidence before any bounded local rerun. |

## Dependency Map

```text
portable plan-sufficiency contract ----\
                                         -> final bounded DDR/DS4 decision
TokenZoom Deepseek4 qualification ------/
```

The workstreams may run independently. The final claim requires both.

## Execution Queue

1. Create and review portable plan-sufficiency requirements.
2. Plan and execute the portable contract manifest.
3. Receive the TokenZoom qualification receipt.
4. Reconcile the final bounded conclusion and refresh project memory.

## Human Checkpoints

None. The user authorized a bounded Deepseek4 evaluation. Paid calls, unsafe
fleet mutation, hardcoded model promotion, and broad Medium claims remain
forbidden.

## Parked / Blocked

- The stale `2026-07-27` EDDR handoff is historical evidence only. Benchmark
  ownership moved to TokenZoom.
- The old `kb-ddr-plan` proposal is not a merge source. Its useful rule is
  absorbed into the generic plan-sufficiency contract.

## Completion Criteria

- The portable manifest is complete and passes repository release/sync gates.
- TokenZoom reports `qualified`, `not-qualified`, or `inconclusive` with
  mechanically checkable evidence and exact exclusions.
- The epic records the final decision without overstating the cohort.
