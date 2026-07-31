---
date: 2026-07-30
topic: proportional-agent-review
brainstorm_style: kb-brainstorm
---

# Proportional Agent Review

## Problem Frame

Routine KB completion currently launches six code-review subagents before any
conditional specialists are added. The same orchestration is duplicated across
`ce-review` and `kb-review`, making ordinary completion disproportionately
expensive and difficult to evolve safely.

## Research Summary

**Findings that shaped requirements:**

- Anthropic reports multi-agent systems use about 15 times the tokens of chat
  and are a weaker fit for most coding tasks than for breadth-first research.
- GitHub Copilot code review exposes effort levels so routine changes do not pay
  the cost of security-sensitive or cross-service analysis.
- Google recommends small, self-contained changes and qualified specialists for
  the specific areas that need them.
- Matt Pocock's `code-review` separates review into Standards and Spec axes,
  demonstrating a compact intent-and-quality framing without a broad persona
  roster.
- Local inspection confirms `ce-review` and `kb-review` each mandate six
  always-on reviewers, while `kb-finalize` unconditionally invokes `kb-review`.

Primary evidence:
`docs/context/research/2026-07-30-proportional-agent-code-review.md`.

**Confidence:** High - the current behavior is directly verified in local skill
files and the replacement pattern is supported by primary-source guidance.

## Requirements

**Assurance that is always required**

- R1. Before slicing, the planning agent must establish authoritative
  requirements, acceptance criteria, material risks, and the verification
  strategy. This self-check must not launch a reviewer merely because planning
  is running or a document crosses a size threshold.
- R2. Every implementation slice must retain its coded tests and deterministic
  proof checks. Reviewer agents must never run after an individual slice.
- R3. After all slices are integrated, the workflow must run the aggregate
  deterministic proof before semantic review.
- R4. A non-trivial integrated change receives one broad semantic review
  covering four questions: does the change satisfy authoritative intent, can
  its tests detect relevant breakage, is the implementation correct, and did
  code health materially regress?
- R5. Existing KB tests, functional checks, exact-tree proof, and repository
  gates remain authoritative. Do not add Phoenix runtime or duplicate proof
  semantics.

**Optional escalation**

- R6. Pre-slice `document-review` is optional. Use it only when unresolved
  ambiguity, high-stakes consequences, a disputed decision, or an unfamiliar
  path-dependent architecture could materially change the plan.
- R7. Each boundary launches at most one reviewer. A pre-slice reviewer is the
  best-fit document reviewer. A post-integration reviewer is either the broad
  profile or one specialist profile that retains every R4 question while adding
  its specialty.
- R8. Thermonuclear review is optional and post-integration only. Use it for
  structural changes, abstraction growth, file sprawl, or architecture work,
  replacing the broad profile while retaining intent, test-validity, and
  correctness coverage.
- R9. Security, migration, performance, reliability, API-contract, and similar
  specialists activate only when the integrated diff crosses their actual risk
  boundary. Diff size or a keyword alone is not sufficient.
- R10. Classification is conservative and evidence-driven. Docs-only,
  generated-only, or mechanically constrained work may skip post-integration
  semantic review only when every changed path is covered by deterministic
  proof and no runtime, contract, configuration, trust-boundary, or behavioral
  surface changed. Unknown classification defaults to the broad reviewer. A
  specialist profile requires a concrete changed risk boundary recorded by the
  plan, changed paths, or deterministic classifier.

**Ownership and cleanup**

- R11. `kb-review` is the sole post-integration review orchestrator and owns
  classification, launch, merge, and receipt policy. Duplicate orchestrators
  with no distinct caller or public contract must be removed rather than
  retained as legacy aliases.
- R12. `kb-finalize` invokes `kb-review` once per integrated tree. Its receipt
  fingerprint includes the base and integrated tree, authoritative requirements
  hash, proof receipt hash, review-policy version, risk classification, and
  selected profile. Any code-affecting fix invalidates proof and review; the
  final exact-tree aggregate must pass after fixes.
- R13. `document-review` must have no always-on multi-persona baseline. Its
  default activated path is one best-fit reviewer; a second persona requires a
  separate material question.
- R14. Compound, learn, evolve, and memory refresh must not all run after every
  completed plan. Invoke them only when the work produced a reusable lesson,
  crossed a configured cadence, or changed project-memory routing.
- R15. Deterministic fixtures must prove boundary-only cadence, one-review
  routine behavior, evidence-triggered escalation, conditional Thermonuclear
  selection, safe receipt reuse, and removal of dead review-skill references.
