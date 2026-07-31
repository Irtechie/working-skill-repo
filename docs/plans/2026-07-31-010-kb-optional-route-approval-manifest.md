---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-07-31-optional-route-approval
brainstorm_path: direct-chat
created: 2026-07-31
status: reviewed
workflow_shape: pipeline-change
scope-verified-files:
  - internal/modelrouting/catalog.go
  - internal/modelrouting/policy.go
  - internal/modelrouting/storage.go
  - cmd/kbrouter/main.go
  - cmd/kbrouter/catalog.go
  - cmd/kbrouter/ddr.go
  - cmd/kbrouter/dispatch.go
  - cmd/kbrouter/catalog_test.go
  - cmd/kbrouter/ddr_test.go
  - .github/skills/kb-models/SKILL.md
  - docs/context/architecture/kbrouter.md
  - LOCAL_MODELS.example.md
  - README.md
  - todo.md
  - docs/plans/2026-07-31-010-kb-optional-route-approval-manifest.md
  - docs/plans/2026-07-31-011-cli-optional-route-approval-plan.md
objective_contract: true
blocker_lifecycle_contract: true
pre_slice_review_contract: true
model_tier_contract: true
workspace_isolation_contract: true
proof_governor_contract: true
pre_slice_review:
  status: not-required
  source: direct-chat
  mode: requirements-wide
  not_required_reason: "The user explicitly rejected permission prompts for bounded routing loops while preserving attended approval as an opt-in mode. The default, bypass scope, retained safety checks, persistence location, and verification surface are fully specified."
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "Proves router behavior, skill contracts, repo quality, and read-only installed-copy drift before propagation."
model_tier_contract:
  allowed: [small, medium, large]
  default: large
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
delivery_authority:
  source: explicit-same-run-user-authorization
  mode: pr
  merge: auto-after-checks
  post_merge_sync: true
  authorized_actions: [edit, test, commit, push-topic, create-or-update-pr, merge-after-required-checks, install-local-router, sync-codex-global, sync-copilot-global, sync-shared-agents-global]
  forbidden_actions: [force-push, bypass-branch-protection, bypass-required-checks, bypass-required-reviews]
plan_run_worktree:
  branch: deaderestpool-make-router-trust-optional
  worktree: C:\Users\marowe\.copilot\repos\copilot-worktrees\working-skill-repo\deaderestpool-verbose-fiesta
  workspace_mode: shared-serial
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Attended endpoint/auth approval remains available but is no longer unavoidable."
      - "Missing mode defaults to disabled so bounded routing loops do not pause for permission."
      - "Users who want the attended boundary can explicitly persist required mode."
      - "Disabled approval does not bypass route denials, endpoint static safety, retention, sensitive-data, one-attempt, or proof controls."
      - "The user's local installed router is switched to disabled mode after verified installation."
    proof:
      - direct-chat
      - cmd/kbrouter/main.go
      - internal/modelrouting/policy.go
      - .github/skills/kb-models/SKILL.md
      - docs/context/architecture/kbrouter.md
    blockers: []
    passed_at: "2026-07-31T23:36:00Z"
    allowed_next_action: "kb-plan direct-chat"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The manifest and one end-to-end CLI slice exist."
      - "The slice includes public CLI behavior, policy enforcement, DDR revalidation, docs, install, and user-local activation."
      - "The single-slice DAG has no blockers or cycles."
      - "The protected CLI oracle is defined before implementation."
    proof:
      - docs/plans/2026-07-31-010-kb-optional-route-approval-manifest.md
      - docs/plans/2026-07-31-011-cli-optional-route-approval-plan.md
      - cmd/kbrouter/catalog.go
      - internal/modelrouting/policy.go
    blockers: []
    passed_at: "2026-07-31T23:36:00Z"
    allowed_next_action: "kb-work docs/plans/2026-07-31-010-kb-optional-route-approval-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: pending
    required_evidence:
      - "The unchanged protected oracle fails before implementation and passes afterward."
      - "Approval-disabled is the missing-field default."
      - "Approval-required remains available as an explicit opt-in."
      - "Approval-disabled permits configured routes without trust receipts while explicit denials and non-approval safety checks remain enforced."
      - "The installed router reports disabled mode and no longer requires trust.json for the configured Deepseek route."
    proof:
      - cmd/kbrouter/catalog_test.go
      - cmd/kbrouter/ddr_test.go
    proof_commands:
      - "go test ./cmd/kbrouter ./internal/modelrouting -run 'ApprovalMode|OptionalRouteApproval' -count=1"
      - "go run ./cmd/kbcheck core"
      - "go run ./cmd/kbcheck local-release"
      - "git diff --check"
    blockers: []
    allowed_next_action: "kb-work docs/plans/2026-07-31-010-kb-optional-route-approval-manifest.md"
slices:
  - id: slice-001
    title: "Let users disable attended route approval without disabling router safety"
    path: docs/plans/2026-07-31-011-cli-optional-route-approval-plan.md
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    execution_class: cli
    model_tier: large
    model_tier_reason: "This changes a trust boundary across catalog persistence, policy evaluation, endpoint/auth validation, dispatch fallback, and DDR replay checks."
    model_requirements: ["security-boundary reasoning", "Go CLI integration", "backward-compatible state migration", "cross-surface documentation"]
    escalation_triggers: ["disabled mode bypasses explicit denials", "static endpoint safety is weakened", "required mode no longer enforces attended approval", "DDR can redispatch or skip proof"]
    workspace_mode: shared-serial
    conflict_domains: ["cmd:kbrouter", "internal:modelrouting", "skill:kb-models", "docs:model-routing", "install:user-router"]
    shared_resources: ["user-state:~/.kb/models.json", "binary:~/.kb/bin/kbrouter.exe", "install:global-kb-models"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbrouter ./internal/modelrouting -run 'ApprovalMode|OptionalRouteApproval' -count=1"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Write the protected CLI oracle, prove RED, implement, run release gates, install, and persist disabled mode."
    human_action: ""
    can_continue_other_slices: true
---

# Optional Route Approval

One vertical CLI slice adds a persisted user-local approval mode, applies it to
selection, endpoint/auth validation, dispatch fallback, and DDR revalidation,
then documents, installs, and activates the user's explicit choice.
