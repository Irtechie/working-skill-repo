# ATV-Phoenix Self-Heal Comparison

Checked: 2026-07-09
Budget mode: standard

## Question

Should KB replace its learning/self-healing loop with ATV-Phoenix, or merge specific Phoenix mechanics into the existing KB app-local learning model?

## Findings

ATV-Phoenix latest inspected commit:
`c083b88eab22c143163d13158559c0ced56b7d1b` (`2026-07-09`, `feat(sense): add UiBehavior CheckKind...`).

## Attribution Boundary

KB's route selection, vertical slicing, dependency DAG, and model-tier planning
are the KB maintainer's work. They are not Phoenix ideas and must not be
described as "absorbed from Phoenix." Phoenix remains credited for the specific
proof-spine mechanics below where its implementation was the inspected prior
art.

Phoenix provides a strong, compact executable proof spine:

- `phoenix_sense`: objective command/hash/regex checks.
- `phoenix_snapshot`: only blesses known-good state when a check passes.
- `phoenix_heal`: bounded retry or rollback, confirmed by an external recheck.
- `phoenix_verify_trace`: tamper-evident hash-chain verification.
- `phoenix_accept`: failure-first gate; a check is accepted only if the trace proves red -> green and it is green now.

This contributes a capability KB can learn from: completion is computed from
trace state rather than authored in prose. KB already has deterministic
verification, protected oracle checks, scoped learning, `kb-repair` iteration
ceilings, and repo-local memory; a generic trace-derived acceptance ledger
complements those project-workflow surfaces.

The two systems have different product boundaries:

- KB's app-local scoped learning model is more appropriate for downstream apps: default to narrow scope, pull ancestors only, promote on recurrence.
- KB's lane routing is richer for brainstorm/plan/work/complete/ship, review gates, handoffs, and repo memory.
- Phoenix's published skill pack supplies its own lifecycle vocabulary, while KB
  already routes those lifecycle phases through its existing skills.
- Phoenix's learning gate focuses on measured adoption eligibility for
  prompt/skill diffs; KB's scoped instincts address project-local memory and
  promotion ownership.

## Complementary Strengths

KB's existing strengths for long-running project work are:

- **Scoped learning:** lessons live at the narrowest workflow/app scope and climb
  only by recurrence. Phoenix contributes a measured adoption gate with a
  different purpose.
- **Route selection:** KB has proportional lanes for fix, troubleshoot,
  brainstorm, plan, work, complete, ship, epic, and durable goals.
- **Repo memory:** KB keeps project maps, handoffs, eval maps, solutions, todo
  ledgers, and scoped instincts in the working repo.
- **Completion discipline:** KB already has review, QA, compound, learn, evolve,
  and sync gates. Phoenix contributes computed failure-first acceptance.

Candidate integration path:

1. Add a KB-native `kb-sense`/`kb-accept` proof layer, preferably inside `cmd/kbcheck`, using Phoenix's trace pattern.
2. Make `kb-goal`, `kb-work`, `kb-repair`, and `kb-troubleshoot` record objective sense events and require failure-first accept for fixes where a failing oracle exists.
3. Add optional snapshot/rollback for files touched by repair loops, but keep git commits/reverts as the primary durable rollback mechanism.
4. Port the measured-gain adoption gate idea into `/learn` + `/evolve`: a candidate shared/global skill change must clear held-out eval gain, no right-to-wrong regressions, and human approval.
5. Keep KB scoped learning as the memory model; Phoenix's measured gate becomes an adoption/promote gate, not the default memory store.

## 2026-07-09 Freshness Audit

Phoenix ideas worth evaluating for KB integration:

- Content-address the prompt/skill surface so drift is reported as explicit
  added/removed/changed files. Compare a native implementation with a focused
  Phoenix/MCP interoperability path using installation and maintenance evidence.
- Isolate proof traces by goal/slice/run. A shared trace keyed only by a check
  digest can let identical checks contaminate each other's evidence.
- Provide a compact read-only monitor over run state, trace integrity, current
  slice, and final acceptance.
- Keep OKF-style index-first Markdown bundles as an optional interchange format
  for project knowledge, with copyable files available alongside any MCP path.
- Preserve the tri-state idea for subjective/generative review: accept, reject,
  or request more evidence. Treat confidence values as qualitative until their
  weights are calibrated against real outcomes.

The new `UiBehavior` check is useful prior art and needs these adaptations for
KB's verifier-integrity contract:

- its default `verify-ui.mjs` returns `ok=true` when no checks are configured;
- `canonical_digest` does not hash the UI verifier for `UiBehavior`;
- command verifier hashing covers only `target[0]`, missing the common
  `node checker.js` / `python checker.py` shape;
- `cwd` and timeout are excluded from Phoenix's check identity, and the timeout
  field is not enforced in-process.

The reusable lesson is to bind acceptance to every oracle input and require a
non-empty behavioral assertion set. KB's Go proof spine already enforces
timeouts, but it has the same interpreter-script hashing gap and should add an
explicit `oracle_files`/verifier-input contract.

## Sources

- https://github.com/All-The-Vibes/ATV-Phoenix/tree/main
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/src/sense.rs
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/src/heal.rs
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/src/snapshot.rs
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/src/trace.rs
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/src/accept.rs
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/src/bin/phoenix_mcp.rs
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/phoenix_learn/gate.py
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/phoenix_learn/optimize.py
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/evals/m1-self-heal/RESULT.md
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/evals/autonomous-workflows/RESULT.md
- https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/evals/c3-phoenix-learn/RESULT.md

## Applies When

- Designing KB self-heal proof, repair gates, long-running goal completion, or skill/learning promotion.
- Deciding whether an agent may mark work done after tests pass.
- Promoting local app learning into shared/global skills.

## Stale When

- Phoenix changes its spine semantics or ships a materially different `phoenix-learn` skill.
- Phoenix updates the verifier-integrity and empty-UI-check behavior described
  above.
- A downstream app proves a different self-heal ledger in production.

## Integration Options Considered

- Use Phoenix as KB's complete lifecycle: not selected because KB must preserve
  its existing route ownership and scoped app-local learning contract.
- Install Phoenix's complete skill pack alongside KB: not selected by default
  because both systems define lifecycle entry points; a focused bridge remains
  eligible.
- Keep KB unchanged: not selected because executable failure-first acceptance is
  valuable and Phoenix demonstrates it well.

## Impact On Current Project

The proof spine, learning-adoption gate, route-history guard, and install doctor
are implemented in Go, and Go remains the selected KB core. A focused
Phoenix/MCP bridge remains eligible as an optional integration over that core.
Candidate features are verifier-input integrity, isolated proof namespaces, a
read-only run monitor, and optional prompt-surface drift reporting. User-facing
docs must credit Phoenix for the specific proof mechanics it informed while
preserving the separate origin of KB routing and slicing.
