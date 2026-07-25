# Eval Map

Checked: 2026-07-21

## App Pattern

Mixed skill/workflow repo:

- primary surface: portable KB skill bundle under `.github\skills\`
- native proof CLIs: `cmd\kbcheck`, `cmd\kbrouter`, `cmd\amrbench`
- install surfaces: `bin\kb-install.mjs`, `scripts\install-kb.ps1`
- deterministic corpora: `evals\**`

There is no product app runtime. Correctness means repo guidance, gates, and
fixture corpora agree across Codex, GHCP, and shared-agent installs.

## Primary Workflows

| Workflow | Surface | Current Proof | Gap | Priority |
|---|---|---|---|---|
| Skill docs/frontmatter stay valid | `.github/skills/**/SKILL.md` | `go run ./cmd/kbcheck skill-lint` | Some inherited long skills still warn instead of failing | P1 |
| Route selection stays calibrated | `evals/route-complexity/*.json` | `go run ./cmd/kbcheck route-eval` | Deterministic fixtures only; not live prompt runs | P0 |
| Installer surface stays healthy | `package.json`, `bin/kb-install.mjs`, `scripts/install-kb.ps1` | `npm run test`; `npm run test:install:core`; `npm run test:install:full` | No checked-in CI workflow runs these automatically | P1 |
| Required skill copies stay synced | working repo + global skill roots | `go run ./cmd/kbcheck local-release`; `go run ./cmd/kbcheck skill-sync-report`; `go run ./cmd/kbcheck doctor` | Depends on local install availability | P1 |
| Skill edits do not regress routing/proof behavior | `evals/skill-eval/**` | `go run ./cmd/kbcheck skill-eval`; `skill-eval-claims`; `skill-eval-quality`; `skill-eval-regression` | Live corpus is still narrow | P0 |
| Live adapter wrappers stay safe | Codex/GHCP adapters | `go run ./cmd/kbcheck eval-run-codex --fixture-id tiny-typo-fix --dry-run`; `go run ./cmd/kbcheck eval-run-ghcp --fixture-id tiny-typo-fix --dry-run`; `go run ./cmd/kbcheck skill-eval-wrap --fixture-id tiny-typo-fix --dry-run --sealed` | Authenticated live runs remain explicit | P0 |
| Repair claims prove RED then GREEN | `.kb/trace.jsonl` and check specs | `go run ./cmd/kbcheck sense`; `trace-verify`; `accept` | Per-slice specs are created as needed, not centrally cataloged | P0 |
| Manifests cannot self-report done | `docs/plans/*manifest*.md` | `go run ./cmd/kbcheck manifest-contract --manifest <manifest>` | Does not execute every recorded proof command yet | P0 |
| False completion is rejected | `evals/dishonest-completion/fixtures.json` | `go run ./cmd/kbcheck dishonest-completion-selftest` | Small corpus only | P0 |
| Run-state loops stop | `.kb/runs/<goal>/route-history.jsonl` | `go run ./cmd/kbcheck run-state --history <history>`; `go run ./cmd/kbcheck run-state-selftest` | Needs more real histories over time | P1 |
| Learning promotion stays evidence-bound | learning result JSON | `go run ./cmd/kbcheck learning-adoption --result-path <results.json>` | More real samples needed | P1 |
| Worker context and telemetry stay bounded | context packet + telemetry JSON | `go run ./cmd/kbcheck context-packet --packet <packet.json>`; `go run ./cmd/kbcheck execution-telemetry --telemetry <telemetry.json>` | Host adapters still expose measured usage inconsistently | P1 |
| Optional providers remain optional | repo/user provider config | `go run ./cmd/kbcheck provider-hygiene`; `go run ./cmd/kbcheck provider-hygiene --include-user` | Host-specific registries may need future adapters | P1 |
| Graph routing stays promotion-safe | `config/graph-route.schema.json`; `evals/graph-routing/**` | `go run ./cmd/kbcheck graph-route --packet <packet.json>`; `graph-routing-lifecycle-selftest`; `graph-routing-eval --require-ready` | Fixture proof only; live providers remain optional | P0 |
| Model-routing release claims stay inside evidence | `evals/model-routing/*.json`; release artifact JSON | `go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json` | Deterministic/no-paid evidence only; not promoted | P0 |
| GHCP AMR benchmark stays safe before spend | `cmd/amrbench`; `evals/amr-model-benchmark/**` | `go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json`; `go run ./cmd/amrbench grade-paired --results <paired-results.jsonl>` | Attended runs remain approval-gated and currently disabled | P0 |

## Existing Harnesses

### Package / installer

- `npm run test`
- `npm run test:install:core`
- `npm run test:install:full`

### KB gate / release / sync

- `go run ./cmd/kbcheck core`
- `go run ./cmd/kbcheck local-release`
- `go run ./cmd/kbcheck live-release`
- `go run ./cmd/kbcheck skill-sync-report`
- `go run ./cmd/kbcheck doctor`
- `go run ./cmd/kbcheck doctor-selftest`
- `git diff --check`

### Workflow proof / state

- `go run ./cmd/kbcheck manifest-contract --manifest <manifest>`
- `go run ./cmd/kbcheck manifest-contract-selftest`
- `go run ./cmd/kbcheck sense --check <check.json> --trace .kb/trace.jsonl`
- `go run ./cmd/kbcheck trace-verify --trace .kb/trace.jsonl`
- `go run ./cmd/kbcheck accept --check <check.json> --trace .kb/trace.jsonl`
- `go run ./cmd/kbcheck run-state --history <history>`
- `go run ./cmd/kbcheck run-state-selftest`
- `go run ./cmd/kbcheck ready-set-selftest`
- `go run ./cmd/kbcheck scope-lease-selftest`
- `go run ./cmd/kbcheck slice-lease-selftest`
- `go run ./cmd/kbcheck workflow-governor-selftest`

### Skill / route / claim evals

- `go run ./cmd/kbcheck route-eval`
- `go run ./cmd/kbcheck skill-eval`
- `go run ./cmd/kbcheck skill-eval-claims`
- `go run ./cmd/kbcheck skill-eval-quality`
- `go run ./cmd/kbcheck skill-eval-regression`
- `go run ./cmd/kbcheck eval-run-codex --fixture-id tiny-typo-fix --dry-run`
- `go run ./cmd/kbcheck eval-run-ghcp --fixture-id tiny-typo-fix --dry-run`
- `go run ./cmd/kbcheck eval-run-live-corpus --dry-run`
- `go run ./cmd/kbcheck skill-eval-wrap --fixture-id tiny-typo-fix --dry-run --sealed`

### Graph / model routing / benchmark

- `go run ./cmd/kbcheck graph-route --packet <packet.json>`
- `go run ./cmd/kbcheck graph-routing-lifecycle-selftest`
- `go run ./cmd/kbcheck graph-routing-eval --require-ready`
- `go test ./cmd/kbrouter -run Catalog|Doctor|Policy`
- `go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json`
- `go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json`
- `go run ./cmd/amrbench grade-paired --results <paired-results.jsonl>`

## Canonical Commands

```powershell
npm run test
go build ./...
go vet ./...
go run ./cmd/kbcheck core
go run ./cmd/kbcheck local-release
go run ./cmd/kbcheck route-eval
go run ./cmd/kbcheck skill-eval
go run ./cmd/kbcheck graph-routing-eval --require-ready
go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json
go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json
git diff --check
```

## Scaffolding Decisions

No new smoke eval was added in this bootstrap refresh. This repo already has
real harnesses for its highest-value workflows:

- skill lint / route-calibration / claim-quality scoring
- deterministic graph-routing readiness
- model-routing release validation
- no-paid AMR conformance
- installer tests in Node

Creating placeholder browser/API smoke tests would not prove anything real here.

## Deterministic vs LLM-Judged

| Check | Class |
|---|---|
| npm installer tests | deterministic |
| skill lint | deterministic |
| route-complexity fixture scoring | deterministic |
| captured skill result scoring | deterministic |
| proof-spine trace acceptance | deterministic |
| manifest done/proof contract | deterministic schema and gate check |
| dishonest completion rejection fixtures | deterministic negative selftest |
| run-state loop guard | deterministic JSONL check |
| doctor install-drift repair/refusal | deterministic fixture selftest |
| graph-route packet validation | deterministic schema/contract check |
| graph-routing lifecycle and readiness | deterministic |
| model-routing release evidence validation | deterministic |
| AMR no-paid conformance | deterministic |
| Codex/GHCP live adapters | mixed: model action plus deterministic scoring |

## Credentials / Session Requirements

- Deterministic checks above require no credentials.
- `eval-run-codex` live mode needs a working/authenticated `codex` CLI.
- `eval-run-ghcp` live mode needs a working/authenticated `copilot` CLI.
- `cmd/amrbench run` non-dry remains unavailable until a trusted approval
  verifier exists.

## Dashboard / Export Options

Keep local fixture JSON/Markdown and `kbcheck`/`amrbench` output as source of
truth. Langfuse, Braintrust, LangSmith, Promptfoo, or DeepEval remain optional
exporters/adapters, not the current judges.

## Open Eval Gaps

- Broaden the live Codex/GHCP corpus beyond the current route fixture set.
- Capture normalized real token/cache/turn usage from live adapters.
- Decide whether release/install proof should stay local-only or gain checked-in
  CI workflows; `.github/workflows/` is currently empty.
- Make `manifest-contract` optionally execute recorded `proof_check` commands
  once the schema is stable.
