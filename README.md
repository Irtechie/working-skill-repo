# KB Workflow Skills

Portable workflow skills that make coding agents easier to route, resume, and
verify across GitHub Copilot and Codex.

**Status:** actively used, pre-1.0. Expect interface changes while the workflow
and release tooling settle.

KB is for developers who want an agent to do more than generate code. It gives
the agent a file-native way to:

- recover project context without replaying old chats;
- choose a small fix, investigation, plan, or larger initiative proportionally;
- execute vertical slices with explicit dependencies and proof;
- review and deliver without calling unverified work done.

Most users only install the Markdown skills. You do not need Go, the eval
harness, or the marketplace tooling to use KB in another repository.

## Install

Install the runtime skills for Codex, GitHub Copilot, and shared agents:

```shell
npx github:Irtechie/working-skill-repo --target all --profile core
```

Then open your project in a supported agent and invoke:

```text
kb-start "what I want done"
```

`kb-start` is an agent skill, not a standalone shell command. The installer
uses Node once to copy files; Node is not required afterward.

## The Six-Skill Loop

| Skill | Responsibility |
|---|---|
| `kb-start` | Route the request to the smallest workflow that can finish it safely |
| `kb-map` | Build or read repo-local memory for fresh-session recovery |
| `kb-fix` | Handle narrow bugs and contained edits |
| `kb-plan` | Turn clear requirements into verifiable vertical slices |
| `kb-work` | Execute ready slices and prove the integrated result |
| `kb-complete` | Resume any phase and carry reviewed work to its configured endpoint |

The core path is single-agent-first. Specialist review, deep research, browser
proof, and multi-agent execution are loaded only when the task warrants them.

## One Concrete Example

Suppose you ask:

```text
kb-complete "Add CSV export to the invoice list, preserving the current filters"
```

1. `kb-map` loads the invoice UI, API, tests, and current project state.
2. `kb-plan` records the behavior as vertical slices, such as filtered export
   plus user-visible error handling, with a proof command for each.
3. `kb-work` executes dependency-ready slices, runs focused checks, and then
   proves the integrated tree.
4. `kb-complete` performs one semantic review, reruns invalidated proof, and
   follows the project's delivery policy.
5. The result is local-only, a review-ready pull request, or an explicitly
   authorized integration—not a prose claim that the work probably passes.

For solo-owner plan-to-accepted-PR delivery, `p2d <idea>` plans first and then
runs the same gated work, review, PR, and permitted merge path.

## What Is In This Repository

| Surface | Purpose |
|---|---|
| `.github/skills/` | 48 portable skills and their lazy references |
| `.github/agents/` | Optional specialist review and execution profiles |
| `bin/` | Cross-runtime installer and package tests |
| `cmd/`, `internal/` | Go verifier, routing, reconciliation, and benchmark tools |
| `evals/` | Deterministic routing, quality, completion, and regression fixtures |
| `config/` | Release, marketplace, architecture, and eval contracts |
| `packages/` | Optional UniversalUI Skills catalog contribution |
| `docs/` | Curated architecture, operations, research, and maintenance references |

The public repository tracks product source and reusable proof. Local workflow
state—todos, brainstorms, plans, handoffs, and run results—is intentionally
Git-ignored in this source repo. Consuming projects may create those files as
their own local or versioned KB memory.

## Maintainer Proof

The optional Go core checks skill structure, route fixtures, eval scoring,
provider hygiene, marketplace boundaries, and contributor-safe repository
behavior:

```shell
go run ./cmd/kbcheck core
```

Installer and package checks:

```shell
npm test
npm pack --dry-run
```

Only Windows has a recorded machine-verification run today. The runtime skills
are Markdown; the installer is Node-based; Go is maintainer tooling.

## Read More

- [Documentation map](docs/README.md)
- [Skill catalog and responsibilities](docs/context/architecture/skills.md)
- [Workflow architecture](docs/context/architecture/kb-workflow.md)
- [Testing and release proof](docs/context/operations/testing.md)
- [Eval surface](docs/context/eval-map.md)
- [Skill bundle maintenance](docs/context/operations/skill-bundle-maintenance.md)
- [Private marketplace boundary](docs/context/architecture/private-skill-marketplace.md)
- [Graph terminology and provenance research](docs/context/research/2026-07-29-graph-engineering-definition-and-provenance.md)

## Credits

KB grew from ideas in the All-The-Vibes StarterKit and Compound Engineering
workflow, with additional influence from
[Matt Pocock's skills](https://github.com/mattpocock/skills),
[G-Stack](https://github.com/garrytan/gstack),
[ATV-Phoenix](https://github.com/All-The-Vibes/ATV-Phoenix),
[HumanLayer](https://www.humanlayer.dev/blog/advanced-context-engineering), and
[TokenMasterX](https://github.com/shyamsridhar123/TokenMasterX). The linked
[research notes](docs/context/research/README.md) separate prior art from KB's
current implementation claims.

MIT licensed.
