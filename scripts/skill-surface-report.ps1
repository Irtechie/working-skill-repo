param(
  [string]$Root = ".",
  [string]$SkillRoot = ".github/skills",
  [string]$Route = "",
  [string]$BaselinePath = "",
  [string]$OutputPath = "",
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

function Get-StringSha256 {
  param([string]$Value)
  $bytes = [System.Text.Encoding]::UTF8.GetBytes($Value)
  $sha = [System.Security.Cryptography.SHA256]::Create()
  try {
    return (($sha.ComputeHash($bytes) | ForEach-Object { $_.ToString("x2") }) -join "")
  } finally {
    $sha.Dispose()
  }
}

function Get-SkillInfo {
  param([string]$SkillRootFull, [string]$Name)
  $skillPath = Join-Path (Join-Path $SkillRootFull $Name) "SKILL.md"
  if (-not (Test-Path $skillPath)) {
    return [pscustomobject]@{
      name = $Name
      exists = $false
      lines = 0
      token_estimate = 0
      hash = ""
    }
  }

  $content = Get-Content $skillPath -Raw
  $lineCount = @($content -split "`r?`n").Count
  $tokens = @($content -split "\s+" | Where-Object { $_ }).Count
  return [pscustomobject]@{
    name = $Name
    exists = $true
    lines = $lineCount
    token_estimate = $tokens
    hash = (Get-StringSha256 $content).Substring(0, 12)
  }
}

function Get-DefaultRoutes {
  return [ordered]@{
    base = @("kb-start", "kb-map", "kb-first-principles", "kb-check")
    "kb-plan" = @("kb-start", "kb-map", "kb-plan", "kb-check")
    "kb-work" = @("kb-start", "kb-map", "kb-work", "kb-check")
    "kb-epic" = @("kb-start", "kb-map", "kb-brainstorm", "kb-plan", "kb-epic", "kb-check")
    "kb-complete" = @("kb-start", "kb-map", "kb-complete", "kb-review", "kb-check", "learn", "evolve")
  }
}

$repoRoot = (Resolve-Path $Root).Path
$skillRootFull = Resolve-RepoPath $repoRoot $SkillRoot
$routes = Get-DefaultRoutes
$selectedRoutes = if ($Route) {
  if (-not $routes.Contains($Route)) {
    throw "Unknown route '$Route'. Known routes: $(@($routes.Keys) -join ', ')"
  }
  [ordered]@{ $Route = $routes[$Route] }
} else {
  $routes
}

$routeRows = [System.Collections.Generic.List[object]]::new()
foreach ($routeName in $selectedRoutes.Keys) {
  $skills = @($selectedRoutes[$routeName])
  $skillRows = @($skills | ForEach-Object { Get-SkillInfo $skillRootFull $_ })
  $missing = @($skillRows | Where-Object { -not $_.exists } | Select-Object -ExpandProperty name)
  $routeRows.Add([pscustomobject]@{
    route = $routeName
    skills = $skillRows
    skill_count = $skillRows.Count
    missing = $missing
    total_lines = (@($skillRows | ForEach-Object { [int]$_.lines }) | Measure-Object -Sum).Sum
    token_estimate = (@($skillRows | ForEach-Object { [int]$_.token_estimate }) | Measure-Object -Sum).Sum
    combined_hash = Get-StringSha256 (@($skillRows | ForEach-Object { "$($_.name):$($_.hash)" }) -join "`n")
  })
}

$comparison = $null
if ($BaselinePath) {
  $baselineFull = Resolve-RepoPath $repoRoot $BaselinePath
  if (-not (Test-Path $baselineFull)) {
    throw "BaselinePath does not exist: $BaselinePath"
  }
  $baseline = Get-Content $baselineFull -Raw | ConvertFrom-Json
  $comparisonRows = [System.Collections.Generic.List[object]]::new()
  foreach ($row in $routeRows) {
    $old = @($baseline.routes | Where-Object { $_.route -eq $row.route }) | Select-Object -First 1
    if ($old) {
      $comparisonRows.Add([pscustomobject]@{
        route = $row.route
        line_delta = [int]$row.total_lines - [int]$old.total_lines
        token_delta = [int]$row.token_estimate - [int]$old.token_estimate
        hash_changed = "$($row.combined_hash)" -ne "$($old.combined_hash)"
      })
    }
  }
  $comparison = $comparisonRows
}

$report = [pscustomobject]@{
  generated_at = (Get-Date).ToString("o")
  skill_root = $skillRootFull
  routes = $routeRows
  comparison = $comparison
}

if ($OutputPath) {
  $outputFull = Resolve-RepoPath $repoRoot $OutputPath
  $outputDir = Split-Path $outputFull -Parent
  if ($outputDir -and -not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
  }
  $report | ConvertTo-Json -Depth 8 | Set-Content -Path $outputFull -Encoding UTF8
}

if ($Json) {
  $report | ConvertTo-Json -Depth 8
} else {
  Write-Host "Skill surface report: routes=$($routeRows.Count)"
  foreach ($row in $routeRows) {
    Write-Host "$($row.route): skills=$($row.skill_count) lines=$($row.total_lines) token_estimate=$($row.token_estimate) hash=$($row.combined_hash.Substring(0, 12))"
    if ($row.missing.Count -gt 0) {
      Write-Host "WARN [$($row.route)] missing skills: $(@($row.missing) -join ', ')"
    }
  }
  if ($comparison) {
    foreach ($row in $comparison) {
      Write-Host "COMPARE $($row.route): lines=$($row.line_delta) tokens=$($row.token_delta) hash_changed=$($row.hash_changed)"
    }
  }
}

exit 0
