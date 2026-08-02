---
date: 2026-08-01
topic: global-cleanup-reconciliation
brainstorm_style: kb-brainstorm
status: ready-for-plan
workflow_shape: pipeline-change
---

# Global Cleanup And Reconciliation

## Problem Frame

Autonomous KB/Copilot workflows can leave a portfolio of sessions, worktrees,
branches, pull requests, queue claims, receipts, and local-only files after the
useful work has finished. Asking the user to classify every artifact defeats the
purpose of autonomy, but treating "autopilot" as permission to delete
automatically risks unique-data loss.

The target operator is a developer or maintainer who wants the harness to
converge routine terminal work with near-zero attention while preserving
anything active, unique, ambiguous, policy-protected, or insufficiently proven.
The reconciler must work across repositories and hosts even when a repository
does not contain `cmd/kbcheck`.

Autopilot means automatic evidence collection, classification, reversible
salvage, policy enforcement, and proof. It does not mean deletion-first cleanup.

## Research Summary

**Findings that shaped requirements:**

- `kb-work` ends execution by invoking `kb-finalize`; `kb-finalize` proves the
  exact tree, cleans only run-owned ephemeral artifacts, and invokes
  `kb-complete`. Delivery remains owned by `kb-complete`, `kb-ship`, and
  `kb-land`, not by work or finalization.
- `kb-complete` already distinguishes local, open-PR, and integrated delivery.
  It registers terminal cleanup only after durable local-branch, pushed-topic,
  or remote-default evidence exists.
- `cmd/kbcheck terminal-cleanup` is deliberately narrow and repo-local. It
  protects the current session and worktree, the primary/default checkout,
  active claims, locked/moved/recreated worktrees, tracked/untracked/ignored
  dirt, rewritten remote evidence, and uncontained commits. It uses non-force
  removal, fresh authoritative remote checks, and exact-SHA local-ref deletion.
- Open PR delivery is intentionally non-terminal for integration. The current
  guard may retire a clean worktree after exact remote-topic containment, but it
  preserves local and remote feature refs until a separately proven integrated
  endpoint exists.
- Host UI session records and race-safe remote feature-ref deletion are
  explicitly outside the current Git-only guard. This is an adapter boundary,
  not evidence that those records may be inferred from Git state.
- The shared work queue currently stores claims under the Git common directory.
  Terminal claims can outlive archived sessions and consume WIP capacity until
  a later actor reconciles them.
- A startup sweep during this brainstorm found an old receipt whose worktree
  registration and admin identity were both missing. The existing guard
  correctly failed closed, but it has no global portfolio repair/classification
  layer.
- The installer already has a precedent for a checksum-verified user-global
  binary under `~/.kb/bin`; skills themselves are not standalone binaries.
- Git worktrees isolate checkout contents, `HEAD`, and indexes, but ordinary
  refs and most repository state remain shared. Worktrees do not coordinate
  semantic responsibility, separate clones, publishers, deployments, or
  external side effects.
- Established schedulers preserve DAG parallelism by combining dependency edges
  with resource constraints. Kubernetes/etcd-style compare-and-swap claims
  select one writer, but lease expiry alone does not fence a paused stale
  worker. A protected mutation endpoint must reject stale monotonically
  increasing generations and idempotently recognize replay.

**Confidence:** High for the existing lifecycle and safety boundaries because
they are directly represented in source and tests. Medium for host session
retirement because available host APIs differ and must remain adapter-owned.

## Recommended Architecture

Use a user-global reconciliation control plane with a small deterministic core
and explicit adapters:

```text
host/repo/provider sensors
          |
          v
normalized evidence ledger -- immutable cutoff snapshot
          |
          v
policy + confidence + risk-budget classifier
          |
    +-----+--------+-----------+-------------+
    |              |           |             |
 preserve      quarantine    salvage    routine action plan
                                              |
                                              v
                                  CAS/fresh-check executor
                                              |
                                              v
                                  verifier + durable receipts
```

1. **Global core:** a checksum-managed tool available independently of any
   consuming repository. It inventories, plans, applies, and verifies using
   plain Git plus installed host/provider adapters.
2. **Normalized evidence ledger:** records canonical repository identity,
   session identity, worktree generation, refs and SHAs, PR identity/state,
   queue ownership, dirt classes, protection reasons, evidence timestamps, and
   source adapters. Classification consumes this ledger, not chat history.
3. **Policy engine:** combines non-overridable safety invariants, optional
   user-global policy, optional repo policy, provider protections, confidence
   thresholds, and a bounded per-run risk budget. Missing policy never expands
   destructive authority.
4. **Action planner:** emits dependency-ordered actions with preconditions,
   rollback or preservation behavior, and an evidence expiry. It groups
   ambiguity into a small decision packet rather than prompting per artifact.
