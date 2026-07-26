---
type: kb-manifest
kb_id: kb-2026-07-26-plan-run-worktree-isolation
brainstorm_path: direct-chat
created: 2026-07-26
status: active
workflow_shape: pipeline-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck plan-worktree-selftest"
  expect: 0
  why: "Proves concurrent manifests use separate plan-run branches, conflicting runs are blocked before mutation, slice receipts integrate only into their owning run branch, and default-branch delivery remains separately authorized."
model_tier_contract:
  allowed: [small, medium, large]
  default: large
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  owner_choice: current-or-delegated
  max_owner_decisions_per_slice: 1
  catalog: active-host-plus-user-local
  delegated_fallback: same-tier-then-higher
  automatic_downward_routing: false
  amr_required: false
  automatic_cross_owner_fallback: false
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
  plan_run_worktree_default: true
  slice_worktrees_optional: true
  internal_integration_target: plan-run-branch
  default_branch_delivery_owner: kb-complete
  allowed_modes: [shared-serial, worktree-required]
context_packet_contract:
  schema_version: 1
execution_preconditions:
  - "Do not execute against the current overlapping dirty baseline."
  - "Contain or complete the DDR route-announcement changes that touch kb-work and workflow documentation."
  - "Restore responsive core/local-release proof or preserve the final release slice as blocked."
  - "The bootstrap implementation runs one mutator at a time until the new plan-run scheduler proves safe concurrency."
