---
name: kb-complete
description: "Single user-facing KB completion command. Takes a feature, plan, or manifest from its current state through planning, work, post-work finalization, and configured delivery: local completion, pushed PR, or explicitly configured direct integration and sync."
argument-hint: "[feature description, plan path, manifest path, or blank to resume active work]"
---

# KB Complete

Take KB work from its current durable state to the project's configured endpoint.
Users should not need to choose between plan, work, finalize, ship, or land.

```text
source/plan/manifest
  -> kb-plan when needed
  -> kb-work
  -> kb-finalize
  -> delivery policy
  -> local | PR | direct integration
  -> post-integration sync when configured
```

This is the user-facing orchestrator. Internal phases remain separate so they
can enforce narrow gates and resume safely.

## Safety Contract

- Explicit `kb-complete` invocation authorizes safe local planning, execution,
  review, repair, proof, learning, and cleanup.
- Publishing authority comes from configured delivery policy or explicit
  run-scoped user authorization. Absence of policy defaults to `pr`: reviewed
  work is committed, pushed, and opened as a PR. Absence of policy never
  authorizes merge, direct-default push, or post-merge sync.
- Never infer direct-default permission from repository ownership or write
  access. Permissions answer where a push can go; policy answers whether a PR is
  required.
- Never merge, direct-push default, deploy, or propagate external copies unless
  the selected policy explicitly authorizes that action.
- Do not stage, commit, revert, or overwrite unrelated dirty work.
- The reviewed manifest-owned plan-run branch is the only delivery candidate.
  Do not reconstruct delivery from a dirty source/default checkout.

## Input Resolution

1. Run `kb-map lookup <request>` and resolve the active project root.
2. Before any mutating phase, claim or resume the stable objective in the shared
   `kb-start` work queue with the current project session ID. If another live
   session owns it, stop successor creation and report that session.
3. If input is a manifest, use it.
4. If input is a slice plan, requirements doc, brainstorm, or handoff, follow
   its source/manifest pointers before creating duplicate artifacts.
5. If input is a clear feature description with no manifest, invoke
   `kb-plan <input>` with execution intent.
6. If material product or architecture questions remain, route through
   `kb-brainstorm` before planning.
7. If input is blank, resume the single active manifest from `todo.md`. Ask only
   when multiple active manifests are genuinely plausible.

## State-Driven Loop

Re-read the manifest after every delegated phase. Durable state, not chat memory,
chooses the next action.

Successful delegated phases return to this state-driven loop automatically.
`kb-work` does not end after slices or finalization, `kb-finalize` does not end
after review, and `kb-ship` does not end an authorized merge run at PR creation.
Each phase records its evidence, then returns control here for the next gated
transition.

Heartbeat the shared work claim after every delegated phase. Publish `done`,
`blocked`, or `superseded` before terminal output.

At the configured endpoint, register exactly one lifecycle delivery state:
`local-durable`, `awaiting-review`, or `delivery-integrated`. Keep delivery,
physical cleanup, ref retirement, and host session retirement as four separate
authorities and reported dimensions. Release or suspend ownership only after
lifecycle registration succeeds. A different/later controller performs
physical cleanup; the current session never deletes its own worktree.

Do not roll a narrow gate up to the whole product. A release-only failure can
prevent `pr-open` or `landed` while local implementation remains complete. An
optional provider/platform gate cannot block the configured core product.
Before repeating any blocker, rerun its recorded recheck sensor.

| Current state | Action |
|---|---|
| no valid manifest | `kb-plan <source>` |
| active with runnable slices | `kb-work <manifest>` |
| completed with `work-to-complete: passed` | `kb-finalize <manifest>` |
| reviewed with `complete-to-ship: passed|quarantined` | apply delivery policy |
| paused | stop without changing technical gate status; resume only on explicit user instruction |
| blocked/human-required on critical path | recheck current state, persist exact owner/scope/resume condition, and stop only affected work |
| release/deployment/signing blocked after review | preserve implementation-complete state and report delivery blocked |
| optional capability/platform blocked | defer or quarantine only that capability and continue the configured core endpoint; use `parked` only after explicit human deferral |
| no state change after one repair | stop with the smallest unblock action |

`kb-work` may invoke `kb-finalize` automatically. Re-read the manifest and skip
already-proven phases rather than repeating review or learning.

Treat a passing exact-tree proof receipt as durable phase evidence. `kb-complete`
must not rerun it merely because orchestration moved from work to finalize,
ship, or land. Rerun only invalidated checks when relevant files, dependencies,
test configuration, environment identity, or the delivery tree changed.

