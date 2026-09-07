# Delivery authority and capability

Portable recovery grants no forge authority and performs no forge mutations.
`deliver-pr` is a pending request to `kb-complete`; authorized `merge-eligible`
attempts ask `kb-land` to obtain actual authenticated forge state. They remain
incomplete until its required checks pass. Imported JSON claiming
PR success, green checks or merged state is never treated as that observation.
A scoped cleanup acceptance cannot expand into another objective's delivery.

The existing native `evaluateRehabDelivery` in
`cmd/kbcheck/rehab_delivery.go` has no CLI entrypoint and remains test-only.
`internal/reconcile/policy.go` keeps native `ActionMerge.Allowed = false`
and its per-run budget at `0`. Do not claim that dormant evaluator ran, raise
its ceiling, or retry its refusal with the portable helper. Native commands,
when present, keep their policy predicates and ownership checks.

Existing `kb-complete`, `kb-ship`, and `kb-land` own their distinct delivery
route. They require current authority and actual exact-head proof. For useful
old work, construct/reuse its own owned delivery workspace and exact candidate;
never carry unrelated lineage into a new feature PR. Observe an existing PR
before creating another, and refresh ambiguous forge outcomes before retries.
A delegate call alone does not establish PR creation or integration.

Configured `merge: auto-after-checks` or a current explicit scoped merge
instruction supplies intent, not capability. Current permissions, remote base,
audited head, checks, reviews and protection must still permit integration.
Local-only policy selected before a resumed publishing/merge action withholds
that action. A revoked acceptance or pause cancels pending mutation while
preserving completed archives and work. Never infer merge permission from
private/solo ownership or write access.

Exact archived local-ref retirement is a different, installed capability: it
requires explicit item rejection, fresh authority, no live occupancy, verified
restore and compare-and-swap deletion. Its receipt cannot authorize publishing,
remote branch deletion, worktree removal or global synchronization.
