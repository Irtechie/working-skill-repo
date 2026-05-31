# Skill Eval Baselines

Baseline reports are selected JSON outputs from
`scripts/skill-eval-regression-report.ps1`. Keep only useful comparison points,
such as the last accepted live corpus before a major skill-routing change.

Live run directories stay under `.atv/eval-runs/` and are ignored. A baseline is
checked in only when it is intentionally useful for regression comparison.

`scripts/skill-eval.ps1` can also persist and enforce deterministic scoring
baselines:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1 -BaselinePath evals/skill-eval/baselines/selftest.json -UpdateBaseline
powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1 -BaselinePath evals/skill-eval/baselines/selftest.json
```

Baseline comparison fails when a baseline row disappears, a passing result
starts failing, or issue counts increase. Intentional baseline changes must use
`-UpdateBaseline`.