## Delivery Policy

Read `docs/context/operations/kb-routing.yaml` when present. `kb-configure`
manages the portable project policy.

```yaml
delivery:
  mode: pr           # local | pr | direct
  merge: manual      # manual | auto-after-checks
  post_merge_sync: false
```

Defaults when absent:

- `mode: pr`
- `merge: manual`
- `post_merge_sync: false`

Finished work reaches a pushed, review-ready PR without asking. Reaching
PR-ready is automatic; accepting a PR never is. A project that genuinely wants
work to stay on disk must opt out with `kb-configure delivery-local`.

### Local

Stop after `kb-finalize` passes. Report the reviewed manifest and exact delivery
command if the user later wants publishing.

### PR

Invoke `kb-ship <manifest>` to audit scope, commit intentional files, push a
topic branch, and create/update a PR.

- With write access, use a same-repository topic branch.
- Without write access, use an authorized fork and upstream PR.
- Write access never bypasses the PR policy.
- With `merge: manual`, stop at the correctly based open PR unless explicit
  run-scoped user authorization permits merge for this run.
- With `merge: auto-after-checks`, invoke `kb-land <manifest>` only after ship
  proof and required checks/approvals pass.
- When same-run authorization permits merge, consume the exact `kb-ship`
  branch/commit/PR evidence and invoke `kb-land <manifest>` after the same
  required checks and approvals pass.

### Direct

Invoke `kb-land <manifest>` with direct-delivery policy.

- Direct mode must be explicitly stored or stated for this run.
- Branch protection, required reviews, stale default, failed release checks, or
  ambiguous scope force PR fallback or block; never bypass protection.
- Do not use force push or admin bypass.

Absent policy is PR/manual: reviewed work reaches an open PR and stops. Write
access never enables direct delivery automatically and no policy authorizes
merge. Only `kb-land`, under explicit direct or authorized auto-merge policy,
owns remote-default integration.

The automatic chain selects phase owners; it never transfers their authority.
`kb-complete` invokes `kb-ship` for PR delivery and then invokes `kb-land` only
when configured delivery policy or explicit run-scoped user authorization
permits integration.

## Terminal Worktree Retirement

Before terminal completion or worktree retirement, branch on the same
`cmd/kbcheck help` capability probe used by `kb-check`. In native mode,
validate the build-storage cleanup receipt produced by `kb-finalize`:

```powershell
go run ./cmd/kbcheck cargo-storage --action validate --run-id <run-id> --root <project-root> --json
```

It must be native `done`, native machine-validated `not-applicable`, portable
`done-portable-fallback` with
`removed_bytes: 0`, or `not-applicable` only when no Cargo command ran. A
missing or `blocked` receipt blocks cleanup completion; it does not downgrade
already-proven implementation. Never compensate by deleting a stable shared
Cargo target.

In portable-fallback mode, do not invoke unavailable native validation.
Recompute the canonical Git common-directory identity and its 24-hex project
key, then require the fallback record's identity and absolute target to match,
the target basename to equal that key, status to be `done-portable-fallback`,
`removed_bytes` to be zero, and no temporary target entries to exist. Any
mismatch blocks cleanup completion.

Register cleanup only after the configured endpoint is durably proven:

- `local`: the worktree is clean, `HEAD` is committed, the exact local feature
  branch ref still points to the delivered commit, and at least one configured
  remote authoritatively identifies a different default branch;
- `pr`: the fetched remote topic contains the delivered commit; an open PR may
  lose its worktree, but its local and remote feature refs remain until
  integration is proven;
- `direct`: the fetched authoritative remote default contains the delivered
  commit. Use this integrated endpoint after either a direct push or a proven
  PR merge; `pr` always means PR-only/open and retains feature refs.

Before cleanup registration, record the corresponding lifecycle endpoint:

- `local` -> `local-durable`;
- open/manual `pr` -> `awaiting-review`;
- proven integrated `direct` -> `delivery-integrated`.

An `awaiting-review` item may remain open for weeks without consuming active
WIP. It may suspend ownership and optionally retire only a proven-clean
worktree, while retaining local and remote refs plus an exact resume packet:
canonical repository, work/claim/session identities, manifest and requirements,
plan-run branch, delivered commit, remote topic and observed SHA, PR identity
and URL, current gate and proof receipts, protected/quarantined paths, and exact
recreation/resume commands. Missing resume evidence blocks worktree retirement,
not the already-proven PR delivery.

