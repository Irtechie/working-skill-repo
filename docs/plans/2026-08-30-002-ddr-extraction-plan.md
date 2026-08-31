---
slice_id: closure-002
title: Extract bounded complexity-aware DDR without evaluator machinery
status: planned
blockers: [closure-001]
verification: tdd
test_level: integration
functional_risk: broad
execution_class: cli
model_tier: medium
model_tier_reason: "The change spans the existing router, its contract tests, and three documentation surfaces, but the policy is fixed: first owner by qualified complexity, parent-owned recovery after a local failure, and no evaluator or provider-specific rule."
model_requirements:
  - "Go refactoring inside an existing router"
  - "reading and reducing a mixed candidate diff"
  - "cross-document contract synchronization"
  - "deterministic CLI and contract testing"
escalation_triggers:
  - "the retained router code cannot be separated from evaluator flags, actor workspaces, or structured evaluator output"
  - "a change permits a second local route or automatic downgrade"
  - "a proposed rule names a provider or model instead of a capability"
  - "the existing tests cannot demonstrate complexity-qualified owner selection"
token_budget: 7000
cost_tier: 2
cheaper_option_ruled_out: "Restoring the old single-attempt wording retains a real UX limitation; PR #36 contains candidate implementation and tests, but the evaluator portion must be removed rather than adopted."
owning_component: kbrouter
expected_files:
  - cmd/kbrouter/dispatch.go
  - cmd/kbrouter/dispatch_test.go
  - cmd/kbrouter/main.go
  - cmd/kbcheck/ddr_contract_test.go
  - README.md
  - .github/skills/kb-work/SKILL.md
  - docs/context/architecture/kb-workflow.md
conflict_domains: [go:kbrouter, go:kbcheck-commands, skills:kb-work, docs:readme, docs:architecture]
shared_resources: [DDR route grammar, route qualification policy]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbrouter ./cmd/kbcheck -run 'Dispatch|ProductionDDRContract' -count=1"
  expect: 0
hitl: none
---

# Extract Bounded Complexity-Aware DDR

## Outcome

The router and documentation select the first execution owner from live,
complexity-qualified routes. A local failure returns to the parent for an
explicit reassessment; it never silently tries a second local route, downgrades
the tier, or treats provider identity as a capability rule.

## Ordered Work

1. Diff PR #36 against `main` and retain only router changes that change
   first-owner selection based on a route satisfying the slice capability
   floor. Exclude evaluator flags, evaluator contract types, actor workspace
   paths, structured evaluator output, evaluation corpus, and ACL changes.
   - Pass criterion: no `evaluator-` CLI flag, evaluator type, or evaluator
     fixture path exists on the replacement branch.
2. Preserve the current bounded recovery semantics while adding the
   complexity-qualified first-choice behavior. Refuse an automatic lower-tier
   or second-local attempt.
   - Pass criterion: router tests cover qualification, local failure returning
     to parent, and refusal of a second local route.
3. Update all four DDR contract surfaces in the same change. Restore an
   explicit no-second-local-route statement in README while retaining the
   complexity-based first-owner statement.
   - Pass criterion: `TestProductionDDRContractExcludesAMRAndSeparatesHostSurfaces`
     passes without weakening its prohibition.
4. Add or retain a focused integration test that proves behavior from the
   emitted router request/receipt, not only documentation text.
   - Pass criterion: the test fails if selection ignores the declared
     capability/complexity floor.

## Acceptance Criteria

- The normal path contains no Opus, Claude, or provider-specific routing rule.
- A qualified route is selected from live capability evidence for the required
  tier and complexity.
- A local route failure does not produce a second local dispatch.
- Router, README, skill, architecture doc, and contract test agree.
- The replacement contains no epistemic evaluator/corpus or Windows ACL file.

## Scope Boundary

Do not repair, promote, baseline, or run the epistemic evaluator. Preserve PR
#36's commits as history until the replacement merges, then close that PR as
superseded rather than merging its mixed tree.
