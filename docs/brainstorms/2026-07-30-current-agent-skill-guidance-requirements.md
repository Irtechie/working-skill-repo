---
date: 2026-07-30
topic: current-agent-skill-guidance
brainstorm_style: kb-brainstorm
---

# Current Agent Skill Guidance Cleanup

## Problem Frame

The bundle has accumulated 46 skills, 51 custom agents, deprecated aliases,
oversized hot-path skill bodies, protected compatibility surfaces, and several
unconditional orchestration loops. Some contracts were reasonable for earlier
agent runtimes but conflict with current Anthropic and OpenAI guidance.

Research:
`docs/context/research/2026-07-30-current-agent-skill-guidance.md`.

## Requirements

**Audit and removal**

- G1. Audit every repo skill against single-agent-first routing, progressive
  disclosure, deterministic proof, bounded iteration, explicit delegation
  contracts, action-sensitive human approval, and curated memory. Store one
  machine-checkable row per observed skill with purpose, inbound callers,
  routing evidence, retained capability, disposition, and proof; reject missing
  or duplicate rows.
- G2. Remove skills that have no distinct behavior, operational caller, or
  intentionally retained user-facing contract. Git history is the recovery
  mechanism; dead aliases are not an archive. Before removal, record repository
  references, route fixtures, installer profiles, current docs, required global
  targets, direct invocation behavior, retained capabilities, and migration
  consequences. Unknown external consumers are not a compatibility promise.
- G3. Remove `ce-review`, `kb-finish`, and `klfg` only after a capability-parity
  matrix proves every still-required behavior has an owner and fixture:
  standalone code review moves to `kb-review`; deprecated completion aliases
  move to `kb-complete`. Then rewrite every caller, route fixture, installer
  profile, protection rule, documentation pointer, and required global target.
- G4. Keep useful explicit edge-case skills even when they are not automatic:
  `document-review`, `repo-critic`, `safe-shell-quoting`, and
  `kb-executive-brief` retain distinct user-facing jobs.

Current-operational reference roots are `.github/skills/`, `AGENTS.md`,
`README.md`, `config/`, active route/eval fixtures, and current architecture or
operations docs. Historical plans, completed work, research, decisions, and
solutions may retain factual references.

**Progressive disclosure**

- G5. Keep each `SKILL.md` below 500 lines. Hot-path skills over 450 lines fail
  unless reduced; move deterministic contracts, templates, and detailed
  mechanics into one-level-deep references. The skill body retains its trigger,
  route/ownership decisions, stop and safety rules, and explicit cues naming
  which reference to load for each phase. Line count is a guardrail, not proof
  of good progressive disclosure.
- G6. Reference files over 100 lines include a compact table of contents when
  navigation is not already obvious from their data format.

The configured hot path is `kb-start`, `kb-map`, `kb-brainstorm`, `kb-plan`,
`kb-work`, `kb-complete`, `kb-finalize`, `kb-review`, `kb-check`, `kb-goal`,
and `kb-qa`. Every extracted hot-path skill records a pre/post contract matrix
for triggers, ownership, inputs, outputs, stop/safety rules, and phase-specific
reference cues.

**Delegation, proof, and loops**

- G7. A skill may dispatch another agent only when it records why separate
  context, tools, permissions, ownership, or parallel independence is needed.
  Delegation prompts specify objective, output, sources/tools, and boundaries.
  Static checks prove only the presence and bounded shape of that contract;
  representative fixtures prove that the declared default route does not
  request fan-out; observed runtime routing remains outside static proof.
- G8. Long-running and repair skills define objective success, retry ceilings,
  stop conditions, and escalation behavior. They do not silently loop.
- G9. Non-trivial mutating skills name executable proof. Prompt-level confidence
  is never the terminal oracle.

**Memory, inventory, and deterministic policy**

- G10. Durable memory updates are conditional on a route-map change, reusable
  lesson, or explicit cadence. Ephemeral run notes remain under `.kb`.
- G11. Extend `kbcheck` with deterministic guidance checks for enforceable
  structure and fixtures. Do not pretend static checks can prove semantic
  quality or runtime value.
- G12. Update the skill architecture inventory from observed files and keep the
  count, groups, aliases, and code-review owner accurate.

**Synchronization**

- G13. Propagate every changed or removed skill consistently to required Codex,
  Copilot, and shared-agent global targets after the repo passes `core`. The
  configured sync-target inventory is authoritative. Propagation deletes named
  removed skills from required targets, verifies expected hashes and stale-file
  absence, then runs `local-release` as post-sync proof. A checked-in removed
  skill inventory makes deleted source folders visible to the drift checker and
  fails closed when required targets cannot be inspected.

## Success Criteria

- Every remaining skill has a distinct purpose or an operational caller.
- The audit matrix has exactly one validated row for every observed skill.
- Every removed skill has a completed evidence record and passing
  capability-parity matrix.
- No deprecated compatibility-only skill remains.
- No `SKILL.md` exceeds 500 lines.
- No configured hot-path `SKILL.md` exceeds 450 lines.
- Default workflow contracts and representative route fixtures contain no
  unjustified fan-out; observed live-model behavior remains a future eval.
- Unattended loop definitions expose bounded transitions, and representative
  fixtures reach objective proof or a typed blocker; universal runtime behavior
  is not claimed without live evaluation.
- `go run ./cmd/kbcheck core` and `go run ./cmd/kbcheck local-release` pass.
- Every required target has matching expected hashes and no stale folder for a
  skill listed in the removed-skill inventory.

## Scope Boundaries

- Do not remove optional specialist agents merely because static workflow
  references disappear; custom agents are a separate capability surface.
- Do remove agent files only when they are implementation-private to a deleted
  skill and have no remaining explicit or operational role.
- Do not add vendor-specific runtime dependencies.
- Do not perform paid live-model evals.
- The user explicitly accepts removal of deprecated aliases and the duplicate
  CE review entry point after the parity evidence above passes.

## Resolve Before Planning

None.

## Deferred to Planning

- Partition overlapping hot-path refactors so each manifest has one integration
  owner and deterministic proof.

## Parked / Out of Scope

- Live token-cost benchmarking remains parked until stable trace telemetry is
  available.
