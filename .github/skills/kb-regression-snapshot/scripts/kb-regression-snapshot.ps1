param(
  [Parameter(Position=0)]
  [ValidateSet("capture", "verify")]
  [string]$Mode = "verify",
  [string]$SliceId,
  [string]$SpecPath,
  [string]$SnapshotDir = ".kb/snapshots",
  [string]$BaseUrl = $env:DEV_SERVER_URL,
  [string[]]$SnapshotId,
  [string]$MilestoneId,
  [int]$CliTimeoutSeconds = 300
)

$ErrorActionPreference = "Stop"

function Resolve-Url([string]$Url) {
  if ($Url -match '^https?://') { return $Url }
  if (-not $BaseUrl) { $script:BaseUrl = "http://localhost:3000" }
  return ($script:BaseUrl.TrimEnd("/") + "/" + $Url.TrimStart("/"))
}

function Assert-RouteStatus($Check) {
  $url = Resolve-Url $Check.url
  try {
    $response = Invoke-WebRequest -Uri $url -UseBasicParsing -Method GET
    $status = [int]$response.StatusCode
  } catch {
    $status = [int]$_.Exception.Response.StatusCode
  }
  if ($status -ne [int]$Check.expected_status) {
    throw "route-status $url expected $($Check.expected_status) observed $status"
  }
}

function Assert-ApiSchema($Check) {
  Assert-RouteStatus $Check
  if ($Check.required_fields) {
    $json = (Invoke-WebRequest -Uri (Resolve-Url $Check.url) -UseBasicParsing).Content | ConvertFrom-Json
    foreach ($field in $Check.required_fields) {
      if (-not ($json.PSObject.Properties.Name -contains $field)) {
        throw "api-schema $($Check.url) missing field $field"
      }
    }
  }
}

function Assert-FileChecksum($Check) {
  if (-not (Test-Path -LiteralPath $Check.path)) { throw "file-checksum missing $($Check.path)" }
  $actual = (Get-FileHash -LiteralPath $Check.path -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($actual -ne [string]$Check.sha256.ToLowerInvariant()) {
    throw "file-checksum $($Check.path) expected $($Check.sha256) observed $actual"
  }
}

function Assert-Cli($Check) {
  $timeoutSeconds = if ($Check.timeout_seconds) { [int]$Check.timeout_seconds } else { $CliTimeoutSeconds }
  $start = [System.Diagnostics.ProcessStartInfo]::new()
  $start.FileName = Join-Path $PSHOME "pwsh.exe"
  $start.UseShellExecute = $false
  $start.CreateNoWindow = $true
  $start.RedirectStandardOutput = $true
  $start.RedirectStandardError = $true
  foreach ($argument in @("-NoProfile", "-NonInteractive", "-Command", [string]$Check.command)) {
    [void]$start.ArgumentList.Add($argument)
  }
  $process = [System.Diagnostics.Process]::new()
  $process.StartInfo = $start
  [void]$process.Start()
  $stdoutTask = $process.StandardOutput.ReadToEndAsync()
  $stderrTask = $process.StandardError.ReadToEndAsync()
  if (-not $process.WaitForExit($timeoutSeconds * 1000)) {
    $process.Kill($true)
    $process.WaitForExit()
    throw "cli $($Check.command) timed out after $timeoutSeconds seconds (exit 124)"
  }
  $output = $stdoutTask.Result + $stderrTask.Result
  $exit = $process.ExitCode
  if ($exit -ne [int]$Check.expected_exit_code) {
    throw "cli $($Check.command) expected exit $($Check.expected_exit_code) observed $exit"
  }
  if ($Check.expected_output_substring -and -not $output.Contains([string]$Check.expected_output_substring)) {
    throw "cli $($Check.command) missing output substring $($Check.expected_output_substring)"
  }
}

function Assert-DomElements($Checks) {
  if ($Checks.Count -eq 0) { return }
  $payload = @($Checks | ForEach-Object {
    [pscustomobject]@{
      url = Resolve-Url $_.url
      selector = [string]$_.selector
      expected_text = [string]$_.expected_text
      expected_text_pattern = [string]$_.expected_text_pattern
    }
  })
  $node = @"
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch({ headless: true });
  try {
    const checks = require(process.argv[2]);
    const page = await browser.newPage();
    const errors = [];
    page.on('console', msg => { if (msg.type() === 'error') errors.push(msg.text()); });
    for (const check of checks) {
      await page.goto(check.url, { waitUntil: 'domcontentloaded' });
      const loc = page.locator(check.selector).first();
      await loc.waitFor({ state: 'visible', timeout: 10000 });
      const text = (await loc.textContent()) || '';
      if (check.expected_text && text.trim() !== check.expected_text) throw new Error(`expected text ${check.expected_text} observed ${text.trim()}`);
      if (check.expected_text_pattern && !(new RegExp(check.expected_text_pattern).test(text))) throw new Error(`text ${text} did not match ${check.expected_text_pattern}`);
    }
    if (errors.length !== 0) throw new Error(`console errors: ${errors.join(' | ')}`);
  } finally {
    await browser.close();
  }
})().catch(err => { console.error(err.message); process.exit(1); });
"@
  $tmp = Join-Path ([System.IO.Path]::GetTempPath()) "kb-snapshot-dom-$([guid]::NewGuid()).js"
  $payloadPath = Join-Path ([System.IO.Path]::GetTempPath()) "kb-snapshot-dom-$([guid]::NewGuid()).json"
  Set-Content -LiteralPath $tmp -Value $node -Encoding UTF8
  $payload | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $payloadPath -Encoding UTF8
  try {
    node $tmp $payloadPath
    if ($LASTEXITCODE -ne 0) { throw "dom-element batch failed" }
  } finally {
    Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $payloadPath -Force -ErrorAction SilentlyContinue
  }
}

