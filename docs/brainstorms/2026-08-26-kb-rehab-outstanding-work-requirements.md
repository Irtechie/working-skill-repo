---
date: 2026-08-26
topic: kb-rehab-outstanding-work
brainstorm_style: kb-brainstorm
status: ready-for-plan
workflow_shape: skill-bundle-change
---

# KB Rehab - Outstanding Work Reconciliation

## Problem Frame

A KB portfolio accumulates two independent kinds of residue, and only one of
them has an owner today.

`cmd/kbreconcile` owns **artifact** residue: worktrees, sessions, queue claims,
receipts, and refs whose work is already terminal. Its requirements
(`docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md`)
deliberately scope out semantic judgment.

Nothing owns **work** residue: a declared workstream whose plan is dead, a
manifest superseded by a later one, a branch carrying real unmerged commits
that never got a PR, or a branch with no declared owner at all. The operator
discovers these by reading `todo.md` next to `git branch` and pairing them by
hand.

This repository is its own evidence. `todo.md` declares nine active
workstreams - two blocked, one skipped and explicitly marked superseded, one
pointing at a session that has ended - while the repository carries roughly ten
unmerged refs across `codex/*`, `deaderestpool-*`, `preserved/*`, and
`parked/*`. No artifact today states which declared stream owns which ref, or
which ref owns no stream.

The target operator wants one lane that answers "what is actually outstanding,
and get it to clean" without hand-pairing, and without a cleanup pass that
deletes work whose only proof of death was its age.

## Research Summary

**Findings that shaped requirements:**

- `kbreconcile` exposes `dry-run`, `plan`, `apply`, `verify`,
  `claim-capability`, and `claim-conformance`. Inventory, cutoff-bound
  planning, risk budget, quarantine, decision packets, and checked apply
  already exist and are proven by `internal/reconcile/reconcile_test.go`.
- `kb-start` already calls the reconciler on startup to reap terminal work, and
  falls back to `kbcheck terminal-cleanup --action sweep` when the global binary
  is absent. Startup reaping is solved; startup *assessment* is not.
- `todo-triage` already defines the decision taxonomy - ready, blocked, parked,
  duplicate, delete - but it reads only `todo.md` and legacy CE todo files. It
  has no git input and no notion of a superseding artifact.
- `kb-complete` already owns configured delivery and distinguishes local,
  open-PR, and integrated endpoints. `kb-ship` opens PRs; `kb-land` verifies the
  remote default branch contains the delivered commit.
- `kbcheck` already carries `session-preserve`, `terminal-cleanup`, and
  `run-state` as repo-local deterministic subcommands, establishing the pattern
  for a new read-only report subcommand.
- `todo.md` already encodes supersession in prose: the "Plan-to-PR finish lane"
  row reads `⊘ skipped` with `Superseded by <manifest path>`. Supersession is
  therefore parseable from existing convention, not a new field to invent.
- The reconciler's scope boundaries forbid inferring merge authority from
  repository ownership and forbid auto-closing long-lived PRs by age. Any
  delivery behavior in this lane must carry its own explicit grant.

**Confidence:** High for the engine boundary and the delivery chain, because
both are represented in source and tests. Medium for PR/CI state, which depends
on the `gh` adapter being present and authorized.

## Recommended Architecture

Three layers, each reusing an existing engine:

```text
declared work                    git reality
todo.md Active Work              local + remote refs
docs/plans/*-manifest.md         containment vs default branch
docs/context/goals/*             open PRs (gh adapter, optional)
docs/handoffs/active/*           work-queue claims
        |                                |
        +--------------+-----------------+
                       v
        kbcheck work-reality  (read-only, deterministic)
                       |
                       v
             classified pairings
   live | unshipped | dead | superseded | orphan-branch | orphan-work
                       |
        +--------------+--------------+
        v              v              v
   kb-complete     todo.md         kbreconcile
   (deliver)       markers         plan/apply/verify
                                   (reap)
```

`kb-rehab` is the orchestrating skill. It owns no engine. It reads the report,
applies triage, drives delivery, and hands terminal artifacts to the reconciler.

## Lifecycle Model

Each declared-work/ref pairing resolves to exactly one state:

