---
name: kb-eval-map
description: Repo-native eval setup mapper. Use during kb-map-bootstrap, when a repo has no meaningful eval harness, or when the user asks how to evaluate an app/workflow. Detects app pattern, chooses the right proof tools, scaffolds one safe smoke eval when possible, and wires the command into KB testing docs/checks.
argument-hint: "[optional project focus, workflow, or eval surface]"
---

# KB Eval Map

Set up the repo's eval strategy. This skill answers:

> What does correctness mean in this project, and what executable proof should
> KB workflows use?

Skills state policy. Repo-local tests, scripts, eval cases, traces, and scorers
are the judges. Do not hide the test suite inside a skill.

## When To Run

Run from `kb-map-bootstrap` after the repo inventory pass and before final
testing docs are written.

Also run when:

- `kb-check` finds no meaningful command for the repo;
- a new app/workflow is added without an eval surface;
- the user asks how to evaluate a non-LLM app, website, CLI, API, or agent flow;
- repeated bugs show that existing tests are mocked theater or too low-level.

## Outputs

Create or update, as appropriate:

```text
docs/context/eval-map.md
docs/context/operations/testing.md
evals/
tests/evals/
scripts/eval-run.ps1
scripts/eval-score.ps1
```

Only create executable files when they are safe and useful. Prefer documenting a
gap over creating fake tests.

## Intent Gate

Inspect before asking. Read existing evidence first:

- `README.md`, `AGENTS.md`, `.github/copilot-instructions.md`;
- package/test/build manifests and CI workflows;
- routes, pages, commands, APIs, jobs, tools, scripts, and existing tests;
- `docs/context/PROJECT.md`, `docs/context/operations/testing.md`, and current
  handoffs or plans.

If the primary workflow is obvious from repo evidence, scaffold one real smoke
eval when safe.

If the repo is new or the primary workflow is unclear, ask exactly one question:

```text
What is this repo supposed to prove works? Name the main workflow, command, page, API, or user job.
```

If the answer is unavailable, write `docs/context/eval-map.md` with the known
surfaces and add a visible `todo.md` item. Do not create placeholder tests that
pass without proving behavior.

## Pattern Matrix

Choose the native proof surface. Do not force one eval framework everywhere.

| Repo Pattern | Default Eval Setup |
|---|---|
| Website/web app | Playwright/Cypress/browser assertions over real routes, DOM state, console/network checks |
| Internal/corporate website | CDP or authenticated browser transport using the user's real session; do not fake SSO |
| API/service | Contract smoke tests: status, schema, required fields, auth/error behavior |
| CLI/tooling | Golden command tests: args, exit code, stdout/stderr, file side effects |
| LLM/agent app | Prompt/output datasets, deterministic scorers, DeepEval/Promptfoo when useful, optional Langfuse/Braintrust export |
| Skill repo | Prompt routing, trace proof, claim verification, output-quality scoring, skill regression matrix, cost telemetry |
| Docs/process repo | Link checks, consistency checks, claim-grounding checks, stale reference checks |
| Mobile/native | Platform UI test hooks when available; otherwise build/install smoke plus HITL gaps |
| Mixed repo | Map each major surface separately and pick the narrowest shared command that exercises the changed surface |

## Harness Discovery

Check for existing tools before scaffolding:

- JavaScript: `package.json`, Playwright, Cypress, Vitest, Jest, pnpm/npm/yarn
  scripts.
- Python: `pyproject.toml`, `pytest`, `tox`, DeepEval, Promptfoo wrappers.
- PowerShell/.NET: Pester, `.sln`, `.csproj`, `*.ps1` smoke scripts.
- API/CLI: curl scripts, Make/just/Taskfile commands, generated clients,
  OpenAPI schemas.
- Eval dashboards/exporters: Langfuse, Braintrust, LangSmith.
- CI: `.github/workflows/**`, existing test matrices, required checks.

Prefer existing project commands over invented commands. If adding a command,
make it non-interactive, scoped, repeatable, and documented.

## Scaffold Policy

Default scaffold level: one real smoke eval for the highest-value workflow.

Safe to scaffold when:

- the primary workflow is known;
- a local/dev target can run without credentials, MFA, production-only systems,
  destructive writes, paid external calls, or private data;
- the assertion is observable and would fail if the workflow is broken;
- the command can exit nonzero on failure.

Do not scaffold when:

- the workflow depends on credentials or session state not available to the
  agent;
- the workflow is destructive or production-only;
- the correct user job is a product decision;
- the repo is too new to identify a meaningful smoke path.

For unsafe cases, write the eval map and add a `todo.md` item with the exact
missing input or access.

## Eval Map Shape

Use this structure for `docs/context/eval-map.md`:

```markdown
# Eval Map

Checked: YYYY-MM-DD

## App Pattern

## Primary Workflows

| Workflow | Surface | Current Proof | Gap | Priority |
|---|---|---|---|---|

## Existing Harnesses

## Canonical Commands

## Scaffolding Decisions

## Deterministic vs LLM-Judged

## Credentials / Session Requirements

## Dashboard / Export Options

## Open Eval Gaps
```

Classify each check as:

- `deterministic` - command, test, browser assertion, file/git/API proof;
- `llm-judged` - rubric quality score, semantic comparison, subjective review;
- `hitl` - true human judgment or credentials/session requirement;
- `exporter-only` - dashboard/reporting integration, not the judge.

## Wiring Rules

Wire into `kb-check` only when the command is runnable and low-risk.

Acceptable wiring:

- add or document a repo-local script such as `scripts/eval-run.ps1`;
- add the command to `docs/context/operations/testing.md`;
- extend the repo's existing `kb-check` helper or package script when that
  pattern already exists.

Do not make `kb-eval-map` the runtime verifier. After setup:

- `kb-check` runs deterministic commands;
- `kb-functional-test` chooses proof level per slice;
- `kb-qa` executes browser/API/CLI workflow checks;
- `kb-regression-snapshot` freezes passing behavior;
- `kb-complete` enforces final machine-verifiable proof.

## Output

Report briefly:

- detected repo pattern and primary workflow confidence;
- existing harnesses found;
- eval map path;
- any scaffolded command/test and how to run it;
- what was wired into `kb-check` or testing docs;
- remaining eval gaps and exact human-required input.