function Invoke-Check($Check, [System.Collections.Generic.List[object]]$DomChecks) {
  switch ($Check.type) {
    "route-status" { Assert-RouteStatus $Check }
    "api-schema" { Assert-ApiSchema $Check }
    "file-checksum" { Assert-FileChecksum $Check }
    "cli" { Assert-Cli $Check }
    "dom-element" { $DomChecks.Add($Check) }
    default { throw "unknown snapshot check type: $($Check.type)" }
  }
}

function Get-SnapshotFingerprint($Files) {
  $rows = @($Files | ForEach-Object {
    "$($_.Name):$((Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256).Hash.ToLowerInvariant())"
  })
  $bytes = [Text.Encoding]::UTF8.GetBytes(($rows -join "`n"))
  $hash = [Security.Cryptography.SHA256]::HashData($bytes)
  return [Convert]::ToHexString($hash).ToLowerInvariant()
}

New-Item -ItemType Directory -Force -Path $SnapshotDir | Out-Null

if ($Mode -eq "capture") {
  if (-not $SliceId -or -not $SpecPath) { throw "capture requires -SliceId and -SpecPath" }
  $snapshot = Get-Content -LiteralPath $SpecPath -Raw | ConvertFrom-Json
  $snapshot | Add-Member -NotePropertyName slice_id -NotePropertyValue $SliceId -Force
  $snapshot | Add-Member -NotePropertyName captured_at -NotePropertyValue ([DateTimeOffset]::UtcNow.ToString("o")) -Force
  $domChecks = [System.Collections.Generic.List[object]]::new()
  foreach ($check in $snapshot.checks) { Invoke-Check $check $domChecks }
  Assert-DomElements $domChecks
  $out = Join-Path $SnapshotDir "$SliceId.json"
  $snapshot | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $out -Encoding UTF8
  "snapshot-capture: PASS $SliceId -> $out"
  exit 0
}

if (($SnapshotId.Count -eq 0) -and -not $MilestoneId) {
  throw "verify requires -SnapshotId <id,...> or -MilestoneId <id>"
}
if (($SnapshotId.Count -gt 0) -and $MilestoneId) {
  throw "verify accepts either -SnapshotId or -MilestoneId, not both"
}

$available = @(Get-ChildItem -Path $SnapshotDir -Filter "*.json" -File -ErrorAction SilentlyContinue |
  Where-Object { $_.Name -notlike "*-spec.json" } |
  Sort-Object Name)
if ($SnapshotId.Count -gt 0) {
  $wanted = @($SnapshotId | ForEach-Object { "$_.json" })
  $files = @($available | Where-Object { $wanted -contains $_.Name })
  $missing = @($wanted | Where-Object { $name = $_; -not ($files | Where-Object Name -eq $name) })
  if ($missing.Count -gt 0) { throw "snapshot not found: $($missing -join ', ')" }
} else {
  if ($MilestoneId -notmatch '^[A-Za-z0-9._-]+$') { throw "MilestoneId contains invalid characters" }
  $files = $available
  $fingerprint = Get-SnapshotFingerprint $files
  $milestoneDir = Join-Path $SnapshotDir ".milestones"
  $milestonePath = Join-Path $milestoneDir "$MilestoneId.sha256"
  if ((Test-Path -LiteralPath $milestonePath) -and ((Get-Content -LiteralPath $milestonePath -Raw).Trim() -eq $fingerprint)) {
    "snapshot-verify: REUSE milestone=$MilestoneId fingerprint=$fingerprint"
    exit 0
  }
}

$domChecks = [System.Collections.Generic.List[object]]::new()
$count = 0
foreach ($file in $files) {
  $snapshot = Get-Content -LiteralPath $file.FullName -Raw | ConvertFrom-Json
  foreach ($check in $snapshot.checks) { Invoke-Check $check $domChecks }
  $count++
}
Assert-DomElements $domChecks
if ($MilestoneId) {
  New-Item -ItemType Directory -Force -Path $milestoneDir | Out-Null
  Set-Content -LiteralPath $milestonePath -Value $fingerprint -Encoding ascii
}
"snapshot-verify: PASS $count/$count snapshots"
exit 0
