---
name: kb-check
description: Deterministic verification harness for KB workflows. Use when code should be tested, linted, typechecked, built, security-checked, or validated by scripts instead of relying on LLM judgment; also use before kb-complete, kb-ship, or after kb-work slices.
argument-hint: "[optional scope, changed files, or command]"
---

# KB Check

Prefer executable truth over model judgment. If a script can check it, run the script.

## Rule

LLM review can find risks, but it does not prove behavior. A slice is not verified until deterministic checks pass or a clear reason is recorded.

When a slice declares protected oracles, deterministic proof must include the
oracle integrity check: the test, fixture, scorer, snapshot, schema, or contract
file used as the behavior target must still match the recorded SHA unless the
plan explicitly updated the oracle. This prevents the model from moving the
target after implementation starts.

## Proof Cadence

Tests stay mandatory, but execution follows three levels:

1. **Slice-local proof** — after a slice's code stabilizes, run the narrowest
   deterministic check that can fail for that slice. Run protected-oracle and
   safety-boundary checks immediately. Do not run the manifest aggregate here.
2. **Proof-batch aggregate** — after a coherent group of dependent slices is
   integrated, run affected integration, functional, smoke, and regression
   checks once for the group. Tightly coupled slices should share this boundary
   instead of replaying the same aggregate after each slice.
3. **Final exact-tree proof** — after review fixes and the last code-affecting
   edit, run one delivery-level aggregate against the exact tree to be shipped.

A passing receipt is reusable across phases, sessions, and worktrees when its
command semantics, relevant-input fingerprint, environment fingerprint, and
tree are unchanged. `REUSE` is proof; do not rerun a command to produce a newer
timestamp, improve provenance, enter another phase, or repeat a summary.

Relevant code, dependency, test-config, generated-contract, environment, merge,
rebase, or conflict-resolution changes invalidate only receipts whose inputs
changed. Docs/status/manifest edits that are not check inputs do not invalidate
code proof. Auth, secrets, destructive data, public contracts, and live/deploy
boundaries still get immediate targeted proof plus final exact-tree proof, but
unchanged receipts are reused between those points.

Do not run the same full suite at slice completion, work completion,
finalization, and shipping. Use `proof-plan` at every phase boundary and execute
only `RUN`; preserve `REUSE`.

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

1. Run `go run ./cmd/kbcheck core --list` when present to inspect discovered commands.
2. Register repeatable commands with their covered checks and relevant inputs,
   then use `kbcheck proof-plan` to select RUN, REUSE, or BLOCK.
3. Execute RUN decisions through `kbcheck proof-run`; do not independently
   replay a command that has a fresh passing receipt or an identical failed
   fingerprint.
4. Run selected checks in this order when available: format/lint, typecheck/static analysis, unit tests, integration/e2e/browser checks, build/package, security/dependency audit.
5. Capture command, exit code, relevant output, and proof receipt.
6. If a check fails, route to `kb-repair` or `kb-fix`; do not ask the user to test normal app behavior.
7. If a check is missing, add a small reusable script or test when practical, then document it in `docs/context/operations/testing.md`.

In this portable skill bundle, the canonical local gate is:

```powershell
go run ./cmd/kbcheck core
```

`cmd/kbcheck` owns top-level orchestration. Existing PowerShell scripts may
still be individual validators until their behavior has separate Go parity
coverage.

For failure-first proof, use the local proof spine:

```powershell
go run ./cmd/kbcheck sense --check <check.json> --trace .kb/trace.jsonl
go run ./cmd/kbcheck accept --check <check.json> --trace .kb/trace.jsonl
go run ./cmd/kbcheck trace-verify --trace .kb/trace.jsonl
```

`accept` passes only when the same check was observed RED and then GREEN, the
trace chain is intact, and the current sensor run is still GREEN. It rejects
vacuous "already green" proof and tampered traces.

For learning changes that claim measurable improvement, use:

```powershell
go run ./cmd/kbcheck learning-adoption --result-path <results.json>
```

The adoption gate requires at least 20 samples, no right-to-wrong regressions,
no holdout string leakage, and either a two-case net gain or a 10 percentage
point gain before a learning rule may be promoted beyond local/scoped use.

## Functional Checks

Use `kb-functional-test` when a change touches user-visible behavior, API/CLI workflows, persistence, auth, streaming, integrations, or any bug that escaped unit tests.

For UI-reachable changes, the check must exercise the rendered UI. Do not substitute a backend/API call, component-handler invocation, mocked request, or direct state assertion for browser proof. If `.tsx`, `.jsx`, `.vue`, or `.svelte` files changed, expect `test_level: functional-browser` and run or call the UI/browser proof path.

Default timing:

- Slice: narrow functional check for the changed path.
- Coherent proof batch: broader smoke tests over the integrated workflows once.
- Final exact tree: full functional/e2e suite when required and practical.
- Ship: reuse the final exact-tree receipt unless shipping changed its inputs.

Headless by default. Do not spawn visible browser windows from multiple workers; serialize browser/e2e checks.

## Script Rule

When the same manual verification would be repeated twice, create a script.

Good scripts accept scope arguments, print concise pass/fail output, exit nonzero on failure, avoid network unless needed, run in CI or from an agent session, and are documented in `docs/context/operations/testing.md`.

For protected-oracle work, prefer reusable SHA/manifest checks over manual
inspection. In this repo, `go run ./cmd/kbcheck skill-eval-manifest-selftest`
proves that tampering with a protected fixture/scorer manifest fails
deterministically.

## Output

Report commands run, pass/fail status, failures fixed or parked, checks added, and remaining manual-only verification with why it cannot be automated.

For every check, include machine proof: command or test file path, exit code, timestamp, and log/artifact path when available. Do not summarize as "tests pass" without the executable proof fields.
