---
kb_id: kb-2026-08-23-epistemic-investigation-gate
slice_id: epistemic-003
title: Add the narrow investigation behavior after baseline freeze
blockers: [epistemic-002]
verification: integration
test_level: integration
functional_risk: broad
model_tier: large
model_tier_reason: The treatment changes supported cross-runtime instruction surfaces while preserving authority and the invisible proceed path.
model_requirements: [instruction design, cross-runtime behavior, compatibility testing, cognitive-load preservation]
escalation_triggers: [always-on checklist emerges, agent-owned research becomes a user question, treatment requires changing kb-cognitive, kb-work, manifest-contract, or frozen eval owners]
workspace_mode: shared-serial
conflict_domains: [skill:kb-plan, skill:kb-gate, file:config/skill-guidance-audit.json]
shared_resources: [git:integration-owner, instruction:planning-trust-brake]
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck skill-lint"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: Require a post-draft challenge loop whose repeated passes must add evidence, revise the plan, or stop inconclusive.
  - path: .github/skills/kb-gate/SKILL.md
    op: edit
    scope: Define planning assurance states, materiality, progress, and the no-justified-conclusion boundary without turning safe planning into ceremony.
  - path: config/skill-guidance-audit.json
    op: edit
    scope: Record ambient and detailed-skill ownership without claiming universal model coverage.
protected_oracles:
  - path: evals/skill-eval/baselines/epistemic-investigation.json
    role: Untouched baseline.
    sha256: filled by kb-work from slice epistemic-002 output
    update_policy: Immutable during treatment; recapture requires restarting the experiment.
  - path: evals/skill-eval/epistemic/oracles/
    role: Frozen hidden labels.
    sha256: filled by kb-work from slice epistemic-002 freeze receipt
    update_policy: Treatment executor may not inspect or change holdout labels.
status: pending
owner: agent
blocked_reason: ""
resume_when: epistemic-002 baseline and hashes are verified
next_agent_action: Protect hashes, add the smallest bounded treatment, and run structural compatibility checks without changing frozen eval surfaces.
human_action: ""
can_continue_other_slices: false
---

# Narrow Investigation Treatment

## Deliverable

A bounded planning-instruction change that challenges a completed draft before
`plan-to-work`, investigates unsupported load-bearing premises autonomously,
revises on evidence, and permits an honest non-conclusion when evidence stalls.

## Acceptance Criteria

- Adequately supported work proceeds without visible verification ceremony.
- Researchable uncertainty is agent-owned and resolved before action or
  mutation.
- User questions remain limited to intent, authority, access, private input,
  irreversible risk, or subjective judgment.
- `no-justified-conclusion` is valid only after available in-scope evidence is
  exhausted or unavailable.
- Routine responses gain no required claim ledger, citation block, confidence
  score, or checklist.
- Existing deterministic proof gates remain authoritative.
- `kb-cognitive` is consumed unchanged as the response contract.
- Repeated challenge passes require an evidence or plan delta; unchanged
  reassurance terminates as inconclusive rather than looping.
- The behavioral claim is bounded to contexts that actually load `kb-plan`
  and `kb-gate`.
- Fixture, oracle, scorer, schema, adapter, baseline, and test hashes remain
  unchanged throughout treatment. Behavioral acceptance is owned by the final
  matched replay, not a post-baseline test rewrite.

## Scope Boundary

No fixture, scorer, oracle, result schema, adapter, baseline, test,
`kb-cognitive`, `kb-work`, manifest-validator, live-run, sync, or delivery
change. Deterministic enforcement is a separate post-promotion slice.
