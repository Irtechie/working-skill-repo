# Completed Work

> Archive of completed items from `todo.md`. Most recent at top.

## 2026-05-30

- Skill Eval Scorer - added `scripts/skill-eval.ps1`, result schema docs, pass/fail self-test fixtures, and `kb-check -All` wiring. Proof: `skill-eval` catches intentional route/proof/claim failures and full `kb-check -All` passed.
- KB Eval Map - added `kb-eval-map`, wired bootstrap to create `docs/context/eval-map.md`, refreshed docs, synced required global/ATV skill copies, and verified with `.\.github\skills\kb-check\scripts\kb-check.ps1 -All` plus `git diff --check` in both touched repos.

## 2026-05-29

- Cross-runtime skill quality - added `config/skill-quality.json`, skill lint, route-complexity fixtures, `kb-check` integration, read-only sync drift reporting, and Codex/GHCP docs. Proof: `.\.github\skills\kb-check\scripts\kb-check.ps1 -All` and `git diff --check` passed.
- Skill repo brutal gap audit - scanned repo structure, skill sizes, sync drift, current official agent docs, and created durable findings in `docs/context/research/2026-05-29-skill-repo-gap-audit.md`. Proof: `git diff --check` passed.
- Initialized repo-local KB memory for the portable skill bundle audit.
