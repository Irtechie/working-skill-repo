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
| Skill edits do not regress behavior | prompt/trace/claim evals | Not built | Need live skill eval suite outside skills | P0 |

## Existing Harnesses

- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`
- `scripts/skill-lint.ps1`
- `scripts/route-complexity-eval.ps1`
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

## Deterministic vs LLM-Judged

| Check | Class |
|---|---|
| skill lint | deterministic |
| route-complexity fixture scoring | deterministic |
| sync drift hashes | deterministic |
| git whitespace/conflict checks | deterministic |
| output quality scoring | LLM-judged, future |
| skill prompt routing live run | mixed: model action plus deterministic route/trace scoring |
| final claim verification | deterministic filesystem/git/log checks |

## Credentials / Session Requirements

None for current deterministic checks.

Future live Codex/GHCP evals may require runtime-specific CLI access, trace
capture, or authenticated platform sessions.

## Dashboard / Export Options

Keep local scripts as source of truth. Langfuse, Braintrust, LangSmith,
Promptfoo, or DeepEval can be adapters/exporters after local fixtures and scores
are stable.

## Open Eval Gaps

- Build live prompt-routing runner for Codex/GHCP.
- Add trace scoring for files read, commands run, and forbidden shortcuts.
- Add claim verifier that checks final answers against git/files/logs/artifacts.
- Add output-quality rubric for maintainability, completeness, relevance, and
  proof quality.
- Track cost: tokens/time/tool calls/retries per verified successful outcome.
