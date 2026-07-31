---
type: kb-manifest
manifest_schema: 2
kb_id: kb-2026-07-30-current-agent-workflow-cleanup
brainstorm_path: docs/brainstorms/2026-07-30-current-agent-skill-guidance-requirements.md
created: 2026-07-30
status: completed
workflow_shape: pipeline-change
objective_contract: true
blocker_lifecycle_contract: true
pre_slice_review_contract: true
pre_slice_review:
  status: passed
  source: docs/brainstorms/2026-07-30-current-agent-skill-guidance-requirements.md
  source_sha256: 6f7b4b1d26beaa3e45416c1cf2956c6aa287825bafa945db11541fde2e99f146
  mode: requirements-wide
  review_id: current-agent-skill-guidance-requirements-2f4e5c2c7e69
  reviewed_at: "2026-07-31T03:08:00Z"
  review_artifact: docs/results/document-reviews/current-agent-skill-guidance-requirements-2f4e5c2c7e69.json
  review_artifact_sha256: 5fc910588767ea187c814dd05a0ea0410c85ba05e4220555dbbecaed5dd66c62
  persona_evidence_json: '{"adversarial-document-reviewer":"adversarial-risk: destructive cleanup, static-certification limits, progressive-disclosure regressions, and stale global copies could invalidate the initiative","coherence-reviewer":"consistency-risk: removal evidence, line limits, static proof boundaries, sync proof, and success criteria must agree","feasibility-reviewer":"delivery-risk: all observed skills, named deletions, hot-path extraction, and global absence proof need executable contracts"}'
  selected_personas_json: '["adversarial-document-reviewer","coherence-reviewer","feasibility-reviewer"]'
  completed_personas_json: '["adversarial-document-reviewer","coherence-reviewer","feasibility-reviewer"]'
  failed_personas_json: '[]'
  findings_resolved: 15
  unresolved_p0: 0
  unresolved_p1: 0
  residual_findings: 1
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "Proves repo contracts, diff hygiene, deterministic reports, and required global target drift after propagation."
model_tier_contract:
  allowed: [small, medium, large]
  default: medium
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  default_owner: current
  owner_choice: current-or-delegated
  max_owner_decisions_per_slice: 1
  catalog: active-host-plus-user-local
  delegated_fallback: same-tier-then-higher
  automatic_downward_routing: false
  automatic_cross_owner_fallback: false
  amr_required: false
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
  plan_run_worktree_default: true
  internal_integration_target: plan-run-branch
  default_branch_delivery_owner: kb-complete
  allowed_modes: [shared-serial]
