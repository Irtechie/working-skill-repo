---
type: kb-manifest
kb_id: kb-2026-07-10-session-model-routing
brainstorm_path: docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md
created: 2026-07-10
status: active
workflow_shape: pipeline-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json"
  expect: 0
  why: "Proves the initial Codex-first advisory pilot without claiming later host cohorts or default routing."
model_tier_contract:
  allowed: [tiny, small, medium, large]
  default: medium
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-brainstorm
    status: passed
    required_evidence:
      - "requirements path exists"
      - "Question Gate classification exists"
      - "Resolve Before Planning is empty"
      - "no unresolved ask-now or research-first items remain"
      - "review findings are resolved"
    proof:
      - docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md
      - docs/context/research/2026-07-09-project-model-routing-surfaces.md
      - "Question Gate and Outstanding Questions sections"
      - "Resolve Before Planning: None"
      - "two completed document-review passes"
    blockers: []
    passed_at: "2026-07-10T03:39:22Z"
    allowed_next_action: "kb-plan docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest and seven slice plans exist"
      - "DAG has no missing blockers or cycles"
      - "every slice has acceptance, expected files, proof, test level, risk, and model tier"
      - "every slice has a validated bounded context packet"
      - "new plans omit concrete model routes and legacy manifests remain readable"
      - "supported cohort and parked claims are explicit"
      - "objective done check exists"
    proof:
      - docs/plans/2026-07-10-030-kb-session-model-routing-manifest.md
      - docs/plans/2026-07-10-031-manifest-neutral-routing-plan.md
      - docs/plans/2026-07-10-032-secure-model-routing-core-plan.md
      - docs/plans/2026-07-10-033-kb-models-catalog-plan.md
      - docs/plans/2026-07-10-034-codex-model-dispatch-plan.md
      - docs/plans/2026-07-10-035-kb-work-model-routing-plan.md
      - docs/plans/2026-07-10-036-model-routing-pilot-plan.md
      - docs/plans/2026-07-10-037-model-routing-distribution-plan.md
    blockers: []
    passed_at: "2026-07-10T04:45:00Z"
    allowed_next_action: "kb-work docs/plans/2026-07-10-030-kb-session-model-routing-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "route-free manifests and stale legacy route hints remain readable"
      - "routing evidence cannot substitute for proof_check"
      - "protected compatibility oracle is hashed after RED/GREEN"
      - "slice diff, deterministic checks, QA, and regression snapshot pass"
    proof:
      - cmd/kbcheck/manifest_contract.go
      - cmd/kbcheck/manifest_contract_test.go
      - evals/dishonest-completion/fixtures.json
      - .github/skills/kb-plan/SKILL.md
      - ".kb/snapshots/session-model-routing-slice-001.json"
      - ".kb/runs/session-model-routing/slice-001-result.md"
      - ".kb/runs/session-model-routing/route-history.jsonl"
      - "qa-browser: skipped — no UI-reachable behavior changed"
      - "qa-lint: PASS — gofmt and git diff --check, 3 implementation files"
    blockers: []
    passed_at: "2026-07-10T05:38:00Z"
    allowed_next_action: "kb-work slice-002"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "automatic selection uses only validated, fingerprinted, dispatch-proven routes"
      - "user/project policy cannot self-promote evidence or activate private routes"
      - "storage, endpoint, receipt, project-identity, and fallback guards are deterministic"
      - "protected security oracle, package checks, QA, and regression replay pass"
    proof:
      - internal/modelrouting/catalog.go
      - internal/modelrouting/storage.go
      - internal/modelrouting/policy.go
      - internal/modelrouting/selector.go
      - internal/modelrouting/receipt.go
      - internal/modelrouting/identity_windows.go
      - internal/modelrouting/identity_unix.go
      - internal/modelrouting/selector_test.go
      - ".kb/snapshots/session-model-routing-slice-002.json"
      - ".kb/runs/session-model-routing/slice-002-result.md"
      - "linux and darwin cross-compile: PASS"
      - "qa-browser: skipped — no UI-reachable behavior changed"
    blockers: []
    passed_at: "2026-07-10T08:22:33Z"
    allowed_next_action: "kb-work slice-003"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "optional routes remain user-local while project policy contains aliases only"
      - "redacted run routes resolve through opaque source IDs and trusted host state"
      - "approval, denial, DNS, ACL, concurrency, and run-root boundaries fail closed"
      - "functional CLI, core, cross-compile, review, QA, and regression replay pass"
    proof:
      - cmd/kbrouter/main.go
      - cmd/kbrouter/catalog.go
      - cmd/kbrouter/catalog_test.go
      - internal/modelrouting/storage.go
      - internal/modelrouting/storage_acl_windows.go
      - internal/modelrouting/storage_acl_unix.go
      - internal/modelrouting/storage_lock_windows.go
      - internal/modelrouting/storage_lock_unix.go
      - .github/skills/kb-models/SKILL.md
      - ".kb/snapshots/session-model-routing-slice-003.json"
      - ".kb/runs/session-model-routing/slice-003-result.md"
      - "P0/P1 coherence review: clear"
      - "Windows tests plus Linux/Darwin amd64 compile: PASS"
      - "qa-browser: skipped — no UI-reachable behavior changed"
    blockers: []
    passed_at: "2026-07-10T11:10:00Z"
    allowed_next_action: "kb-work slice-004"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Windows external dispatch is Job Object contained and dispatch-proven"
      - "Linux/macOS external dispatch stays unavailable-before-start while discovery, preview, and current fallback remain portable"
      - "dispatcher-owned route attestation lives in private user-local state"
      - "telemetry verifies but cannot mint capability credit"
      - "ordinary proof stays authoritative"
    proof:
      - .kb/runs/session-model-routing/c1-final-report.md
      - .kb/runs/session-model-routing/c2-final-report.md
      - cmd/kbrouter/dispatch_test.go
      - .kb/runs/session-model-routing/slice-004-result.md
      - ".kb/snapshots/session-model-routing-slice-004.json"
      - "Windows package tests, native Linux package tests, Linux vet, Darwin cross-compile, live native Codex discovery, and final P0/P1-clear reviews are inherited from C1/C2"
    blockers: []
    passed_at: "2026-07-10T20:05:00Z"
    allowed_next_action: "kb-work slice-005"
  - gate_id: slice-slice-005-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "new plans retain difficulty tiers while durable model routes remain optional legacy hints"
      - "work-time selection uses live host evidence, same-tier then upward fallback, and no automatic downward route"
      - "run-scoped overrides, honest provenance, private-route locality, and ordinary proof authority are explicit"
      - "portable Go sandbox guidance, protected fixture, skill eval, lint, and packet checks pass"
    proof:
      - .github/skills/kb-goal/SKILL.md
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-work/references/execution-prompt.md
      - .github/skills/kb-work/references/go-sandbox.md
      - .github/skills/kb-complete/SKILL.md
      - .github/skills/kb-models/SKILL.md
      - evals/skill-eval/selftest/pass-session-model-routing.json
      - "focused and full skill-eval selftests, skill-lint, context-packet validation, and scoped diff check: PASS"
    blockers: []
    passed_at: "2026-07-10T21:36:00Z"
    allowed_next_action: "kb-work slice-006"
