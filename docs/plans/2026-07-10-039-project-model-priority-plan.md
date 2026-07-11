---
kb_id: kb-2026-07-10-session-model-routing
slice_id: slice-005b
title: "Persist honest user-local project source priority"
blockers: [slice-003, slice-005a]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
context_packet_path: docs/plans/2026-07-10-session-model-routing-context/slice-005b.json
proof_check:
  kind: command_exit
  command: "go test ./internal/modelrouting ./cmd/kbrouter -run 'Qualified|Hosting|Origin|Priority|Preference|QuickAdd'"
  expect: 0
hitl: false
expected_files:
  - path: internal/modelrouting/catalog.go
    op: edit
    scope: "Separate native/extra origin, hosting class, and dispatch-qualified evidence from trust and exact dispatch proof."
  - path: internal/modelrouting/selector.go
    op: edit
    scope: "Apply automatic, self-hosted-first, and native-first only among already eligible routes."
  - path: internal/modelrouting/selector_test.go
    op: edit
    scope: "Prove source priority never infers locality from private trust or changes tier/trust eligibility."
  - path: cmd/kbrouter/catalog.go
    op: edit
    scope: "Store personal project priority in a versioned user-local collection keyed by canonical project identity."
  - path: cmd/kbrouter/catalog_test.go
    op: edit
    scope: "Prove quick-add defaults, no-repo-write project priority, isolation, clear/reset, and metadata redaction."
  - path: cmd/kbrouter/main.go
    op: edit
    scope: "Expose explicit project priority configuration without adding a normal-work prompt."
  - path: cmd/kbrouter/select.go
    op: edit
    scope: "Load saved project priority when no run override exists and keep run overrides authoritative."
  - path: cmd/kbrouter/select_test.go
    op: edit
    scope: "Prove saved priority, automatic default, override precedence, and honest selection output."
protected_oracles:
  - path: cmd/kbrouter/catalog_test.go
    role: "user-local project preference, quick-add, and no-repo-write oracle"
    sha256: "0ab7eebc9d94f1c91a4bffb11fc77b30bc118eca95140a587c2472cf545569e6"
    update_policy: "requires explicit plan update"
status: done
owner: agent
can_continue_other_slices: false
---

# Slice 005b — User-Local Project Source Priority

## Outcome

The current master automatically chooses host-native models. Optional extra
routes are configured once in user-local state. Ordinary work silently uses
`automatic`; explicit setup may save `self-hosted-first` or `native-first` for
one canonical project without modifying the repository or granting trust.

## Acceptance Criteria

1. Route origin (`native|extra`), hosting (`self-hosted|provider-hosted|unknown`),
   trust boundary, and readiness evidence are independent fields.
2. Adapter evidence establishes `dispatch-qualified`; only an exact route-bound
   receipt linked to deterministic proof establishes `dispatch-proven`.
3. User-local project priorities are a versioned collection keyed by canonical
   project identity. `automatic` is an explicit storable value.
4. Normal discovery, map, planning, and work create no preference file and ask no
   priority question. Explicit priority setup writes no repository file.
5. `self-hosted-first` matches only explicitly self-hosted eligible routes;
   private/provider-hosted and unknown routes do not match by inference.
6. `native-first` prefers eligible host-native routes. `automatic` preserves the
   evidence-based order. No preference changes trust, tier, tools, context,
   authority, proof, or fallback eligibility.
7. Run-scoped `use`, `require`, `prefer self-hosted|native`, and ignore override
   saved priority without weakening their existing safety boundaries.
8. Quick-add uses minimal inputs and conservative unknown/unqualified defaults;
   approval and capability evidence remain separate.

## Test Scenarios

- Two canonical projects retain different priorities in the same user profile.
- A worktree shares preference with its repository; an unrelated clone or
  identity replacement does not inherit approval.
- A private provider-hosted route does not win `self-hosted-first`.
- An explicitly self-hosted LAN route wins only within the same eligible tier.
- No saved priority behaves as automatic without writing state or prompting.
- Run override wins saved priority; exact require remains trust-bound.
- Adapter-qualified and route-proven evidence remain distinct in output and
  capability credit.

## Scope Boundary

No generic MCP dispatch, direct chat-completions worker, fleet controller,
hosted-model version ladder, or tracked personal preference.