scope-verified-files:
  - .github/skills/ce-compound-refresh/references/reporting-and-commit.md
  - .github/skills/ce-review
  - .github/skills/document-review
  - .github/skills/kb-brainstorm
  - .github/skills/kb-complete/SKILL.md
  - .github/skills/kb-finalize/SKILL.md
  - .github/skills/kb-finish
  - .github/skills/kb-gate/references/gate-ledger.md
  - .github/skills/kb-goal/SKILL.md
  - .github/skills/kb-map
  - .github/skills/kb-map-bootstrap
  - .github/skills/kb-plan
  - .github/skills/kb-review
  - .github/skills/kb-start/SKILL.md
  - .github/skills/kb-task/SKILL.md
  - .github/skills/kb-work
  - .github/skills/klfg
  - AGENTS.md
  - README.md
  - cmd/kbcheck
  - cmd/kbrouter/dispatch.go
  - config/removed-skills.json
  - config/skill-guidance-audit.json
  - config/skill-guidance.json
  - config/skill-quality.json
  - docs/brainstorms/2026-07-30-current-agent-skill-guidance-requirements.md
  - docs/brainstorms/2026-07-30-proportional-agent-review-requirements.md
  - docs/context/architecture
  - docs/context/epics/current-agent-workflow-cleanup.md
  - docs/context/goals/proportional-review-workflow.md
  - docs/context/research/2026-07-30-current-agent-skill-guidance.md
  - docs/context/research/2026-07-30-proportional-agent-code-review.md
  - docs/plans/2026-07-30-010-kb-current-agent-workflow-cleanup-manifest.md
  - docs/plans/2026-07-30-011-tool-skill-guidance-guard-plan.md
  - docs/plans/2026-07-30-012-skill-proportional-review-lifecycle-plan.md
  - docs/plans/2026-07-30-013-skill-dead-surface-cleanup-plan.md
  - docs/plans/2026-07-30-014-skill-progressive-disclosure-plan.md
  - docs/plans/2026-07-30-015-release-skill-propagation-plan.md
  - docs/results/document-reviews
  - docs/results/proof/current-agent-workflow-cleanup-slice-001.txt
  - docs/results/proof/current-agent-workflow-cleanup-slice-002.txt
  - docs/results/proof/current-agent-workflow-cleanup-slice-003.txt
  - docs/results/proof/current-agent-workflow-cleanup-slice-004.txt
  - docs/results/proof/current-agent-workflow-cleanup-slice-005.txt
  - evals/cross-model-benchmarks/route-selection.json
  - evals/route-complexity
  - todo.md
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-epic
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Current Anthropic and OpenAI guidance is captured from primary sources."
      - "Review and bundle-wide requirements have zero unresolved P0/P1 findings."
      - "The user authorized removal of dead skills and continuous execution."
      - "No ask-now or research-first decisions remain."
    proof:
      - docs/context/research/2026-07-30-current-agent-skill-guidance.md
      - docs/context/research/2026-07-30-proportional-agent-code-review.md
      - docs/brainstorms/2026-07-30-proportional-agent-review-requirements.md
      - docs/brainstorms/2026-07-30-current-agent-skill-guidance-requirements.md
      - docs/results/document-reviews/proportional-agent-review-requirements-08166ece884c.json
      - docs/results/document-reviews/current-agent-skill-guidance-requirements-2f4e5c2c7e69.json
    blockers: []
    passed_at: "2026-07-31T03:12:00Z"
    allowed_next_action: "kb-plan docs/context/epics/current-agent-workflow-cleanup.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The manifest and five slice plans exist."
      - "The dependency DAG is serial and acyclic."
      - "Every slice defines acceptance, expected files, verification, risk, model tier, proof, and HITL."
      - "Removal parity and required global propagation are explicit."
    proof:
      - docs/plans/2026-07-30-010-kb-current-agent-workflow-cleanup-manifest.md
      - docs/plans/2026-07-30-011-tool-skill-guidance-guard-plan.md
      - docs/plans/2026-07-30-012-skill-proportional-review-lifecycle-plan.md
      - docs/plans/2026-07-30-013-skill-dead-surface-cleanup-plan.md
      - docs/plans/2026-07-30-014-skill-progressive-disclosure-plan.md
      - docs/plans/2026-07-30-015-release-skill-propagation-plan.md
    blockers: []
    passed_at: "2026-07-31T03:12:00Z"
    allowed_next_action: "kb-work docs/plans/2026-07-30-010-kb-current-agent-workflow-cleanup-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Skill guidance policy tests pass."
      - "The guard distinguishes structural enforcement from semantic judgment."
      - "The observed 46-skill audit has one row per skill."
    proof:
      - cmd/kbcheck/skill_guidance.go
      - cmd/kbcheck/skill_guidance_test.go
      - config/skill-guidance.json
      - config/skill-guidance-audit.json
      - docs/results/proof/current-agent-workflow-cleanup-slice-001.txt
    blockers: []
    passed_at: "2026-07-31T03:22:00Z"
    allowed_next_action: "slice-002"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Planning defaults to self-check and zero reviewers."
      - "Optional document review and integrated code review dispatch at most one profile."
      - "Review receipts and final proof invalidate on code-affecting fixes."
    proof:
      - docs/results/proof/current-agent-workflow-cleanup-slice-002.txt
      - .github/skills/kb-review/SKILL.md
      - .github/skills/document-review/SKILL.md
      - .github/skills/kb-finalize/SKILL.md
    blockers: []
    passed_at: "2026-07-31T04:05:00Z"
    allowed_next_action: "slice-003"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Removed skills have replacement owners and no unique retained behavior."
      - "Current routes, configs, tests, README, and architecture docs do not advertise removed names."
      - "Historical artifacts remain intact."
    proof:
      - docs/results/proof/current-agent-workflow-cleanup-slice-003.txt
      - docs/context/architecture/skills.md
      - config/removed-skills.json
    blockers: []
    passed_at: "2026-07-31T04:24:00Z"
    allowed_next_action: "slice-004"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Every skill body is below the enforced hard limit and configured hot paths are below their lower limit."
      - "Moved contracts remain reachable through one-level references and navigation."
      - "Planner difficulty classification and route evals agree on small, medium, and large."
      - "The full contributor core gate passes on the exact tree."
    proof:
      - docs/results/proof/current-agent-workflow-cleanup-slice-004.txt
      - .github/skills/kb-brainstorm/SKILL.md
      - .github/skills/kb-plan/SKILL.md
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-finalize/SKILL.md
    blockers: []
    passed_at: "2026-07-31T05:30:00Z"
    allowed_next_action: "slice-005"
  - gate_id: slice-slice-005-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Every required global target was reviewed before overwrite."
      - "All required skill hashes match the repository source."
      - "Retired skill folders are absent from all required targets."
      - "The local release gate passes."
    proof:
      - docs/results/proof/current-agent-workflow-cleanup-slice-005.txt
      - README.md
      - config/removed-skills.json
      - cmd/kbcheck/model_routing_release.go
    blockers: []
    passed_at: "2026-07-31T06:15:00Z"
    allowed_next_action: "work-to-complete"
  - gate_id: work-to-complete
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Every non-skipped slice has a passing slice-to-done gate."
      - "The final verification command and result are recorded."
      - "No unresolved P0/P1 exists."
      - "Board and manifest are synchronized."
      - "scope-verified-files is populated."
    proof:
      - docs/plans/2026-07-30-010-kb-current-agent-workflow-cleanup-manifest.md
      - docs/results/proof/current-agent-workflow-cleanup-slice-001.txt
      - docs/results/proof/current-agent-workflow-cleanup-slice-002.txt
      - docs/results/proof/current-agent-workflow-cleanup-slice-003.txt
      - docs/results/proof/current-agent-workflow-cleanup-slice-004.txt
      - docs/results/proof/current-agent-workflow-cleanup-slice-005.txt
    proof_commands:
      - "go run ./cmd/kbcheck core"
      - "go run ./cmd/kbcheck local-release"
    blockers: []
    passed_at: "2026-07-31T06:15:00Z"
    allowed_next_action: "kb-finalize docs/plans/2026-07-30-010-kb-current-agent-workflow-cleanup-manifest.md"
