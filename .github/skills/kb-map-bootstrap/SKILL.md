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

Run this preflight before the deep inventory pass:

```powershell
$root = git rev-parse --show-toplevel
$skip = @('.git','.token-master','.codegraph','node_modules','bin','obj','dist','build','venv','__pycache__','.uv-cache','.uv-tools','.uv-bin','.go-cache','.tmp','tmp','runs','graphify-out')
$ext = @('.cs','.fs','.go','.java','.js','.jsx','.kt','.mjs','.php','.ps1','.py','.rb','.rs','.ts','.tsx')
$codeFiles = Get-ChildItem -LiteralPath $root -Recurse -File -ErrorAction SilentlyContinue |
  Where-Object {
    $relative = $_.FullName.Substring($root.Length).TrimStart('\','/')
    $parts = $relative -split '[\\/]'
    -not ($parts | Where-Object { ($skip -contains $_) -or ($_ -like '.venv*') }) -and
    ($ext -contains $_.Extension.ToLowerInvariant())
  }
$codeFileCount = @($codeFiles).Count
```

Do not count generated run folders, dependency installs, tool caches, or copied
benchmark worktrees as repo size. The threshold is about source the agent would
otherwise need to understand, not local artifacts.

Default decision:

- Fewer than 80 code files: skip graphify unless the user explicitly asks for
  graph routing or the task is a hard structural traversal.
- 80 to 199 code files: consider graphify when bootstrap needs caller, callee,
  blast-radius, dependency, or subsystem-boundary discovery.
- 200 or more code files: use graphify during bootstrap when prerequisites are
  available.

When the decision says to use graphify, prefer the cheapest local graph path:

```powershell
if (Get-Command graphify -ErrorAction SilentlyContinue) {
  graphify update .
  New-Item -ItemType Directory -Force -Path (Join-Path $root '.token-master') | Out-Null
  Copy-Item -LiteralPath (Join-Path $root 'graphify-out/graph.json') -Destination (Join-Path $root '.token-master/graph.json') -Force
}
```

Use TokenMasterX setup instead of raw graphify only when the user wants the host
routing agent installed for GHCP or Claude:

```powershell
python <TokenMasterX>/token-master-plugin/skills/token-master/setup.py <repo-root> --host=copilot
```

For Codex bootstrap, raw graphify output is enough to reduce structural
rediscovery. For GHCP or Claude live-token benchmarks, the TokenMasterX routing
agent must be active and verified separately.

If graphify is used, treat graph output as candidate evidence, not final truth.
Verify load-bearing callers, callees, and impact edges against source files
before writing `PROJECT.md`, architecture docs, or todos. If graphify coverage
is sparse, fall back to normal source inspection for unsupported areas and
record the limitation in `docs/context/memory-maintenance.md`.

When graphify is the right orientation surface, do not make `PROJECT.md`
duplicate the graph by listing every source file. Keep `PROJECT.md` as the
router:

- Write a short human-readable subsystem summary.
- Point structural traversal to a named `graph_route`.
- Name only source-of-truth files, entry points, tests, and docs that a fresh
  session should read directly.
- Leave callers, callees, blast radius, and dense dependency fan-out to the
  graph route, with source verification required for load-bearing claims.

Suggested row:

```markdown
| Subsystem | Purpose | Orientation | Source |
|---|---|---|---|
| Plugin routing | Chooses host/plugin behavior | graph_route: plugin-routing | `.token-master/graph.json`; verify source edges |
```

This avoids double orientation: `kb-map` reads `PROJECT.md`, then follows either
the doc/file pointer or the graph pointer selected by the map.

If prerequisites such as `uv` or `graphify` are missing, do not block bootstrap.
Record that graphify was skipped and continue with the normal inventory
workflow.

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

2.5. **Validate dependency and runtime chains**
   Bootstrap is not only a file crawl. For each high-risk subsystem, connect
   what is built, installed, downloaded, configured, and used at runtime.

   Build a compact chain table for installer, release, runtime, integration,
   auth, data, and tool subsystems when they exist:

   ```text
   dependency/artifact | build source | install location | first-launch need | runtime consumer | version/arch pin | validation
   ```

   Check these edges before writing the architecture doc:
   - build-time environment variables vs runtime spawn args and process env;
   - bundled binaries/assets with size, source URL, arch, update path, and owner;
   - install-time vs first-launch vs ongoing runtime dependencies;
   - CI workflow commands vs clean-clone local build commands;
   - hardcoded version strings, DLL names, URLs, and arch labels;
   - requirements/dependency manifests vs real imports and lazy imports;
   - transitional migration code, with sunset criteria or removal trigger;
   - comments/docs that claim a download/bundle path different from code;
   - smoke-test commands that prove embedded runtimes can import/load required
     packages and binaries.

   Flag mismatches in `todo.md` or `docs/context/memory-maintenance.md` instead
   of documenting the happy path as fact. If the subsystem doc cannot answer
   "what must exist on disk after install?", "what may download on first run?",
   and "what is hardcoded vs derived?", keep auditing before route-test passes.

