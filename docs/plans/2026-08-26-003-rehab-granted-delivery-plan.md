---
kb_id: kb-2026-08-26-kb-rehab-outstanding-work
slice_id: slice-003
title: "Granted delivery and delegated reaping"
blockers: [slice-002]
verification: tdd
test_level: contract-plus-functional-cli
functional_risk: severe
execution_class: cli
model_tier: large
model_tier_reason: "This slice defines a privileged trust boundary: an explicit user grant authorizes merges to a default branch that propagates into three global agent install roots. Proof provenance, adapter-absence polarity, predicate inheritance, grant binding, and TOCTOU fencing must hold together, and a single weak predicate is an irreversible supply-chain write."
model_requirements: ["security-boundary reasoning about self-attesting proof and adapter absence", "compare-and-swap and TOCTOU fencing design", "faithful inheritance of an existing versioned predicate manifest", "adversarial fixture construction for negative security cases"]
escalation_triggers:
  - "A predicate would be satisfied by the absence of an adapter."
  - "A proof command would be resolved from the candidate branch's tree."
  - "A grant would authorize a pairing it does not enumerate by ref and tip SHA."
  - "A grant would be persisted, inherited, or read by a later run."
  - "Merge would proceed on a pairing touching a protected path."
  - "Post-merge sync, propagation, or release would be attempted under a grant."
token_budget: 32000
cost_tier: 2
cost_tier_evidence: >-
  Prior art in this repo. The mandatory merge predicate set already exists as
  ActionMerge in internal/reconcile/policy.go with explicit-merge-authority,
  exact-pr-head-base, required-checks-green, required-reviews-green,
  fresh-mergeability, remote-default-authority, not-post-cutoff, and
  exact-final-tree, shipped with Allowed false and zero budgets. Risk budgets,
  quarantine, plan TTL, and receipt validation exist in policy.go, plan.go, and
  receipt.go. Delivery exists in kb-complete, kb-ship, and kb-land. Reaping
  exists as kbreconcile plan/apply/verify. The shared repository lock exists via
  acquireRepositoryLock in internal/reconcile/git.go. Ruled out tier 6, a new
  merge engine with its own predicates, because the review established that
  relocating merge to a path with fewer predicates is precisely how the grant
  fails; this slice must consume the existing manifest rather than restate it.
