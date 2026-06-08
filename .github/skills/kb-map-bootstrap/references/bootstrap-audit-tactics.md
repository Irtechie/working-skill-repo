# Bootstrap Audit Tactics

Use this reference during `kb-map-bootstrap` when a repo has build/install,
runtime, release, packaging, cross-process, or large coverage-discovery risk.

## Dependency And Runtime Chains

Bootstrap is not only a file crawl. For each high-risk subsystem, connect what
is built, installed, downloaded, configured, and used at runtime.

Build a compact chain table for installer, release, runtime, integration, auth,
data, and tool subsystems when they exist:

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

Flag mismatches in `todo.md` or `docs/context/memory-maintenance.md` instead of
documenting the happy path as fact. If the subsystem doc cannot answer "what
must exist on disk after install?", "what may download on first run?", and "what
is hardcoded vs derived?", keep auditing before route-test passes.

## Eval Surface Mapping

Invoke `kb-eval-map` after repo inventory has enough evidence to identify app
patterns, public workflows, existing tests, and likely proof surfaces. Run this
even when runtime-chain checks find broken dependencies. The eval map should
record what can be evaluated now and which checks are blocked.

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

## Build, Install, And Runtime Tactics

Use these checks when a subsystem builds, ships, downloads, installs, or
launches runtime artifacts.

1. **Cross-reference environment variables** — Grep env vars set in build
   scripts, CI, package config, and installer scripts. Grep runtime process
   spawns for the same vars and diff resolved paths.
2. **Inventory bundled blobs** — Build a table: filename, size, source
   URL/recipe, arch, install destination, update mechanism, owner.
3. **Map hardcoded versions** — Grep build/release/runtime files for literal
   versions, DLL names, URLs, filenames, and arch labels. Map each literal to
   its source-of-truth constant.
4. **Check dependency deadweight** — For every production dependency manifest
   entry, grep shipped runtime code for imports, including lazy imports inside
   functions. Classify as shipped+used, shipped+unused, or test-only-but-shipped.
5. **Audit the architecture matrix** — When builds target multiple
   architectures, grep for `x64|amd64|arm64|x86_64|aarch64`; flag asymmetric
   handling.
6. **Compare comments to lifecycle code** — Grep comments/docstrings near
   build, install, update, and launch code for `downloads|fetches|installs|bundles|requires|ships`.
   Verify the adjacent code does that verb now.
7. **Require embedded-runtime smoke tests** — Any shipped language/runtime embed
   needs a build-time command proving it can import/load expected packages and
   binaries.
8. **Write first-clean-clone runbooks** — Subsystem docs must include literal
   clean-machine commands, expected cold/warm durations, cache locations, and
   network endpoints.
9. **Expire optional, disabled, or legacy code** — Record what reactivates it,
   what removes it, and who owns the decision.
10. **Small-model doc test** — Feed only the subsystem doc to a small model and
    ask operational triage questions about lookup paths, artifact size, arch
    differences, offline behavior, and version-pin coupling.

Bugs discovered by these checks should be recorded in a "Confirmed Bugs Found &
Fixed" table in the subsystem doc.

## Coverage Discovery Tactics

Use these checks before declaring coverage inventory complete. The goal is to
map concepts, not merely top-level folders.

First compare `code-intel.ps1` entry point hints, symbol samples, largest files,
and language-server availability against manual inventory. Any likely entry
point or large source file from the helper must be represented in the coverage
inventory, folded into a named parent row, or explicitly skipped with a reason.

1. **Descend into substantial child directories** — If a child directory has
   more than about 30 source files, inspect it as a candidate subsystem group.
2. **Cluster cross-cutting concepts** — Sweep for auth, token, credential,
   session, storage, telemetry, browser/runtime control, IPC, settings, cache,
   queue, worker, model, tool, and integration.
3. **Pattern-match filenames** — Group files by shared prefixes/suffixes such
   as `*_map`, `*_bridge`, `telemetry_*`, `*_adapter`, `*_client`, or
   `*_provider`.
4. **Mine repo memory and prior docs** — Diff named architecture topics in
   memories, AGENTS files, READMEs, and notes against architecture docs.
5. **Enumerate user-visible surfaces** — List each route, page, screen,
   command, playbook, or workflow and its backend/tool entry point.
6. **Run hotspot discovery** — Produce a top-20 list of largest source files
   and directories. Files over about 800 lines or directories over about 50
   files must be documented or explicitly skipped.
7. **Use a must-cover checklist** — Check auth, storage, IPC, browser/HTTP,
   telemetry, settings, build/install, LLM/model, background worker, and
   integration concerns.
8. **Detect cross-process concerns** — Search for matching API surfaces across
   runtime boundaries such as desktop/web, frontend/backend, language/runtime
   bridges, worker/server, or native/web.
9. **Check coverage ratio** — A rough architecture-doc to source-file ratio
   above 25:1 in a substantial area suggests undermapping unless child pointers
   and skip reasons are strong.
10. **Test small-model triage by subsystem** — Ask whether a small model could
    triage failures for auth, runtime control, telemetry, storage, install, and
    top user workflows from the KB alone.
11. **Respect small native glue** — Small files touching OS APIs, native embeds,
    security storage, COM, browser/runtime embedding, device APIs, or process
    injection can be critical even when short.
12. **Record known-unknowns** — When a meaningful file, command, page, or
    workflow is found but not documented, add it to `PROJECT.md` or
    `docs/context/memory-maintenance.md`.
