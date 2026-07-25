---
type: kb-manifest
kb_id: kb-2026-07-19-evidence-graph-routing
brainstorm_path: docs/brainstorms/2026-07-19-evidence-graph-routing-requirements.md
created: 2026-07-19
status: done
workflow_shape: pipeline-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "Proves the provider-neutral routing, multi-session isolation, skill sync, and repo release contract together."
model_tier_contract:
  allowed: [small, medium, large]
  default: large
model_selection_contract:
  timing: work-time
  catalog: host-native-plus-user
  fallback: same-tier-then-higher-then-current
  automatic_downgrade: false
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
  allowed_modes: [shared-serial, worktree-required]
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-brainstorm
    status: passed
    required_evidence:
      - "requirements path exists"
      - "Question Gate classification exists"
      - "Resolve Before Planning is empty"
      - "no unresolved ask-now or research-first items remain"
      - "safe assumptions, deferred details, and parked scope are recorded"
    proof:
      - docs/brainstorms/2026-07-19-evidence-graph-routing-requirements.md
      - "Question Gate classifies ask-now, research-first, safe-assumption, defer-to-planning, and parked items"
      - "Resolve Before Planning: None"
      - "Safe assumptions and deferred details are recorded in the requirements artifact"
      - "No unresolved ask-now or research-first items remain before planning"
    blockers: []
    passed_at: "2026-07-19T14:30:54Z"
    allowed_next_action: "kb-plan docs/brainstorms/2026-07-19-evidence-graph-routing-requirements.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest and eight slice plans exist"
      - "DAG has no missing blockers or cycles"
      - "every slice has rationale, acceptance criteria, expected_files, proof, test level, functional risk, and model tier"
      - "every slice has a bounded context packet"
      - "temporary single-coordinator rule protects execution until atomic claims and worktree isolation land"
      - "provider adapters remain optional and file-native fallback remains authoritative"
      - "objective done check exists"
    proof:
      - docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md
      - docs/plans/2026-07-19-001-tool-impact-packet-contract-plan.md
      - docs/plans/2026-07-19-002-tool-atomic-slice-lease-plan.md
      - docs/plans/2026-07-19-003-tool-worktree-isolation-plan.md
      - docs/plans/2026-07-19-004-integration-exact-symbol-adapter-plan.md
      - docs/plans/2026-07-19-005-integration-structural-flow-recipes-plan.md
      - docs/plans/2026-07-19-006-skill-impact-lifecycle-plan.md
      - docs/plans/2026-07-19-007-eval-routing-correctness-plan.md
      - docs/plans/2026-07-19-008-doc-release-sync-plan.md
      - docs/plans/2026-07-19-evidence-graph-routing-context/slice-001.json
      - docs/plans/2026-07-19-evidence-graph-routing-context/slice-002.json
      - docs/plans/2026-07-19-evidence-graph-routing-context/slice-003.json
      - docs/plans/2026-07-19-evidence-graph-routing-context/slice-004.json
      - docs/plans/2026-07-19-evidence-graph-routing-context/slice-005.json
      - docs/plans/2026-07-19-evidence-graph-routing-context/slice-006.json
      - docs/plans/2026-07-19-evidence-graph-routing-context/slice-007.json
      - docs/plans/2026-07-19-evidence-graph-routing-context/slice-008.json
    blockers: []
    passed_at: "2026-07-19T14:30:54Z"
    allowed_next_action: "kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "protected oracle hash recorded"
      - "deterministic proof passed"
      - "functional CLI proof passed"
      - "memory impact classified"
    proof:
      - internal/graphrouting/contract.go
      - internal/graphrouting/contract_test.go
      - cmd/kbcheck/graph_route.go
      - cmd/kbcheck/graph_route_test.go
      - cmd/kbcheck/main.go
      - config/graph-route.schema.json
      - evals/graph-routing/impact-packet-valid.json
      - evals/graph-routing/impact-packet-stale.json
      - ".github/skills/kb-map/SKILL.md"
      - ".github/skills/kb-map/references/graph-routing.md"
      - "proof command graph route impact packet tests passed"
      - "proof command valid packet CLI passed"
      - "proof command stale packet CLI failed closed with stale fallback issue"
      - "protected oracle sha256=40CCC36E0F19494AA08DACC372140E81FDFC086DA051F4378D367C55836FBE66"
      - "scope-check: forecast=9 changed=12 discovered=3 unexplained=0"
      - "memory-impact: durable; areas=kb-map graph-routing contract, kbcheck CLI; refresh=pending"
    blockers: []
    passed_at: "2026-07-19T17:28:00Z"
    allowed_next_action: "kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "protected oracle hash recorded"
      - "deterministic proof passed"
      - "functional CLI proof passed"
      - "memory impact classified"
    proof:
      - cmd/kbcheck/slice_lease.go
      - cmd/kbcheck/slice_lease_test.go
      - cmd/kbcheck/swarm.go
      - cmd/kbcheck/swarm_test.go
      - cmd/kbcheck/main.go
      - cmd/kbcheck/checks.go
      - cmd/kbcheck/checks_test.go
      - cmd/kbcheck/skill_repo_contract_test.go
      - ".github/skills/kb-work/SKILL.md"
      - ".github/skills/kb-work/references/execution-prompt.md"
      - "proof command kbcheck SliceLease ScopeLease tests passed"
      - "proof command kbcheck slice-lease-selftest passed"
      - "protected oracle sha256=9371669B79F75CF9F43D5A5CB0949933EEB25458E9B63BCC5E6939154B64C8A1"
      - "scope-check: forecast=8 changed=11 discovered=3 unexplained=0"
      - "memory-impact: durable; areas=kb-work atomic slice lease, kbcheck CLI, core release checks; refresh=pending"
    blockers: []
    passed_at: "2026-07-19T19:05:00Z"
    allowed_next_action: "kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "protected oracle hash recorded"
      - "deterministic proof passed"
      - "functional CLI proof passed"
      - "regression snapshot captured"
      - "memory impact classified"
    proof:
      - cmd/kbcheck/worktree_isolation.go
      - cmd/kbcheck/worktree_isolation_test.go
      - cmd/kbcheck/main.go
      - cmd/kbcheck/swarm.go
      - cmd/kbcheck/manifest_contract.go
      - cmd/kbcheck/manifest_contract_test.go
      - ".github/skills/kb-plan/SKILL.md"
      - ".github/skills/kb-work/SKILL.md"
      - ".github/skills/kb-work/references/worktree-isolation.md"
      - ".github/skills/kb-regression-snapshot/SKILL.md"
      - docs/context/architecture/kb-workflow.md
      - ".kb/snapshots/evidence-graph-routing-slice-003.json"
      - "proof command kbcheck Worktree IntegrationOwner tests passed"
      - "proof command kbcheck manifest contract passed"
      - "protected oracle sha256=F52A314AC9E21D6EEAB78728189C7953FFB6E89FA79351B4E798F8DF30542D30"
      - "scope-check: forecast=10 changed=16 discovered=6 unexplained=0"
      - "memory-impact: durable; areas=kb-work worktree isolation, kb-plan workspace contract, kb-regression-snapshot canonical root, kbcheck CLI; refresh=pending"
    blockers: []
    passed_at: "2026-07-19T20:25:00Z"
    allowed_next_action: "kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "protected oracle hash recorded"
      - "deterministic proof passed"
      - "functional CLI proof passed"
      - "memory impact classified"
    proof:
      - internal/graphrouting/provider.go
      - internal/graphrouting/scip.go
      - internal/graphrouting/scip_test.go
      - evals/graph-routing/exact-symbol-index.json
      - ".github/skills/kb-map/references/graph-routing.md"
      - docs/context/operations/testing.md
      - "proof command graphrouting exact symbol SCIP fallback tests passed"
      - "protected oracle sha256=8D3296B66A3E045803E0B0C337611FAF88EADFB351F374DD81A794B8CEA2729D"
      - "scope-check: forecast=6 changed=6 discovered=0 unexplained=0"
      - "workspace note: serialized in canonical checkout because the uncommitted slice 001-003 baseline is not available from a fresh worktree"
      - "memory-impact: durable; areas=kb-map exact-symbol precedence, graphrouting optional SCIP adapter, testing proof docs; refresh=pending"
    blockers: []
    passed_at: "2026-07-19T20:34:58-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"
  - gate_id: slice-slice-005-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "protected oracle hash recorded"
      - "deterministic proof passed"
      - "functional CLI proof passed"
      - "memory impact classified"
    proof:
      - internal/graphrouting/recipe.go
      - internal/graphrouting/recipe_test.go
      - internal/graphrouting/graphify.go
      - internal/graphrouting/graphify_test.go
      - evals/graph-routing/traversal-recipes.json
      - cmd/kbcheck/graph_route.go
      - cmd/kbcheck/graph_route_test.go
      - ".github/skills/kb-map/SKILL.md"
      - ".github/skills/kb-map/references/graph-routing.md"
      - "proof command traversal recipe flow budget graphify tests passed"
      - "protected oracle sha256=02F0DB4ED6188F1E8C8E048D258E3BD4C089B6436AEB47554BFEE8988B325C6A"
      - "scope-check: forecast=8 changed=9 discovered=1 unexplained=0"
      - "workspace note: serialized in canonical checkout after slice 004 lease release"
      - "memory-impact: durable; areas=kb-map traversal recipes, graphrouting Graphify adapter, graph-route CLI annotations; refresh=pending"
    blockers: []
    passed_at: "2026-07-19T20:49:22-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"
  - gate_id: slice-slice-006-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "protected oracle hash recorded"
      - "deterministic proof passed"
      - "functional CLI proof passed"
      - "memory impact classified"
    proof:
      - ".github/skills/kb-map/SKILL.md"
      - ".github/skills/kb-plan/SKILL.md"
      - ".github/skills/kb-work/SKILL.md"
      - ".github/skills/kb-review/SKILL.md"
      - ".github/skills/kb-review/references/diff-scope.md"
      - cmd/kbcheck/manifest_contract.go
      - cmd/kbcheck/manifest_contract_test.go
      - cmd/kbcheck/graph_routing_lifecycle.go
      - cmd/kbcheck/main.go
      - cmd/kbcheck/checks.go
      - cmd/kbcheck/swarm.go
      - evals/skill-eval/selftest/pass-evidence-graph-routing.json
      - docs/context/architecture/kb-workflow.md
      - "proof command graph routing lifecycle selftest passed"
      - "proof command manifest contract focused tests passed"
      - "protected oracle sha256=80B306772B62E1035BD39DB36D4239A498E37198830F39AAFCAE63A872ACF786"
      - "scope-check: forecast=10 changed=13 discovered=3 unexplained=0"
      - "memory-impact: durable; areas=kb-map kb-plan kb-work kb-review graph routing lifecycle, kbcheck manifest impact contract; refresh=pending"
    blockers: []
    passed_at: "2026-07-19T21:00:39-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"
  - gate_id: slice-slice-007-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "protected oracle hash recorded"
      - "deterministic proof passed"
      - "functional CLI proof passed"
      - "memory impact classified"
    proof:
      - evals/graph-routing/README.md
      - evals/graph-routing/expected-results.json
      - evals/graph-routing/fixtures/symbol-impact.json
      - evals/graph-routing/fixtures/stale-index.json
      - evals/graph-routing/fixtures/multisession-race.json
      - cmd/kbcheck/graph_routing_eval.go
      - cmd/kbcheck/graph_routing_eval_test.go
      - cmd/kbcheck/checks.go
      - cmd/kbcheck/main.go
      - docs/context/eval-map.md
      - "proof command graph routing eval require ready passed"
      - "proof command graph routing eval negative tests passed"
      - "protected oracle sha256=0A01D0543B363E41EF801A2E823AC908286A6F674A10EDF52DE426F719519CBF"
      - "scope-check: forecast=9 changed=10 discovered=1 unexplained=0"
      - "memory-impact: durable; areas=graph routing eval corpus, kbcheck readiness gate, eval map; refresh=pending"
    blockers: []
    passed_at: "2026-07-19T21:12:20-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"
  - gate_id: slice-slice-008-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "implementation finished"
      - "scope check passed"
      - "documentation refreshed"
      - "required skill sync drift resolved"
      - "deterministic proof passed"
      - "release proof passed"
      - "memory impact classified"
    proof:
      - README.md
      - docs/context/PROJECT.md
      - docs/context/architecture/kb-workflow.md
      - docs/context/operations/testing.md
      - docs/context/eval-map.md
      - docs/context/memory-maintenance.md
      - docs/plans/2026-07-19-008-doc-release-sync-plan.md
      - config/skill-quality.json
      - todo.md
      - todo-done.md
      - "proof command graph routing lifecycle selftest passed"
      - "proof command graph routing eval require-ready passed"
      - "proof command skill sync report passed with 129 comparisons and 0 required issues"
      - "proof command core passed with 36 checks"
      - "proof command local-release passed"
      - "scope-check: forecast=9 changed=9 discovered=0 unexplained=0; skill-quality config inspected and reused without edit"
      - "memory-impact: durable; areas=README project map testing eval map memory maintenance, release sync state; refresh=complete"
    blockers: []
    passed_at: "2026-07-19T22:27:46-04:00"
    allowed_next_action: "kb-complete docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"