5. **Checked executor:** reacquires authoritative evidence immediately before
   every mutation, uses exact identity and compare-and-swap where available,
   never force-removes, and stops only the affected dependency chain on drift.
6. **Verifier:** proves postconditions, writes append-only receipts, and keeps
   delivery, physical cleanup, ref retirement, and host session-record
   retirement as separate states.
7. **Adapters:** optional repo-native `kbcheck`, Copilot/session host, GitHub or
   other forge, queue/lease, and repo-policy adapters add evidence or actions.
   An unavailable adapter reduces authority; it never creates a success-shaped
   fallback.
8. **Semantic scheduler:** each DAG node declares operational resource keys,
   including `publisher:<product>`, `release-manifest:<product>`, and
   `deploy:<environment>`. Disjoint nodes may run in parallel; conflicting
   writers require one authoritative claim even when their paths do not overlap.
9. **Fenced side-effect gateway:** protected publication and deployment adapters
   accept only a current controller-issued generation, validate it immediately
   before the side effect, and atomically reject stale workers. Agents do not
   receive a bypass path around the gateway.

The exact executable name and package layout are planning decisions. The
required boundary is a global deterministic core with optional repo adapters,
not a skill that shells into a repository-specific `go run ./cmd/kbcheck`.

The trust model treats workers as potentially stale, buggy, or compromised
within their unprivileged execution boundary. The controller, policy trust root,
claim store, and fenced gateway are privileged components and must minimize
their APIs, authenticate each other, protect integrity, and retain auditable
evidence. No worker-provided ledger field, generation number, or summary is
authority by itself.

## Lifecycle Model

Delivery and cleanup are separate dimensions. A work item may move through:

| State | Meaning | Automatic behavior |
|---|---|---|
| `active` | A live session/claim owns mutation | Preserve; refresh heartbeat only |
| `local-durable` | Clean committed branch exists, no publication promised | May retire a clean worktree; preserve branch |
| `awaiting-review` | Exact topic commit is on the remote and PR remains open | May suspend/archive execution and retire a clean worktree; preserve local/remote refs and resume packet |
| `delivery-integrated` | Provider/Git proof shows accepted integration on remote default | Eligible for checked local-ref retirement under policy |
| `cleanup-eligible` | Delivery state plus all physical safety predicates pass | A later/different executor may apply the planned cleanup |
| `quarantined` | Unique or ambiguous work is preserved outside routine action | No destructive action; retry when named evidence appears |
| `physically-retired` | Worktree/ref action has verified receipts | Does not imply the host session record is retired |
| `session-retired` | Host adapter proved the execution record safe to archive/delete | Does not alter delivery history or receipts |

An open PR may remain in `awaiting-review` for weeks. Age alone never makes its
work unique, stale, superseded, merged, or deletion-eligible. Long review waits
must not consume an active mutation slot when exact remote-topic containment
and a durable resume packet allow the execution session to suspend safely.

### Self-Cleaning Workflow

- `kb-work` must continue to `kb-finalize`; neither phase may delete its current
  worktree or infer delivery.
- `kb-finalize` may remove only exact run-owned ephemeral artifacts and must
  return reviewed state to `kb-complete`.
- `kb-complete` must register lifecycle intent after the configured durable
  endpoint:
  - local completion registers `local-durable`;
  - manual/open PR registers `awaiting-review`;
  - authorized and proven landing registers `delivery-integrated`.
- The current session must release or suspend its active queue claim only after
  registration succeeds. It must never remove its own worktree.
- A different session, later `kb-start`, scheduled run, or explicit global
  reconcile command performs `plan -> apply -> verify`.
- A forge event or later poll may upgrade `awaiting-review` to
  `delivery-integrated` after fresh merge/containment proof. Only then may the
  integrated-ref policy run.
- Physical cleanup failure does not roll back or misreport proven delivery.
  Session-record retirement failure does not roll back physical cleanup.

## Requirements

### Inventory And Evidence

- **R1.** Inventory all discoverable Copilot/agent sessions, canonical repository
  identities, primary checkouts, linked worktrees, local and remote branches,
  PRs, queue claims and leases, lifecycle/cleanup receipts, tracked changes,
  untracked files, ignored files, remote defaults, and containment relationships.
- **R2.** Capture an immutable scan cutoff. Any artifact created, rewritten,
  checked out, dirtied, claimed, or updated after that cutoff is protected from
  mutation for that run and must be reclassified from fresh evidence.
- **R3.** Protect the current/primary session by both host session ID and resolved
  real worktree path. Protect primary/default checkouts, active processes and
  claims, locked worktrees, and paths whose Git-admin identity cannot be proven
  bidirectionally.
- **R4.** Treat tracked, untracked, and ignored files as distinct evidence.
  Ignored content is data, not permission to remove it. Credential-like,
  model/runtime, learning/memory, database, socket/lock, and other live-state
  paths must be preserved without reading secret contents and classified by
  metadata plus explicit adapters.
