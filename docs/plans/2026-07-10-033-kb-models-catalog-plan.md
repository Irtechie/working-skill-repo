---
kb_id: kb-2026-07-10-session-model-routing
slice_id: slice-003
title: "Manage optional extra routes without a setup questionnaire"
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
expected_files:
  - path: cmd/kbrouter/main.go
    op: create
    scope: "public models show/add/remove/prefer/ignore-routing/doctor/calibrate/discover CLI"
  - path: cmd/kbrouter/catalog.go
    op: create
    scope: "Codex CLI and configured OpenAI-compatible discovery adapters with bounded deadlines"
  - path: cmd/kbrouter/catalog_test.go
    op: create
    scope: "functional CLI and slow/dead adapter fixtures"
  - path: .github/skills/kb-models/SKILL.md
    op: create
    scope: "low-ceremony optional extra-route workflow and schema guidance"
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "register routing conformance selftests in core"
protected_oracles:
  - path: cmd/kbrouter/catalog_test.go
    role: "secure CLI and discovery oracle"
    sha256: "6d13708dd3eee4d01fc5c89d1217b9614e13ed8c8b0de6da09a1036d62666006"
    update_policy: "requires explicit plan update"
status: done
owner: agent
can_continue_other_slices: true
---

# Optional Extra-Route Catalog

## What To Build

Provide a small `kbrouter models` CLI and `kb-models` skill. Native discovery happens without questions; users add only routes the active surface cannot know.

## Acceptance Criteria

- `discover`/`show` creates a redacted run catalog and never creates `~/.kb/models.json` by default.
- CRUD/preferences modify explicit user or project scopes atomically and reject secrets/arbitrary commands.
- `doctor` is non-mutating and separately reports discovery, configuration, selection, dispatch, auth, and control readiness.
- Dead adapters respect per-adapter and session budgets while current-model work remains available.
- Project files store aliases/policy, not private connection details.

## Test Scenarios

- Captured `codex debug models` fixture, current-model-only fixture, generic `/v1/models` fixture, dead/slow server, fingerprint refresh.
- Global add/remove/prefer/clear and project alias approval/denial.
- Doctor with missing env key, unreachable endpoint, model absent, and dispatch unavailable.

## Tier Rationale

Large: user-local storage, network discovery, trust boundaries, and public CLI behavior require security-aware integration.

## Scope Boundary

No model inference, Codex process dispatch, controller mutation, or support-label promotion.