Write that resume packet as versioned JSON outside the worktree being retired;
never infer or synthesize missing fields. Bind its immutable packet identity and
SHA-256 digest into the cleanup receipt. It must explicitly contain the
canonical delivery repository; work, claim, and session identities; branch;
delivered SHA; remote/ref/observed SHA; provider, PR identity, and PR URL;
manifest and requirements pointers; current gate and proof pointers;
protected and quarantined path arrays (including explicit empty arrays); and
exact worktree recreation and `kb-start` resume commands. Registration validates
all identities against live delivery evidence. Sweep rereads the packet and
blocks retirement if its path, identity, digest, or required contents drift.

When `cmd/kbcheck` provides the guard, register the exact terminal target before
releasing ownership:

```powershell
go run ./cmd/kbcheck terminal-cleanup --action register `
  --work-id <work-id> --session-id <project-session-id> `
  --worktree <exact-worktree> --branch <exact-feature-branch> `
  --commit-sha <delivered-commit> --delivery-mode <local|pr|direct> `
  [--remote <delivery-remote>] `
  [--claim-id <claim-id> --provider <provider> --pr-id <pr-id> `
   --pr-url <pr-url> --resume-packet <path-for-pr>] `
  --root <project-root>
```

Then release or suspend the shared work claim with the matching lifecycle state
(`local-durable`, `awaiting-review`, or `delivery-integrated`), preserving the
registered endpoint rather than collapsing it to `done`. Do not register cleanup for
`blocked`, `delivery-blocked`, pending/unverified delivery, or unrelated dirty
work. Ask a coordinator or later `kb-start` session to run
`terminal-cleanup --action sweep --session-id <current-project-session-id>`;
the current executing session cannot delete itself even when its process runs
outside the recorded worktree. The guard removes no worktree with force and
preserves ignored files as dirt. Registration records the linked-worktree admin
directory, random generation marker, real path, authoritative remote-default
names and SHAs, and any required topic SHA. Sweep requires remote refs to advance
monotonically from that evidence, validates the bidirectional Git admin
round-trip, refreshes remote authority immediately before each destructive
action, and removes refs only with exact-SHA compare-and-swap. A missing,
rewritten, or renamed default blocks and requires re-registration. Local-only completion
retains its durable branch ref. `pr` completion retains feature refs even if the
commit later appears on default. Only a receipt registered as the proven
integrated `direct` endpoint deletes the exact matching merged local feature
ref; squash/rebase integration remains blocked until provider-backed merge
proof exists. Remote feature-ref deletion remains provider/host-owned because
plain Git deletion has no race-safe compare-and-swap.

Never release the shared work claim before lifecycle registration succeeds.

If the native guard is unavailable, do not improvise filesystem deletion.
Report the exact session ID, worktree, branch, commit, delivery proof, and
`cleanup: deferred-host` for an external owner.

## Terminal Outcomes

```text
KB complete: local|pr-open|landed|nothing-to-deliver|delivery-blocked|blocked
Manifest: <path>
Finalization: <complete-to-ship status>
Implementation: complete|incomplete
Delivery policy: <local|pr|direct>
Branch: <branch or none>
Commit: <sha or none>
PR: <url or none>
Integration: <not-requested|pending-review|merged|direct>
Sync: <not-configured|done|blocked>
Build storage: <done retained_bytes=N removed_bytes=N|not-applicable|blocked>
Cleanup: <registered|deferred-current-session|deferred-host|blocked>
Next: none|<exact resume action>
```

Do not report `landed` unless the remote default branch contains the delivered
commit and any configured post-integration sync has been verified.

## Terminal Integration Question

After reporting `pr-open`, ask exactly one question and then stop:

```text
PR is ready for review: <url>
Sync now, or wait for PR review?
```

Rules:

- Ask only when the PR is genuinely review-ready: pushed, correctly based,
  proof green, and scope audited. Never ask to paper over a failed gate.
- Ask once. A `local`, `nothing-to-deliver`, `delivery-blocked`, or `blocked`
  outcome has nothing to decide, so do not ask.
- "Wait for PR review" is the default on silence. Leaving a PR open is a
  finished state, not an unfinished one.
- "Sync now" authorizes integration for this run only. It never writes policy,
  never enables `auto-after-checks`, and never implies future auto-merge.
- Honor the answer through `kb-land`, which still verifies that the remote
  default branch contains the delivered commit.

This is the only terminal question. Do not also ask whether to commit, push,
open a PR, clean up, or retire a worktree: those are agent-owned.

## Compatibility

- Older prompts that say "kb finish" or "full KB pipeline" route here directly.
- `kb-finalize`, `kb-ship`, and `kb-land` remain internal phase skills.
