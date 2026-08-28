---
type: kb-manifest
kb_id: kb-2026-07-26-plan-run-worktree-isolation
brainstorm_path: direct-chat
created: 2026-07-26
status: completed
workflow_shape: pipeline-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck plan-worktree-selftest"
  expect: 0
  why: "Proves concurrent manifest groups use separate plan-run branches, conflicting runs are blocked before mutation, slices advance only their owning run branch, and default-branch delivery remains separately authorized."
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
  slice_worktrees_optional: false
  internal_integration_target: plan-run-branch
  default_branch_delivery_owner: kb-complete
  allowed_modes: [shared-serial]
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
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Conflicting manifest groups cannot both acquire mutation authority."
      - "Disjoint manifest groups remain concurrently admissible."
      - "File, prefix, domain, and resource claims compose with slice leases under one run lineage."
      - "Owner token and generation protect renew, expansion, release, and recovery."
      - "The protected contention oracle showed RED before implementation and remained unchanged through GREEN."
    proof:
      - cmd/kbcheck/plan_run_scheduler.go
      - cmd/kbcheck/cross_manifest_scheduler_test.go
      - cmd/kbcheck/slice_lease.go
      - cmd/kbcheck/slice_lease_test.go
      - cmd/kbcheck/swarm.go
      - cmd/kbcheck/swarm_test.go
      - .github/skills/kb-work/SKILL.md
      - "Focused scheduler, scope lease, and plan-run contract proof passed."
      - "Ten consecutive contention runs and the full command package passed."
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "A slice commit is accepted only from the exact owning plan-run branch and worktree."
      - "Owner lineage, clean state, prior-head CAS, descendant/current HEAD, and proof receipt all fail closed."
      - "Slice and aggregate proof replay before the receipt head advances."
      - "The advance path creates no branch/worktree and performs no merge, reset, stash, cleanup, push, or delivery."
      - "The protected oracle showed RED before implementation and remained unchanged through GREEN."
    proof:
      - cmd/kbcheck/plan_run_integration_test.go
      - cmd/kbcheck/plan_run_workspace.go
      - cmd/kbcheck/main.go
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-work/references/worktree-isolation.md
      - .github/skills/kb-work/references/execution-prompt.md
      - "Focused plan-run advance, integration-head, and slice-commit proof passed."
      - "Full command package and diff checks passed."
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Internal targets resolving to local or fetched remote defaults fail closed."
      - "Absent delivery policy remains local-only and PR/manual never merges."
      - "Only explicitly authorized kb-land owns remote-default integration."
      - "Relevant dirty WIP blocks without hidden commit, stash, reset, or omission; unrelated dirt remains unchanged."
      - "The protected oracle showed RED before implementation and remained unchanged through GREEN."
    proof:
      - cmd/kbcheck/delivery_boundary_test.go
      - cmd/kbcheck/worktree_isolation.go
      - cmd/kbcheck/plan_run_workspace.go
      - cmd/kbcheck/skill_repo_contract_test.go
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-complete/SKILL.md
      - .github/skills/kb-ship/SKILL.md
      - .github/skills/kb-land/SKILL.md
      - .github/skills/kb-configure/SKILL.md
      - "Focused default-branch, dirty-authority, and delivery-owner proof passed."
      - "Full command package and diff checks passed."
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md"
  - gate_id: slice-slice-005-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "The disposable two-manifest lifecycle proves one worktree per manifest group and shared-serial slice commits."
      - "The final command package and contributor core pass after review repairs."
      - "Useful target-only global drift is merged back before required skill roots are synchronized."
      - "The blocking local release gate passes with zero required sync issues."
      - "Independent correctness, structure, and testing review has no unresolved P0/P1."
    proof:
      - "kbcheck plan-worktree-selftest result: exit 0; runs=2 commits=4 collisions=2 source-unchanged=true delivery-stopped=true"
      - "Go command package test result: exit 0; duration 212.369s"
      - docs/solutions/workflow-issues/plan-run-worktree-isolation-2026-07-26.md
      - "kbcheck skill-sync-report result: ok true; required issues zero"
      - "kbcheck local-release result: profile local-release; ok true"
      - "kb-review: multi-agent; final unresolved P0=0 P1=0 P2=0"
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md"
  - gate_id: work-to-complete
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Every non-skipped slice has a passing terminal gate."
      - "The top-level objective and final release checks pass."
      - "No unresolved P0/P1 review finding remains."
      - "Board, manifest, project memory, and handoff lifecycle agree."
      - "The final changed-file scope is recorded."
    proof:
      - "slice-slice-001-to-done through slice-slice-005-to-done: passed"
      - "kbcheck plan-worktree-selftest result: exit 0"
      - "kbcheck local-release result: profile local-release; ok true"
      - "kb-review multi-agent resolution: P0=0 P1=0 unresolved"
      - todo-done.md
      - docs/handoffs/done/2026-07-26-plan-run-worktree-isolation.md
      - "scope-verified-files: exact git diff against containment base 3f1d916; no unrelated shared-checkout files included"
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-finalize docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md"
  - gate_id: complete-to-ship
    owner_skill: kb-finalize
    status: passed
    required_evidence:
      - "Final deterministic and release checks pass."
      - "Objective and per-slice proof results are recorded."
      - "Functional CLI/full coverage exercises the public workflow."
      - "Multi-agent review and finding resolution are recorded."
      - "Follow-up resolution has no blocked item."
      - "Proof and demo evidence is durable."
      - "Compound, learn, and evolve cadence results are recorded."
      - "Project memory and maintenance are refreshed."
      - "Cleanup and todo/handoff hygiene are complete."
      - "Alerts are explicit."
    proof:
      - "kbcheck local-release result: profile local-release; ok true"
      - "kbcheck plan-worktree-selftest result: exit 0; four accepted commits"
      - "All five slice proof checks and terminal gates passed"
      - cmd/kbcheck/plan_worktree_selftest_test.go
      - "review mode multi-agent; final unresolved critical, important, and suggestion findings zero"
      - "follow-up resolution: resolved 9; logged 0; blocked 0"
      - docs/solutions/workflow-issues/plan-run-worktree-isolation-2026-07-26.md
      - docs/context/kb/instincts/scoped/workflow/plan-execution.yaml
      - "evolve cadence: no candidates met the confidence and observation thresholds"
      - docs/context/PROJECT.md
      - docs/context/memory-maintenance.md
      - docs/handoffs/done/2026-07-26-plan-run-worktree-isolation.md
      - "cleanup: no screenshots; five current observations retained; todo contains no completed workstream"
      - "alert: memory review recommended because maintenance thresholds were already crossed"
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-complete docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md"
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
    notes: "execution_owner=current; owner_reason=reasoner-required for Git and dirty-work authority boundaries; route announcement emitted before mutation; slice lease generation 1 acquired for the eight expected files plus git integration ownership; scope-forecast: loaded 8 expected files plus 3 lifecycle files; scope-discovery: plan-run workspace test is the protected convention-matched oracle; RED: undefined plan-run workspace types and executor; GREEN: focused plan-run workspace and manifest contract tests pass; full cmd/kbcheck package passes; explicit plan amendment: review-driven fresh-repository, authorization, terminal-proof, and completion-retry cases updated the protected oracle to SHA256 f3b5375fa6a89e0e40975aa903443d80497bcc40f1957a694413011eb1ab159e; functional proof: public plan-worktree status command parsed and failed closed with an explicit migration message when no receipt existed; scope-check: forecast=11 changed=11 discovered=0 unexplained=0; test-level: functional-cli; functional-risk: broad; qa-lint: PASS gofmt and git diff --check; qa-browser: skipped - no UI behavior changed; memory-impact: durable; docs refreshed in kb-plan, kb-work, and the worktree isolation reference."
    protected_oracles:
      - path: cmd/kbcheck/plan_run_workspace_test.go
        role: "manifest-owned workspace and immutable-base oracle"
        sha256: "f3b5375fa6a89e0e40975aa903443d80497bcc40f1957a694413011eb1ab159e"
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
    status: done
    workspace_mode: shared-serial
    conflict_domains: [file:cmd/kbcheck/plan_run_scheduler.go, file:cmd/kbcheck/slice_lease.go, skill:kb-work]
    shared_resources: [git:integration-owner, git:plan-run-lease]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Proceed to slice-003 and protect serialized slice-commit advancement on the owning plan-run branch."
    human_action: ""
    can_continue_other_slices: false
    notes: "execution_owner=delegated; owner_reason=bounded native worker had exact packet and deterministic contention proof while coordinator retained lease, board, and commit authority; route announcement emitted before dispatch; slice lease generation 1 acquired for eight expected files plus integration and plan-run lease resources; scope-forecast: loaded 8 expected files plus 3 lifecycle files; RED: missing plan-run lease API; GREEN: exact PlanRunLease, CrossManifestScheduler, and ScopeLease proof passed; contention proof passed 10 consecutive runs; full cmd/kbcheck package and plan-run, slice-lease, and scope-lease selftests passed; protected-oracle SHA256 3816b2eb97f511390a04d43f94a3df18a419b88cc3a43e977f19860ba31b848a preserved; forecast hydration prevents silent underclaiming; observed expansion requeues before a colliding write; separate clones remain explicitly uncoordinated; user architecture amendment: one manifest group owns one worktree and all slices are shared-serial, with no per-slice worktrees; amendment refreshed manifest contract, kb-plan, worktree reference, slice-003, slice-005, and the slice-003 packet; scope-check: expected=8 lifecycle=3 amendment=8 changed=19 unexplained=0; qa-lint: PASS gofmt and git diff --check; qa-browser: skipped - no UI behavior changed; memory-impact: durable."
    protected_oracles:
      - path: cmd/kbcheck/cross_manifest_scheduler_test.go
        role: "cross-manifest path and shared-resource exclusion oracle"
        sha256: "3816b2eb97f511390a04d43f94a3df18a419b88cc3a43e977f19860ba31b848a"
        update_policy: "requires explicit plan amendment"
  - id: slice-003
    title: "Advance slice commits only on the owning plan-run branch"
    path: docs/plans/2026-07-26-013-tool-plan-run-integration-plan.md
    blockers: [slice-002]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Serialized commit acceptance, compare-and-swap head state, and proof replay on the sole plan-run branch are data-loss-sensitive architecture."
    model_requirements: ["plan-run branch identity tests", "compare-and-swap integration-head state", "single-worktree commit receipts", "post-commit proof enforcement"]
    escalation_triggers: ["a slice runs in another worktree or branch", "a commit is accepted after unexpected integration-head movement", "the plan-run worktree is dirty at receipt time", "proof is accepted only from worker self-report"]
    context_packet_path: docs/plans/2026-07-26-plan-run-worktree-context/slice-003.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'PlanRunIntegrate|IntegrationHead|ParallelReceipt' -count=1"
      expect: 0
    hitl: false
    status: done
    workspace_mode: shared-serial
    conflict_domains: [file:cmd/kbcheck/plan_run_workspace.go, git:plan-run-branch, skill:kb-work]
    shared_resources: [git:integration-owner, git:plan-run-branch]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Proceed to slice-004 and enforce default-branch delivery and dirty-WIP authority boundaries."
    human_action: ""
    can_continue_other_slices: false
    notes: "execution_owner=delegated; owner_reason=bounded native worker had the amended same-worktree packet and deterministic head/proof oracle while coordinator retained manifest lease, board, and commit acceptance; route announcement emitted before dispatch; plan-run lease generation 1 and slice lease generation 1 were active; scope-forecast: loaded 6 expected files plus 3 lifecycle files; RED: missing advance and proof-receipt API; GREEN: exact PlanRunAdvance, IntegrationHead, and SliceCommit proof passed; full cmd/kbcheck package passed; explicit plan amendment: review-driven exact-write, live-head, release, archive-tamper, and completion-journal cases updated the protected oracle to SHA256 0240a836e45d8483115ea1737d81f6ebebe682b4968a1a83931922bb8f399bd0; exact owner/run/worktree/ref lineage, clean state, prior-head CAS, strict descendant/current-HEAD checks, mandatory proof receipt, slice plus aggregate proof replay, and atomic receipt-head advancement are enforced; advance creates no branch/worktree and performs no merge, reset, stash, cleanup, push, or delivery; scope-check: expected=6 lifecycle=3 changed=9 unexplained=0; qa-lint: PASS gofmt and git diff --check; qa-browser: skipped - no UI behavior changed; memory-impact: durable."
    protected_oracles:
      - path: cmd/kbcheck/plan_run_integration_test.go
        role: "serialized same-worktree slice commit and integration-head oracle"
        sha256: "0240a836e45d8483115ea1737d81f6ebebe682b4968a1a83931922bb8f399bd0"
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
    status: done
    workspace_mode: shared-serial
    conflict_domains: [git:default-branch, skill:kb-work, skill:kb-complete, skill:kb-ship, skill:kb-land]
    shared_resources: [git:integration-owner, git:delivery-owner]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Proceed to slice-005 and run the disposable multi-manifest lifecycle plus release/sync proof."
    human_action: ""
    can_continue_other_slices: false
    notes: "execution_owner=delegated; owner_reason=bounded native worker had exact delivery-boundary packet and deterministic Git/skill contract proof while coordinator retained leases, lifecycle, review, and commit authority; route announcement emitted before dispatch; manifest and slice lease generation 1 were active; scope-forecast: loaded 9 expected files plus 3 lifecycle files and 1 corrected context packet; RED: missing default-branch and delivery-policy API; GREEN: exact DefaultBranchBoundary, DirtyBaseAuthority, and DeliveryOwner proof passed; full cmd/kbcheck package passed; explicit plan amendment: review-driven authorization and unresolved-remote-default cases updated the protected oracle to SHA256 fca79a7cdbcc5839d5ceabb15b1ca8408d554ccd42e549d698f1cd40520fd7a7; local and fetched remote defaults are forbidden internal targets; absent policy is local-only; PR/manual stops at an open PR without merge; only explicitly authorized kb-land owns remote-default integration; relevant dirty WIP blocks before creation while unrelated dirt remains untouched; local common-directory leases are explicitly not team-wide locks; scope-check: expected=9 lifecycle=3 packet=1 changed=13 unexplained=0; qa-lint: PASS gofmt and git diff --check; qa-browser: skipped - no UI behavior changed; memory-impact: durable."
    protected_oracles:
      - path: cmd/kbcheck/delivery_boundary_test.go
        role: "default-branch refusal and dirty-work authority oracle"
        sha256: "fca79a7cdbcc5839d5ceabb15b1ca8408d554ccd42e549d698f1cd40520fd7a7"
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
    status: done
    workspace_mode: shared-serial
    conflict_domains: [eval:plan-worktree-lifecycle, docs:workflow, sync:global-skills]
    shared_resources: [git:integration-owner, sync:global-skills]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Run kb-finalize against the completed manifest."
    human_action: ""
    can_continue_other_slices: true
    notes: "execution_owner=delegated; owner_reason=bounded native worker produced the initial disposable harness while the coordinator retained manifest/slice leases, global drift review, propagation, release gates, lifecycle, review, and commit authority; route announcement emitted before dispatch; scope-discovery expanded into commit-authority provenance, exact claim/write equality, live integration-head slice binding, durable slice-release evidence, immutable per-slice proof archives, and failure-idempotent completion journaling after independent review; objective proof: go run ./cmd/kbcheck plan-worktree-selftest exit 0 with runs=2 commits=4 collisions=2 source-unchanged=true delivery-stopped=true; functional proof: go test ./cmd/kbcheck -count=1 exit 0 in 212.369s; sync proof: skill-sync-report ok=true required_issues=0 after merging useful Copilot-only kb-gate command/path parsing back into source; release proof: local-release profile ok=true; review-mode: multi-agent; review: P0=0 P1=5(resolved) P2=4(resolved) P3=0, final unresolved P0/P1/P2=0; follow-up-resolution: resolved 9, logged 0, blocked 0; qa-browser: skipped - no rendered UI changed; demo evidence: CLI selftest output; compound: docs/solutions/workflow-issues/plan-run-worktree-isolation-2026-07-26.md; learn: 1 new scoped instinct and existing-confidence decay applied; evolve: no candidates ready; kb-map-refresh: done - PROJECT, workflow architecture, testing, eval map, solution, superseded plan, instinct, board, and handoff; memory-maintenance: one contradiction signal recorded and completion counter advanced to 15; compact: skipped - no new startup bloat; cleanup: no screenshots, observations retained within 90 days; alerts: memory review recommended by existing thresholds; bootstrap exception: this implementation worktree predates plan-worktree receipts, so no receipt was fabricated; scope-check: forecast plus review-discovered files all covered by expanded manifest/slice claims; memory-impact: durable."
    protected_oracles:
      - path: cmd/kbcheck/plan_worktree_selftest_test.go
        role: "end-to-end multi-plan lifecycle oracle"
        sha256: "50943d314434e7b87f8bc3b32e408c5b0bede35d5e08cd43ef272b4fec8fa451"
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
- One manifest group owns one worktree; its slices mutate that worktree
  shared-serial. Per-slice worktrees are not part of this workflow.
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
| 2 | Block cross-manifest conflicts before mutation | slice-001 | tdd / functional-cli | no | done |
| 3 | Advance slice commits only on the owning plan-run branch | slice-002 | tdd / functional-cli | no | done |
| 4 | Keep default-branch delivery and dirty-WIP authority outside kb-work | slice-003 | tdd / functional-cli | no | done |
| 5 | Prove and release the multi-plan worktree lifecycle | slice-004 + external release gate | integration / full | no | done |