slices:
  - id: slice-001
    title: "Return a provider-neutral evidence and impact packet"
    path: docs/plans/2026-07-19-001-tool-impact-packet-contract-plan.md
    blockers: []
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Defines a durable cross-skill contract whose wrong shape would couple every later adapter and lifecycle phase."
    model_requirements: ["cross-skill architecture judgment", "Go schema and CLI tests", "backward-compatible file-native fallback", "bounded source-evidence design"]
    escalation_triggers: ["the contract requires a mandatory provider or daemon", "existing manifest/context-packet contracts cannot represent the impact packet without migration", "freshness cannot be bound to canonical repo identity"]
    context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-001.json
    proof_check:
      kind: command_exit
      command: "go test ./internal/graphrouting ./cmd/kbcheck -run 'GraphRoute|ImpactPacket'"
      expect: 0
    hitl: false
    status: done
    workspace_mode: shared-serial
    conflict_domains: [file:internal/graphrouting, file:cmd/kbcheck/graph_route.go, skill:kb-map]
    shared_resources: [git:integration-owner]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Implement the smallest file-native packet path and its protected fixtures before adding external providers."
    human_action: ""
    can_continue_other_slices: false
    notes: "Temporary execution rule: one coordinator only until slices 002 and 003 are integrated. scope-forecast: loaded 9 expected files + convention Go tests. router-unavailable: binary-not-found; current execution reason=no-qualified-route; preflight fixed 2026-07-19 after cold Go cache completed and brainstorm-to-plan proof count was repaired. proof: go test ./internal/graphrouting ./cmd/kbcheck -run 'GraphRoute|ImpactPacket' -count=1 -v passed; CLI valid packet passed; stale packet failed closed. protected-oracle: evals/graph-routing/impact-packet-valid.json sha256=40CCC36E0F19494AA08DACC372140E81FDFC086DA051F4378D367C55836FBE66. scope-discovery: cmd/kbcheck/graph_route_test.go - convention CLI test for graph_route.go. scope-discovery: docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md - lifecycle status/gate update. scope-discovery: todo.md - board claim/status update. scope-check: forecast=9 changed=12 discovered=3 unexplained=0. memory-impact: durable; areas=kb-map graph-routing contract, kbcheck CLI; refresh=pending."
    protected_oracles:
      - path: evals/graph-routing/impact-packet-valid.json
        role: "provider-neutral accepted packet"
        sha256: "40CCC36E0F19494AA08DACC372140E81FDFC086DA051F4378D367C55836FBE66"
        update_policy: "requires explicit plan amendment"
  - id: slice-002
    title: "Atomically claim and release slices across local sessions"
    path: docs/plans/2026-07-19-002-tool-atomic-slice-lease-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Cross-process ownership, stale recovery, path safety, and Windows/Unix atomicity are concurrency-critical and can corrupt workflow state if subtly wrong."
    model_requirements: ["cross-process concurrency tests", "Windows and Unix filesystem semantics", "canonical Git common-dir identity", "fail-closed ownership tokens"]
    escalation_triggers: ["atomicity requires a daemon or network service", "same-slice double claims still pass", "lease recovery can delete or steal a live owner", "separate clones are accidentally treated as coordinated"]
    context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-002.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'SliceLease|ScopeLease'"
      expect: 0
    hitl: false
    status: done
    workspace_mode: shared-serial
    conflict_domains: [file:cmd/kbcheck/slice_lease.go, file:cmd/kbcheck/swarm.go, skill:kb-work]
    shared_resources: [git:integration-owner, git:slice-lease]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Proceed to slice 003 worktree isolation and serialized integration."
    human_action: ""
    can_continue_other_slices: false
    notes: "This closes the current todo.md and passive-ledger TOCTOU hole. scope-forecast: loaded 8 expected files + convention Go tests. router-unavailable: binary-not-found; current execution reason=no-qualified-route. pre-edit drift: kb-work execution prompt matched globals; kb-work SKILL.md repo copy was ahead and merged useful global-only AMR optional sentence before atomic lease edits. proof: go test ./cmd/kbcheck -run 'SliceLease|ScopeLease' -count=1 -v passed; go run ./cmd/kbcheck slice-lease-selftest passed. protected-oracle: cmd/kbcheck/slice_lease_test.go sha256=9371669B79F75CF9F43D5A5CB0949933EEB25458E9B63BCC5E6939154B64C8A1. scope-discovery: cmd/kbcheck/checks_test.go - native check registry oracle for added selftest. scope-discovery: cmd/kbcheck/skill_repo_contract_test.go - skill repo contract oracle for added selftest. scope-discovery: docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md - lifecycle status/gate update. scope-check: forecast=8 changed=11 discovered=3 unexplained=0. memory-impact: durable; areas=kb-work atomic slice lease, kbcheck CLI, core release checks; refresh=pending."
    protected_oracles:
      - path: cmd/kbcheck/slice_lease_test.go
        role: "cross-process ownership and stale-recovery oracle"
        sha256: "9371669B79F75CF9F43D5A5CB0949933EEB25458E9B63BCC5E6939154B64C8A1"
        update_policy: "requires explicit plan amendment"
  - id: slice-003
    title: "Isolate mutating slices in worktrees and serialize integration"
    path: docs/plans/2026-07-19-003-tool-worktree-isolation-plan.md
    blockers: [slice-002]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Creates and integrates Git branches/worktrees around dirty user state, so cleanup, base revision, ownership, and conflict behavior need high-authority safeguards."
    model_requirements: ["Git worktree lifecycle expertise", "non-destructive Windows path handling", "serialized integration and post-merge proof", "skill and CLI contract alignment"]
    escalation_triggers: ["the adapter needs force removal or reset", "canonical lifecycle files can be edited by workers without merge conflict protection", "dirty checkout preservation cannot be proved", "integration cannot be replayed deterministically"]
    context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-003.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'Worktree|IntegrationOwner'"
      expect: 0
    hitl: false
    status: done
    workspace_mode: shared-serial
    conflict_domains: [file:cmd/kbcheck/worktree_isolation.go, file:cmd/kbcheck/manifest_contract.go, skill:kb-plan, skill:kb-work, skill:kb-regression-snapshot]
    shared_resources: [git:integration-owner, git:worktree, git:slice-lease]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Proceed to slices 004 and 005 through the new lease/worktree path when safe."
    human_action: ""
    can_continue_other_slices: false
    notes: "After this slice integrates, slices 004 and 005 may execute concurrently through the new lease/worktree path. scope-forecast: loaded 10 expected files + convention Go tests. router-unavailable: binary-not-found; current execution reason=no-qualified-route. pre-edit drift: kb-regression-snapshot matched globals; kb-plan globals had useful model_tier_reason contract text merged into repo copy; kb-work repo copy was ahead and preserved atomic lease/current-reason routing text. pre-slice snapshot-verify: PASS 15/15 snapshots from canonical root E:/Dev/Tools/working-skill-repo after E:/working-skill-repo alias failed a GHCP evidence symlink containment check. proof: go test ./cmd/kbcheck -run 'Worktree|IntegrationOwner' -count=1 -v passed; manifest contract passed. protected-oracle: cmd/kbcheck/worktree_isolation_test.go sha256=F52A314AC9E21D6EEAB78728189C7953FFB6E89FA79351B4E798F8DF30542D30. snapshot: .kb/snapshots/evidence-graph-routing-slice-003.json captured using file checksums after non-verbose embedded worktree proof was too slow/flaky under the snapshot runner. scope-discovery: cmd/kbcheck/swarm.go - manifest parser needed workspace/model selection fields. scope-discovery: docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md - workspace isolation contract and lifecycle status/gate update. scope-discovery: docs/plans/2026-07-19-003-tool-worktree-isolation-plan.md - protected oracle/proof status update. scope-discovery: todo.md - board status update. scope-discovery: docs/handoffs/active/2026-07-19-evidence-graph-routing.md - restart state update. scope-discovery: .kb/snapshots/evidence-graph-routing-slice-003.json - generated regression snapshot artifact. scope-check: forecast=10 changed=16 discovered=6 unexplained=0. memory-impact: durable; areas=kb-work worktree isolation, kb-plan workspace contract, kb-regression-snapshot canonical root, kbcheck CLI; refresh=pending."
    protected_oracles:
      - path: cmd/kbcheck/worktree_isolation_test.go
        role: "dirty checkout, branch isolation, integration, and cleanup oracle"
        sha256: "F52A314AC9E21D6EEAB78728189C7953FFB6E89FA79351B4E798F8DF30542D30"
        update_policy: "requires explicit plan amendment"
  - id: slice-004
    title: "Prefer exact-symbol evidence through an optional adapter"
    path: docs/plans/2026-07-19-004-integration-exact-symbol-adapter-plan.md
    blockers: [slice-003]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "The provider boundary is fixed by slice 001; this is bounded adapter and fixture work with objective fallback and provenance checks."
    model_requirements: ["SCIP/LSP symbol semantics", "fixture-driven adapter parsing", "optional-tool detection", "precise source-span provenance"]
    escalation_triggers: ["adapter setup would download tools or add a required daemon", "symbol identities cannot be made stable across supported fixture languages", "provider output bypasses source verification"]
    context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-004.json
    proof_check:
      kind: command_exit
      command: "go test ./internal/graphrouting -run 'ExactSymbol|SCIP|Fallback'"
      expect: 0
    hitl: false
    status: done
    workspace_mode: shared-serial
    conflict_domains: [file:internal/graphrouting, file:evals/graph-routing/exact-symbol-index.json, skill:kb-map]
    shared_resources: [git:integration-owner, graph:index]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Proceed to slice 005 structural and flow traversal recipes."
    human_action: ""
    can_continue_other_slices: true
    notes: "Completed with slice-004 lease generation 1. Worktree-required isolation was unavailable because the slice 001-003 baseline is uncommitted in the canonical checkout; execution used the allowed shared-serial fallback and no slice 005 mutation ran concurrently. proof: graphrouting exact symbol SCIP fallback tests passed. protected-oracle: evals/graph-routing/exact-symbol-index.json sha256=8D3296B66A3E045803E0B0C337611FAF88EADFB351F374DD81A794B8CEA2729D. scope-check: forecast=6 changed=6 discovered=0 unexplained=0. memory-impact: durable; areas=kb-map exact-symbol precedence, graphrouting optional SCIP adapter, testing proof docs; refresh=pending."
    protected_oracles:
      - path: evals/graph-routing/exact-symbol-index.json
        role: "definition/reference/implementation resolution fixture"
        sha256: "8D3296B66A3E045803E0B0C337611FAF88EADFB351F374DD81A794B8CEA2729D"
        update_policy: "requires explicit plan amendment"
  - id: slice-005
    title: "Add budgeted structural and flow traversal recipes"
    path: docs/plans/2026-07-19-005-integration-structural-flow-recipes-plan.md
    blockers: [slice-003]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Typed traversal, dynamic uncertainty, provider budgets, and flow over-approximation are architecture-sensitive and easy to make confidently incomplete."
    model_requirements: ["AST and code-property-graph concepts", "bounded path-query design", "Graphify/Tree-sitter compatibility", "uncertainty and cost controls"]
    escalation_triggers: ["a recipe requires whole-program flow by default", "LLM-inferred edges become authoritative", "dynamic/reflection limitations are hidden", "provider-specific data leaks into durable plan state"]
    context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-005.json
    proof_check:
      kind: command_exit
      command: "go test ./internal/graphrouting ./cmd/kbcheck -run 'TraversalRecipe|FlowBudget|Graphify'"
      expect: 0
    hitl: false
    status: done
    workspace_mode: shared-serial
    conflict_domains: [file:internal/graphrouting, file:evals/graph-routing/traversal-recipes.json, skill:kb-map]
    shared_resources: [git:integration-owner, graph:index]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Proceed to slice 006 KB lifecycle integration."
    human_action: ""
    can_continue_other_slices: true
    notes: "Completed with slice-005 lease generation 1 after slice 004 was complete and released. Execution used shared-serial fallback in the canonical checkout. proof: traversal recipe flow budget graphify tests passed, including graph-route CLI annotation rejection. protected-oracle: evals/graph-routing/traversal-recipes.json sha256=02F0DB4ED6188F1E8C8E048D258E3BD4C089B6436AEB47554BFEE8988B325C6A. scope-discovery: cmd/kbcheck/graph_route_test.go - CLI regression for Graphify exact-evidence rejection. scope-check: forecast=8 changed=9 discovered=1 unexplained=0. memory-impact: durable; areas=kb-map traversal recipes, graphrouting Graphify adapter, graph-route CLI annotations; refresh=pending."
    protected_oracles:
      - path: evals/graph-routing/traversal-recipes.json
        role: "intent-specific edge and budget fixture"
        sha256: "02F0DB4ED6188F1E8C8E048D258E3BD4C089B6436AEB47554BFEE8988B325C6A"
        update_policy: "requires explicit plan amendment"
  - id: slice-006
    title: "Carry impact and isolation contracts through KB lifecycle"
    path: docs/plans/2026-07-19-006-skill-impact-lifecycle-plan.md
    blockers: [slice-004, slice-005]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "Changes routing, planning, execution, and review semantics across multiple synchronized skills and must preserve existing proof and consent gates."
    model_requirements: ["cross-skill lifecycle architecture", "manifest compatibility", "impact reconciliation", "multi-session coordinator/worker separation"]
    escalation_triggers: ["existing manifests become unreadable", "workers can independently overwrite canonical lifecycle files", "expected_files becomes a hard allowlist", "graph evidence can replace source or functional proof"]
    context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-006.json
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck graph-routing-lifecycle-selftest"
      expect: 0
    hitl: false
    status: done
    workspace_mode: shared-serial
    conflict_domains: [skill:kb-map, skill:kb-plan, skill:kb-work, skill:kb-review, file:cmd/kbcheck]
    shared_resources: [git:integration-owner, lifecycle:manifest, lifecycle:todo]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Proceed to slice 007 routing correctness and concurrency evals."
    human_action: ""
    can_continue_other_slices: false
    notes: "Completed with slice-006 lease generation 1. Pre-edit drift: kb-review matched required global roots; kb-map/kb-plan/kb-work repo copies were ahead of global roots with prior slice work, no useful global-only drift found; Codex global copies missing. proof: graph routing lifecycle selftest passed; focused manifest contract tests passed. protected-oracle: evals/skill-eval/selftest/pass-evidence-graph-routing.json sha256=80B306772B62E1035BD39DB36D4239A498E37198830F39AAFCAE63A872ACF786. scope-discovery: cmd/kbcheck/main.go - CLI command registration. scope-discovery: cmd/kbcheck/checks.go - core check registration. scope-discovery: cmd/kbcheck/swarm.go - manifest impact packet fields. scope-check: forecast=10 changed=13 discovered=3 unexplained=0. memory-impact: durable; areas=kb-map kb-plan kb-work kb-review graph routing lifecycle, kbcheck manifest impact contract; refresh=pending."
    protected_oracles:
      - path: evals/skill-eval/selftest/pass-evidence-graph-routing.json
        role: "end-to-end KB lifecycle routing fixture"
        sha256: "80B306772B62E1035BD39DB36D4239A498E37198830F39AAFCAE63A872ACF786"
        update_policy: "requires explicit plan amendment"
  - id: slice-007
    title: "Falsify routing correctness and multi-session safety claims"
    path: docs/plans/2026-07-19-007-eval-routing-correctness-plan.md
    blockers: [slice-006]
    verification: functional
    test_level: functional-cli
    functional_risk: broad
    model_tier: medium
    model_tier_reason: "The implementation contract is settled; this slice is fixture-heavy falsification with explicit metrics and deterministic CLI proof."
    model_requirements: ["adversarial fixture design", "two-process race harness", "impact recall and false-positive scoring", "stale-index and dirty-worktree cases"]
    escalation_triggers: ["the harness cannot distinguish token savings from correctness", "race tests are nondeterministic", "provider absence changes required core results", "a claimed safety property lacks a negative fixture"]
    context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-007.json
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck graph-routing-eval --require-ready"
      expect: 0
    hitl: false
    status: done
    workspace_mode: shared-serial
    conflict_domains: [file:evals/graph-routing, file:cmd/kbcheck]
    shared_resources: [git:integration-owner, eval:graph-routing]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Proceed to slice 008 documentation, sync, and release proof."
    human_action: ""
    can_continue_other_slices: false
    notes: "Completed with slice-007 lease generation 1. proof: graph routing eval require-ready passed; focused graph routing eval tests passed. protected-oracle: evals/graph-routing/expected-results.json sha256=0A01D0543B363E41EF801A2E823AC908286A6F674A10EDF52DE426F719519CBF. scope-discovery: cmd/kbcheck/main.go - CLI command and require-ready flag registration. scope-check: forecast=9 changed=10 discovered=1 unexplained=0. memory-impact: durable; areas=graph routing eval corpus, kbcheck readiness gate, eval map; refresh=pending."
    protected_oracles:
      - path: evals/graph-routing/expected-results.json
        role: "routing and concurrency promotion oracle"
        sha256: "0A01D0543B363E41EF801A2E823AC908286A6F674A10EDF52DE426F719519CBF"
        update_policy: "requires explicit plan amendment"
  - id: slice-008
    title: "Document, synchronize, and release the optional routing contract"
    path: docs/plans/2026-07-19-008-doc-release-sync-plan.md
    blockers: [slice-007]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    model_tier: medium
    model_tier_reason: "The implementation is complete; this slice requires careful drift merge, cross-install hash proof, and release validation rather than new architecture."
    model_requirements: ["skill drift comparison", "documentation consistency", "hash verification", "core and local-release gates"]
    escalation_triggers: ["a global skill copy contains newer useful work", "required install hashes differ after sync", "core or local-release fails", "docs imply a mandatory provider or automatic worktree cleanup"]
    context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-008.json
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck local-release"
      expect: 0
    hitl: false
    status: done
    workspace_mode: shared-serial
    conflict_domains: [skill:kb-map, skill:kb-plan, skill:kb-work, skill:kb-review, docs:release]
    shared_resources: [git:integration-owner, sync:global-skills]
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Complete; all slices are done and local-release passed."
    human_action: ""
    can_continue_other_slices: false
    notes: "Completed with slice-008 lease generation 1. Reviewed required global drift for ce-review, kb-map, kb-plan, kb-regression-snapshot, kb-review, and kb-work; repo copies contained the approved graph-routing/workspace updates and no useful global-only work was found. Codex global missing targets repaired with doctor; Copilot/shared-agent stale copies repaired after reviewed-drift markers. proof: graph routing lifecycle selftest passed; graph routing eval require-ready passed; skill-sync-report passed with 129 comparisons and 0 required issues; core passed with 36 checks; local-release passed. scope-check: forecast=9 changed=9 discovered=0 unexplained=0; skill-quality config inspected and reused without edit. memory-impact: durable; areas=README project map testing eval map memory maintenance, release sync state; refresh=complete. ATV repositories were not sync, release, inspection, or gate targets."
    protected_oracles: []
