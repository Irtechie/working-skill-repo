---
kb_id: kb-2026-08-26-kb-rehab-outstanding-work
slice_id: slice-001
title: "Pair declared work against git reality, read-only"
blockers: []
verification: tdd
test_level: unit-plus-functional-cli
functional_risk: broad
execution_class: cli
model_tier: medium
model_tier_reason: "Bounded additive work over an existing exported inventory API and existing remote-authority helpers. Acceptance criteria, classification states, and fail-closed rules are fully enumerated in the requirements; no product or architecture choice remains open."
model_requirements: ["Go implementation against an existing internal package", "git plumbing reasoning including ancestry and patch equivalence", "deterministic fixture repository construction", "fail-closed classification design"]
escalation_triggers:
  - "A classification would be produced without the evidence its lifecycle row names."
  - "Containment or remote-authority logic would be reimplemented instead of calling internal/reconcile helpers."
  - "A missing adapter would add a conclusion rather than remove one."
  - "The preservation set would fall short of terminalCleanupSafetyPredicates()."
token_budget: 24000
cost_tier: 2
cost_tier_evidence: >-
  Prior art in this repo. reconcile.Inventory(InventoryOptions) (Ledger, error)
  in internal/reconcile/inventory.go already returns Repository{DefaultBranch,
  DefaultBranchState, Branches, Remotes, QueueClaims, Worktrees, Artifacts} with
  ProtectionReasons, Dirt, UniqueWork, and Ambiguity populated. Remote authority
  and monotonic-default checks exist in internal/reconcile/git.go
  (refreshRemotes, remoteDefaultsMonotonic, freshDeliveryState) and
  cmd/kbcheck/terminal_cleanup.go (fetchAuthoritativeRemoteDefault,
  validateTerminalDeliveryMonotonic). Ruled out tier 6, a fresh git walker and
  a second containment implementation, because R4 and R19 forbid a divergent
  second implementation and the exported Ledger already carries every artifact
  field the pairing needs. The only genuinely new code is declared-work parsing
  and the pairing/classification layer.
