---
type: kb-manifest
kb_id: kb-2026-05-31-atv-security-marketplace
brainstorm_path: chat:2026-05-31-atv-security-marketplace
created: 2026-05-31
status: reviewed
workflow_shape: "skill-bundle-change"
scope-verified-files:
  - docs/plans/2026-05-31-050-kb-atv-security-marketplace-manifest.md
  - docs/plans/2026-05-31-051-marketplace-security-promotion-plan.md
  - docs/plans/2026-05-31-052-atv-security-osv-propagation-plan.md
  - docs/plans/2026-05-31-053-global-atv-security-install-plan.md
  - docs/plans/2026-05-31-054-promotion-proof-completion-plan.md
  - todo.md
  - todo-done.md
  - .atv/kb-completions.txt
  - docs/context/PROJECT.md
  - docs/context/operations/testing.md
  - docs/context/architecture/private-skill-marketplace.md
  - docs/context/memory-maintenance.md
  - E:/agent-marketplace/README.md
  - E:/agent-marketplace/catalog/approved-skills.json
  - E:/agent-marketplace/catalog/harness-index.json
  - E:/agent-marketplace/harnesses/README.md
  - E:/agent-marketplace/harnesses/dependency-vulnerability-osv.json
  - E:/agent-marketplace/skills/README.md
  - E:/agent-marketplace/skills/atv-security/SKILL.md
  - E:/all-the-vibes/plugins/atv-skill-atv-security/skills/atv-security/SKILL.md
  - E:/all-the-vibes/pkg/scaffold/templates/skills/atv-security/SKILL.md
  - E:/all-the-vibes/plugins/atv-pack-security/skills/atv-security/SKILL.md
  - E:/all-the-vibes/plugins/atv-everything/skills/atv-security/SKILL.md
  - C:/Users/marowe/.codex/skills/atv-security/SKILL.md
  - C:/Users/marowe/.copilot/skills/atv-security/SKILL.md
  - C:/Users/marowe/.agents/skills/atv-security/SKILL.md
slices:
  - id: slice-051
    title: "Promote ATV security into the approved marketplace catalog"
    path: docs/plans/2026-05-31-051-marketplace-security-promotion-plan.md
    blockers: []
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Validate marketplace metadata, OSV harness registration, and approved skill hash."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=6 changed=6 discovered=0 unexplained=0; proof: marketplace JSON parsed with ConvertFrom-Json; catalog pinned to source sha256 1b0bcd8a05f79059e7bb8029ecb58d8a4e7492e3812664e3e577943fe9e650b0; memory-impact: durable; areas=marketplace-security-catalog; refresh=pending"
    protected_oracles: []
  - id: slice-052
    title: "Add OSV dependency-vulnerability proof to ATV security"
    path: docs/plans/2026-05-31-052-atv-security-osv-propagation-plan.md
    blockers: []
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Propagate the final atv-security SKILL.md across ATV shipped copies."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=4 changed=4 discovered=0 unexplained=0; proof: OSV guidance present at source line 345, skipped-unavailable line 354, no-fix line 352, version-age guard line 364; ATV shipped copy hashes all equal 1b0bcd8a05f79059e7bb8029ecb58d8a4e7492e3812664e3e577943fe9e650b0; git diff --check passed with existing line-ending warnings; memory-impact: durable; areas=atv-security-a06; refresh=pending"
    protected_oracles: []
  - id: slice-053
    title: "Install the approved ATV security skill into global targets"
    path: docs/plans/2026-05-31-053-global-atv-security-install-plan.md
    blockers: [slice-051, slice-052]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Copy the approved skill to Codex, Copilot, shared agents, and marketplace targets; prove hash equality."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=5 changed=5 discovered=0 unexplained=0; proof: trusted source, marketplace, Codex, Copilot, and shared agents hashes all equal 1b0bcd8a05f79059e7bb8029ecb58d8a4e7492e3812664e3e577943fe9e650b0; firebreak passed issues=0; firebreak selftest passed; memory-impact: durable; areas=global-atv-security-install; refresh=pending"
    protected_oracles: []
  - id: slice-054
    title: "Run completion proof and record memory state"
    path: docs/plans/2026-05-31-054-promotion-proof-completion-plan.md
    blockers: [slice-053]
    verification: verification-only
    test_level: full
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Run json parsing, firebreak checks, kb-check -All, diff checks, and marketplace/global hash proof."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=7 changed=7 discovered=5 unexplained=0; scope-discovery: .atv/kb-completions.txt - kb-complete cadence counter; scope-discovery: E:/agent-marketplace/* - marketplace promotion outputs; scope-discovery: E:/all-the-vibes/* - trusted ATV shipped copies; scope-discovery: C:/Users/marowe/.codex/.copilot/.agents skills - explicit global install targets; proof: marketplace JSON parse passed; hash equality unique_hashes=1 across 8 source/install targets; firebreak and selftest passed; kb-check -All passed with 0 required sync issues; git diff --check passed in working repo, marketplace, and ATV with line-ending warnings only; osv-scanner installed version=2.3.8 after completion; review-mode: local-fallback P0=0 P1=0 P2=0 P3=0; follow-up-resolution: resolved 0, logged 0, blocked 0; kb-map-refresh: done - PROJECT.md, operations/testing.md, architecture/private-skill-marketplace.md, memory-maintenance.md; compound: skipped - promotion/proof wiring, no novel implementation pattern beyond documented marketplace contract; learn: no new generic instincts; evolve: skipped - completion counter 4; compact: skipped - no startup bloat; cleanup: done"
    protected_oracles: []
