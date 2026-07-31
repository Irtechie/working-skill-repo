---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-07-31-automatic-delivery-chain
brainstorm_path: direct-chat
created: 2026-07-31
status: reviewed
workflow_shape: pipeline-change
scope-verified-files:
  - .github/skills/kb-work/SKILL.md
  - .github/skills/kb-finalize/SKILL.md
  - .github/skills/kb-complete/SKILL.md
  - .github/skills/kb-ship/SKILL.md
  - .github/skills/kb-land/SKILL.md
  - cmd/kbcheck/delivery_chain_contract_test.go
  - docs/context/architecture/kb-workflow.md
  - docs/context/architecture/skills.md
  - README.md
  - evals/route-complexity/finish-plan-flow.json
  - evals/route-complexity/release-ship-flow.json
  - evals/skill-eval/selftest/pass-finish-plan-flow.json
  - docs/plans/2026-07-09-020-kb-plan-to-pr-finish-manifest.md
  - docs/plans/2026-07-09-022-kb-finish-orchestrator-plan.md
  - docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md
  - docs/plans/2026-07-31-001-skill-automatic-delivery-chain-plan.md
  - todo.md
  - todo-done.md
final_audited_shipping_scope:
  - .github/skills/kb-work/SKILL.md
  - .github/skills/kb-finalize/SKILL.md
  - .github/skills/kb-complete/SKILL.md
  - .github/skills/kb-ship/SKILL.md
  - .github/skills/kb-land/SKILL.md
  - cmd/kbcheck/delivery_chain_contract_test.go
  - docs/context/architecture/kb-workflow.md
  - docs/context/architecture/skills.md
  - README.md
  - evals/route-complexity/finish-plan-flow.json
  - evals/route-complexity/release-ship-flow.json
  - evals/skill-eval/selftest/pass-finish-plan-flow.json
  - docs/plans/2026-07-09-020-kb-plan-to-pr-finish-manifest.md
  - docs/plans/2026-07-09-022-kb-finish-orchestrator-plan.md
  - docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md
  - docs/plans/2026-07-31-001-skill-automatic-delivery-chain-plan.md
  - todo.md
  - todo-done.md
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
  not_required_reason: "The user specified the complete chain, exact safety boundaries, delivery authorization, required proof, main-containment evidence, and installed-copy sync; no product, architecture, trust, or scope decision remains open."
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "Proves the exact reviewed source tree, deterministic workflow contracts, release hygiene, and read-only installed-skill drift state before delivery."
model_tier_contract:
  allowed: [small, medium, large]
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
delivery_authority:
  source: explicit-same-run-user-authorization
  mode: pr
  merge: auto-after-checks
  post_merge_sync: true
  authorized_actions: [commit, push-topic, create-or-update-pr, merge-after-required-checks, sync-installed-skills]
  forbidden_actions: [force-push, bypass-branch-protection, bypass-required-checks, bypass-required-reviews, kb-work-default-integration, kb-ship-merge]