## Execution Gate

`plan-to-work` passed after the overlapping DDR scope was reviewed and contained
on local branch `codex/ddr-route-announcement-containment` at commit `3f1d916`.
Bootstrap execution remained shared-serial. Reviewed target drift was
reconciled into the source bundle, required global roots were synchronized, and
the final local release gate passed without touching the default branch.

## Scope-Verified Files

Exact final feature diff against containment base `3f1d916`:

- `.github/skills/kb-complete/SKILL.md`
- `.github/skills/kb-configure/SKILL.md`
- `.github/skills/kb-gate/scripts/check_gate_ledger.py`
- `.github/skills/kb-land/SKILL.md`
- `.github/skills/kb-plan/SKILL.md`
- `.github/skills/kb-ship/SKILL.md`
- `.github/skills/kb-start/SKILL.md`
- `.github/skills/kb-work/SKILL.md`
- `.github/skills/kb-work/references/execution-prompt.md`
- `.github/skills/kb-work/references/worktree-isolation.md`
- `README.md`
- `cmd/kbcheck/checks.go`
- `cmd/kbcheck/checks_test.go`
- `cmd/kbcheck/cross_manifest_scheduler_test.go`
- `cmd/kbcheck/delivery_boundary_test.go`
- `cmd/kbcheck/main.go`
- `cmd/kbcheck/manifest_contract.go`
- `cmd/kbcheck/manifest_contract_test.go`
- `cmd/kbcheck/plan_run_integration_test.go`
- `cmd/kbcheck/plan_run_scheduler.go`
- `cmd/kbcheck/plan_run_workspace.go`
- `cmd/kbcheck/plan_run_workspace_test.go`
- `cmd/kbcheck/plan_worktree_selftest.go`
- `cmd/kbcheck/plan_worktree_selftest_test.go`
- `cmd/kbcheck/skill_repo_contract_test.go`
- `cmd/kbcheck/slice_lease.go`
- `cmd/kbcheck/slice_lease_test.go`
- `cmd/kbcheck/swarm.go`
- `cmd/kbcheck/swarm_test.go`
- `cmd/kbcheck/worktree_isolation.go`
- `cmd/kbcheck/worktree_isolation_test.go`
- `config/skill-quality.json`
- `docs/context/PROJECT.md`
- `docs/context/architecture/kb-workflow.md`
- `docs/context/eval-map.md`
- `docs/context/kb/instincts/project.yaml`
- `docs/context/kb/instincts/scoped/kbcheck-proof-spine.yaml`
- `docs/context/kb/instincts/scoped/model-routing/evaluation.yaml`
- `docs/context/kb/instincts/scoped/model-routing/orchestration.yaml`
- `docs/context/kb/instincts/scoped/skill-bundle/provider-hygiene.yaml`
- `docs/context/kb/instincts/scoped/workflow/plan-execution.yaml`
- `docs/context/kb/kb-completions.txt`
- `docs/context/memory-maintenance.md`
- `docs/context/operations/testing.md`
- `docs/handoffs/done/2026-07-26-plan-run-worktree-isolation.md`
- `docs/plans/2026-07-19-003-tool-worktree-isolation-plan.md`
- `docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md`
- `docs/plans/2026-07-26-011-tool-plan-run-workspace-plan.md`
- `docs/plans/2026-07-26-012-tool-cross-manifest-scheduler-plan.md`
- `docs/plans/2026-07-26-013-tool-plan-run-integration-plan.md`
- `docs/plans/2026-07-26-014-tool-delivery-boundary-plan.md`
- `docs/plans/2026-07-26-015-eval-plan-worktree-lifecycle-plan.md`
- `docs/plans/2026-07-26-plan-run-worktree-context/slice-001.json`
- `docs/plans/2026-07-26-plan-run-worktree-context/slice-002.json`
- `docs/plans/2026-07-26-plan-run-worktree-context/slice-003.json`
- `docs/plans/2026-07-26-plan-run-worktree-context/slice-004.json`
- `docs/plans/2026-07-26-plan-run-worktree-context/slice-005.json`
- `docs/solutions/workflow-issues/plan-run-worktree-isolation-2026-07-26.md`
- `todo-done.md`
