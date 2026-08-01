---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-08-01-global-cleanup-reconciliation
brainstorm_path: docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md
created: 2026-08-01
status: active
workflow_shape: pipeline-change
objective_contract: true
blocker_lifecycle_contract: true
pre_slice_review_contract: true
model_tier_contract: true
workspace_isolation_contract: true
proof_governor_contract: true
source_requirements_sha256: 11545592e2fe3f184babfc2f9e7b2f9e175fa8b9f6a88a6c7174cf8bf8676d98
pre_slice_review:
  status: passed
  source: docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md
  source_sha256: 11545592e2fe3f184babfc2f9e7b2f9e175fa8b9f6a88a6c7174cf8bf8676d98
  mode: requirements-wide
  review_id: global-cleanup-reconciliation-requirements-11545592e2fe
  reviewed_at: "2026-08-01T16:21:57Z"
  review_artifact: docs/results/document-reviews/global-cleanup-reconciliation-requirements-11545592e2fe.json
  review_artifact_sha256: 6bda7d5b6143be0805ad542e4922fe4c837b542d8ed31ef7d275fdf4090ec717
  persona_evidence_json: '{"adversarial-document-reviewer":"adversarial-risk: cross-repository merge, salvage, PR supersession, worktree/ref retirement, and host session-record retirement require stress-testing evidence, race, authority, and recovery boundaries."}'
  selected_personas_json: '["adversarial-document-reviewer"]'
  completed_personas_json: '["adversarial-document-reviewer"]'
  failed_personas_json: '[]'
  findings_resolved: 11
  unresolved_p0: 0
  unresolved_p1: 0
  residual_findings: 0
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck core"
  expect: 0
  why: "Proves the global reconciler, existing terminal-cleanup safety, installer behavior, workflow contracts, and documentation gates together."
model_tier_contract:
  allowed: [large]
  default: large
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  default_owner: delegated
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
plan_run_worktree:
  base_sha: 8bf72ab96ec2e925604169882bd32a6271924543
  branch: codex/kb-2026-08-01-global-cleanup-reconciliation
  workspace_mode: shared-serial
  commit_authorized: true
  commit_authorized_by: user
  commit_approval_ref: "2026-08-01 explicit kb-complete request"
delivery_authority:
  source: project-policy-absent
  mode: local
  merge: manual
  post_merge_sync: false
  authorized_actions: [create-plan-worktree, local-commit]
  forbidden_actions: [push-topic, create-pr, merge, push-default, remote-ref-delete, host-session-delete]
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The reviewed requirements define fail-closed inventory, exact containment, policy thresholds, risk budgets, compact decision packets, salvage, and separate delivery/cleanup/ref/session gates."
      - "The requirements-wide adversarial review has no unresolved P0/P1 findings."
      - "The global baseline remains useful without repo-native kbcheck or host/forge adapters."
    proof:
      - docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md
      - docs/results/document-reviews/global-cleanup-reconciliation-requirements-11545592e2fe.json
      - cmd/kbcheck/terminal_cleanup.go
    blockers: []
    passed_at: "2026-08-01T16:27:00Z"
    allowed_next_action: "kb-plan docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Three vertical slices cover read-only portfolio convergence, checked apply/verify, and lifecycle/distribution integration."
      - "Every requirement maps to a slice and every destructive path retains existing terminal-cleanup predicates or fails closed."
      - "The dependency graph is acyclic and serializes shared Git, installer, skill, and documentation surfaces."
      - "Local plan-run commits are explicitly authorized; publishing remains unauthorized."
    proof:
      - docs/plans/2026-08-01-001-global-reconciler-inventory-plan.md
      - docs/plans/2026-08-01-002-global-reconciler-apply-plan.md
      - docs/plans/2026-08-01-003-global-reconciler-lifecycle-plan.md
      - docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md
    blockers: []
    passed_at: "2026-08-01T16:27:00Z"
    allowed_next_action: "kb-work docs/plans/2026-08-01-000-kb-global-cleanup-reconciliation-manifest.md"
slices:
  - id: slice-001
    title: "Inventory and plan a portfolio without deleting first"
    path: docs/plans/2026-08-01-001-global-reconciler-inventory-plan.md
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    execution_class: cli
    model_tier: large
    model_tier_reason: "Defines cross-repository evidence, confidence, risk, and exception-packet contracts consumed by later mutation."
    model_requirements: ["cross-repository Git reasoning", "fail-closed policy design", "functional CLI fixtures"]
    escalation_triggers: ["missing authoritative evidence", "confidence overriding a failed predicate", "unbounded decision packets"]
    workspace_mode: shared-serial
    conflict_domains: ["go:reconcile-core", "cli:kbreconcile", "config:reconcile-contract"]
    proof_check:
      kind: command_exit
      command: "go test ./internal/reconcile ./cmd/kbreconcile -run 'Inventory|Plan|DecisionPacket|NoKBRepo' -count=1"
      expect: 0
    status: pending
    owner: agent
    can_continue_other_slices: false
  - id: slice-002
    title: "Apply and verify only unchanged high-confidence actions"
    path: docs/plans/2026-08-01-002-global-reconciler-apply-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: functional-cli
    functional_risk: destructive
    execution_class: cli
    model_tier: large
    model_tier_reason: "Mutates Git worktrees, exact refs, and queue metadata under concurrent races and partial failures."
    model_requirements: ["Git worktree internals", "compare-and-swap refs", "remote containment", "cross-platform locks"]
    escalation_triggers: ["force would be required", "remote authority is unresolved", "plan identity drift", "terminal-cleanup parity gap"]
    workspace_mode: shared-serial
    conflict_domains: ["go:reconcile-core", "go:terminal-cleanup", "git:worktrees", "git:refs", "state:work-queue"]
    proof_check:
      kind: command_exit
      command: "go test ./internal/reconcile ./cmd/kbreconcile ./cmd/kbcheck -run 'Apply|Verify|Reconcile|TerminalCleanup' -count=1"
      expect: 0
    status: pending
    owner: agent
    can_continue_other_slices: false
  - id: slice-003
    title: "Make KB lifecycle and global install converge automatically"
    path: docs/plans/2026-08-01-003-global-reconciler-lifecycle-plan.md
    blockers: [slice-001, slice-002]
    verification: tdd
    test_level: integration
    functional_risk: broad
    execution_class: cli
    model_tier: large
    model_tier_reason: "Connects checked convergence to queue lifecycle, managed binary distribution, skills, and workflow documentation."
    model_requirements: ["cross-skill lifecycle design", "managed binary installation", "queue CAS", "documentation synchronization"]
    escalation_triggers: ["install becomes mandatory", "open PR consumes active WIP", "delivery and cleanup states collapse", "global drift is newer"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:kb-start", "skill:kb-complete", "skill:kb-finalize", "installer:managed-binaries", "docs:workflow"]
    proof_check:
      kind: command_exit
      command: "go test ./... && node --test ./bin/kb-install.test.mjs && go run ./cmd/kbcheck skill-lint --root ."
      expect: 0
    status: pending
    owner: agent
    can_continue_other_slices: false
---

# Global Cleanup Reconciliation

The manifest delivers one fail-closed global reconciliation product in three
observable increments. The current session coordinates serial execution because
the slices share the Go core, Git safety contract, installer, and KB lifecycle
surfaces.
