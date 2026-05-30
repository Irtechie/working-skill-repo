# Testing Operations

Checked: 2026-05-29

## Current Commands

```powershell
git status --short
git diff --check
.\.github\skills\kb-check\scripts\kb-check.ps1 -List
.\.github\skills\kb-check\scripts\kb-check.ps1 -All
```

## Skill Quality Contract

The cross-runtime quality contract lives in `config/skill-quality.json`.
It defines:

- Codex and GHCP instruction surfaces.
- Skill lint budgets and allowlists.
- Route/complexity eval locations.
- Required and optional sync targets.

`kb-check.ps1` discovers this repo as a skill repo when `.github/skills` and
`config/skill-quality.json` exist.

The sync drift report is read-only:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\skill-sync-report.ps1
```

## Current Result

`kb-check.ps1 -List` now reports skill-repo checks when run here:

- `skill-lint`
- `route-complexity-eval`
- `skill-sync-report`

`kb-check.ps1 -All` runs all three and exits nonzero when a required check fails.
Expected current warnings:

- missing `argument-hint` on inherited/older skills;
- hot-path skill size warnings;
- optional ATV scaffold/plugin drift or omissions.

Required targets should report zero required sync issues.

## Eval Mapping

`kb-eval-map` is the bootstrap-owned setup skill for repo-native evals. It
creates or updates `docs/context/eval-map.md`, detects the app pattern and
existing harnesses, chooses the right proof surface, and scaffolds one real smoke
eval only when the primary workflow is known and safe to run.

Runtime proof still belongs to:

- `kb-check` for deterministic commands;
- `kb-functional-test` for per-slice proof-level classification;
- `kb-qa` for browser/API/CLI workflow checks;
- `kb-regression-snapshot` for replaying previous passing behavior;
- `kb-complete` for final machine-verifiable proof.

Eval frameworks such as Langfuse, Braintrust, LangSmith, DeepEval, or Promptfoo
are optional adapters/exporters unless the target repo is an LLM/agent app where
prompt/output datasets are the native proof surface.

## Required Harness

Add a repo-local `scripts/` or `.github/skills/kb-check` mode that can validate:

- every `SKILL.md` has `name`, `description`, and a clear output contract;
- route skills have decision tables and escalation rules;
- execution skills name deterministic proof requirements;
- lazy references exist and are linked only when needed;
- skill bodies stay under agreed line/token budgets or justify exceptions;
- copied skills match expected global and ATV targets;
- route eval prompts classify to the expected lane.

## Route Eval Seeds

Minimum prompt matrix:

| Prompt Shape | Expected Route |
|---|---|
| "Fix this failing unit test" | `kb-fix` |
| "The UI sometimes loses state; figure it out" | `kb-troubleshoot` |
| "Build this bounded feature; don't ask many questions" | `kb-plan` -> `kb-work` |
| "I have a vague product idea" | `kb-brainstorm` |
| "Migrate auth, billing, and deploy flow" | `kb-epic` |
| "Run this existing manifest" | `kb-work` |
| "Review and finish this diff" | `kb-complete` or `kb-review` depending state |
