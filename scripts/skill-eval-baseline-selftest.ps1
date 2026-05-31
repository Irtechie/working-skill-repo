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
$tempRoot = Resolve-RepoPath $repoRoot ".atv/eval-baseline-selftest"

try {
  New-Item -ItemType Directory -Force -Path $tempRoot | Out-Null
  $baselinePath = Join-Path $tempRoot "baseline.json"
  $resultRoot = Join-Path $tempRoot "results"
  New-Item -ItemType Directory -Force -Path $resultRoot | Out-Null

  foreach ($file in Get-ChildItem (Resolve-RepoPath $repoRoot "evals/skill-eval/selftest") -Filter "*.json") {
    Copy-Item -LiteralPath $file.FullName -Destination (Join-Path $resultRoot $file.Name)
  }

  $updateCommand = "powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1 -ResultRoot `"$resultRoot`" -BaselinePath `"$baselinePath`" -UpdateBaseline -Json"
  $update = Invoke-JsonCommand $updateCommand
  if ($update.exit_code -ne 0 -or -not $update.json.ok) {
    throw "Failed to create valid baseline."
  }

  $compareCommand = "powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1 -ResultRoot `"$resultRoot`" -BaselinePath `"$baselinePath`" -Json"
  $compare = Invoke-JsonCommand $compareCommand
  if ($compare.exit_code -ne 0 -or -not $compare.json.ok) {
    throw "Unchanged baseline comparison failed."
  }

  $resultPath = Join-Path $resultRoot "pass-tiny-typo-fix.json"
  $result = Get-Content $resultPath -Raw | ConvertFrom-Json
  $result.actual.proof = @()
  $result | ConvertTo-Json -Depth 8 | Set-Content -Path $resultPath -Encoding UTF8

  $regression = Invoke-JsonCommand $compareCommand
  if ($regression.exit_code -eq 0 -or $regression.json.ok) {
    throw "Baseline comparison accepted a proof regression."
  }

  Copy-Item -LiteralPath (Resolve-RepoPath $repoRoot "evals/skill-eval/selftest/fail-proof-missing.json") -Destination (Join-Path $resultRoot "fail-proof-missing.json") -Force
  $negativeResultPath = Join-Path $resultRoot "fail-proof-missing.json"
  $negative = Get-Content $negativeResultPath -Raw | ConvertFrom-Json
  $negative.actual.proof = @("git diff --check")
  $negative.trace.commands = @("git diff --check")
  $negative | ConvertTo-Json -Depth 8 | Set-Content -Path $negativeResultPath -Encoding UTF8

  $negativeRegression = Invoke-JsonCommand $compareCommand
  if ($negativeRegression.exit_code -eq 0 -or $negativeRegression.json.ok) {
    throw "Baseline comparison accepted a negative-fixture regression."
  }

  Write-Host "Skill eval baseline selftest: baseline update passed; unchanged compare passed; proof regression failed; negative-fixture regression failed."
} finally {
  if (Test-Path $tempRoot) {
    $atvRoot = (Resolve-Path (Resolve-RepoPath $repoRoot ".atv")).Path
    $resolvedTemp = (Resolve-Path $tempRoot).Path
    if ($resolvedTemp.StartsWith($atvRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
      Remove-Item -LiteralPath $resolvedTemp -Recurse -Force
    }
  }
}

exit 0
