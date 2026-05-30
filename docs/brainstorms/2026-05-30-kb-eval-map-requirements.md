---
date: 2026-05-30
topic: kb-eval-map
brainstorm_style: kb-brainstorm
---

# KB Eval Map

## Problem Frame

KB workflows already enforce proof during work, but they do not reliably set up a
repo-specific eval system before serious work starts. `kb-map-bootstrap` is the
right expensive-once moment to inspect a project, decide what correctness means,
and leave behind a native eval map.

The goal is not to force LangChain, Langfuse, Braintrust, DeepEval, Promptfoo, or
Playwright into every repo. The goal is right tool for the right job: web apps
need browser workflow proof, CLIs need command/output proof, APIs need contract
proof, skill repos need prompt routing and trace/claim proof, and LLM apps may
need prompt/output scoring frameworks.

## Requirements

**Skill Boundary**
- R1. Add a new `kb-eval-map` skill that owns repo eval-surface discovery,
  native harness selection, minimal scaffolding, and `kb-check` wiring guidance.
- R2. Keep executable eval harnesses outside skills. Skills state policy; repo
  scripts/tests/evals are the judges.
- R3. `kb-map-bootstrap` must invoke `kb-eval-map` during bootstrap after repo
  inventory and before final testing docs are written.

**Intent Gate**
- R4. `kb-eval-map` must inspect existing repo evidence first: README, tests,
  scripts, routes/screens, commands, APIs, CI, and project memory.
- R5. If the primary workflow is obvious from repo evidence, scaffold one real
  smoke eval when safe.
- R6. If the repo is new or the primary workflow is unclear, ask one question:
  "What is this repo supposed to prove works? Name the main workflow, command,
  page, API, or user job."
- R7. If the answer is unavailable, write the eval map and a visible todo instead
  of creating fake tests.

**Native Harness Selection**
- R8. Detect project patterns: website, internal/corporate website, API/service,
  CLI/tooling, LLM/agent app, skill repo, docs/process repo, mobile/native, or
  mixed.
- R9. Detect existing harnesses: Playwright, Cypress, pytest, Vitest/Jest,
  Pester/PowerShell scripts, curl/API probes, Promptfoo, DeepEval, Langfuse,
  Braintrust, LangSmith, CI workflows, and repo-specific browser transports.
- R10. Recommend the narrowest native harness that can prove the highest-value
  workflow. Framework dashboards are optional exporters, not the source of truth.

**Outputs**
- R11. Create or update `docs/context/eval-map.md` with app pattern, eval
  surfaces, existing harnesses, chosen native proof commands, missing gaps,
  credential/session requirements, and optional dashboard/export choices.
- R12. Update `docs/context/operations/testing.md` when a canonical eval command
  is added or discovered.
- R13. Scaffold minimal repo-local eval files only when safe: examples include
  `evals/`, `tests/evals/`, `scripts/eval-run.*`, or a single Playwright/API/CLI
  smoke test.
- R14. Wire one canonical eval command into `kb-check` only when the command is
  runnable and low-risk. Otherwise document the command and todo.

**Quality Bar**
- R15. The first scaffolded eval should be real, not a placeholder: one primary
  workflow, observable assertion, nonzero exit on failure, and reproducible
  command.
- R16. Classify each eval as deterministic, LLM-judged, HITL, or exporter-only.
- R17. For corporate/internal web apps, prefer authenticated CDP/session
  transport over fake Playwright login when SSO/Conditional Access blocks clean
  browser sessions.
- R18. For skill repos, eval surfaces must include prompt routing, trace proof,
  claim verification, output quality, skill regression matrix, and cost telemetry.

## Success Criteria

- Fresh bootstrap leaves a project with `docs/context/eval-map.md`.
- `kb-check`, `kb-functional-test`, `kb-qa`, and `kb-complete` can read the eval
  map to choose repo-native proof.
- New/unclear repos do not receive fake smoke tests.
- Existing repos with obvious primary workflows get one runnable smoke eval when
  safe.
- Skill edits can be evaluated by tests outside the skills themselves.

## Scope Boundaries

- Do not implement full Langfuse/Braintrust/LangSmith integration in v1.
- Do not make a universal eval framework for all apps.
- Do not make `kb-eval-map` replace `kb-functional-test`, `kb-qa`, or
  `kb-complete`; it gives them repo-local eval context.
- Do not scaffold tests that need credentials, production data, paid APIs, MFA,
  or destructive actions without explicit user input.

## Key Decisions

- `kb-eval-map` gets authority to map, scaffold, and wire into `kb-check` when
  safe.
- The default scaffold is one real smoke eval for the highest-value workflow,
  not an empty skeleton and not a full suite.
- Bootstrap should spend the tokens because this setup is expected to happen once
  or after major repo drift.

## Next Steps

-> /kb-plan
