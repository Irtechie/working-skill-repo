---
kb_id: kb-2026-07-29-cargo-build-storage-contract
slice_id: slice-001
title: "Enforce stable Cargo build storage across KB workflows"
blockers: []
verification: tdd
test_level: unit
functional_risk: narrow
model_tier: medium
model_tier_reason: "The edit is text-focused but crosses seven workflow owners, delegated prompts, cleanup semantics, and global propagation."
model_requirements:
  - "Can preserve cross-skill contracts and add a deterministic Go contract test."
  - "Can inspect and synchronize all required runtime skill roots without overwriting unknown drift."
escalation_triggers:
  - "A required global skill differs from the repository source before propagation."
  - "The release gate reveals an incompatible existing cleanup or proof contract."
workspace_mode: shared-serial
conflict_domains:
  - "skill:cargo-build-storage"
  - "file:cmd/kbcheck/skill_repo_contract_test.go"
shared_resources:
  - "git:integration-owner"
  - "sync:global-skills"
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'CargoStorage|CargoBuildStorage' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/skill_repo_contract_test.go
    op: edit
    scope: "Add a contract test that fails when Cargo storage rules disappear from owning skills."
  - path: cmd/kbcheck/cargo_storage.go
    op: create
    scope: "Resolve deterministic shared targets and guard exact run-owned cleanup with versioned receipts."
  - path: cmd/kbcheck/cargo_storage_test.go
    op: create
    scope: "Prove cross-worktree resolution, explicit absolute targets, safe cleanup, and forged-marker rejection."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Expose the native cargo-storage command and validate its CLI contract."
  - path: config/skill-quality.json
    op: edit
    scope: "Record the already-over-limit kb-finalize orchestration surface as an intentional reviewed long skill."
  - path: .github/skills/kb-check/SKILL.md
    op: edit
    scope: "Define the canonical stable Cargo target and temporary-target exception contract."
  - path: .github/skills/kb-fix/SKILL.md
    op: edit
    scope: "Reuse the canonical Cargo target for reproduction and verification."
  - path: .github/skills/kb-troubleshoot/SKILL.md
    op: edit
    scope: "Reuse the canonical Cargo target across diagnostic iterations."
  - path: .github/skills/kb-repair/SKILL.md
    op: edit
    scope: "Forbid repair-specific Cargo targets and preserve the failing check environment."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Resolve and pass one build-storage contract to every slice owner."
  - path: .github/skills/kb-work/references/execution-prompt.md
    op: edit
    scope: "Carry the resolved build-storage contract into delegated execution."
  - path: .github/skills/kb-finalize/SKILL.md
    op: edit
    scope: "Remove only recorded run-owned temporary targets and emit a disk receipt."
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "Require the build-storage cleanup receipt before completion."
  - path: docs/context/operations/skill-bundle-maintenance.md
    op: edit
    scope: "Document the portable Cargo storage invariant and safe cleanup boundary."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document the native Cargo storage behavioral proof."
  - path: docs/context/architecture/kbcheck.md
    op: edit
    scope: "Add cargo-storage to the native kbcheck architecture map."
  - path: docs/context/PROJECT.md
    op: edit
    scope: "Route fresh sessions to the Cargo storage command and tests."
  - path: README.md
    op: edit
    scope: "Document native and fail-closed portable Cargo storage runtime behavior."
protected_oracles:
  - path: cmd/kbcheck/skill_repo_contract_test.go
    role: "Cross-skill Cargo build storage contract"
    sha256: "f04909d4d3cf414ac064415ff105fd125c70009ec839c84c260ecdb57f2f46af"
    update_policy: "Requires an explicit plan update when the storage contract changes."
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Enforce Stable Cargo Build Storage Across KB Workflows

## What to Build

Create one portable build-storage contract for Rust/Cargo commands. Every workflow must reuse a stable per-project target directory across phases, sessions, workers, and worktrees. Phase-named or run-named targets are forbidden by default. A separate target is allowed only for a documented technical isolation requirement, must be recorded as run-owned, and must be removed during finalization.

## Acceptance Criteria

1. `kb-check` owns the canonical contract and explicitly forbids `target-check`, `target-repair`, `target-repro`, and equivalent phase/run-specific Cargo target paths.
2. `kb-fix`, `kb-troubleshoot`, and `kb-repair` preserve the same Cargo target directory between reproduction, fixes, retries, and verification.
3. `kb-work` resolves the contract once and includes it in every delegated slice prompt.
4. `kb-finalize` deletes only recorded run-owned temporary targets, preserves shared targets, and records retained bytes plus removed paths/bytes.
5. `kb-complete` cannot report terminal completion without the cleanup receipt or an explicit not-applicable reason.
6. A deterministic Go contract test protects all owning surfaces.
7. Required Codex, Copilot, and shared-agent installs match the reviewed repository source after propagation.

## Test Scenarios

1. Add the contract test first and observe it fail because required policy tokens are absent.
2. Add the cross-skill and native behavioral contract and observe
   `go test ./cmd/kbcheck -run 'CargoStorage|CargoBuildStorage' -count=1` pass.
3. Run `go run ./cmd/kbcheck core`.
4. Run the pre-propagation `go run ./cmd/kbcheck local-release` gate, accepting
   only the expected reviewed sync drift as the reason it cannot become green.
5. Propagate the reviewed complete skill directories, then rerun
   `local-release` and verify repository/global hashes.

Propagation stages a complete copy of each runtime skill root beside the live
root, overlays and verifies all changed skill directories there, then swaps
that root into place with a retained rollback directory. A failed swap restores
the prior root before continuing; no runtime observes partially copied skill
directories.

## Scope Boundary

This slice does not delete the user's existing 19.05 GB directory, change Cargo itself, modify consuming repositories, or introduce a daemon. Existing build artifacts require a separate explicitly authorized cleanup after this prevention fix.

## Context Packet

- Source: user-measured duplicate targets and named owning skills.
- Memory: `todo.md`, `docs/context/PROJECT.md`, `docs/context/architecture/skills.md`, and `docs/context/operations/skill-bundle-maintenance.md`.
- Constraint: repository source must win only after all global drift is reviewed.
- Proof: targeted Go contract test, `kbcheck core`, `kbcheck local-release`, and hash equality.
- Search policy: inspect only the named skills, delegated prompt, contract tests, and maintenance docs.