| State | Evidence required |
|---|---|
| `live` | Any work-queue claim in `queued`/`in_progress`/`active` exists for the ref, branch, or worktree, **regardless of heartbeat freshness**, or manifest status `in_progress` with commits after the last integration |
| `unshipped` | Ref has commits that are neither ancestors of the authoritative remote default **nor patch-equivalent to it** (`git cherry <remote>/<default> <branch>` yields zero `+` lines ⇒ not `unshipped`), tree is clean, and no open PR exists |
| `superseded` | Declared work nominates a superseding artifact that exists **and** the superseded work's uncontained commits are proven contained in, or patch-equivalent to, that artifact's delivered ref |
| `dead` | Ref is proven contained in the authoritative remote default branch per R4/R4a **and** its declared work is `done`/`skipped`. There is no second path to `dead`. |
| `orphan-branch` | Ref has uncontained commits and no declared work item references it |
| `orphan-work` | Declared work item has no ref, no commits, and no manifest |
| `human-required` | Any pairing whose evidence is missing, stale, self-attesting, or contradictory |

Age is never evidence. Absence of a predicate preserves. Heartbeat staleness is
age wearing a different name and never confers takeover authority.

## Requirements

### R1-R6: Assessment

- **R1** A read-only subcommand pairs every declared work item with every local
  and remote ref in the current repository and emits one classification per
  pairing, with the evidence that produced it.
- **R1a** Every run captures an immutable evidence cutoff. Any ref, claim,
  worktree, PR, or `todo.md` row created, moved, dirtied, or updated after the
  cutoff is protected for that run and must be reclassified from fresh evidence
  in a later run. `not-post-cutoff` is a mandatory predicate for every delivery
  and marking action.
- **R1b** The pairing report and any grant or decision-packet record inherit
  reconciler R23a: minimum fields, redaction of protected-path and
  credential-like details, least-privilege permissions, and policy-defined
  retention.
- **R2** Declared work is read from `todo.md` Active Work, Queued Improvements,
  Blocked, and Parked sections, `docs/plans/*-manifest.md` frontmatter status,
  `docs/context/goals/*`, and `docs/handoffs/active/*`.
- **R3** Supersession is *nominated* by `todo.md` prose convention
  (`Superseded by <path>`) or manifest `status: superseded`, and *held* only
  when the named artifact exists **and** the superseded work's uncontained
  commits are proven contained in, or patch-equivalent to, the superseding
  artifact's delivered ref per R4/R4a. A nomination without that containment
  proof classifies `human-required` and enters the packet. A supersession claim
  introduced by the same commit that creates the artifact it names is never
  self-proving.
- **R4** A ref is classified `dead` only on containment proven against a
  *freshly fetched* authoritative remote default: resolve default via
  `ls-remote --symref <remote> HEAD`, fetch, and require the fetched ref to
  equal the advertised SHA; reject if the recorded default SHA is no longer an
  ancestor of the current default. Remote-tracking refs, local symbolic
  defaults, and `@{u}` are not authority. Where multiple remotes are configured,
  every configured remote is resolved and the strictest outcome applies. A
  repository with no configured remote, an unreachable remote, or an
  unresolvable default **fails closed**: no pairing may be classified `dead` and
  no delivery may occur in that run. The report must call `internal/reconcile`'s
  remote-authority and monotonic-default helpers rather than reimplementing
  containment; a divergent second implementation is a defect.
- **R4a** Containment is proven by ancestry **or** patch equivalence against the
  freshly fetched authoritative default, per `kb-complete`'s portable merge
  proof. Never invert the patch-equivalence test to act on `+`. A branch that is
  patch-equivalent to default is `dead`, never `unshipped`, and is never a
  delivery candidate.
- **R5** Missing evidence downgrades to `orphan-*`, `live`, or `human-required`,
  never to `dead`.
- **R5a** A named manifest, session, handoff, or goal that cannot be found is
  *missing evidence*, not proof of death. It classifies the declared work
  `orphan-work` and the paired ref `orphan-branch`, both of which preserve.
  Absence of a file may never produce a `dead` or `superseded` classification,
  and may never authorize a `todo.md` removal under R8.
- **R6** The report runs without `gh`, without network, and without
  `kbreconcile`; a missing adapter removes conclusions, never adds them. **This
  polarity governs every phase of the lane — assessment, triage, delivery, and
  reaping. No requirement may make adapter absence satisfy, skip, or weaken a
  predicate.**

### R7-R11: Triage

