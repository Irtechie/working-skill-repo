param(
  [switch]$List,
  [switch]$All
)

$ErrorActionPreference = "Stop"

function Add-Check {
  param([string]$Name, [string]$Command, [string]$Reason)
  [PSCustomObject]@{ Name = $Name; Command = $Command; Reason = $Reason }
}

$checks = New-Object System.Collections.Generic.List[object]

if (Test-Path "package.json") {
  $pkg = Get-Content "package.json" -Raw | ConvertFrom-Json
  $scripts = $pkg.scripts
  foreach ($name in @("lint", "typecheck", "test", "test:unit", "test:integration", "test:e2e", "build")) {
    if ($scripts -and $scripts.PSObject.Properties.Name -contains $name) {
      $runner = if (Test-Path "pnpm-lock.yaml") { "pnpm" } elseif (Test-Path "yarn.lock") { "yarn" } else { "npm run" }
      $cmd = if ($runner -eq "npm run") { "npm run $name" } else { "$runner $name" }
      $checks.Add((Add-Check $name $cmd "package.json script"))
    }
  }
}

if ((Test-Path "pyproject.toml") -or (Test-Path "pytest.ini")) {
  $checks.Add((Add-Check "pytest" "python -m pytest" "Python test config detected"))
}

if (Get-ChildItem -Filter "*.sln" -ErrorAction SilentlyContinue) {
  $sln = (Get-ChildItem -Filter "*.sln" | Select-Object -First 1).Name
  $checks.Add((Add-Check "dotnet-test" "dotnet test `"$sln`"" ".NET solution detected"))
  $checks.Add((Add-Check "dotnet-build" "dotnet build `"$sln`" --no-restore" ".NET solution detected"))
} elseif (Get-ChildItem -Filter "*.csproj" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1) {
  $checks.Add((Add-Check "dotnet-test" "dotnet test" ".NET project detected"))
}

if (Test-Path "Makefile") {
  $checks.Add((Add-Check "make-test" "make test" "Makefile detected"))
}

if ((Test-Path ".github/skills") -and (Test-Path "config/skill-quality.json")) {
  if (Test-Path "scripts/skill-lint.ps1") {
    $checks.Add((Add-Check "skill-lint" "powershell -ExecutionPolicy Bypass -File scripts\skill-lint.ps1" "skill quality config detected"))
  }
  if (Test-Path "scripts/route-complexity-eval.ps1") {
    $checks.Add((Add-Check "route-complexity-eval" "powershell -ExecutionPolicy Bypass -File scripts\route-complexity-eval.ps1" "route complexity eval fixtures detected"))
  }
  if (Test-Path "scripts/skill-eval.ps1") {
    $checks.Add((Add-Check "skill-eval" "powershell -ExecutionPolicy Bypass -File scripts\skill-eval.ps1" "skill eval selftest fixtures detected"))
  }
  if (Test-Path "scripts/skill-eval-run-codex.ps1") {
    $checks.Add((Add-Check "skill-eval-codex-dry-run" "powershell -ExecutionPolicy Bypass -File scripts\skill-eval-run-codex.ps1 -FixtureId tiny-typo-fix -DryRun" "Codex skill eval adapter detected"))
  }
  if (Test-Path "scripts/skill-sync-report.ps1") {
    $checks.Add((Add-Check "skill-sync-report" "powershell -ExecutionPolicy Bypass -File scripts\skill-sync-report.ps1" "skill sync target config detected"))
  }
}

if ($List -or -not $All) {
  $checks | Format-Table -AutoSize
  exit 0
}

foreach ($check in $checks) {
  Write-Host "==> $($check.Name): $($check.Command)"
  Invoke-Expression $check.Command
  if ($LASTEXITCODE -ne 0) {
    Write-Error "check failed: $($check.Name)"
    exit 1
  }
}

exit 0
