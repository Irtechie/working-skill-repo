# Skill Eval Results

This directory contains deterministic scoring fixtures for skill behavior.

The runner does not call a model. It scores captured agent results against the
route-complexity dataset:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1
```

Default mode is a self-test under `evals/skill-eval/selftest/`. The self-test
contains one valid result and several intentionally bad results. The runner must
pass the good result and fail the bad ones; otherwise the scorer is too weak.

For real captured runs, write a result JSON and pass it explicitly:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1 -ResultPath path\to\captured-result.json
```

## Result Shape

```json
{
  "id": "run-id",
  "fixture_id": "tiny-typo-fix",
  "expected_result": "pass",
  "actual": {
    "route": "kb-fix",
    "user_questions": 0,
    "artifacts": ["changed file", "verification note"],
    "proof": ["git diff --check", "targeted text/render check if UI-visible"]
  },
  "trace": {
    "files_read": ["todo.md"],
    "commands": ["git diff --check"],
    "tools": ["shell"]
  },
  "claim_checks": [
    {
      "type": "command_ran",
      "path": "",
      "contains": "git diff --check",
      "expected": true,
      "claim": "Agent claimed git diff --check was run"
    }
  ]
}
```

Supported claim checks:

- `file_exists`
- `command_ran`
- `file_read`

Live adapters should produce this result shape from transcripts/traces, then let
this deterministic scorer decide pass/fail. The Codex adapter exists; GHCP is
still a planned adapter.

## Codex Adapter

Run the Codex adapter in safe dry-run mode:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\skill-eval-run-codex.ps1 -FixtureId tiny-typo-fix -DryRun
```

Run one live Codex eval:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\skill-eval-run-codex.ps1 -FixtureId tiny-typo-fix -KeepRun
```

The live adapter creates a disposable git worktree under `.atv/eval-runs/`, runs
`codex exec` in read-only mode with a JSON output schema, writes `result.json`,
then calls `scripts/skill-eval.ps1 -ResultPath <result.json>`.

Dry-run mode is part of `kb-check -All`; live mode is explicit because it calls a
model.