external_dependencies:
  - path: docs/plans/2026-07-26-000-kb-ddr-route-announcement-manifest.md
    reason: "Current dirty edits overlap kb-work and workflow architecture surfaces required by this plan."
  - path: docs/plans/2026-07-20-000-kb-harness-validation-recovery-manifest.md
    reason: "The final release gate needs responsive core and local-release checks."
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "The direct conversation defines the plan-run worktree, cross-manifest scheduling, integration, and delivery boundaries."
      - "Current source and tests prove the existing helper merges into the source checkout and rejects expected integration-head movement."
      - "No unresolved product or safety decision is required to decompose the work."
    proof:
      - "Direct user decision in the active task"
      - cmd/kbcheck/worktree_isolation.go
      - .github/skills/kb-work/references/worktree-isolation.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-plan plan-run worktree isolation and team-safe integration"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "Manifest and five slice plans exist."
      - "All five context packets validate."
      - "The DAG has no missing blockers or cycles."
      - "Every slice declares acceptance criteria, expected files, proof, risk, model tier, workspace mode, and conflict domains."
      - "The current overlapping dirty baseline is contained before mutation."
    proof:
      - docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md
      - docs/plans/2026-07-26-011-tool-plan-run-workspace-plan.md
      - docs/plans/2026-07-26-012-tool-cross-manifest-scheduler-plan.md
      - docs/plans/2026-07-26-013-tool-plan-run-integration-plan.md
      - docs/plans/2026-07-26-014-tool-delivery-boundary-plan.md
      - docs/plans/2026-07-26-015-eval-plan-worktree-lifecycle-plan.md
      - docs/plans/2026-07-26-plan-run-worktree-context/slice-001.json
      - docs/plans/2026-07-26-plan-run-worktree-context/slice-002.json
      - docs/plans/2026-07-26-plan-run-worktree-context/slice-003.json
      - docs/plans/2026-07-26-plan-run-worktree-context/slice-004.json
      - docs/plans/2026-07-26-plan-run-worktree-context/slice-005.json
      - "The five-slice DAG is a linear dependency chain with no missing references or cycle."
      - "Containment commit 3f1d916528453664a22560097e5f4b593215e74a on the local DDR topic branch was reviewed and its worktree is clean."
      - "Contributor core and the full command package pass in the clean containment worktree; final release remains blocked only by unrelated global drift already preserved for slice 005."
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "A manifest maps idempotently to one plan-run receipt, topic branch, and worktree."
      - "The receipt records immutable base identity and an explicit non-default integration ref and head."
      - "Dirty source state is preserved and excluded from the plan-run worktree."
      - "Owner mismatch, default-branch targeting, and unsafe release fail closed."
      - "The protected oracle showed RED before implementation and remained unchanged through GREEN."
    proof:
      - cmd/kbcheck/plan_run_workspace.go
      - cmd/kbcheck/plan_run_workspace_test.go
      - cmd/kbcheck/manifest_contract.go
      - cmd/kbcheck/manifest_contract_test.go
      - .github/skills/kb-plan/SKILL.md
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-work/references/worktree-isolation.md
      - "Focused plan-run workspace and manifest contract proof passed."
      - "The full command package passed and the protected oracle hash remained unchanged."
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md"
slices:
  - id: slice-001
    title: "Create a manifest-owned plan-run workspace"
    path: docs/plans/2026-07-26-011-tool-plan-run-workspace-plan.md
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "This establishes Git ownership and recovery state around dirty user work, where an incorrect base or target can lose or silently omit changes."
    model_requirements: ["Git worktree and branch-lifecycle reasoning", "fail-closed Go CLI contracts", "Windows and Unix path handling", "backward-compatible receipt migration"]
    escalation_triggers: ["a plan-run workspace requires force cleanup", "dirty uncommitted changes must be copied implicitly", "the default branch becomes the internal integration target", "legacy slice receipts cannot be migrated safely"]
    context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-001.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'PlanRunWorkspace|PlanRunManifestContract' -count=1"
      expect: 0
    hitl: false
    status: done
    workspace_mode: shared-serial
    conflict_domains: [file:cmd/kbcheck/plan_run_workspace.go, file:cmd/kbcheck/manifest_contract.go, skill:kb-plan, skill:kb-work]
    shared_resources: [git:integration-owner]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Proceed to slice-002 and extend plan-run ownership with atomic cross-manifest conflict claims."
    human_action: ""
    can_continue_other_slices: false
    notes: "execution_owner=current; owner_reason=reasoner-required for Git and dirty-work authority boundaries; route announcement emitted before mutation; slice lease generation 1 acquired for the eight expected files plus git integration ownership; scope-forecast: loaded 8 expected files plus 3 lifecycle files; scope-discovery: plan-run workspace test is the protected convention-matched oracle; RED: undefined plan-run workspace types and executor; GREEN: focused plan-run workspace and manifest contract tests pass; full cmd/kbcheck package passes; protected-oracle SHA256 82645d4bdbce01a665076e4dbb24794e90108b32e8a858bb9a2673c90a26b6db preserved; functional proof: public plan-worktree status command parsed and failed closed with an explicit migration message when no receipt existed; scope-check: forecast=11 changed=11 discovered=0 unexplained=0; test-level: functional-cli; functional-risk: broad; qa-lint: PASS gofmt and git diff --check; qa-browser: skipped - no UI behavior changed; memory-impact: durable; docs refreshed in kb-plan, kb-work, and the worktree isolation reference."
    protected_oracles:
      - path: cmd/kbcheck/plan_run_workspace_test.go
        role: "manifest-owned workspace and immutable-base oracle"
        sha256: "82645d4bdbce01a665076e4dbb24794e90108b32e8a858bb9a2673c90a26b6db"
        update_policy: "requires explicit plan amendment"
  - id: slice-002
    title: "Block cross-manifest conflicts before mutation"
    path: docs/plans/2026-07-26-012-tool-cross-manifest-scheduler-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Cross-process scheduling and shared-resource claims must be atomic and fail closed or multiple valid manifests will still race."
    model_requirements: ["cross-process lease design", "Git common-directory identity", "path and resource conflict normalization", "deterministic contention tests"]
    escalation_triggers: ["two conflicting manifests can both acquire mutation authority", "separate clones are falsely reported as coordinated", "a stale run can be stolen without token and generation proof", "shared resources cannot be represented without a daemon"]
    context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-002.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'PlanRunLease|CrossManifestScheduler|ScopeLease' -count=1"
      expect: 0
    hitl: false
    status: pending
    workspace_mode: shared-serial
    conflict_domains: [file:cmd/kbcheck/plan_run_scheduler.go, file:cmd/kbcheck/slice_lease.go, skill:kb-work]
    shared_resources: [git:integration-owner, git:plan-run-lease]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Extend common-dir ownership from slice-only claims to manifest-level conflict and shared-resource claims."
    human_action: ""
    can_continue_other_slices: false
    notes: "Local leases coordinate sibling worktrees only; remote team coordination remains branch/PR based."
    protected_oracles:
      - path: cmd/kbcheck/cross_manifest_scheduler_test.go
        role: "cross-manifest path and shared-resource exclusion oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan amendment"
  - id: slice-003
    title: "Integrate slice receipts only into the owning plan-run branch"
    path: docs/plans/2026-07-26-013-tool-plan-run-integration-plan.md
    blockers: [slice-002]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Serialized Git integration with expected concurrent head movement, conflict recovery, and proof replay is data-loss-sensitive architecture."
    model_requirements: ["three-way Git integration tests", "compare-and-swap integration-head state", "recoverable conflict handling", "post-integration proof enforcement"]
    escalation_triggers: ["integration can run in an arbitrary source branch", "the second disjoint receipt is rejected only because the first moved the integration head", "a conflict destroys the worker branch or lease", "proof is accepted only from the worker checkout"]
    context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-003.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'PlanRunIntegrate|IntegrationHead|ParallelReceipt' -count=1"
      expect: 0
    hitl: false
    status: pending
    workspace_mode: shared-serial
    conflict_domains: [file:cmd/kbcheck/worktree_isolation.go, git:plan-run-branch, skill:kb-work]
    shared_resources: [git:integration-owner, git:plan-run-branch]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Replace source-HEAD-equals-base integration with serialized integration-head ownership on the explicit plan-run branch."
    human_action: ""
    can_continue_other_slices: false
    notes: "Workers return commits/diffs/proof; only the plan-run coordinator integrates and reruns proof."
    protected_oracles:
      - path: cmd/kbcheck/plan_run_integration_test.go
        role: "two-receipt serialized integration and conflict-recovery oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan amendment"
  - id: slice-004
    title: "Keep default-branch delivery and dirty-WIP authority outside kb-work"
    path: docs/plans/2026-07-26-014-tool-delivery-boundary-plan.md
    blockers: [slice-003]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "This is the team-safety boundary between reversible internal integration and authorized PR or direct-default delivery."
    model_requirements: ["Git remote/default-branch detection", "delivery-policy reasoning", "dirty-work and commit-authority safety", "cross-skill contract testing"]
    escalation_triggers: ["kb-work can merge or push the resolved remote default", "worktree execution silently commits user-owned dirty files", "absence of delivery policy becomes direct delivery", "team mode depends on local common-dir locks"]
    context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-004.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'DefaultBranchBoundary|DirtyBaseAuthority|DeliveryOwner' -count=1"
      expect: 0
    hitl: false
    status: pending
    workspace_mode: shared-serial
    conflict_domains: [git:default-branch, skill:kb-work, skill:kb-complete, skill:kb-ship, skill:kb-land]
    shared_resources: [git:integration-owner, git:delivery-owner]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Fail closed on default-branch internal integration and route completed plan branches through configured local, PR, or explicitly authorized direct delivery."
    human_action: ""
    can_continue_other_slices: false
    notes: "No policy defaults to local; PR/manual is the portable team recommendation; direct remains explicit."
    protected_oracles:
      - path: cmd/kbcheck/delivery_boundary_test.go
        role: "default-branch refusal and dirty-work authority oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan amendment"
  - id: slice-005
    title: "Prove and release the multi-plan worktree lifecycle"
    path: docs/plans/2026-07-26-015-eval-plan-worktree-lifecycle-plan.md
    blockers: [slice-004]
    verification: integration
    test_level: full
    functional_risk: full
    model_tier: large
    model_tier_reason: "The final claim spans Git lifecycle, concurrency, skill behavior, team delivery policy, documentation, and global skill propagation."
    model_requirements: ["disposable multi-repo functional harness", "release-gate and sync verification", "cross-skill documentation audit", "failure-artifact preservation"]
    escalation_triggers: ["the functional harness mutates the real repository", "core or local-release remains unresponsive", "global skill drift contains useful target-only work", "any test requires force cleanup or default-branch delivery"]
    context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-005.json
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck plan-worktree-selftest"
      expect: 0
    hitl: false
    status: pending
    workspace_mode: shared-serial
    conflict_domains: [eval:plan-worktree-lifecycle, docs:workflow, sync:global-skills]
    shared_resources: [git:integration-owner, sync:global-skills]
    owner: agent
    blocked_reason: "Final release proof also depends on responsive core and local-release gates."
    resume_when: "Slices 001-004 are done and the harness-validation recovery has restored bounded core/local-release execution."
    next_agent_action: "Run the disposable multi-plan lifecycle, refresh public docs, review global drift, sync approved skills, and pass core plus local-release."
    human_action: ""
    can_continue_other_slices: true
    notes: "The selftest is the objective oracle; release remains incomplete until repo and sync gates also pass."
    protected_oracles:
      - path: cmd/kbcheck/plan_worktree_selftest_test.go
        role: "end-to-end multi-plan lifecycle oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan amendment"
