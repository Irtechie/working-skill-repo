# Testing Operations

Checked: 2026-07-09

## Current Commands

```powershell
git status --short
git diff --check
go run ./cmd/kbcheck core --list
go run ./cmd/kbcheck core
go run ./cmd/kbcheck local-release
go run ./cmd/kbcheck dishonest-completion-selftest
go run ./cmd/kbcheck manifest-contract --manifest <manifest>
go run ./cmd/kbcheck run-state --history <history>
go run ./cmd/kbcheck doctor
go run ./cmd/kbcheck sense --check <check.json> --trace .kb/trace.jsonl
go run ./cmd/kbcheck accept --check <check.json> --trace .kb/trace.jsonl
go run ./cmd/kbcheck trace-verify --trace .kb/trace.jsonl
go run ./cmd/kbcheck learning-adoption --result-path <results.json>
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
records `<agent-marketplace>` as the approved catalog root and defines the
project-local-first promotion policy for learned skills and reusable pipelines.
`go run ./cmd/kbcheck marketplace-firebreak` enforces the quarantine boundary:
active skill roots, approved catalog paths, and loadable skill links must never
resolve into `<agent-marketplace>/quarantine`. This is a blocking `core` gate,
not a naming convention.

The approved `atv-security` marketplace skill also has a dependency
vulnerability proof harness:

```powershell
osv-scanner scan source -r <repo-or-scope-path> --format json --output-file docs/security/osv-YYYY-MM-DD.json
```

Run it when dependency manifests or lockfiles are in scope and `osv-scanner` is
available. If the scanner is not installed, record `skipped-unavailable` and the
official install command instead of using model judgment as proof.

`cmd/kbcheck` discovers this repo as a skill repo when `.github/skills` and
`config/skill-quality.json` exist. The skill-repo quality harness is Go-native.

The sync drift report is read-only: it never copies files. It exits nonzero for
required-target drift and is part of the blocking `local-release` gate, not the
fresh-clone contributor `core` gate.
Current policy expects required roots to match the working skill root. ATV
scaffold/plugin roots may report warning-only differences unless the current
change explicitly ships that packaging surface.

```powershell
go run ./cmd/kbcheck skill-sync-report
go run ./cmd/kbcheck skill-sync-report --verbose-optional
```

Default output prints required issues in detail and summarizes any optional ATV
scaffold/plugin differences. Current expected state is full match across
working, global, and ATV `.github` roots; ATV scaffold/plugin roots may report
warning-only differences unless the current change explicitly ships that
surface. Use `--verbose-optional` when reviewing a deliberate packaging
exception.

## Current Result

2026-07-09 local caveat resolved: on this Windows workstation with `go version
go1.26.2 windows/amd64`, module-scoped Go commands timed out with no output:

```powershell
go list ./cmd/kbcheck
go list -buildvcs=false ./cmd/kbcheck
go test ./cmd/kbcheck -run TestProofAcceptsRedThenGreenTrace -count=1
go test -buildvcs=false ./cmd/kbcheck -run TestProofAcceptsRedThenGreenTrace -count=1
go run ./cmd/kbcheck core --list
```

The workaround was:

```powershell
go env -w GOTOOLCHAIN=go1.25.4+auto
```

After that, `go list ./cmd/kbcheck`, the targeted proof-spine test, and
`go run ./cmd/kbcheck core --list` returned normally.

If the timeout recurs, narrow non-Go proof for snapshot-path only changes is:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .github\skills\kb-regression-snapshot\scripts\kb-regression-snapshot.ps1 verify
rg -n "\.atv/snapshots|\.atv\\snapshots" .github\skills README.md docs\context\PROJECT.md docs\context\architecture\kb-workflow.md
git diff --check
```

Do not expand `cmd/kbcheck` harness behavior while module-scoped Go commands are
timing out; first fix the toolchain or use a bounded replacement runner.

