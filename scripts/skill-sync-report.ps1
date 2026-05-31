param(
  [string]$Root = ".",
  [string]$ConfigPath = "config/skill-quality.json",
  [switch]$VerboseOptional,
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

function Get-SkillHash {
  param([string]$SkillRoot, [string]$SkillName)
  $skillDir = Join-Path $SkillRoot $SkillName
  if (-not (Test-Path $skillDir)) {
    return $null
  }
  $lines = Get-ChildItem $skillDir -Recurse -File |
    Sort-Object FullName |
    ForEach-Object {
      $relative = $_.FullName.Substring($skillDir.Length + 1) -replace "\\", "/"
      $hash = (Get-FileHash $_.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
      "$relative`t$hash"
    }
  return Get-StringSha256 ($lines -join "`n")
}

$repoRoot = (Resolve-Path $Root).Path
$configFullPath = Resolve-RepoPath $repoRoot $ConfigPath
$config = Get-Content $configFullPath -Raw | ConvertFrom-Json
$sourceTarget = @($config.sync_targets | Where-Object { $_.classification -eq "source" }) | Select-Object -First 1

if (-not $sourceTarget) {
  throw "No source sync target configured."
}

$sourceRoot = $sourceTarget.path
if (-not (Test-Path $sourceRoot)) {
  $sourceRoot = Resolve-RepoPath $repoRoot $sourceRoot
}
if (-not (Test-Path $sourceRoot)) {
  throw "Source skill root not found: $($sourceTarget.path)"
}

$skillNames = Get-ChildItem $sourceRoot -Directory | Sort-Object Name | Select-Object -ExpandProperty Name
$rows = [System.Collections.Generic.List[object]]::new()
$requiredIssues = [System.Collections.Generic.List[object]]::new()

foreach ($skill in $skillNames) {
  $sourceHash = Get-SkillHash $sourceRoot $skill
  foreach ($target in $config.sync_targets) {
    if ($target.classification -eq "source") {
      continue
    }

    $targetRoot = $target.path
    if (-not (Test-Path $targetRoot)) {
      $targetRoot = Resolve-RepoPath $repoRoot $targetRoot
    }

    $status = "missing-target"
    $targetHash = $null
    $suggestion = "target path unavailable"

    if (Test-Path $targetRoot) {
      $targetHash = Get-SkillHash $targetRoot $skill
      if (-not $targetHash) {
        $status = if ($target.required) { "missing-required" } else { "missing-optional" }
        $suggestion = "copy source -> target if this skill is meant to ship there"
      } elseif ($targetHash -eq $sourceHash) {
        $status = "match"
        $suggestion = "none"
      } else {
        $status = if ($target.required) { "drift-required" } else { "drift-optional" }
        $suggestion = "review diff before copying source -> target or merging target -> source"
      }
    }

    $row = [pscustomobject]@{
      skill = $skill
      target = $target.id
      classification = $target.classification
      required = [bool]$target.required
      status = $status
      source_hash = $sourceHash.Substring(0, 12)
      target_hash = if ($targetHash) { $targetHash.Substring(0, 12) } else { "" }
      suggestion = $suggestion
    }
    $rows.Add($row)

    if ($target.required -and $status -ne "match") {
      $requiredIssues.Add($row)
    }
  }
}

$result = [pscustomobject]@{
  ok = ($requiredIssues.Count -eq 0)
  required_issues = $requiredIssues.Count
  rows = $rows
}

if ($Json) {
  $result | ConvertTo-Json -Depth 6
} else {
  Write-Host "Skill sync report: $($rows.Count) comparisons, $($requiredIssues.Count) required issues"
  $rows |
    Group-Object target, status |
    Sort-Object Name |
    ForEach-Object {
      Write-Host "$($_.Name): $($_.Count)"
    }

  if ($requiredIssues.Count -gt 0) {
    Write-Host ""
    Write-Host "Required issues:"
    foreach ($issue in $requiredIssues) {
      Write-Host "ERROR [$($issue.target)] $($issue.skill): $($issue.status) source=$($issue.source_hash) target=$($issue.target_hash) :: $($issue.suggestion)"
    }
  }

  $optionalIssues = @($rows | Where-Object { -not $_.required -and $_.status -ne "match" })
  if ($optionalIssues.Count -gt 0) {
    Write-Host ""
    Write-Host "Optional target differences: $($optionalIssues.Count) warning-only differences. Use -VerboseOptional for per-skill rows."
    $optionalIssues |
      Group-Object target, status |
      Sort-Object Name |
      ForEach-Object {
        Write-Host "WARN  $($_.Name): $($_.Count)"
      }
    if ($VerboseOptional) {
      Write-Host ""
      Write-Host "Optional target details:"
      foreach ($issue in $optionalIssues) {
        Write-Host "WARN  [$($issue.target)] $($issue.skill): $($issue.status) source=$($issue.source_hash) target=$($issue.target_hash) :: $($issue.suggestion)"
      }
    }
  }
}

if ($requiredIssues.Count -gt 0) {
  exit 1
}

exit 0