2.55. **Map eval surfaces**
   Invoke `kb-eval-map` after the repo inventory has enough evidence to identify
   app patterns, public workflows, existing tests, and likely proof surfaces.
   Run this step even when Step 2.5 finds broken dependency/runtime chains. The
   eval map should record what can be evaluated now and which checks are blocked
   by the broken chain instead of disappearing because bootstrap found a problem.

   `kb-eval-map` owns:
   - classifying the repo as website, internal/corporate website, API, CLI,
     LLM/agent app, skill repo, docs/process repo, mobile/native, or mixed;
   - detecting existing Playwright/Cypress/pytest/Vitest/Pester/API/CLI/eval
     harnesses;
   - creating or updating `docs/context/eval-map.md`;
   - scaffolding one real smoke eval only when the primary workflow is known and
     safe to exercise;
   - documenting eval gaps when credentials, sessions, production-only systems,
     destructive actions, or unclear product intent block scaffolding;
   - updating `docs/context/operations/testing.md` with canonical eval commands
     when they exist.

   Do not let bootstrap invent fake tests. If the primary workflow is unclear,
   `kb-eval-map` asks the user what the repo is supposed to prove and records a
   todo if that answer is unavailable.

2.6. **Audit tactics for build/install/runtime subsystems**
   Use these tactical checks when a subsystem builds, ships, downloads, installs,
   or launches runtime artifacts.

   1. **Cross-reference environment variables**
      Grep env vars set in build scripts, CI, package config, and installer
      scripts. Grep runtime process spawns for the same vars and diff resolved
      paths. Catches bundled assets that runtime ignores and re-downloads.

   2. **Inventory bundled blobs**
      Build a table: filename, size, source URL/recipe, arch, install
      destination, update mechanism, owner. Every binary, generated asset,
      vendored runtime, or external blob must have a row; unowned blobs are
      removal candidates.

   3. **Map hardcoded versions**
      Grep build/release/runtime files for literal versions, DLL names, URLs,
      filenames, and arch labels. Map each literal to its source-of-truth
      constant; duplicated literals are future silent failures on version bumps.

   4. **Check dependency deadweight**
      For every production dependency manifest entry, grep shipped runtime code
      for imports, including lazy imports inside functions. Classify as
      shipped+used, shipped+unused, or test-only-but-shipped; ignore archives,
      examples, and tools that are not shipped.

   5. **Audit the architecture matrix**
      When builds target multiple architectures, grep for
      `x64|amd64|arm64|x86_64|aarch64`. Classify each match as arch-aware,
      hardcoded, or should-derive; flag asymmetric handling across files.

   6. **Compare comments to lifecycle code**
      Grep comments/docstrings near build, install, update, and launch code for
      `downloads|fetches|installs|bundles|requires|ships`. Verify the adjacent
      code does that verb now; drifted comments mislead future sessions.

   7. **Require embedded-runtime smoke tests**
      Any shipped language/runtime embed needs a build-time command proving it
      can import/load expected packages and binaries. Document the exact command;
      if none exists, propose one and add a `todo.md` item.

   8. **Write first-clean-clone runbooks**
      Subsystem docs must include literal clean-machine commands, expected cold
      and warm durations, cache locations, and network endpoints. A vague
      "run install/build" instruction fails this check.

   9. **Expire optional, disabled, or legacy code**
      For each optional/disabled/legacy tag, record what reactivates it, what
      removes it, and who owns the decision. If any answer is unknown, create a
      `todo.md` item instead of leaving silent debt.

   10. **Small-model doc test**
      Feed only the subsystem doc to a small model and ask five operational
      triage questions about lookup paths, artifact size, arch differences,
      offline behavior, and version-pin coupling. If it cannot answer from the
      doc alone, add detail before route-test passes.

   Bugs discovered by these checks should be recorded in a "Confirmed Bugs Found
   & Fixed" table in the subsystem doc so future sessions know what was audited
   and do not re-debate resolved findings.

