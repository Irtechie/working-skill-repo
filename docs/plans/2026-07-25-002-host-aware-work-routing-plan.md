---
type: kb-slice-plan
kb_id: kb-2026-07-25-orchestrator-directed-ddr
slice_id: ddr-002
title: Host-aware work routing
status: done
blockers: []
verification: verification-only
test_level: functional-cli
functional_risk: narrow
model_tier: large
model_tier_reason: "The policy controls every later worker selection and reconciles active host and CLI capability surfaces."
model_requirements: ["cross-skill consistency", "host capability reasoning", "documentation contract review"]
escalation_triggers: ["active host schema is unavailable", "global skill drift conflicts with source policy"]
workspace_mode: worktree-required
conflict_domains: [".github/skills", "README.md", "LOCAL_MODELS.example.md", "docs/context/architecture"]
shared_resources: ["git:integration-owner", "sync:global-skills"]
context_packet: docs/plans/2026-07-25-orchestrator-directed-ddr-context/ddr-002.json
expected_files:
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: define minimum capability
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: replace normal AMR loop with one ownership decision
  - path: .github/skills/kb-work/references/execution-prompt.md
    op: edit
    scope: make worker prompt owner-first
  - path: .github/skills/kb-models/SKILL.md
    op: edit
    scope: document active surface and user-local catalog
  - path: .github/skills/kb-configure/SKILL.md
    op: edit
    scope: isolate AMR experiment configuration
  - path: .github/skills/kb-configure/references/kb-routing-example.yaml
    op: edit
    scope: mark AMR state experimental-only
  - path: .github/skills/kb-functional-test/SKILL.md
    op: edit
    scope: remove model-attempt coupling from proof classification
  - path: README.md
    op: edit
    scope: explain orchestrator-directed DDR
  - path: LOCAL_MODELS.example.md
    op: create
    scope: document private route configuration
  - path: docs/context/architecture/kbrouter.md
    op: edit
    scope: record owner-first subsystem contract
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: replace the canonical normal AMR loop with owner-first DDR
  - path: evals/skill-eval/selftest/pass-session-model-routing.json
    op: edit
    scope: assert owner-first production routing
  - path: cmd/kbcheck/ddr_contract_test.go
    op: create
    scope: prevent AMR leakage and App/CLI surface conflation
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck core"
  expect: 0
---

# DDR-002: Host-aware work routing

## Goal

Make KB planning and work instructions use the active orchestrator's callable
surface plus user-local model state.

## Scope

- `.github/skills/kb-plan/SKILL.md`
- `.github/skills/kb-work/SKILL.md`
- `.github/skills/kb-work/references/execution-prompt.md`
- `.github/skills/kb-models/SKILL.md`
- `.github/skills/kb-configure/SKILL.md`
- `.github/skills/kb-configure/references/kb-routing-example.yaml`
- `.github/skills/kb-functional-test/SKILL.md`
- `README.md`
- `LOCAL_MODELS.example.md`
- `docs/context/architecture/kbrouter.md`

## Contract

1. The orchestrator reasons about the minimum tier and whether its own
   reasoning, context, tools, or authority are needed.
2. It chooses `current` or `delegated` once.
3. A delegated slice gets one exact qualified route from the active host schema
   or the live CLI/user-local catalog.
4. App-specific and CLI-specific aliases are not conflated.
5. AMR is documented as an unpromoted experiment, not a prerequisite.

## Proof

- Skill contract checks.
- Exact text checks for prohibited automatic AMR/downward fallback language.
- `git diff --check`.

## Acceptance criteria

- Production skills never require or trigger an AMR attempt.
- The orchestrator owns both tier and owner decisions.
- App and CLI model catalogs are not conflated.
- Private endpoints and credentials remain user-local.
