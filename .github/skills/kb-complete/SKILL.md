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
- Publishing authority comes from project delivery policy or an explicit
  run-scoped user instruction. Absence of policy defaults to `local`.
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

Heartbeat the shared work claim after every delegated phase. Publish `done`,
`blocked`, or `superseded` before terminal output.

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
  mode: local        # local | pr | direct
  merge: manual      # manual | auto-after-checks
  post_merge_sync: false
```

Defaults when absent:

- `mode: local`
- `merge: manual`
- `post_merge_sync: false`

### Local

Stop after `kb-finalize` passes. Report the reviewed manifest and exact delivery
command if the user later wants publishing.

### PR

Invoke `kb-ship <manifest>` to audit scope, commit intentional files, push a
topic branch, and create/update a PR.

- With write access, use a same-repository topic branch.
- Without write access, use an authorized fork and upstream PR.
- Write access never bypasses the PR policy.
- With `merge: manual`, stop at the correctly based open PR.
- With `merge: auto-after-checks`, invoke `kb-land <manifest>` only after ship
  proof and required checks/approvals pass.

### Direct

Invoke `kb-land <manifest>` with direct-delivery policy.

- Direct mode must be explicitly stored or stated for this run.
- Branch protection, required reviews, stale default, failed release checks, or
  ambiguous scope force PR fallback or block; never bypass protection.
- Do not use force push or admin bypass.

Absent policy is always local-only. PR/manual is the recommended team policy,
but write access never enables it automatically and it never authorizes merge.
Only `kb-land`, under explicit direct or authorized auto-merge policy, owns
remote-default integration.

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

When `cmd/kbcheck` provides the guard, register the exact terminal target before
releasing ownership:

```powershell
go run ./cmd/kbcheck terminal-cleanup --action register `
  --work-id <work-id> --session-id <project-session-id> `
  --worktree <exact-worktree> --branch <exact-feature-branch> `
  --commit-sha <delivered-commit> --delivery-mode <local|pr|direct> `
  [--remote <delivery-remote>] --root <project-root>
```

Then release the shared work claim as `done`. Do not register cleanup for
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

## Compatibility

- `kb-finish` is a deprecated alias that delegates here.
- `klfg` may delegate here for its full idea-to-endpoint loop.
- `kb-finalize`, `kb-ship`, and `kb-land` remain internal phase skills.
