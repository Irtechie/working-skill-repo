---
kb_id: kb-2026-07-30-current-agent-workflow-cleanup
slice_id: slice-002
title: "Make review and finalization proportional"
blockers: [slice-001]
verification: tdd
test_level: integration
functional_risk: broad
model_tier: large
model_tier_reason: "Core planning and completion boundaries change while proof and review coverage must remain intact."
model_requirements: ["workflow reasoning", "review prompt design", "receipt contracts", "integration fixtures"]
escalation_triggers: ["multiple reviewers can launch at one boundary", "unknown code risk can skip review", "stale proof can authorize completion"]
workspace_mode: shared-serial
conflict_domains: ["skill:kb-review", "skill:document-review", "skill:kb-finalize", "skill:kb-plan", "skill:kb-brainstorm"]
shared_resources: ["git:integration-owner", "config:skill-quality"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'Review|DocumentReview|Finalize' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-review/SKILL.md
    op: edit
    scope: "One profile per integrated review with four mandatory questions"
  - path: .github/skills/kb-review/references/review-process.md
    op: edit
    scope: "Conservative classification, profile selection, receipt, and single dispatch"
  - path: .github/skills/document-review/SKILL.md
    op: edit
    scope: "One best-fit reviewer only for explicit unresolved uncertainty"
  - path: .github/skills/kb-finalize/SKILL.md
    op: edit
    scope: "One review call, final proof invalidation, conditional learning and memory"
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Self-check by default; one optional reviewer for named uncertainty"
  - path: .github/skills/kb-brainstorm/SKILL.md
    op: edit
    scope: "Remove automatic materiality/persona review fan-out"
  - path: .github/skills/ce-review
    op: delete
    scope: "Remove duplicate review orchestrator and private references"
  - path: config/skill-quality.json
    op: edit
    scope: "Remove CE review forks and encode one review owner"
protected_oracles:
  - path: cmd/kbcheck/skill_guidance_test.go
    role: "review lifecycle and dispatch-count oracle"
    sha256: "db8b6ac30c163d724a0bed25919bcc807dddcdf5e66279b5d5b72ec9cc5e2a33"
    update_policy: "requires explicit plan update"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Extend failing fixtures, then simplify review and finalization."
human_action: ""
can_continue_other_slices: true
---

# Make Review and Finalization Proportional

Replace reviewer rosters and automatic learning ceremony with one bounded
semantic judgment at each justified plan boundary.

## Acceptance Criteria

- Per-slice execution runs coded proof but no reviewer.
- Planning self-checks by default; unresolved high-cost uncertainty may launch
  one best-fit document reviewer.
- Integrated review selects exactly one broad or specialist profile.
- Every profile covers intent, test validity, correctness, and code health.
- Unknown code-risk classification reviews rather than skips.
- Review receipts bind tree, requirements, proof, policy, and profile.
- Code-affecting fixes invalidate review and final exact-tree proof.
- Compound, learn, evolve, memory refresh, and compaction run only on their
  explicit signals.
- `ce-review` is deleted after standalone review parity moves to `kb-review`.

## Test Scenarios

- Routine code change dispatches one reviewer.
- Structural change dispatches one Thermonuclear profile, not broad plus
  Thermonuclear.
- Mechanical docs-only change with complete proof dispatches zero.
- Unknown classification dispatches broad review.
- Pre-slice uncertainty dispatches one reviewer; ordinary planning dispatches
  none.
- Changed requirements or proof invalidates a receipt.

## Scope Boundary

No claim that one reviewer has equal live defect yield to six without future
runtime evals.
