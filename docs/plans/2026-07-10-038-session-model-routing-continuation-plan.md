---
kb_id: kb-2026-07-10-session-model-routing
title: "Continue model routing with difficulty-sized workers"
continuation_of: docs/plans/2026-07-10-030-kb-session-model-routing-manifest.md
status: in_progress
created: 2026-07-10
current_gate: c5-advisory-pilot
protected_oracle:
  path: cmd/kbrouter/dispatch_test.go
  sha256: a8c38d5d13edb9c9ed39e2c3752157af68ca9eb8548d99d7bfce1cfedb4a2eba
---

# Session Model Routing Continuation

## Current State

- Slices 001-003 are complete and were independently reviewed P0/P1-clear.
- Slice 004 is incomplete. A GPT-5.5/high worker was stopped during its second
  remediation pass at the user's request so remaining work could be delegated
  by difficulty.
- Preserve the dirty worktree. Do not reset, revert, or recreate the current
  dispatch files. The interrupted edits are useful but unverified.
- C0 recovery is complete. The apparent Go hang was a cold sandboxed build plus
  an inaccessible user-level cache, not a persistent driver deadlock. With a
  warmed repo-local cache and parent workspace authority, the focused dispatch
  suite passed in 23.777s. Evidence is in
  `.kb/runs/session-model-routing/c0-recovery-report.md`.
- No background remediation or Go test process should be running. Recheck at
  recovery start.
- Do not mark slice 004 complete until both security and coherence reviewers
  report no P0/P1 findings.
- C1 is complete and P0/P1-clear. Windows runtime tests, native Linux tests,
  vet, Linux/Darwin cross-compiles, and the protected oracle hash passed.
  Evidence is in `.kb/runs/session-model-routing/c1-final-report.md`.
- C2 is complete and P0/P1-clear. Native Codex discovery, dispatcher-owned
  exact-receipt attestation, verification-only telemetry, canonical containment,
  Windows/Linux tests, vet, Darwin cross-compile, and live discovery passed.
  Evidence is in `.kb/runs/session-model-routing/c2-final-report.md`.
- C3 is complete. The Small `gpt-5.4-mini` closure updated the reviewed packet,
  plan, and manifest; the planner recorded private run state. Slice proof,
  contributor core (34 checks), protected hash, and snapshot replay (8/8)
  passed.
- C4 is complete. The Medium `gpt-5.4` worker implemented work-time discovery,
  tiered selection, overrides, provenance, completion, and portable Go sandbox
  policy. Focused/full skill eval, lint, packet validation, and diff checks pass.
- C4 review remediation used a launched Large `gpt-5.5` worker for mapping
  (thread `019f4dfb-c41c-7563-8c28-ac6cbd91f99e`) and the planner landed the
  bounded fixes: public Go selection, reusable run-root artifacts, honest
  override commands, and hardened sandbox paths.

## Work-Time Model Ladders

These are advisory choices, not frozen plan routes. Resolve actual availability
immediately before work and fall upward when an option is unavailable.

| Tier | Preferred options | Use |
|---|---|---|
| Small | Luna, Haiku, Qwen 3.6, Gemma 4, mini-class models | deterministic cleanup, fixtures, docs, narrow tests |
| Medium | Terra, Sonnet, DeepSeek 4, GPT-5.4 | bounded integration using already-proven security primitives |
| Large | GPT-5.5 high, Opus | auth, process execution, trust boundaries, live release evidence |
| Planner | current planner, Sol, Fable, Opus, GPT-5.5 high | dependency routing, review synthesis, gate decisions |

If a preferred model is unavailable, try another eligible same-tier model,
then move upward. Never move a security/auth/process-boundary slice downward.

## Policy Ownership And Instruction Scope

- `kb-plan` owns durable slice classification (`small`, `medium`, `large`) and
  risk constraints. Plans do not freeze a provider or model name.
- `kb-work` owns session discovery, work-time worker selection, same-tier then
  upward fallback, user `use`/`require`/`ignore` overrides, and receipts for
  what actually ran.
- `$CODEX_HOME/AGENTS.md` is only a global bootstrap and safety net for Codex
  work that bypasses KB. It should direct agents to the routing skills when
  available, forbid silent down-routing of risky work, allow user opt-out, and
  fall back safely when routing is unavailable. It must not duplicate ladders,
  endpoints, or the selection algorithm.
