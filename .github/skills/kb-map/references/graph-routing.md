# Graph Routing Reference

Use graph routing only when repo size or structural complexity makes normal
project-memory lookup too expensive.

## Size Preflight

Run this check during `kb-map-bootstrap` or targeted `kb-map refresh`, not
ordinary lookup:

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
benchmark worktrees as repo size. The threshold is source the agent would
otherwise need to understand.

Default decisions:

- `<80` code files: skip graphify unless explicitly requested or the task is a
  hard structural traversal.
- `80-199` code files: consider graphify when bootstrap needs caller, callee,
  blast-radius, dependency, or subsystem-boundary discovery.
- `>=200` code files: use graphify during bootstrap when prerequisites are
  available.

Record the decision in `docs/context/memory-maintenance.md`:

```text
graphify-size-check: YYYY-MM-DD code_files=<n> project_md_bytes=<n> decision=skip|consider|use reason=<short reason>
```

## Raw Graphify Path

Prefer the cheapest local graph path:

```powershell
if (Get-Command graphify -ErrorAction SilentlyContinue) {
  graphify update .
  New-Item -ItemType Directory -Force -Path (Join-Path $root '.token-master') | Out-Null
  Copy-Item -LiteralPath (Join-Path $root 'graphify-out/graph.json') -Destination (Join-Path $root '.token-master/graph.json') -Force
}
```

If prerequisites such as `uv` or `graphify` are missing, do not block
bootstrap. Record the skip and continue with normal inventory.

## TokenMasterX Path

Use TokenMasterX setup instead of raw graphify only when the user wants the host
routing agent installed for GHCP or Claude:

```powershell
python <TokenMasterX>/token-master-plugin/skills/token-master/setup.py <repo-root> --host=copilot
```

For Codex bootstrap, raw graphify output is enough to reduce structural
rediscovery. For GHCP or Claude live-token benchmarks, TokenMasterX must be
active and verified separately.

## Evidence Rules

Graph output is candidate evidence, not final truth.

- `kb-map` consumes provider-neutral impact packets, not raw provider output.
  Validate packets with `go run ./cmd/kbcheck graph-route --packet <packet.json>`.
- Packet fields must include repository identity, revision, dirty/worktree
  freshness, seed files/symbols, typed edges, source spans, direct/reverse
  impact, tests/docs, confidence, limitations, budget, and fallback mode.
- Evidence classes are distinct: `exact`, `observed`, `structural`,
  `heuristic`, and `llm-inferred`. LLM-inferred edges are never exact.
- Exact-symbol evidence from an optional SCIP/LSP-grade index has precedence
  over structural, heuristic, semantic, or inferred candidates when the index
  matches repository identity, revision, and worktree fingerprint.
- Optional exact-symbol adapters consume already-produced index snapshots only.
  They must not download tools, start language servers, or require a daemon.
- Missing, unsupported, stale, or fingerprint-mismatched exact-symbol indexes
  must return an explicit `file-native` fallback packet rather than failing base
  lookup or claiming provider authority.
- Load-bearing exact or observed edges require source spans. Missing repo
  identity, revision, freshness, or source provenance fails closed.
- Stale or unsupported providers must set an explicit non-authoritative fallback
  such as `file-native`; they must not produce authoritative packets.
- Verify load-bearing callers, callees, and impact edges against source files
  before writing `PROJECT.md`, architecture docs, or todos.
- If graphify coverage is sparse, fall back to normal source inspection for
  unsupported areas and record the limitation in
  `docs/context/memory-maintenance.md`.
- Do not duplicate dense graph output in `PROJECT.md`. Keep `PROJECT.md` as a
  router and point structural traversal to named `graph_route` entries.

## Traversal Recipes

Select a task-shaped recipe before expanding structural or flow evidence. The
protected recipe fixture lives at `evals/graph-routing/traversal-recipes.json`
and covers:

- API changes: implementation, config, tests, docs, and downstream consumers.
- Bugs: observed calls first, then static calls, refs, tests, and config.
- Deletions: reverse refs, config registrations, build/test/doc edges.
- Security flow: bounded source, guard, sink, and supporting call/config edges.
- UI behavior: component, route, state, generated output, tests, and docs.

Recipe output must preserve typed edges such as `CALLS_STATIC`,
`CALLS_OBSERVED`, `REFERENCES`, `IMPLEMENTS`, `OVERRIDES`, `READS_CONFIG`,
`GENERATES`, `BUILDS`, `TESTS`, and `DOCUMENTS`. Inferred provider edges remain
heuristic and cannot claim exact confidence.

Graphify output is optional structural evidence. It must match repository
identity, revision, and worktree fingerprint. Missing output, stale output,
multigraph collapse risk, dynamic dispatch gaps, generated code, reflection,
dependency injection, aliases, and missing library models must stay visible as
fallbacks or limitations.

Suggested `PROJECT.md` row:

```markdown
| Subsystem | Purpose | Orientation | Source |
|---|---|---|---|
| Plugin routing | Chooses host/plugin behavior | graph_route: plugin-routing | `.token-master/graph.json`; verify source edges |
```
