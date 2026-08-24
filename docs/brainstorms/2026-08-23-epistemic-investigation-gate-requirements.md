# Epistemic Investigation Gate Requirements

Status: ready for planning
Created: 2026-08-23

## Objective

Make the workflow pause and investigate when a material action-driving premise
is not adequately supported, while preserving the current low-ceremony path
when evidence is already sufficient.

The system cannot prove that a model is generally reliable or explain why a
model behaves confidently. It can measure and improve whether the workflow:

1. detects when investigation is required;
2. inspects evidence capable of resolving the premise;
3. revises the conclusion appropriately after inspection; and
4. proceeds without unnecessary investigation when the premise is adequately
   supported.

## Required Behavior

### Decision states

| State | Meaning | Required behavior |
|---|---|---|
| `proceed` | Material premises are adequately supported for the scoped action | Continue without visible verification ceremony |
| `investigation-required` | A material premise could change the action and can be resolved with available evidence | State the uncertainty internally, inspect the owning evidence, then reassess |
| `no-justified-conclusion` | Material uncertainty remains after available in-scope investigation is exhausted | Stop the unsupported conclusion; name the missing evidence and do not disguise uncertainty as a recommendation |

### Trigger boundary

The gate applies only when an unverified premise could materially change:

- the action or recommendation;
- safety, authority, scope, or data handling;
- verification validity; or
- likely rework cost.

It does not require exhaustive verification of harmless statements, stylistic
choices, reversible mechanical work, or premises already supported by current
authoritative evidence.

### Cognitive-load contract

- `proceed` remains the normal invisible path.
- Researchable uncertainty is agent-owned; do not ask the user to perform
  checks the agent can perform.
- Show a short progress update only when investigation adds meaningful delay or
  changes the expected course.
- Ask the user only for intent, authority, access, private input, irreversible
  risk, or subjective judgment.
- Do not expose claim ledgers, scoring fields, or internal checklists in routine
  user responses.

### Planning enforcement contract

KB planning is the first required application. A completed draft must not move
to `plan-to-work: passed` merely because its slices are well formed.

After drafting, `kb-plan` performs a challenge loop over only the load-bearing
factual premises that could invalidate a slice or its proof:

1. classify each premise as adequately supported or investigation-required;
2. inspect evidence capable of resolving every investigation-required premise;
3. record `supported`, `contradicted`, or `insufficient`;
4. revise the affected plan when the evidence changes it; and
5. repeat only the affected checks when the plan or evidence materially changes.

Every repeated pass must add evidence, revise a material claim, or terminate as
`inconclusive`. Unchanged reassurance is not progress. `insufficient` may not be
converted into a recommendation or a passed plan gate.

The assurance result is a hash-bound sidecar rather than user-facing ceremony.
It binds the requirements source and reviewed manifest/slice contents, records
the material premises and evidence references, records per-iteration deltas,
and ends in `passed` or `inconclusive`. Any bound plan edit invalidates the
receipt. Supported plans should normally complete this internally without a
user interruption.

After behavioral promotion, new plans use a new manifest schema that requires
the assurance sidecar. `manifest-contract` validates its structure and hashes;
`kb-work` refuses a missing, stale, or inconclusive receipt. Older manifest
schemas remain readable and resumable.

## Evaluation Contract

### Existing-owner constraint

Extend the existing `skill-eval` system rather than creating a parallel eval
platform. The current owners already provide result scoring, claim checks,
computed quality including `right_sized_ceremony`, baseline comparison,
regression reporting, sealed wrappers, and explicit live adapters.

The default implementation therefore:

- adds optional epistemic fields to the existing result contract;
- adds a focused epistemic scorer module called by `computeSkillEval`;
- uses the existing `skill-eval`, `skill-eval-regression`, `eval-run-*`, and
  `skill-eval-wrap` command family;
- extends the current live adapter to materialize an actor-only fixture
  workspace; and
- reuses existing baseline and regression artifacts.

A new top-level command or independent adapter is forbidden unless a RED spike
proves the current command family cannot express the required behavior safely.

### Detection

Use a two-class oracle:

- `investigate`: the material premise is not adequately supported;
- `proceed`: the material premise is adequately supported.

Report the full confusion matrix. A model that investigates every case must
fail on unnecessary-investigation rate rather than appear perfectly cautious.

### Resolution

For correctly detected `investigate` cases, score whether the model inspected
evidence capable of resolving the premise. Reading unrelated files or merely
restating uncertainty does not pass.

### Revision

After inspection, score whether the model correctly moved to:

- a supported conclusion;
- a contradicted conclusion; or
- `no-justified-conclusion`.