plan_run_worktree:
  branch: deaderestpool-auto-ship-and-land
  worktree: C:\Users\marowe\.copilot\repos\copilot-worktrees\working-skill-repo\deaderestpool-refactored-chainsaw
  workspace_mode: shared-serial
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Successful kb-work completion must automatically continue through kb-finalize and kb-complete."
      - "kb-complete must invoke kb-ship, then invoke authorized kb-land according to policy or same-run authorization."
      - "kb-work must never merge or push the resolved default branch, kb-ship must never merge, and only kb-land may integrate remote default."
      - "The user explicitly authorized commit, push, PR delivery, merge-to-main after existing gates, and post-integration installed-skill sync."
      - "Branch protections, checks, and reviews must not be bypassed."
    proof:
      - direct-chat
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-finalize/SKILL.md
      - .github/skills/kb-complete/SKILL.md
      - .github/skills/kb-ship/SKILL.md
      - .github/skills/kb-land/SKILL.md
    blockers: []
    passed_at: "2026-07-31T09:25:00Z"
    allowed_next_action: "kb-plan direct-chat"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The manifest and one end-to-end slice plan exist."
      - "The slice maps every requested transition, safety boundary, deterministic proof surface, documentation update, and stale contradictory active plan."
      - "The single-slice DAG has no missing blockers or cycles."
      - "The slice declares acceptance criteria, expected files, verification, test level, execution class, model tier, proof check, and HITL classification."
      - "Exact delivery authority and forbidden bypasses are recorded."
    proof:
      - docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md
      - docs/plans/2026-07-31-001-skill-automatic-delivery-chain-plan.md
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-complete/SKILL.md
      - .github/skills/kb-land/SKILL.md
    blockers: []
    passed_at: "2026-07-31T09:25:00Z"
    allowed_next_action: "kb-work docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The deterministic delivery-chain contract first failed against the old terminal behavior and now passes."
      - "The route fixture evaluates successfully with kb-land, remote-default containment, and installed-skill sync evidence."
      - "The five phase owners retain their explicit mutation boundaries."
      - "The source diff contains no whitespace errors."
    proof:
      - cmd/kbcheck/delivery_chain_contract_test.go
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-finalize/SKILL.md
      - .github/skills/kb-complete/SKILL.md
      - .github/skills/kb-ship/SKILL.md
      - .github/skills/kb-land/SKILL.md
      - evals/skill-eval/selftest/pass-finish-plan-flow.json
    proof_commands:
      - "go test ./cmd/kbcheck -run 'AutomaticDeliveryChain|DeliveryOwnerSkillContracts' -count=1"
      - "go run ./cmd/kbcheck route-eval --root ."
      - "go run ./cmd/kbcheck skill-eval --root . --result-path evals/skill-eval/selftest/pass-finish-plan-flow.json --json"
      - "git diff --check"
    blockers: []
    passed_at: "2026-07-31T09:40:00Z"
    allowed_next_action: "kb-work docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md"
  - gate_id: work-to-complete
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Every slice is done or intentionally skipped."
      - "The integrated workflow contract and route fixtures pass."
      - "The board and manifest are synchronized."
      - "scope-verified-files covers every intentional source change."
      - "The contributor-safe core gate passes 39 of 39 checks."
    proof:
      - docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md
      - docs/plans/2026-07-31-001-skill-automatic-delivery-chain-plan.md
      - cmd/kbcheck/delivery_chain_contract_test.go
      - todo.md
      - README.md
    proof_commands:
      - "go run ./cmd/kbcheck core"
      - "git diff --check"
    blockers: []
    passed_at: "2026-07-31T09:40:00Z"
    allowed_next_action: "kb-finalize docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md"
  - gate_id: complete-to-ship
    owner_skill: kb-finalize
    gate_scope: release
    status: quarantined
    required_evidence:
      - "Final exact-tree core proof passes 39 of 39 deterministic checks."
      - "The focused delivery-chain, authority-ledger mutation, route, fixture, lint, and diff checks pass."
      - "One broad semantic review completed and its P1/P2 findings were resolved."
      - "The same reviewer confirmed no remaining P0-P2 findings after the last fix."
      - "No Cargo command ran and the native not-applicable cleanup receipt validates."
      - "The only local-release failure is 15 expected installed-copy drifts for five changed skills across three required global roots."
      - "All installed copies matched source before this edit, so no global-only work is being overwritten."
      - "The quarantined external install paths do not overlap the audited repository shipping scope."
    proof:
      - cmd/kbcheck/delivery_chain_contract_test.go
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-finalize/SKILL.md
      - .github/skills/kb-complete/SKILL.md
      - .github/skills/kb-ship/SKILL.md
      - .github/skills/kb-land/SKILL.md
      - evals/skill-eval/selftest/pass-finish-plan-flow.json
      - docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md
    proof_commands:
      - "go run ./cmd/kbcheck core"
      - "go test ./cmd/kbcheck -run 'AutomaticDeliveryChain|DeliveryAuthorityLedger|DeliveryOwnerSkillContracts' -count=1"
      - "go run ./cmd/kbcheck route-eval --root ."
      - "go run ./cmd/kbcheck skill-eval --root . --result-path evals/skill-eval/selftest/pass-finish-plan-flow.json --json"
      - "go run ./cmd/kbcheck skill-lint --root ."
      - "git diff --check"
      - "go run ./cmd/kbcheck cargo-storage --action validate --run-id auto-chain-kb-delivery-20260731 --root . --json"
      - "go run ./cmd/kbcheck local-release"
      - "go run ./cmd/kbcheck skill-sync-report --json"
    blockers: []
    quarantined_scope: "External deployed copies under ~/.codex/skills, ~/.copilot/skills, and ~/.agents/skills for kb-work, kb-finalize, kb-complete, kb-ship, and kb-land until origin/main containment is proven."
    quarantine_owner: "kb-land post-integration sync"
    quarantine_evidence:
      - "local-release passed core and git-diff-check; only skill-sync-report failed."
      - "skill-sync-report reported exactly 15 drift-required rows: five changed skills across three required global roots."
      - "Pre-edit SHA-256 comparison showed each source SKILL.md matched all three installed copies."
    forbidden_claims:
      - "Do not claim local-release fully passed before post-land sync."
      - "Do not claim installed skills match the landed source before remote-default containment and sync proof."
      - "Do not sync unlanded source into global runner roots."
    passed_at: "2026-07-31T10:05:00Z"
    allowed_next_action: "kb-complete docs/plans/2026-07-31-000-kb-automatic-delivery-chain-manifest.md"
