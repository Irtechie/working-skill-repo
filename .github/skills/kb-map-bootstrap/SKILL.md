---
name: kb-map-bootstrap
description: Token-expensive bootstrap skill that deeply indexes a new or existing project and creates the standard KB memory layout. Use when kb-map reports missing or badly stale memory, when entering an existing project without todo.md/docs/context/PROJECT.md, or when the user says "bootstrap this project", "deep map this repo", "build project memory", or "index this app".
argument-hint: "[optional project focus or subsystem hints]"
---

# KB Map Bootstrap

Build parity: after this runs, a fresh session should use the same files whether the app is new or years old.

Use `kb-map` for normal startup. Use this only for missing or badly stale memory.

## Automatic Invocation

When `kb-map`, `AGENTS.md`, or `.github/copilot-instructions.md` detects missing `todo.md` or `docs/context/PROJECT.md`, run this skill immediately. Do not ask the user first unless a non-empty user file would be overwritten or moved.

Run bootstrap in the active project root only. Prefer `git rev-parse --show-toplevel`; otherwise use the current working directory only if it is clearly a project directory. Never bootstrap a drive root such as `E:\`, `~`, `%USERPROFILE%`, `.copilot`, `.codex`, `.agents`, the whole drive, or a sibling repo unless the user explicitly chose that path.

## Graphify Bootstrap Decision

Use graphify or TokenMasterX only when the repo is large or structurally complex
enough for graph routing to pay for itself. Normal `kb-map` lookup remains
doc-first and must not invoke graphify.

Load `../kb-map/references/graph-routing.md` only when the repo may cross the
graph thresholds or the user explicitly asks for graph routing. That reference
contains the preflight command, graphify/TokenMasterX split, `graph_route` row
shape, and evidence rules.

Record the preflight result in `docs/context/memory-maintenance.md` so future
`kb-map` lookups do not repeat the size check every time:

```text
graphify-size-check: 2026-06-03 code_files=<n> project_md_bytes=<n> decision=skip|consider|use reason=<short reason>
```

## Create Layout

```text
todo.md
todo-done.md
docs/context/
  PROJECT.md
  eval-map.md
  architecture/
    README.md
  research/
    README.md
  decisions/
    README.md
  operations/
    README.md
    testing.md
docs/brainstorms/
docs/plans/
docs/handoffs/
  active/
  parked/
  done/