- **R5.** Resolve remote authority from the provider or authoritative Git remote,
  not local symbolic assumptions. Record exact observed SHAs and timestamps.
  Refresh remote/PR evidence immediately before every merge, close, ref
  deletion, worktree retirement, or session-record retirement that depends on
  delivery state.
- **R6.** Every conclusion must name its evidence sources, freshness, exact
  predicates, limitations, and confidence. LLM judgment, naming similarity,
  age, inactivity, or a successful prior run cannot independently prove
  duplication, containment, supersession, or safe deletion.
- **R6a.** Controller-owned policy, evidence ledgers, adapter assertions, claim
  records, receipts, and audit events require schema validation, authenticated
  source identity, capability-scoped write APIs, tamper-evident ordering, and
  rollback detection. Workers may propose evidence but cannot rewrite accepted
  authority or postconditions. A failed integrity check invalidates the affected
  evidence and downgrades dependent mutation to preserve or quarantine.

### Exact Deduplication And Containment

- **R7.** Deduplicate only with deterministic proof: identical blobs/trees for
  the relevant path set; a byte-identical, full-index, binary-capable patch
  stream computed from the same recorded merge-base and path set; commit
  ancestry; exact remote-topic containment; remote-default ancestry; or
  provider-backed PR merge identity combined with exact tree/patch-delta proof
  for squash/rebase. Record the merge-base, path set, proof algorithm/version,
  and resulting identity. Whitespace-normalizing patch IDs may nominate
  candidates but cannot authorize retirement.
- **R8.** Do not treat similar descriptions, overlapping filenames, equivalent
  intent, or model-generated summaries as exact duplicate proof. Fuzzy evidence
  may prioritize review but can only lead to preservation or quarantine.
- **R9.** When one artifact is exactly contained by another, preserve the
  containing artifact and its receipt before marking the redundant artifact
  superseded. Recheck both identities immediately before any mutation.

### Classification, Confidence, And Risk Budget

- **R10.** Classify each artifact or coherent group as `preserve-active`,
  `protected`, `routine-retire`, `safe-supersede`, `salvage`, `quarantine`, or
  `human-required`. Classification must be deterministic from the ledger and
  policy version.
- **R11.** Default thresholds are fail-closed:
  - destructive or integration actions require complete authoritative evidence
    for every mandatory predicate and computed confidence `1.00`;
  - reversible remote metadata actions require confidence at least `0.98`;
  - additive salvage actions require confidence at least `0.90`;
  - lower confidence, stale evidence, disagreement between adapters, or a
    missing mandatory predicate yields preservation or quarantine.
  Policy may raise thresholds but may not lower the `1.00` requirement for
  unique-data deletion, merge, or ref retirement.
- **R11a.** Publish a versioned mandatory-predicate manifest for every action
  class in the decision policy. It must identify each required predicate, the
  authoritative adapters capable of supplying it, optional enrichment
  predicates, freshness limits, and the exact downgrade on absence. A missing
  mandatory predicate blocks that action. Missing optional enrichment may
  reduce confidence but cannot convert a failed mandatory predicate into a
  pass. The plan and receipt must name the predicate-manifest version.
- **R12.** Confidence must be computed from evidence completeness, freshness,
  source authority, and agreement. It must not be an unconstrained model
  self-score.
- **R13.** Apply a bounded risk budget per run and per repository. The budget is
  an action-class cap, not permission to override a failed invariant. Inventory,
  receipt repair, preservation, and quarantine consume no destructive budget.
  Merge, PR close, ref deletion, worktree removal, and host-record deletion each
  require explicit policy allowance and consume their configured class budget.
  Exhaustion defers remaining actions without asking one question per artifact.
- **R13a.** Report portfolio convergence health after every run: eligible but
  budget-deferred count by action class, oldest deferred age, net backlog change
  since the previous verified run, and an evidence-based runs-to-convergence
  projection at the current cap. These metrics inform policy tuning but never
  increase authority or budget automatically.
- **R14.** Execute independent safe actions even when another artifact is
  blocked. Propagate a blocker only through actions that depend on that evidence
  or target.

### Decision Policy