- Repo or nested `AGENTS.md` files own only project-specific constraints and
  override the global bootstrap where necessary.
- Private/local endpoints, startup commands, credentials, and extra model
  capabilities remain in the private user model registry, not any
  `AGENTS.md`. Codex global guidance is per OS account/`CODEX_HOME`; other
  accounts and non-Codex hosts require their own installed instruction layer.

## Remaining DAG

```text
C0 recovery (Small) [done]
  -> C1 dispatch trust kernel (Large) [done]
     -> C2 public dispatch wiring (Medium) [done]
        -> C3 slice-004 proof/docs (Small) [done]
           -> C4 kb-work routing integration (Medium) [done]
              -> C5 advisory pilot/release gate (Large)
                 -> C6 distribution/sync/PR (Large)
```

## C0 - Recover The Interrupted Checkpoint

**Tier:** Small. Escalate to Medium after 15 minutes without a concrete cause.

**Scope:** diagnostics and mechanical repair only. No trust-policy redesign.

1. Confirm the GPT-5.5 remediation process and orphaned `go test` processes are
   gone. Do not kill unrelated user processes.
2. Read the current scoped status and the protected oracle without changing it.
3. Explain why even `go version`/focused `go test` became slow or stuck. Check
   toolchain selection, cache locks, orphan processes, and repo-local
   `.gocache`/`.gotmp`. Remove only verified repo-local disposable caches.
4. Run `gofmt -d` on touched Go files, then compile and run the smallest focused
   test with a bounded outer timeout.
5. Record the exact current compile/test failure in the slice-004 result. Do not
   weaken tests or change the protected oracle hash without updating this plan.

**Proof:** `go test ./cmd/kbrouter -run 'Dispatch|Receipt|Fallback|ModelsAddProfile' -count=1 -timeout 45s`

## C1 - Finish The Dispatch Trust Kernel

**Status:** Complete.

**Tier:** Large only.

**Scope:** `internal/modelrouting/{catalog,dispatch,receipt}.go`,
`cmd/kbrouter/dispatch*.go`, and their security oracles.

Complete and verify the interrupted high-risk primitives:

- pin one canonical non-project Codex executable identity for the dispatch;
- strict packet/run/slice/packet-id binding and reserved/distinct attempt
  artifacts;
- single-open bounded child reads with identity checks;
- per-attempt catalog, state, approval, profile, auth, and destination recheck;
- immutable, strict user-local profile revision and endpoint/auth binding;
- post-run host-owned session attribution, never worker self-report;
- sample-free typed handoff passed to a real fallback without changing the
  original packet hash;
- no automatic downward fallback or less-trusted transition without approval;
- `ProofUnknown` until independent deterministic proof is attached;
- read-only/workspace-write authority only, explicit `codex-harness`, honest
  network enforcement, and fixed bounded `--output-schema`;
- Windows Job Object containment for the external Codex worker adapter. On
  Linux/macOS, discovery, cataloging, preview, and current-model fallback remain
  supported, but this adapter is not `dispatch-proven` and must return typed
  `dispatch-unavailable` before process start until a Unix containment adapter
  passes native security proof;
- no direct OpenAI-compatible agent loop.

Add/retain RED tests for every item. Run Windows runtime tests, native Linux
fail-closed and routing-fixture tests, plus Linux/Darwin cross-compiles. The
worker must not edit skills, release docs, lifecycle state, or git metadata.

**Proof:** protected dispatch oracle, package tests, vet, cross-compile, and two
read-only P0/P1 reviews.

## C2 - Wire Public Catalog And Dispatch Surfaces

**Status:** Complete.

**Tier:** Medium, after C1 exposes reviewed helpers. Escalate to Large if a
security primitive must change.

**Scope:** `cmd/kbrouter/main.go`, `cmd/kbrouter/catalog.go`, discovery/private
state plumbing, and telemetry receipt linkage.

- Finish `models add --profile` using the C1 validator; project scope rejects
  connection/profile data.
- Complete discover -> redacted catalog -> private per-project/run route-state
  writer -> dispatch flow with freshness and catalog binding.
