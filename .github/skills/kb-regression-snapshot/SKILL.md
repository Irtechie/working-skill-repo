---
name: kb-regression-snapshot
description: Capture and replay deterministic regression snapshots at coherent KB proof-batch boundaries. Use after an integrated slice group passes QA and before a later batch touches covered behavior.
argument-hint: "[capture|verify plus slice id/spec path]"
---

# KB Regression Snapshot

Freeze what passed so later proof batches cannot quietly break it without
replaying the same snapshots between tightly coupled slices.

This is not a test-design skill. The LLM defines the smallest useful snapshot spec. The runner executes it mechanically.

Use this for cross-slice application state snapshots. Use the skill-eval
baseline path for skill-harness fixture regressions. The two mechanisms solve
different problems: app/workflow replay between slices versus scorer output
comparison across eval runs.

## Runner

Use:

```powershell
.github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1 capture -SliceId <id> -SpecPath <spec.json>
.github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1 verify -SnapshotId <impacted-id>
.github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1 verify -MilestoneId <manifest-or-release-id>
```

Store snapshots at:

```text
.kb/snapshots/<slice-id>.json
```

## Snapshot Shape

```json
{
  "slice_id": "JE3",
  "captured_at": "ISO-8601",
  "checks": [
    {"type": "dom-element", "url": "/dashboard", "selector": ".margin-value", "expected_text": "42%"},
    {"type": "route-status", "url": "/api/deals/AIG", "expected_status": 200},
    {"type": "file-checksum", "path": "src/config.ts", "sha256": "abc123..."}
  ]
}
```

## Capture

After a coherent slice group passes its proof-batch `kb-check`,
`kb-functional-test`, and `kb-qa`, build a compact spec from what changed:

| Change | Snapshot checks |
|---|---|
| Frontend/UI | route URL, key DOM selector, expected text or text pattern, console error count `0` |
| API | endpoint URL, expected status, required response fields or schema shape |
| CLI | command, expected exit code, expected output substring |
| Files | path and SHA-256 checksum for generated/config/runtime files |

Use behavioral checks for UI snapshots. Prefer "the margin value is visible and numeric" over "a div has class X." A class selector is acceptable only as a stable locator, not as the behavior being proven.

Do not store secrets, cookies, tokens, credentials, or large response bodies. Store only deterministic assertions and small metadata.

## Verify

Before the next proof batch starts execution, select only snapshots whose
declared inputs or covered behavior are affected by the planned batch, then run
`verify -SnapshotId`. Do not replay snapshots between slices in the same batch.
A no-scope verify request fails closed.

Run the complete snapshot set once at a genuine manifest/release milestone with
`-MilestoneId`. The runner fingerprints the selected snapshot definitions and
returns `snapshot-verify: REUSE` without executing checks when that exact
milestone proof was already completed unchanged.

The runner must:

- batch headless DOM checks into one Playwright browser lifecycle;
- verify API/route status with `fetch`, `curl`, or platform equivalent;
- verify CLI checks with a bounded child process and check exit code/output;
- verify file checksums with SHA-256;
- exit nonzero on the first failed snapshot.

If any snapshot fails, STOP. Mark the current slice `🔒 blocked` with the failing snapshot path, check type, expected value, observed value, and log/trace path. Do not edit implementation files until the regression is resolved, parked by the human, or explicitly skipped.

When a repo is accessed through a Windows junction, subst drive, or other alias,
run verification from the canonical Git worktree path when snapshot commands use
path containment or symlink checks. Record the canonical path used. Do not
rewrite snapshots merely because an alias path fails.

For isolated worktree batches, snapshot verification runs before worktree
preparation in the coordinator checkout. Snapshot capture after integration runs
from the source checkout after the coordinator has run or reused proof. Workers may
return snapshot specs or artifacts in their receipts, but coordinator capture is
the authoritative cross-batch replay surface.

## Output

Capture:

```text
snapshot-capture: PASS JE3 -> .kb/snapshots/JE3.json
```

Verify:

```text
snapshot-verify: PASS 7/7 snapshots
```

or:

```text
snapshot-verify: FAIL .kb/snapshots/JE3.json
failed: dom-element /dashboard .margin-value
expected: 42%
observed: <missing>
```

Record the result in the manifest notes. Snapshot verification is acceptable machine proof for `kb-complete`.
