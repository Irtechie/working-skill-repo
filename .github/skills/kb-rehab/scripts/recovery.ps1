[CmdletBinding()]
param(
  [ValidateSet('survey','prepare')][string]$Action = 'survey',
  [Parameter(Mandatory=$true)][string]$Root,
  [switch]$Json
)

$ErrorActionPreference = 'Stop'
$rootPath = [IO.Path]::GetFullPath($Root)
function Invoke-Git([string[]]$Arguments) {
  $output = & git -C $rootPath @Arguments 2>&1
  if ($LASTEXITCODE -ne 0) { return @{ ok=$false; output=($output -join "`n") } }
  return @{ ok=$true; output=($output -join "`n") }
}

$top = Invoke-Git @('rev-parse','--show-toplevel')
if (-not $top.ok) { throw "survey root is not a Git repository: $rootPath" }
if ($Action -eq 'prepare') {
  $result = [ordered]@{ schema_version=1; action='prepare'; status='dependency-needed'; reason='prepare requires a separately verified baseline SHA and explicit relative artifact allowlist'; preserved_source=$top.output.Trim(); next_action='continue independent work or supply verified recovery receipt' }
  if ($Json) { $result | ConvertTo-Json -Depth 5 } else { $result }
  exit 0
}
$common = Invoke-Git @('rev-parse','--git-common-dir')
$branch = Invoke-Git @('branch','--show-current')
$head = Invoke-Git @('rev-parse','HEAD')
$dirty = Invoke-Git @('status','--porcelain=v1','--ignored=no')
$remote = Invoke-Git @('remote')
$worktrees = Invoke-Git @('worktree','list','--porcelain')
$limitations = @('read-only survey: no stage, commit, checkout, reset, removal, or rewrite was attempted')
if (-not $remote.ok -or [string]::IsNullOrWhiteSpace($remote.output)) { $limitations += 'remote authority unavailable; no merge or retirement eligibility' }
$result = [ordered]@{
  schema_version = 1
  action = 'survey'
  repository = $top.output.Trim()
  common_dir = $common.output.Trim()
  branch = $branch.output.Trim()
  head = $head.output.Trim()
  dirty_paths = @($dirty.output -split "`n" | Where-Object { $_ })
  worktrees = @($worktrees.output -split "`n" | Where-Object { $_ })
  authority = if ($remote.ok -and -not [string]::IsNullOrWhiteSpace($remote.output)) { 'unverified' } else { 'unavailable' }
  pairing_status = 'candidates-unproven'
  limitations = $limitations
  next_actions = @('offer cleanup once only after a separately authorized decision','continue independent work')
}
if ($Json) { $result | ConvertTo-Json -Depth 6 } else { $result }
