# GHCP AIC Snapshot Blocker - Resolved

## State

- Manifest: `docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md`
- Gate: `plan-to-work` passed.
- Current slice: `slice-001`, blocked before implementation.
- Paid/model calls: none.

Resolved by proving the checksum drift came from reviewed later routing slices,
removing stale test-source checksum assertions while preserving executable CLI
checks and immutable fixture/result hashes, and replaying all snapshots:
`snapshot-verify: PASS 11/11 snapshots`.

## Failure

Command:

```powershell
.github\skills\kb-regression-snapshot\scripts\kb-regression-snapshot.ps1 verify
```

Observed:

```text
file-checksum internal/modelrouting/selector_test.go
expected b421387416e485307d28a45d8ea63a1b190796ea639a36ae5fd23059efec01d2
observed 4f10cc21874ed6f2295f9886049abbe076aa892c7176aabefabc3a0b61ca4f29
```

The stale expected hash appears in:

- `.kb/snapshots/session-model-routing-slice-002.json`
- `.kb/snapshots/session-model-routing-slice-005a.json`
- `.kb/snapshots/session-model-routing-slice-005b-spec.json`
- `.kb/snapshots/session-model-routing-slice-005b.json`

## Resume

1. Prove whether the checksum change came from reviewed later session-routing
   slices included in landed commit `51a49de`.
2. If intended, refresh only the superseded snapshot checksum assertions.
3. Rerun the complete snapshot verification.
4. Return `slice-001` to `pending` only after verification passes.
5. Continue deterministic/no-paid work. Any AIC spend still requires the
   user's attended approval after the requested readiness preview.
