---
name: kb-ddr-plan
description: "Choose the minimum capable execution tier and make a delegated plan detailed enough for that tier to succeed unaided without writing the implementation into the plan. Use from kb-plan for runnable slices, or when evaluating DDR tier choice, plan sufficiency, and whole-route economics."
argument-hint: "[slice requirements and repository evidence]"
---

# DDR Planning Guide

DDR is a planning discipline. A stronger planner chooses the minimum capable
execution tier and writes a **minimum sufficient execution contract** for that
tier. Runtime fallback or stronger-model repair is AMR, not DDR proof.

## Success Contract

Plan success means the assigned tier succeeds without stronger-model rescue.
The executor may inspect the named source, choose local implementation details,
write code, and run proof. It must not need another model to discover omitted
requirements, repair the plan, or supply fixture-specific answers.

The planner owns two independent judgments:

1. **Tier correctness:** the selected tier has enough residual reasoning,
   context, tools, and authority after reading the plan.
2. **Plan sufficiency:** the plan removes product and repository ambiguity
   without doing the executor's implementation work.

Do not treat a later successful escalation as DDR success.

## Build The Execution Contract

For every runnable slice, provide:

- **Objective:** one observable outcome, not a broad phase.
- **Repository route:** relevant files, symbols, callers, and contracts already
  established by evidence. Mark forecasts as forecasts.
- **Behavior contract:** inputs, outputs, state transitions, compatibility, and
  error behavior that the implementation must preserve.
- **Invariants and edge cases:** only source-derived or requirement-derived
  cases known before execution.
- **Implementation boundary:** allowed files or scope forecast, public APIs that
  must remain stable, non-goals, and forbidden shortcuts.
- **Proof:** exact command and acceptance assertions, with protected-oracle
  hashes when available.
- **Stop conditions:** observable evidence that the task is underspecified,
  exceeds the selected tier, or requires a planner decision.

Write enough that the executor can begin at the named code boundary without
rediscovering product intent. Stop before choosing ordinary local code
structure that tests can judge.

## Do Not Write The Solution

Do not include patches, function bodies, or fixture answers. Avoid code fences,
diff hunks, hidden-oracle deductions, and line-by-line implementation recipes.
API signatures, data shapes, and short pseudocode are allowed only when they are
part of the public contract or resolve otherwise material ambiguity.

Every instruction must reduce execution uncertainty. If making a lower tier
viable requires the planner to solve the implementation, select a higher tier
instead. A cheap executor carrying an expensive prewritten solution is not a
DDR savings result.

## Choose The Tier From Residual Work

Classify the work left after the execution contract—not the task title alone.

| Tier | Residual work the plan may leave |
|---|---|
| `small` | Mechanical local implementation with explicit contracts, narrow files, and focused proof |
| `medium` | Repository reasoning across a few boundaries, ordinary refactoring choices, or integration details with settled behavior |
| `large` | Unresolved architecture, security/data authority, broad migration, weak proof, or decisions whose mistakes have wide blast radius |

Choose the higher tier when evidence is incomplete. Record a falsifiable
`model_tier_reason` that names the residual reasoning burden and why the plan
makes that tier sufficient. Requirements, tools, context, and proof remain
portable; never record a model, provider, route, endpoint, adapter, or
transport.

## Preregister And Diagnose Failures

Freeze the plan, selected tier, task inputs, and protected proof before the
executor runs. No post-failure hint counts as DDR success.

When a trial fails, diagnose it with controlled follow-ups:

- If the same tier with a richer preregistered plan succeeds, classify
  `plan-insufficient`.
- If the next tier succeeds with the original plan, classify
  `tier-underestimated`.
- If neither succeeds, classify `task-or-oracle-defect` until the fixture,
  specification, packet, and proof are independently cleared.

Do not rewrite the original result after a follow-up succeeds. Keep setup and
infrastructure defects outside model correctness.

## Count Whole-Route Economics

DDR cost includes planner input and output, executor input and output, repeated
plan ingestion, verification, invalid attempts, and failed accepted-scope
trials. Report provider-billed credits or dollars when available; tokens explain
movement but do not replace billed cost.

The primary KPI is dollars per accepted result against the same-task direct
large-tier baseline. Also report:

- assigned-tier acceptance rate;
- planner cost share;
- plan input/output tokens;
- executor tokens;
- plan-to-artifact token ratio;
- right-to-wrong regressions;
- setup or infrastructure not-run counts.

Do not claim savings from a successful lower-tier call alone.

## Return To `kb-plan`

Return one compact decision per slice:

```text
tier: small|medium|large
tier_reason: <residual reasoning burden and sufficiency claim>
execution_contract: <objective, route, behavior, invariants, boundaries, proof, stop conditions>
plan_token_budget: <bounded output ceiling or explicit no-ceiling reason>
leakage_check: pass|fail
```

`kb-plan` incorporates the decision into the slice plan. `kb-work` resolves a
qualified live route and executes the frozen plan; it does not reinterpret a
failed DDR trial as success through fallback.