| Decision | Safe automatic policy | Otherwise |
|---|---|---|
| Merge | Only `kb-land`/equivalent adapter has explicit configured authority; exact PR/head/base match; required checks and reviews are green; mergeability and remote default are freshly resolved; no post-cutoff drift; final tree proof is valid | Leave open or quarantine; never infer authority from write access |
| Close/supersede PR | Exact patch/blob/merge containment proves no unique code; comments/checks contain no unresolved policy-defined blocker; provider supports reversible reopen; policy permits close | Label/link as possible duplicate and preserve; human packet only if closure is policy-required |
| Retire worktree | Different executor; clean tracked/untracked/ignored state; terminal/suspended claim; exact worktree generation and Git-admin round-trip; durable branch/topic/integration proof; not current, primary, default, locked, moved, or post-cutoff; non-force removal succeeds | Preserve or quarantine with exact failed predicate |
| Retire local ref | Integrated endpoint is freshly proven; exact local ref still equals recorded SHA; ref is not default or checked out; compare-and-swap deletion succeeds; squash/rebase also has provider-backed PR merge identity plus exact tree/byte-preserving patch-delta equivalence | Preserve when any predicate is absent |
| Delete remote ref | Provider adapter supplies race-safe exact-ref mutation and repo policy explicitly permits it | Preserve; plain Git delete is insufficient |
| Quarantine | Work is unique/ambiguous, evidence is stale/missing, live/protected paths are present, remote history rewrote, or an adapter disagrees | Record resume sensor; do not count as cleanup failure |
| Salvage | Unique coherent work can be isolated without reading/committing credentials or live state; base and scope are provable; secret/protected-path checks pass; additive action is within budget | Preserve in place or private quarantine |
| Retire session record | Host adapter proves no active process/input, durable delivery/resume state exists, physical state no longer depends on the record, retention policy permits it, and exact record CAS/version still matches | Archive/suspend or preserve; Git evidence alone is insufficient |

### Salvage

- **R15.** Salvage coherent unique commits or safe dirty work before considering
  retirement. Group by canonical repo, base, intent/scope evidence, and
  dependency; never create one omnibus PR merely to reduce artifact count.
- **R16.** Salvage is additive and bounded. Create a protected salvage branch and
  at most the policy-configured number of focused PRs per run. Each PR must state
  provenance, base, included commits/paths, exclusions, proof state, and why it
  is not already contained elsewhere.
- **R17.** Do not automatically commit credential-like files, private model
  state, learning/memory databases, live locks/sockets, unknown binaries, or
  ignored runtime data. Preserve them in place or quarantine metadata-only and
  make the exact protected path class visible.
- **R18.** A failed salvage attempt must leave the source artifact and refs
  untouched. Successful salvage must be verified on the remote before the
  source becomes eligible for later physical cleanup.
- **R18a.** Creating or publishing a salvage branch/PR requires explicit
  `salvage.publish` authority in global or repo policy plus freshly verified push
  and PR-creation permission. Without that authority, salvage may create only a
  local protected ref or private quarantine receipt. Policy must cap both new
  salvage PRs per run and total live salvage PRs per repository.

### Human Attention

- **R19.** Ask a human only when an unresolved decision is irreversible,
  policy-defined, authority-bearing, or cannot be answered by available
  evidence. Agent-owned inventory, Git diagnosis, deduplication, proof, and
  retry work must not become user classification work.
- **R20.** Produce one decision packet per run, capped by policy and defaulting
  to five grouped decisions rather than one prompt per artifact. Each decision
  contains the recommended choice, affected artifacts, exact evidence and
  uncertainty, irreversible consequence, safe default, and expiry/recheck
  sensor. Excess ambiguity is quarantined for later runs.
- **R21.** The safe default for an unanswered or expired decision is preserve or
  quarantine, never merge, close, delete, overwrite, or retire.

### Modes, Receipts, And Recovery

- **R22.** Support:
  - `dry-run`: sensor-only summary with no durable mutations;
  - `plan`: write a cutoff-bound action plan and decision packet;
  - `apply`: execute only an unchanged, unexpired plan after fresh precondition
    checks;
  - `verify`: re-sense postconditions, reconcile partial outcomes, and emit the
    final ledger without adding new mutations.
- **R23.** Plans and receipts must bind schema/policy versions, host and repo
  identity, cutoff, evidence fingerprints, exact target identities, action
  preconditions, observed-before and observed-after state, actor/session,
  timestamps, and result. Repeated apply/verify must be idempotent.
- **R23a.** Treat global ledger and receipt metadata as sensitive operational
  data. Store the minimum fields, redact secret values and protected-path
  details from packets/logs, use least-privilege file/service permissions and
  authenticated transport, and apply policy-defined retention to claims,
  idempotency records, receipts, audit events, and backups. Retention expiry
  cannot erase evidence still required to prevent replay, generation rollback,
  or unique-data loss.
- **R24.** Ref deletion must use exact-SHA compare-and-swap. Worktree retirement
  must use non-force Git removal with bounded retries. Missing registration,
  replaced paths, non-empty residuals, identity drift, or failed postconditions
  must preserve data and create a repairable partial receipt.
- **R25.** Never use force worktree removal, recursive broad deletion, hard reset,
  force push, bypass flags, or deletion based solely on filesystem age.
- **R26.** Distinguish and report four independent outcomes:
  `delivery_state`, `physical_cleanup_state`, `ref_retirement_state`, and
  `session_record_state`. Success in one dimension must not hide a blocker in
  another or erase already-proven completion.
