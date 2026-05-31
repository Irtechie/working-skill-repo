param(
  [string]$Root = ".",
  [string]$ConfigPath = "config/skill-quality.json",
  [string]$ResultRoot = "",
  [string]$ResultPath = "",
  [string]$RequiredRunId = "",
  [string]$ManifestPath = "",
  [string]$BaselinePath = "",
  [switch]$UpdateBaseline,
  [switch]$Json
)

$ErrorActionPreference = "Stop"

function Resolve-RepoPath {
  param([string]$Base, [string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return $Path
  }
  return (Join-Path $Base $Path)
}

function Has-Property {
  param($Object, [string]$Name)
  return ($Object -and ($Object.PSObject.Properties.Name -contains $Name))
}

function Add-Issue {
  param(
    [System.Collections.Generic.List[object]]$List,
    [string]$ResultId,
    [string]$Message
  )
  $List.Add([pscustomobject]@{
    result = $ResultId
    message = $Message
  })
}

function Get-FileSha256 {
  param([string]$Path)
  return (Get-FileHash $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

function Normalize-Text {
  param([string]$Value)
  if ($null -eq $Value) {
    $Value = ""
  }
  return (($Value -replace "\s+", " ").Trim().ToLowerInvariant())
}

function Test-RunManifest {
  param(
    [string]$RepoRoot,
    [string]$Path,
    [string]$RequiredRunId
  )

  $issues = [System.Collections.Generic.List[object]]::new()
  if (-not $Path) {
    return $issues
  }

  $fullPath = Resolve-RepoPath $RepoRoot $Path
  if (-not (Test-Path $fullPath)) {
    Add-Issue $issues "manifest" "ManifestPath does not exist: $Path"
    return $issues
  }

  $manifest = Get-Content $fullPath -Raw | ConvertFrom-Json
  $manifestRunId = if (Has-Property $manifest "run_id") { "$($manifest.run_id)" } else { "" }
  if ($RequiredRunId -and ($manifestRunId -ne $RequiredRunId)) {
    Add-Issue $issues "manifest" "Expected manifest run_id '$RequiredRunId' but got '$manifestRunId'."
  }

  if (-not (Has-Property $manifest "protected_files")) {
    Add-Issue $issues "manifest" "Manifest is missing protected_files."
    return $issues
  }

  foreach ($entry in @($manifest.protected_files)) {
    $role = if (Has-Property $entry "role") { "$($entry.role)" } else { "<missing-role>" }
    $pathValue = if (Has-Property $entry "path") { "$($entry.path)" } else { "" }
    $expectedHash = if (Has-Property $entry "sha256") { "$($entry.sha256)".ToLowerInvariant() } else { "" }
    if (-not $pathValue) {
      Add-Issue $issues "manifest" "Protected file entry '$role' is missing path."
      continue
    }
    if (-not $expectedHash) {
      Add-Issue $issues "manifest" "Protected file '$pathValue' is missing sha256."
      continue
    }

    $protectedPath = Resolve-RepoPath $RepoRoot $pathValue
    if (-not (Test-Path $protectedPath)) {
      Add-Issue $issues "manifest" "Protected file '$pathValue' is missing."
      continue
    }

    $actualHash = Get-FileSha256 $protectedPath
    if ($actualHash -ne $expectedHash) {
      Add-Issue $issues "manifest" "Protected file '$pathValue' changed for role '$role': expected $expectedHash but got $actualHash."
    }
  }

  return $issues
}

function New-BaselineRow {
  param($Row)
  return [pscustomobject]@{
    file = "$($Row.file)"
    fixture_id = "$($Row.fixture_id)"
    expected_result = "$($Row.expected_result)"
    actual_result = "$($Row.actual_result)"
    issue_count = [int]$Row.issue_count
  }
}

function New-SkillEvalBaseline {
  param(
    [string]$RepoRoot,
    [string]$ResultRoot,
    [string]$ResultPath,
    $Rows
  )

  return [pscustomobject]@{
    schema_version = 1
    generated_at = (Get-Date).ToString("o")
    result_root = $ResultRoot
    result_path = $ResultPath
    result_count = @($Rows).Count
    rows = @($Rows | ForEach-Object { New-BaselineRow $_ })
  }
}

function Compare-SkillEvalBaseline {
  param(
    $Baseline,
    $Rows
  )

  $issues = [System.Collections.Generic.List[object]]::new()
  $currentByFile = @{}
  foreach ($row in @($Rows)) {
    $currentByFile["$($row.file)"] = $row
  }

  foreach ($baselineRow in @($Baseline.rows)) {
    $file = "$($baselineRow.file)"
    if (-not $currentByFile.ContainsKey($file)) {
      Add-Issue $issues "baseline" "Baseline row '$file' is missing from current results."
      continue
    }

    $current = $currentByFile[$file]
    if ("$($current.fixture_id)" -ne "$($baselineRow.fixture_id)") {
      Add-Issue $issues $file "Fixture changed from '$($baselineRow.fixture_id)' to '$($current.fixture_id)'."
    }
    if ("$($current.expected_result)" -ne "$($baselineRow.expected_result)") {
      Add-Issue $issues $file "Expected result changed from '$($baselineRow.expected_result)' to '$($current.expected_result)'."
    }
    if ("$($baselineRow.expected_result)" -eq "fail" -and "$($current.actual_result)" -ne "fail") {
      Add-Issue $issues $file "Negative fixture regressed from fail to '$($current.actual_result)'."
    }
    if ("$($baselineRow.actual_result)" -eq "pass" -and "$($current.actual_result)" -ne "pass") {
      Add-Issue $issues $file "Result regressed from pass to '$($current.actual_result)'."
    }
    if ([int]$current.issue_count -gt [int]$baselineRow.issue_count) {
      Add-Issue $issues $file "Issue count regressed from $($baselineRow.issue_count) to $($current.issue_count)."
    }
  }

  return $issues
}

function Test-ContainsAny {
  param($Items, [string]$Expected)
  $needle = Normalize-Text $Expected
  foreach ($item in @($Items)) {
    if ((Normalize-Text "$item").Contains($needle)) {
      return $true
    }
  }
  return $false
}

function Test-TraceRules {
  param(
    [System.Collections.Generic.List[object]]$Issues,
    [string]$ResultId,
    $Result
  )

  if (-not (Has-Property $Result "trace_rules")) {
    return
  }

  $rules = $Result.trace_rules
  $checks = @(
    @{ field = "required_files_read"; trace = $Result.trace.files_read; label = "required file read" },
    @{ field = "required_commands"; trace = $Result.trace.commands; label = "required command" },
    @{ field = "required_tools"; trace = $Result.trace.tools; label = "required tool" }
  )

  foreach ($check in $checks) {
    if (Has-Property $rules $check.field) {
      foreach ($expected in @($rules.($check.field))) {
        if (-not (Test-ContainsAny $check.trace "$expected")) {
          Add-Issue $Issues $ResultId "Missing trace rule evidence for $($check.label) containing '$expected'."
        }
      }
    }
  }

  $forbiddenChecks = @(
    @{ field = "forbidden_files_read"; trace = $Result.trace.files_read; label = "forbidden file read" },
    @{ field = "forbidden_commands"; trace = $Result.trace.commands; label = "forbidden command" },
    @{ field = "forbidden_tools"; trace = $Result.trace.tools; label = "forbidden tool" }
  )

  foreach ($check in $forbiddenChecks) {
    if (Has-Property $rules $check.field) {
      foreach ($forbidden in @($rules.($check.field))) {
        if (Test-ContainsAny $check.trace "$forbidden") {
          Add-Issue $Issues $ResultId "Forbidden trace rule matched $($check.label) containing '$forbidden'."
        }
      }
    }
  }
}

function Get-FixtureMap {
  param([string]$FixtureRoot)
  $map = @{}
  foreach ($file in (Get-ChildItem $FixtureRoot -Filter "*.json" | Sort-Object Name)) {
    $fixture = Get-Content $file.FullName -Raw | ConvertFrom-Json
    if (Has-Property $fixture "id") {
      $map[$fixture.id] = $fixture
    }
  }
  return $map
}

function Get-ResultFiles {
  param([string]$RepoRoot, [string]$ResultRoot, [string]$ResultPath)
  if ($ResultPath) {
    $full = Resolve-RepoPath $RepoRoot $ResultPath
    if (-not (Test-Path $full)) {
      throw "ResultPath does not exist: $ResultPath"
    }
    return @(Get-Item $full)
  }
  $root = Resolve-RepoPath $RepoRoot $ResultRoot
  if (-not (Test-Path $root)) {
    throw "ResultRoot does not exist: $ResultRoot"
  }
  return @(Get-ChildItem $root -Filter "*.json" | Sort-Object Name)
}

function Test-ClaimCheck {
  param(
    [string]$RepoRoot,
    $Result,
    $Check
  )

  $type = "$($Check.type)"
  $expected = if (Has-Property $Check "expected") { [bool]$Check.expected } else { $true }

  if ($type -eq "file_exists") {
    $path = "$($Check.path)"
    $actual = Test-Path (Resolve-RepoPath $RepoRoot $path)
    return ($actual -eq $expected)
  }

  if ($type -eq "command_ran") {
    $contains = "$($Check.contains)"
    $actual = Test-ContainsAny $Result.trace.commands $contains
    return ($actual -eq $expected)
  }

  if ($type -eq "file_read") {
    $contains = "$($Check.contains)"
    $actual = Test-ContainsAny $Result.trace.files_read $contains
    return ($actual -eq $expected)
  }

  throw "Unknown claim check type '$type'."
}

function Test-Result {
  param(
    [string]$RepoRoot,
    $Result,
    $FixtureMap,
    [string]$RequiredRunId
  )

  $issues = [System.Collections.Generic.List[object]]::new()
  $resultId = if (Has-Property $Result "id") { "$($Result.id)" } else { "<missing-id>" }

  foreach ($field in @("id", "fixture_id", "actual", "trace")) {
    if (-not (Has-Property $Result $field)) {
      Add-Issue $issues $resultId "Missing top-level field '$field'."
    }
  }

  if ($RequiredRunId) {
    if (-not (Has-Property $Result "eval_run_id")) {
      Add-Issue $issues $resultId "Missing eval_run_id sentinel for required run '$RequiredRunId'."
    } elseif ("$($Result.eval_run_id)" -ne $RequiredRunId) {
      Add-Issue $issues $resultId "Expected eval_run_id '$RequiredRunId' but got '$($Result.eval_run_id)'."
    }
  }

  if (-not (Has-Property $Result "fixture_id")) {
    return $issues
  }

  $fixtureId = "$($Result.fixture_id)"
  if (-not $FixtureMap.ContainsKey($fixtureId)) {
    Add-Issue $issues $resultId "Unknown fixture_id '$fixtureId'."
    return $issues
  }

  $fixture = $FixtureMap[$fixtureId]
  $expected = $fixture.expected

  if (-not (Has-Property $Result.actual "route")) {
    Add-Issue $issues $resultId "Missing actual.route."
  } elseif ("$($Result.actual.route)" -ne "$($expected.route)") {
    Add-Issue $issues $resultId "Expected route '$($expected.route)' but got '$($Result.actual.route)'."
  }

  if (-not (Has-Property $Result.actual "user_questions")) {
    Add-Issue $issues $resultId "Missing actual.user_questions."
  } elseif ([int]$Result.actual.user_questions -gt [int]$expected.max_user_questions) {
    Add-Issue $issues $resultId "Expected at most $($expected.max_user_questions) user questions but got $($Result.actual.user_questions)."
  }

  foreach ($artifact in @($expected.artifacts)) {
    if (-not (Test-ContainsAny $Result.actual.artifacts "$artifact")) {
      Add-Issue $issues $resultId "Missing expected artifact evidence containing '$artifact'."
    }
  }

  foreach ($proof in @($expected.proof)) {
    if (-not (Test-ContainsAny $Result.actual.proof "$proof")) {
      Add-Issue $issues $resultId "Missing expected proof evidence containing '$proof'."
    }
  }

  if (-not (Has-Property $Result "trace")) {
    Add-Issue $issues $resultId "Missing trace."
  } else {
    if (-not (Has-Property $Result.trace "files_read")) {
      Add-Issue $issues $resultId "Missing trace.files_read."
    }
    if (-not (Has-Property $Result.trace "commands")) {
      Add-Issue $issues $resultId "Missing trace.commands."
    }
  }

  foreach ($check in @($Result.claim_checks)) {
    try {
      if (-not (Test-ClaimCheck $RepoRoot $Result $check)) {
        $claim = if (Has-Property $check "claim") { "$($check.claim)" } else { "$($check.type)" }
        Add-Issue $issues $resultId "Claim check failed: $claim"
      }
    } catch {
      Add-Issue $issues $resultId $_.Exception.Message
    }
  }

  Test-TraceRules $issues $resultId $Result

  if (Has-Property $Result "claim_artifacts") {
    foreach ($artifact in @($Result.claim_artifacts)) {
      $artifactPath = Resolve-RepoPath $RepoRoot "$artifact"
      if (-not (Test-Path $artifactPath)) {
        Add-Issue $issues $resultId "Missing claim artifact '$artifact'."
        continue
      }
      $claimOutput = powershell -ExecutionPolicy Bypass -File (Join-Path $RepoRoot "scripts/skill-eval-claims.ps1") -Root $RepoRoot -ClaimPath $artifactPath -Json | ConvertFrom-Json
      if (-not $claimOutput.ok) {
        Add-Issue $issues $resultId "Claim artifact failed deterministic verification: $artifact"
      }
    }
  }

  return $issues
}

$repoRoot = (Resolve-Path $Root).Path
$configFullPath = Resolve-RepoPath $repoRoot $ConfigPath
$config = Get-Content $configFullPath -Raw | ConvertFrom-Json
$fixtureRoot = Resolve-RepoPath $repoRoot $config.route_complexity.fixture_root
$defaultResultRoot = if (Has-Property $config "skill_eval") { "$($config.skill_eval.selftest_result_root)" } else { "evals/skill-eval/selftest" }
if (-not $ResultRoot) {
  $ResultRoot = $defaultResultRoot
}

$fixtureMap = Get-FixtureMap $fixtureRoot
$resultFiles = Get-ResultFiles $repoRoot $ResultRoot $ResultPath
$allIssues = [System.Collections.Generic.List[object]]::new()
$rows = [System.Collections.Generic.List[object]]::new()
$selfTestMode = -not $ResultPath

foreach ($issue in @(Test-RunManifest $repoRoot $ManifestPath $RequiredRunId)) {
  $allIssues.Add($issue)
}

foreach ($file in $resultFiles) {
  $result = Get-Content $file.FullName -Raw | ConvertFrom-Json
  $issues = @(Test-Result $repoRoot $result $fixtureMap $RequiredRunId)
  $actualPass = ($issues.Count -eq 0)
  $expectedOutcome = if ((Has-Property $result "expected_result") -and $result.expected_result) { "$($result.expected_result)" } else { "pass" }
  $expectedPass = ($expectedOutcome -eq "pass")

  if ($selfTestMode -and ($actualPass -ne $expectedPass)) {
    if ($expectedPass) {
      Add-Issue $allIssues $file.Name "Self-test expected pass but scorer found issues."
    } else {
      Add-Issue $allIssues $file.Name "Self-test expected failure but scorer passed it."
    }
  }

  foreach ($issue in $issues) {
    if (-not $selfTestMode -or $expectedPass) {
      $allIssues.Add($issue)
    }
  }

  $rows.Add([pscustomobject]@{
    file = $file.Name
    fixture_id = if (Has-Property $result "fixture_id") { $result.fixture_id } else { "" }
    expected_result = $expectedOutcome
    actual_result = if ($actualPass) { "pass" } else { "fail" }
    issue_count = $issues.Count
  })
}

$baselineFullPath = ""
$baselineOutput = $null
$baselineComparisonIssues = [System.Collections.Generic.List[object]]::new()
if ($BaselinePath) {
  $baselineFullPath = Resolve-RepoPath $repoRoot $BaselinePath
  if ($UpdateBaseline) {
    $baselineOutput = New-SkillEvalBaseline $repoRoot $ResultRoot $ResultPath $rows
    $baselineDir = Split-Path $baselineFullPath -Parent
    if ($baselineDir -and -not (Test-Path $baselineDir)) {
      New-Item -ItemType Directory -Force -Path $baselineDir | Out-Null
    }
    $baselineOutput | ConvertTo-Json -Depth 8 | Set-Content -Path $baselineFullPath -Encoding UTF8
  } else {
    if (-not (Test-Path $baselineFullPath)) {
      Add-Issue $baselineComparisonIssues "baseline" "BaselinePath does not exist: $BaselinePath"
    } else {
      $baseline = Get-Content $baselineFullPath -Raw | ConvertFrom-Json
      foreach ($issue in @(Compare-SkillEvalBaseline $baseline $rows)) {
        $baselineComparisonIssues.Add($issue)
      }
    }
  }
  foreach ($issue in $baselineComparisonIssues) {
    $allIssues.Add($issue)
  }
}

$output = [pscustomobject]@{
  ok = ($allIssues.Count -eq 0)
  result_count = $resultFiles.Count
  selftest = $selfTestMode
  results = $rows
  baseline = if ($BaselinePath) {
    [pscustomobject]@{
      path = $baselineFullPath
      updated = [bool]$UpdateBaseline
      issue_count = $baselineComparisonIssues.Count
    }
  } else { $null }
  issues = $allIssues
}

if ($Json) {
  $output | ConvertTo-Json -Depth 8
} else {
  $mode = if ($selfTestMode) { "selftest" } else { "results" }
  Write-Host "Skill eval: $($resultFiles.Count) $mode files, $($allIssues.Count) issues"
  foreach ($row in $rows) {
    Write-Host "$($row.file): fixture=$($row.fixture_id) expected=$($row.expected_result) actual=$($row.actual_result) issues=$($row.issue_count)"
  }
  foreach ($issue in $allIssues) {
    Write-Host "ERROR [$($issue.result)] $($issue.message)"
  }
  if ($BaselinePath) {
    if ($UpdateBaseline) {
      Write-Host "Baseline updated: $baselineFullPath"
    } else {
      Write-Host "Baseline compared: $baselineFullPath issues=$($baselineComparisonIssues.Count)"
    }
  }
}

if ($allIssues.Count -gt 0) {
  exit 1
}

exit 0
