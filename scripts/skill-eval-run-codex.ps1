param(
  [string]$Root = ".",
  [string]$ConfigPath = "config/skill-quality.json",
  [string]$FixtureId = "",
  [switch]$All,
  [switch]$DryRun,
  [switch]$KeepRun,
  [string]$RunRoot = ".atv/eval-runs",
  [string]$CodexCommand = "codex",
  [string]$Model = "",
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

function Get-SafeId {
  param([string]$Value)
  return ($Value -replace '[^A-Za-z0-9_.-]', '-')
}

function ConvertTo-JsonFile {
  param($Object, [string]$Path, [int]$Depth = 12)
  $Object | ConvertTo-Json -Depth $Depth | Set-Content -Path $Path -Encoding UTF8
}

function Get-FileSha256 {
  param([string]$Path)
  return (Get-FileHash $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

function Get-RelativeRepoPath {
  param([string]$RepoRoot, [string]$Path)
  $resolved = (Resolve-Path $Path).Path
  if ($resolved.StartsWith($RepoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    return $resolved.Substring($RepoRoot.Length + 1)
  }
  return $resolved
}

function New-ProtectedFileEntry {
  param(
    [string]$RepoRoot,
    [string]$Role,
    [string]$Path
  )
  return [pscustomobject]@{
    role = $Role
    path = Get-RelativeRepoPath $RepoRoot $Path
    sha256 = Get-FileSha256 $Path
  }
}

function New-RunManifest {
  param(
    [string]$RepoRoot,
    [string]$RunId,
    [string]$AdapterPath,
    [string]$ConfigPath,
    [string]$FixtureRoot,
    [string]$SchemaPath,
    $Fixture
  )

  $fixturePath = Join-Path $FixtureRoot "$($Fixture.id).json"
  return [pscustomobject]@{
    run_id = $RunId
    fixture_id = $Fixture.id
    created_at = (Get-Date).ToString("o")
    protected_files = @(
      New-ProtectedFileEntry $RepoRoot "fixture" $fixturePath
      New-ProtectedFileEntry $RepoRoot "scorer" (Join-Path $RepoRoot "scripts/skill-eval.ps1")
      New-ProtectedFileEntry $RepoRoot "result_schema" $SchemaPath
      New-ProtectedFileEntry $RepoRoot "adapter" $AdapterPath
      New-ProtectedFileEntry $RepoRoot "config" $ConfigPath
    )
  }
}

function Read-JsonObjectFromText {
  param([string]$Text)
  $trimmed = $Text.Trim()
  try {
    return ($trimmed | ConvertFrom-Json)
  } catch {
    $match = [regex]::Match($trimmed, '(?s)\{.*\}')
    if ($match.Success) {
      return ($match.Value | ConvertFrom-Json)
    }
    throw
  }
}

function Get-ActionableClaimChecks {
  param($Checks)
  $actionable = [System.Collections.Generic.List[object]]::new()
  foreach ($check in @($Checks)) {
    $type = "$($check.type)"
    if (($type -eq "file_exists") -and "$($check.path)") {
      $actionable.Add($check)
    } elseif ((($type -eq "command_ran") -or ($type -eq "file_read")) -and "$($check.contains)") {
      $actionable.Add($check)
    }
  }
  return @($actionable)
}

function Join-ProcessArguments {
  param([string[]]$Arguments)
  $quoted = foreach ($arg in $Arguments) {
    if ($null -eq $arg) {
      '""'
    } elseif ($arg -match '[\s"]') {
      '"' + ($arg -replace '"', '\"') + '"'
    } else {
      $arg
    }
  }
  return ($quoted -join " ")
}

function Get-Fixtures {
  param([string]$FixtureRoot, [string]$FixtureId, [bool]$All)
  $fixtures = @(Get-ChildItem $FixtureRoot -Filter "*.json" | Sort-Object Name | ForEach-Object {
      Get-Content $_.FullName -Raw | ConvertFrom-Json
    })
  if ($FixtureId) {
    $selected = @($fixtures | Where-Object { $_.id -eq $FixtureId })
    if ($selected.Count -eq 0) {
      throw "Unknown fixture id: $FixtureId"
    }
    return $selected
  }
  if ($All) {
    return $fixtures
  }
  throw "Pass -FixtureId <id> or -All."
}

function New-EvalPrompt {
  param($Fixture, [string]$RunId)
  $fixtureJson = $Fixture | ConvertTo-Json -Depth 12
  return @"
You are running a KB skill-routing evaluation.

Rules:
- Do not edit files.
- Do not run destructive commands.
- Decide the smallest correct KB route for the request.
- Return only JSON matching the provided output schema.
- Set eval_run_id exactly to "$RunId".
- Use the route fixture as ground truth input; do not execute the requested work.
- Fill trace.files_read and trace.commands only with files/commands you actually used.
- Include proof expectations that the selected route should eventually produce.

Route fixture:
$fixtureJson

Return a result object with:
- id: "codex-live-$($Fixture.id)"
- fixture_id: "$($Fixture.id)"
- expected_result: "pass"
- eval_run_id: "$RunId"
- actual.route
- actual.user_questions
- actual.artifacts
- actual.proof
- trace.files_read
- trace.commands
- trace.tools
- claim_checks for concrete claims only; use [] when there are no concrete file or command claims
- claim_checks must include both path and contains on every check, using an empty string when the field does not apply
"@
}

function New-DryRunResult {
  param($Fixture, [string]$RunId)
  return [pscustomobject]@{
    id = $RunId
    fixture_id = $Fixture.id
    expected_result = "pass"
    eval_run_id = $RunId
    actual = [pscustomobject]@{
      route = $Fixture.expected.route
      user_questions = [int]$Fixture.expected.max_user_questions
      artifacts = @($Fixture.expected.artifacts)
      proof = @($Fixture.expected.proof)
    }
    trace = [pscustomobject]@{
      files_read = @("evals/route-complexity/$($Fixture.id).json")
      commands = @("dry-run")
      tools = @("skill-eval-run-codex")
    }
    claim_checks = @(
      [pscustomobject]@{
        type = "file_exists"
        path = "evals/route-complexity/$($Fixture.id).json"
        contains = ""
        expected = $true
        claim = "Fixture file exists"
      },
      [pscustomobject]@{
        type = "command_ran"
        path = ""
        contains = "dry-run"
        expected = $true
        claim = "Dry-run command was recorded"
      }
    )
  }
}

function Invoke-CodexFixture {
  param(
    [string]$RepoRoot,
    [string]$RunDir,
    [string]$WorkspaceDir,
    [string]$CodexCommand,
    [string]$Model,
    [string]$SchemaPath,
    $Fixture,
    [string]$RunId
  )

  $promptPath = Join-Path $RunDir "prompt.txt"
  $stdoutPath = Join-Path $RunDir "codex.stdout.jsonl"
  $stderrPath = Join-Path $RunDir "codex.stderr.log"
  $finalPath = Join-Path $RunDir "codex-final.json"
  New-EvalPrompt $Fixture $RunId | Set-Content -Path $promptPath -Encoding UTF8

  $modelArgs = @()
  if ($Model) {
    $modelArgs += @("-m", $Model)
  }

  $args = @(
    "exec",
    "--cd", $WorkspaceDir,
    "--sandbox", "read-only",
    "--ephemeral",
    "--json",
    "--output-schema", $SchemaPath,
    "--output-last-message", $finalPath
  ) + $modelArgs + @("-")

  $prompt = Get-Content $promptPath -Raw
  $commandInfo = Get-Command $CodexCommand -ErrorAction Stop
  $baseArgs = @()
  $executable = $commandInfo.Source
  if ([System.IO.Path]::GetExtension($executable) -eq ".ps1") {
    $basedir = Split-Path $executable -Parent
    $codexJs = Join-Path $basedir "node_modules/@openai/codex/bin/codex.js"
    if (-not (Test-Path $codexJs)) {
      throw "Could not resolve Codex npm shim target: $codexJs"
    }
    $nodeExe = Join-Path $basedir "node.exe"
    if (-not (Test-Path $nodeExe)) {
      $nodeExe = "node.exe"
    }
    $baseArgs = @($codexJs)
    $executable = $nodeExe
  }

  $psi = [System.Diagnostics.ProcessStartInfo]::new()
  $psi.FileName = $executable
  $psi.Arguments = Join-ProcessArguments ($baseArgs + $args)
  $psi.RedirectStandardInput = $true
  $psi.RedirectStandardOutput = $true
  $psi.RedirectStandardError = $true
  $psi.UseShellExecute = $false
  $psi.WorkingDirectory = $RepoRoot

  $process = [System.Diagnostics.Process]::new()
  $process.StartInfo = $psi
  [void]$process.Start()
  $process.StandardInput.Write($prompt)
  $process.StandardInput.Close()
  $stdout = $process.StandardOutput.ReadToEnd()
  $stderr = $process.StandardError.ReadToEnd()
  $process.WaitForExit()

  $stdout | Set-Content -Path $stdoutPath -Encoding UTF8
  $stderr | Set-Content -Path $stderrPath -Encoding UTF8

  if ($process.ExitCode -ne 0) {
    throw "codex exec failed for $($Fixture.id) with exit code $($process.ExitCode). See $stderrPath"
  }
  if (-not (Test-Path $finalPath)) {
    throw "codex exec did not write final message: $finalPath"
  }

  $result = Read-JsonObjectFromText (Get-Content $finalPath -Raw)
  $result.id = $RunId
  $result.fixture_id = $Fixture.id

  $commandText = "$CodexCommand $($args -join ' ')"
  $commands = @($result.trace.commands)
  if (($commands | Where-Object { "$_" -like "*codex exec*" }).Count -eq 0) {
    $commands += $commandText
  }
  $result.trace.commands = $commands

  $tools = @($result.trace.tools)
  if ($tools -notcontains "codex") {
    $tools += "codex"
  }
  $result.trace.tools = $tools

  $checks = @(Get-ActionableClaimChecks $result.claim_checks)
  $checks += [pscustomobject]@{
    type = "file_exists"
    path = $finalPath
    contains = ""
    expected = $true
    claim = "Codex final message was captured"
  }
  $checks += [pscustomobject]@{
    type = "command_ran"
    path = ""
    contains = "codex exec"
    expected = $true
    claim = "Codex exec command was recorded"
  }
  $result.claim_checks = $checks

  return $result
}

$repoRoot = (Resolve-Path $Root).Path
$configFullPath = Resolve-RepoPath $repoRoot $ConfigPath
$config = Get-Content $configFullPath -Raw | ConvertFrom-Json
$fixtureRoot = Resolve-RepoPath $repoRoot $config.route_complexity.fixture_root
$schemaPath = Resolve-RepoPath $repoRoot "evals/skill-eval/result.schema.json"
$adapterPath = Join-Path $repoRoot "scripts/skill-eval-run-codex.ps1"
$fixtures = Get-Fixtures $fixtureRoot $FixtureId ([bool]$All)
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$runRootFull = Resolve-RepoPath $repoRoot $RunRoot
New-Item -ItemType Directory -Force -Path $runRootFull | Out-Null

$rows = [System.Collections.Generic.List[object]]::new()

foreach ($fixture in $fixtures) {
  $runId = "$timestamp-$($fixture.id)-codex"
  if ($DryRun) {
    $runId = "$timestamp-$($fixture.id)-codex-dry-run"
  }
  $runDir = Join-Path $runRootFull (Get-SafeId $runId)
  $workspaceDir = Join-Path $runDir "workspace"
  New-Item -ItemType Directory -Force -Path $runDir | Out-Null

  ConvertTo-JsonFile $fixture (Join-Path $runDir "fixture.json")
  New-EvalPrompt $fixture $runId | Set-Content -Path (Join-Path $runDir "prompt.txt") -Encoding UTF8
  $manifestPath = Join-Path $runDir "manifest.json"
  ConvertTo-JsonFile (New-RunManifest $repoRoot $runId $adapterPath $configFullPath $fixtureRoot $schemaPath $fixture) $manifestPath

  try {
    if ($DryRun) {
      $result = New-DryRunResult $fixture $runId
    } else {
      git -C $repoRoot worktree add --detach $workspaceDir HEAD | Out-Null
      $result = Invoke-CodexFixture $repoRoot $runDir $workspaceDir $CodexCommand $Model $schemaPath $fixture $runId
    }

    $resultPath = Join-Path $runDir "result.json"
    ConvertTo-JsonFile $result $resultPath
    powershell -ExecutionPolicy Bypass -File (Join-Path $repoRoot "scripts/skill-eval.ps1") -Root $repoRoot -ResultPath $resultPath -RequiredRunId $runId -ManifestPath $manifestPath | Tee-Object -FilePath (Join-Path $runDir "score.log") | Out-Null
    if ($LASTEXITCODE -ne 0) {
      throw "skill-eval scoring failed for $($fixture.id)."
    }

    $rows.Add([pscustomobject]@{
      fixture_id = $fixture.id
      run_id = $runId
      result_path = $resultPath
      manifest_path = $manifestPath
      mode = if ($DryRun) { "dry-run" } else { "live" }
      ok = $true
    })
  } finally {
    if ((-not $DryRun) -and (Test-Path $workspaceDir)) {
      if (-not $KeepRun) {
        git -C $repoRoot worktree remove $workspaceDir --force | Out-Null
      }
    } elseif ($DryRun -and (-not $KeepRun)) {
      Remove-Item -LiteralPath $runDir -Recurse -Force
    }
  }
}

if ($Json) {
  [pscustomobject]@{ ok = $true; runs = $rows } | ConvertTo-Json -Depth 8
} else {
  Write-Host "Codex skill eval adapter: $($rows.Count) run(s)"
  foreach ($row in $rows) {
    Write-Host "$($row.fixture_id): mode=$($row.mode) result=$($row.result_path)"
  }
}

exit 0
