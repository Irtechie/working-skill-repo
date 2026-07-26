---
kb_id: kb-2026-07-26-low-cognitive-burden-communication
slice_id: slice-002
title: "Apply low-burden structure to PR and companion review artifacts"
blockers: [slice-001]
verification: verification-only
test_level: integration
functional_risk: narrow
model_tier: medium
model_tier_reason: "Review artifacts must preserve high-leverage human decisions while avoiding duplicated tactical history."
model_requirements: ["HumanLayer prior-art synthesis", "PR workflow policy editing", "deterministic contract coverage", "cross-install hash verification"]
escalation_triggers: ["PR structure hides release blockers or proof", "companion document duplicates source-of-truth state", "installed copy contains newer useful drift"]
workspace_mode: shared-serial
conflict_domains: ["skill:kb-ship", "file:README.md", "file:cmd/kbcheck/communication_contract_test.go"]
shared_resources: ["git:integration-owner", "sync:global-skills"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run TestLowCognitiveBurdenCommunicationContract"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-ship/SKILL.md
    op: edit
    scope: "Make PR descriptions a low-burden first screen with explicit reviewer-attention and agent-handled sections."
  - path: docs/context/operations/low-burden-review-artifacts.md
    op: create
    scope: "Define the companion research/design/plan/proof document contract and its relationship to the PR body."
  - path: README.md
    op: edit
    scope: "Document low-burden PR and companion review behavior and credit HumanLayer prior art."
  - path: cmd/kbcheck/communication_contract_test.go
    op: edit
    scope: "Protect PR and companion review-artifact requirements."
  - path: docs/context/research/2026-07-26-low-cognitive-burden-agent-communication.md
    op: edit
    scope: "Record HumanLayer review-leverage findings and source links."
status: done
owner: agent
---

# Apply Low-Burden Structure To Review Artifacts

## Acceptance

- The PR body is a low-burden first screen, not a dump of the implementation
  transcript or tactical plan.
- The PR states what changed, why it matters, what genuinely needs reviewer
  attention, what the agent already handled, verification, risks, and deferred
  work.
- Every reviewer-attention item explains why the human should inspect or decide
  it; mechanical/proven work does not masquerade as a question.
- Detailed research, design, plan, and proof remain in source-owned companion
  documents linked from the PR.
- Companion documents identify unresolved human decisions and update their
  status when review resolves them.
- Proof, release blockers, risks, and safety context remain visible.

## Test Scenarios

1. Contract test fails when `kb-ship` omits reviewer-attention classification.
2. Contract test fails when the companion document omits the PR/companion
   source-of-truth boundary.
3. Focused test and `local-release` pass after required global skill copies are
   synchronized.

## Scope Boundary

No GitHub deployment, PR creation, comment-ingestion runtime, HumanLayer
installation, or replacement of KB's existing plan and manifest ownership.

## Result

Focused proof, core, local-release, and required global synchronization pass.