- **R26a.** Enforce an explicit action dependency graph. A proven durable
  endpoint (`local-durable`, exact remote topic, or integrated default) gates
  claim suspension/release and physical worktree retirement. Integrated-default
  proof additionally gates merged-ref retirement. Verified physical retirement
  gates local-ref deletion and destructive host session-record deletion.
  Non-destructive host suspension/archive may occur earlier when a complete
  resume packet exists. A blocker defers only its real dependents.

### Global Operation And Optional Adapters

- **R27.** The baseline reconciler must run from a user-global installation and
  handle repositories with no `cmd/kbcheck`. It must provide its own deterministic
  inventory, ledger, policy, CAS, receipt, and non-force Git safety core.
- **R28.** Repo-native `cmd/kbcheck`, queue schemas, manifests, and cleanup
  receipts are optional evidence/action adapters. When available and compatible,
  reuse them rather than duplicating mutation. When absent or incompatible,
  downgrade only the unsupported action.
- **R28a.** The global core and repo-native guard must consume one versioned
  safety-invariant contract or shared implementation, not drift-prone parallel
  policy. The global worktree-retirement predicate set must match or exceed the
  existing `terminal-cleanup` protections, including generation token,
  Git-admin round-trip, canonical real path, current/primary/default/locked
  exclusions, queue ownership, tracked/untracked/ignored inspection, exact
  branch/HEAD, fresh and monotonic remote evidence, non-force removal, and
  empty-residual-only partial reconciliation. When both guards contribute
  predicates, the union and stricter outcome apply. One conformance corpus must
  prove both entrypoints.
- **R29.** Host/session and forge actions require versioned adapters with declared
  capabilities. No adapter means no session-record deletion, PR mutation,
  provider merge proof, or remote-ref deletion; local inventory and preservation
  still proceed.
- **R30.** Merge user-global policy, optional repo policy, and run-scoped
  authority using least privilege. Repo policy may protect more or narrow
  automation. Only explicit configured delivery authority or same-run user
  authority may enable merge.
- **R31.** Support portfolio-wide inventory while serializing mutations by
  canonical Git common directory and by provider resource. Separate clones or
  machines are distinct until a provider-backed identity/coordination adapter
  proves otherwise.
- **R31a.** Global reconcile, `kb-start` sweep, queue mutation, and repo-native
  cleanup must use the same compatible repo lock namespace/protocol. An actor
  that cannot acquire the lock within its bounded wait must record `contended`
  and skip that repository. After lock acquisition it must discard stale plan
  evidence and rerun the action's mandatory predicates before mutation.

### Semantic Writer Coordination And Fencing

- **R31b.** Preserve DAG and worktree parallelism for proven-disjoint work, but
  require every mutating node to declare versioned canonical semantic read/write
  resource keys. Human-readable examples include `publisher:<product>`,
  `release-manifest:<product>`, and `deploy:<environment>`; the authoritative
  identity additionally binds provider, tenant/account, resource type, and
  provider-canonical resource ID. The schema defines Unicode normalization,
  case handling, escaping, maximum length, and collision rejection. Unknown
  aliases quarantine rather than creating a second authority domain.
  Conflicting writers serialize even when they edit different repositories,
  branches, or paths.
- **R31c.** A protected semantic writer requires an authoritative claim acquired
  through compare-and-swap against the current claim version. The claim binds
  resource key, holder/session, objective/work ID, source revision, claim
  revision, monotonically increasing fencing generation, issued/expiry times,
  adapter identity, controller incarnation, and authenticated workload identity.
  Local worktree or Git-common-directory leases may add coordination but cannot
  satisfy global authority.
- **R31d.** Lease expiry is a recheck signal, not takeover authority. A successor
  may become active only after authoritative compare-and-swap advances the claim
  revision and fencing generation. Read-then-write replacement, age-only
  takeover, and generation reuse are forbidden.
- **R31e.** Every publish, release-manifest promotion, deployment, or other
  policy-protected external side effect must carry the resource key, exact
  holder, fencing generation, source/manifest digest, idempotency key, and an
  unforgeable short-lived authorization bound to audience, controller/workload
  identity, work ID, operation, resource key, generation, request digest, and
  expiry. The authoritative endpoint owns a per-resource high-water generation
  and serializes authorization validation, generation admission, idempotency
  reservation, and side-effect commit, or uses a provider-native conditional
  mutation with equivalent atomicity. Stale, missing, expired, mismatched,
  forged, or unverifiable authority is rejected without mutation. Providers
  unable to enforce this boundary are unsupported for protected automation.
- **R31f.** Replaying the same idempotency key and identical payload returns the
  recorded result without repeating the side effect. Reusing a key with a
  different payload fails closed. The gateway atomically reserves the key and
  digest before side effects and exposes durable `in-flight`, `committed`, and
  `failed-safe` states. `unknown`, missing-after-admission, or still-in-flight
  status preserves the claim and blocks retry; it never proves the original
  request cannot commit. Receipts bind the claim revision, generation, request
  digest, endpoint result, and postcondition evidence.