- **R7** `kb-rehab` writes `todo.md` markers for `superseded` and `dead`
  declared work, preserving the existing status-marker vocabulary.
- **R8** Removal of a declared work item requires the superseding or completing
  artifact to be named in the same edit **and** the item's paired ref to hold no
  commits uncontained in the default branch. Any uncontained commit blocks
  removal; the item is re-marked, never removed.
- **R9** Ambiguity is batched into one bounded decision packet of at most five
  grouped items. Each decision carries the recommended choice, affected
  artifacts, exact evidence and uncertainty, **irreversible consequence**, safe
  default, and expiry/recheck sensor. For any merge decision in this repository,
  the irreversible-consequence field must state the global install targets
  affected per `AGENTS.md`.
- **R10** Unanswered packet items preserve the work item and the ref.
- **R11** The lane never writes implementation code. It classifies, marks,
  delivers, and reaps.

### R12-R16: Delivery and reaping

- **R12** Proof for an `unshipped` pairing is executed from a proof command
  resolved **only** from the authoritative remote default branch's tree, never
  from the candidate branch's tree, manifest, frontmatter, or any file it
  introduces or modifies. If the default branch names no proof command for that
  scope, the pairing is reported and is not delivery-eligible. Proof execution
  runs with no forge credentials in the environment. A nonzero exit, a timeout,
  or an attempt by the branch to modify the resolved proof command classifies
  the pairing `human-required` and forbids merge for that run.
- **R13** Delivery uses `kb-complete` and its configured policy. `kb-rehab`
  never opens a PR directly.
- **R13a** Only a pairing in state `unshipped` whose declared work item is
  present, parsed, and attributable to this operator's own KB workstreams is
  delivery-eligible. `orphan-branch`, `orphan-work`, `dead`, `superseded`,
  `live`, and `human-required` pairings are never delivered under any grant. An
  `orphan-branch` is reported and preserved; it may be adopted only by an
  explicit decision-packet answer naming the ref, and adoption does not itself
  confer merge authority for that run.
- **R13b** A ref whose tip commit is not authored or pushed by an identity the
  grant names is reported only. The lane never delivers a ref it did not observe
  a declared KB workstream produce.
- **R14** Auto-merge and auto-delete require a grant binding: run ID, operator
  identity, issue time, an expiry not exceeding the plan TTL, an immutable
  evidence cutoff, per-action caps defaulting to zero, and **the exact
  enumerated list of pairings (ref name + tip SHA + PR identity) it
  authorizes**. A grant is single-use and non-replayable: a grant whose run ID,
  cutoff, or any listed tip SHA no longer matches observed state is void and the
  run stops at PR. A pairing not enumerated is never merged, even if it
  satisfies every predicate. Grants are never persisted as policy and never
  inherited by a later run. Without a valid grant the lane stops at PR and emits
  the approval packet.
- **R14a** A grant never authorizes post-merge sync, propagation to global
  install targets, release, or publication. Those remain separately authorized
  per `AGENTS.md`, and the run report must state `Sync: not-authorized` after
  every granted merge.
- **R14b** A pairing whose diff touches `.github/skills/**`, `.github/agents/**`,
  `.github/instructions/**`, `cmd/**`, `internal/**`, `scripts/**`, or
  `config/skill-quality.json` is not auto-merge-eligible under any grant. It
  reaches an open PR and enters the decision packet. Where a grant permits merge
  for any pairing in this repository, `go run ./cmd/kbcheck local-release` must
  pass on the **post-merge default tree**, resolved from default per R12, as an
  additional mandatory predicate.
- **R14c** Absent an explicit cap the merge cap is zero and the delete cap is
  zero. A grant may raise a cap up to the ceiling in
  `internal/reconcile/policy.go`'s risk budget but never above it, may never set
  `Allowed` for an action the shipped policy disallows, and may never substitute
  for a failed predicate. Budget exhaustion defers; budget availability
  authorizes nothing.
- **R15** A grant bounds *asking*, never *evidence*. Auto-merge fires only when
  every mandatory predicate in the versioned merge predicate manifest of
  `internal/reconcile/policy.go` (`ActionMerge`) holds for that exact pairing —
  `explicit-merge-authority`, `exact-pr-head-base`, `required-checks-green`,
  `required-reviews-green`, `fresh-mergeability`, `remote-default-authority`,
  `not-post-cutoff`, `exact-final-tree` — plus default-resolved proof per R12
  and exact remote containment re-verified after merge. **Adapter absence never
  satisfies a predicate.** A repository with no reachable check adapter, no
  review adapter, or no forge adapter has an unsatisfied mandatory predicate and
  is merge-ineligible for that run. Removal, disablement, or absence of CI
  configuration is a blocker, never a pass. The plan and receipt must name the
  predicate-manifest version consumed.
