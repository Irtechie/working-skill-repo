# Eval Map

Checked: 2026-06-01

## App Pattern

Skill/workflow repo. There is no app runtime. Correctness is defined by skill
structure, route-complexity fixtures, sync drift, and whether repo guidance gives
Codex/GHCP the same workflow contract.

## Primary Workflows

| Workflow | Surface | Current Proof | Gap | Priority |
|---|---|---|---|---|
| Skill structure remains valid | `.github/skills/**/SKILL.md` | `go run ./cmd/kbcheck skill-lint` | Warnings remain for inherited older skills | P1 |
| Route complexity stays calibrated | `evals/route-complexity/*.json` | `go run ./cmd/kbcheck route-eval` | Fixtures are deterministic metadata, not live prompt runs; workflow-shape fixtures cover skill edit, skill-bundle, proof pipeline, and multi-stream epic prompts | P0 |
| Required skill copies stay synced | global installs and ATV `.github` skills | `go run ./cmd/kbcheck local-release`; `go run ./cmd/kbcheck skill-sync-report` | ATV scaffold/plugin shipping policy unresolved | P1 |
| Skill edits do not regress behavior | prompt/trace/claim evals | `go run ./cmd/kbcheck skill-eval`; `go run ./cmd/kbcheck eval-run-codex`; `go run ./cmd/kbcheck eval-run-ghcp` | Need broader live corpus and richer trace/claim scoring | P0 |

## Existing Harnesses

- `go run ./cmd/kbcheck core`
- `go run ./cmd/kbcheck local-release`
- `go run ./cmd/kbcheck skill-lint`
- `go run ./cmd/kbcheck route-eval`
- `go run ./cmd/kbcheck skill-eval`
- `go run ./cmd/kbcheck eval-run-codex --fixture-id tiny-typo-fix --dry-run`
- `go run ./cmd/kbcheck eval-run-ghcp --fixture-id tiny-typo-fix --dry-run`
- `go run ./cmd/kbcheck eval-run-live-corpus --dry-run`
- `go run ./cmd/kbcheck skill-eval-claims`
- `go run ./cmd/kbcheck skill-eval-quality`
- `go run ./cmd/kbcheck skill-eval-regression`
- `go run ./cmd/kbcheck skill-sync-report`
- `git diff --check`

## Canonical Commands

```powershell
go run ./cmd/kbcheck core
go run ./cmd/kbcheck local-release
git diff --check
```

For touched ATV repo copies:

```powershell
git -C <atv-repo> diff --check
```

## Scaffolding Decisions

No new smoke eval is scaffolded by `kb-eval-map` for this repo yet because the
next useful smoke is not a placeholder; it is the planned live skill eval suite:
prompt routing, trace capture, claim verification, output quality scoring, and
cost telemetry.

The deterministic scorer exists as `go run ./cmd/kbcheck skill-eval`. It scores
captured agent result JSON against route fixtures and claim checks, and its
self-test includes intentionally bad route/proof/claim outputs that must fail.
The Codex and GHCP adapters exist as `eval-run-codex` and `eval-run-ghcp`;
dry-run mode is included in `core`, while live mode explicitly invokes
authenticated local CLIs. `eval-run-live-corpus` runs selected fixtures across
Codex and GHCP adapters and summarizes pass/fail/skip categories.
`skill-eval-claims` checks transcript-derived claim artifacts deterministically
and reports ambiguous claims without counting them as proof.
`skill-eval-quality` scores output-quality fixtures separately from
deterministic route/proof/claim pass/fail. `skill-eval-regression` summarizes
local live-run artifacts and compares pass/non-pass plus size/time proxies
against selected baselines.

## Deterministic vs LLM-Judged

| Check | Class |
|---|---|
| skill lint | deterministic |
| route-complexity fixture scoring | deterministic |
| captured skill result scoring | deterministic |
| sync drift hashes | deterministic |
| git whitespace/conflict checks | deterministic |
| output quality scoring | deterministic rubric selftest |
| Codex skill prompt routing live run | mixed: model action plus deterministic route/trace scoring |
| final claim verification | deterministic filesystem/git/log checks |

## Credentials / Session Requirements

None for current deterministic checks.

The Codex adapter requires a working `codex` CLI and authenticated Codex session.
The GHCP adapter requires a working `copilot` CLI and authenticated Copilot
session. The deterministic scorer can run without those sessions when given
captured result JSON.

## Dashboard / Export Options

Keep local scripts as source of truth. Langfuse, Braintrust, LangSmith,
Promptfoo, or DeepEval can be adapters/exporters after local fixtures and scores
are stable.

## Open Eval Gaps

- Grow the live Codex/GHCP corpus beyond the current route fixture set.
- Optional exporters can be added for Langfuse, Braintrust, LangSmith,
  Promptfoo, or DeepEval after local JSON/Markdown reports remain stable.
