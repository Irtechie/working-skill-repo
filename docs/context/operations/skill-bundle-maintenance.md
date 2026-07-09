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
- `<atv-repo>\.github\skills\<skill>\`

ATV shipped copies:

- `<atv-repo>\pkg\scaffold\templates\skills\<skill>\`
- `<atv-repo>\plugins\atv-everything\skills\<skill>\`

Before overwriting a global or ATV copy, review drift. Newer useful work found
only in a global install must be merged back into this repo first, not
discarded.

Source-of-truth invariant:

- `<working-skill-repo>\.github\skills` is the source for KB-owned skills.
- Required global/ATV installs are deployed copies for runners, not authorship
  locations.
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

After editing this repo, sync the final approved copy to the required targets
and ATV shipped copies when that skill intentionally ships there.

Verify:

```powershell
go run ./cmd/kbcheck local-release
git diff --check
```

## ATV Upstream Policy

This is a hand-curated ATV-derived snapshot. There is no automatic upstream
merge bot.

Original ATV `upstream/main` is authoritative for ATV-native changes to inspect,
not a mirror target. KB-owned skills are this repo's overlay.

Use the read-only upstream report before deciding what to port:

```powershell
go run ./cmd/kbcheck atv-delta
```

Classifications:

- `kb-owned-reject` - upstream changed a skill KB owns locally; do not apply it
  over the KB copy.
- `shared-overlap-review` - manually review and port useful improvements.
- `atv-native-candidate` - upstream change may matter to an ATV-native skill.
- `superseded-workflow-reject` - old workflow skill replaced by KB lanes.
- `unknown-review` - needs human review.

Superseded workflow skills such as `lfg`, `slfg`, and `workflows-*` stay out
unless a current app uses them or a focused porting plan proves value.

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

Full global install:

```shell
npx github:Irtechie/working-skill-repo --target all --profile full
```

Repo-local GitHub Copilot install:

```shell
npx github:Irtechie/working-skill-repo --target repo --repo <path-to-your-project> --profile core
```
