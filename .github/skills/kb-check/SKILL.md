---
name: kb-check
description: Deterministic verification harness for KB workflows. Use when code should be tested, linted, typechecked, built, security-checked, or validated by scripts instead of relying on LLM judgment; also use before kb-complete, kb-ship, or after kb-work slices.
argument-hint: "[optional scope, changed files, or command]"
---

# KB Check

Prefer executable truth over model judgment. If a script can check it, run the script.

## Rule

LLM review can find risks, but it does not prove behavior. A slice is not verified until deterministic checks pass or a clear reason is recorded.

`cmd/kbcheck` belongs to this bundle's source repo and does not ship with an
installed skill. Every `go run ./cmd/kbcheck ...` command below is therefore
conditional: run it when the repo provides it, and otherwise substitute the
project's own equivalent test, lint, typecheck, or build command and record
which command produced the proof. A missing harness never lowers the bar — it
changes which command you run, not whether you prove the slice.

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

## Cargo Build Storage

Before any selected command invokes Cargo, use the native resolver only after a
capability probe confirms the project's `cmd/kbcheck help` output includes
`cargo-storage`:

```powershell
go run ./cmd/kbcheck cargo-storage --action resolve --run-id <run-id> --root <project-root> --json
go run ./cmd/kbcheck cargo-storage --action validate-ready --run-id <run-id> --root <project-root> --json
```

The native resolver derives one collision-resistant target from canonical
repository identity, treats an external absolute `CARGO_TARGET_DIR` as a cache
root keyed by that identity, shares the result across linked worktrees,
fingerprints Cargo config, serializes receipt updates, and returns the exact
`CARGO_TARGET_DIR` to apply. Its receipt under the Git common directory is the
only build-storage handoff between checks, workers, sessions, and phases.

If the capability probe fails because `cmd/kbcheck` is absent or predates
`cargo-storage`, use the portable fail-closed fallback: require an existing
absolute external
`CARGO_TARGET_DIR`, treat it as a cache root, append the first 24 lowercase hex
characters of SHA-256 over the canonical absolute Git common-directory path,
create that project-keyed child, and apply it unchanged to every Cargo command.
If the configured path already ends with that exact project key, reuse it
instead of appending the key again.
Use the host's built-in SHA-256 utility; if canonicalization or hashing is
unavailable, block Cargo rather than inventing a target. Record
`portable-fallback`, the canonical identity, and the exact applied path in the
workflow state. Temporary targets and automated deletion are prohibited in
fallback mode, so finalization records retained bytes and zero removed bytes.

Never create check-, repair-, reproduction-, probe-, worker-, slice-, or
run-specific targets such as `target-check`, `target-repair`, `target-repro`,
`release-api-probe-target`, or a `cargo-target` directory under an agent temp
root. A new target forces Cargo to rebuild the dependency graph and is not test
isolation. Do not run `cargo clean` against the stable shared target while
another consumer may be active.

A temporary target is allowed only in native mode after the native
`register-temp` action creates an ownership marker under an approved temporary
root for a documented technical incompatibility. Its basename must be
`kb-cargo-temp-<24-lowercase-hex>` so phase, worker, slice, and run labels
cannot masquerade as approved isolation. Finalization owns deletion through the
same native guard. Never delete or rotate the stable shared target as routine
KB cleanup.

When the selected proof set contains no Cargo command and native `cmd/kbcheck`
is present, create the machine-validated terminal state with
`cargo-storage --action not-applicable --run-id <run-id> --reason <reason>`.

## Workflow

1. Run `go run ./cmd/kbcheck core --list` when present to inspect discovered commands.
2. Register repeatable commands with their covered checks and relevant inputs,
   then use `kbcheck proof-plan` to select RUN, REUSE, or BLOCK.
3. Execute RUN decisions through `kbcheck proof-run`; do not independently
   replay a command that has a fresh passing receipt or an identical failed
   fingerprint.
4. Run selected checks in this order when available: format/lint, typecheck/static analysis, unit tests, integration/e2e/browser checks, build/package, security/dependency audit.
5. Capture command, exit code, relevant output, proof receipt, and the resolved
   Cargo target path when Cargo ran.
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
