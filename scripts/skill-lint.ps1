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
    [string]$Severity,
    [string]$Path,
    [string]$Message
  )
  $List.Add([pscustomobject]@{
    severity = $Severity
    path = $Path
    message = $Message
  })
}

function Get-FrontmatterValue {
  param([string]$Frontmatter, [string]$Field)
  if (-not $Frontmatter) {
    return $null
  }
  foreach ($line in ($Frontmatter -split "`r?`n")) {
    if ($line -match "^\s*$([regex]::Escape($Field))\s*:\s*(.+?)\s*$") {
      return $Matches[1].Trim().Trim("'").Trim('"')
    }
  }
  return $null
}

$repoRoot = (Resolve-Path $Root).Path
$configFullPath = Resolve-RepoPath $repoRoot $ConfigPath

if (-not (Test-Path $configFullPath)) {
  throw "Missing skill quality config: $ConfigPath"
}

$config = Get-Content $configFullPath -Raw | ConvertFrom-Json
$issues = [System.Collections.Generic.List[object]]::new()
$warnings = [System.Collections.Generic.List[object]]::new()

$skillRoot = Resolve-RepoPath $repoRoot $config.lint.skill_root
if (-not (Test-Path $skillRoot)) {
  Add-Issue $issues "error" $config.lint.skill_root "Skill root is missing."
} else {
  Get-ChildItem $skillRoot -Directory | Sort-Object Name | ForEach-Object {
    $skillName = $_.Name
    $skillPath = $_.FullName
    $skillFile = Join-Path $skillPath "SKILL.md"
    $relativeSkillFile = $skillFile.Substring($repoRoot.Length + 1)

    if (-not (Test-Path $skillFile)) {
      Add-Issue $issues "error" $relativeSkillFile "Missing SKILL.md."
      return
    }

    $content = Get-Content $skillFile -Raw
    $lines = ($content -split "`r?`n").Count
    $frontmatter = $null

    if ($content -match '(?s)^---\s*(.*?)\s*---') {
      $frontmatter = $Matches[1]
    } else {
      Add-Issue $issues "error" $relativeSkillFile "Missing YAML frontmatter."
    }

    foreach ($field in $config.lint.required_frontmatter) {
      if (-not $frontmatter -or $frontmatter -notmatch "(?m)^$([regex]::Escape($field))\s*:") {
        Add-Issue $issues "error" $relativeSkillFile "Missing required frontmatter field '$field'."
      }
    }

    $declaredName = Get-FrontmatterValue $frontmatter "name"
    if ($declaredName) {
      if ($declaredName -ne $skillName) {
        Add-Issue $issues "error" $relativeSkillFile "Frontmatter name '$declaredName' does not match folder '$skillName'."
      }
    }

    if ($config.lint.require_argument_hint -eq "error" -and $frontmatter -notmatch '(?m)^argument-hint\s*:') {
      Add-Issue $issues "error" $relativeSkillFile "Missing argument-hint frontmatter."
    } elseif ($config.lint.require_argument_hint -eq "warning" -and $frontmatter -notmatch '(?m)^argument-hint\s*:') {
      Add-Issue $warnings "warning" $relativeSkillFile "Missing argument-hint frontmatter."
    }

    foreach ($ref in [regex]::Matches($content, '@\./([^\s)]+)')) {
      $refPath = Join-Path $skillPath $ref.Groups[1].Value
      if (-not (Test-Path $refPath)) {
        Add-Issue $issues "error" $relativeSkillFile "Broken lazy reference '@./$($ref.Groups[1].Value)'."
      }
    }

    if ($lines -gt [int]$config.lint.hot_path_fail_lines) {
      $allow = $config.lint.allow_long_skills.PSObject.Properties.Name -contains $skillName
      if (-not $allow) {
        Add-Issue $issues "error" $relativeSkillFile "Skill has $lines lines, exceeding hard limit $($config.lint.hot_path_fail_lines)."
      } else {
        Add-Issue $warnings "warning" $relativeSkillFile "Skill has $lines lines, exceeding hard limit but allowlisted: $($config.lint.allow_long_skills.$skillName)"
      }
    } elseif ($lines -gt [int]$config.lint.hot_path_warning_lines) {
      Add-Issue $warnings "warning" $relativeSkillFile "Skill has $lines lines, exceeding warning limit $($config.lint.hot_path_warning_lines)."
    }
  }
}

$extensions = @($config.lint.scan_extensions)
Get-ChildItem $repoRoot -Recurse -File -Force |
  Where-Object {
    $full = $_.FullName
    ($full -notmatch '\\.git\\') -and
    ($extensions -contains $_.Extension)
  } |
  ForEach-Object {
    $relative = $_.FullName.Substring($repoRoot.Length + 1)
    $lineNo = 0
    Get-Content $_.FullName | ForEach-Object {
      $lineNo++
      if ($_ -match '^(<<<<<<<|>>>>>>>)(\s|$)' -or $_ -match '^=======$') {
        Add-Issue $issues "error" $relative "Unresolved conflict marker at line $lineNo."
      }
    }
  }

$result = [pscustomobject]@{
  ok = ($issues.Count -eq 0)
  errors = $issues
  warnings = $warnings
  summary = [pscustomobject]@{
    errors = $issues.Count
    warnings = $warnings.Count
  }
}

if ($Json) {
  $result | ConvertTo-Json -Depth 6
} else {
  Write-Host "Skill lint: $($issues.Count) errors, $($warnings.Count) warnings"
  foreach ($issue in $issues) {
    Write-Host "ERROR [$($issue.path)] $($issue.message)"
  }
  foreach ($warning in $warnings) {
    Write-Host "WARN  [$($warning.path)] $($warning.message)"
  }
}

if ($issues.Count -gt 0) {
  exit 1
}

exit 0
