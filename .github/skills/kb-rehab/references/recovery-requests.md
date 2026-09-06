# Portable recovery requests

Run the script from the **loaded installed `kb-rehab` directory**, not from a
consumer's assumed `.github` path. Windows PowerShell 5.1 and Git suffice.
These requests record current user intent; a prior plan's passed gate is not
execution or cleanup authorization. Values marked `<...>` must come from the
actual source, current run, or user decision.

## Survey and prepare

```powershell
$helper = '<loaded-kb-rehab-dir>/scripts/recovery.ps1'
$source = '<absolute-source-repository>'
$survey = (& $helper -Action survey -Root $source -Json | Out-String) | ConvertFrom-Json
$selected = @($survey.dirty_paths | Where-Object { $_.path -eq 'docs/plans/current-manifest.md' } |
  ForEach-Object { @{ path = $_.path; sha256 = $_.sha256 } })
$prepare = @{
  schema_version = 1
  run_id = 'current-feature'
  objective = 'Implement the selected feature'
  survey = $survey
  branch = 'codex/current-feature'
  destination = '<source-parent>/.kb-recovery-worktrees/current-feature'
  dependency_status = 'independent'
  authority = @{ source = 'current-run'; prepare = $true; objective = 'Implement the selected feature' }
  artifacts = $selected
}
$prepare | ConvertTo-Json -Depth 12 | Set-Content '<absolute-prepare.json>' -Encoding UTF8
& $helper -Action prepare -Root $source -Request '<absolute-prepare.json>' -Json
```

Select only actual current-task plan/document paths and preserve their returned
hashes. `independent` requires an assessed dependency relationship; use another
value to retain a scoped prerequisite. Destination is one direct child of the
source parent's `.kb-recovery-worktrees` directory. Existing compatible owned
receipts resume; incompatible or live destinations remain preserved. Selected
paths already present in the baseline are refused rather than overwritten.

Before first implementation, use the same request with `-Action verify`.
Require `prepared`. Source/index/default drift or modified copied artifacts
requires refresh, not an invented successful receipt. Implementation then uses
`kb-work`'s ordinary exact-head proof; preparation is not code completion.

## Continue without awaiting cleanup

```powershell
$continue = @{
  schema_version = 1
  run_id = 'current-feature'
  objective = 'Implement the selected feature'
  authority = @{ source = 'current-run'; continue = $true; objective = 'Implement the selected feature' }
  preparation_request = '<absolute-prepare.json>'
  manifest = 'docs/plans/current-manifest.md'
}
$continue | ConvertTo-Json -Depth 8 | Set-Content '<absolute-continue.json>' -Encoding UTF8
& $helper -Action continue -Root $source -Request '<absolute-continue.json>' -Json
```

Display the enumerated offer only when `notice.emit` is true; continue with the
returned workspace/manifest and existing gate revalidation. Silence never
permits cleanup. To record an actual reply, add:

```json
{
  "reply": {
    "source": "current-user-reply",
    "offer_id": "<notice.offer_id>",
    "response": "review",
    "item_ids": ["<shown-item-id>"]
  }
}
```

Known responses are `unanswered`, `decline`, `review`, `cleanup`, and ambiguous
`accept` (review only). `revoked`/`paused` withdraw cleanup; unknown responses
create no accepted scope. A boolean `authority.paused: true` stops the whole
continuation. Changed tips/hashes are excluded from late replies. An accepted
cleanup field is a pending next action, never a claim cleanup completed.

## Accepted disposition

Use exact current refs/tips and artifact hashes from survey. An example of an
explicitly rejected branch with selected notes to preserve is:

```json
{
  "schema_version": 1,
  "run_id": "accepted-backlog",
  "objective": "Resolve the explicitly selected backlog",
  "repository_id": "<survey.repository_id>",
  "acceptance": {
    "source": "current-user-reply",
    "accepted": true,
    "items": [
      {"kind": "branch", "ref": "refs/heads/codex/old-work", "tip": "<exact-tip>"},
      {"kind": "artifact", "path": "docs/plans/old-notes.md", "sha256": "<exact-hash>"}
    ]
  },
  "items": [{
    "kind": "branch",
    "ref": "refs/heads/codex/old-work",
    "tip": "<exact-tip>",
    "decision": "reject",
    "rejected": true,
    "artifacts": [{"path": "docs/plans/old-notes.md", "sha256": "<exact-hash>"}]
  }]
}
```

Call `-Action dispose -Root <source> -Request <disposition.json> -Json`.
`explicit-kb-rehab` is the other accepted authority source for an actual direct
invocation. `accepted`, `rejected`, `revoked`, `paused`, and optional `merge`
are JSON booleans. A changed or revoked acceptance prevents pending retirement.
The archive and receipt paths are returned; originals remain in place.

For review, use `decision: review`. For delivery, use `decision: deliver`,
`manifest: docs/plans/<owner>.md`, and `manifest_sha256`. The actual contained
manifest must declare `ref:` or `branch:` and `tip:` or `head:` matching the
item. This validates identity binding, not code proof. `kb-complete` reviews
and proves it in its own clean delivery workspace, then delegates PR/merge.

`decision: merge` with explicit boolean `acceptance.merge: true` or configured
`merge: auto-after-checks` produces only a pending `merge-eligible` owner
attempt with `forge_checks_required: true` and `completed: false`. It does not
assert that the PR currently passes checks. `kb-land` must obtain actual forge
evidence; upstream proof/PR creation remain with `kb-complete`/`kb-ship`.
Persistent local-only or unknown delivery policy withholds publishing until
resolved by current policy ownership. Imported forge success fields do nothing.

## Exact generated output

A generated artifact additionally needs `decision: discard-generated`,
`rejected: true`, its exact `sha256`, and `provenance: {"path": "<relative-record>",
"sha256": "<record-hash>"}`. Only `.kb/generated/` or `.kb/tmp/` leaf files are
eligible. The current acceptance must enumerate that exact artifact/hash.
The pre-existing producer record must match:

```json
{
  "kind": "generated-output",
  "repository_id": "<survey.repository_id>",
  "producer": "<actual-producing-tool>",
  "path": ".kb/generated/<exact-file>",
  "sha256": "<exact-output-hash>"
}
```

The producer label alone never establishes disposable work: explicit rejection,
record bytes/hash, output hash, contained path, no live claim and a restored
recovery copy are required together. Credentials, uncertain provenance, changed
bytes and ordinary source/documents stay preserved. No directory or worktree
removal is performed.
