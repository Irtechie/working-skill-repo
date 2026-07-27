---
kb_id: kb-2026-07-26-blocker-lifecycle-contract
slice_id: slice-032
title: "Apply pause, ownership, and propagation rules across KB skills"
blockers: [slice-031]
verification: integration
test_level: integration
functional_risk: narrow
model_tier: medium
model_tier_reason: "The same status semantics must survive planning, execution, durable goals, completion, and user-facing summaries."
model_requirements: ["workflow policy editing", "status transition reasoning", "contract-test coverage", "global skill drift reconciliation"]
escalation_triggers: ["a real safety approval is weakened", "pause mutates after a plain stop", "optional gates still roll up", "new global drift appears"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'TestBlockerLifecycle|TestLowCognitiveBurdenCommunicationContract' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-gate/SKILL.md
    op: edit
    scope: "Define ownership, scope, freshness, and propagation before a blocked or human-required claim."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Generate opt-in lifecycle metadata and prevent optional/release dependencies from contaminating implementation."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Use repairing and paused states, continue unrelated work, and recheck before reporting a blocker."
  - path: .github/skills/kb-goal/SKILL.md
    op: edit
    scope: "Keep explicit pause distinct from blocked durable-goal state."
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "Stop only the affected scope and keep release/optional gates from rolling up."
  - path: .github/skills/kb-brainstorm/SKILL.md
    op: edit
    scope: "Keep an explicit pause distinct from an unresolved planning question."
  - path: .github/skills/kb-handoff/SKILL.md
    op: edit
    scope: "Recheck and classify blockers before copying them into durable handoffs."
  - path: .github/skills/kb-qa/SKILL.md
    op: edit
    scope: "Keep missing agent-owned verification tooling in repair and reserve human-required for human-only access or judgment."
  - path: .github/skills/kb-finalize/SKILL.md
    op: edit
    scope: "Recheck terminal claims and prevent delivery-only gates from erasing implementation completion."
  - path: AGENTS.md
    op: edit
    scope: "Add the ambient blocker classification contract beside the question responsibility rules."
  - path: cmd/kbcheck/communication_contract_test.go
    op: edit
    scope: "Protect user-facing blocker language and pause semantics."
  - path: README.md
    op: edit
    scope: "Describe the visible blocker lifecycle behavior."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Record durable workflow status semantics."
status: done
owner: agent
---

# Apply Pause, Ownership, and Propagation Rules Across KB Skills

## Acceptance Criteria

- Plain `pause` stops mutation without converting the task or durable goal to
  blocked. `pause and handoff` may write only the requested handoff/state.
- Agent-owned test, code, controller, UI, and reproducibility failures stay
  repairing while meaningful safe work remains.
- Human-required is limited to authority, credentials/access, unavailable
  private input, irreversible risk, or subjective judgment.
- Every blocker is rechecked against its named sensor before it is repeated in
  a summary or handoff.
- A blocked slice stops only dependent slices. Unrelated ready work continues.
- Release, deployment, signing, optional provider, and optional platform gates
  block only their own promotion or capability scope.
- User-facing output says whether the user must act, what exact scope is
  affected, and what the agent can continue independently.
- Global skill copies are diff-reviewed before synchronization.

## Scope Boundary

Do not remove hard safety gates or claim unsupported platforms, visual quality,
credentials, deployments, or production changes are verified.
