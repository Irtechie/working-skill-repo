param(
  [string]$Root = ".",
  [string]$Start = "",
  [switch]$Status,
  [string]$RunId = ""
)

$ErrorActionPreference = "Stop"

function Resolve-RepoPath {
  param([string]$Base, [string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return $Path
  }
  return (Join-Path $Base $Path)
}

function Get-FileSha256 {
  param([string]$Path)
  return (Get-FileHash $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

function ConvertTo-Slug {
  param([string]$Value)
  $slug = ($Value.ToLowerInvariant() -replace "[^a-z0-9]+", "-").Trim("-")
  if (-not $slug) {
    return "pipeline"
  }
  return $slug
}

function Get-LatestRunDir {
  param([string]$RunRoot)
  if (-not (Test-Path $RunRoot)) {
    return $null
  }
  return Get-ChildItem $RunRoot -Directory | Sort-Object Name -Descending | Select-Object -First 1
}

function Write-PhasePrompt {
  param(
    [string]$Path,
    $Pipeline,
    $Phase
  )

  $skills = @($Phase.skills) -join ", "
  $outputs = @($Phase.required_outputs) -join ", "
  $fresh = if ($Phase.fresh_context) { "yes" } else { "no" }
  $lines = @(
    "# Phase: $($Phase.id)",
    "",
    "Pipeline: $($Pipeline.id)",
    "Fresh context: $fresh",
    "Skills: $skills",
    "Required outputs: $outputs",
    "",
    "## Instructions",
    "",
    "- Read pipeline.json in this run directory.",
    "- Read only the context needed for this phase.",
    "- Produce the required outputs listed above.",
    "- Do not execute later phases from this prompt."
  )
  $lines | Set-Content -Path $Path -Encoding UTF8
}

function Start-Pipeline {
  param(
    [string]$RepoRoot,
    [string]$PipelineId
  )

  $pipelinePath = Resolve-RepoPath $RepoRoot ("config/pipelines/{0}.json" -f $PipelineId)
  if (-not (Test-Path $pipelinePath)) {
    throw "Pipeline '$PipelineId' not found at config/pipelines/$PipelineId.json"
  }

  $pipeline = Get-Content $pipelinePath -Raw | ConvertFrom-Json
  $runRoot = Resolve-RepoPath $RepoRoot ".atv/pipeline-runs"
  New-Item -ItemType Directory -Force -Path $runRoot | Out-Null

  $runId = "{0}-{1}-{2}" -f (Get-Date -Format "yyyyMMdd-HHmmss-fff"), ([guid]::NewGuid().ToString("N").Substring(0, 8)), (ConvertTo-Slug "$($pipeline.id)")
  $runDir = Join-Path $runRoot $runId
  New-Item -ItemType Directory -Force -Path $runDir | Out-Null
  New-Item -ItemType Directory -Force -Path (Join-Path $runDir "phase-prompts") | Out-Null

  $protected = [System.Collections.Generic.List[object]]::new()
  foreach ($entry in @($pipeline.protected_files)) {
    $pathValue = "$($entry.path)"
    $fullPath = Resolve-RepoPath $RepoRoot $pathValue
    $protected.Add([pscustomobject]@{
      role = "$($entry.role)"
      path = $pathValue
      sha256 = if (Test-Path $fullPath) { Get-FileSha256 $fullPath } else { "" }
      exists = Test-Path $fullPath
    })
  }

  $run = [pscustomobject]@{
    run_id = $runId
    pipeline_id = "$($pipeline.id)"
    started_at = (Get-Date).ToString("o")
    status = "started"
    run_dir = $runDir
    phase_count = @($pipeline.phases).Count
    current_phase = @($pipeline.phases)[0].id
    proof_commands = @($pipeline.proof_commands)
    protected_files = $protected
  }

  $pipeline | ConvertTo-Json -Depth 10 | Set-Content -Path (Join-Path $runDir "pipeline.json") -Encoding UTF8
  $run | ConvertTo-Json -Depth 10 | Set-Content -Path (Join-Path $runDir "run.json") -Encoding UTF8
  [pscustomobject]@{
    run_id = $runId
    protected_files = $protected
  } | ConvertTo-Json -Depth 8 | Set-Content -Path (Join-Path $runDir "protected-files.json") -Encoding UTF8
  [pscustomobject]@{
    run_id = $runId
    proof_commands = @($pipeline.proof_commands)
    results = @()
  } | ConvertTo-Json -Depth 8 | Set-Content -Path (Join-Path $runDir "proof.json") -Encoding UTF8

  foreach ($phase in @($pipeline.phases)) {
    Write-PhasePrompt (Join-Path $runDir ("phase-prompts/{0}.md" -f $phase.id)) $pipeline $phase
  }

  $phaseRows = @($pipeline.phases | ForEach-Object {
    $phaseSkills = @($_.skills) -join ", "
    $phaseOutputs = @($_.required_outputs) -join ", "
    "| $($_.id) | $($_.fresh_context) | $phaseSkills | $phaseOutputs |"
  })
  $selected = @(
    "# Selected Pipeline",
    "",
    "- Run ID: $runId",
    "- Pipeline: $($pipeline.id)",
    "- Status: started",
    "- Run directory: $runDir",
    "",
    "| Phase | Fresh Context | Skills | Required Outputs |",
    "|---|---|---|---|"
  ) + $phaseRows
  $selected | Set-Content -Path (Join-Path $runDir "selected-pipeline.md") -Encoding UTF8

  Write-Host "KB pipeline started: $runId"
  Write-Host "Run directory: $runDir"
}

function Show-PipelineStatus {
  param(
    [string]$RepoRoot,
    [string]$RunId
  )

  $runRoot = Resolve-RepoPath $RepoRoot ".atv/pipeline-runs"
  $runDir = $null
  if ($RunId) {
    $candidate = Join-Path $runRoot $RunId
    if (Test-Path $candidate) {
      $runDir = Get-Item $candidate
    }
  } else {
    $runDir = Get-LatestRunDir $runRoot
  }

  if (-not $runDir) {
    throw "No pipeline run found."
  }

  $runPath = Join-Path $runDir.FullName "run.json"
  if (-not (Test-Path $runPath)) {
    throw "Pipeline run is missing run.json: $($runDir.FullName)"
  }

  $run = Get-Content $runPath -Raw | ConvertFrom-Json
  Write-Host "KB pipeline status: $($run.run_id)"
  Write-Host "Pipeline: $($run.pipeline_id)"
  Write-Host "Status: $($run.status)"
  Write-Host "Current phase: $($run.current_phase)"
  Write-Host "Run directory: $($run.run_dir)"
}

$repoRoot = (Resolve-Path $Root).Path

if ($Start) {
  Start-Pipeline $repoRoot $Start
  exit 0
}

if ($Status) {
  Show-PipelineStatus $repoRoot $RunId
  exit 0
}

Write-Host "Usage:"
Write-Host "  scripts/kb-pipeline.ps1 -Start skill-bundle-proof-spike"
Write-Host "  scripts/kb-pipeline.ps1 -Status [-RunId <id>]"
exit 1