---

# KB: Plan-Run Worktree Isolation and Team-Safe Integration

## Origin

Direct conversation decision: multiple active sets of plans need independent
workspaces, but worktree isolation must not silently defer collisions to a merge
into `main`.

## Workflow Shape

`pipeline-change` — the correction spans manifest contracts, local concurrency,
Git worktree lifecycle, integration receipts, delivery authority, deterministic
functional proof, documentation, and required skill propagation.

## Decisions Carried Forward

- A plan set means one KB manifest/workstream, not every individual plan file.
- Every concurrently mutating manifest gets one plan-run branch and worktree.
- One mutator per plan-run is the default; child slice worktrees are optional.
- Expected files are forecasts; observed path/resource overlap wins.
- Slice results integrate serially into the owning plan-run branch.
- `kb-work` never implicitly integrates or delivers to the remote default branch.
- No delivery policy means local-only; PR/manual is the team-safe recommendation.
- Common-directory leases coordinate one clone only; branches, PR checks, and
  protection own cross-machine coordination.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Create a manifest-owned plan-run workspace | - | tdd / functional-cli | no | done |
| 2 | Block cross-manifest conflicts before mutation | slice-001 | tdd / functional-cli | no | pending |
| 3 | Integrate slice receipts only into the owning plan-run branch | slice-002 | tdd / functional-cli | no | pending |
| 4 | Keep default-branch delivery and dirty-WIP authority outside kb-work | slice-003 | tdd / functional-cli | no | pending |
| 5 | Prove and release the multi-plan worktree lifecycle | slice-004 + external release gate | integration / full | no | pending |

## Execution Gate

`plan-to-work` passed after the overlapping DDR scope was reviewed and contained
on local branch `codex/ddr-route-announcement-containment` at commit `3f1d916`.
Bootstrap execution remains shared-serial, and final release remains separately
blocked on the pre-existing non-`kb-work` global drift recorded by
`local-release`.