`go run ./cmd/kbcheck core --list` now reports contributor-safe skill-repo
checks when run here:

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
- `dishonest-completion-selftest`
- `kb-doctor-selftest`
- `kb-run-state-selftest`
- `manifest-contract-selftest`
`go run ./cmd/kbcheck core` runs every discovered check and exits nonzero when a
required check fails.
Expected current warnings:

- hot-path skill size warnings.

All tracked skill roots should report matches before release or propagation.

The release gate is intentionally separate from `core` because it composes the
core check with sync drift, line-ending, and optional report surfaces:

```powershell
go run ./cmd/kbcheck local-release
go run ./cmd/kbcheck live-release
```

`local-release` is the pre-sync proof command and includes
`skill-sync-report`. `live-release` is explicit and
may call authenticated Codex/GHCP CLIs; unavailable live surfaces must be labeled
`skipped-explicit`, not silently treated as verified.

Landmine lifecycle changes are verified with the same structural gate:

```powershell
git diff --check
go run ./cmd/kbcheck core
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

All current harness commands are native `cmd/kbcheck` commands:

- `skill-lint` validates required skill frontmatter, conflict markers,
  configured line budgets, allowlisted long skills, and referenced local files.
- `route-eval` validates route-complexity fixtures, computes deterministic
  complexity tiers, and checks over/under-planning guard coverage.
- `skill-eval`, `skill-eval-claims`, `skill-eval-quality`, and
  `skill-eval-regression` score captured result JSON, trace evidence, claim
  artifacts, rubric quality, baselines, and protected verifier-file hashes.
- `eval-run-codex`, `eval-run-ghcp`, `eval-run-live-corpus`, and
  `skill-eval-wrap` provide dry-run/live adapters and observed-trace wrapping.
- `release-selftest` proves profile selection, explicit-live skip labeling, and
  required-check failure propagation through the Go release path.
- `pipeline` and `pipeline-selftest` create/read coded pipeline spike runs
  under `.atv/pipeline-runs/` and reject unknown pipeline IDs.
- `ready-set` and `scope-lease` prove KB manifest DAG ready-set selection and
  bounded-swarm write-overlap guards.
- `surface-report` and `minimality` report loaded skill surface and static
  skill/agent minimality classifications.
- `atv-delta` compares local ATV/fork state to original ATV upstream with
  read-only git diff commands and classifies changed skills as KB-owned rejects,
  shared-overlap reviews, ATV-native candidates, superseded workflow rejects, or
  unknown reviews.
- `marketplace-firebreak`, `marketplace-firebreak-selftest`,
  `marketplace-promote`, and `marketplace-promote-selftest` enforce quarantine
  boundaries and prove the reviewed promotion path.
- `skill-sync-report` validates required skill-copy hashes across the working
  repo, Codex global, Copilot global, shared agents global, and ATV `.github`
  skills; it is release-blocking through `local-release`.
- `doctor` reports the same configured install drift and `doctor --fix` repairs
  only missing or marker-proven stale required copies from this repo source.
- `dishonest-completion-selftest` validates negative fixtures for vacuous-green
  proof, missing slice proof checks, invalid model routes, and route-history
  oscillation.
- `manifest-contract` validates objective `done_check`, per-slice
  `proof_check`/`no_check_reason`, model routes, and terminal gate proof.
- `run-state` validates `.kb/runs/<goal>/route-history.jsonl` for oscillation,
  repeated low confidence, and no-progress loops.
- `sense`, `accept`, and `trace-verify` implement the failure-first proof spine:
  a runnable check must be observed RED, then GREEN, with an intact trace before
  a repair claim is accepted.
- `learning-adoption` scores measured learning changes and blocks promotion
  unless the candidate has enough samples, no right-to-wrong regressions, no
  holdout leakage, and a meaningful net gain.

For unattended runners, `skill-sync-report` is a release blocker, not a cleanup
task. Required drift means source and deployed runner behavior disagree. If a
global copy is newer, merge useful drift back into the repo first, prove it, and
only then sync outward from the repo source.

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
