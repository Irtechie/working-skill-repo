# Skill Bundle Cleanup Audit Requirements

Date: 2026-06-10
Source: chat-attached external audit summary
Route: kb-plan

## Intent

Turn the fresh-clone, minimalism, and documentation-layout audit into a
machine-verifiable cleanup plan for this portable skill bundle.

## Findings To Triage

1. Fresh-clone `go run ./cmd/kbcheck core` may fail because sync targets and
   ATV repo paths are absent in uninstalled contributor environments.
2. `go.mod` may require a newer Go toolchain than the code actually needs.
3. The documented core install profile may not include the dependency closure
   implied by installed skill instructions.
4. The agent surface may contain orphaned or near-duplicate agents that conflict
   with the repo's minimalism contract.
5. `ce-review` and `kb-review` may have forked shared doctrine instead of
   delegating or sharing a canonical reference set.
6. Some skills may reference infrastructure that the bundle does not ship, such
   as hook configuration paths.
7. Some successful negative selftests may print wording that looks like a real
   failure.
8. Repo memory and documented handoff layout should be checked for consistency
   before claiming the repo passes its own preflight.
9. Historical process documents may be too visible for clone-based consumers.
10. Cross-platform claims should match current script/runtime reality.

## Question Gate

Resolve Before Planning: empty.

Safe assumptions:

- Treat audit claims as unverified until slice 001 reproduces, confirms, or
  retires each one against the current repo.
- Do not delete agents, skills, or process history until a deterministic report
  proves the surface is unused or a documented retention decision exists.
- Do not sync global or ATV copies during this plan unless a later execution
  slice explicitly passes the canonical skill-repo gate first.

Deferred planning questions:

- Whether to execute the manifest immediately is deferred to the user after
  planning.

Parked:

- Live Codex/GHCP runs remain parked unless deterministic proof shows a live
  adapter is necessary.

## Acceptance Criteria

- The audit is converted into independently executable slices.
- Stale claims are retired with evidence instead of implemented blindly.
- Every implementation slice names expected files and agent-runnable proof.
- The manifest can be handed to `kb-work` without rediscovering the whole repo.

