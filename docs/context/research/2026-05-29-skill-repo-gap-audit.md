# Skill Repo Gap Audit

Checked: 2026-05-29
Budget mode: standard

## Question

Where is this portable skill repo weakest compared with current agent-workflow best practice, especially around complexity and avoiding the wrong amount of ceremony?

## Findings

### P0: No Deterministic Skill Eval Harness

At audit time, the repo described benchmarking in `README.md`, but had no runnable harness. `kb-check.ps1 -List` found no checks because this repo had no conventional app manifest.

Current status: resolved by the cross-runtime quality work. `kb-check.ps1 -List` now reports `skill-lint`, `route-complexity-eval`, and `skill-sync-report`.

Impact: route quality, complexity, and verification discipline are currently enforced by prose. Prose is not enough for "best on the planet."

Required fix:

- Add a `scripts/skill-lint.ps1` or extend `kb-check.ps1` for this repo.
- Add route eval fixtures: prompt, repo state, expected skill, expected asks, expected proof.
- Make `git diff --check` plus skill lint plus route eval the standard pre-sync gate.

### P0: Route Complexity Is Conceptual, Not Calibrated

`kb-start` has small/feature/large categories, and `kb-plan` enforces vertical slices. Missing is a measurable complexity rubric that catches:

- under-sizing: a multi-subsystem change routed to `kb-fix`;
- over-sizing: a typo or one-line bug routed to `kb-brainstorm`;
- false execution intent: "go straight to work" skipping requirements or manifest;
- stale handoff execution without freshness check.

Required fix:

- Add a scoring table to `kb-start` or a lazy sizing reference.
- Inputs: file-count forecast, subsystem count, user-facing behavior, unknown count, external dependency, data/auth/security risk, verification surface, rollback difficulty, expected elapsed time.
- Outputs: `kb-fix`, `kb-troubleshoot`, `kb-brainstorm`, `kb-plan`, `kb-epic`, or `kb-ship`.
- Add eval prompts for every boundary case.

### P1: Portable Repo Memory Contract Was Self-Contradictory

Before this audit, the repo told agents to invoke `kb-map-bootstrap` when `todo.md` or `docs/context/PROJECT.md` was missing, but the skill bundle itself shipped without those files.

Impact: fresh sessions in this repo had to choose between following KB rules and preserving portable-repo hygiene.

Current mitigation: this audit created repo-local memory for maintaining the skill bundle itself.

Remaining fix:

- Add a clear README/AGENTS exception: this repo may contain memory only for skill-bundle maintenance, never consuming-project work.

### P1: Drift Reporting Is Manual

Hash probe result:

- working repo, `E:\all-the-vibes\.github\skills`, and personal/global installs match for KB skills;
- ATV scaffold/plugin copies are missing many KB skills or carry older inherited skill variants.

Impact: no machine-readable distinction between "intentional not shipped there" and "forgot to propagate."

Required fix:

- Add `scripts/skill-sync-report.ps1`.
- Output: per-skill hashes, target classification, missing intentional/unknown, and suggested copy direction.
- Fail when required targets drift; warn when optional targets differ.

### P1: Skill Bodies Are Still Too Heavy In The Hot Path

Longest skill bodies:

- `kb-brainstorm`: 501 lines
- `kb-plan`: 458 lines
- `kb-map-bootstrap`: 415 lines
- `kb-work`: 413 lines
- `kb-complete`: 367 lines

Some of that is justified. Some should move to lazy references, especially examples and templates that are only needed after route choice.

Required fix:

- Keep route-choice rules, safety gates, and output contracts in `SKILL.md`.
- Move long templates, examples, and phase mechanics into `references/`.
- Add a lint budget: warning over 250 lines; fail over 400 unless allowlisted with reason.

### P1: Verification Is Strong For Apps, Weak For Skills

The workflow is strict about UI/browser/API verification in consuming apps. The skill repo itself lacks equivalent proof:

- no static prompt validation;
- no synthetic repo fixtures;
- no expected-output snapshots;
- no route confusion regression tests;
- no parser for manifest examples embedded in skill docs.

Required fix:

- Create scratch fixture repos for tiny bug, ambiguous bug, UI change, broad refactor, stale handoff, and release flow.
- Run route evals against those fixtures and record pass/fail.
- Add manifest/YAML examples to parser tests so docs cannot rot.

### P2: Copilot Instruction Surface Could Be More Granular

Official VS Code/Copilot docs support path-specific `.instructions.md` files. This repo uses root `.github/copilot-instructions.md` and root `AGENTS.md` only.

Impact: acceptable today, but as the repo grows, always-on instructions may become too broad.

Candidate fix:

- Add `.github/instructions/skills.instructions.md` for `.github/skills/**`.
- Add `.github/instructions/agents.instructions.md` for `.github/agents/**`.
- Keep root instructions terse.

### P2: Source Attribution Is In README But Not Operationalized

README credits Matt Pocock, G-Stack, and kevin-copilot, but the workflow does not yet compare behavior against external benchmark prompts or current official guidance.

Required fix:

- Maintain a research note per major external agent platform behavior when it affects design.
- Refresh on official docs changes or after repeated workflow failures.

## Sources

- OpenAI Codex AGENTS.md docs: https://developers.openai.com/codex/guides/agents-md
- GitHub Copilot repository instructions docs: https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/add-custom-instructions/add-repository-instructions
- VS Code custom instructions docs: https://code.visualstudio.com/docs/copilot/customization/custom-instructions
- Anthropic Claude Code context/prompt docs: https://support.claude.com/en/articles/14553240-give-claude-context-claude-md-and-better-prompts
- Anthropic Claude Code power-user tips: https://support.claude.com/en/articles/14554000-claude-code-power-user-tips

## Applies When

- Editing KB routing, planning, execution, verification, review, or memory skills.
- Syncing skills to global installs or ATV copies.
- Deciding whether a user request is a fix, troubleshoot loop, brainstorm, plan, epic, or ship task.

## Stale When

- Codex/Copilot/Claude instruction discovery semantics change.
- The Codex/GHCP skill quality contract changes materially.
- ATV distribution targets are clarified or removed.

## Rejected Approaches

- Do not solve complexity with more prose only. Add evals.
- Do not make every request `kb-brainstorm`. That is overfitting to safety.
- Do not make "go straight to work" skip planning. It should compress questions, not delete gates.
- Do not remove reviewer agents just because they look bulky; prove dispatch is unused first.

## Impact On Current Project

The next highest-leverage implementation is a deterministic skill-quality harness plus route eval corpus. The second is a tighter complexity rubric in `kb-start` and `kb-task`.
