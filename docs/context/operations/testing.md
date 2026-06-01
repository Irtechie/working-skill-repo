# Testing Operations

Checked: 2026-05-31

## Current Commands

```powershell
git status --short
git diff --check
go run .\cmd\kbcheck core --list
go run .\cmd\kbcheck core
go run .\cmd\kbcheck local-release
go test ./...
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
`scripts/skill-marketplace-firebreak.ps1` enforces the quarantine boundary:
active skill roots, approved catalog paths, and loadable skill links must never
resolve into `E:/agent-marketplace/quarantine`. This is a blocking
`kb-check -All` gate, not a naming convention.

The approved `atv-security` marketplace skill also has a dependency
vulnerability proof harness:

```powershell
osv-scanner scan source -r <repo-or-scope-path> --format json --output-file docs/security/osv-YYYY-MM-DD.json
```

Run it when dependency manifests or lockfiles are in scope and `osv-scanner` is
available. If the scanner is not installed, record `skipped-unavailable` and the
official install command instead of using model judgment as proof.

`cmd/kbcheck` discovers this repo as a skill repo when `.github/skills` and
`config/skill-quality.json` exist. Top-level gate orchestration is native Go;
several individual validators are still PowerShell scripts until ported.

The sync drift report is read-only: it never copies files. It still exits
nonzero for required-target drift and is part of the blocking `kb-check -All`
gate. Current policy expects ATV scaffold/plugin copies to match the working
skill root for all tracked skills.

```powershell
pwsh -NoProfile -File scripts\skill-sync-report.ps1
pwsh -NoProfile -File scripts\skill-sync-report.ps1 -VerboseOptional
```

Default output prints required issues in detail and summarizes any optional ATV
scaffold/plugin differences. Current expected state is full match across
working, global, ATV `.github`, ATV scaffold, and ATV plugin skill roots.
Use `-VerboseOptional` when reviewing a deliberate packaging exception.

## Current Result

`go run .\cmd\kbcheck core --list` now reports skill-repo checks when run here:

- `skill-lint`
- `go-test`
- `route-complexity-eval`
- `skill-eval`
- `skill-eval-manifest-selftest`
- `skill-eval-baseline-selftest`
- `skill-eval-codex-dry-run`
- `skill-eval-ghcp-dry-run`
- `skill-eval-observed-trace-dry-run`
- `skill-eval-quality`
- `kb-pipeline-selftest`
- `skill-surface-report`
- `skill-marketplace-firebreak`
- `skill-marketplace-firebreak-selftest`
- `marketplace-promotion-selftest`
- `kb-release-gate-selftest`
- `skill-surface-minimality-selftest`
- `skill-surface-minimality`
- `atv-upstream-delta-selftest`
- `atv-upstream-delta`
- `skill-sync-report`

`go run .\cmd\kbcheck core` runs every discovered check and exits nonzero when a
required check fails.
Expected current warnings:

- hot-path skill size warnings.

All tracked skill roots should report matches.

The release gate is intentionally separate from `core` because it composes the
core check with sync drift, line-ending, and optional report surfaces:

```powershell
go run .\cmd\kbcheck local-release
go run .\cmd\kbcheck live-release
```

`local-release` is the pre-sync proof command. `live-release` is explicit and
may call authenticated Codex/GHCP CLIs; unavailable live surfaces must be labeled
`skipped-explicit`, not silently treated as verified.

Landmine lifecycle changes are verified with the same structural gate:

```powershell
git diff --check
go run .\cmd\kbcheck core
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
- `scripts/skill-eval-quality.ps1` computes deterministic output-quality scores
  from raw captured result JSON for completeness, maintainability-shape,
  relevance, proof quality, and right-sized ceremony.
- `scripts/skill-eval-regression-report.ps1 -RunRoot .atv/eval-runs`
  summarizes local live-run artifacts and compares them to a selected baseline
  when `-BaselinePath` is provided.
- `cmd/kbcheck` composes the native `core`, `local-release`, and `live-release`
  gates. `scripts/kb-release-gate-selftest.ps1` proves profile selection,
  explicit-live skip labeling, and required-check failure propagation through
  the Go path.
- `scripts/go-ps1-parity-report.ps1` recorded the Windows parity proof used to
  retire the old top-level PowerShell wrappers.
- `scripts/kb-pipeline.ps1 -Start skill-bundle-proof-spike` creates a
  non-agent pipeline spike run under `.atv/pipeline-runs/` with selected
  pipeline metadata, phase prompts, protected-file hashes, and proof command
  placeholders. `scripts/kb-pipeline.ps1 -Status` reads the latest run back.
- `scripts/kb-pipeline-selftest.ps1` proves the pipeline spike can start, can
  read status, writes required run artifacts, and rejects unknown pipeline IDs.
- `scripts/skill-surface-report.ps1` reports route-level loaded skill surface
  with line counts, rough token estimates, and content hashes. It can compare
  against a JSON baseline when `-BaselinePath` is provided.
- `scripts/skill-surface-minimality.ps1` statically classifies skills and
  reviewer agents as required, conditional, unproven, unused candidates, or trim
  candidates, with protected repo-policy skills excluded from cold-storage
  deletion candidates. It produces candidates only; it does not approve
  deletion.
- `scripts/atv-upstream-delta.ps1` compares local ATV/fork state to original
  ATV upstream with read-only git diff commands and classifies changed skills as
  KB-owned rejects, shared-overlap reviews, ATV-native candidates, superseded
  workflow rejects, or unknown reviews. Security-sensitive rows, including
  `atv-security`, warn when OSV proof may have changed.
- `scripts/skill-marketplace-firebreak.ps1` enforces the marketplace quarantine
  boundary by failing when any active/approved skill root or approved catalog
  entry resolves into quarantine. Quarantine entries cannot become loadable by
  being marked approved in place.
- `scripts/skill-marketplace-firebreak-selftest.ps1` proves the firebreak trips
  by generating a temporary config that points an active skill root at
  quarantine and requiring a nonzero result.
- `scripts/promote-marketplace-skill-selftest.ps1` proves the one-command
  marketplace promotion path with a temp marketplace: happy path promotion,
  catalog hash pinning, temp global sync, and quarantine destination refusal.
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