slices:
  - id: slice-001
    title: "Keep manifests model-neutral with legacy compatibility"
    path: docs/plans/2026-07-10-031-manifest-neutral-routing-plan.md
    blockers: []
    verification: tdd
    test_level: unit
    functional_risk: narrow
    model_tier: medium
    context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-001.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run ModelRoute"
      expect: 0
    hitl: false
    status: done
    owner: agent
    can_continue_other_slices: true
    notes: "actual-route: codex-cli/gpt-5.4, provider=openai, effort=medium, session=019f4a68-8286-7db0-857a-d7769ad2ada8, exact CLI selector evidence captured, token telemetry unavailable; fallback-reserved: codex-cli/gpt-5.5; TDD: RED missing/invalid route enforcement then GREEN route-neutral contract; scope-ledger: forecast=4 changed=4 discovered=1 unexplained=0, cmd/kbcheck/swarm.go unused, evals/dishonest-completion/fixtures.json required by core regression; kb-repair: commit 763178f migrated stale invalid-model-route fixture to routing-receipt-is-not-proof; kb-check: narrow, package, dishonest-completion, and core tests PASS; kb-qa: gofmt/diff-check PASS, browser skipped because no UI-reachable behavior; regression: repaired contributor-safe core baseline then snapshot-verify PASS 7/7; sync drift is an expected intermediate state owned by slice-007 and remains a blocking final local-release condition; memory-impact: durable routing contract, project memory refresh pending final pilot"
    protected_oracles:
      - path: cmd/kbcheck/manifest_contract_test.go
        role: "new/legacy manifest compatibility oracle"
        sha256: "cbfd6b3265e47b1bba0d97e0590780c8ebbc6978d62041d6e75021536e338b4b"
        update_policy: "requires explicit plan update"
  - id: slice-002
    title: "Select routes conservatively from a secure session catalog"
    path: docs/plans/2026-07-10-032-secure-model-routing-core-plan.md
    blockers: []
    verification: tdd
    test_level: unit
    functional_risk: broad
    model_tier: large
    context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-002.json
    proof_check:
      kind: command_exit
      command: "go test ./internal/modelrouting"
      expect: 0
    hitl: false
    status: done
    owner: agent
    can_continue_other_slices: true
    notes: "actual-route phase 1: codex-cli/gpt-5.5 high, provider=openai, session=019f4abe-c233-7070-a058-6801a2e61752, startup selector evidence captured, 82,596 reported tokens; remediation redispatch to the same verified route failed before work because the account usage ceiling was reached, so context/diff/audit were preserved and the current planner completed the repair as degraded-current with model identity unreported; TDD: original compile RED, initial GREEN, security-audit RED, final GREEN; security: three read-only passes, initial P0 persistence bug and P1 trust/provenance gaps resolved, final findings resolved by evidence-state fingerprints, duplicate refusal, complete request envelopes, and immutable validated catalogs; scope-ledger: forecast=6 changed=8 discovered=2 explained identity_windows.go/identity_unix.go for replacement-resistant cross-platform project identity; kb-check: package, full go test, vet, Windows proof, Linux/macOS compile PASS; race: honest skip because CGO is disabled; kb-qa: gofmt/diff-check PASS, browser skipped because no UI-reachable behavior; regression replay PASS. Slice-003 explicitly extended the protected oracle without removing existing assertions: opaque route identity, exact host-derived route/current state, bounded DNS, Windows/Unix private ACL profiles, project-storage separation, confined run roots, and cross-process mutation locking; dispatcher consumption of pinned IPs/redirect policy and trusted receipt authorship remain slice-004 work."
    protected_oracles:
      - path: internal/modelrouting/selector_test.go
        role: "difficulty, trust, fallback, and current-model degradation oracle"
        sha256: "3e241a0c74d2020fcc1b535c45b45604b042c912292698be3b66bdae3e9a5465"
        update_policy: "requires explicit plan update"
  - id: slice-003
    title: "Manage optional extra routes without a setup questionnaire"
    path: docs/plans/2026-07-10-033-kb-models-catalog-plan.md
    blockers: [slice-002]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-003.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbrouter -run 'Catalog|Doctor|Policy'"
      expect: 0
    hitl: false
    status: done
    owner: agent
    can_continue_other_slices: true
    notes: "planned=large; actual initial route codex-cli/gpt-5.5 high session 019f4b25-8687-78f0-a50a-64da3918965d, 185,429 reported tokens; remediation route codex-cli/gpt-5.5 high session 019f4b44-1c0d-7291-8666-599cc4a48a06, final token telemetry unavailable after sandbox-local Go cache stall; current planner completed preserved remediation. Security/coherence review resolved self-authorization, opaque source identity, exact trusted current/receipt state, attended-policy honesty, denial-before-probe, bounded DNS, current-user ACL/DACL, private/project storage split, run-root containment/revalidation, and cross-process RMW locking. Full packages, core, vet, skill validation/lint, Windows tests, Linux/Darwin compile, diff check, and snapshot replay PASS; UI QA skipped because no UI behavior. No model inference or dispatch is in slice scope."
    protected_oracles:
      - path: cmd/kbrouter/catalog_test.go
        role: "secure CRUD, discovery, and non-mutating doctor CLI oracle"
        sha256: "6d13708dd3eee4d01fc5c89d1217b9614e13ed8c8b0de6da09a1036d62666006"
        update_policy: "requires explicit plan update"
  - id: slice-004
    title: "Dispatch Codex and custom-provider workers with route-bound receipts"
    path: docs/plans/2026-07-10-034-codex-model-dispatch-plan.md
    blockers: [slice-001, slice-002, slice-003]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-004.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbrouter -run 'Dispatch|Receipt|Fallback'"
      expect: 0
    hitl: false
    status: done
    owner: agent
    can_continue_other_slices: true
    notes: "closure-only update: C1 final report proved Windows Job Object containment, native Linux tests, vet, Linux/Darwin cross-compile, and protected oracle hash; C2 final report proved native Codex discovery, dispatcher-owned exact-receipt attestation, verification-only telemetry, canonical containment, Windows/Linux tests, vet, Darwin cross-compile, and live discovery; implementation route used Medium gpt-5.4, trust remediation escalated to Large gpt-5.5 high, and this closure used Small gpt-5.4-mini; scope-discovery: .github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1 - ignore *-spec.json inputs so replay executes captured snapshots only; proof: dispatch regex PASS 27.334s, full go test PASS kbcheck=14.519s kbrouter=95.120s modelrouting=3.636s, core PASS 34 checks, snapshot capture PASS and replay PASS 8/8; no GHCP, TinyBoss, MCP, Unix external dispatch, or direct OpenAI-compatible agent loop claim is added here"
    protected_oracles:
      - path: cmd/kbrouter/dispatch_test.go
        role: "least-privilege host dispatch and receipt binding oracle"
        sha256: "a8c38d5d13edb9c9ed39e2c3752157af68ca9eb8548d99d7bfce1cfedb4a2eba"
        update_policy: "requires explicit plan update"
  - id: slice-005
    title: "Route kb-work slices by live difficulty and preserve proof"
    path: docs/plans/2026-07-10-035-kb-work-model-routing-plan.md
    blockers: [slice-001, slice-004]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    model_tier: medium
    context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-005.json
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck skill-lint"
      expect: 0
    hitl: false
    status: done
    owner: agent
    can_continue_other_slices: true
    notes: "Medium Codex CLI gpt-5.4 worker thread 019f4de1-f9ce-7e53-88db-722752518e10 implemented live work-time routing policy; plan tiers are small/medium/large with planner separate; exact/same-tier/upward/current-degraded fallback, run-scoped overrides, no automatic down-route, honest receipts, already-proven external work handling, private-route locality, and per-Go-invocation .kb/runtime sandboxing are explicit. RED/GREEN focused fixture, 14-file skill-eval selftest, skill-lint, context-packet validation, and scoped diff checks passed. Model identity is launch evidence from the parent Codex CLI invocation; token telemetry unavailable. The selftest trace is intentionally self-reported rather than mislabeled as externally observed."
    protected_oracles:
      - path: evals/skill-eval/selftest/pass-session-model-routing.json
        role: "work-time routing, override, fallback, and provenance behavior oracle"
        sha256: "342ba9e2bf109433e145453f0102ea814acc7c3fd41ec52d31660214201ef669"
        update_policy: "requires explicit plan update"
  - id: slice-006
    title: "Prove the Codex-first advisory pilot and promotion boundary"
    path: docs/plans/2026-07-10-036-model-routing-pilot-plan.md
    blockers: [slice-004, slice-005]
    verification: functional
    test_level: full
    functional_risk: full
    model_tier: large
    context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-006.json
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck model-routing-release --cohort initial-pilot --evidence docs/results/2026-07-10-session-model-routing-initial-pilot.json"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: cmd/kbcheck/model_routing_release_test.go
        role: "support-cohort and no-regression release oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan update"
  - id: slice-007
    title: "Ship portable binaries, installer fallback, docs, and sync"
    path: docs/plans/2026-07-10-037-model-routing-distribution-plan.md
    blockers: [slice-003, slice-005, slice-006]
    verification: integration
    test_level: full
    functional_risk: broad
    model_tier: large
    context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-007.json
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck local-release"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: bin/kb-install.test.mjs
        role: "install, upgrade, uninstall, and router-unavailable fallback oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan update"
