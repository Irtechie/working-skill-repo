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
| Route complexity stays calibrated | `evals/route-complexity/*.json` | `scripts/route-complexity-eval.ps1` | Fixtures are deterministic metadata, not live prompt runs | P0 |
| Required skill copies stay synced | global installs and ATV `.github` skills | `scripts/skill-sync-report.ps1` | ATV scaffold/plugin shipping policy unresolved | P1 |
| Skill edits do not regress behavior | prompt/trace/claim evals | `scripts/skill-eval.ps1` self-test scorer | Need live Codex/GHCP adapters and broader corpus | P0 |

## Existing Harnesses

- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`
- `scripts/skill-lint.ps1`
- `scripts/route-complexity-eval.ps1`
- `scripts/skill-eval.ps1`
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
It does not yet run Codex or GHCP itself.

## Deterministic vs LLM-Judged

| Check | Class |
|---|---|
| skill lint | deterministic |
| route-complexity fixture scoring | deterministic |
| captured skill result scoring | deterministic |
| sync drift hashes | deterministic |
| git whitespace/conflict checks | deterministic |
| output quality scoring | LLM-judged, future |
| skill prompt routing live run | mixed: model action plus deterministic route/trace scoring |
| final claim verification | deterministic filesystem/git/log checks |

## Credentials / Session Requirements

None for current deterministic checks.

Future live Codex/GHCP adapters may require runtime-specific CLI access, trace
capture, or authenticated platform sessions. The deterministic scorer can run
without those sessions when given captured result JSON.

## Dashboard / Export Options

Keep local scripts as source of truth. Langfuse, Braintrust, LangSmith,
Promptfoo, or DeepEval can be adapters/exporters after local fixtures and scores
are stable.

## Open Eval Gaps

- Build live prompt-routing adapters for Codex/GHCP that produce
  `evals/skill-eval` result JSON.
- Expand trace scoring for forbidden shortcuts and required workflow reads.
- Expand claim verification from structured claim checks to transcript-derived
  claims against git/files/logs/artifacts.
- Add output-quality rubric for maintainability, completeness, relevance, and
  proof quality.
- Track cost: tokens/time/tool calls/retries per verified successful outcome.
- Add scaffold negative-check validation to future consuming-repo eval maps: any
  generated smoke eval must fail when its expected selector/status/output/schema
  is intentionally broken.
