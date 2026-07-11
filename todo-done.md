# Completed Work

> Archive of completed items from `todo.md`. Most recent at top.

## 2026-07-09

- Finish Skill Repo Hardening - removed Phoenix runtime/MCP/global skills,
  preserved CCE as an opt-in adapter, deduplicated ambient Copilot
  instructions, added bounded context-packet and separate telemetry contracts,
  added semantic provider hygiene, corrected loaded-surface accounting,
  reconciled stale goal/manifest/handoff state, and refreshed compound/learning
  memory. Multi-agent review found and resolved eight P1 contract defects.
  Proof: `go test ./cmd/kbcheck`, `go run ./cmd/kbcheck core` (33 checks),
  `go run ./cmd/kbcheck local-release`, manifest/packet/telemetry/provider CLI
  probes, required skill hash sync, and working/ATV `git diff --check`.
- Phoenix routing/slicing absorption - absorbed useful routing, snapshot,
  manifest-proof, run-state, and doctor mechanics without retaining a Phoenix
  runtime or MCP dependency.
- Skill bundle hardening - completed/reviewed the hot-path rent audit, review
  reference guard, native gate extraction, plan archive, and contributor/release
  gate split; the H2 experiment remains intentionally parked.
- RTK-inspired token efficiency - compact passing check output while preserving
  full failure evidence and verbose diagnostics.

## 2026-07-05

- Phoenix Proof Spine Merge - mined ATV-Phoenix for self-healing primitives without replacing KB. Added `kbcheck sense`, `trace-verify`, `accept`, and `learning-adoption`; wired failure-first proof into repair/troubleshoot/goal/work/complete/gate; added model-tier decomposition contracts; preserved scoped-local learning with a measured adoption gate for promotion. Review found and fixed one proof digest issue: behavior-changing check fields such as `timeout_ms` must be part of the digest. Proof: `go test ./cmd/kbcheck`, proof-spine RED->GREEN smoke, learning-adoption ADOPT_ELIGIBLE smoke, `go run ./cmd/kbcheck manifest-contract`, `go run ./cmd/kbcheck local-release`, working/ATV `git diff --check`, `work-to-complete`, and `complete-to-ship` gates passed.

## 2026-07-01

- Live Steering Learning Loop - mined HumanLayer `design-control-loop` for portable mechanics and added KB live steering without importing its runner assumptions. Updated `kb-goal`, `kb-plan`, `kb-complete`, and `learn`; refreshed README, AGENTS, workflow architecture, goal/manifest docs, and added a solution note. Review found and fixed duplicate observation logging risk. Proof: `go run ./cmd/kbcheck core`, `git diff --check`, `work-to-complete` gate, and `complete-to-ship` gate passed.

## 2026-06-10
- Review reference closed contract - added root sweep mode to `kbcheck review-reference-guard`, classified `document-review` overlaps, and filed durable decisions under `docs/context/decisions/`. Proof: `go test ./cmd/kbcheck`, `go run ./cmd/kbcheck review-reference-guard`, `go run ./cmd/kbcheck core`, `go run ./cmd/kbcheck local-release`, `git diff --check`.
- Review reference drift guard - added `kbcheck review-reference-guard`, wired it into `core`, declared shared/forked `ce-review`/`kb-review` reference ownership in `config/skill-quality.json`, and updated memory-maintenance. Proof: `go test ./cmd/kbcheck`, `go run ./cmd/kbcheck review-reference-guard`, `go run ./cmd/kbcheck core`, `go run ./cmd/kbcheck local-release`.

- Skill Bundle Cleanup Audit Follow-up - made `core` contributor-safe by removing release-only sync drift from the core gate, kept sync blocking in `local-release`, lowered the Go directive to `1.22`, changed the installer core profile to install every runtime skill plus baseline review/document agents, reworded successful negative selftests as `correctly rejected`, fixed the optional hook claim in `learn`, refreshed docs/memory, and synced required skill roots plus the changed ATV shipped learn copies. Proof: `go run ./cmd/kbcheck core`, `go run ./cmd/kbcheck local-release`, `go run ./cmd/kbcheck skill-sync-report --verbose-optional`, `git diff --check`, `git -C E:\all-the-vibes diff --check`, installer core/full dry-runs, and old-wording grep passed. Optional ATV scaffold/plugin drift remains warning-only for unrelated pre-existing skill surfaces.

