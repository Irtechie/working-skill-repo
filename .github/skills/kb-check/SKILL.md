---
name: kb-check
description: Deterministic verification harness for KB workflows. Use when code should be tested, linted, typechecked, built, security-checked, or validated by scripts instead of relying on LLM judgment; also use before kb-complete, kb-ship, or after kb-work slices.
argument-hint: "[optional scope, changed files, or command]"
---

# KB Check

Prefer executable truth over model judgment. If a script can check it, run the script.

## Rule

LLM review can find risks, but it does not prove behavior. A slice is not verified until deterministic checks pass or a clear reason is recorded.

## Check Sources

Discover commands from:

- `package.json`, `pnpm-workspace.yaml`, `turbo.json`, `nx.json`
- `pyproject.toml`, `requirements*.txt`, `pytest.ini`, `tox.ini`
- `.csproj`, `.sln`, `global.json`
- `Makefile`, `justfile`, `Taskfile.yml`
- repo docs: `README.md`, `AGENTS.md`, `docs/context/operations/testing.md`
- existing CI files under `.github/workflows/`

Prefer existing project commands over invented commands.

## Workflow

1. Run `.github/skills/kb-check/scripts/kb-check.ps1 -List` when present to inspect discovered commands.
2. Pick the narrowest commands that verify the touched behavior.
3. Run checks in this order when available: format/lint, typecheck/static analysis, unit tests, integration/e2e/browser checks, build/package, security/dependency audit.
4. Capture command, exit code, and relevant output.
5. If a check fails, route to `kb-repair` or `kb-fix`; do not ask the user to test normal app behavior.
6. If a check is missing, add a small reusable script or test when practical, then document it in `docs/context/operations/testing.md`.

## Functional Checks

Use `kb-functional-test` when a change touches user-visible behavior, API/CLI workflows, persistence, auth, streaming, integrations, or any bug that escaped unit tests.

Default timing:

- Slice: narrow functional check for the changed path.
- Manifest complete: broader smoke tests over changed workflows.
- Ship: full functional/e2e suite when practical.

Headless by default. Do not spawn visible browser windows from multiple workers; serialize browser/e2e checks.

## Script Rule

When the same manual verification would be repeated twice, create a script.

Good scripts accept scope arguments, print concise pass/fail output, exit nonzero on failure, avoid network unless needed, run in CI or from an agent session, and are documented in `docs/context/operations/testing.md`.

## Output

Report commands run, pass/fail status, failures fixed or parked, checks added, and remaining manual-only verification with why it cannot be automated.
