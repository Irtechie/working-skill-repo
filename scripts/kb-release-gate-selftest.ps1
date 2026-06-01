param()

$ErrorActionPreference = "Stop"

function Write-TextFile {
  param([string]$Path, [string]$Text)
  $dir = Split-Path $Path -Parent
  if ($dir -and -not (Test-Path $dir)) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
  }
  $Text | Set-Content -Path $Path -Encoding Ascii
}

function New-TestRepo {
  param([bool]$FailGoTest)

  $root = Join-Path ([System.IO.Path]::GetTempPath()) "kb-release-gate-$([guid]::NewGuid())"
  New-Item -ItemType Directory -Force -Path $root | Out-Null
  Write-TextFile (Join-Path $root "go.mod") "module releasefixture`n"
  $testBody = if ($FailGoTest) {
    @"
package releasefixture

import "testing"

func TestFixture(t *testing.T) { t.Fatal("expected failure") }
"@
  } else {
    @"
package releasefixture

import "testing"

func TestFixture(t *testing.T) {}
"@
  }
  Write-TextFile (Join-Path $root "fixture_test.go") $testBody
  Write-TextFile (Join-Path $root "scripts/skill-sync-report.ps1") "param()`nexit 0`n"
  git -C $root init | Out-Null
  git -C $root config user.email test@example.com | Out-Null
  git -C $root config user.name "Release Gate Test" | Out-Null
  git -C $root add . | Out-Null
  git -C $root commit -m "fixture" | Out-Null
  return $root
}

function Invoke-GoGate {
  param([string]$Root, [string]$Profile)

  $oldPreference = $ErrorActionPreference
  $oldNativePreference = $null
  if (Get-Variable -Name PSNativeCommandUseErrorActionPreference -Scope Global -ErrorAction SilentlyContinue) {
    $oldNativePreference = $Global:PSNativeCommandUseErrorActionPreference
    $Global:PSNativeCommandUseErrorActionPreference = $false
  }
  $ErrorActionPreference = "Continue"
  try {
    $output = & go run .\cmd\kbcheck $Profile --root $Root --json 2>&1
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $oldPreference
    if ($null -ne $oldNativePreference) {
      $Global:PSNativeCommandUseErrorActionPreference = $oldNativePreference
    }
  }
  $jsonStart = ($output | Select-String -Pattern '^\{' | Select-Object -First 1).LineNumber
  if (-not $jsonStart) {
    throw "Go gate did not emit JSON: $($output -join "`n")"
  }
  $jsonLines = @($output | Select-Object -Skip ($jsonStart - 1))
  $jsonEnd = ($jsonLines | Select-String -Pattern '^\}\s*$' | Select-Object -Last 1).LineNumber
  if (-not $jsonEnd) {
    throw "Go gate emitted incomplete JSON: $($output -join "`n")"
  }
  $json = ($jsonLines | Select-Object -First $jsonEnd) -join "`n"
  return [pscustomobject]@{ exit_code = $exitCode; result = ($json | ConvertFrom-Json) }
}

$successRoot = New-TestRepo -FailGoTest $false
$failureRoot = New-TestRepo -FailGoTest $true

try {
  $local = Invoke-GoGate -Root $successRoot -Profile "local-release"
  if ($local.exit_code -ne 0 -or -not $local.result.ok) {
    throw "local-release should pass with successful required checks; exit=$($local.exit_code) ok=$($local.result.ok) results=$(($local.result.results | ConvertTo-Json -Compress -Depth 6))"
  }
  if (@($local.result.results | Where-Object { $_.name -eq "live-codex-ghcp-corpus" }).Count -ne 0) {
    throw "local-release must not include live corpus checks"
  }
  $skipped = @($local.result.results | Where-Object { $_.status -eq "skipped-explicit" })
  if ($skipped.Count -lt 1) {
    throw "local-release should label unavailable optional checks as skipped-explicit"
  }

  $live = Invoke-GoGate -Root $successRoot -Profile "live-release"
  if ($live.exit_code -ne 0 -or -not $live.result.ok) {
    throw "live-release should pass when live corpus runner is explicitly unavailable"
  }
  $liveCorpus = @($live.result.results | Where-Object { $_.name -eq "live-codex-ghcp-corpus" }) | Select-Object -First 1
  if (-not $liveCorpus -or $liveCorpus.status -ne "skipped-explicit") {
    throw "live-release should report unavailable live corpus as skipped-explicit"
  }

  $failed = Invoke-GoGate -Root $failureRoot -Profile "local-release"
  if ($failed.exit_code -eq 0 -or $failed.result.ok) {
    throw "required native core failure should make the gate fail"
  }
  $failedKbCheck = @($failed.result.results | Where-Object { $_.name -eq "kb-check-all" }) | Select-Object -First 1
  if (-not $failedKbCheck -or $failedKbCheck.status -ne "failed") {
    throw "required native core failure was not reported as failed"
  }
} finally {
  Remove-Item -LiteralPath $successRoot -Recurse -Force -ErrorAction SilentlyContinue
  Remove-Item -LiteralPath $failureRoot -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "kb-release-gate selftest passed"
exit 0
