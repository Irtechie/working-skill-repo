---
name: kb-rehab
description: Review and reconcile explicitly accepted unfinished repository work. Survey portable Git lineage, preserve selected artifacts, classify every accepted item, deliver useful work through existing KB owners, and retire explicitly rejected local work only after restoration proof. Ordinary startup offers scoped cleanup without stopping independent work.
argument-hint: "[accepted scope, or blank for explicit repository-wide review]"
---

# KB Rehab

Resolve unfinished work to evidenced dispositions while independent work keeps
moving. Scope comes from the current invocation or a bound user reply, never
from an unanswered offer or the age of a branch.

## Authority and outcome

An explicit invocation authorizes review, preservation, and policy-owned
reconciliation within this repository. A scoped acceptance names exact refs,
tips and artifact hashes; it does not authorize unrelated work. Explicit
rejection of unique work is separate from merely accepting a cleanup offer.
Current pause/revocation stops pending mutation. Recheck current policy and
acceptance on resume; an old receipt is not irrevocable permission.

Complete each accepted item to a real result or a named scoped limitation:
`preserve-live`, `review-needed`, `salvage`, `discard-confirmed-junk`,
`deliver-pr`, `merge-eligible`, or `retained-blocked`. A delivery next action is pending until its
owner returns fresh observed evidence. The portable helper never claims a
forge action completed or treats an imported receipt as observed forge proof.
`merge-eligible` means an authorized pending owner attempt with required fresh
forge checks; it does not assert that those checks already pass.

A dirty primary checkout or retained branch is not by itself failure. Preserve
source/current worktrees, active claims, credentials and uncertain work. Never
use age, ignore status, an empty file, a missing manifest or failing tests as
proof that work is disposable. No broad clean/reset/stash, force push, shared
history rewrite, current-worktree removal, or automatic global propagation.

## Installed sequence

Resolve the project with `kb-map`. Resolve `<rehab-dir>` from the loaded skill's
actual location, then use its shipped PowerShell 5.1/Git helper:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "<rehab-dir>/scripts/recovery.ps1" -Action survey -Root "<repo>" -Json
```

1. Read the advertised/fetched default, local refs/tips, dirty fingerprints,
   candidates and live protections. Unknown authority withholds destructive
   actions; it does not stop independent local work.
2. During ordinary startup, use `kb-start`'s `continue` contract: one advisory,
   silence/decline preserves backlog, independent preparation returns `kb-work`.
3. For accepted cleanup, construct the exact request in
   [recovery requests](references/recovery-requests.md), then call `-Action
   dispose -Request <request.json>`. Consume every item result; a proposed
   delegate is not a completed disposition.
4. Review unknown pairings. For useful work, establish a contained owning
   manifest with exact ref/tip and hash binding before requesting delivery.
   `kb-complete` must review/prove the old work in its **own clean owned delivery
   workspace/branch**, copying the selected manifest when needed. Never append
   the old commits to the independent new feature branch. `kb-ship` owns
   exact-head PR lookup/reuse; `kb-land` owns actual current policy/check/review
   observation and eligible integration. Re-survey after each delegate returns.
5. Explicitly rejected, unoccupied non-default local refs can be archived and
   restoration-tested before compare-and-swap retirement. Exact generated
   output can be removed only with explicit rejection, matching hash-bound
   provenance, no live claim, and a verified recovery copy. Source documents
   and credentials remain in place. Report the exact recovery location.
6. If an item cannot safely advance, retain it with its precise reason and
   next owner. Continue other accepted items and the independent feature.
   Report actual archive/ref/PR outcomes, never a blanket “tree clean” claim.

Read [classification](references/classification.md) for evidence boundaries and
[grant](references/grant.md) for native versus portable delivery limits.

## Native tooling and delivery

`cmd/kbcheck` and `kbreconcile` are optional maintainer/native capabilities,
not prerequisites in installed consumers. When present, use their actual
owners and predicates; a native refusal cannot be retried through a weaker
portable destructive path. The helper therefore withholds portable retirement
when the native owner is present. Missing binaries do not waive proof.

Explicit `w2d` grants delivery intent only for its named objective. A private or
solo repository alone grants no merge. Existing configured auto-after-checks
or an explicit scoped merge instruction still requires actual permissions,
exact PR head/base, checks, reviews and protection. Local-only policy keeps
publishing withheld. No dormant native evaluator or self-reported forge result
can replace those observations.