- **R15a** A repository-native deterministic gate may serve as the check adapter
  when no forge CI exists, provided it is resolved from the authoritative remote
  default per R12 and never from the candidate branch. In this repository that
  gate is `go run ./cmd/kbcheck local-release`. Satisfying
  `required-checks-green` this way requires the pairing to also hold R13a
  ownership and R13b identity attribution; an unowned or unattributed ref is
  never eligible for native-gate substitution. A repository with neither a forge
  CI adapter nor a declared native gate has an unsatisfied predicate and is
  merge-ineligible. The receipt must name which adapter satisfied the predicate.
- **R16** Reaping is delegated to `kbreconcile plan`/`apply`/`verify`. The lane
  adds no deletion path of its own and never deletes remote refs by plain Git
  inference.
- **R16a** The lane acquires the same repo lock namespace used by
  `kbreconcile`, the `kb-start` sweep, and `kbcheck terminal-cleanup`, and holds
  it across the final re-verification preceding any mutation. Failure to acquire
  within the bounded wait records `contended` and skips the repository. After
  acquisition, stale classification evidence is discarded and R15's predicates
  are re-run.

### R17-R19: Safety

- **R17** The lane operates on the current repository only, including its
  orphaned and unreferenced branches. Sibling checkouts are reported read-only
  and never mutated.
- **R18** The following are preserved unconditionally and are never classified
  `dead`/`superseded`, never delivered, and never reaped: the current executing
  session identified by **both** host session ID **and** canonical resolved real
  path (symlinks evaluated, Git-admin round-trip proven bidirectionally); its
  branch and worktree; the primary and default checkouts; any locked, moved, or
  recreated worktree; any tracked, untracked, or ignored dirt; and the branch
  and worktree of **any other session holding a work-queue claim in a
  non-terminal state**. This set must equal or exceed
  `terminalCleanupSafetyPredicates()` in `cmd/kbcheck/terminal_cleanup.go`;
  where both guards contribute predicates, the union and stricter outcome apply.
- **R18a** A stale heartbeat is a review signal, never takeover authority. A
  pairing whose claim is stale classifies `human-required` and enters the
  packet; it is never delivered, marked `dead`, or reaped in that run.
- **R19** Existing `terminal-cleanup` and reconciler predicates are reused
  as-is. This lane may not weaken them, and may not reimplement containment,
  remote authority, or safety predicates in a second code path.

## Success Criteria

- Running the lane on this repository pairs all nine declared workstreams and
  all unmerged refs, with zero hand-pairing.
- The `⊘ skipped / Superseded by` row is detected as a supersession *nomination*
  from existing prose, and resolves to `superseded` only with R3 containment
  proof, otherwise `human-required`.
- A fixture repository with mixed live, unshipped, dead, superseded, and orphan
  artifacts preserves every annotated unique artifact.
- Ungranted runs produce zero merges and zero deletions.
- An ungranted run against a fixture containing an unowned remote ref performs
  zero proof executions, zero pushes, and zero PR creations for that ref; it
  reports only.
- A fixture branch supplying its own manifest and proof command cannot satisfy
  R12; the resolved command comes from default or the pairing is ineligible.
- A fixture repository with no CI adapter and no declared native gate produces
  zero auto-merges; a fixture with a default-resolved native gate auto-merges
  only owned, attributed, non-protected-path pairings.
- A fixture branch that is patch-equivalent to default classifies `dead`, never
  `unshipped`, and is never delivered.
- A fixture branch touching `.github/skills/**` is never auto-merged under a
  valid grant.
- Granted runs merge only pairings holding every R15 predicate and enumerated in
  the grant.

## Scope Boundaries

- Do not reimplement inventory, planning, apply, verify, quarantine, or the risk
  budget. Delegate to `kbreconcile`.
- Do not reimplement PR creation, merge verification, or post-merge sync.
  Delegate to `kb-complete`, `kb-ship`, and `kb-land`.
