# Eval Map

Checked: 2026-05-30

## App Pattern

Skill/workflow repo. There is no app runtime. Correctness is defined by skill
structure, route-complexity fixtures, sync drift, and whether repo guidance gives
Codex/GHCP the same workflow contract.

## Primary Workflows

| Workflow | Surface | Current Proof | Gap | Priority |
|---|---|---|---|---|
| Skill structure remains valid | `.github/skills/**/SKILL.md` | `scripts/skill-lint.ps1` | Warnings remain for inherited older skills | P1 |
| Route complexity stays calibrated | `evals/route-complexity/*.json` | `scripts/route-complexity-eval.ps1` | Fixtures are deterministic metadata, not live prompt runs; workflow-shape fixtures cover skill edit, skill-bundle, proof pipeline, and multi-stream epic prompts | P0 |
| Required skill copies stay synced | global installs and ATV `.github` skills | `scripts/skill-sync-report.ps1` | ATV scaffold/plugin shipping policy unresolved | P1 |
| Skill edits do not regress behavior | prompt/trace/claim evals | `scripts/skill-eval.ps1`; `scripts/skill-eval-run-codex.ps1`; `scripts/skill-eval-run-ghcp.ps1` | Need broader live corpus and richer trace/claim scoring | P0 |

## Existing Harnesses

- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`
- `scripts/skill-lint.ps1`
- `scripts/route-complexity-eval.ps1`
- `scripts/skill-eval.ps1`
- `scripts/skill-eval-run-codex.ps1`
- `scripts/skill-eval-run-ghcp.ps1`
- `scripts/skill-eval-run-live-corpus.ps1`
- `scripts/skill-eval-claims.ps1`
- `scripts/skill-eval-quality.ps1`
- `scripts/skill-eval-regression-report.ps1`
- `scripts/skill-sync-report.ps1`
- `git diff --check`

## Canonical Commands

```powershell
.\.github\skills\kb-check\scripts\kb-check.ps1 -All
git diff --check
```

For touched ATV repo copies:

```powershell
git -C E:\all-the-vibes diff --check
```

## Scaffolding Decisions

No new smoke eval is scaffolded by `kb-eval-map` for this repo yet because the
next useful smoke is not a placeholder; it is the planned live skill eval suite:
prompt routing, trace capture, claim verification, output quality scoring, and
cost telemetry.

The first deterministic scorer now exists at `scripts/skill-eval.ps1`. It scores
captured agent result JSON against route fixtures and claim checks, and its
self-test includes intentionally bad route/proof/claim outputs that must fail.
The Codex adapter exists at `scripts/skill-eval-run-codex.ps1`; dry-run mode is
included in `kb-check -All`, and live mode explicitly invokes `codex exec` in a
disposable read-only worktree. The GHCP adapter exists at
`scripts/skill-eval-run-ghcp.ps1`; it uses GitHub Copilot CLI prompt-level JSON
constraints because the observed local CLI does not expose a Codex-style
`--output-schema` flag. `scripts/skill-eval-run-live-corpus.ps1` runs selected
fixtures across Codex and GHCP adapters and summarizes pass/fail/skip categories.
`scripts/skill-eval-claims.ps1` checks transcript-derived claim artifacts
deterministically and reports ambiguous claims without counting them as proof.
`scripts/skill-eval-quality.ps1` scores output-quality fixtures separately from
deterministic route/proof/claim pass/fail.
`scripts/skill-eval-regression-report.ps1` summarizes local live-run artifacts
and compares pass/non-pass plus size/time proxies against selected baselines.

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