- R16. The caller migration is explicit: `kb-finalize` continues through
  `kb-review`; standalone `kb-review` remains available; `ce-review` and its
  installer, documentation, guard, and global-install entries are removed.

## Success Criteria

- Mechanical/proof-complete work may launch zero reviewer subagents.
- A routine bounded code change launches no more than one post-integration
  review subagent.
- No reviewer subagent is launched after an individual slice.
- Routine plans launch at most one reviewer subagent after integration.
- Additional reviewer calls identify the concrete evidence that justified each
  escalation.
- No plan launches more than one reviewer at either boundary or more than two
  reviewer agents across its full lifecycle.
- Intent, proof validity, and code-health coverage remain visible in findings
  and output.
- Only one post-integration review orchestrator remains operational.
- Learning and memory work run only for a reusable lesson, configured cadence,
  or changed project-memory route; ordinary completion skips them.
- Existing KB proof, completion, sync, and release gates still pass.

## Scope Boundaries

- Remove review skills and compatibility surfaces that have no real caller or
  distinct contract; Git history is the recovery path.
- Preserve `document-review` as an explicit edge-case capability, but remove
  automatic activation based only on document size or generic materiality.
- Do not add Phoenix runtime, MCP, trace, or test semantics.
- Do not weaken deterministic tests, browser/API/CLI proof, or final exact-tree
  checks.
- Do not optimize model choice or pricing in this change.

## Key Decisions

- Review dimensions do not imply separate agents. One generalist reviews intent,
  proof validity, correctness, and code health; separate reviewers are
  evidence-driven escalation.
- Boundary-only reviewer cadence: optional pre-slice review protects uncertain
  decomposition and one post-integration review judges the integrated diff;
  per-slice feedback remains
  executable tests rather than repeated semantic review.
- Specialists are optional capabilities selected for a concrete unresolved
  question, not a roster selected from diff labels.
- Thermonuclear remains post-integration and conditional: structural review
  needs a real code diff and is valuable only when that diff changes
  architecture.
- Existing KB tests stay authoritative: the user clarified that plan-declared
  verification is native KB behavior, not a Phoenix adoption.
- One operational review orchestrator: dead aliases and duplicate engines are
  removed instead of protected indefinitely.

## Dependencies / Assumptions

- [safe-assumption] A compact generalist rubric can cover Standards and Spec
  without concatenating existing persona prompts. Reversible because specialist
  escalation remains available. Evidence/proof: captured review fixtures must
  retain both output axes and high-confidence finding classes.

## Alternatives Considered

- Keep six always-on reviewers: rejected because it pays multi-agent cost before
  risk classification.
- Remove semantic review entirely: rejected because executable checks cannot
  detect every intent, design, or maintainability mismatch.
- Adopt Phoenix review semantics: rejected for this change because KB already
  plans and executes its own tests and proof gates.
- Keep two independent Matt-style agents: better than six, but routine Standards
  and Spec review can share one bounded context and one reviewer call.
- Treat passing tests as complete proof of intent or test quality: rejected
  because tests cannot establish that the right behavior was specified or that
  the assertions would fail for relevant breakage.

## Slice Candidates (advisory for /kb-plan)

- Boundary-only review cadence - users receive at most one pre-slice document
  reviewer for the explicit R6 triggers and one post-integration reviewer,
  never reviewer agents per slice.
- Single-agent routine review - users receive intent, proof-validity,
  correctness, and code-health findings from one bounded integrated review.
- Risk-bounded profile selection - sensitive changes select one specialist
  profile that retains the broad review questions.
- Review ownership cleanup - one operational review skill remains and stale
  references, protection rules, installer entries, and drift guards are removed.
- Completion integration - KB finalization reuses proof and review receipts
  without duplicate review calls.
- Review policy proof - deterministic fixtures demonstrate reviewer counts,
  tier selection, wrapper behavior, and unchanged-tree reuse.

## Outstanding Questions

### Resolve Before Planning

None.

### Deferred to Planning

- [defer-to-planning][Affects R15][Technical] Reuse or extend the existing
  `kbcheck` review-reference guard and route fixtures rather than creating a
  separate test harness.

### Parked / Out of Scope

- [parked][Affects R5] Live token-cost benchmarking - Forbidden claim: this
  change proves a specific percentage cost reduction without measured runs.

## Next Steps

-> /kb-plan
