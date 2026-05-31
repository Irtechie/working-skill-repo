param(
  [string]$Root = ".",
  [string]$ConfigPath = "config/skill-quality.json",
  [string]$FixtureId = "",
  [switch]$All,
  [switch]$DryRun,
  [switch]$KeepRun,
  [string]$RunRoot = ".atv/eval-runs",
  [string]$CopilotCommand = "copilot",
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

function Assert-ResultShape {
  param($Result)
  foreach ($field in @("id", "fixture_id", "expected_result", "eval_run_id", "actual", "trace", "claim_checks")) {
    if (-not (Has-Property $Result $field)) {
      throw "GHCP result JSON is missing required field '$field'."
    }
  }
  foreach ($field in @("route", "user_questions", "artifacts", "proof")) {
    if (-not (Has-Property $Result.actual $field)) {
      throw "GHCP result JSON is missing required field 'actual.$field'."
    }
  }
  foreach ($field in @("files_read", "commands", "tools")) {
    if (-not (Has-Property $Result.trace $field)) {
      throw "GHCP result JSON is missing required field 'trace.$field'."
    }
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

function Test-GhcpUnavailable {
  param([string]$Text)
  return ($Text -match '(?i)(not logged in|login|authenticate|authentication|unauthorized|forbidden|copilot.*policy|policy.*copilot|device code|subscription|license|not enabled)')
}

function Resolve-NpmShimCommand {
  param([string]$Command, [string]$PackageEntry)
  $commandInfo = Get-Command $Command -ErrorAction Stop
  $executable = $commandInfo.Source
  $baseArgs = @()
  if ([System.IO.Path]::GetExtension($executable) -eq ".ps1") {
    $basedir = Split-Path $executable -Parent
    $entry = Join-Path $basedir $PackageEntry
    if (-not (Test-Path $entry)) {
      throw "Could not resolve npm shim target: $entry"
    }
    $nodeExe = Join-Path $basedir "node.exe"
    if (-not (Test-Path $nodeExe)) {
      $nodeExe = "node.exe"
    }
    $baseArgs = @($entry)
    $executable = $nodeExe
  }
  return [pscustomobject]@{
    executable = $executable
    base_args = $baseArgs
  }
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
You are running a KB skill-routing evaluation for GitHub Copilot/GHCP.

Rules:
- Do not edit files.
- Do not run destructive commands.
- Do not execute the requested work.
- Decide the smallest correct KB route for the request.
- Return exactly one JSON object and no markdown, prose, or code fences.
- Set eval_run_id exactly to "$RunId".
- Use the route fixture as ground truth input.
- Fill trace.files_read and trace.commands only with files/commands you actually used.
- Include proof expectations that the selected route should eventually produce.
- If you did not read files or run commands, use empty arrays for those trace fields.
- claim_checks is for concrete claims only; use [] when there are no concrete file or command claims.
- Every claim_check object must include type, path, contains, expected, and claim. Use an empty string when path or contains does not apply.

Route fixture:
$fixtureJson

Return this result object shape:
{
  "id": "ghcp-live-$($Fixture.id)",
  "fixture_id": "$($Fixture.id)",
  "expected_result": "pass",
  "eval_run_id": "$RunId",
  "actual": {
    "route": "<selected KB route>",
    "user_questions": 0,
    "artifacts": ["<expected artifact evidence>"],
    "proof": ["<expected proof evidence>"]
  },
  "trace": {
    "files_read": [],
    "commands": [],
    "tools": []
  },
  "claim_checks": []
}
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
      tools = @("skill-eval-run-ghcp")
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

function Invoke-GhcpFixture {
  param(
    [string]$RepoRoot,
    [string]$RunDir,
    [string]$WorkspaceDir,
    [string]$CopilotCommand,
    [string]$Model,
    $Fixture,
    [string]$RunId
  )

  $promptPath = Join-Path $RunDir "prompt.txt"
  $stdoutPath = Join-Path $RunDir "ghcp.stdout.log"
  $stderrPath = Join-Path $RunDir "ghcp.stderr.log"
  $finalPath = Join-Path $RunDir "ghcp-final.json"
  $sharePath = Join-Path $RunDir "ghcp-session.md"
  New-EvalPrompt $Fixture $RunId | Set-Content -Path $promptPath -Encoding UTF8

  $modelArgs = @()
  if ($Model) {
    $modelArgs += @("--model", $Model)
  }

  $args = @(
    "-C", $WorkspaceDir,
    "-s",
    "--no-ask-user",
    "--no-color",
    "--stream", "off",
    "--allow-tool=read",
    "--deny-tool=write,shell,url,memory",
    "--share=$sharePath"
  ) + $modelArgs

  $prompt = Get-Content $promptPath -Raw
  $command = Resolve-NpmShimCommand $CopilotCommand "node_modules/@github/copilot/npm-loader.js"

  $psi = [System.Diagnostics.ProcessStartInfo]::new()
  $psi.FileName = $command.executable
  $psi.Arguments = Join-ProcessArguments ($command.base_args + $args)
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
    $combined = "$stdout`n$stderr"
    if (Test-GhcpUnavailable $combined) {
      ConvertTo-JsonFile ([pscustomobject]@{
          status = "unavailable"
          runtime = "ghcp"
          fixture_id = $Fixture.id
          exit_code = $process.ExitCode
          stdout = $stdoutPath
          stderr = $stderrPath
          transcript = $sharePath
          reason = "GitHub Copilot CLI authentication, policy, subscription, or runtime access is unavailable."
        }) (Join-Path $RunDir "unavailable.json")
      throw "ghcp unavailable for $($Fixture.id) with exit code $($process.ExitCode). See $(Join-Path $RunDir "unavailable.json")"
    }
    throw "copilot failed for $($Fixture.id) with exit code $($process.ExitCode). See $stderrPath"
  }

  $result = Read-JsonObjectFromText $stdout
  Assert-ResultShape $result

  $result.id = $RunId
  $result.fixture_id = $Fixture.id
  $result.expected_result = "pass"

  $commandText = "$CopilotCommand $($args -join ' ')"
  $commands = @($result.trace.commands)
  if (($commands | Where-Object { "$_" -like "*copilot*" }).Count -eq 0) {
    $commands += $commandText
  }
  $result.trace.commands = $commands

  $tools = @($result.trace.tools)
  if ($tools -notcontains "ghcp") {
    $tools += "ghcp"
  }
  if ($tools -notcontains "copilot") {
    $tools += "copilot"
  }
  $result.trace.tools = $tools

  $checks = @(Get-ActionableClaimChecks $result.claim_checks)
  $checks += [pscustomobject]@{
    type = "file_exists"
    path = $stdoutPath
    contains = ""
    expected = $true
    claim = "GHCP stdout was captured"
  }
  $checks += [pscustomobject]@{
    type = "command_ran"
    path = ""
    contains = "copilot"
    expected = $true
    claim = "Copilot command was recorded"
  }
  $result.claim_checks = $checks

  ConvertTo-JsonFile $result $finalPath
  return $result
}

$repoRoot = (Resolve-Path $Root).Path
$configFullPath = Resolve-RepoPath $repoRoot $ConfigPath
$config = Get-Content $configFullPath -Raw | ConvertFrom-Json
$fixtureRoot = Resolve-RepoPath $repoRoot $config.route_complexity.fixture_root
$schemaPath = Resolve-RepoPath $repoRoot "evals/skill-eval/result.schema.json"
$adapterPath = Join-Path $repoRoot "scripts/skill-eval-run-ghcp.ps1"
$fixtures = Get-Fixtures $fixtureRoot $FixtureId ([bool]$All)
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$runRootFull = Resolve-RepoPath $repoRoot $RunRoot
New-Item -ItemType Directory -Force -Path $runRootFull | Out-Null

$rows = [System.Collections.Generic.List[object]]::new()

foreach ($fixture in $fixtures) {
  $runId = "$timestamp-$($fixture.id)-ghcp"
  if ($DryRun) {
    $runId = "$timestamp-$($fixture.id)-ghcp-dry-run"
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
      $result = Invoke-GhcpFixture $repoRoot $runDir $workspaceDir $CopilotCommand $Model $fixture $runId
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
  Write-Host "GHCP skill eval adapter: $($rows.Count) run(s)"
  foreach ($row in $rows) {
    Write-Host "$($row.fixture_id): mode=$($row.mode) result=$($row.result_path)"
  }
}

exit 0