evals/
```

Optional:

```text
docs/context/decisions/starter-kit-deltas.md
docs/context/epics/
docs/context/history/
```

## Workflow

1. **Inventory the repo**
   - Top-level structure, entry points, frameworks, package managers.
   - Build/test/dev commands.
   - Routes, screens, commands, tools, actions, jobs, integrations.
   - Tests, docs, existing TODOs, brainstorms, plans, ADRs, and handoffs.
   - Packaging, installer, updater, release, deployment, and CI workflows.
   - Run the optional code-intel helper when available:

     ```powershell
     $skillDir = Split-Path -Parent $PSCommandPath
     pwsh -NoProfile -File (Join-Path $skillDir 'scripts/code-intel.ps1') -Root <project-root> -Json
     ```

     If executing from a copied prompt where `$PSCommandPath` is unavailable,
     resolve the helper relative to this `SKILL.md`. The helper is static
     code-intelligence, not a mandatory LSP dependency: it reports language
     server availability, symbol samples, likely entry points, largest source
     files, and file-extension inventory. Use it to seed the coverage inventory;
     do not block bootstrap if no language server is installed.

   This is a repo-wide inventory pass. Do not stop after finding the first
   obvious app surface. The point of bootstrap is to discover what major systems
   exist before writing the map.

   Create a temporary coverage inventory while scanning:

   ```text
   discovered area | evidence files | should map? | target doc | reason
   ```

   Every discovered route/screen/command/tool/action/job/integration/runtime
   shell/build-release flow must end up as either:
   - a row in `PROJECT.md`'s subsystem index;
   - a row in `docs/context/architecture/README.md`;
   - folded into a named parent doc with a clear pointer; or
   - explicitly marked "not mapped" with a reason, such as generated, vendor,
     duplicate, dead code, or trivial support file.

2. **Identify subsystems**
   - User-facing workflows.
   - Backend domains.
   - Tool/action layers.
   - Data/storage layers.
   - External integrations.
   - Runtime shells such as desktop, browser, mobile, service, worker, or CLI.
   - Build, package, installer, updater, release, and deployment flows.

   Treat complex operational flows as first-class subsystems. A runtime,
   installer, update, or packaging flow is not "just build stuff" if it spans
   config, CI workflows, packaging scripts, release assets, embedded runtimes,
   startup checks, and update delivery. Create a child architecture doc when one
   parent doc would force a fresh session to rediscover those files.

2.5. **Apply high-risk audit tactics when needed**
   For build/install/runtime, release, packaging, cross-process, or large
   coverage-discovery risk, load `references/bootstrap-audit-tactics.md`.
   It contains dependency-chain checks, eval-surface mapping rules, runtime
   artifact audits, and coverage-discovery tactics.

   Always invoke `kb-eval-map` after the repo inventory has enough evidence to
   identify app patterns, public workflows, existing tests, and likely proof
   surfaces. Do not invent fake tests; record blocked evals instead.

3. **Create or merge memory files**
   - Preserve existing user docs.
   - Do not overwrite non-empty files without reading and merging.
   - Move stale or completed active work out of the active board.
   - Use lowercase kebab-case except `PROJECT.md` and folder `README.md`.

3.5. **Coverage reconciliation**
   - Compare the temporary coverage inventory against `PROJECT.md`,
     `docs/context/architecture/README.md`, and planned subsystem docs.
   - No major discovered area may be silently missing from the map.
   - If a child doc is created, add it to `docs/context/architecture/README.md`.
   - If an area is folded into a parent doc, the parent doc must name it so a
     keyword lookup like `installer`, `MCP`, `actions`, or `auth` can find the
     right pointer without broad repo search.
   - Record unresolved coverage gaps in `docs/context/memory-maintenance.md`
     with type `stale-doc` or `repeated-rediscovery`.

4. **Write `docs/context/PROJECT.md`**
   - Keep it short.
   - Include run/test commands.
   - Include subsystem, research, operation, and active-work pointers.
   - Mark confidence as verified, inferred, or unknown.

5. **Write testing operations**
   - Create/update `docs/context/operations/testing.md`.
   - Include deterministic commands discovered by `go run ./cmd/kbcheck core --list` when present, or equivalent manifest inspection.
   - Note which checks are fast, broad, flaky, external-service dependent, or CI-only.

6. **Write subsystem docs**
   - One concise doc per major subsystem.
   - Parent docs summarize and point to child docs.
   - Include known sharp edges, rejected approaches, and first files to read.
   - For high-risk build/release/runtime flows, include:
     - source of truth and current mode;
     - key scripts/config/workflows;
     - generated artifacts and where they come from;
     - manual or CI steps required to populate release assets;
     - common failure modes and what not to assume.
     - dependency/runtime chain table when artifacts move across build,
       install, first launch, and runtime.

7. **Write board and handoff structure**
   - `todo.md` for active work and handoff queue pointers.
   - `todo-done.md` for compact completion summaries.
   - `docs/handoffs/active/`, `parked/`, and `done/`.
   - If `todo_rules.md`, `todo-rules.md`, `docs/todo-rules.md`, or another separate todo rules file exists, inline current board rules at the top of `todo.md`. Delete the separate rules file only after moving any unique project content into `todo.md` or `docs/context/*`.

8. **Starter-kit deltas**
   - If the app is based on ATV, another starter kit, or a fork, create/update `docs/context/decisions/starter-kit-deltas.md`.

9. **Review**
   - Run `document-review` on `PROJECT.md` and large architecture docs when available.
   - Record unresolved findings in `todo.md` or an active handoff.

10. **Route test**
   - Run a cheap `kb-map lookup` against the new memory.
   - Confirm a fresh session can answer: what this app is, how to run it, how to test it, what work is active, and which subsystem docs to read first.
   - Route-test every area in the coverage inventory marked `should map=yes`.
     Use the area name as the lookup prompt, such as `installer`, `release`,
     `auth`, `workflows`, `actions`, `tools`, `runtime`, or `deployment`.
   - A passing route test means a fresh session can name the exact subsystem doc,
     source-of-truth files, current mode, known sharp edges, and next files to
     read without broad repo search.
   - For high-risk subsystem docs, ask five small-model-grade questions from the
     doc alone, such as where a runtime comes from, what happens offline, what
     differs by architecture, which versions are pinned, and how to validate a
     clean install. If the answers require rediscovery, refine the doc.
   - If any mapped area fails lookup, write or refine the missing index/doc
     before declaring bootstrap complete. Do not pass bootstrap with known
     invisible subsystems.

## Templates

Load `references/memory-templates.md` only when creating missing memory files.
Do not load templates during inventory, coverage discovery, or route testing.

## Output

Finish with:

- Files created or updated.
- Major subsystems discovered.
- Uncertain areas.
- Stale or completed work moved.
- Result of the `kb-map lookup` route test.