- Do not mutate sibling repositories or any ATV checkout.
- Do not infer merge authority from repository ownership or from an earlier
  request to plan this lane.
- Do not use age, inactivity, or branch count as evidence of death.
- Do not make deletion volume or branch count a success metric.
- Do not add a daemon, MCP server, or required network adapter.

## Key Decisions

- **Assess, then converge:** the primary product is the pairing report. Marking,
  delivery, and reaping are downstream actions over its output.
- **Two engines, one lane:** `kbreconcile` owns artifacts, `kb-complete` owns
  delivery, `kb-rehab` owns only the semantic pairing and the sequencing.
- **Supersession is already written down:** parse the existing convention rather
  than introducing a field the operator must maintain.
- **The grant bounds asking, not proof:** the user authorized auto-merge and
  auto-delete on 2026-08-26; that removes prompts for proven-terminal work and
  changes no predicate. A grant relocating an action to a code path with fewer
  predicates is the failure mode this decision exists to forbid.
- **Blast radius is the bundle, not the repo:** merging to this repository's
  default branch propagates to `~/.codex/skills/`, `~/.copilot/skills/`, and
  `~/.agents/skills/` per `AGENTS.md`, changing the instruction set of every
  future agent session on the host. Delivery decisions in this lane are
  supply-chain decisions, and this repository has no CI adapter today.

## Dependencies / Assumptions

- **[safe-assumption]** `todo.md` conventions are stable enough to parse.
  Reversible because unparsed rows classify as `orphan-work` and surface in the
  packet rather than being dropped. Evidence/proof: fixtures containing
  malformed and unconventional rows must round-trip to the packet.
- **[safe-assumption]** `gh` is available and authorized in the operator's
  environment for PR and CI state. Reversible because its absence downgrades
  `unshipped` conclusions and forbids auto-merge. Evidence/proof: a fixture with
  the adapter disabled must preserve, not merge.
- **[safe-assumption]** `internal/reconcile` git helpers can be reused for ref
  and containment evidence. Reversible because the report may call plumbing
  directly. Evidence/proof: the report runs in a fixture repo without
  duplicating containment logic.

## Alternatives Considered

- **Extend `todo-triage` instead:** rejected because triage is deliberately
  git-blind and disable-model-invocation, and adding refs, containment, PR
  state, and delivery to it would make one skill own two lanes.
- **Extend `kbreconcile` instead:** rejected because its own requirements scope
  out semantic judgment, and adding `todo.md` parsing would make the global
  binary depend on KB document conventions it is designed to work without.
- **Ship it in `agent-marketplace`:** rejected because the lane depends on
  `kbreconcile`, `kbcheck`, and six KB skills versioned in this bundle; a copy
  drifts immediately.
- **Skill-only, no runtime:** rejected because pairing, containment, and
  supersession detection are deterministic and belong in `kbcheck` per the
  repository's markdown-to-runtime extraction rule.
- **Delete refs older than a threshold:** rejected for the same reason the
  reconciler rejected it - age does not prove delivery or duplication.
- **One slice for the whole lane:** rejected because the read-only report must
  be provable before any destructive action consumes its output.

## Slice Candidates

- Deterministic read-only work/ref pairing report in `kbcheck`.
- `kb-rehab` skill lane: triage, `todo.md` markers, decision packet, routing.
- Granted delivery and reaping loop over the report's output.

## Outstanding Questions

### Resolve Before Planning

None.

### Deferred to Planning

- **[defer-to-planning][Affects R2][Technical]** Decide the exact declared-work
  parse surface and how unparsed rows are represented.

R14's grant format deferral is resolved inline by R14/R14a/R14b/R14c.

### Parked / Out Of Scope

- **[parked][Affects R17]** Cross-repository rehab across sibling checkouts.
  The first version reports siblings read-only.
- **[parked][Affects R16]** Remote feature-ref deletion, which stays with the
  reconciler's existing parked provider-race-safety requirement.

## Evidence

- `cmd/kbreconcile/main.go`, `internal/reconcile/inventory.go`
- `docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md`
- `.github/skills/kb-start/SKILL.md`, `.github/skills/todo-triage/SKILL.md`
- `.github/skills/kb-complete/SKILL.md`
- `todo.md` Active Work table, 2026-08-26 repository ref listing

## Next Steps

`kb-plan docs/brainstorms/2026-08-26-kb-rehab-outstanding-work-requirements.md`