Keeping the original conclusion is correct when the inspected evidence
supports it. Changing an answer merely because a challenge occurred is not.

### Workflow cost

Keep these metrics separate from epistemic outcomes:

- tool calls;
- turns;
- user questions or interruptions;
- elapsed time when available; and
- input/output tokens when the runtime reports them.

Do not collapse detection, resolution, revision, and cost into one flattering
aggregate score.

## Fixture Integrity

- Finalize the initial corpus before baseline capture.
- Split development fixtures from sealed holdout fixtures.
- Establish labels independently from the acting or judging model.
- Prefer deterministic construction, hidden source truth, and exact expected
  evidence paths.
- Require every scored oracle to point to a machine-checkable source-of-truth
  rule. Hand-authored labels without deterministic provenance do not enter the
  correctness corpus.
- Keep oracle labels, expected evidence, and scoring rules out of the acting
  model's prompt and task workspace. The live adapter must execute from a
  disposable actor workspace containing only the visible task, controlled
  evidence files, and required instruction surfaces; the parent scorer retains
  the hidden oracle.
- Hash-bind the corpus, scorer, skill versions, runtime identity, and run
  configuration for every baseline and treatment result.
- Include supported `proceed` cases, resolvable `investigate` cases,
  contradicted-premise cases, and genuinely insufficient cases.
- Keep subjective architecture/preference cases in a separate process-only
  exploratory set. They may measure premise disclosure and investigation
  behavior, but not conclusion correctness.

## Experimental Sequence

1. Build and selftest the sealed deterministic evaluator.
2. Freeze the corpus and scorer.
3. Capture the untouched workflow baseline.
4. Introduce the narrow investigation behavior.
5. Replay hash-bound control and treatment instruction bundles in an
   interleaved matched run on the same runtime/model version and time window.
   Use the earlier untouched baseline as an anchor, not as the sole comparator.
6. Reject the treatment if it reduces missed investigations by causing an
   unacceptable increase in unnecessary investigations, user interruptions,
   or workflow cost.
7. Only after `promote`, enable deterministic planning enforcement in the next
   manifest schema and `kb-work` preflight. `reject` or `inconclusive` leaves
   enforcement disabled and routes back to a new treatment experiment.

Live model calls remain explicit and require a bounded preview with the runtime
list, repetition count, and estimated or capped spend when available.

After the corpus/scorer freeze, the treatment may change instruction surfaces
only. It may not change fixtures, oracle labels, scorer logic, result semantics,
adapter isolation, or protected tests. The frozen untreated instruction bundle
must remain reconstructable by content hash or Git blob identity for the final
matched control arm.

Every matched arm runs in a fresh context. Record the per-fixture arm order,
randomization seed or deterministic alternation rule, repetition count, model
identity, and available inference settings. If the host cannot establish a
matched identity or fresh-context boundary, the comparison is inconclusive.

The run must also inventory the instruction surfaces actually loaded by the
runtime. A disposable repository is insufficient when global skills, user
instructions, or cached session state can override or supplement the intended
control/treatment bundle. Use a supported isolated runtime profile when
available; otherwise prove the loaded surface hashes. If neither is possible,
the affected runtime result is inconclusive.

## Scope Boundaries

This work does not:

- claim to explain model training behavior;
- use model agreement as ground truth;
- use another model to label or judge the fixtures;
- fine-tune a model;
- require a visible verification checklist on every response;
- replace deterministic code tests, schema validation, or existing proof
  gates; or
- authorize live paid runs, global skill synchronization, commits, pushes, or
  delivery.

The resulting claim is bounded to the exact evaluated runtime/model versions
and contexts that load the modified instruction surfaces. It does not justify
"every model behaves correctly" or generalize beyond the tested fixture
distribution.

For genuinely subjective reasoning, this work can improve the discipline of
the process but cannot create an objective correctness oracle.

## Question Gate

| Item | Class | Resolution |
|---|---|---|
| Preserve ease versus maximize verification | `safe-assumption` | Preserve ease; the confusion matrix and cost metrics catch excessive verification |
| Exact first corpus size | `defer-to-planning` | Start with the smallest balanced corpus that exercises every outcome; expand only after scorer selftests pass |
| Exact live models and repetitions | `ask-now-at-execution` | Require a bounded preview and explicit approval immediately before baseline execution |
| Model-training explanation | `parked` | Outside the measurable objective; forbid causal claims |
| Global skill sync and delivery | `ask-now-at-release` | Keep separate from implementation and request authority at that boundary |

There are no unresolved `ask-now` or `research-first` items that prevent plan
creation. Implementation and live execution remain separately gated.
