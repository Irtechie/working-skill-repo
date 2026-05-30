param(
  [string]$Root = ".",
  [string]$ConfigPath = "config/skill-quality.json",
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
    [System.Collections.Generic.List[object]]$List,
    [string]$Fixture,
    [string]$Message
  )
  $List.Add([pscustomobject]@{
    fixture = $Fixture
    message = $Message
  })
}

function Has-Property {
  param($Object, [string]$Name)
  return ($Object -and ($Object.PSObject.Properties.Name -contains $Name))
}

function Get-Score {
  param($Signals)
  $verificationWeights = @{
    none = 0
    unit = 1
    integration = 2
    functional = 3
    full = 4
  }
  $duration = [double]$Signals.expected_duration_hours
  $durationScore = if ($duration -le 0.5) { 0 } elseif ($duration -le 2) { 1 } elseif ($duration -le 8) { 2 } else { 4 }
  $externalScore = if ($Signals.external_dependency) { 1 } else { 0 }
  $userScore = if ($Signals.user_visible) { 1 } else { 0 }
  return [int]$Signals.subsystem_count +
    [int]$Signals.uncertainty +
    [int]$Signals.data_auth_security_risk +
    [int]$Signals.rollback_difficulty +
    $externalScore +
    $userScore +
    $durationScore +
    [int]$verificationWeights["$($Signals.verification_surface)"]
}

function Get-Tier {
  param([int]$Score, $Rubric)
  if ($Score -le [int]$Rubric.small_max) {
    return "small"
  }
  if ($Score -le [int]$Rubric.standard_max) {
    return "standard"
  }
  return "large"
}

$repoRoot = (Resolve-Path $Root).Path
$configFullPath = Resolve-RepoPath $repoRoot $ConfigPath
$config = Get-Content $configFullPath -Raw | ConvertFrom-Json
$fixtureRoot = Resolve-RepoPath $repoRoot $config.route_complexity.fixture_root
$issues = [System.Collections.Generic.List[object]]::new()
$results = [System.Collections.Generic.List[object]]::new()

if (-not (Test-Path $fixtureRoot)) {
  throw "Missing route sizing fixture root: $($config.route_complexity.fixture_root)"
}

$fixtures = Get-ChildItem $fixtureRoot -Filter "*.json" | Sort-Object Name
if ($fixtures.Count -eq 0) {
  throw "No route sizing fixtures found in $($config.route_complexity.fixture_root)"
}

$requiredTop = @("id", "platforms", "prompt", "repo_state", "expected", "complexity_signals", "guards")
$requiredExpected = @("route", "complexity_tier", "max_user_questions", "artifacts", "proof")
$requiredSignals = @("subsystem_count", "uncertainty", "user_visible", "data_auth_security_risk", "external_dependency", "verification_surface", "rollback_difficulty", "expected_duration_hours")

foreach ($file in $fixtures) {
  $fixture = Get-Content $file.FullName -Raw | ConvertFrom-Json
  foreach ($field in $requiredTop) {
    if (-not (Has-Property $fixture $field)) {
      Add-Issue $issues $file.Name "Missing top-level field '$field'."
    }
  }
  if (Has-Property $fixture "expected") {
    foreach ($field in $requiredExpected) {
      if (-not (Has-Property $fixture.expected $field)) {
        Add-Issue $issues $file.Name "Missing expected field '$field'."
      }
    }
  }
  if (Has-Property $fixture "complexity_signals") {
    foreach ($field in $requiredSignals) {
      if (-not (Has-Property $fixture.complexity_signals $field)) {
        Add-Issue $issues $file.Name "Missing complexity_signals field '$field'."
      }
    }
  }

  if (Has-Property $fixture "platforms") {
    foreach ($platform in $fixture.platforms) {
      if (@($config.route_complexity.allowed_platforms) -notcontains $platform) {
        Add-Issue $issues $file.Name "Unknown platform '$platform'."
      }
    }
  }

  if ((Has-Property $fixture "expected") -and (Has-Property $fixture.expected "route") -and (@($config.route_complexity.allowed_routes) -notcontains $fixture.expected.route)) {
    Add-Issue $issues $file.Name "Unknown expected route '$($fixture.expected.route)'."
  }

  $canScore = (Has-Property $fixture "complexity_signals") -and
    (Has-Property $fixture "expected") -and
    (Has-Property $fixture.expected "complexity_tier") -and
    (($requiredSignals | Where-Object { -not (Has-Property $fixture.complexity_signals $_) }).Count -eq 0)

  if ($canScore) {
    $score = Get-Score $fixture.complexity_signals
    $tier = Get-Tier $score $config.route_complexity.rubric
    if ($tier -ne $fixture.expected.complexity_tier) {
      Add-Issue $issues $file.Name "Expected tier '$($fixture.expected.complexity_tier)' but rubric computed '$tier' (score $score)."
    }

    $results.Add([pscustomobject]@{
      id = $fixture.id
      route = $fixture.expected.route
      tier = $tier
      score = $score
      guards = @($fixture.guards) -join ","
    })
  }
}

$allGuards = @($fixtures | ForEach-Object { (Get-Content $_.FullName -Raw | ConvertFrom-Json).guards } | ForEach-Object { $_ })
foreach ($guard in @("over-planning", "under-planning")) {
  if ($allGuards -notcontains $guard) {
    Add-Issue $issues "suite" "Missing required guard '$guard'."
  }
}

$result = [pscustomobject]@{
  ok = ($issues.Count -eq 0)
  fixture_count = $fixtures.Count
  issues = $issues
  results = $results
}

if ($Json) {
  $result | ConvertTo-Json -Depth 6
} else {
  Write-Host "Route complexity eval: $($fixtures.Count) fixtures, $($issues.Count) issues"
  foreach ($row in $results) {
    Write-Host "$($row.id): route=$($row.route) tier=$($row.tier) score=$($row.score) guards=$($row.guards)"
  }
  foreach ($issue in $issues) {
    Write-Host "ERROR [$($issue.fixture)] $($issue.message)"
  }
}

if ($issues.Count -gt 0) {
  exit 1
}

exit 0