workspace_mode: shared-serial
conflict_domains: ["go:kbcheck-commands", "go:reconcile-policy", "cli:kbcheck", "git:refs"]
shared_resources: ["git:common-directory", "git:refs", "forge:pull-requests", "filesystem:repo-root"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck ./internal/reconcile -run 'RehabDelivery|RehabGrant|WorkReality|TerminalCleanup|ReconcileContract' -count=1"
  expect: 0
hitl: false
hitl_note: >-
  The grant is a runtime input supplied by the operator when the lane runs, not
  a build-time input to this slice. Building the grant mechanism is agent-owned
  work. The lane refuses to merge without a valid grant.
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Implement grant validation, predicate inheritance, delegated delivery, and delegated reaping."
human_action: ""
can_continue_other_slices: false
protected_oracles:
  - path: cmd/kbcheck/rehab_delivery_test.go
    purpose: "Negative security fixtures prove that self-supplied proof, absent adapters, unenumerated pairings, protected paths, and unowned refs can never reach a merge."
  - path: cmd/kbcheck/terminal_cleanup_test.go
    purpose: "The existing safety corpus must still pass unchanged, proving this slice weakened no predicate."
expected_files:
  - path: cmd/kbcheck/rehab_delivery.go
    op: create
    scope: "Grant parsing and validation, predicate inheritance from ActionMerge, delivery eligibility, delegation to kb-complete and kbreconcile, and receipt emission."
  - path: cmd/kbcheck/rehab_delivery_test.go
    op: create
    scope: "Negative security fixtures and the granted happy path."
  - path: .github/skills/kb-rehab/SKILL.md
    op: modify
    scope: "Add the delivery and reaping sequence and the grant contract."
  - path: .github/skills/kb-rehab/references/grant.md
    op: create
    scope: "Lazy reference for the grant record fields and refusal rules."
  - path: config/rehab-policy.json
    op: modify
    scope: "Declare the native check adapter and the protected-path list."
---

# Slice 003 - Granted Delivery and Delegated Reaping

## Observable Outcome

With a valid grant, owned unshipped pairings reach merged and their terminal
artifacts are reaped by `kbreconcile`. Without one, the run stops at PR and
emits the approval packet. In every case, no predicate was satisfied by the
absence of a thing.

## Trust Boundary

Merging to this repository's default branch propagates `.github/skills/**` to
`~/.codex/skills/`, `~/.copilot/skills/`, and `~/.agents/skills/`. Every merge
decision here is a supply-chain decision. Design accordingly.

## Ordered Steps

1. **Validate the grant.** Require run ID, operator identity, issue time, expiry
   within the plan TTL, immutable evidence cutoff, per-action caps defaulting to
   zero, and an enumerated list of pairings by ref name, tip SHA, and PR
   identity (R14). Void the grant if the run ID, cutoff, or any listed tip SHA
   no longer matches observed state. Never persist it, never read one written by
   an earlier run.
   *Pass criterion:* fixtures for expired, replayed, tip-SHA-drifted, and
   unenumerated-pairing grants each produce zero merges.

2. **Filter by state before anything else.** Only `unshipped` pairings with a
   present, parsed, attributable declared work item are delivery-eligible.
   `orphan-branch`, `orphan-work`, `dead`, `superseded`, `live`, and
   `human-required` are never delivered under any grant (R13a). A ref whose tip
   commit is not authored or pushed by an identity the grant names is reported
   only (R13b).
   *Pass criterion:* an unowned remote ref fixture asserts zero proof
   executions, zero pushes, and zero PR creations for that ref.

3. **Resolve proof from default only.** The proof command comes from the
   authoritative remote default tree or from repo policy on default, never from
   the candidate branch's tree, manifest, frontmatter, or any file it introduces
   or modifies (R12). Execute with no forge credentials in the environment. A
   nonzero exit, a timeout, or a branch attempt to modify the resolved command
   classifies `human-required` and forbids merge.
   *Pass criterion:* a fixture branch shipping its own manifest with
   `proof: exit 0` asserts the self-supplied command is never executed and the
   pairing is ineligible.

4. **Inherit the predicate set.** Consume `ActionMerge` from
   `internal/reconcile/policy.go` and require every mandatory predicate to hold:
   `explicit-merge-authority`, `exact-pr-head-base`, `required-checks-green`,
   `required-reviews-green`, `fresh-mergeability`, `remote-default-authority`,
   `not-post-cutoff`, `exact-final-tree` (R15). Do not restate or fork the list.
   Record the predicate-manifest version in the receipt.
   *Pass criterion:* a test asserts the slice's consumed predicate set is
   identical to `ActionMerge`'s mandatory set, and fails if `policy.go` gains a
   predicate this path does not consult.

5. **Adapter absence is a blocker.** No reachable check adapter, review adapter,
   or forge adapter means an unsatisfied mandatory predicate and merge
   ineligibility. Removal, disablement, or absence of CI configuration is a
   blocker, never a pass (R15, R6).
   *Pass criterion:* a fixture with no CI and no declared native gate asserts
   zero auto-merges.

6. **Permit the native gate, narrowly.** Where no forge CI exists, a
   repository-native deterministic gate declared in `config/rehab-policy.json`
   and resolved from the authoritative default may satisfy
   `required-checks-green`, but only for a pairing that also holds R13a
   ownership and R13b attribution (R15a). In this repository that gate is
   `go run ./cmd/kbcheck local-release`. Name the satisfying adapter in the
   receipt.
   *Pass criterion:* a fixture asserts the native gate satisfies the predicate
   for an owned pairing and does not for an unowned one.

7. **Stop on protected paths.** A pairing whose diff touches
   `.github/skills/**`, `.github/agents/**`, `.github/instructions/**`,
   `cmd/**`, `internal/**`, `scripts/**`, or `config/skill-quality.json` is not
   auto-merge-eligible under any grant; it reaches an open PR and enters the
   packet (R14b). Where a grant permits merge for any pairing in this
   repository, `go run ./cmd/kbcheck local-release` must additionally pass on
   the post-merge default tree.
   *Pass criterion:* a fixture branch modifying a `SKILL.md` asserts an open PR
   and zero merges under a valid grant.

8. **Cap and fence.** Absent an explicit cap, merge and delete caps are zero. A
   grant may raise a cap only up to the shipped policy ceiling, may never enable
   a disallowed action, and may never substitute for a failed predicate (R14c).
   Acquire the same repo lock namespace used by `kbreconcile`, the `kb-start`
   sweep, and `kbcheck terminal-cleanup`, hold it across final re-verification,
   discard stale evidence after acquisition, and re-run the predicates (R16a).
   Record `contended` and skip on lock timeout.
   *Pass criterion:* a concurrent-writer fixture asserts the second actor
   records `contended` and mutates nothing.

9. **Delegate delivery, then reaping.** Delivery goes through `kb-complete` and
   its configured policy; the lane never opens a PR directly (R13). Reaping goes
   through `kbreconcile plan`/`apply`/`verify`; the lane adds no deletion path
   and never deletes a remote ref by plain Git inference (R16).
   *Pass criterion:* a test asserts the lane invokes the delegated paths and
   contains no direct ref-deletion or PR-creation call.

10. **Refuse sync.** A grant never authorizes post-merge sync, propagation to
    global install targets, release, or publication. The run report states
    `Sync: not-authorized` after every granted merge (R14a).
    *Pass criterion:* a granted-merge fixture asserts the literal
    `Sync: not-authorized` line and zero writes to any global skills path.

## Acceptance Criteria

- No predicate is ever satisfied by the absence of an adapter, a config file, or
  a CI directory.
- No proof command is ever resolved or executed from a candidate branch's tree.
- No pairing is merged that the grant does not enumerate by ref and tip SHA.
- No pairing touching a protected path is auto-merged under any grant.
- The existing `terminal-cleanup` and reconciler test corpora pass unchanged.
- An ungranted run performs zero merges, zero deletions, zero proof executions
  on unowned refs, zero pushes for unowned refs, and zero PR creations for
  unowned refs.

## Test Scenarios

| Scenario | Expected |
|---|---|
| Branch supplies its own manifest and proof command | command never executed, pairing ineligible |
| No CI, no declared native gate | zero auto-merges |
| Native gate declared, owned and attributed pairing | predicate satisfied, adapter named in receipt |
| Native gate declared, unowned ref | predicate unsatisfied |
| Branch modifies `.github/skills/**` under a valid grant | open PR, zero merges |
| Grant expired past plan TTL | void, stop at PR |
| Grant replayed from an earlier run | void, stop at PR |
| Grant enumerates ref but tip SHA has moved | void for that pairing |
| Pairing satisfies every predicate but is not enumerated | never merged |
| Second concurrent actor holds the repo lock | `contended`, zero mutation |
| Granted merge completes | `Sync: not-authorized`, zero global-root writes |
| `policy.go` gains a new mandatory `ActionMerge` predicate | predicate-parity test fails |

## Scope Boundary

This slice adds no inventory engine, no deletion engine, and no merge engine. It
validates a grant, inherits an existing predicate manifest, and delegates. Any
new predicate logic here is a defect.

## Requirements Owned

R12, R13, R13a, R13b, R14, R14a, R14b, R14c, R15, R15a, R16, R16a.
