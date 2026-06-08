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

- Verify load-bearing callers, callees, and impact edges against source files
  before writing `PROJECT.md`, architecture docs, or todos.
- If graphify coverage is sparse, fall back to normal source inspection for
  unsupported areas and record the limitation in
  `docs/context/memory-maintenance.md`.
- Do not duplicate dense graph output in `PROJECT.md`. Keep `PROJECT.md` as a
  router and point structural traversal to named `graph_route` entries.

Suggested `PROJECT.md` row:

```markdown
| Subsystem | Purpose | Orientation | Source |
|---|---|---|---|
| Plugin routing | Chooses host/plugin behavior | graph_route: plugin-routing | `.token-master/graph.json`; verify source edges |
```
