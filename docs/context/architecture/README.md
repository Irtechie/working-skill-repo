# Architecture Index

## Skill Bundle Surface

Primary files:

- `.github/skills/*/SKILL.md` - portable skills.
- `.github/skills/*/references/*` - lazy-loaded detailed mechanics.
- `.github/skills/*/scripts/*` - deterministic helper scripts.
- `.github/agents/*.agent.md` - reviewer and specialist agents dispatched by review/planning skills.
- `AGENTS.md` - Codex/agent repo contract and sync workflow.
- `.github/copilot-instructions.md` - Copilot always-on repo instructions.
- `README.md` - user-facing install, workflow, and design contract.
- `config/skill-marketplace.json` - private marketplace path, trust model, and
  promotion policy.

## Main Workflow Lanes

| Lane | Entry Skill | Notes |
|---|---|---|
| Startup/routing | `kb-start` | Calls `kb-map`, then routes by task shape. |
| First-principles autonomous task | `kb-task` | Uses map, frames assumptions, delegates to the smallest correct lane. |
| Project memory | `kb-map`, `kb-map-bootstrap`, `kb-memory-review` | Creates and maintains repo-local memory in consuming projects. |
| Requirements/planning | `kb-brainstorm`, `kb-plan`, `kb-gate` | Converts unclear intent into requirements and vertical slices. |
| Execution | `kb-work`, `kb-fix`, `kb-troubleshoot`, `kb-repair` | Executes slices or smaller repair loops with proof gates. |
| Verification setup | `kb-eval-map` | Maps repo-native eval surfaces during bootstrap and documents/scaffolds the right harness for the app pattern. |
| Verification | `kb-check`, `kb-functional-test`, `kb-qa`, `kb-regression-snapshot` | Chooses and runs deterministic proof where available. |
| Completion | `kb-complete`, `kb-review`, `ce-compound`, `learn`, `evolve` | Review, memory, learning, and cleanup. |
| Release | `kb-ship`, `klfg` | Ship readiness or full pipeline orchestration. |

## Private Marketplace

See `docs/context/architecture/private-skill-marketplace.md`.

`E:\agent-marketplace` is the private approved catalog for reusable skills and
pipelines. New learned skills should prove themselves project-local first, then
move into the marketplace only after evidence, review, hash pinning, and human
approval. Public marketplace imports go to quarantine, never directly to global
skill directories.

## Distribution Targets

Working source:

- `E:\working-skill-repo\.github\skills\<skill>\`

Sync targets:

- `C:\Users\marowe\.codex\skills\<skill>\`
- `C:\Users\marowe\.copilot\skills\<skill>\`
- `C:\Users\marowe\.agents\skills\<skill>\`
- `E:\all-the-vibes\.github\skills\<skill>\`
- ATV scaffold/plugin copies when the skill is intentionally shipped there.

Approved reusable catalog:

- `E:\agent-marketplace\skills\<skill>\`
- `E:\agent-marketplace\pipelines\<pipeline>.json`

## Current Coverage Gaps

- `kb-eval-map` is new; consuming repos still need eval maps created during bootstrap or refresh.
