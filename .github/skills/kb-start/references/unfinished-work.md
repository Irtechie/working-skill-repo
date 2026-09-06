# Continue independently of unfinished work

Use the `recovery.ps1` shipped under the loaded `kb-rehab` skill directory.
Resolve the loaded skill path; do not assume a consumer `.github` directory or
require Go, Node, `cmd/kbcheck`, or `kbreconcile`.

1. Survey the source repository with `-Action survey -Root <source> -Json`.
   Local branch refs/tips include merge commits and unknown topology. A clean
   default checkout does not mean there is no backlog. Divergence is evidence
   for review, never sufficient evidence for deletion.
2. Under current mutating execution intent, write an explicit preparation
   request using the survey's repository/head/index/dirty/baseline identity,
   the run/objective, a non-default destination, selected document hashes and
   `dependency_status: independent`. The preparation contract lives with the
   installed helper. Unknown prerequisites remain scoped dependencies.
3. Call the same helper with `-Action continue -Root <source> -Request
   <continuation.json> -Json`. The request has `schema_version: 1`, `run_id`,
   `objective`, `preparation_request` (absolute path to that request), `manifest`
   (relative destination manifest), and `authority` containing
   `source: current-run`, boolean `continue: true`, and the same `objective`.
   Boolean `paused: true` stops continuation. Never turn a passed gate into
   current user authorization.
4. Display `notice.message` with its enumerated branch/path scope only when
   `notice.emit` is true. Use an asynchronous choice UI when available. Otherwise
   put the offer in a progress update and continue safe work in the same turn;
   do not wait for a reply. Read-only questions use only survey and session
   context, without creating requests or notice files.
5. `status: ready` returns an actual `workspace`, copied `manifest`, and
   `next_action: kb-work`. Revalidate existing manifest gates against the new
   inputs, then dispatch that owner. Never invent proof or a plan-to-work gate.
   `dependency-needed` identifies only the preparation that needs repair; keep
   other ready work moving. Advisory-state failure also leaves preparation
   runnable. Repeated entry reuses owned state instead of looping over offers.

An optional `reply` binds `offer_id` and selected `item_ids` to the inventory
shown to the user. Set `source: current-user-reply` only for an actual reply;
use `response: unanswered`, `decline`, `review`, or `cleanup`. Ambiguous
acceptance defaults to review/preservation. Cleanup requires explicitly
selected IDs. Changed tips/hashes and new items are excluded from late replies.
No response never becomes acceptance.

`cleanup.status: accepted-scope` is a **next action**, not cleanup completion:
invoke `kb-rehab` only for its returned items and mode. Revalidate that current
acceptance before later mutations; the notice record is not irrevocable
authority. Merge remains with existing delivery policy and real checks; private
or solo ownership alone cannot grant it. `merge_authorized` is always false in
this advisory receipt. A cleanup repair, protected source worktree, unavailable
native tool, or unrelated pending PR never blocks independent feature work.

Keep the same source root/run/objective on resume. Notice state is scoped under
the owning Git common directory; it is not global memory. The objective's own
new branch is excluded from notice fingerprints, so preparation itself cannot
repeat the backlog offer.
