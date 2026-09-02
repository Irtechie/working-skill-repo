# Agent Skills Git Distribution

Checked: 2026-05-30
Budget mode: lean

## Question

Are we using Git correctly for this portable skill bundle, or are we creating
avoidable drift by treating local and global installs as peer source trees?

## Findings

Agent Skills are intentionally simple, portable directories: a required
`SKILL.md` plus optional `scripts/`, `references/`, `assets/`, and other bundled
files. The open format is designed for cross-product reuse, with agents loading
only skill metadata at startup and reading the full skill only when relevant.

That implies this repo should treat skills like distributable source packages:

- one canonical source repo owns skill content;
- installed global skill directories are install artifacts, not authoring
  locations;
- sync should be deterministic and hash-verified after merge, not hand-managed
  during implementation.

The official specification says `SKILL.md` must contain YAML frontmatter and
Markdown body, and recommends progressive disclosure: keep the main skill body
bounded, move detail into focused reference files, and reference bundled files
with relative paths. It also states that `scripts/`, `references/`, and
`assets/` are optional directories. This supports the current direction of
moving deterministic checks into scripts and keeping long skill bodies under
pressure.

The best-practices guidance reinforces that skills should be grounded in real
work, not generic LLM output. It recommends extracting skills from hands-on
tasks, project artifacts, execution traces, failures, and version-control
history. It also says every token in an activated `SKILL.md` competes with the
rest of the context window, so skill content should include what the agent would
otherwise miss and cut what the model already knows.

The eval guidance says skill quality needs structured test cases, assertions,
grading evidence, and iteration. It distinguishes verifiable checks from softer
human-review judgments and explicitly recommends verification scripts for
mechanical checks such as valid JSON, row counts, and file existence. That lines
up with this repo's split between model actions and deterministic judges.

The description-optimization guidance treats `description` as the primary
triggering mechanism. It recommends realistic should-trigger and
should-not-trigger query sets, train/validation separation, and avoiding
keyword-level overfitting. That maps directly onto this repo's route-complexity
and live skill-eval fixtures.

The scripts guidance recommends pinning one-off command versions, stating
environment prerequisites, moving complex commands into tested scripts, exposing
`--help`, avoiding interactive prompts, using helpful errors, and emitting
structured output. This supports keeping sync and eval behavior in tested deterministic tooling
rather than relying on agent prose.

## Sources

- [Agent Skills overview](https://agentskills.io/home) - skill folders,
  cross-product reuse, progressive disclosure.
- [Agent Skills specification](https://agentskills.io/specification) - required
  frontmatter, directory structure, optional bundled directories, progressive
  disclosure, file-reference rules, validation.
- [Best practices for skill creators](https://agentskills.io/skill-creation/best-practices)
  - real expertise, project artifacts, execution traces, context economy,
  coherent scope, scripts for repeated logic.
- [Evaluating skill output quality](https://agentskills.io/skill-creation/evaluating-skills)
  - eval cases, assertions, grading evidence, script-based mechanical checks,
  iteration loop.
- [Optimizing skill descriptions](https://agentskills.io/skill-creation/optimizing-descriptions)
  - trigger eval queries, train/validation split, overfitting risk, description
  limits.
- [Using scripts in skills](https://agentskills.io/skill-creation/using-scripts)
  - pinned commands, prerequisites, script interfaces, `--help`, structured
  output.
- [Anthropic engineering: Equipping agents for the real world with Agent Skills](https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills)
  - skills as composable procedural knowledge, scripts as deterministic
  repeatable machinery, progressive disclosure.

## Applies When

- Changing skill content in `.github/skills/**`.
- Syncing to `~/.codex/skills`,
  `~/.copilot/skills`, or
  `~/.agents/skills`.
- Designing the next distribution contract and sync/check scripts.
- Deciding whether a behavior belongs in `SKILL.md`, `references/`, `assets/`,
  or `scripts/`.

## Stale When

- Agent Skills changes the directory/spec/frontmatter contract.
- Codex or Copilot stops using Agent Skills-compatible discovery.
- A package-manager-style installer replaces local global skill directories.

## Rejected Approaches

- Editing global installs directly as source. This creates unreviewed drift and
  makes Git history incomplete.
- Hash-gating line-ending churn without normalization. Byte-level gates are
  useful only when the sync path is byte-stable.
- Adding more skill prose to compensate for missing deterministic scripts. If
  the agent repeats the same parsing, scoring, syncing, or validation work,
  write a script and make the skill invoke it.
- Optimizing skill trigger descriptions only by intuition. Trigger behavior
  needs should-trigger and should-not-trigger prompts with holdout queries.

## Current Adoption

The dated recommendation is now implemented through the Node installer,
`config/skill-quality.json`, and the Go-native `kbcheck skill-sync-report`,
`doctor`, and `local-release` commands. The working repository remains the
canonical source; Codex, Copilot, and shared-agent global roots are deployed
copies. Current commands and target policy live in the architecture and
operations docs, not this note.
