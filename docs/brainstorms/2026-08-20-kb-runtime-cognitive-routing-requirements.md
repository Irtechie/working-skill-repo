# KB Runtime, Cognitive Routing, and Surface Reduction Requirements

Status: ready for planning
Source: user direction, 2026-08-20

## Goal

Make the portable KB bundle cheaper and clearer to operate across hosts: use
deterministic state and tool contracts for inner-loop execution, route bounded
work to the lowest qualified live model, and preserve concise human-facing
artifacts at planning and delivery boundaries.

## User Requirements

1. Rename `kb-compact` to `kb-cognitive`. It owns low-cognitive-burden
   organization for brainstorming, planning, work reporting, and PR completion.
2. Keep `kb-simplify` distinct: it is explicit code-debt work, not an automatic
   drive-by-edit mechanism.
3. Require DDR delegation when a lower qualified live route is available. Plans
   express capabilities and tiers, never host-specific model names. The active
   host discovers actual routes at work time.
4. Make the execution inner loop state-machine/tool oriented. Phase transitions,
   heartbeats, proof reuse, routing receipts, and blockers must be structured
   data. Human Markdown is a boundary projection, not an inner-loop protocol.
5. Audit the remaining skill surface before retiring anything. `tdd` remains
   active unless normal planning/work proves parity for its explicit protected
   RED-to-GREEN contract. Retain `kb-first-principles` as the explicit,
   portable conversational trust boundary. Replace the large `kb-start` routing
   table only after a deterministic, fixture-tested classifier proves route
   parity.
6. Prefer serial execution in one branch/worktree. DAGs expose dependencies;
   parallel worktrees require explicit opt-in and proven isolation/benefit.
7. Retire AMR from active KB planning, routing, benchmark, and acceptance paths;
   DDR is the supported lower-tier execution model. Historical records may stay
   factual but must not be current routes.
8. Delivery is legible and role-specific: solo `p2d`/`w2d` continues to its
   permitted endpoint by opening and accepting its PR without a second merge
   question when protections permit; collaborative `kb-complete` opens a
   reviewable PR and stops at `awaiting-review` unless separately authorized to
   merge.
9. Limit each repository to one unsettled delivery candidate by default. A DAG
   orders slices; it does not authorize branch, worktree, or PR fan-out. In a
   solo P2D/W2D train, a later unit cannot begin while an earlier unit is merely
   `local-durable` or awaiting integration.

## Non-Goals

- Do not hard-code provider model lists, endpoints, credentials, or private
  routing state into plans or repository configuration.
- Do not merge planning, proof setup, research, memory repair, review, and
  delivery into one undifferentiated skill.
- Do not replace deterministic routing with an embedding dependency unless a
  later evidence-backed decision proves the rule classifier insufficient.
- Do not make `kb-simplify` automatic.

## Acceptance Criteria

1. Inner-loop transition and route/proof records validate as versioned,
   machine-readable artifacts; boundary Markdown is derived from those records.
2. A slice cannot mutate through normal KB execution without a delegated
   exact-tier-or-higher route receipt or a validated typed retain-current
   exception.
3. Live route selection remains host-local and records only a redacted runtime
   receipt in project artifacts.
4. `kb-cognitive` is the live named owner for cognitive organization, with
   callers and route fixtures updated; `kb-compact` is not a live route.
5. The explicit TDD request produces protected RED-to-GREEN behavior; `tdd` is
   retained unless an evidence-backed parity audit proves another owner enforces
   that exact contract.
6. `kb-first-principles` remains a distinct, portable conversational route and
   is not pulled into the execution inner loop.
7. The routing classifier gives deterministic, fixture-backed results and keeps
   the visible skill router short.
8. AMR has no active normal-work route; DDR is the only supported lower-tier
   execution contract.
9. A clean serial plan run requires no linked worktree; parallel execution is
   opt-in and documented by an isolation/benefit receipt.
10. Focused proof, `kbcheck core`, release/sync checks, and cross-install hashes
    pass before any delivery attempt.
11. Runtime state distinguishes implemented, locally durable, awaiting review,
    and delivery-integrated. Only the terminal delivery state is reported as
    done; a one-candidate gate prevents stranded completed branches.
12. Every reviewable PR presents a decision-first `Should we merge?` block from
    structured proof, changed scope, risk, and follow-ups. UI-visible changes
    include executed Playwright or host-browser evidence and selected rendered
    screenshots; non-UI changes use their executable proof without decorative
    images.
13. The harness is low-cognitive-burden by default: it emits a compact human
    projection with state, decision, proof, and next action, while complete
    JSON and raw logs remain available only through explicit detail/verbose
    outputs. Internal automation must not require humans to read JSON.
14. A P2D/W2D merge response reports PR URL, merged commit or exact unmet merge
    condition, proof summary, and selected Playwright/host-browser screenshots
    for UI-visible changes.

## Planning Decisions

- Use an epic with serial workstreams because runtime contracts, routing, and
  delivery touch shared policy and proof surfaces.
- Treat actual code/test/skill changes as future W2D work. This document and
  its manifests authorize planning only.
- Use standard Markdown as the portable reviewer format; HTML is optional only
  when a host proves its renderer supports it.
