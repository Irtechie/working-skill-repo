---
name: safe-shell-quoting
description: Run quote-heavy, multiline, or mixed PowerShell/Bash commands from task-owned temporary script files instead of fragile inline command strings. Use for nested quoting, JSON/regex/SQL payloads, multiple interpolation layers, Bash invoked from PowerShell or PowerShell invoked from Bash, or whenever shell escaping has already failed once.
---

# Safe Shell Quoting

Move fragile quoting out of the command line and into a real script file.

## Mandatory workflow

1. Choose the interpreter before writing:
   - PowerShell payload: `.ps1`, executed with `pwsh -NoProfile -NonInteractive -File` or Windows PowerShell `-File` when `pwsh` is unavailable.
   - Bash payload: `.sh`, executed with `bash --noprofile --norc`.
   - Mixed shells: write one file per interpreter. Pass values as positional arguments or environment variables; never embed the inner program in an outer quoted command string.
2. Create one unique direct child of the operating-system temp directory. Use a task-specific prefix such as `codex-shell-<slug>-<random>`; never use the workspace, home directory, drive root, or a shared fixed temp path.
3. Resolve and record the absolute temp base and task directory. Before writing, require all of these:
   - the task directory's parent is the resolved OS temp directory;
   - its leaf name starts with the chosen `codex-shell-<slug>-` prefix;
   - it was created by this run and did not pre-exist.
4. Use the file-edit tool to write the complete payload to a script file inside that directory. Do not construct it with `echo`, `cat`, a heredoc, `Set-Content`, `Out-File`, or another inline shell string—the point is to remove a quoting layer.
5. Read back the written file when exact quoting or byte content matters. Keep secrets in environment variables or existing secret files, never in the generated script.
6. Execute the file by absolute path with an argument array. Do not use `bash -c`, `sh -c`, `pwsh -Command`, `Invoke-Expression`, or `eval` for the payload.
7. Capture stdout, stderr, and the real exit code. Treat a timeout or silent hang as failure; bound execution when the tool supports a timeout.
8. Materialize any requested durable output outside the task directory before cleanup.
9. After the child process exits, clean up in a controller-owned `finally`/`trap` path. Re-resolve and revalidate the exact directory against step 3 before recursive removal. Never delete an unresolved variable, glob, temp base, workspace root, home directory, or drive root.
10. Verify the task directory no longer exists. If cleanup fails, report the exact path and error. Preserve the payload's failure as the primary result while also reporting cleanup failure.

## Controller patterns

PowerShell controller:

```powershell
$ErrorActionPreference = 'Stop'
$tempBase = [IO.Path]::GetFullPath([IO.Path]::GetTempPath()).TrimEnd('\')
$taskName = 'codex-shell-<slug>-' + [guid]::NewGuid().ToString('N')
$taskDir = Join-Path $tempBase $taskName
$created = New-Item -ItemType Directory -Path $taskDir -ErrorAction Stop
try {
    # Write <payload.ps1|payload.sh> with the file-edit tool, then execute by absolute path.
} finally {
    $resolved = [IO.Path]::GetFullPath($created.FullName).TrimEnd('\')
    $item = Get-Item -LiteralPath $resolved -Force -ErrorAction Stop
    if ($item.Parent.FullName.TrimEnd('\') -ne $tempBase -or $item.Name -notlike 'codex-shell-<slug>-*') {
        throw "Refusing unsafe cleanup target: $resolved"
    }
    Remove-Item -LiteralPath $resolved -Recurse -Force -ErrorAction Stop
}
```

Bash controller:

```bash
temp_base="$(realpath "${TMPDIR:-/tmp}")"
prefix="codex-shell-<slug>-"
task_dir="$(mktemp -d "$temp_base/${prefix}XXXXXX")"
cleanup() {
  resolved="$(realpath -e "$task_dir")" || return 1
  case "$resolved" in
    "$temp_base"/"$prefix"*) rm -rf -- "$resolved" ;;
    *) printf 'Refusing unsafe cleanup target: %s\n' "$resolved" >&2; return 1 ;;
  esac
}
trap cleanup EXIT
# Write <payload.sh|payload.ps1> with the file-edit tool, then execute by absolute path.
```

Adapt syntax to the host, but preserve the invariants: unique owned directory,
file-backed payload, argument-based execution, exit capture, validated cleanup,
and absence verification.