- **R31g.** Failure recovery is bounded and dependency-aware. Ambiguous timeout
  or transport failure triggers status lookup by idempotency key before retry.
  Retry count and elapsed time are policy-capped; exhaustion preserves the claim
  or enters quarantine and blocks only dependent nodes. It must not create a
  second publisher, advance the generation speculatively, or ask the user to
  classify routine retry state.
- **R31h.** If the authoritative claim adapter or endpoint validation is
  unavailable, protected writer work may be planned or preserved but may not
  acquire authority or perform its side effect. The compact human exception
  packet remains reserved for unresolved irreversible ambiguity or
  policy-defined authority, not coordinator outages that have a safe retry.
- **R31i.** All credentialed production paths must pass through the fenced
  endpoint. Repository concurrency groups, merge queues, environments, and
  remote Git claims are defense in depth; none independently proves a global
  fence. Production IAM and network policy must admit only the gateway principal;
  agents receive gateway-invocation authority only; direct production
  credentials are revoked and rotated; alternate workflows lack a permitted
  principal. The adapter must prove these controls or downgrade protected
  mutation to unavailable. Conformance tests attempt direct and alternate-path
  bypass.
- **R31j.** The first implementation must expose a versioned provider-neutral
  claim/fence adapter contract and deterministic conformance fixtures. A concrete
  controller may implement it with a linearizable database, Kubernetes
  `resourceVersion`, etcd transaction, or equivalent CAS. No consuming
  repository must run a new daemon; missing global coordination fails closed.
- **R31k.** Claim and gateway authorization keys require a pinned trust root,
  rotation identifiers, bounded overlap, explicit revocation, and audience
  separation. Rotation or compromise advances the controller incarnation and
  invalidates outstanding worker authority unless a policy-defined,
  audit-recorded reissue proves the same claim and request.
- **R31l.** Claim storage disaster recovery cannot restore an old generation as
  current. Persist a monotonic controller incarnation or high-water epoch outside
  the claim snapshot rollback domain. On restore, reject lower epochs, invalidate
  outstanding authority, reconcile gateway high-water state, and block protected
  mutation until rollback checks pass.

### Proactive Prevention

- **R32.** Register every autonomous run at lifecycle start with canonical repo,
  session, worktree, branch, objective, owner, cutoff, heartbeat, and declared
  semantic resources. Mutation without durable registration fails before work;
  protected semantic writers additionally require R31b-R31j authority.
- **R33.** Enforce WIP caps on active execution owners, not accidentally on
  multiple claims belonging to the same owner. `awaiting-review`, terminal,
  quarantined, and suspended states must not consume active mutation capacity.
- **R34.** Require explicit session terminal states and exact resume conditions:
  `active`, `paused`, `awaiting-review`, `blocked`, `superseded`, `done`, and
  `retired`. Staleness is a recheck signal, never automatic takeover authority.
- **R34a.** Before retiring an `awaiting-review` worktree, write and validate a
  resume packet containing canonical repo identity, manifest/requirements
  pointers when present, work ID and claim identity, plan-run branch, delivered
  commit SHA, remote/topic ref and observed SHA, PR provider/repo/number/URL,
  current slice/gate and proof-receipt pointers, protected/quarantined paths,
  and exact environment recreation/resume commands. Missing fields block
  worktree retirement but not PR delivery.
- **R35.** A successful KB chain must register the correct durable endpoint and
  release/suspend ownership before terminal output. A later session or scheduled
  controller must sweep eligible receipts automatically; users must not need to
  remember a cleanup command.
- **R36.** Startup sweeping is opportunistic, not the only trigger. Provide
  scheduled and on-demand portfolio reconciliation so repositories with no
  later interactive session still converge.
- **R37.** Detect queue/session/PR divergence early: archived session with active
  claim, merged PR with `awaiting-review`, missing receipt for a terminal claim,
  or removed worktree with unreconciled registration. Repair metadata
  autonomously only when host/provider identity and terminal evidence are exact.
- **R37a.** Releasing an orphaned active claim requires authoritative host proof
  that the exact owning session is terminal or absent, no live process/heartbeat,
  and compare-and-swap of the unchanged work ID, session ID, branch/worktree,
  status, and `updated_at`. Transient or retryable host-adapter outage remains
  preserved or quarantined with a retry sensor. Only proven capability absence
  or unresolved authority-bearing ambiguity may enter the grouped decision
  packet after a configurable threshold (default 72 hours); age never authorizes
  takeover or release.

## Success Criteria

- A fixture portfolio of at least 20 mixed sessions/worktrees/branches/PRs is
  reduced to the policy-safe minimum with no per-artifact prompts. Every fixture
  artifact annotated as containing unique work remains in salvage, quarantine,
  or preservation after apply and verify. No artifact lacking exact
  deterministic containment proof is retired.