workspace_mode: shared-serial
conflict_domains: ["go:kbcheck-commands", "cli:kbcheck", "config:rehab-policy"]
shared_resources: ["git:common-directory", "filesystem:repo-root"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'WorkReality' -count=1"
  expect: 0
hitl: false
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: false
protected_oracles:
  - path: cmd/kbcheck/work_reality_test.go
    sha256: "d9592f7cfbab222800ba58b09707af1d8ddd0d241c8aea5ded495d6608c1e439"
    update_policy: "requires an explicit slice-plan amendment"
    purpose: "A mixed fixture repository proves every lifecycle classification, every fail-closed downgrade, and that no adapter absence adds a conclusion."
expected_files:
  - path: cmd/kbcheck/work_reality.go
    op: create
    scope: "Declared-work parsing, pairing, evidence-bound classification, and stable JSON report."
  - path: cmd/kbcheck/work_reality_test.go
    op: create
    scope: "Fixture repositories proving classification, fail-closed downgrades, cutoff protection, and preservation."
  - path: cmd/kbcheck/main.go
    op: modify
    scope: "Register the work-reality subcommand in the existing dispatch switch."
  - path: config/rehab-policy.json
    op: create
    scope: "Versioned classification predicate manifest and protected-path list."
---

# Slice 001 - Read-Only Work/Reality Pairing

## Observable Outcome

`go run ./cmd/kbcheck work-reality --root . --json` emits one classification per
declared-work/ref pairing, with the evidence that produced it, and mutates
nothing.

## Ordered Steps

1. **Register the subcommand.** Add `case "work-reality": return
   runWorkRealityCommand(root, opts, stdout, stderr)` to the dispatch switch in
   `cmd/kbcheck/main.go`, following the existing `terminal-cleanup` and
   `session-preserve` entries.
   *Pass criterion:* `go run ./cmd/kbcheck work-reality --root . --json` exits
   without an unknown-command error.

2. **Capture the cutoff and acquire evidence.** Record an immutable RFC3339
   cutoff at run start (R1a). Call `reconcile.Inventory` for repository,
   worktree, branch, remote, queue-claim, and artifact evidence. Do not
   reimplement any of it.
   *Pass criterion:* a fixture asserts the report's cutoff is stable across the
   run and that artifacts updated after the cutoff carry a `post-cutoff`
   protection reason.

3. **Parse declared work.** Read `todo.md` Active Work, Queued Improvements,
   Blocked, and Parked sections; `docs/plans/*-manifest.md` frontmatter
   `status`; `docs/context/goals/*`; and `docs/handoffs/active/*` (R2). A row
   that does not parse becomes a declared item in state `orphan-work`, never a
   dropped row.
   *Pass criterion:* a fixture containing malformed and unconventional rows
   yields one declared item per row, with unparsed rows marked `orphan-work`.

4. **Prove remote authority.** Resolve the default branch via
   `ls-remote --symref`, fetch, and require the fetched SHA to equal the
   advertised SHA; reject a rewritten default; resolve every configured remote
   and take the strictest outcome; fail closed when no remote, an unreachable
   remote, or an unresolvable default exists (R4). Call the existing
   `internal/reconcile` helpers.
   *Pass criterion:* fixtures for no-remote, unreachable-remote, stale-tracking
   ref, and force-pushed default each produce zero `dead` classifications and a
   fail-closed report status.

5. **Classify each pairing.** Implement the Lifecycle Model table exactly.
   Containment is ancestry **or** patch equivalence via
   `git cherry <remote>/<default> <branch>`; zero `+` lines means contained, so
   the pairing is `dead`, never `unshipped` (R4a). A missing manifest, session,
   handoff, or goal is missing evidence and downgrades to `orphan-*`, never
   `dead` or `superseded` (R5a). A supersession nomination holds only with
   containment proof, otherwise `human-required` (R3). Any non-terminal
   work-queue claim makes a pairing `live` regardless of heartbeat age; a stale
   claim is `human-required`, never takeover authority (R18a).
   *Pass criterion:* the mixed fixture asserts one expected state per pairing,
   including a squash-merged branch classified `dead` and a stale-claim pairing
   classified `human-required`.

6. **Enforce preservation.** The report never classifies as `dead` or
   `superseded`: the current session resolved by both host session ID and
   canonical real path with symlinks evaluated; its branch and worktree; the
   primary and default checkouts; locked, moved, or recreated worktrees; any
   tracked, untracked, or ignored dirt; and the branch or worktree of any peer
   session holding a non-terminal claim (R18). The set must equal or exceed
   `terminalCleanupSafetyPredicates()`.
   *Pass criterion:* a test asserts the slice's predicate name set is a superset
   of `terminalCleanupSafetyPredicates()`, and a symlinked-worktree fixture and a
   `E:` versus `e:` path fixture both preserve.

7. **Emit a redacted, stable report.** Stable-ordered JSON with minimum fields,
   protected-path and credential-like detail redacted, and least-privilege file
   permissions when written (R1b). Publish the predicate manifest version from
   `config/rehab-policy.json`.
   *Pass criterion:* two consecutive runs on an unchanged fixture produce
   byte-identical JSON, and a fixture containing a credential-like path asserts
   redaction.

## Acceptance Criteria

- The subcommand is read-only. A test asserts the fixture repository's git
  state, worktrees, refs, and working tree are byte-identical before and after.
- Every pairing carries its classification, the predicate names consulted, and
  the evidence source for each.
- Absence of `gh`, of network, or of `kbreconcile` removes conclusions and never
  adds one; affected pairings degrade to `human-required` or `orphan-*`.
- No pairing reaches `dead` or `superseded` without containment proof against a
  freshly fetched authoritative default.
- Running against this repository pairs the declared workstreams in `todo.md`
  against the `codex/*`, `deaderestpool-*`, `preserved/*`, and `parked/*` refs
  with no hand-pairing.

## Test Scenarios

| Scenario | Expected |
|---|---|
| Squash-merged branch, patch-equivalent to default | `dead`, not `unshipped` |
| Branch with uncontained commits, work item present | `unshipped` |
| Branch with uncontained commits, no work item | `orphan-branch`, preserved |
| Work item naming a manifest that does not exist | `orphan-work`, never `dead` |
| `Superseded by <path>` where the path exists but commits are uncontained | `human-required` |
| `Superseded by <path>` introduced by the same commit that created the path | `human-required`, never self-proving |
| Peer session holds an `in_progress` claim, heartbeat 6 hours old | `live` |
| Peer session claim stale past the review threshold | `human-required` |
| No configured remote | fail closed, zero `dead` |
| Remote default force-pushed between resolve and fetch | fail closed |
| Artifact modified after the cutoff | protected, reported for a later run |
| Current session's own branch and worktree | preserved unconditionally |

## Scope Boundary

Read-only. This slice writes no `todo.md` marker, opens no PR, deletes nothing,
and executes no proof command from any branch. It produces one report.

## Requirements Owned

R1, R1a, R1b, R2, R3, R4, R4a, R5, R5a, R6, R17, R18, R18a, R19.
