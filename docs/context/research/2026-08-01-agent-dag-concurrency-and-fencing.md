# Agent DAG Concurrency And Fencing

Checked: 2026-08-01
Budget mode: standard

## Question

How should an autonomous coding harness preserve useful DAG parallelism and
worktree isolation without allowing duplicate publishers, conflicting
production responsibilities, or stale actors to mutate external systems?

## Findings

### Worktrees isolate checkouts, not responsibility

Git linked worktrees have separate working directories, `HEAD`, and indexes,
but share the object database and most repository metadata, including ordinary
refs. `git worktree lock` protects worktree administration from pruning or
movement; it is not a semantic task lock.

Therefore worktrees prevent local checkout collisions. They do not coordinate
separate clones or machines, define the sole publisher, order deployments, or
fence registry and production side effects.

### A DAG needs resource constraints as well as dependency edges

Dependency edges answer "what must finish first." They do not answer "which
otherwise-independent nodes compete for the same responsibility." Established
schedulers add resource or concurrency constraints:

- GitHub Actions `needs` expresses prerequisites.
- Matrix `max-parallel` bounds concurrency.
- Concurrency groups serialize participating jobs in one repository.
- Merge queues retest integration against the latest base and preceding queued
  changes.

The harness should retain DAG parallelism for disjoint nodes while requiring
each node to declare semantic resources such as:

- `code:<module>`
- `publisher:<product>`
- `release-manifest:<product>`
- `deploy:<environment>`

Conflicting writers serialize even when their file paths do not overlap.

### Production exclusion requires a fenced mutation boundary

Kubernetes Leases and `resourceVersion` demonstrate shared claims acquired with
optimistic compare-and-swap. etcd transactions provide a linearizable
compare/then/else primitive. A lease alone is insufficient: an expired holder
can resume after a takeover.

For protected external mutations, the coordinator must issue a monotonically
increasing fencing generation. Every publish or deploy request carries that
generation, and the mutation endpoint atomically rejects stale generations.
The endpoint also enforces an idempotency key because retryable work may run
more than once.

This guarantee requires all production credentials and mutation paths to be
behind the fenced gateway. If agents retain bypass credentials, serialization
is policy rather than enforcement.

### Agent guidance supports narrower parallelism

Anthropic's multi-agent guidance reports that vague delegation causes duplicate
work and gaps, and that coding has fewer safely parallelizable tasks than
research. Its agent-team guidance uses dependencies, claiming, and file
locking, while recommending serialization for overlapping or
dependency-heavy work. GitHub Copilot's cloud agent similarly isolates each
task to one environment, branch, and pull request.

These mechanisms improve local coordination. None claims that workspace
isolation supplies global semantic or production ownership.

### Preventive controls

- Complete DAG dependencies plus semantic read/write resource declarations.
- Atomic responsibility claims with holder, source revision, expiry, and
  monotonically increasing generation.
- Endpoint-validated fencing for every external mutation.
- One credentialed publish/deploy gateway.
- Immutable artifacts followed by atomic promotion of one complete,
  digest-bound release manifest.
- Endpoint idempotency keys.
- Repository-wide deployment concurrency groups, protected environments, and
  merge queues as defense in depth.
- Fail closed when the coordinator or fence-validation store is unavailable.

Worktree scans, duplicate-code searches, drift reports, reconciliation, logs,
and alerts are detective or corrective. They cannot replace the preventive
boundary.

## Sources

- Git, [Git Worktree](https://git-scm.com/docs/git-worktree) and
  [Git Repository Layout](https://git-scm.com/docs/gitrepository-layout):
  per-worktree checkout state versus shared repository state.
- GitHub,
  [Using jobs](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-jobs):
  explicit DAG prerequisites.
- GitHub,
  [Control workflow concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency):
  repository-scoped concurrency groups.
- GitHub,
  [Managing a merge queue](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-a-merge-queue):
  integration validation, not external-side-effect serialization.
- GitHub,
  [Control deployments](https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/control-deployments):
  environments and concurrency are independent controls.
- Kubernetes, [Leases](https://kubernetes.io/docs/concepts/architecture/leases/)
  and [API concepts](https://kubernetes.io/docs/reference/using-api/api-concepts/#updates-to-existing-resources):
  coordination and `resourceVersion` compare-and-swap.
- etcd, [API transactions](https://etcd.io/docs/v3.6/learning/api/#transaction):
  linearizable conditional updates.
- Martin Kleppmann,
  [How to do distributed locking](https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html):
  why stale lease holders require fencing tokens.
- Kubernetes client-go,
  [Leader election](https://pkg.go.dev/k8s.io/client-go/tools/leaderelection):
  leader election does not itself guarantee fencing.
- Temporal,
  [Activity idempotency](https://docs.temporal.io/activity-definition#idempotency):
  retryable activities may execute more than once.
- Anthropic,
  [Building a multi-agent research system](https://www.anthropic.com/engineering/multi-agent-research-system)
  and [Agent teams](https://code.claude.com/docs/en/agent-teams):
  explicit ownership, dependency-aware claiming, and narrower coding
  parallelism.
- GitHub,
  [About Copilot cloud agent](https://docs.github.com/en/copilot/concepts/agents/cloud-agent/about-cloud-agent):
  one isolated environment, branch, and pull request per task.

## Applies When

Use this decision for parallel coding agents, release automation, publishers,
deployment controllers, migrations, shared mutable infrastructure, and any DAG
whose nodes can produce external side effects.

## Stale When

Recheck if Git changes linked-worktree sharing, the harness adopts an
authoritative remote coordinator, GitHub expands concurrency guarantees beyond
one repository and participating workflows, or the publish/deploy endpoint
gains a different atomic fencing primitive.

## Rejected Approaches

- **Serialize the entire DAG:** safe but discards independent parallelism.
- **Treat worktrees as locks:** they isolate files and indexes, not semantic or
  external responsibilities.
- **Use path overlap alone:** disjoint files can implement competing publishers
  or mutate the same environment.
- **Use lease expiry without fencing:** a paused stale holder can resume after
  takeover.
- **Use only GitHub concurrency groups:** callers outside the participating
  repository or workflow can bypass them.
- **Use only a merge queue:** it protects branch integration, not publication
  or deployment side effects.
- **Keep direct production credentials on agents:** this leaves the safety
  boundary bypassable.

## Impact On Current Project

Keep manifest DAGs and plan-run worktrees. Extend scheduling with declared
semantic resources and exclusive writers. Local leases remain useful for
same-repository coordination, but protected publication and deployment require
an optional authoritative coordinator adapter plus endpoint-validated fencing.
When that adapter is absent or unavailable, the reconciler must preserve work
and deny protected mutation rather than inventing authority.
