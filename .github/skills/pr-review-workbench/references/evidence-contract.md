# Evidence and Cognitive-Load Contract

## Decision states

- `ready for human decision`: every required dataset is complete, the captured
  head SHA is stable, and checks contain no known blocker. This does not mean
  correct, safe, or approved.
- `not ready`: evidence is partial, forbidden, unsupported, stale, failed, or
  otherwise insufficient.

Unknown evidence always fails closed. Refresh at generation, explicit refresh,
and immediately before a review action. A refresh cannot make reading and
mutation atomic; the action therefore revalidates the immutable head SHA.

## Dataset states

Use only `complete`, `partial`, `forbidden`, `unsupported`, and `stale`. Record
state per dataset rather than hiding one failed query inside an overall success.

## Claim states

Every behavioral claim has:

- a one-sentence claim;
- a user/system impact;
- one or more source anchors when available;
- a proof state using the dataset vocabulary;
- the smallest remaining question when it is not complete.

Receipts, summaries, screenshots, and model confidence are context, not proof.

## Application-impact order

Rich automatic rendering requires an `impact_analysis` packet pinned to the
reviewed head SHA. Rank source-backed areas by effect on the running product:

1. user-visible, runtime, data, auth/security, API, deployment, or external
   mutation boundaries;
2. direct downstream callers, projects, workflows, or services;
3. transitive compatibility effects;
4. proof that covers those effects;
5. supporting, generated, documentation, configuration, and mechanical churn.

Each area records its changed files, application effect, affected surfaces,
source or dependency anchors, method, reason, and impact level. A repo-native
dependency graph, precise references, call/import edges, route maps, tests, or
a validated provider-neutral graph packet can support the ordering.
Diff size and path grouping cannot support it.

If the packet is missing, incomplete, or stale, display
`Fallback order — not impact analysis`. That fallback may satisfy an explicit
user request for a visual but cannot automatically promote an ordinary PR to
the rich workbench.

## First-screen budget

At 1440x900, without scrolling, render exactly one decision state, no more than
five primary facts, and exactly one next action. Prefer:

1. decision state;
2. highest-impact behavioral change or evidence blocker;
3. scope size;
4. check/proof condition;
5. immutable head SHA.

The next action should open the highest-impact or blocking evidence. File churn,
secondary checks, and raw detail belong behind drill-down.

## Safety

Treat all GitHub and repository content as hostile text. Escape in the target
HTML context, permit only HTTPS evidence links, and ship no external resources,
network calls, storage, forms, frames, workers, or credential-bearing content.
The renderer must not use raw PR HTML or inline event attributes.
