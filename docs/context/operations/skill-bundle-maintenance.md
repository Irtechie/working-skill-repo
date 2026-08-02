# Skill Bundle Maintenance

This document holds operational detail that should not live in the root README.

## Repo Boundary

This repo should contain skills, agents, native gate tooling, templates, and durable
references needed by the workflow. It should not carry consuming-project
brainstorms, plans, handoffs, research notes, or context maps unless the work is
explicitly about maintaining this skill bundle.

Consuming projects own their local:

- `todo.md`
- `todo-done.md`
- `docs/context/*`
- `docs/handoffs/*`
- `.github/skills/learned-*`
- `config/pipelines/*.json`
- `.atv/pipeline-runs`
- `.agent-marketplace/skill-lock.json`

## Canonical Gates

Core:

```powershell
go run ./cmd/kbcheck core
```

Local release:

```powershell
go run ./cmd/kbcheck local-release
```

Live release:

```powershell
go run ./cmd/kbcheck live-release
```

`cmd/kbcheck` owns quality, release, eval, marketplace, and drift-report
orchestration. The current skill-repo quality/release harness is Go-native.
Remaining `.ps1` files are narrow helper scripts, not the top-level gate.

## Optional Managed Binaries

`bin\kb-install.mjs` manages optional `kbrouter` and `kbreconcile` artifacts
under `~/.kb/bin`. Automatic mode preserves a skill-only install when either
asset is absent. Checksums, managed-state hashes, backups, reconciler downgrade
refusal, and drift-safe uninstall protect existing files.

No signed provenance artifact for `kbreconcile` is published by this repository.
Its install record therefore says `checksum-only` and
`protected_writer_capable: false`. Source/dev inventory, planning, apply,
verify, and reference conformance remain usable. Privileged protected-writer
operation fails closed until a pinned signed provenance source and a live
authoritative adapter both exist.

Live model evals are explicit because they shell to authenticated local CLIs.
Dry-run adapters are part of the local gate; live calls are not implied by a
local green run.

## Sync Targets

Working source:

- `<working-skill-repo>\.github\skills\<skill>\`

Required targets:

- `~/.codex/skills/<skill>/`
- `~/.copilot/skills/<skill>/`
- `~/.agents/skills/<skill>/`

Before overwriting a global copy, review drift. Newer useful work found
only in a global install must be merged back into this repo first, not
discarded.

Source-of-truth invariant:

- `<working-skill-repo>\.github\skills` is the source for KB-owned skills.
- Required global installs are deployed copies for runners, not authorship locations.
- A red `skill-sync-report` is a release blocker for unattended runners. It may
  mean a global-only production fix exists and must be merged back, or it may
  mean a stale global copy would downgrade the runner.
- Never reinstall or sync from globals to other targets. First merge useful
  global-only drift into this repo, prove it here, then sync from this repo
  outward.

Use the read-only report when deciding what drift exists:

```powershell
go run ./cmd/kbcheck skill-sync-report
```

Use doctor when you want the same evidence plus a safe repair path:

```powershell
go run ./cmd/kbcheck doctor
go run ./cmd/kbcheck doctor --fix
```

`doctor --fix` is intentionally conservative. It repairs missing required
targets from `<working-skill-repo>\.github\skills` and repairs stale targets
only when `.kb-sync/<skill>.sha256` proves the target was last deployed from
this repo. Unknown drift is refused with a merge-back instruction so useful
global-only edits are not silently overwritten.

After editing this repo, sync the final approved copy to the required global targets.

Verify:

```powershell
go run ./cmd/kbcheck local-release
git diff --check
```

## Cargo Build Storage

KB workflows reuse one stable per-project Cargo target across checks, repair,
troubleshooting, workers, sessions, and worktrees. Resolve it with
`go run ./cmd/kbcheck cargo-storage --action resolve --run-id <run-id> --json`.
Before Cargo execution, require `validate-ready`.
A new phase- or run-specific target causes a full dependency rebuild and is not
an isolation mechanism.

- The native resolver treats an external absolute `CARGO_TARGET_DIR` as a cache
  root and appends a collision-resistant repository key. Relative or
  worktree-local values map to the same project-keyed cache.
- The receipt lives under the Git common directory, has a collision-resistant
  run filename, serializes mutations, and includes the applied environment plus
  Cargo config fingerprint.
- When a consuming repo lacks `cmd/kbcheck`, the portable fallback applies the
  same repository-key formula to an existing external absolute cache root,
  forbids temporary targets, and performs no deletion.
- Never create `target-check`, `target-repair`, `target-repro`,
  `release-api-probe-target`, or equivalent agent-run targets.
- Allow a run-owned temporary target only through native `register-temp`, using
  the basename `kb-cargo-temp-<24-lowercase-hex>`, then remove it through native
  `finalize` after its final consumer.
- Finalization reports stable bytes retained and temporary bytes removed. It
  never removes the shared target as routine cleanup.
- Record no-Cargo runs with native `not-applicable --reason`; `validate` accepts
  only complete cleanup or a reason-bearing no-Cargo receipt.

## Marketplace

`<agent-marketplace>` is a private approved catalog, not a global install.

Promotion requires:

1. evidence;
2. human approval;
3. `SKILL.md` review;
4. hash pin;
5. approved copy placed under `<agent-marketplace>\skills`;
6. runtime roots synced only from the approved copy.

Use the promotion command so the safe path is also the fast path:

```powershell
go run ./cmd/kbcheck marketplace-promote `
  --source <reviewed-skill-dir> `
  --skill-id <skill-id> `
  --approval-reason "<why this is approved>" `
  --install-targets codex,copilot,agents `
  --approved
```

Quarantine is a firebreak, not a category label. Active and approved skill roots
must not resolve into `<agent-marketplace>\quarantine`.

## Security

`atv-security` is the current approved single-skill exception from ATV. It is
hash-pinned in `<agent-marketplace>\catalog\approved-skills.json`, mirrored in
`<agent-marketplace>\skills\atv-security`, and installed into the Codex,
Copilot, and shared agents global skill directories.

Do not bulk-install ATV skills globally. Promote each skill through the
marketplace boundary first.

Dependency vulnerability proof prefers OSV Scanner:

```powershell
osv-scanner scan source -r <repo-or-scope-path> --format json --output-file docs/security/osv-YYYY-MM-DD.json
```

If OSV is unavailable, record `skipped-unavailable` rather than inventing
vulnerability findings from version age alone.

## Install Snippets

Core global install:

```shell
npx github:Irtechie/working-skill-repo --target all --profile core
```

Skills-only install:

```shell
npx github:Irtechie/working-skill-repo --target all --profile core --router skip --reconciler skip
```

Full global install:

```shell
npx github:Irtechie/working-skill-repo --target all --profile full
```

Repo-local GitHub Copilot install:

```shell
npx github:Irtechie/working-skill-repo --target repo --repo <path-to-your-project> --profile core
```
