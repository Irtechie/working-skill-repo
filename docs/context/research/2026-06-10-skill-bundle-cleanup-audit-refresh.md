# Skill Bundle Cleanup Audit Refresh

Date: 2026-06-10
Source: pasted external audit

## Evidence Summary

| Finding | Status | Evidence | Action |
|---|---|---|---|
| Fresh-clone `core` gate fails on missing install targets | confirmed | `go run ./cmd/kbcheck core` exited 1 because `skill-sync-report` reported 38 `codex-global` `missing-required` rows. | Moved sync enforcement out of `core`; keep it blocking in `local-release` and direct `skill-sync-report`. |
| `go.mod` pinned a too-new Go directive | confirmed | `go.mod` contained a higher Go directive; current code uses no known feature requiring it. | Lowered to `go 1.22`. |
| Core install profile is not dependency-closed | confirmed | Installer core list was six skills: `kb-start`, `kb-map`, `kb-fix`, `kb-plan`, `kb-work`, `kb-complete`; README also listed direct dependencies such as `kb-check`, `kb-gate`, `kb-review`, `learn`, `evolve`, `document-review`, `todo-create`, and `todo-triage`. | Core now installs every runtime skill plus baseline review/document agents; full adds all specialist agents. |
| Agent surface has unproven/orphan candidates | confirmed-static-only | `go run ./cmd/kbcheck minimality` reported 52 agents, 11 unproven agents, and one trim candidate. | Keep as cold-storage review input; no deletion without human approval and runtime proof. |
| `ce-review` / `kb-review` reference trees are partly forked | confirmed | Hash comparison showed `diff-scope.md` and `resolve-base.sh` match; `findings-schema.json`, `persona-catalog.md`, `post-review-flow.md`, `review-output-template.md`, `review-process.md`, and `subagent-template.md` differ. | Documented as an intentional split for now; future canonicalization should add a drift check for shared files only. |
| `learn` references unshipped hook file | confirmed | `.github/skills/learn/SKILL.md` references `.github/hooks/copilot-hooks.json`; targeted search found no shipped hook file. | Leave as follow-up unless hook work is in scope; docs now avoid claiming hook enforcement from core gate. |
| Negative selftest wording looks like failure | confirmed | Core output contained successful negative-case messages that ended with the word `failed`, including the pipeline unknown-id and marketplace quarantine cases. | Reworded successful negative checks to `correctly rejected`. |
| Repo handoff layout missing | stale | Exact-path preflight showed `docs/handoffs/active`, `docs/handoffs/parked`, and `docs/handoffs/done` all exist. | No work needed. |
| Process history ships to consumers | partially stale | `package.json` excludes `docs/` from npm files; clone-based users still see repo-local process history. | No deletion in this run; archival policy remains a future decision. |
| Cross-platform claim has PowerShell holes | already documented | README already says two narrow PowerShell helpers remain: `kb-regression-snapshot.ps1` and `code-intel.ps1`; targeted search also found the installer fallback `scripts/install-kb.ps1`. | Keep platform caveat; do not claim those helper paths are Go-native. |

## Verification Commands Used

```powershell
go run ./cmd/kbcheck core --list
go run ./cmd/kbcheck minimality
go run ./cmd/kbcheck core
rg --files -g "*.ps1"
```

## Decision

Implement confirmed contributor-onramp, install-profile, toolchain, and output
wording fixes. Do not delete agents or archive process history without a
separate human-approved deletion/retention decision.
