param(
  [string]$Root = ".",
  [string]$FixtureRoot = "evals/cross-model-benchmarks",
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

function Add-Issue {
  param(
    [System.Collections.Generic.List[object]]$Issues,
    [string]$Path,
    [string]$Message
  )
  $Issues.Add([pscustomobject]@{
      path = $Path
      message = $Message
    })
}

$repoRoot = (Resolve-Path $Root).Path
$fixtureRootFull = Resolve-RepoPath $repoRoot $FixtureRoot
$issues = [System.Collections.Generic.List[object]]::new()
$caseCount = 0

if (-not (Test-Path $fixtureRootFull)) {
  throw "Fixture root not found: $FixtureRoot"
}

$files = @(Get-ChildItem $fixtureRootFull -Filter "*.json" -File | Sort-Object Name)
if ($files.Count -eq 0) {
  Add-Issue $issues $FixtureRoot "No benchmark fixture JSON files found."
}

$ids = @{}
foreach ($file in $files) {
  $relative = $file.FullName.Substring($repoRoot.Length + 1) -replace "\\", "/"
  try {
    $fixture = Get-Content $file.FullName -Raw | ConvertFrom-Json
  } catch {
    Add-Issue $issues $relative "Invalid JSON: $($_.Exception.Message)"
    continue
  }

  if ($fixture.schema_version -ne 1) {
    Add-Issue $issues $relative "schema_version must be 1."
  }
  if (-not $fixture.suite) {
    Add-Issue $issues $relative "Missing suite."
  }
  if (-not $fixture.cases -or @($fixture.cases).Count -eq 0) {
    Add-Issue $issues $relative "Missing cases."
    continue
  }

  foreach ($case in @($fixture.cases)) {
    $caseCount++
    foreach ($field in @("id", "category", "prompt", "expected", "forbidden_failures", "scoring")) {
      if (-not $case.PSObject.Properties.Name.Contains($field) -or $null -eq $case.$field -or $case.$field -eq "") {
        Add-Issue $issues $relative "Case missing required field '$field'."
      }
    }

    if ($case.id) {
      if ($ids.ContainsKey($case.id)) {
        Add-Issue $issues $relative "Duplicate case id '$($case.id)'."
      }
      $ids[$case.id] = $true
    }

    if ($case.expected -and -not $case.expected.PSObject.Properties.Name.Contains("must_include")) {
      Add-Issue $issues $relative "Case '$($case.id)' expected block must include must_include."
    }
    if (@($case.forbidden_failures).Count -eq 0) {
      Add-Issue $issues $relative "Case '$($case.id)' must list forbidden_failures."
    }
    if ($case.scoring -and @($case.scoring.PSObject.Properties.Name).Count -eq 0) {
      Add-Issue $issues $relative "Case '$($case.id)' scoring must define at least one dimension."
    }
  }
}

$result = [pscustomobject]@{
  ok = ($issues.Count -eq 0)
  fixture_root = $FixtureRoot
  files = $files.Count
  cases = $caseCount
  issues = $issues
}

if ($Json) {
  $result | ConvertTo-Json -Depth 6
} else {
  Write-Host "Cross-model benchmark fixtures: files=$($files.Count) cases=$caseCount issues=$($issues.Count)"
  foreach ($issue in $issues) {
    Write-Host "ERROR [$($issue.path)] $($issue.message)"
  }
}

if ($issues.Count -gt 0) {
  exit 1
}

exit 0
