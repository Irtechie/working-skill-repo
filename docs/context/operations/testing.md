# Testing Operations

Checked: 2026-07-21

## Fast Contributor Commands

```powershell
git status --short
git diff --check
npm run test
go build ./...
go vet ./...
go run ./cmd/kbcheck core --list
go run ./cmd/kbcheck core
```

Use this set for ordinary repo-local validation.

## Canonical Release / Sync Commands

```powershell
go run ./cmd/kbcheck local-release
go run ./cmd/kbcheck live-release
go run ./cmd/kbcheck skill-sync-report
go run ./cmd/kbcheck skill-sync-report --verbose-optional
go run ./cmd/kbcheck doctor
go run ./cmd/kbcheck doctor --fix
```

`core` is contributor-safe. `local-release` adds blocking sync drift and release
surfaces. `live-release` is explicit because it may call authenticated CLIs.

## `kbcheck core --list` Current Checks

```text
test
go-test
context-packet-selftest
cross-model-benchmark-validate
dishonest-completion-selftest
execution-telemetry-selftest
graph-routing-eval
graph-routing-lifecycle-selftest
kb-doctor-selftest
kb-pipeline-selftest
kb-release-gate-selftest
kb-run-state-selftest
kb-work-ready-set-selftest
kb-work-scope-lease-selftest
kb-work-slice-lease-selftest
kbrouter-catalog-tests
manifest-contract-selftest
marketplace-promotion-selftest
provider-hygiene
provider-hygiene-selftest
review-reference-guard
route-complexity-eval
skill-eval
skill-eval-baseline-selftest
skill-eval-codex-dry-run
skill-eval-ghcp-dry-run
skill-eval-manifest-selftest
skill-eval-observed-trace-dry-run
skill-eval-quality
skill-lint
skill-marketplace-firebreak
skill-marketplace-firebreak-selftest
skill-surface-minimality
skill-surface-minimality-selftest
skill-surface-report
workflow-governor-selftest
```

## Targeted Commands By Surface

### Package / installer

```powershell
npm run test
npm run test:install:core
npm run test:install:full
node ./bin/check-release-tag.mjs --tag v<package-version>
```

### Workflow proof and state

```powershell
go run ./cmd/kbcheck manifest-contract --manifest <manifest>
go run ./cmd/kbcheck run-state --history <history>
go run ./cmd/kbcheck sense --check <check.json> --trace .kb/trace.jsonl
go run ./cmd/kbcheck trace-verify --trace .kb/trace.jsonl
go run ./cmd/kbcheck accept --check <check.json> --trace .kb/trace.jsonl
go run ./cmd/kbcheck learning-adoption --result-path <results.json>
go run ./cmd/kbcheck context-packet --packet cmd\kbcheck\testdata\context-packet-valid.json
go run ./cmd/kbcheck execution-telemetry --telemetry cmd\kbcheck\testdata\execution-telemetry-valid.json
```

### Graph / model routing

```powershell
go run ./cmd/kbcheck graph-route --packet <packet.json>
go run ./cmd/kbcheck graph-routing-lifecycle-selftest
go run ./cmd/kbcheck graph-routing-eval --require-ready
go test ./cmd/kbrouter -run Catalog|Doctor|Policy
go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json
```

### Skill / route evals

```powershell
go run ./cmd/kbcheck route-eval
go run ./cmd/kbcheck skill-eval
go run ./cmd/kbcheck skill-eval-claims
go run ./cmd/kbcheck skill-eval-quality
go run ./cmd/kbcheck skill-eval-regression
go run ./cmd/kbcheck eval-run-codex --fixture-id tiny-typo-fix --dry-run
go run ./cmd/kbcheck eval-run-ghcp --fixture-id tiny-typo-fix --dry-run
go run ./cmd/kbcheck eval-run-live-corpus --dry-run
go run ./cmd/kbcheck skill-eval-wrap --fixture-id tiny-typo-fix --dry-run --sealed
```

### AMR benchmark

```powershell
go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json
go run ./cmd/amrbench run --dry-run --experiment-id <id> --routes <routes.json> --context baseline --mode direct --task <task> --model <runtime-model>
go run ./cmd/amrbench grade-paired --results <paired-results.jsonl>
```

## Skill Quality Contract

`config/skill-quality.json` is the source of truth for:

- supported instruction surfaces
- skill lint budgets and long-skill allowlists
- route-complexity fixture roots
- review-reference ownership
- required and optional sync targets

`config/skill-marketplace.json` is the source of truth for marketplace
promotion, quarantine firebreak rules, and global-install boundaries.

## Operational Notes

- `skill-sync-report` is read-only. If a global copy is newer, merge back into
  the repo first; do not overwrite it blindly.
- `provider-hygiene --include-user` is the explicit user-global inspection mode.
  Plain `provider-hygiene` remains repo-local.
- `graph-routing-eval --require-ready` is fixture proof, not live-provider proof.
- `model-routing-release` validates evidence boundaries; a green result does not
  mean live AMR is promoted.
- `.github/workflows\` is currently empty, so local CLI commands are the only
  checked-in deterministic proof surface for release/install behavior.

## Eval Mapping

`kb-eval-map` owns repo-native eval documentation in this repo:

- `docs/context/eval-map.md` = inventory + app-pattern classification
- `docs/context/operations/testing.md` = canonical commands

## Route Eval Seeds

| Prompt Shape | Expected Route |
|---|---|
| "Fix this failing unit test" | `kb-fix` |
| "The UI sometimes loses state; figure it out" | `kb-troubleshoot` |
| "Build this bounded feature; don't ask many questions" | `kb-plan` -> `kb-work` |
| "I have a vague product idea" | `kb-brainstorm` |
| "Migrate auth, billing, and deploy flow" | `kb-epic` |
| "Run this existing manifest" | `kb-work` |
| "Review and finish this diff" | `kb-complete` or `kb-review` depending state |