## 2026-06-03

- Cross-Platform Adoption On-Ramp - added the `npx`/Node installer with `core` and `full` profiles, non-destructive backup-on-replace behavior, repo-local install support, and CI proof for Windows/macOS/Linux. Updated front-door docs to make `core` the shallow start path and switched canonical Go commands to the portable `go run ./cmd/kbcheck` form. Proof: installer core/full/repo-local/conflict smoke checks, `go test ./...` with in-repo `GOCACHE`, `go run ./cmd/kbcheck core`, working/ATV `git diff --check`, and required sync report passed.

## 2026-06-01

- Go Validator Full Replacement - ported the remaining skill-repo validators, eval adapters, marketplace promotion/firebreak checks, ATV upstream delta report, pipeline proof, ready-set/scope-lease utilities, release selftests, surface/minimality reports, and sync drift report into native `cmd/kbcheck`; deleted all `.ps1` files from the repo. Proof: `go test ./...`, `go run ./cmd/kbcheck core`, `go run ./cmd/kbcheck local-release --json`, `go run ./cmd/kbcheck ready-set --manifest docs\plans\archive\2026-06\2026-06-01-130-kb-go-validator-full-replacement-manifest.md --json`, `rg --files -g "*.ps1"` returned no files, and `git diff --check` passed.
- Go Validator Port Wave 1 - started the PowerShell-to-Go migration by moving `kb-work` ready-set and scope-lease proof utilities into native `cmd/kbcheck` commands, wiring their selftests into `core`, and deleting the four superseded `.ps1` files. Proof: `go run ./cmd/kbcheck ready-set-selftest`, `go run ./cmd/kbcheck scope-lease-selftest`, and `go test ./...` passed.
- KB Work Swarm Ready Set - inverted `kb-work` from one-active-slice default to bounded ready-set swarming. Added deterministic ready-set and scope-lease proof scripts, wired their selftests into `cmd/kbcheck core`, updated workflow docs, and synced `kb-work` across globals and ATV tracked roots. Proof: `kb-work-ready-set-selftest`, `kb-work-scope-lease-selftest`, `go test ./...`, `go run ./cmd/kbcheck core --list`, `go run ./cmd/kbcheck local-release --json`, working/ATV `git diff --check`, and `skill-sync-report` passed.
- Cold Storage Follow-Through and Go Native Gate Rewrite - added explicit minimality evidence classes, cross-model benchmark fixtures, path-scoped Copilot instructions, native Go `core`/`local-release`/`live-release` orchestration, and Windows parity proof against the old PS wrappers. Removed `kb-check.ps1` and `kb-release-gate.ps1` only after `docs/reports/go-gate-parity-2026-06-01.md` recorded PASS. Fixed a concurrent `kb-pipeline-selftest` run-id collision and removed non-global `media-*` skills from global/working roots. Proof: `go test ./...`, `skill-surface-minimality-selftest`, `cross-model-benchmark-validate`, `kb-release-gate-selftest`, `go-ps1-parity-report`, `go run ./cmd/kbcheck core`, `go run ./cmd/kbcheck local-release --json`, `skill-sync-report`, and `git diff --check` passed.
- Claude Remaining Hardening Follow-Through - unparked the Go core-gate wrapper and trim/deletion queue. Added `cmd/kbcheck` as a thin Go CLI for `core`, `local-release`, and `live-release`, wired Go tests into `kb-check -All`, and tightened minimality classification so protected CE/document-review dependencies do not appear as deletion candidates. No skills or agents were deleted; remaining cold-storage candidates require runtime usage proof or focused trimming. Proof: `go test ./...`, `go build ./cmd/kbcheck`, wrapper help/dry-runs, `skill-surface-minimality-selftest`, `skill-surface-minimality`, `kb-check -All`, `go run ./cmd/kbcheck local-release`, required/optional sync report, and working/ATV `git diff --check` passed.
- Claude Remaining Hardening - added `kb-release-gate.ps1` with local/live profiles, static skill/agent minimality classification, read-only ATV upstream delta reporting, and wired the new selftests into `kb-check -All`. Parked Go wrapper/module and trim/deletion execution remain in cold storage. Also fixed a concurrent baseline-selftest temp-directory collision found during completion. Proof: `kb-check -All`, `kb-release-gate.ps1 -Profile local-release`, `skill-sync-report`, all three protected selftests, and `git diff --check` passed.