- Rehydrate auth/profile fields only from opaque trusted user sources.
- Make native Codex routes dispatchable only through a versioned adapter prior
  or matching KB receipt; visible-only App models remain informative.
- Finish receipt-linked telemetry validation. Self-asserted `credited`, actual
  model/session/route, or proof-pass fields cannot satisfy a release gate.
- Add public CLI end-to-end fixtures without prompts or hand-edited models.json.

**Proof:** `go test ./cmd/kbrouter ./cmd/kbcheck -count=1` and `go vet`.

## C3 - Close Slice 004

**Status:** Complete.

**Tier:** Small.

**Scope:** deterministic fixtures, context packet, docs, and lifecycle evidence.

- Update the slice-004 packet to the reviewed dispatch envelope and
  `allowed_tools: [codex-harness]`.
- Remove obsolete fixture claims and document unavailable telemetry honestly.
- Recompute the protected oracle hash only after unchanged GREEN proof.
- Run `go run ./cmd/kbcheck core`, capture the slice result/history/snapshot,
  and replay all snapshots.
- Request final security/coherence read-only reviews. Any P0/P1 returns to C1
  or C2 according to ownership.

**Proof:** slice-004 plan proof, core, snapshot replay, no P0/P1.

## C4 - Integrate Work-Time Difficulty Routing

**Tier:** Medium (Terra-class is the intended first choice).

Execute existing slice 005:
`docs/plans/2026-07-10-035-kb-work-model-routing-plan.md`.

Make `kb-plan` write a difficulty tier and risk constraints for each slice while
keeping model names out of durable plans. `kb-work` discovers/selects
immediately before each bounded subagent dispatch, shows only non-empty route
groups, honors `use`/`require`/`ignore`, falls upward, and preserves ordinary
proof when the router or provenance is unavailable.

Ship and prove the portable Go sandbox bootstrap: Go cache/temp paths are
derived from the canonical project and slice, applied inside each Go command
rather than to the agent launcher, and remain ephemeral for downloaded installs.

**Proof:** skill lint plus the protected routing behavior eval.

## C5 - Prove The Advisory Pilot

**Tier:** Large.

Execute existing slice 006:
`docs/plans/2026-07-10-036-model-routing-pilot-plan.md`.

Require a live native Codex canary and one Codex-hosted OpenAI-compatible
profile canary, both linked to deterministic proof. GHCP, TinyBoss control, MCP,
and automatic-default claims remain unsupported until separately proven.

**Proof:** `go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json`

**Current blocker:** no attended, approved Codex-hosted OpenAI-compatible
profile canary was available in this session. Native Codex alone cannot satisfy
the two-route pilot contract. Do not fabricate or self-approve that evidence.

## C6 - Complete, Distribute, Sync, And Land

**Tier:** Large, coordinated by Planner.

1. Execute existing slice 007:
   `docs/plans/2026-07-10-037-model-routing-distribution-plan.md`.
2. Run kb-complete: structured review, resolution gate, proof/demo evidence,
   compound/learn/evolve, memory refresh, and cleanup.
3. Compare drift before syncing. Propagate approved skills to Codex, Copilot,
   shared agents, and ATV targets; verify hashes and every touched repo's
   `git diff --check`.
4. Install or document a short managed routing bootstrap for the user's global
   Codex `AGENTS.md` without overwriting unrelated guidance. Keep the complete
   policy in `kb-plan`/`kb-work`, and provide equivalent skill-based behavior
   for Copilot/shared-agent installations rather than assuming Codex's global
   file governs them.
5. Run `go run ./cmd/kbcheck local-release`.
6. Audit intentional files, commit on `codex/session-model-routing`, push, and
   open/update a PR. Do not merge without separate authority.

**Proof:** local-release, sync hashes, checked commit, pushed branch, PR URL.

## Stop Rules

- Do not redo correct work solely to improve model telemetry.
- Do not let a Small/Medium worker modify auth, endpoint, process containment,
  executable identity, profile parser, or trust-transition primitives.
- If a lower-tier worker encounters a security boundary, stop that slice and
  move only that bounded issue to Large.
- Preserve unrelated dirty-worktree changes and stage only audited feature
  files during C6.
