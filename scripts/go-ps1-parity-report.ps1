param(
  [string]$Root = ".",
  [string]$OutputPath = "docs/reports/go-gate-parity-2026-06-01.md"
)

$ErrorActionPreference = "Stop"

function Resolve-RepoPath {
  param([string]$Base, [string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return $Path
  }
  return (Join-Path $Base $Path)
}

function Invoke-Captured {
  param(
    [string]$RepoRoot,
    [string]$FileName,
    [string[]]$Arguments
  )
  $psi = [System.Diagnostics.ProcessStartInfo]::new()
  $psi.FileName = $FileName
  $psi.Arguments = ($Arguments | ForEach-Object {
      if ($_ -match '[\s"]') { '"' + ($_ -replace '"', '\"') + '"' } else { $_ }
    }) -join " "
  $psi.RedirectStandardOutput = $true
  $psi.RedirectStandardError = $true
  $psi.UseShellExecute = $false
  $psi.WorkingDirectory = $RepoRoot

  $process = [System.Diagnostics.Process]::new()
  $process.StartInfo = $psi
  [void]$process.Start()
  $stdout = $process.StandardOutput.ReadToEnd()
  $stderr = $process.StandardError.ReadToEnd()
  $process.WaitForExit()
  return [pscustomobject]@{
    command = "$FileName $($psi.Arguments)".Trim()
    exit_code = $process.ExitCode
    stdout = $stdout
    stderr = $stderr
  }
}

function Get-CheckNames {
  param([string]$Text)
  $names = [System.Collections.Generic.List[string]]::new()
  foreach ($line in ($Text -split "`r?`n")) {
    $trimmed = $line.Trim()
    if (-not $trimmed -or $trimmed -match '^(Name|----)') {
      continue
    }
    if ($trimmed -match '^([A-Za-z0-9:_-]+)\s+') {
      $names.Add($Matches[1])
    }
  }
  return @($names | Sort-Object -Unique)
}

function Format-List {
  param([string[]]$Values)
  if (-not $Values -or $Values.Count -eq 0) {
    return "- none"
  }
  return (($Values | Sort-Object) | ForEach-Object { "- $_" }) -join "`n"
}

$repoRoot = (Resolve-Path $Root).Path
$outputFull = Resolve-RepoPath $repoRoot $OutputPath
$outputDir = Split-Path $outputFull -Parent
if (-not (Test-Path $outputDir)) {
  New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
}

$ps = (Get-Command pwsh -ErrorAction SilentlyContinue | Select-Object -First 1).Source
if (-not $ps) {
  $ps = (Get-Command powershell -ErrorAction Stop | Select-Object -First 1).Source
}
$psArgs = if ([System.IO.Path]::GetFileName($ps).ToLowerInvariant() -like "powershell*") {
  @("-NoProfile", "-ExecutionPolicy", "Bypass", "-File")
} else {
  @("-NoProfile", "-File")
}

$psCheck = ".github/skills/kb-check/scripts/kb-check.ps1"
$psRelease = "scripts/kb-release-gate.ps1"

$psList = Invoke-Captured $repoRoot $ps (@($psArgs) + @($psCheck, "-List"))
$goList = Invoke-Captured $repoRoot "go" @("run", ".\cmd\kbcheck", "core", "--list")
$psAll = Invoke-Captured $repoRoot $ps (@($psArgs) + @($psCheck, "-All"))
$goCore = Invoke-Captured $repoRoot "go" @("run", ".\cmd\kbcheck", "core")
$psLocal = Invoke-Captured $repoRoot $ps (@($psArgs) + @($psRelease, "-Profile", "local-release", "-Root", $repoRoot, "-Json"))
$goLocal = Invoke-Captured $repoRoot "go" @("run", ".\cmd\kbcheck", "local-release", "--json")

$psNames = Get-CheckNames $psList.stdout
$goNames = Get-CheckNames $goList.stdout
$missingInGo = @($psNames | Where-Object { $goNames -notcontains $_ })
$extraInGo = @($goNames | Where-Object { $psNames -notcontains $_ })
$allExitCodes = @($psList.exit_code, $goList.exit_code, $psAll.exit_code, $goCore.exit_code, $psLocal.exit_code, $goLocal.exit_code)
$parityOK = ($missingInGo.Count -eq 0 -and $extraInGo.Count -eq 0 -and @($allExitCodes | Where-Object { $_ -ne 0 }).Count -eq 0)

$report = @"
# Go Gate Parity Report

Generated: $(Get-Date -Format o)
Root: $repoRoot
Result: $(if ($parityOK) { "PASS" } else { "FAIL" })

## Commands

| Surface | Command | Exit |
|---|---|---:|
| PS list | $($psList.command) | $($psList.exit_code) |
| Go list | $($goList.command) | $($goList.exit_code) |
| PS core | $($psAll.command) | $($psAll.exit_code) |
| Go core | $($goCore.command) | $($goCore.exit_code) |
| PS local release | $($psLocal.command) | $($psLocal.exit_code) |
| Go local release | $($goLocal.command) | $($goLocal.exit_code) |

## Check Name Diff

Missing in Go:

$(Format-List $missingInGo)

Extra in Go:

$(Format-List $extraInGo)

## PS Check Names

$(Format-List $psNames)

## Go Check Names

$(Format-List $goNames)

## Removal Gate

PS gate wrapper removal is allowed only when this report says `Result: PASS`.
This report proves Windows parity on this machine. It does not claim macOS or
Linux runtime proof.
"@

$report | Set-Content -Path $outputFull -Encoding UTF8
Write-Host "Wrote parity report: $OutputPath"
if (-not $parityOK) {
  exit 1
}
exit 0
