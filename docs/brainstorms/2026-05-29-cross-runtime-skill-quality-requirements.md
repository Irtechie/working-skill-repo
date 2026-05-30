---
date: 2026-05-29
topic: cross-runtime-skill-quality
brainstorm_style: kb-brainstorm
---

# Cross-Runtime Skill Quality

## Problem Frame

This repo has strong KB workflow doctrine, but too much of its quality depends on prose and agent obedience. The skill bundle needs a deterministic quality system that works for both Codex and GitHub Copilot/GHCP so route selection, complexity, verification discipline, and sync drift are testable before skills are propagated.

The affected users are the repo maintainer and future agents entering consuming projects. The core failure to prevent is mis-sized work: tiny fixes becoming ceremony, large work becoming shortcuts, and platform-specific instructions working in Codex but failing or degrading in GHCP.

## Research Summary

**Findings that shaped requirements:**
- Official Codex guidance supports layered `AGENTS.md` discovery and size limits, so repo instructions need to stay lean and verifiable. Affected: R1, R6, R9. Source: https://developers.openai.com/codex/guides/agents-md
- GitHub/Copilot and VS Code support repository-wide instructions, `AGENTS.md`, and path-specific `.instructions.md`, so the repo needs a cross-runtime contract rather than Codex-only assumptions. Affected: R1, R2, R4, R9. Sources: https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/add-custom-instructions/add-repository-instructions and https://code.visualstudio.com/docs/copilot/customization/custom-instructions
- Claude/agent workflow guidance emphasizes verification loops and persistent notes/memory, matching the existing KB direction but exposing the gap that this repo has no runnable skill eval harness. Affected: R3, R5, R7. Source: https://support.claude.com/en/articles/14554000-claude-code-power-user-tips
- Local audit found `kb-check.ps1 -List` returns no checks for this repo, and route-complexity behavior has no executable regression tests. Affected: R3, R5, R7. Source: `docs/context/research/2026-05-29-skill-repo-gap-audit.md`

**Confidence:** High - the local gaps are verified and the external platform constraints are from current official docs.

## Requirements

**Cross-Runtime Contract**
- R1. The quality system must validate skill behavior for both Codex and GHCP, including the instruction surfaces each runtime actually reads: `AGENTS.md`, `.github/copilot-instructions.md`, `.github/skills/**/SKILL.md`, and any future `.github/instructions/*.instructions.md`.
- R2. Each requirement, lint rule, and eval fixture must declare whether it applies to Codex, GHCP, or both.
- R3. The system must avoid platform-only tool assumptions in core checks. When a check depends on a runtime-specific capability, it must degrade to a clear warning or skip with reason instead of pretending the behavior was tested.
- R4. The system must maintain a compatibility matrix that states, for each runtime, which instruction files, skill directories, agents, scripts, and eval modes are actually supported, simulated, warning-only, or unsupported.

**Skill Lint And Structure**
- R5. Add a deterministic skill lint command that validates frontmatter, required sections, output contracts, lazy-reference links, script references, and forbidden unresolved conflict markers.
- R6. Add hot-path size budgets for `SKILL.md` files: warn above the agreed line/token threshold, fail only when an unallowlisted hot-path skill exceeds the hard threshold.
- R7. The lint command must be callable from `kb-check` so standard verification for this repo is not empty.

**Route Complexity Evals**
- R8. Add an eval fixture format for prompt, repo state assumptions, expected route, expected user questions, expected artifacts, expected verification proof, and platform applicability.
- R9. Add route evals for at least these boundary cases: tiny typo/fix, known failing test, unclear broken behavior, bounded feature with "don't ask many questions", stale handoff, broad migration, release/ship flow, and cross-runtime instruction update.
- R10. Add a complexity rubric that scores subsystem count, uncertainty, user-visible surface, data/auth/security risk, external dependencies, verification surface, rollback difficulty, and expected duration.
- R11. Evals must catch both over-planning and under-planning. A passing suite must prove that `kb-start`/`kb-task` choose the smallest correct lane, not just the safest or most ceremonial lane.

