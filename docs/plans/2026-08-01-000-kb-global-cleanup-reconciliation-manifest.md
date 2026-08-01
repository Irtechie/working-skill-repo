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
source_requirements_sha256: edc3f8a084547dcb35f664f3ca1fc77c90cee1048c6d965128eccb8633f997de
pre_slice_review:
  status: passed
  source: docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md
  source_sha256: edc3f8a084547dcb35f664f3ca1fc77c90cee1048c6d965128eccb8633f997de
  mode: requirements-wide
  review_id: global-cleanup-reconciliation-requirements-edc3f8a08454
  reviewed_at: "2026-08-01T17:20:00Z"
  review_artifact: docs/results/document-reviews/global-cleanup-reconciliation-requirements-edc3f8a08454.json
  review_artifact_sha256: 7af9ad1feb938cd5742964e6841e4c21f46873bc59f67bfbcbb7f5057a7302a5
  persona_evidence_json: '{"security-lens-reviewer":"security-risk: authoritative semantic-writer claims, scoped caller authority, stale-worker fencing, sole-path credentials, idempotency, rollback, and gateway commit atomicity define a new privileged trust boundary."}'
  selected_personas_json: '["security-lens-reviewer"]'
  completed_personas_json: '["security-lens-reviewer"]'
  failed_personas_json: '[]'
  findings_resolved: 10
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
  branch: codex/the-worktrees-filed-a-grievance
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
      - "The amended trust contract preserves disjoint DAG/worktree parallelism while requiring authoritative CAS claims, monotonically fenced generations, scoped caller authority, endpoint high-water validation, idempotency, rollback detection, and sole-path credential proof for protected semantic writers."
    proof:
      - docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md
      - docs/results/document-reviews/global-cleanup-reconciliation-requirements-edc3f8a08454.json
      - docs/context/research/2026-08-01-agent-dag-concurrency-and-fencing.md
      - cmd/kbcheck/terminal_cleanup.go
    blockers: []
    passed_at: "2026-08-01T17:22:00Z"
    allowed_next_action: "kb-plan docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Three vertical slices cover read-only portfolio convergence, checked apply/verify, and lifecycle/distribution plus fenced semantic-writer integration."
      - "Every requirement maps to a slice and every destructive path retains existing terminal-cleanup predicates or fails closed."
      - "The dependency graph is acyclic, preserves proven-disjoint parallelism, and serializes shared Git, installer, skill, documentation, and semantic-writer resources."
      - "Local plan-run commits are explicitly authorized; publishing remains unauthorized."
    proof:
      - docs/plans/2026-08-01-001-global-reconciler-inventory-plan.md
      - docs/plans/2026-08-01-002-global-reconciler-apply-plan.md
      - docs/plans/2026-08-01-003-global-reconciler-lifecycle-plan.md
      - docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md
    blockers: []
    passed_at: "2026-08-01T17:22:00Z"
    allowed_next_action: "kb-work docs/plans/2026-08-01-000-kb-global-cleanup-reconciliation-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "A standalone kbreconcile binary inventories and plans a plain non-KB Git repository without repo-local kbcheck files."
      - "The mixed 20-artifact oracle preserves active, protected, dirty, ignored, post-cutoff, credential/model/learning/live, unique, ambiguous, and unproven work."
      - "Routine exact cases produce zero prompts; ambiguity is grouped into one bounded decision packet and excess ambiguity defaults to quarantine."
      - "The versioned predicate manifest, deterministic confidence, and per-run/per-repository risk budget fail closed without enabling mutation."
    proof:
      - docs/results/proofs/global-cleanup-reconcile-20260801-slice-001.md
      - internal/reconcile/reconcile_test.go
      - cmd/kbreconcile/main_test.go
      - config/reconcile-predicates.json
    blockers: []
    passed_at: "2026-08-01T17:00:01Z"
    allowed_next_action: "kb-work docs/plans/2026-08-01-000-kb-global-cleanup-reconciliation-manifest.md"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Apply and verify accept only an unexpired cutoff-bound plan/receipt whose schema, policy, target identities, fingerprints, and mandatory predicates still match fresh evidence acquired under the compatible repository lock."
      - "Only allowlisted local non-force worktree retirement and exact-SHA local-ref CAS are available; every protected external action remains unavailable."
      - "Delivery, physical cleanup, ref retirement, and session-record states remain independent, with idempotent exact-empty residual repair and fail-closed preservation on nonempty or identity-drift residuals."
      - "The shared global predicate policy matches or exceeds the established terminal-cleanup safety corpus."
    proof:
      - docs/results/proofs/global-cleanup-reconcile-20260801-slice-002.md
      - internal/reconcile/apply_test.go
      - cmd/kbreconcile/main_test.go
      - cmd/kbcheck/reconcile_contract_test.go
      - cmd/kbcheck/terminal_cleanup_test.go
    blockers: []
    passed_at: "2026-08-01T18:11:44Z"
    allowed_next_action: "kb-work docs/plans/2026-08-01-000-kb-global-cleanup-reconciliation-manifest.md"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Canonical provider/tenant/account resource keys, authoritative CAS claims, monotonic generations, controller epochs, scoped verifier capability, endpoint high-water fencing, and durable idempotency fail closed under stale, forged, rollback, outage, and ambiguous-retry scenarios."
      - "Distinct semantic resources run concurrently while alias-equivalent or identical resources serialize; local queue state reports that it is not global authority and counts distinct active owners."
      - "Lifecycle registration preserves local-durable, awaiting-review, and delivery-integrated separately from physical cleanup, ref retirement, and host session retirement."
      - "The optional kbreconcile installer is checksum-managed, downgrade-aware, drift-safe, and skill-only compatible while reporting the absence of signed privileged provenance and live adapters."
    proof:
      - docs/results/proofs/global-cleanup-reconcile-20260801-slice-003.md
      - internal/reconcile/claim_test.go
      - cmd/kbreconcile/main_test.go
      - cmd/kbcheck/semantic_claim_contract_test.go
      - bin/kb-install.test.mjs
      - cmd/kbcheck/delivery_chain_contract_test.go
    blockers: []
    passed_at: "2026-08-01T19:19:51Z"
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
    status: done
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: internal/reconcile/reconcile_test.go
        role: "mixed-portfolio preservation, compact decision packet, policy parity, and no-KB inventory oracle"
        sha256: "ba84b73ddd5e04607a94f118193cab03824c30e0b363fe1aeb044393b928ac12"
        update_policy: "requires an explicit slice-plan amendment"
      - path: cmd/kbreconcile/main_test.go
        role: "standalone binary, stable JSON plan, fail-closed CLI, and plain-repository oracle"
        sha256: "181eda229d05f27bef109ea56a57a7beec23f876f05f2f3e24186b0cdb4f431c"
        update_policy: "requires an explicit slice-plan amendment"
    notes: "DDR route=current orchestrator; exact slice proof PASS; standalone no-KB binary dry-run PASS; gofmt/go vet/go build/git diff --check PASS; qa-browser skipped no UI-reachable behavior; forecast implementation=8 actual implementation=8 discovered lifecycle/proof=3 unused=0 unexplained=0; policy-sha256=57a11f5e79164efde34713ededfbed972eaab3e0053f5b5d8b6c0bc70d5d0c16; memory-impact=durable reconciler architecture, context refresh deferred to slice-003 lifecycle integration; aggregate core/local-release intentionally not run; slice-002 explicitly rebinds cmd/kbreconcile/main_test.go from 04e3576e383cd841750fb832bf11260864f23e660d896458b76f4bd433f7cc4d to f0fcbc98de2ba78741592178d13723bbc388b1ea2b1f0f1194fac0f0dbf3ca74 solely to add apply/verify CLI fail-closed and deterministic-JSON coverage; slice-003 explicitly rebinds it from f0fcbc98de2ba78741592178d13723bbc388b1ea2b1f0f1194fac0f0dbf3ca74 to 181eda229d05f27bef109ea56a57a7beec23f876f05f2f3e24186b0cdb4f431c solely to add stable fail-closed claim-capability and reference-conformance JSON coverage."
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
    model_requirements: ["Git worktree internals", "compare-and-swap refs", "remote containment", "cross-platform locks", "protected action authority downgrade"]
    escalation_triggers: ["force would be required", "remote authority is unresolved", "plan identity drift", "terminal-cleanup parity gap", "protected external action lacks fenced authority"]
    workspace_mode: shared-serial
    conflict_domains: ["go:reconcile-core", "go:terminal-cleanup", "git:worktrees", "git:refs", "state:work-queue"]
    proof_check:
      kind: command_exit
      command: "go test ./internal/reconcile ./cmd/kbreconcile ./cmd/kbcheck -run 'Apply|Verify|Reconcile|TerminalCleanup' -count=1"
      expect: 0
    status: done
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: cmd/kbcheck/terminal_cleanup_test.go
        role: "existing current, primary, dirty, ignored, remote, and exact-SHA CAS preservation oracle"
        sha256: "6779c3904a3aa087bbae9b2e2f7f6271d947665338dd55f66d1000c7631d922d"
        update_policy: "preserved unchanged"
      - path: internal/reconcile/apply_test.go
        role: "fresh lock/evidence, mutation gates, partial recovery, and idempotency oracle"
        sha256: "4598510ed34823993c5baaf0c7cea608de5718b87f4cf796ad337e0f1b6ab0dd"
        update_policy: "requires an explicit slice-plan amendment"
    notes: "DDR route=current orchestrator; exact narrow proof PASS; complete touched-package tests PASS; gofmt/go vet/go build/git diff --check PASS; destructive behavior exercised only in disposable test repositories/worktrees; protected external adapters unavailable; forecast implementation=7 actual forecast=6 discovered implementation=2 unused forecast=1 lifecycle/proof=2 unexplained=0; terminal_cleanup_test.go preserved byte-for-byte and its shared-policy obligation moved to reconcile_contract_test.go; memory-impact=durable checked reconciliation and shared terminal-safety policy, context refresh deferred to slice-003 lifecycle integration; aggregate core/local-release intentionally not run."
  - id: slice-003
    title: "Make KB lifecycle and global install converge automatically"
    path: docs/plans/2026-08-01-003-global-reconciler-lifecycle-plan.md
    blockers: [slice-001, slice-002]
    verification: tdd
    test_level: integration
    functional_risk: broad
    execution_class: cli
    model_tier: large
    model_tier_reason: "Connects checked convergence to queue lifecycle, globally fenced semantic-writer authority, managed binary distribution, skills, and workflow documentation."
    model_requirements: ["cross-skill lifecycle design", "managed binary installation", "queue and authoritative CAS", "fencing and idempotency", "documentation synchronization"]
    escalation_triggers: ["install becomes mandatory", "open PR consumes active WIP", "delivery and cleanup states collapse", "global drift is newer", "claim adapter cannot prove atomic high-water fencing or sole-path enforcement"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:kb-start", "skill:kb-complete", "skill:kb-finalize", "installer:managed-binaries", "docs:workflow", "authority:semantic-writers"]
    proof_check:
      kind: command_exit
      command: "go test ./... && node --test ./bin/kb-install.test.mjs && go run ./cmd/kbcheck skill-lint --root . && go test ./internal/reconcile ./cmd/kbreconcile ./cmd/kbcheck -run 'SemanticClaim|Fence|Idempot|DeliveryChain' -count=1"
      expect: 0
    status: done
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: bin/kb-install.test.mjs
        role: "optional checksum-managed reconciler install, fallback, downgrade, and drift-safe lifecycle oracle"
        sha256: "4df2a3bec9da0458f7bfbe42820078c0366a8a9f249b39e77b48af29d7c4cd8b"
        update_policy: "requires an explicit slice-plan amendment"
      - path: cmd/kbcheck/delivery_chain_contract_test.go
        role: "lifecycle registration and independent delivery/cleanup/ref/session authority oracle"
        sha256: "91db4ace0ad7a448a0a314d708902e05818b61f4b573ce3e5d69f450e56ed85b"
        update_policy: "requires an explicit slice-plan amendment"
      - path: internal/reconcile/claim_test.go
        role: "canonical claim, fencing, authorization, idempotency, recovery, concurrency, rollback, and bypass-denial oracle"
        sha256: "b5ccc6e7adf3349ad46b31bc71dd140c811974492355865165bb2f66de1aadd0"
        update_policy: "requires an explicit slice-plan amendment"
    notes: "DDR route=current orchestrator; planned proof PASS; gofmt/go test ./.../node installer tests/skill-lint/targeted conformance/go vet/go build/git diff --check PASS; PowerShell queue AST parse PASS and no pre-existing queue behavior suite was present; protected slice-001 reconcile_test.go and slice-002 apply_test.go/terminal_cleanup_test.go hashes preserved; forecast implementation=18 actual implementation=22 discovered=4 (internal/reconcile/policy.go, internal/reconcile/reconcile.go, internal/reconcile/apply.go, cmd/kbcheck/terminal_cleanup.go) unused=0 unexplained=0; plan-run lease expanded for discovered paths; first advance failed closed on the original forecast-only slice lease, which has no expansion action, so the official lease was released and immediately reacquired with all 24 exact committed paths before retry; initial lifecycle skill copies were identical across repo/Copilot/agents/Codex, final global copies remain mutually identical and intentionally unsynchronized; limitations=no live claim/provider adapter, no signed kbreconcile provenance, local queue/lease is not global authority, protected mutation unavailable; memory-impact=durable architecture/project/operations route maps refreshed; aggregate kbcheck core/local-release intentionally deferred to parent finalization."
---

# Global Cleanup Reconciliation

The manifest delivers one fail-closed global reconciliation product in three
observable increments. The current session coordinates serial execution because
the slices share the Go core, Git safety contract, installer, KB lifecycle, and
semantic-writer authority surfaces. Runtime DAG nodes remain parallel only when
their declared semantic resources are proven disjoint.
