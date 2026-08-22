---
kb_id: kb-2026-08-20-routing-cognitive-delivery
slice_id: routing-002
title: Rename kb-compact to kb-cognitive and route boundary presentation
blockers: [routing-001]
verification: integration
test_level: verification-only
functional_risk: medium
model_tier: medium
model_tier_reason: Cross-skill references and installed-copy synchronization require careful mechanical work.
model_requirements: [skill maintenance, repository search, deterministic test execution]
escalation_triggers: [a live old-name route remains, global copy contains newer drift]
token_budget: 4000
workspace_mode: shared-serial
conflict_domains: [skill:kb-compact, skill:kb-cognitive, file:README.md, sync:global-skills]
proof_check: {kind: command_exit, command: "go test ./cmd/kbcheck -run LowCognitiveBurden -count=1", expect: 0}
hitl: false
status: pending
---

# Rename and Route KB Cognitive

## Acceptance Criteria

- `kb-cognitive` is the sole live skill name; historical references stay factual.
- Brainstorm, plan, work, and PR boundaries use its low-burden projection rules
  when information shape warrants them.
- The skill is not an obligatory inner-loop phase.
- Source and all required global installs hash-identically match.
- Review-facing projections make the merge decision visible before supporting
  detail, using a compact decision block, responsibility table, and a visual
  only when it reduces reviewer reconstruction.

## Expected Files

- `.github/skills/kb-cognitive/**`; current caller skills; README; active
  architecture docs; skill audit; communication tests; required global copies.