- Routine exact cases complete in `plan -> apply -> verify` without human input;
  ambiguous cases produce at most one five-decision packet and otherwise
  quarantine safely.
- Current, primary, active, protected, post-cutoff, dirty, ignored-data,
  credential-like, model/runtime, learning/memory, locked, moved/recreated,
  rewritten, uncontained, and adapter-unproven artifacts remain untouched.
- An open PR can remain awaiting review for weeks without losing local/remote
  recovery refs and without consuming an active WIP slot. Its clean worktree may
  be reclaimed only after exact remote-topic containment and durable resume
  state are proven, and the fixture can recreate the worktree and resume from
  the packet without human archaeology.
- Exact duplicate branches/PRs are superseded only by deterministic
  blob/patch/ancestry/provider evidence; fuzzy similarity never authorizes
  mutation.
- Squash/rebase retirement requires provider-backed PR merge identity plus exact
  tree or byte-preserving patch-delta equivalence; provider state or normalized
  patch similarity alone retains the ref.
- Unique coherent work is salvaged into bounded focused PRs or preserved in
  quarantine before any source retirement.
- Authorized auto-merge occurs only with fresh exact PR/base/head, green gates,
  explicit authority, and remote-default verification. Manual-policy PRs remain
  open.
- Every destructive action fails closed on post-plan drift and exact-SHA CAS
  mismatch. Repeated apply/verify is idempotent.
- Verify output exposes budget-deferred backlog, oldest age, net growth, and
  projected runs to convergence without changing the configured budget.
- Delivery completion, worktree cleanup, local/remote ref retirement, and host
  session-record retirement are separately observable and independently
  recoverable.
- The global baseline works in a fixture repository with no KB files or
  `cmd/kbcheck`; repo-native adapters add evidence without becoming mandatory.
- Two DAG nodes with disjoint files and resources may execute concurrently. Two
  nodes claiming the same protected semantic writer serialize globally; only
  the current fencing generation can commit a side effect.
- Conformance fixtures reject stale workers after takeover, reject expiry-only
  takeover, return the recorded result for identical idempotent replay, reject
  idempotency-key payload mismatch, and fail closed during coordinator or
  endpoint-validation outage.
- Security fixtures reject forged or wrong-audience authority, direct credential
  and alternate-workflow bypass, resource-key aliases/collisions, claim-store
  rollback, lower controller incarnations, validation-to-commit races, and an
  original timed-out request that commits after a status miss.

## Scope Boundaries

- Do not implement the reconciler in this brainstorm.
- Do not make deletion volume, disk-space reduction, or artifact count the
  primary success metric.
- Do not infer merge authority from repository ownership, write access, or an
  earlier request to implement code.
- Do not auto-close ordinary long-lived PRs because of age or inactivity.
- Do not require every consuming repository to vendor this bundle or build Go
  source locally.
- Do not introduce a required per-repository daemon, MCP server, or vector
  database. A versioned authoritative coordination adapter is required only for
  protected global semantic writers; ordinary local inventory and preservation
  remain available without it.
- Do not read or copy credential contents, private model state, or live runtime
  data into salvage PRs or decision packets.
- Do not delete host session records or remote refs through Git-only inference.
- Do not weaken the existing `terminal-cleanup` guard while introducing the
  global layer.

## Key Decisions

- **Reconcile, do not clean:** the primary product is an evidence-driven
  convergence controller. Deletion is one late action class.
- **Different actor for physical retirement:** the finishing session records
  intent and durable evidence; a later/different controller removes eligible
  work. This preserves current-session safety while still making the workflow
  self-cleaning.
- **Open PR is a durable waiting state:** it may release active execution
  capacity and a clean worktree, but it retains refs and a resume packet until
  integration or explicit supersession is proven.
- **Proof dominates confidence:** confidence thresholds cannot compensate for a
  missing destructive-action invariant.
- **Global core, optional repo depth:** global operation is the baseline;
  `kbcheck` and repo policy are richer adapters.
- **Risk budget caps action, not evidence:** budget exhaustion defers work;
  budget availability never authorizes a failed proof.
- **Human attention is batched:** unresolved irreversible choices become one
  bounded packet; unanswered items preserve data.
- **Parallel where proven, fenced where shared:** worktrees and DAG edges retain
  useful parallelism, while semantic writer keys serialize conflicting
  responsibility and endpoint fencing rejects stale authority.

## Dependencies / Assumptions

- **[safe-assumption]** A managed user-global binary can follow the existing
  `~/.kb/bin` installation pattern. Reversible because planning may choose a
  separate executable or extend a compatible installed binary. Distribution
  must bind the checksum manifest to a signed release or verifiable provenance
  anchored in a pinned trust root, enforce rollback protection, and verify
  destination permissions. Evidence/proof: installer tests must reject altered
  binaries, manifests, signatures/provenance, and downgraded releases while
  proving drift-safe upgrade and global execution in a non-KB fixture repo.
