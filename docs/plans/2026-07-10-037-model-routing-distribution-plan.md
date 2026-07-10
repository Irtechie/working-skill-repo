---
kb_id: kb-2026-07-10-session-model-routing
slice_id: slice-007
title: "Ship portable binaries, installer fallback, docs, and sync"
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
expected_files:
  - path: bin/kb-install.mjs
    op: edit
    scope: "optional verified router binary install/upgrade/uninstall with skill-only fallback"
  - path: bin/kb-install.test.mjs
    op: create
    scope: "cross-platform install lifecycle and missing-binary fallback tests"
  - path: package.json
    op: edit
    scope: "package router metadata/tests without requiring a Go toolchain"
  - path: .github/workflows/cross-platform.yml
    op: edit
    scope: "build and smoke-test kbrouter on Windows, macOS, and Linux"
  - path: .github/workflows/release.yml
    op: create
    scope: "tagged multi-platform binaries, checksums, and build provenance/attestation"
  - path: README.md
    op: edit
    scope: "document zero-question discovery, optional extras, difficulty routing, overrides, fallbacks, support matrix, and respectful prior-art credit"
  - path: docs/context/operations/testing.md
    op: edit
    scope: "document routing and release proof commands"
  - path: config/skill-quality.json
    op: edit
    scope: "track kb-models and synchronized KB skill copies"
protected_oracles:
  - path: bin/kb-install.test.mjs
    role: "install, upgrade, uninstall, and fallback oracle"
    sha256: "filled by kb-work after RED/protection"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
can_continue_other_slices: false
---

# Model Routing Distribution

## What To Build

Make the advisory pilot easy to copy/install: platform binaries are optional enhancements, the installer verifies them, and the file-copy skill path remains fully functional with `router-unavailable` fallback.

## Acceptance Criteria

- One installer command installs skills and the matching prebuilt router when available without requiring Go.
- Upgrade/uninstall are deterministic, checksummed, backup-safe, and cross-platform.
- Missing/incompatible binary leaves ordinary KB work usable on the current model.
- Tagged artifacts have checksums and GitHub build provenance; docs do not claim code signing that was not performed.
- Working, global Codex/Copilot/agents, and ATV copies are diff-reviewed then hash-synced; release gates pass.

## Test Scenarios

- Windows/macOS/Linux install, same-version skip, upgrade, uninstall, offline/missing release, checksum mismatch, and skill-only use.
- Global/ATV drift before sync and hash equality after sync.
- README support matrix distinguishes Codex pilot, later cohorts, and router-unavailable fallback.

## Tier Rationale

Large: release automation, supply-chain integrity, cross-platform install, and multi-repo sync are broad/high-risk.

## Scope Boundary

No tag publication, merge, default-branch push, forced sync overwrite, or claim of unavailable signing/provider support.
