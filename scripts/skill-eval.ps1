param(
  [string]$Root = ".",
  [string]$ConfigPath = "config/skill-quality.json",
  [string]$ResultRoot = "",
  [string]$ResultPath = "",
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

function Normalize-Text {
  param([string]$Value)
  if ($null -eq $Value) {
    $Value = ""
  }
  return (($Value -replace "\s+", " ").Trim().ToLowerInvariant())
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
    $FixtureMap
  )

  $issues = [System.Collections.Generic.List[object]]::new()
  $resultId = if (Has-Property $Result "id") { "$($Result.id)" } else { "<missing-id>" }

  foreach ($field in @("id", "fixture_id", "actual", "trace")) {
    if (-not (Has-Property $Result $field)) {
      Add-Issue $issues $resultId "Missing top-level field '$field'."
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

foreach ($file in $resultFiles) {
  $result = Get-Content $file.FullName -Raw | ConvertFrom-Json
  $issues = @(Test-Result $repoRoot $result $fixtureMap)
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

$output = [pscustomobject]@{
  ok = ($allIssues.Count -eq 0)
  result_count = $resultFiles.Count
  selftest = $selfTestMode
  results = $rows
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
}

if ($allIssues.Count -gt 0) {
  exit 1
}

exit 0