---

# KB Session Model Routing

## Origin

Brainstorm: `docs/brainstorms/2026-07-09-session-model-discovery-and-routing-requirements.md`

## Workflow Shape

`pipeline-change` - deterministic routing binary, host adapter, skill behavior, proof contract, and public distribution.

## Supported Cohort

The first shipped cohort is an advisory pilot for Codex CLI plus one generic OpenAI-compatible provider routed through Codex's coding-agent harness. Codex App-only model names remain informative unless a selector is exposed. GHCP, TinyBoss controller actions, MCP model dispatch, generated named agents, and default automatic routing require later conformance evidence.

## Slice Overview

| # | Slice | Blocked By | Verification | Tier | HITL | Status |
|---|---|---|---|---|---|---|
| 1 | Model-neutral manifests | - | tdd | medium | no | done |
| 2 | Secure routing core | - | tdd | large | no | done |
| 3 | Optional extra-route catalog | 2 | integration | large | no | done |
| 4 | Codex/custom-provider dispatch | 1, 2, 3 | integration | large | no | done |
| 5 | KB work-time routing | 1, 4 | integration | medium | no | done |
| 6 | Advisory pilot proof | 4, 5 | functional | large | no | pending |
| 7 | Distribution and sync | 3, 5, 6 | integration | large | no | pending |

## Model Selection Contract

The planner records only difficulty. `kb-work` discovers routes before dispatch, prefers evidence-qualified routes that meet tier/risk/tool/trust needs, and records actual runtime/model evidence when available. Here Large prefers a dispatch-proven GPT-5.5 High route; Medium prefers a dispatch-proven Terra route when the active surface exposes a selector, otherwise an evidence-qualified same-class route such as GPT-5.4, then a stronger route. Visibility alone never proves dispatch.

## Parked Claims

- GHCP supported label until independent catalog/dispatch/receipt fixtures pass.
- TinyBoss reserve/wake/start and lease cleanup until controller conformance passes.
- Outbound MCP model dispatch until a versioned LLM capability exists.
- Default automatic routing until the pilot baseline proves no correctness/rework/intervention regression and at least one material benefit.
