# Run state layout

Ephemeral, git-ignored control-loop state for one active `kb-goal` run, at
`.kb/runs/<goal-slug>/`. `kb-goal` owns the loop rules and the `run-state`
guard; this file is the layout only.

## Required files

- `goal.md` - pointer to the durable goal ledger and current objective.
- `done-check.json` - optional `kbcheck sense/accept` check spec when the done
  check can be expressed as JSON.
- `catalog.json` - the redacted live run catalog for this goal/run only.
- `catalog-fingerprint.txt` - the last accepted host/config fingerprint used to
  decide whether the run catalog can be reused.
- `backlog.json` - small queue of candidate work units with route, priority,
  blockers, and source artifact.
- `progress.md` - compact current state, last accepted proof, and next allowed
  action.
- `route-history.jsonl` - one JSON object per route decision.

## Route history row

```json
{"ts":"<ISO-8601>","route":"kb-work","confidence":0.82,"state_changed":true,"progress_key":"slice-003-done"}
```

## Run catalog rules

The run catalog stays redacted, ephemeral, and project-local. It records only
what this host and this run can select, plus any project-allowed aliases the
user already configured. It is never a trust source; credentials, approvals,
and trust stay under the OS user's private KB state. Refresh it only when the
surface/provider/configuration or generated agent fingerprint changes.