slices:
  - id: slice-001
    title: "Enforce current skill-guidance structure"
    path: docs/plans/2026-07-30-011-tool-skill-guidance-guard-plan.md
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "The guard changes release-blocking policy across every skill and must distinguish enforceable structure from semantic judgment."
    model_requirements: ["Go policy testing", "skill-contract parsing", "false-positive control", "cross-platform CLI proof"]
    escalation_triggers: ["the guard claims semantic quality", "existing valid skills cannot express an exception", "core becomes dependent on live model output"]
    workspace_mode: shared-serial
    conflict_domains: ["path:cmd/kbcheck", "file:config/skill-quality.json"]
    shared_resources: ["git:integration-owner", "release:core"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'SkillGuidance|Minimality|ReviewReference' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Write failing policy fixtures, then implement the smallest deterministic guidance guard."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=5 changed=7 discovered=2 unexplained=0; test-level: functional-cli; proof: go test ./cmd/kbcheck -run 'SkillGuidance|Minimality|ReviewReference' -count=1 exit=0; memory-impact: durable; areas=skill-quality; refresh=pending"
  - id: slice-002
    title: "Make review and finalization proportional"
    path: docs/plans/2026-07-30-012-skill-proportional-review-lifecycle-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: integration
    functional_risk: broad
    model_tier: large
    model_tier_reason: "The slice rewrites planning and completion boundaries while preserving proof, review coverage, and exact-tree invalidation."
    model_requirements: ["workflow call-graph reasoning", "review prompt design", "receipt contracts", "fixture-driven routing"]
    escalation_triggers: ["more than one reviewer can run at a boundary", "review can be skipped for unknown code risk", "final proof can remain stale after fixes"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:kb-review", "skill:document-review", "skill:kb-finalize", "skill:kb-plan", "skill:kb-brainstorm"]
    shared_resources: ["git:integration-owner", "config:skill-quality"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'Review|DocumentReview|Finalize' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Protect boundary and reviewer-count behavior, then simplify orchestration and remove ce-review."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-discovery: cmd/kbcheck/manifest_contract.go and tests - schema 3 enforces one reviewer while schema 2 receipts remain readable; scope-discovery: cmd/kbcheck/review_reference_guard.go and tests - duplicate CE owner removed; scope-discovery: AGENTS.md and config/skill-guidance-audit.json - protection and audit updated after ce-review removal; scope-check: forecast=8 changed=23 discovered=15 unexplained=0; test-level: integration; proof: go test ./cmd/kbcheck -run 'Review|DocumentReview|Finalize' -count=1 exit=0; memory-impact: durable; areas=review-lifecycle,finalization; refresh=pending"
  - id: slice-003
    title: "Remove dead completion aliases and stale surfaces"
    path: docs/plans/2026-07-30-013-skill-dead-surface-cleanup-plan.md
    blockers: [slice-002]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "Deletion is mechanically bounded but spans routing fixtures, installer profiles, docs, and required target inventories."
    model_requirements: ["call inventory", "capability-parity fixtures", "installer/profile updates", "stale-reference detection"]
    escalation_triggers: ["a named alias has distinct behavior not owned by kb-complete", "a current operational caller has no migration", "required install profiles cannot remove stale folders"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:kb-finish", "skill:klfg", "docs:skill-inventory", "config:routes"]
    shared_resources: ["git:integration-owner", "config:skill-quality"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'WorkflowGovernor|SkillLint|Minimality' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Record parity, delete aliases, and remove every current operational reference while preserving historical artifacts."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-discovery: cmd/kbcheck/workflow_governor.go and skill_validators.go - alias-specific contracts removed; scope-discovery: docs/context/architecture/README.md and kb-workflow.md - current route map updated; scope-check: forecast=9 changed=16 discovered=7 unexplained=0; test-level: functional-cli; proof: targeted Go tests exit=0, skill-lint errors=0, operational git grep removed names exit=1/no matches; proof amendment: replaced dependency-inverted core check with route-local tests, while core remains the terminal aggregate; memory-impact: durable; areas=routing,skill-inventory; refresh=done"
  - id: slice-004
    title: "Refactor oversized hot-path skills"
    path: docs/plans/2026-07-30-014-skill-progressive-disclosure-plan.md
    blockers: [slice-003]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Moving contracts out of hot-path prompts can silently break discovery, safety, and phase ownership unless reference reachability is proven."
    model_requirements: ["progressive-disclosure design", "large Markdown refactoring", "contract-preservation fixtures", "skill lint interpretation"]
    escalation_triggers: ["a moved safety rule lacks an explicit load cue", "a skill remains above 500 lines", "route fixtures change behavior unintentionally"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:kb-work", "skill:kb-plan", "skill:kb-brainstorm", "skill:kb-finalize"]
    shared_resources: ["git:integration-owner", "release:core"]
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck core"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Move deterministic mechanics into one-level references, preserve hot-path decisions, and validate all skill bodies."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-discovery: README.md - visible lifecycle, tier classifier, and worktree naming contract; scope-discovery: cmd/kbcheck/checks.go and cmd/kbrouter/dispatch.go - Windows nested process proof made reproducible; scope-discovery: config/skill-quality.json and evals route fixtures - legacy standard tier migrated to medium; scope-check: forecast=8 changed=34 discovered=26 unexplained=0; test-level: functional-cli; proof: go run ./cmd/kbcheck core exit=0 checks=39; protected-oracle: cmd/kbcheck/skill_guidance_test.go sha256=db8b6ac30c163d724a0bed25919bcc807dddcdf5e66279b5d5b72ec9cc5e2a33 unchanged; memory-impact: durable; areas=skill-loading,model-tier-routing,worktree-naming,windows-proof; refresh=done"
  - id: slice-005
    title: "Propagate and prove the cleaned bundle"
    path: docs/plans/2026-07-30-015-release-skill-propagation-plan.md
    blockers: [slice-004]
    verification: verification-only
    test_level: full
    functional_risk: broad
    model_tier: medium
    model_tier_reason: "The final slice applies an established release procedure but must prove changed hashes and stale-folder deletion across three required targets."
    model_requirements: ["release gate operation", "safe exact-path synchronization", "hash verification", "drift diagnosis"]
    escalation_triggers: ["a global target contains newer useful work", "sync requires broad recursive deletion", "local-release reports unexplained drift"]
    workspace_mode: shared-serial
    conflict_domains: ["sync:global-skills", "docs:release-state"]
    shared_resources: ["git:integration-owner", "sync:global-skills", "release:local-release"]
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck local-release"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Run core, compare required targets, propagate exact skill folders and deletions, then run local-release."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-discovery: cmd/kbcheck/model_routing_release.go - release proof must not nest kbrouter process-containment tests inside another Windows job; scope-check: forecast=5 changed=4 discovered=1 unexplained=0; test-level: full; proof: skill-sync-report comparisons=129 issues=0, removed global folders absent, go run ./cmd/kbcheck local-release exit=0; memory-impact: operational; refresh=done"
---

# KB: Current Agent Workflow Cleanup

## Origin

Epic: `docs/context/epics/current-agent-workflow-cleanup.md`

Requirements:

- `docs/brainstorms/2026-07-30-proportional-agent-review-requirements.md`
- `docs/brainstorms/2026-07-30-current-agent-skill-guidance-requirements.md`

## Workflow Shape

`pipeline-change` - skills, callers, deterministic guards, documentation,
installer/sync surfaces, and release proof change together.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Enforce current skill-guidance structure | - | tdd | no | done |
| 2 | Make review and finalization proportional | slice-001 | tdd | no | done |
| 3 | Remove dead completion aliases and stale surfaces | slice-002 | integration | no | done |
| 4 | Refactor oversized hot-path skills | slice-003 | integration | no | done |
| 5 | Propagate and prove the cleaned bundle | slice-004 | verification-only | no | done |
