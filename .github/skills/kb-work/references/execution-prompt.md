# KB Work Slice Execution Prompt

## Contents

- Prompt template
- Ownership receipt
- Build storage
- Scope and proof return

Use this as the per-slice execution checklist.

```text
You are executing a single vertical slice. Complete it fully.

KB: <kb_id>
Slice: <slice_id> - <title>
Verification mode: <tdd|integration|functional|verification-only>

Plan contents:
<full slice plan content>

Context packet:
<validated packet, or explicit small legacy/no-packet reason>

Route request:
<required tier and tier reason, execution_owner current|delegated, owner reason,
saved user-local project source priority, user overrides, and active host
surface>

Router receipt:
<current/no-router reason; selected native host target and provenance; or exact
kbrouter dispatch receipt with slice-unique artifact names>

Route announcement:
DDR route: <current|subagent> | primary: <current orchestrator|evidence-backed-route> | return: <none|parent-on-first-local-failure|required-alias-block> | tier: <small|medium|large> | proof: <short-proof-target>

Slice lease:
<exact slice-lease acquire command, owner token source, generation, renewal/release command, or non-mutating/no-lease reason>

Plan-run workspace:
<exact manifest-owned worktree, integration ref, run ID, owner token, and
expected integration head>

Build-storage contract:
<not-applicable, native execution-ready cargo-storage receipt path, or
portable-fallback record; plus the exact applied CARGO_TARGET_DIR>

Pre-slice review receipt:
<review ID, requirements source and SHA-256, durable review artifact path and
SHA-256, selected/completed/failed persona lifecycle, P0/P1 state, and each
structured residual constraint from the artifact; or the manifest's specific
not-required reason. Resolved constraints belong in the reviewed source and
the context packet.>

Execution policy:
<one owner decision, capability evidence, exact proof, and escalation triggers>

Instructions:
1. You are an implementation owner, not a document-review persona. Read the
   plan completely and implement only this bounded slice.
2. Work only in the exact manifest-owned worktree and integration ref above.
   Do not create or switch branches or create another worktree.
3. If the slice runs Go inside a workspace sandbox, load
   `references/go-sandbox.md` and apply its environment inside every Go shell
   invocation. Never put its temp overrides on the agent launcher.
4. If the slice runs Cargo, require a native receipt accepted by
   `cargo-storage --action validate-ready` or a `kb-check` portable-fallback
   record, then use its exact `CARGO_TARGET_DIR` for every build, test, repair,
   and probe. Do not replace it with a phase-, worker-, or run-specific target.
   Return a conflict if the contract is missing or invalid; do not invent
   `target-check`, `target-repair`, `target-repro`, or a probe target.
5. Use the packet's files and deterministic prefetch before broad search.
   Escalate when an escalation trigger fires or the packet is insufficient;
   do not silently expand authority.
6. Treat the route request, router receipt, execution owner, and route
   announcement as immutable orchestration receipts. The orchestrator already
   made the owner-first live decision from host/router evidence.
7. Do not re-decide ownership, discover or select a route, dispatch, or delegate
   again. If a receipt is missing or conflicts with the active execution
   surface, stop and return the conflict to the orchestrator.
8. The route announcement above was already emitted by the orchestrator before
   dispatch. Do not emit or repeat it, and never replace an evidence-backed
   route name from model memory.
9. Treat the pre-slice review receipt and its resolved constraints as immutable
   planning input; consume actionable constraints from the context packet. Do
   not rerun plan-wide specialist review or dispatch product,
   design, flow, security, scope, feasibility, coherence, or adversarial
   document reviewers for this slice. Return newly discovered plan-level risk
   as a replan envelope with requirements source and SHA-256, manifest path and
   run ID, affected slices and their invalidated status transitions, evidence,
   required source change, resume checkpoint, and `progress_key`.
10. Run the exact deterministic proof. Proof—not model self-review or a routing
   receipt—is the acceptance oracle. If proof fails, use ordinary bounded repair
   under the same owner. Re-plan, block, or record a new explicit ownership
   decision if the required authority changes.
11. For files marked `op: edit` in expected_files:
   - Read the current file content first.
   - Make only the change described in the `scope` field.
   - Preserve all existing behavior not mentioned in scope.
   - Current disk content is authoritative over stale plan text.
12. For files marked `op: create`, create the planned file.
13. Apply the verification mode:
   - tdd: failing test -> implementation -> passing test -> refactor.
   - integration: integration test proves the wired path.
   - functional: workflow/API/CLI/UI path is proven from public surface.
   - verification-only: build/check proves no regression.
14. Run relevant deterministic checks first, then broader checks when practical.
15. Before any observed write exceeds the plan-run claim, return it for
    coordinator expansion. A failed expansion requeues before the write.
16. Stage only files changed for this slice.
17. Commit in the current plan-run worktree only if the user authorized commits.
    Do not merge, reset, stash, clean, rebase, or amend another slice's commit.
18. Return the current commit SHA, observed writes, and a machine-readable proof
    receipt tied to the run, slice, and commit. Do not edit `todo.md`, manifests,
    handoffs, plan-run receipts, or other lifecycle state.
19. Renew the slice lease before handing back. The coordinator releases it only
    after it reruns slice and aggregate proof and atomically advances the
    expected plan-run integration head.

Do not modify other slices' files unless required for this slice.
Do not add scope beyond what the plan specifies.
Do not stage unrelated dirty or untracked files.
Do not route below the planned tier or silently change execution owners.
```
