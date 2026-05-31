# Completed Work

> Archive of completed items from `todo.md`. Most recent at top.

## 2026-05-31

- ATV Security Marketplace Promotion - promoted trusted `atv-security` into `E:/agent-marketplace`, added `dependency-vulnerability-osv`, installed the single approved skill into Codex/Copilot/shared agents globals, and synced ATV shipped copies. Proof: marketplace JSON parse, hash equality across source/ATV/marketplace/globals, firebreak + negative selftest, `kb-check -All`, and `git diff --check` passed. OSV Scanner is now locally installed; target repos still need dependency manifests or lockfiles for live scan proof.
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