- **[safe-assumption]** Hosts and forges can expose versioned adapters for at
  least read-only session/PR inventory. Reversible because missing adapters
  downgrade only their owned mutations. Evidence/proof: capability negotiation
  fixtures must show unsupported actions preserved, not simulated.
- **[safe-assumption]** Suspended `awaiting-review` work can release active WIP
  while retaining durable resume state. Reversible because the remote topic,
  PR, local ref, receipt, and resume packet remain. Evidence/proof: a fixture
  resumes after worktree retirement and then reconciles after merge.
- **[safe-assumption]** Protected publishers/deployers can expose a versioned
  claim and fencing adapter even when its storage implementation differs.
  Reversible because the first release defines a provider-neutral contract and
  fails closed when no adapter exists. Evidence/proof: conformance fixtures must
  prove CAS acquisition, monotonic generations, stale-worker rejection,
  idempotent replay, bounded recovery, and outage preservation.

## Alternatives Considered

- **Delete everything older than a threshold:** rejected because age does not
  prove delivery, duplication, inactivity, or lack of unique data.
- **Prompt once per worktree/session:** rejected because it transfers inventory
  and Git classification work to the user and scales cognitive load with mess.
- **Run repo-local `kbcheck` everywhere:** rejected because global operation must
  work in ordinary repositories and installed skills are not standalone
  executables.
- **Have `kb-finalize` delete the current worktree after PR creation:** rejected
  because finalization neither owns delivery nor can safely remove the current
  executing worktree.
- **Keep every worktree until PR merge:** safe but unnecessarily expensive.
  Exact remote-topic containment plus a durable resume packet permits clean
  worktree retirement while preserving delivery refs.
- **Treat passing tests or PR state as sufficient containment proof:** rejected
  because tests do not establish remote identity and provider state may race.
- **Serialize the whole DAG:** rejected because dependencies and disjoint
  semantic resources permit useful parallel work.
- **Treat worktrees or path-disjoint edits as semantic locks:** rejected because
  separate paths and clones can still create competing publishers or mutate the
  same environment.
- **Use lease expiry without endpoint fencing:** rejected because a paused stale
  worker can resume after takeover.

## Slice Candidates

- Global evidence ledger and read-only portfolio inventory.
- Policy, confidence, protection, and risk-budget planner with decision packets.
- Exact Git deduplication, salvage, quarantine, and checked local retirement.
- Host/session and forge adapters for PR lifecycle and session-record retirement.
- KB lifecycle registration, WIP-state repair, and later-session/scheduled sweep.
- End-to-end mixed-portfolio fixtures, partial recovery, and no-KB-repo proof.

## Outstanding Questions

### Resolve Before Planning

None.

### Deferred to Planning

- **[defer-to-planning][Affects R27-R29][Technical]** Choose whether the global
  command is a dedicated executable or a subcommand of an existing installed
  binary after mapping release and adapter boundaries.
- **[defer-to-planning][Affects R13][Technical]** Define the concrete policy
  schema for per-action caps and default portfolio mutation budgets without
  weakening mandatory evidence predicates.
- **[defer-to-planning][Affects R16][Technical]** Choose conservative default
  salvage-PR and decision-packet caps and prove they remain configurable.
- **[defer-to-planning][Affects R20][Technical]** Select the durable decision
  packet format and host presentation surface.

### Parked / Out Of Scope

- **[parked][Affects R31b-R31j]** Automatically discovering that independently
  named resource keys represent the same semantic responsibility. The first
  version requires normalized policy-declared keys and may quarantine aliases;
  model similarity cannot merge authority domains.
- **[parked][Affects R29]** Automatic deletion of remote feature refs without a
  provider race-safe exact-ref API. Forbidden claim: plain Git branch deletion
  provides compare-and-swap safety.
- **[parked][Affects R4]** Salvaging credentials, private model state, or live
  runtime databases. Forbidden claim: secret scanning makes those files safe to
  publish.

## Evidence

- `.github/skills/kb-start/SKILL.md`
- `.github/skills/kb-work/SKILL.md`
- `.github/skills/kb-finalize/SKILL.md`
- `.github/skills/kb-complete/SKILL.md`
- `.github/skills/kb-ship/SKILL.md`
- `.github/skills/kb-land/SKILL.md`
- `.github/skills/kb-start/scripts/work_queue.ps1`
- `cmd/kbcheck/terminal_cleanup.go`
- `cmd/kbcheck/terminal_cleanup_test.go`
- `cmd/kbcheck/delivery_chain_contract_test.go`
- `docs/context/architecture/kb-workflow.md`
- `docs/context/architecture/kbcheck.md`
- `docs/context/research/2026-08-01-agent-dag-concurrency-and-fencing.md`
- `README.md`

## Next Steps

-> `kb-plan docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md`