## 2026-05-31

- ATV Upstream Resync Correction - treated original `All-The-Vibes/ATV-StarterKit` `upstream/main` as a source to mine, not a mirror target. Backed out transient `lfg`, `slfg`, and `workflows-*` imports because KB replacements own those lanes, preserved the local `atv-security` OSV proof gate, and confirmed useful upstream `ce-review` mechanics are already present locally. Proof: `kb-check -All`, `skill-sync-report`, working/ATV/marketplace `git diff --check`, focused review-skill diff, and direct `Test-Path` checks for removed workflow dirs.
- Future-work Hardening - replaced authored quality-rubric fixtures with computed output-quality scoring from raw result JSON, collapsed `kb-start` routing into one ranked decision list, and made harness subprocesses prefer PowerShell 7 with Windows PowerShell fallback. Proof: `kb-check -All`, `skill-sync-report`, and `git diff --check` passed.
- Marketplace Promotion CLI - added `scripts/promote-marketplace-skill.ps1` so reviewed skills can be promoted, hash-pinned, globally synced, and firebreak-verified in one command. Added `promote-marketplace-skill-selftest` to `kb-check -All` to prove happy path and quarantine refusal.
- ATV Security Marketplace Promotion - promoted trusted `atv-security` into `<agent-marketplace>`, added `dependency-vulnerability-osv`, installed the single approved skill into Codex/Copilot/shared agents globals, and synced ATV shipped copies. Proof: marketplace JSON parse, hash equality across source/ATV/marketplace/globals, firebreak + negative selftest, `kb-check -All`, and `git diff --check` passed. OSV Scanner is now locally installed; target repos still need dependency manifests or lockfiles for live scan proof.
- Warning Quality Cleanup - added missing `argument-hint` frontmatter to older skills, codified `review-mode: local-fallback` for `kb-review`/`kb-complete`, and compacted optional ATV scaffold/plugin sync warnings behind `-VerboseOptional`. Proof: `kb-check -All`, `git diff --check`, and required sync report passed with 0 required sync issues.
- Skill Minimalism and Proof Harness - completed four manifests covering persisted skill-eval baselines, protected verifier SHA manifests, a coded pipeline spike, repo-local landmines, workflow-shape routing, loaded-surface reporting, `kb-first-principles` trim, architecture-deepening lazy lane, TDD/todo lane consolidation, and optional thin ATV scaffold/plugin policy. Review found and fixed one baseline comparison gap for negative fixtures. Proof: `kb-check -All`, `git diff --check`, and required sync report passed with 0 required sync issues.

## 2026-05-30

- Live Cross-Runtime Skill Eval Harness - added GHCP live adapter, Codex/GHCP corpus runner, deterministic trace scoring, transcript claim verification, output-quality rubric selftests, regression reporting, and `kb-eval-map` scaffold negative-validation evidence. Proof: `kb-check -All`, working/ATV `git diff --check`, and required skill hash sync passed.
- Skill Eval Scorer - added `scripts/skill-eval.ps1`, result schema docs, pass/fail self-test fixtures, and `kb-check -All` wiring. Proof: `skill-eval` catches intentional route/proof/claim failures and full `kb-check -All` passed.
- KB Eval Map - added `kb-eval-map`, wired bootstrap to create `docs/context/eval-map.md`, refreshed docs, synced required global/ATV skill copies, and verified with `.\.github\skills\kb-check\scripts\kb-check.ps1 -All` plus `git diff --check` in both touched repos.

## 2026-05-29

- Cross-runtime skill quality - added `config/skill-quality.json`, skill lint, route-complexity fixtures, `kb-check` integration, read-only sync drift reporting, and Codex/GHCP docs. Proof: `.\.github\skills\kb-check\scripts\kb-check.ps1 -All` and `git diff --check` passed.
- Skill repo brutal gap audit - scanned repo structure, skill sizes, sync drift, current official agent docs, and created durable findings in `docs/context/research/2026-05-29-skill-repo-gap-audit.md`. Proof: `git diff --check` passed.
- Initialized repo-local KB memory for the portable skill bundle audit.
