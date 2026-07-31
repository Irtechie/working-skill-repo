# Reporting and Commit

## Contents

- Output Format
- Autofix reporting
- Commit rules

## Output Format

**The full report MUST be printed as markdown output.** Do not summarize findings internally and then output a one-liner. The report is the deliverable — print every section in full, formatted as readable markdown with headers, tables, and bullet points.

After processing the selected scope, output the following report:

```text
Compound Refresh Summary
========================
Scanned: N learnings

Kept: X
Updated: Y
Consolidated: C
Replaced: Z
Deleted: W
Skipped: V
Marked stale: S
```

Then for EVERY file processed, list:
- The file path
- The classification (Keep/Update/Consolidate/Replace/Delete/Stale)
- What evidence was found -- tag any memory-sourced findings with "(auto memory)" to distinguish them from codebase-sourced evidence
- What action was taken (or recommended)
- For Consolidate: which doc was canonical, what unique content was merged, what was deleted

For **Keep** outcomes, list them under a reviewed-without-edits section so the result is visible without creating git churn.

### Autofix mode report

In autofix mode, the report is the sole deliverable — there is no user present to ask follow-up questions, so the report must be self-contained and complete. **Print the full report. Do not abbreviate, summarize, or skip sections.**

Split actions into two sections:

**Applied** (writes that succeeded):
- For each **Updated** file: the file path, what references were fixed, and why
- For each **Consolidated** cluster: the canonical doc, what unique content was merged from each subsumed doc, and the subsumed docs that were deleted
- For each **Replaced** file: what the old learning recommended vs what the current code does, and the path to the new successor
- For each **Deleted** file: the file path and why it was removed (problem domain gone, fully redundant, etc.)
- For each **Marked stale** file: the file path, what evidence was found, and why it was ambiguous

**Recommended** (actions that could not be written — e.g., permission denied):
- Same detail as above, but framed as recommendations for a human to apply
- Include enough context that the user can apply the change manually or re-run the skill interactively

If all writes succeed, the Recommended section is empty. If no writes succeed (e.g., read-only invocation), all actions appear under Recommended — the report becomes a maintenance plan.

**Legacy cleanup** (if `docs/solutions/_archived/` exists):
- List archived files found and recommend disposition: restore (if still relevant), delete (if truly obsolete), or consolidate (if overlapping with active docs)

## Phase 5: Commit Changes

After all actions are executed and the report is generated, handle committing the changes. Skip this phase if no files were modified (all Keep, or all writes failed).

### Detect git context

Before offering options, check:
1. Which branch is currently checked out (main/master vs feature branch)
2. Whether the working tree has other uncommitted changes beyond what compound-refresh modified
3. Recent commit messages to match the repo's commit style

### Autofix mode

Use sensible defaults — no user to ask:

| Context | Default action |
|---------|---------------|
| On main/master | Create a branch named for what was refreshed (e.g., `docs/refresh-auth-and-ci-learnings`), commit, attempt to open a PR. If PR creation fails, report the branch name. |
| On a feature branch | Commit as a separate commit on the current branch |
| Git operations fail | Include the recommended git commands in the report and continue |

Stage only the files that compound-refresh modified — not other dirty files in the working tree.

### Interactive mode

First, run `git branch --show-current` to determine the current branch. Then present the correct options based on the result. Stage only compound-refresh files regardless of which option the user picks.

**If the current branch is main, master, or the repo's default branch:**

1. Create a branch, commit, and open a PR (recommended) — the branch name should be specific to what was refreshed, not generic (e.g., `docs/refresh-auth-learnings` not `docs/compound-refresh`)
2. Commit directly to `{current branch name}`
3. Don't commit — I'll handle it

**If the current branch is a feature branch, clean working tree:**

1. Commit to `{current branch name}` as a separate commit (recommended)
2. Create a separate branch and commit
3. Don't commit

**If the current branch is a feature branch, dirty working tree (other uncommitted changes):**

1. Commit only the compound-refresh changes to `{current branch name}` (selective staging — other dirty files stay untouched)
2. Don't commit

### Commit message

Write a descriptive commit message that:
- Summarizes what was refreshed (e.g., "update 3 stale learnings, consolidate 2 overlapping docs, delete 1 obsolete doc")
- Follows the repo's existing commit conventions (check recent git log for style)
- Is succinct — the details are in the changed files themselves
