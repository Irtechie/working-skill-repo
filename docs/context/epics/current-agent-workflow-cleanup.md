# Current Agent Workflow Cleanup

Status: active
Created: 2026-07-30
Last refreshed: 2026-07-30

## Intent

Align the portable skill bundle with current Anthropic and OpenAI agent
guidance while preserving KB's useful planning, proof, recovery, and delivery
pattern.

## Success Criteria

- Review is proportional and single-agent-first.
- Dead compatibility skills are removed.
- Every skill is audited against the current guidance rubric.
- Hot-path skills use progressive disclosure and remain below 500 lines.
- Long-running loops have objective proof and bounded stop behavior.
- Required global skill targets match the reviewed repository source.

## Architecture Decisions

- Deterministic proof remains per slice; semantic review happens at plan
  boundaries.
- One reviewer profile runs per justified boundary.
- Optional edge-case skills remain discoverable but do not run automatically.
- Git history replaces dead aliases and duplicate orchestration engines.
- Repo source is canonical; global copies are release targets.

## Research

- `docs/context/research/2026-07-30-proportional-agent-code-review.md`
- `docs/context/research/2026-07-30-current-agent-skill-guidance.md`

## Workstreams

| Workstream | Brainstorm | Manifest | Status | Notes |
|---|---|---|---|---|
| Proportional review and finalization | `docs/brainstorms/2026-07-30-proportional-agent-review-requirements.md` | `docs/plans/2026-07-30-010-kb-current-agent-workflow-cleanup-manifest.md` | queued | Remove `ce-review`; one profile per boundary |
| Skill-surface cleanup and guidance guard | `docs/brainstorms/2026-07-30-current-agent-skill-guidance-requirements.md` | `docs/plans/2026-07-30-010-kb-current-agent-workflow-cleanup-manifest.md` | queued | Audit all skills; remove aliases; add deterministic structural guard |
| Progressive-disclosure refactor | same guidance source | `docs/plans/2026-07-30-010-kb-current-agent-workflow-cleanup-manifest.md` | queued | Reduce oversized hot-path skill bodies without weakening contracts |

## Dependency Map

1. Guidance guard and call inventory establish the enforceable baseline.
2. Review/finalization simplification updates the core lifecycle.
3. Progressive-disclosure refactors consume the settled lifecycle and guard.
4. One release/sync gate closes the epic.

## Execution Queue

Serial execution: these workstreams share skill routing, `kbcheck`, docs, and
global synchronization surfaces.

## Human Checkpoints

None. The user authorized removal of dead skills and continuous execution
through completion.

## Parked / Blocked

- Paid live-model cost comparison is parked.
- Custom-agent catalog deletion requires evidence beyond loss of one static
  caller and is not authorized by this skill-focused cleanup.

## Completion Criteria

- Every workstream is complete with deterministic proof.
- `core` passes before propagation.
- Required global targets are synchronized.
- `local-release` passes after propagation.
