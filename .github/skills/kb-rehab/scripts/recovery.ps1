[CmdletBinding()]
param([ValidateSet('survey','prepare','verify','continue')][string]$Action='survey', [Parameter(Mandatory=$true)][string]$Root, [string]$Request, [switch]$Json)
$ErrorActionPreference='Stop'
$rootPath=[IO.Path]::GetFullPath($Root)
$gitPath=(Get-Command git -CommandType Application -ErrorAction Stop).Source
$gitSafeConfig=@()
function Quote-Argument([string]$Value) {
  return '"' + [regex]::Replace([regex]::Replace($Value, '(\\*)"', '$1$1\"'), '(\\+)$', '$1$1') + '"'
}
function Invoke-Git([string[]]$Arguments) {
  # Bounded native argv; no shell interpolation, credential prompt or stderr leakage.
  $s=New-Object Diagnostics.ProcessStartInfo
  $s.FileName=$gitPath
  $s.Arguments=((@('-C',$rootPath,'-c','core.quotePath=false','-c','core.fsmonitor=false','-c','core.hooksPath=NUL')+$gitSafeConfig+$Arguments | ForEach-Object { Quote-Argument $_ }) -join ' ')
  $s.UseShellExecute=$false; $s.CreateNoWindow=$true
  $s.RedirectStandardOutput=$true; $s.RedirectStandardError=$true
  $s.StandardOutputEncoding=New-Object Text.UTF8Encoding($false)
  foreach ($name in @('GIT_DIR','GIT_WORK_TREE','GIT_INDEX_FILE','GIT_COMMON_DIR','GIT_OBJECT_DIRECTORY','GIT_ALTERNATE_OBJECT_DIRECTORIES','GIT_CONFIG_COUNT','GIT_CONFIG_PARAMETERS')) { $s.EnvironmentVariables.Remove($name) }
  foreach ($name in @($s.EnvironmentVariables.Keys)) { if ($name -match '^GIT_CONFIG_(KEY|VALUE)_') { $s.EnvironmentVariables.Remove($name) } }
  $s.EnvironmentVariables['GIT_OPTIONAL_LOCKS']='0'
  $s.EnvironmentVariables['GIT_TERMINAL_PROMPT']='0'
  $s.EnvironmentVariables['GCM_INTERACTIVE']='Never'
  $p=New-Object Diagnostics.Process; $p.StartInfo=$s
  try {
    [void]$p.Start(); $o=$p.StandardOutput.ReadToEndAsync(); $e=$p.StandardError.ReadToEndAsync()
    if (-not $p.WaitForExit(20000)) { $p.Kill(); return @{ok=$false; output=''; reason='git-timeout'} }
    $p.WaitForExit()
    return @{ok=($p.ExitCode -eq 0); output=$o.Result}
  } finally { $p.Dispose() }
}
function Get-Hash([string]$Text) {
  $h=[Security.Cryptography.SHA256]::Create()
  try { return ([BitConverter]::ToString($h.ComputeHash([Text.Encoding]::UTF8.GetBytes($Text)))).Replace('-','').ToLowerInvariant() } finally { $h.Dispose() }
}
function Get-FileDigest([string]$Path) {
  $h=[Security.Cryptography.SHA256]::Create(); $stream=$null
  try { $stream=[IO.File]::OpenRead($Path); return ([BitConverter]::ToString($h.ComputeHash($stream))).Replace('-','').ToLowerInvariant() }
  finally { if ($stream) { $stream.Dispose() }; $h.Dispose() }
}
function Get-Divergence([string]$Base,[string]$Tip) {
  $v=Invoke-Git @('rev-list','--left-right','--count',"$Base...$Tip",'--')
  if ($v.ok -and $v.output.Trim() -match '^(\d+)\s+(\d+)$') { return @{status='verified';behind=[int]$Matches[1];ahead=[int]$Matches[2];base=$Base} }
  return @{status='unavailable';behind=$null;ahead=$null;base=$Base}
}
$filterKeys=Invoke-Git @('config','--name-only','--get-regexp','^filter\..*\.(process|smudge|clean|required)$')
foreach ($key in ($filterKeys.output -split '\r?\n')) {
  if ($key) { $value=''; if ($key.EndsWith('.required')) { $value='false' }; $gitSafeConfig+=@('-c',($key+'='+$value)) }
}
$top=Invoke-Git @('rev-parse','--show-toplevel')
if (-not $top.ok) { throw 'survey root is not a Git repository' }
$rootPath=[IO.Path]::GetFullPath($top.output.Trim())
$common=(Invoke-Git @('rev-parse','--git-common-dir')).output.Trim()
if (-not [IO.Path]::IsPathRooted($common)) { $common=Join-Path $rootPath $common }
$common=[IO.Path]::GetFullPath($common)
$branch=(Invoke-Git @('branch','--show-current')).output.Trim()
$head=(Invoke-Git @('rev-parse','--verify','HEAD')).output.Trim()
$limitations=New-Object 'Collections.Generic.List[string]'
$limitations.Add('Survey grants no merge or deletion authority. Candidate links are not proof.')
$policy=@{status='absent-conservative-default';sha256=$null}; $policyValue=$null
$policyPath=Join-Path $rootPath 'config/rehab-policy.json'
if (Test-Path -LiteralPath $policyPath -PathType Leaf) {
  try { $policyValue=[IO.File]::ReadAllText($policyPath) | ConvertFrom-Json; $policy=@{status='present';sha256=(Get-FileDigest $policyPath)} }
  catch { $policy.status='invalid'; $limitations.Add('Invalid policy: policy-dependent writes require repair.') }
}
$remotes=@((Invoke-Git @('remote')).output -split '\r?\n' | Where-Object { $_ })
$upstreamRemote=''; if ($branch) { $upstreamRemote=(Invoke-Git @('config','--get',"branch.$branch.remote")).output.Trim() }
$remote=''; $selection=''
if ($upstreamRemote -and $upstreamRemote -ne '.' -and $remotes -contains $upstreamRemote) { $remote=$upstreamRemote; $selection='branch-upstream' }
elseif ($policyValue -and $policyValue.remote -and $remotes -contains [string]$policyValue.remote) { $remote=[string]$policyValue.remote; $selection='project-policy' }
elseif ($remotes.Count -eq 1) { $remote=$remotes[0]; $selection='only-remote' }
$authority=@{status='unavailable';remote=$remote;selection=$selection;default_ref=$null;advertised_sha=$null;baseline_sha=$null;verified_at=$null}
if (-not $remote) { $limitations.Add('Absent or ambiguous remotes: independent local work can continue.') }
else {
  $ad=Invoke-Git @('ls-remote','--symref','--',$remote,'HEAD'); $defaultRef=''; $sha=''
  foreach ($line in ($ad.output -split '\r?\n')) {
    if ($line -match '^ref: (refs/heads/[^\s]+)\s+HEAD$') { $defaultRef=$Matches[1] }
    if ($line -match '^([a-f0-9]{40,64})\s+HEAD$') { $sha=$Matches[1] }
  }
  if ($ad.ok -and $defaultRef -and $sha) {
    $authority.default_ref=$defaultRef; $authority.advertised_sha=$sha
    $temporaryRef='refs/kb-recovery-survey/'+[Guid]::NewGuid().ToString('N')
    try {
      $fetch=Invoke-Git @('fetch','--quiet','--no-tags','--no-prune','--no-auto-maintenance','--refmap=','--no-recurse-submodules','--no-write-fetch-head','--',$remote,($defaultRef+':'+$temporaryRef))
      $fetched=Invoke-Git @('rev-parse','--verify',($temporaryRef+'^{commit}'))
      if ($fetch.ok -and $fetched.ok -and $fetched.output.Trim() -eq $sha) { $authority.status='verified'; $authority.baseline_sha=$sha; $authority.verified_at=[DateTimeOffset]::UtcNow.ToString('o') }
      else { $limitations.Add('Fetch or advertised/fetched equality failed; cached refs cannot authorize integration.') }
    } finally { $removed=Invoke-Git @('update-ref','-d',$temporaryRef); if (-not $removed.ok) { $limitations.Add('Survey temporary ref cleanup failed; preserve it for inspection.') } }
  } else { $limitations.Add('Remote advertisement unavailable: independent local work can continue.') }
}
$defaultDivergence=@{status='unavailable';behind=$null;ahead=$null}
if ($authority.status -eq 'verified') { $defaultDivergence=Get-Divergence $authority.baseline_sha $head }
$upstream=Invoke-Git @('rev-parse','--verify','@{upstream}')
$upstreamDivergence=@{status='unavailable';behind=$null;ahead=$null}
if ($upstream.ok) { $upstreamDivergence=Get-Divergence $upstream.output.Trim() $head }
$upstreamDivergence.freshness='local-tracking-cache'
# NUL records preserve unusual filenames and both sides of a rename.
$paths=@{}; $dirty=Invoke-Git @('status','--porcelain=v1','-z','--untracked-files=all')
$records=$dirty.output.Split([char]0)
for ($i=0;$i -lt $records.Length;$i++) {
  $r=$records[$i]; if ($r.Length -lt 4) { continue }
  $code=$r.Substring(0,2); $paths[$r.Substring(3)]=$code
  if ($code -match '[RC]' -and $i+1 -lt $records.Length) { $i++; $paths[$records[$i]]='rename-source' }
}
$known=@('todo.md','todo-done.md','docs/plans','docs/brainstorms','docs/results','docs/handoffs','docs/context/kb')
$ignored=Invoke-Git (@('ls-files','--others','--ignored','--exclude-standard','-z','--')+$known)
foreach ($path in $ignored.output.Split([char]0)) { if ($path) { $paths[$path]='ignored-workflow' } }
$inventory=@(); $complete=$dirty.ok -and $ignored.ok
foreach ($path in @($paths.Keys | Sort-Object)) {
  if ($inventory.Count -ge 512) { $complete=$false; break }
  $full=Join-Path $rootPath $path
  $entry=[ordered]@{path=$path;status=$paths[$path];sha256=$null;protection='preserve';size=$null}
  try {
    $item=Get-Item -LiteralPath $full -Force -ErrorAction Stop
    if (($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -or $item.PSIsContainer) { $entry.protection='preserve-link-or-directory' }
    elseif ($item.Length -gt 5MB) { $entry.size=$item.Length; $entry.protection='preserve-large'; $complete=$false }
    else { $entry.size=$item.Length; $entry.sha256=Get-FileDigest $full }
  } catch { if (Test-Path -LiteralPath $full) { $complete=$false }; $entry.protection='preserve-missing-or-unreadable' }
  $inventory+=$entry
}
if (-not $complete) { $limitations.Add('Inventory incomplete or bounded limit reached; dependent copies need a fresh explicit allowlist.') }
$dirtyFingerprint=Get-Hash (ConvertTo-Json -InputObject @($inventory) -Depth 5 -Compress)
$worktrees=(Invoke-Git @('worktree','list','--porcelain')).output; $protections=@()
foreach ($block in ($worktrees -split '\r?\n\r?\n')) {
  if ($block -match '(?m)^worktree (.+)$') {
    $wt=$Matches[1].Trim(); $wtBranch=''; if ($block -match '(?m)^branch (.+)$') { $wtBranch=$Matches[1].Trim() }
    $protections+=@{kind='occupied-worktree';path=$wt;branch=$wtBranch;protection='preserve-live'}
  }
}
$queuePath=Join-Path $common '.copilot-kb/work-queue.json'
if (Test-Path -LiteralPath $queuePath) {
  try {
    foreach ($claim in @([IO.File]::ReadAllText($queuePath) | ConvertFrom-Json)) {
      if ($claim.status -notin @('done','retired','superseded','delivery-integrated')) { $protections+=@{kind='work-claim';branch=[string]$claim.branch;protection='preserve-live'} }
    }
  } catch { $protections+=@{kind='unreadable-work-queue';protection='preserve-unknown'}; $limitations.Add('Unreadable claim queue: resolve ownership before writes.') }
}
# Read bounded known declarations only; never execute their commands or report content.
$candidates=@(); $declared=Invoke-Git (@('ls-files','--cached','--others','--exclude-standard','-z','--')+$known)
$declarationPaths=@(($declared.output.Split([char]0)+$ignored.output.Split([char]0)) | Where-Object { $_ } | Sort-Object -Unique | Select-Object -First 256)
foreach ($path in $declarationPaths) {
  if ($path -notmatch '\.(md|json|ya?ml)$') { continue }
  $full=Join-Path $rootPath $path; if (-not (Test-Path -LiteralPath $full -PathType Leaf)) { continue }
  $item=Get-Item -LiteralPath $full -Force
  if ($item.Length -gt 256KB -or ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)) { continue }
  $content=[IO.File]::ReadAllText($full)
  $refs=@([regex]::Matches($content,'(?m)^\s*(?:branch|ref|head|tip|source_sha256|proof_sha256)\s*:\s*["\x27]?([A-Za-z0-9_./-]+)') | ForEach-Object { $_.Groups[1].Value })
  $links=@([regex]::Matches($content,'docs/(?:plans|results)/[A-Za-z0-9_./-]+\.(?:md|json)') | ForEach-Object { $_.Value } | Sort-Object -Unique)
  $candidates+=@{path=$path;status='candidate-unproven';declared_refs_or_hashes=$refs;artifact_links=$links;evidence='not-validated'}
}
$localBranches=@()
$branchRefs=Invoke-Git @('for-each-ref','--format=%(refname)%09%(objectname)','refs/heads/')
$branchesComplete=$branchRefs.ok
foreach ($line in ($branchRefs.output -split '\r?\n')) {
  if (-not $line) { continue }
  if ($localBranches.Count -ge 256) { $branchesComplete=$false; break }
  $parts=$line.Split([char]9)
  if ($parts.Length -ne 2) { $branchesComplete=$false; continue }
  $div=@{status='unavailable';ahead=$null;behind=$null}
  if ($authority.status -eq 'verified') { $div=Get-Divergence $authority.baseline_sha $parts[1] }
  $live=@($protections | Where-Object { $_.branch -eq $parts[0] -or ('refs/heads/'+$_.branch) -eq $parts[0] }).Count -gt 0
  $localBranches+=[ordered]@{ref=$parts[0];tip=$parts[1];divergence=$div;protected=$live;containment='not-disposal-proof'}
}
if (-not $branchesComplete) { $limitations.Add('Local branch inventory incomplete; unenumerated branches remain preserved.') }
$surveyIndex=(Invoke-Git @('rev-parse','--git-path','index')).output.Trim()
if (-not [IO.Path]::IsPathRooted($surveyIndex)) { $surveyIndex=Join-Path $rootPath $surveyIndex }
$surveyIndexHash=$null
if (Test-Path -LiteralPath $surveyIndex -PathType Leaf) { $surveyIndexHash=Get-FileDigest $surveyIndex }
$result=[ordered]@{
  schema_version=1;action='survey';status='surveyed';repository=$rootPath;common_dir=$common
  repository_id=(Get-Hash $common.ToLowerInvariant());branch=$branch;head=$head;authority=$authority
  default_divergence=$defaultDivergence;upstream_divergence=$upstreamDivergence
  dirty_paths=@($inventory);dirty_fingerprint=$dirtyFingerprint;inventory_complete=$complete;index_sha256=$surveyIndexHash
  protections=@($protections);candidates=@($candidates);pairing_status='candidates-unproven';policy=$policy
  local_branches=@($localBranches);branches_complete=$branchesComplete
  capabilities=@{git=$true;native_kbcheck=(Test-Path -LiteralPath (Join-Path $rootPath 'cmd/kbcheck'));native_required=$false}
  eligibility=@{merge=$false;delete=$false};limitations=@($limitations)
  next_actions=@('continue-independent-work','offer-scoped-cleanup-once')
}
if ($Action -ne 'survey') {
  $entrypoint=$PSCommandPath
  . (Join-Path $PSScriptRoot 'recovery_prepare.ps1')
  if ($Action -eq 'continue') {
    . (Join-Path $PSScriptRoot 'recovery_continue.ps1')
    $result=Invoke-Continuation $result
  } else { $result=Invoke-Preparation $result }
}
if ($Json) { $result | ConvertTo-Json -Depth 10 } else { $result }
