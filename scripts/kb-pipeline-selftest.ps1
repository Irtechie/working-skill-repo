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

$repoRoot = (Resolve-Path $Root).Path
$runId = ""

try {
  $startOutput = powershell -ExecutionPolicy Bypass -File scripts\kb-pipeline.ps1 -Start skill-bundle-proof-spike | Out-String
  if ($LASTEXITCODE -ne 0) {
    throw "Pipeline start failed."
  }
  if ($startOutput -notmatch "KB pipeline started: (?<run>[^\r\n]+)") {
    throw "Pipeline start output did not include a run id."
  }
  $runId = $Matches["run"].Trim()
  $runDir = Resolve-RepoPath $repoRoot ".atv/pipeline-runs/$runId"
  foreach ($required in @("run.json", "pipeline.json", "selected-pipeline.md", "protected-files.json", "proof.json", "phase-prompts/map.md")) {
    if (-not (Test-Path (Join-Path $runDir $required))) {
      throw "Pipeline run missing required artifact: $required"
    }
  }

  $statusOutput = powershell -ExecutionPolicy Bypass -File scripts\kb-pipeline.ps1 -Status -RunId $runId | Out-String
  if ($LASTEXITCODE -ne 0 -or $statusOutput -notmatch "Status: started") {
    throw "Pipeline status failed."
  }

  $oldErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  powershell -ExecutionPolicy Bypass -File scripts\kb-pipeline.ps1 -Start does-not-exist 1>$null 2>$null
  $badPipelineExitCode = $LASTEXITCODE
  $ErrorActionPreference = $oldErrorActionPreference
  if ($badPipelineExitCode -eq 0) {
    throw "Pipeline accepted an unknown id."
  }

  Write-Host "KB pipeline selftest: start/status passed; unknown pipeline id failed."
} finally {
  if ($runId) {
    $runDir = Resolve-RepoPath $repoRoot ".atv/pipeline-runs/$runId"
    if (Test-Path $runDir) {
      $runRoot = (Resolve-Path (Resolve-RepoPath $repoRoot ".atv/pipeline-runs")).Path
      $resolvedRunDir = (Resolve-Path $runDir).Path
      if ($resolvedRunDir.StartsWith($runRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
        Remove-Item -LiteralPath $resolvedRunDir -Recurse -Force
      }
    }
  }
}

exit 0
