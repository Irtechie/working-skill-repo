# GHCP AIC No-Paid Readiness Handoff

## State

- Manifest: `docs/plans/2026-07-11-040-kb-ghcp-aic-falsification-manifest.md`
- `plan-to-work`: passed with artifact-backed evidence.
- Slice 001: done; exact trace+span accounting, AIU normalization, redaction,
  telemetry integration, QA, and regression snapshot pass.
- Slices 001-004: deterministic implementation complete.
- Work-to-complete: passed.
- Post-work review: complete; `complete-to-ship` passed with no reachable P0/P1.
- Paid/model calls: one hosted direct leaf call (15.37 AIC); Qwen calls: zero.
- 2026-07-13 attended direct canary attempted after explicit approval:
  - requested `gpt-5.4`, observed `gpt-5.6-sol`;
  - approved per-call maximum 5 AIC, observed 15.37 AIC;
  - classified `invalid-direct`;
  - no proof, correction, AMR, or Qwen inference ran;
  - Qwen reservation and deployment were released.

## Passing Proof

```powershell
go test ./cmd/amrbench -run 'Isolation|Oracle|Fixture|Budget|Containment|InvalidRoute|DisabledRunner|ConformanceNoPaid' -count=1
```

The test passes. Protected `cmd/amrbench/main_test.go` SHA-256:
`8ac3d397ab5ec204196880bc0f3fc216268e2f8fba6e80d85fd0031dc297a1ab`.

## Blocking Proof

```powershell
go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json
```

The command exits zero with `ready: true`, `runner: disabled`,
`paid_calls: 0`, and `release_decision: not-promoted`.

## Resume Contract

1. Preserve the completed no-paid proof and post-work review evidence.
2. Keep `DisabledRunner` for every `--no-paid` path; profile/secret/provider
   construction must remain unreachable.
3. Implement an independent trusted human-approval verifier before enabling any
   non-dry run.
4. Restore and reserve attended routes only immediately before a user-approved
   run, regenerate the preview, then wait for explicit AIC approval.
5. Any attended command must carry a strict route catalog, exact approval
   receipt, experiment ID, and durable budget ledger.
6. Do not use the mislabeled GPT-5.4 canary for pairing or promotion until GHCP
   hard-enforces requested model identity. The user reclassified credit as a
   measurement rather than a cap. Evidence:
   `docs/results/2026-07-13-qwen-canary-budget-route-failure.md`.
7. Exploratory Qwen/Sol cost evidence is complete, but correctness was zero. Do
   not collect more samples until provider response-contract smoke proof passes.
   Evidence:
   `docs/results/2026-07-13-qwen-sol-exploratory-cost.md`.

## Temporary Podman Cleanup Inventory

The user selected Podman as the temporary proof runtime and asked that its
location be retained for later deletion.

- Winget package: `RedHat.Podman`
- Installed version: `5.8.3`
- Executable: `C:\Program Files\RedHat\Podman\podman.exe`
- Machine storage root:
  `C:\Users\marowe\.local\share\containers\podman\machine\wsl`
- Machine: `podman-machine-default` (WSL, rootless)
- Machine identity:
  `C:\Users\marowe\.local\share\containers\podman\machine\machine`
- WSL disk:
  `C:\Users\marowe\.local\share\containers\podman\machine\wsl\wsldist\podman-machine-default\ext4.vhdx`
- Container graph root inside the machine:
  `/home/user/.local/share/containers/storage`
- Pulled image:
  `docker.io/library/golang@sha256:a7ecaac5efda22510d8c903bdc6b19026543f1eac3317d47363680df22161bd8`

Later cleanup must first remove the recorded Podman machine, then uninstall the
exact package:

```powershell
& 'C:\Program Files\RedHat\Podman\podman.exe' machine rm --force podman-machine-default
winget uninstall --exact --id RedHat.Podman
Test-Path 'C:\Program Files\RedHat\Podman\podman.exe'
Test-Path 'C:\Users\marowe\.local\share\containers\podman\machine\wsl\wsldist\podman-machine-default\ext4.vhdx'
```

Both final `Test-Path` checks must return `False`.
