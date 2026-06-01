param(
  [string]$Root = ".",
  [string]$QualityRoot = "evals/skill-eval/quality",
  [string]$QualityPath = "",
  [string]$FixtureRoot = "evals/route-complexity",
  [int]$MinScore = 3,
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
  param([System.Collections.Generic.List[object]]$List, [string]$CaseId, [string]$Message)
  $List.Add([pscustomobject]@{ case = $CaseId; message = $Message })
}

function Normalize-Text {
  param([string]$Value)
  return (($Value -replace '\s+', ' ').Trim()).ToLowerInvariant()
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

function New-QualityEntry {
  param([int]$Score, [string]$Reason)
  return [pscustomobject]@{
    score = $Score
    judge = "deterministic"
    reason = $Reason
  }
}

function Get-FixtureMap {
  param([string]$RepoRoot, [string]$FixtureRoot)
  $root = Resolve-RepoPath $RepoRoot $FixtureRoot
  if (-not (Test-Path $root)) {
    throw "FixtureRoot does not exist: $FixtureRoot"
  }
  $map = @{}
  Get-ChildItem $root -Filter "*.json" | ForEach-Object {
    $fixture = Get-Content $_.FullName -Raw | ConvertFrom-Json
    if (Has-Property $fixture "id") {
      $map["$($fixture.id)"] = $fixture
    }
  }
  return $map
}

function Measure-ComputedQuality {
  param($Result, $FixtureMap)

  $fixture = $null
  if ((Has-Property $Result "fixture_id") -and $FixtureMap.ContainsKey("$($Result.fixture_id)")) {
    $fixture = $FixtureMap["$($Result.fixture_id)"]
  }

  $actual = if (Has-Property $Result "actual") { $Result.actual } else { $null }
  $route = if ($actual -and (Has-Property $actual "route")) { "$($actual.route)" } else { "" }
  $artifacts = if ($actual -and (Has-Property $actual "artifacts")) { @($actual.artifacts) } else { @() }
  $proof = if ($actual -and (Has-Property $actual "proof")) { @($actual.proof) } else { @() }
  $questions = if ($actual -and (Has-Property $actual "user_questions")) { [int]$actual.user_questions } else { 99 }

  $completenessScore = 1
  $completenessReason = "Missing core actual result fields."
  if ($route -and $artifacts.Count -gt 0 -and $proof.Count -gt 0) {
    $completenessScore = 5
    $completenessReason = "Route, artifacts, and proof are present."
  } elseif ($route -and ($artifacts.Count -gt 0 -or $proof.Count -gt 0)) {
    $completenessScore = 3
    $completenessReason = "Route is present but artifacts or proof are incomplete."
  }

  $maintainabilityScore = 5
  $maintainabilityReason = "Output is concise and concrete."
  $allText = (@($artifacts) + @($proof)) -join " "
  if ($artifacts.Count -gt 8 -or $proof.Count -gt 8) {
    $maintainabilityScore = 2
    $maintainabilityReason = "Output has too many artifact or proof entries for a captured result."
  }
  if ($allText -match '(?i)\b(stuff|things|misc|various|whatever|\?\?\?|todo later)\b') {
    $maintainabilityScore = 1
    $maintainabilityReason = "Output uses vague placeholder wording instead of reviewable artifacts or proof."
  }

  $relevanceScore = 3
  $relevanceReason = "Fixture expectation unavailable; relevance is only partially checked."
  if ($fixture -and (Has-Property $fixture "expected") -and (Has-Property $fixture.expected "route")) {
    if ($route -eq "$($fixture.expected.route)") {
      $relevanceScore = 5
      $relevanceReason = "Route matches the fixture expectation."
    } else {
      $relevanceScore = 1
      $relevanceReason = "Route '$route' does not match expected route '$($fixture.expected.route)'."
    }
  }

  $proofScore = 1
  $proofReason = "No executable proof was recorded."
  if ($proof.Count -gt 0) {
    $proofScore = 3
    $proofReason = "Proof is present but does not cover every expected proof item."
    if ($fixture -and (Has-Property $fixture "expected") -and (Has-Property $fixture.expected "proof")) {
      $missingProof = @($fixture.expected.proof | Where-Object { -not (Test-ContainsAny $proof "$_") })
      if ($missingProof.Count -eq 0) {
        $proofScore = 5
        $proofReason = "Proof covers every expected fixture proof item."
      }
    } else {
      $proofScore = 4
      $proofReason = "Proof is present; no fixture proof list was available."
    }
  }

  $ceremonyScore = 4
  $ceremonyReason = "No clear ceremony mismatch detected."
  if ($fixture -and (Has-Property $fixture "expected")) {
    $maxQuestions = if (Has-Property $fixture.expected "max_user_questions") { [int]$fixture.expected.max_user_questions } else { 99 }
    $tier = if (Has-Property $fixture.expected "complexity_tier") { "$($fixture.expected.complexity_tier)" } else { "" }
    if ($questions -gt $maxQuestions) {
      $ceremonyScore = 1
      $ceremonyReason = "Asked $questions user questions; fixture allows $maxQuestions."
    } elseif (($tier -eq "small") -and (@("kb-brainstorm", "kb-plan", "kb-epic") -contains $route)) {
      $ceremonyScore = 1
      $ceremonyReason = "Small fixture was escalated into planning or brainstorm ceremony."
    } elseif (($tier -eq "large") -and (@("kb-fix", "kb-work") -contains $route)) {
      $ceremonyScore = 1
      $ceremonyReason = "Large fixture was routed into an under-planned lane."
    } elseif ($route -and $questions -le $maxQuestions) {
      $ceremonyScore = 5
      $ceremonyReason = "Route and question count fit the fixture size."
    }
  }

  return [pscustomobject]@{
    completeness = New-QualityEntry $completenessScore $completenessReason
    maintainability = New-QualityEntry $maintainabilityScore $maintainabilityReason
    relevance = New-QualityEntry $relevanceScore $relevanceReason
    proof_quality = New-QualityEntry $proofScore $proofReason
    right_sized_ceremony = New-QualityEntry $ceremonyScore $ceremonyReason
  }
}

function Get-QualityFiles {
  param([string]$RepoRoot, [string]$QualityRoot, [string]$QualityPath)
  if ($QualityPath) {
    $full = Resolve-RepoPath $RepoRoot $QualityPath
    if (-not (Test-Path $full)) {
      throw "QualityPath does not exist: $QualityPath"
    }
    return @(Get-Item $full)
  }
  $root = Resolve-RepoPath $RepoRoot $QualityRoot
  if (-not (Test-Path $root)) {
    throw "QualityRoot does not exist: $QualityRoot"
  }
  return @(Get-ChildItem $root -Filter "*.json" | Sort-Object Name)
}

function Test-QualityCase {
  param($Case, [int]$MinScore, $FixtureMap)
  $issues = [System.Collections.Generic.List[object]]::new()
  $caseId = if (Has-Property $Case "id") { "$($Case.id)" } else { "<missing-id>" }
  $required = @("completeness", "maintainability", "relevance", "proof_quality", "right_sized_ceremony")
  $computed = $false

  $quality = $null
  if (Has-Property $Case "input_result") {
    $quality = Measure-ComputedQuality $Case.input_result $FixtureMap
    $computed = $true
  } elseif (Has-Property $Case "quality") {
    Add-Issue $issues $caseId "Hand-authored quality objects are not accepted; provide input_result so scores are computed."
    return [pscustomobject]@{ issues = $issues; computed = $computed; quality = $null }
  }

  if (-not $quality) {
    Add-Issue $issues $caseId "Missing input_result; quality scores must be computed from captured output."
    return [pscustomobject]@{ issues = $issues; computed = $computed; quality = $null }
  }

  foreach ($dimension in $required) {
    if (-not (Has-Property $quality $dimension)) {
      Add-Issue $issues $caseId "Missing quality dimension '$dimension'."
      continue
    }
    $entry = $quality.$dimension
    foreach ($field in @("score", "judge", "reason")) {
      if (-not (Has-Property $entry $field)) {
        Add-Issue $issues $caseId "Missing '$field' for quality dimension '$dimension'."
      }
    }
    if (Has-Property $entry "judge") {
      $judge = "$($entry.judge)"
      if (@("deterministic") -notcontains $judge) {
        Add-Issue $issues $caseId "Unsupported judge '$judge' for quality dimension '$dimension'; this scorer only implements deterministic judges."
      }
    }
    if (Has-Property $entry "score") {
      $score = [int]$entry.score
      if ($score -lt 0 -or $score -gt 5) {
        Add-Issue $issues $caseId "Score for '$dimension' must be 0-5."
      } elseif ($score -lt $MinScore) {
        Add-Issue $issues $caseId "Quality dimension '$dimension' scored $score, below threshold $MinScore."
      }
    }
    if (Has-Property $Case "expected_quality") {
      $expectedQuality = $Case.expected_quality
      if ((Has-Property $expectedQuality $dimension) -and (Has-Property $expectedQuality.$dimension "score")) {
        $expectedScore = [int]$expectedQuality.$dimension.score
        if ((Has-Property $entry "score") -and ([int]$entry.score -ne $expectedScore)) {
          Add-Issue $issues $caseId "Computed '$dimension' score $($entry.score), expected $expectedScore."
        }
      }
    }
  }

  return [pscustomobject]@{ issues = $issues; computed = $computed; quality = $quality }
}

$repoRoot = (Resolve-Path $Root).Path
$fixtureMap = Get-FixtureMap $repoRoot $FixtureRoot
$qualityFiles = Get-QualityFiles $repoRoot $QualityRoot $QualityPath
$allIssues = [System.Collections.Generic.List[object]]::new()
$rows = [System.Collections.Generic.List[object]]::new()
$selfTestMode = -not $QualityPath

foreach ($file in $qualityFiles) {
  $case = Get-Content $file.FullName -Raw | ConvertFrom-Json
  $caseResult = Test-QualityCase $case $MinScore $fixtureMap
  $issues = @($caseResult.issues)
  $actualPass = ($issues.Count -eq 0)
  $expectedOutcome = if ((Has-Property $case "expected_result") -and $case.expected_result) { "$($case.expected_result)" } else { "pass" }
  $expectedPass = ($expectedOutcome -eq "pass")

  if ($selfTestMode -and ($actualPass -ne $expectedPass)) {
    if ($expectedPass) {
      Add-Issue $allIssues $file.Name "Self-test expected pass but quality scorer found issues."
    } else {
      Add-Issue $allIssues $file.Name "Self-test expected failure but quality scorer passed it."
    }
  }

  foreach ($issue in $issues) {
    if (-not $selfTestMode -or $expectedPass) {
      $allIssues.Add($issue)
    }
  }

  $rows.Add([pscustomobject]@{
    file = $file.Name
    case_id = if (Has-Property $case "id") { $case.id } else { "" }
    expected_result = $expectedOutcome
    actual_result = if ($actualPass) { "pass" } else { "fail" }
    issue_count = $issues.Count
    computed = [bool]$caseResult.computed
  })
}

$output = [pscustomobject]@{
  ok = ($allIssues.Count -eq 0)
  case_count = $qualityFiles.Count
  selftest = $selfTestMode
  min_score = $MinScore
  results = $rows
  issues = $allIssues
}

if ($Json) {
  $output | ConvertTo-Json -Depth 8
} else {
  $mode = if ($selfTestMode) { "selftest" } else { "quality cases" }
  Write-Host "Skill quality eval: $($qualityFiles.Count) $mode, $($allIssues.Count) issues"
  foreach ($row in $rows) {
    Write-Host "$($row.file): case=$($row.case_id) expected=$($row.expected_result) actual=$($row.actual_result) issues=$($row.issue_count)"
  }
  foreach ($issue in $allIssues) {
    Write-Host "ERROR [$($issue.case)] $($issue.message)"
  }
}

if ($allIssues.Count -gt 0) {
  exit 1
}

exit 0