---

# KB: Evidence-Backed Graph Routing And Multi-Session Isolation

## Origin

Brainstorm: `docs/brainstorms/2026-07-19-evidence-graph-routing-requirements.md`

## Outcome

`kb-map` returns compact, source-citable impact packets through a provider-
neutral contract; `kb-plan`, `kb-work`, and `kb-review` consume and reconcile
them; and local concurrent mutating sessions use atomic claims plus isolated
worktrees with one serialized integration owner.

## Direct Concurrency Answer

Yes: multiple sessions can duplicate or overwrite work today because the board
claim is not atomic and the existing scope-lease command only evaluates a
ledger supplied after the fact. Worktrees reduce filesystem collisions but do
not solve task ownership or mergeback. This manifest therefore builds atomic
claims before it enables automatic worktree concurrency.

## Worktree Ownership Decision

| Surface | Owns | Does not own | Reason |
|---|---|---|---|
| `kb-plan` | isolation intent, conflict domains, shared resources, integration dependencies | live paths, branches, session IDs, cleanup | plans are durable and portable; worktree state is live and machine-specific |
| `kb-work` | atomic claim, workspace choice, branch/worktree lifecycle, receipts, serialized integration, post-integration proof, safe cleanup | architecture intent | only execution time has current Git, dirty-state, and competing-session evidence |
| coordinator | canonical `todo.md`, manifest, handoff, integration order | slice implementation in every worker | prevents lifecycle-file divergence across worktrees |
| worker | isolated implementation, commit/diff, proof receipt, observed-write report | direct canonical merge or forced cleanup | makes worker results reviewable and integration recoverable |