---

# KB: ATV Security Marketplace Promotion

## Origin

Brainstorm: `chat:2026-05-31-atv-security-marketplace`

## Workflow Shape

`skill-bundle-change` - this promotes one trusted security skill, adds one deterministic security harness descriptor, and syncs known install targets without changing application runtime behavior.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Promote ATV security into the approved marketplace catalog | - | integration | no | done |
| 2 | Add OSV dependency-vulnerability proof to ATV security | - | integration | no | done |
| 3 | Install the approved ATV security skill into global targets | slice-051, slice-052 | integration | no | done |
| 4 | Run completion proof and record memory state | slice-053 | verification-only | no | done |

## Completion Review

- Review mode: `local-fallback` because no authorized reviewer subagent was invoked in this runtime.
- Findings: P0=0, P1=0, P2=0, P3=0.
- Residual risk: none for installation. A repo scan still needs dependency manifests or lockfiles in the target repo.

## Final Proof

- Marketplace JSON parsed with `ConvertFrom-Json`.
- Hash equality across trusted ATV source, ATV shipped copies, marketplace, Codex, Copilot, and shared agents: `1b0bcd8a05f79059e7bb8029ecb58d8a4e7492e3812664e3e577943fe9e650b0`.
- `osv-scanner --version` reports `2.3.8`; smoke scan against this repo executed but exited `128` because no package sources were present.
- `scripts/skill-marketplace-firebreak.ps1` passed with `issues=0`.
- `scripts/skill-marketplace-firebreak-selftest.ps1` passed by requiring the quarantined active-root case to fail.
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All` passed with 0 required sync issues.
- `git diff --check` passed in `E:/working-skill-repo`, `E:/agent-marketplace`, and `E:/all-the-vibes`; output contained line-ending warnings only.

## Assumptions

- `E:/all-the-vibes/plugins/atv-skill-atv-security/skills/atv-security/SKILL.md` is the trusted source for this promotion.
- `E:/agent-marketplace` is the approved private catalog, not an auto-loaded global skill root.
- OSV Scanner is an optional executable proof path: absence of `osv-scanner` is recorded as `skipped-unavailable`, not silently replaced by model judgment.
