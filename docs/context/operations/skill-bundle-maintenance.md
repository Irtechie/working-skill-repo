# Skill Bundle Maintenance

This document holds operational detail that should not live in the root README.

## Repo Boundary

This repo should contain skills, agents, scripts, templates, and durable
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
go run .\cmd\kbcheck core
```

Local release:

```powershell
go run .\cmd\kbcheck local-release
```

Live release:

```powershell
go run .\cmd\kbcheck live-release
```

`cmd/kbcheck` owns top-level orchestration. Individual validators are still
mixed Go and PowerShell.

Important scripts:

- `scripts/skill-lint.ps1`
- `scripts/route-complexity-eval.ps1`
- `scripts/skill-eval.ps1`
- `scripts/skill-eval-run-codex.ps1`
- `scripts/skill-eval-run-ghcp.ps1`
- `scripts/skill-eval-wrap.ps1`
- `scripts/skill-eval-run-live-corpus.ps1`
- `scripts/skill-eval-claims.ps1`
- `scripts/skill-eval-quality.ps1`
- `scripts/skill-eval-regression-report.ps1`
- `scripts/skill-surface-minimality.ps1`
- `scripts/atv-upstream-delta.ps1`
- `scripts/skill-sync-report.ps1`

Live model evals are explicit because they shell to authenticated local CLIs.
Dry-run adapters are part of the local gate; live calls are not implied by a
local green run.

## Sync Targets

Working source:

- `E:\working-skill-repo\.github\skills\<skill>\`

Required targets:

- `C:\Users\marowe\.codex\skills\<skill>\`
- `C:\Users\marowe\.copilot\skills\<skill>\`
- `C:\Users\marowe\.agents\skills\<skill>\`
- `E:\all-the-vibes\.github\skills\<skill>\`

ATV shipped copies:

- `E:\all-the-vibes\pkg\scaffold\templates\skills\<skill>\`
- `E:\all-the-vibes\plugins\atv-everything\skills\<skill>\`

Before overwriting a global or ATV copy, review drift. Newer useful work found
only in a global install must be merged back into this repo first, not
discarded.

After editing this repo, sync the final approved copy to the required targets
and ATV shipped copies when that skill intentionally ships there.

Verify:

```powershell
go run .\cmd\kbcheck local-release
git diff --check
```

## ATV Upstream Policy

This is a hand-curated ATV-derived snapshot. There is no automatic upstream
merge bot.

Original ATV `upstream/main` is authoritative for ATV-native changes to inspect,
not a mirror target. KB-owned skills are this repo's overlay.

Use the read-only upstream report before deciding what to port:

```powershell
.\scripts\atv-upstream-delta.ps1
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

`E:\agent-marketplace` is a private approved catalog, not a global install.

Promotion requires:

1. evidence;
2. human approval;
3. `SKILL.md` review;
4. hash pin;
5. approved copy placed under `E:\agent-marketplace\skills`;
6. runtime roots synced only from the approved copy.

Use the promotion script so the safe path is also the fast path:

```powershell
.\scripts\promote-marketplace-skill.ps1 `
  -Source <reviewed-skill-dir> `
  -SkillId <skill-id> `
  -ApprovalReason "<why this is approved>" `
  -InstallTargets codex,copilot,agents `
  -Approved
```

Quarantine is a firebreak, not a category label. Active and approved skill roots
must not resolve into `E:\agent-marketplace\quarantine`.

## Security

`atv-security` is the current approved single-skill exception from ATV. It is
hash-pinned in `E:\agent-marketplace\catalog\approved-skills.json`, mirrored in
`E:\agent-marketplace\skills\atv-security`, and installed into the Codex,
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

GitHub Copilot personal install:

```powershell
$src = 'E:\working-skill-repo'
Copy-Item "$src\.github\skills\*" "$env:USERPROFILE\.copilot\skills" -Recurse -Force
Copy-Item "$src\.github\agents\*" "$env:USERPROFILE\.copilot\agents" -Force
```

Codex personal install:

```powershell
$src = 'E:\working-skill-repo'
Copy-Item "$src\.github\skills\*" "$env:USERPROFILE\.codex\skills" -Recurse -Force
Copy-Item "$src\.github\agents\*" "$env:USERPROFILE\.codex\agents" -Force
```

Repo-local GitHub Copilot install:

```powershell
$src = 'E:\working-skill-repo'
$dst = 'E:\path\to\your\project'
Copy-Item "$src\.github\skills" "$dst\.github\skills" -Recurse -Force
Copy-Item "$src\.github\agents" "$dst\.github\agents" -Recurse -Force
Copy-Item "$src\AGENTS.md" "$dst\AGENTS.md" -Force
Copy-Item "$src\.github\copilot-instructions.md" "$dst\.github\copilot-instructions.md" -Force
```