## Workflow Shape

`pipeline-change` — schemas, Go validators/adapters, routing skills, work
isolation, evals, docs, and synchronized global installs change together.

## Slice Overview

| # | Slice | Blocked By | Verification | Test level | Reason | Status |
|---|---|---|---|---|---|---|
| 1 | Provider-neutral impact packet | - | integration | functional-cli | Stabilize semantics before selecting providers | done |
| 2 | Atomic slice claims | 1 | tdd | functional-cli | Close the actual multi-session race | done |
| 3 | Worktree isolation/integration | 2 | integration | functional-cli | Turn isolation wording into safe runtime behavior | done |
| 4 | Exact-symbol adapter | 3 | integration | functional-cli | Highest-precision relationship source first | done |
| 5 | Structural/flow recipes | 3 | integration | functional-cli | Cover behavior AST/call graphs miss, with budgets | done |
| 6 | KB lifecycle integration | 4, 5 | integration | functional-cli | Make map intelligence affect edits and review | done |
| 7 | Correctness/concurrency eval | 6 | functional | functional-cli | Prevent token-saving or safety claims without recall/race proof | done |
| 8 | Docs/sync/release | 7 | integration | functional-cli | Ship one consistent portable contract | done |

## Temporary Execution Safety

- Run slices 001-003 under one coordinator and one mutating session.
- Do not start this manifest concurrently in multiple sessions before slice 002
  atomic claims and slice 003 worktree integration are merged and proved.
- After slice 003, slices 004 and 005 are the first permitted parallel pair.
- If execution cannot use the new isolation path, serialize all mutating slices.

## Non-Negotiables

- File-native lookup remains functional without any optional provider.
- No required daemon, MCP server, vector database, CPG service, or auto-install.
- Exact/source evidence outranks heuristics; LLM-inferred edges are never exact.
- Graph/index state is namespaced by repository identity, revision, and
  worktree/dirty fingerprint.
- `expected_files` remains a forecast, not a safety oracle.
- Worktrees are isolation; atomic claims and serialized integration provide
  coordination.
- No forced worktree deletion, reset, clean, checkout, or overwrite.
- Separate clones/machines are not claimed to share the local lease.
- ATV repositories remain out of scope.

## Done

- All eight slices are done or explicitly skipped with reason.
- Routing correctness and concurrency claims pass negative and positive fixtures.
- Required global skill copies hash-match the working source.
- `go run ./cmd/kbcheck local-release` exits 0.