**Sync And Drift**
- R12. Add a sync report that compares required skill copies across the working repo, Codex global, Copilot global, shared agents global, and ATV targets.
- R13. The sync report must distinguish required targets from optional/intentional omissions so missing KB skills in ATV scaffold/plugin copies are either explained or flagged.
- R14. The sync report must include hashes and suggested copy direction, but must not overwrite anything by itself.

**Repo Memory And Docs**
- R15. Update repo guidance so this repo may contain KB memory only for skill-bundle maintenance, never consuming-project work.
- R16. Update testing docs so a fresh session knows the canonical quality command and what a clean result means.
- R17. Keep requirements and docs portable: repo-relative paths in generated artifacts, no absolute machine paths except in install/sync target documentation where the path itself is the contract.

## Success Criteria

- `kb-check` reports at least one meaningful check for this repo.
- A single command can run skill lint plus route-complexity evals locally.
- The eval suite includes Codex and GHCP applicability metadata.
- The system fails on at least one deliberately bad fixture for under-sizing and one for over-sizing.
- Sync drift output separates "must fix" from "intentional optional target."
- A fresh session can read `docs/context/PROJECT.md` and find the quality command without broad repo crawling.

## Scope Boundaries

- Do not implement a full model benchmark runner in the first slice; fixture definitions and deterministic lint/eval scaffolding come first.
- Do not rewrite all hot-path skills during this work. Add budgets and allowlists first; refactor bloated skills in follow-up slices.
- Do not auto-copy or overwrite global/ATV skill targets in the first pass.
- Do not create consuming-project memory or handoffs in this repo.

## Key Decisions

- Build a cross-runtime harness, not a Codex-only harness. Rationale: the user explicitly requires Codex and GHCP support; official docs show different but overlapping instruction surfaces. Evidence: user input and platform docs.
- Treat complexity as an eval problem, not a prose problem. Rationale: the existing prose is good but cannot detect regressions. Evidence: `docs/context/research/2026-05-29-skill-repo-gap-audit.md`.
- Make sync reporting read-only first. Rationale: the repo's AGENTS contract requires drift review before overwriting copies. Evidence: `AGENTS.md`.

## Dependencies / Assumptions

- GHCP means GitHub Copilot / Copilot Chat / Copilot coding agent behavior that reads `.github/copilot-instructions.md`, `AGENTS.md`, and repo skill files where supported.
- Codex means this Codex app/CLI skill and `AGENTS.md` workflow.
- PowerShell is acceptable for the first deterministic harness because this repo already has `kb-check.ps1` and runs on Windows.

## Alternatives Considered

- Only extend `kb-start` with more sizing prose: rejected because it does not produce regression proof.
- Start with cross-model benchmark automation: deferred because deterministic fixtures and lint need to exist first.
- Immediately propagate every KB skill into every ATV scaffold/plugin target: rejected until sync target intent is machine-readable.

## Slice Candidates

- Skill lint command - validates `SKILL.md` structure, references, conflict markers, and size budgets.
- Route/complexity eval fixtures - defines fixture schema and initial boundary-case corpus for Codex and GHCP applicability.
- `kb-check` integration - makes the skill repo return meaningful checks from the existing verification entry point.
- Sync drift report - read-only hash report across required and optional distribution targets.
- Memory/docs contract cleanup - clarifies the portable repo's own KB memory exception and canonical quality command.
- Hot-path budget allowlist - records justified exceptions for long skills and flags unapproved bloat.

## Outstanding Questions

### Resolve Before Planning

None.

### Deferred to Planning

- [Affects R5-R11][Technical] Decide exact fixture file format: JSON, YAML, or Markdown with frontmatter.
- [Affects R6][Technical] Set the initial line/token thresholds and allowlist format.
- [Affects R12-R14][Technical] Decide target configuration file path for required vs optional sync targets.

## Next Steps

-> /kb-plan