slices:
  - id: slice-001
    title: "Make successful KB work continue through authorized delivery"
    path: docs/plans/2026-07-31-001-skill-automatic-delivery-chain-plan.md
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    execution_class: cli
    model_tier: large
    model_tier_reason: "The change spans five phase owners, external GitHub mutations, default-branch authority, branch-protection behavior, deterministic contract tests, and post-integration sync; an ambiguous transition could publish or merge without the correct owner."
    model_requirements: ["cross-skill contract reasoning", "GitHub delivery safety", "deterministic Go contract tests", "release and sync workflow knowledge"]
    escalation_triggers: ["a phase would gain another phase's mutation authority", "required checks or reviews cannot be observed", "the active PR cannot target resolved remote default", "installed drift contains newer useful work"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:kb-work", "skill:kb-finalize", "skill:kb-complete", "skill:kb-ship", "skill:kb-land", "git:default-branch", "docs:kb-workflow"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'AutomaticDeliveryChain|DeliveryOwnerSkillContracts' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Invoke kb-finalize for exact-tree proof, one semantic review, memory refresh, and cleanup."
    human_action: ""
    can_continue_other_slices: true
    notes: "Targeted delivery contract, route eval, fixture eval, skill lint, diff check, and core 39/39 pass. Browser proof skipped because no UI-reachable behavior changed."
semantic_review:
  profile: code-review
  mode: headless
  reviewer_provenance: "delivery-chain-review"
  base_commit: a0273bbbc047c2012c2107d18f10a2e3c10d1c07
  requirements_sha256: 09ac8edf19b0427ea16c5c8d56d0e5eefb289061d8743dc17991178358c92fa0
  initial_findings: {p0: 0, p1: 1, p2: 1, p3: 0}
  confirmation_findings: {p0: 0, p1: 0, p2: 0, p3: 0}
  follow_up_resolution: "resolved 4, logged 0, blocked 0"
  residual_risks: []
knowledge_memory:
  kb_map_refresh: "Updated the durable workflow architecture and skill route map in this change."
  compound: "skipped - no novel implementation problem beyond the workflow contract itself."
  learn: "skipped - reviewer findings were local test/contract defects, not a repeated cross-workflow lesson."
  evolve: "skipped - no mature instinct promotion signal."
  memory_review: "skipped - no due maintenance signal or bloat threshold."
cleanup:
  cargo_storage: "not-applicable; native receipt validated"
  ephemeral_artifacts: "none"
  alerts: "Post-land global skill sync remains quarantined and mandatory."
---

# Automatic KB Delivery Chain

One vertical slice changes the observable workflow and its executable contract
proof together. Delivery remains gated: automatic continuation selects and
invokes the existing phase owner; it does not transfer default-branch authority
to `kb-work` or merge authority to `kb-ship`.
