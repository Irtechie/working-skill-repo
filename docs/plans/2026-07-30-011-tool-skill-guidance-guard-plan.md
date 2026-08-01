---
kb_id: kb-2026-07-30-current-agent-workflow-cleanup
slice_id: slice-001
title: "Enforce current skill-guidance structure"
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Release-blocking policy across every skill requires precise structural checks and honest limits."
model_requirements: ["Go policy testing", "skill parsing", "fixture design", "false-positive control"]
escalation_triggers: ["semantic quality is inferred from syntax", "valid exceptions cannot be scoped", "live models become required"]
workspace_mode: shared-serial
conflict_domains: ["path:cmd/kbcheck", "file:config/skill-quality.json"]
shared_resources: ["git:integration-owner", "release:core"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'SkillGuidance|Minimality|ReviewReference' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/skill_guidance.go
    op: create
    scope: "Deterministic enforceable guidance checks and concise report"
  - path: cmd/kbcheck/skill_guidance_test.go
    op: create
    scope: "Passing and failing fixtures for size, references, delegation, loops, and aliases"
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "Register the guidance check in contributor proof"
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Expose the focused guidance command when needed"
  - path: config/skill-quality.json
    op: edit
    scope: "Declare enforceable thresholds and retained explicit lanes"
protected_oracles:
  - path: cmd/kbcheck/skill_guidance_test.go
    role: "skill guidance policy oracle"
    sha256: "db8b6ac30c163d724a0bed25919bcc807dddcdf5e66279b5d5b72ec9cc5e2a33"
    update_policy: "requires explicit plan update"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Write failing fixtures before implementing the guard."
human_action: ""
can_continue_other_slices: true
---

# Enforce Current Skill-Guidance Structure

Build a deterministic guard for the parts of current guidance that code can
honestly prove.

## Acceptance Criteria

- Fails skill bodies over 500 lines and hot-path bodies over the configured
  threshold without broad permanent exemptions.
- Validates one-level reference reachability and navigation cues.
- Detects prohibited deprecated aliases and unbounded orchestration markers.
- Delegation checks validate contract shape only and do not claim necessity.
- Representative fixtures prove default review/finalization routes do not fan
  out.
- Existing core proof invokes the guard.

## Test Scenarios

- Oversized skill fails; compact skill with reachable references passes.
- Nested or missing reference fails.
- Boilerplate delegation rationale cannot satisfy required scenario evidence.
- Deprecated alias appears and fails.
- A bounded retry loop passes; an unbounded loop fixture fails.

## Scope Boundary

No live model grading and no semantic claim that a skill is useful or that a
delegation is necessary.
