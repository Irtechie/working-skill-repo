# Testing Operations

Checked: 2026-05-31

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

The private marketplace contract lives in `config/skill-marketplace.json`. It
records `E:/agent-marketplace` as the approved catalog root and defines the
project-local-first promotion policy for learned skills and reusable pipelines.

`kb-check.ps1` discovers this repo as a skill repo when `.github/skills` and
`config/skill-quality.json` exist.

The sync drift report is read-only: it never copies files. It still exits
nonzero for required-target drift and is part of the blocking `kb-check -All`
gate. Optional ATV scaffold/plugin differences remain warnings until their
shipping contract is decided.

```powershell
powershell -ExecutionPolicy Bypass -File scripts\skill-sync-report.ps1
powershell -ExecutionPolicy Bypass -File scripts\skill-sync-report.ps1 -VerboseOptional
```

Default output prints required issues in detail and summarizes optional ATV
scaffold/plugin differences. Use `-VerboseOptional` when reviewing optional
per-skill drift.

## Current Result

`kb-check.ps1 -List` now reports skill-repo checks when run here:

- `skill-lint`
- `route-complexity-eval`
- `skill-eval`
- `skill-eval-manifest-selftest`
- `skill-eval-baseline-selftest`
- `skill-eval-codex-dry-run`
- `skill-eval-ghcp-dry-run`
- `skill-eval-quality`
- `kb-pipeline-selftest`
- `skill-surface-report`
- `skill-sync-report`

`kb-check.ps1 -All` runs every discovered check and exits nonzero when a
required check fails.
Expected current warnings:

- hot-path skill size warnings;
- summarized optional ATV scaffold/plugin drift or omissions.

Required targets should report zero required sync issues.

Landmine lifecycle changes are verified with the same structural gate:

```powershell
git diff --check
.\.github\skills\kb-check\scripts\kb-check.ps1 -All
```

`docs/context/landmines.md` must contain owner, evidence, fix condition, and
verification fields for active entries. Generic advice should not be recorded as
an active landmine.

Generated `learned-*` skills require explicit human approval before they are
committed or synced from this portable bundle. The default prompt is:

```text
Promote these generated learned-* skills and allow them to be committed or
synced from this portable bundle? yes/no
```

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

## Implemented Harness

- `scripts/skill-lint.ps1` validates required skill frontmatter, conflict
  markers, configured line budgets, allowlisted long skills, and referenced
  local files.
- `scripts/route-complexity-eval.ps1` validates the route-complexity fixture
  schema, computes deterministic complexity tiers, and checks over/under
  planning guard coverage.
- `scripts/skill-eval.ps1` scores captured skill result JSON against route
  fixtures, trace evidence, and structured claim checks. Its default self-test
  includes intentionally bad route/proof/claim results that must fail. When
  passed `-ManifestPath`, it also verifies protected verifier-file SHA256 hashes
  before accepting the result.
- `scripts/skill-eval-manifest-selftest.ps1` creates a dry-run eval manifest,
  proves the valid manifest passes, then corrupts the fixture SHA in a copied
  manifest and requires deterministic scoring to fail.
- `scripts/skill-eval-baseline-selftest.ps1` creates a temporary baseline from
  selftest results, proves unchanged comparison passes, removes proof from a
  copied passing result, and mutates a negative fixture so it incorrectly
  passes. Both mutations must fail baseline comparison.
- `scripts/skill-eval-run-codex.ps1 -FixtureId tiny-typo-fix -DryRun` validates
  the Codex live-adapter plumbing without calling a model and writes a
  protected-file hash manifest. Live mode is explicit because it invokes
  `codex exec`.
- `scripts/skill-eval-run-ghcp.ps1 -FixtureId tiny-typo-fix -DryRun` validates
  the GHCP live-adapter plumbing without calling a model and writes a
  protected-file hash manifest. Live mode is explicit because it invokes GitHub
  Copilot CLI and relies on prompt-level JSON constraints plus deterministic
  parsing.
- `scripts/skill-eval-run-live-corpus.ps1 -All -Runtime codex,ghcp -DryRun`
  validates corpus orchestration across both adapters. Live mode is explicit and
  not part of `kb-check -All`.
- `scripts/skill-eval-claims.ps1` self-tests transcript-derived claim artifacts:
  true deterministic claims pass, false deterministic claims fail, and ambiguous
  claims are reported without becoming proof.
- `scripts/skill-eval-quality.ps1` self-tests output-quality rubric scoring for
  completeness, maintainability, relevance, proof quality, and right-sized
  ceremony.
- `scripts/skill-eval-regression-report.ps1 -RunRoot .atv/eval-runs`
  summarizes local live-run artifacts and compares them to a selected baseline
  when `-BaselinePath` is provided.
- `scripts/kb-pipeline.ps1 -Start skill-bundle-proof-spike` creates a
  non-agent pipeline spike run under `.atv/pipeline-runs/` with selected
  pipeline metadata, phase prompts, protected-file hashes, and proof command
  placeholders. `scripts/kb-pipeline.ps1 -Status` reads the latest run back.
- `scripts/kb-pipeline-selftest.ps1` proves the pipeline spike can start, can
  read status, writes required run artifacts, and rejects unknown pipeline IDs.
- `scripts/skill-surface-report.ps1` reports route-level loaded skill surface
  with line counts, rough token estimates, and content hashes. It can compare
  against a JSON baseline when `-BaselinePath` is provided.
- `scripts/skill-sync-report.ps1` validates required skill-copy hashes across
  the working repo, Codex global, Copilot global, shared agents global, and ATV
  `.github` skills.

## Remaining Harness Growth

The current harness covers the core planned eval stack. Remaining growth areas:

- route skills have decision tables and escalation rules;
- execution skills name deterministic proof requirements;
- lazy references exist and are linked only when needed;
- broader live Codex/GHCP corpus covers more than the initial fixture;
- optional dashboard exporters can be added after local reports stay stable.

`kb-eval-map` scaffolded smoke evals must record pass-command, pass-result,
negative-check, negative-command, negative-result=failed-as-expected, and
reverted=true evidence before reporting success.

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
