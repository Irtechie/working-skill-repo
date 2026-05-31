param(
  [string]$Root = "."
)

$ErrorActionPreference = "Stop"

function Resolve-RepoPath {
  param([string]$Base, [string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return $Path
  }
  return (Join-Path $Base $Path)
}

function Invoke-JsonCommand {
  param([string]$Command)
  $output = Invoke-Expression $Command | Out-String
  $exitCode = $LASTEXITCODE
  $json = $null
  if ($output.Trim()) {
    $json = $output | ConvertFrom-Json
  }
  return [pscustomobject]@{
    exit_code = $exitCode
    json = $json
    raw = $output
  }
}

$repoRoot = (Resolve-Path $Root).Path
$runDir = $null

try {
  $adapterCommand = "powershell -ExecutionPolicy Bypass -File scripts\skill-eval-run-codex.ps1 -FixtureId tiny-typo-fix -DryRun -KeepRun -Json"
  $adapter = Invoke-JsonCommand $adapterCommand
  if ($adapter.exit_code -ne 0 -or -not $adapter.json.ok) {
    throw "Codex dry-run adapter failed before manifest selftest."
  }

  $run = @($adapter.json.runs)[0]
  $runDir = Split-Path "$($run.manifest_path)" -Parent
  $resultPath = "$($run.result_path)"
  $manifestPath = "$($run.manifest_path)"
  $runId = "$($run.run_id)"

  $goodCommand = "powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1 -ResultPath `"$resultPath`" -RequiredRunId `"$runId`" -ManifestPath `"$manifestPath`" -Json"
  $good = Invoke-JsonCommand $goodCommand
  if ($good.exit_code -ne 0 -or -not $good.json.ok) {
    throw "Skill eval rejected a valid manifest."
  }

  $badManifestPath = Join-Path $runDir "manifest-bad.json"
  $manifest = Get-Content $manifestPath -Raw | ConvertFrom-Json
  $manifest.protected_files[0].sha256 = ("0" * 64)
  $manifest | ConvertTo-Json -Depth 8 | Set-Content -Path $badManifestPath -Encoding UTF8

  $badCommand = "powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1 -ResultPath `"$resultPath`" -RequiredRunId `"$runId`" -ManifestPath `"$badManifestPath`" -Json"
  $bad = Invoke-JsonCommand $badCommand
  if ($bad.exit_code -eq 0 -or $bad.json.ok) {
    throw "Skill eval accepted a manifest with a tampered protected-file SHA."
  }

  Write-Host "Skill eval manifest selftest: valid manifest passed; tampered fixture SHA failed."
} finally {
  if ($runDir -and (Test-Path $runDir)) {
    $evalRunRoot = (Resolve-Path (Resolve-RepoPath $repoRoot ".atv/eval-runs")).Path
    $resolvedRunDir = (Resolve-Path $runDir).Path
    if ($resolvedRunDir.StartsWith($evalRunRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
      Remove-Item -LiteralPath $resolvedRunDir -Recurse -Force
    }
  }
}

exit 0