2.7. **Coverage discovery tactics**
   Use these checks before declaring the coverage inventory complete. The goal
   is to map concepts, not merely top-level folders.

   First compare the `code-intel.ps1` entry point hints, symbol samples, largest
   files, and language-server availability against the manual inventory. Any
   likely entry point or large source file from the helper must be represented
   in the coverage inventory, folded into a named parent row, or explicitly
   skipped with a reason. Treat language-server output as a precision boost when
   available, not as a source of truth that replaces reading the code.

   1. **Descend into substantial child directories**
      If a child directory has more than about 30 source files, inspect it as a
      candidate subsystem group instead of folding it into its parent. Large
      child dirs often hide the real architecture below a generic folder name.

   2. **Cluster cross-cutting concepts**
      Sweep for common concepts such as auth, token, credential, session,
      storage, telemetry, browser/runtime control, IPC, settings, cache, queue,
      worker, model, tool, and integration. Hits across multiple layers usually
      need a concept doc even when no single directory owns them.

   3. **Pattern-match filenames**
      Group files by shared prefixes/suffixes such as `*_map`, `*_bridge`,
      `telemetry_*`, `*_adapter`, `*_client`, or `*_provider`. Any group with
      three or more shipped files is a candidate subsystem or shared-pattern doc.

   4. **Mine repo memory and prior docs**
      Read existing memories, AGENTS files, root READMEs, and project notes for
      named architecture topics. Diff mentioned topics against architecture docs;
      a remembered subsystem with no doc is a coverage gap.

   5. **Enumerate user-visible surfaces**
      For apps with routes, pages, screens, commands, playbooks, or workflows,
      list each surface and its backend/tool entry point. Missing UI-route maps
      make small models rediscover the product one page at a time.

   6. **Run hotspot discovery**
      Produce a top-20 list of largest source files and largest source
      directories. Files over about 800 lines or directories over about 50 files
      must be documented or explicitly justified as not architectural.

   7. **Use a must-cover checklist**
      Check whether the repo has auth, storage, IPC, browser/HTTP, telemetry,
      settings, build/install, LLM/model, background worker, and integration
      concerns. If present, each needs a doc, parent pointer, or skip reason.

   8. **Detect cross-process concerns**
      Search for matching API surfaces across runtime boundaries such as
      desktop/web, frontend/backend, language/runtime bridges, worker/server, or
      native/web. Cross-process flows deserve docs because no one file explains
      them.

   9. **Check coverage ratio**
      Compare architecture-doc count to source-file count by major section. A
      rough ratio above 25:1 in a substantial area suggests undermapping unless
      the parent doc has strong child pointers and skip reasons.

   10. **Test small-model triage by subsystem**
      Ask whether a small model could triage failures for auth, runtime control,
      telemetry, storage, install, and top user workflows from the KB alone. Any
      "no" means the map is not done for that subsystem.

   11. **Respect small native glue**
      Small files touching OS APIs, native embeds, security storage, COM,
      browser/runtime embedding, device APIs, or process injection can be
      critical even when short. Size alone is not a reason to skip native glue.

   12. **Record known-unknowns**
      When a meaningful file, command, page, or workflow is found but not
      documented, add it to `PROJECT.md` or `docs/context/memory-maintenance.md`
      as a known-unknown with a reason and revisit trigger.

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

### `todo.md`

```markdown
# Todo

## Rules

**Conventions:** these match the KB skill spec. Keep them inline here; do not split into `todo_rules.md`, `todo-rules.md`, or any separate rules file.

**This file is the single source of truth for active work** — not chat history, not session SQL, and not stale manifests. Any agent should be able to claim a row from here cold.

**Status markers** (applied to individual rows):

| Marker | Meaning |
|--------|---------|
| ⬜ pending | Ready when blockers clear |
| 🔧 in_progress | Agent claimed and actively working |
| ✅ done | Complete and verified — move summary to `todo-done.md` promptly |
| 🔒 blocked | Cannot proceed — explain in `## Blocked` with `Depends on:` |
| ⊘ skipped | Intentionally skipped with reason |
| 🛑 human-required | Needs human action (HITL) — also surface under `## Human Required` |

**Section icons** (section headers, not row markers):

- 💡 Feature Ideas — not yet brainstormed; a human promotes to active
- 📋 Queued Improvements — approved but not yet planned
- 🧊 Parked / Cold Storage — intentionally out of bounds today; never auto-runs, human-promote only
- 🛑 Human Required — items only a person can complete (logins, approvals, decisions)
- 📝 Work Log — short dated entries for cross-session visibility

**Task metadata** lines under a row when relevant: `Task ID:`, `Ready: yes|no`, `Depends on:`, `Discovered from:`, `Validation:`.

**Promotion rules:**
- Newly discovered work goes to 🧊 Parked / Cold Storage first. Never auto-execute from there.
- Items stalled because another agent, dependency, tool failure, or missing input must finish first go to 🔒 Blocked, not Parked.
- Human-required work must not be silently folded into generic blocked notes.
- Detailed handoffs live under `docs/handoffs/`; link them here instead of pasting content.
- Refresh cold or parked work older than 72 hours before execution.
- Keep this file current and small. When all active todos are done, check the handoff queue.

## Objective
## Current Focus
## Current Truth
## Active Work
## Handoff Queue
| Handoff | Status | Route | Created | Stale Check | Link |
|---|---|---|---|---|---|
## Human Required
## Parked / Cold Storage
## Blocked
## Work Log
```

### `todo-done.md`

```markdown
# Completed Work

> Archive of completed items from `todo.md`. Most recent at top.

## YYYY-MM-DD
- <feature or slice group> — <compact outcome, important proof, commit/link if available>
```

### `PROJECT.md`

```markdown
# Project Map

Bootstrap: YYYY-MM-DD
Bootstrap confidence: verified|mixed|rough

## What This Is
## How To Run
## How To Test
## Current Architecture
## Subsystem Index
| Area | Read This | Use When | Confidence |
|---|---|---|---|
## Current Work Pointers
## Known Sharp Edges
## Research Index
## Do Not Repeat
## Maintenance Notes
```

## Output

Finish with:

- Files created or updated.
- Major subsystems discovered.
- Uncertain areas.
- Stale or completed work moved.
- Result of the `kb-map lookup` route test.
