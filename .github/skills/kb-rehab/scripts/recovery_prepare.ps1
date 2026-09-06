# Loaded by recovery.ps1; no independent entrypoint or external runtime.
function Assert-PlainPath([string]$Path) {
  $p=[IO.Path]::GetFullPath($Path)
  while ($p) {
    if (Test-Path -LiteralPath $p) {
      $item=Get-Item -LiteralPath $p -Force
      if ($item.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'path-link-or-junction' }
    }
    $parent=[IO.Directory]::GetParent($p)
    if (-not $parent) { break }; $p=$parent.FullName
  }
}
function Get-ContainedPath([string]$Base,[string]$Relative) {
  if (-not $Relative -or [IO.Path]::IsPathRooted($Relative) -or $Relative -match '[:\\]|(^|/)\.\.?(/|$)|[. ]($|/)' ) { throw 'invalid-relative-artifact' }
  $full=[IO.Path]::GetFullPath((Join-Path $Base $Relative))
  if (-not $full.StartsWith($Base.TrimEnd('\','/')+[IO.Path]::DirectorySeparatorChar,[StringComparison]::OrdinalIgnoreCase)) { throw 'path-escape' }
  Assert-PlainPath $full
  return $full
}
function Write-RecoveryReceipt([string]$Path,$Value) {
  $temp=$Path+'.'+[Guid]::NewGuid().ToString('N')+'.tmp'
  [IO.File]::WriteAllText($temp,($Value | ConvertTo-Json -Depth 12),(New-Object Text.UTF8Encoding($false)))
  if (Test-Path -LiteralPath $Path) { [IO.File]::Replace($temp,$Path,[NullString]::Value) } else { [IO.File]::Move($temp,$Path) }
}
function Invoke-Preparation($Survey) {
  $lock=$null; $receiptPath=$null
  $failure=@{schema_version=1;action=$Action;status='dependency-needed';next_action='continue-independent-work';reason='request-required';receipt=$null}
  try {
    if (-not $Request -or -not (Test-Path -LiteralPath $Request -PathType Leaf)) { return $failure }
    Assert-PlainPath $Request
    $requestHash=Get-FileDigest $Request
    $q=[IO.File]::ReadAllText($Request) | ConvertFrom-Json
    if ($q.schema_version -ne 1 -or $q.run_id -isnot [string] -or $q.run_id -notmatch '^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$' -or $q.objective -isnot [string] -or -not $q.objective) { throw 'invalid-request-identity' }
    if ($q.authority.source -ne 'current-run' -or $q.authority.prepare -isnot [bool] -or -not $q.authority.prepare -or $q.authority.objective -ne $q.objective -or ($null -ne $q.authority.paused -and ($q.authority.paused -isnot [bool] -or $q.authority.paused))) { throw 'current-preparation-authority-required' }
    if ($q.dependency_status -ne 'independent') { throw 'dependency-needed' }
    if ($Survey.authority.status -ne 'verified') { throw 'remote-baseline-unavailable' }
    if ($q.survey.repository_id -ne $Survey.repository_id -or $q.survey.head -ne $Survey.head -or $q.survey.index_sha256 -ne $Survey.index_sha256 -or $q.survey.dirty_fingerprint -ne $Survey.dirty_fingerprint -or $q.survey.authority.baseline_sha -ne $Survey.authority.baseline_sha) { throw 'survey-refresh-required' }
    if (-not $Survey.inventory_complete -or $Survey.policy.status -eq 'invalid') { throw 'inventory-or-policy-incomplete' }
    $baseline=[string]$Survey.authority.baseline_sha
    $tree=Invoke-Git @('ls-tree','-r','-z',$baseline)
    if (-not $tree.ok -or $tree.output -match '(?:^|\x00)(120000|160000) ') { throw 'unsupported-baseline-link-or-submodule' }
    $topic=[string]$q.branch
    if ($topic -notmatch '^codex/[a-zA-Z0-9][a-zA-Z0-9/_-]*$' -or ('refs/heads/'+$topic) -eq $Survey.authority.default_ref) { throw 'nondefault-topic-required' }
    # A single sibling container prevents arbitrary output roots and nesting.
    $container=Join-Path ([IO.Directory]::GetParent($rootPath).FullName) '.kb-recovery-worktrees'
    $dest=[IO.Path]::GetFullPath([string]$q.destination)
    if (-not $dest.StartsWith($container+[IO.Path]::DirectorySeparatorChar,[StringComparison]::OrdinalIgnoreCase) -or [IO.Directory]::GetParent($dest).FullName -ne $container) { throw 'destination-outside-recovery-container' }
    Assert-PlainPath $rootPath; Assert-PlainPath $common; Assert-PlainPath $dest
    $artifacts=@(); $seen=@{}
    foreach ($a in @($q.artifacts)) {
      $relative=[string]$a.path
      if ($relative -notmatch '^(docs/(plans|brainstorms|handoffs)/.+\.(md|json|ya?ml)|todo(-done)?\.md)$' -or $relative -match '(?i)(credential|secret|token|password|\.env|id_rsa|id_ed25519|\.pem|\.key)' -or $a.sha256 -notmatch '^[a-f0-9]{64}$') { throw 'artifact-not-allowed' }
      if ($seen.ContainsKey($relative)) { throw 'artifact-case-collision' }; $seen[$relative]=$true
      $source=Get-ContainedPath $rootPath $relative; $target=Get-ContainedPath $dest $relative
      if (-not (Test-Path -LiteralPath $source -PathType Leaf) -or (Get-FileDigest $source) -ne $a.sha256) { throw 'source-artifact-changed' }
      # Do not overwrite a baseline-owned path, even if selected accidentally.
      if ((Invoke-Git @('cat-file','-e',($baseline+':'+$relative))).ok) { throw 'artifact-already-in-baseline' }
      $artifacts+=@{path=$relative;sha256=[string]$a.sha256;source=$source;target=$target}
    }
    $queueDir=Join-Path $common '.copilot-kb'; Assert-PlainPath $queueDir
    [void][IO.Directory]::CreateDirectory($queueDir)
    $lockPath=Join-Path $queueDir 'work-queue.lock'; Assert-PlainPath $lockPath
    try { $lock=[IO.File]::Open($lockPath,[IO.FileMode]::OpenOrCreate,[IO.FileAccess]::ReadWrite,[IO.FileShare]::None) } catch { throw 'ownership-lock-busy' }
    $fresh=(& $entrypoint -Action survey -Root $rootPath -Json | Out-String) | ConvertFrom-Json
    if ($fresh.head -ne $Survey.head -or $fresh.dirty_fingerprint -ne $Survey.dirty_fingerprint -or $fresh.authority.baseline_sha -ne $baseline -or $fresh.authority.status -ne 'verified') { throw 'survey-changed-before-write' }
    foreach ($p in @($fresh.protections)) {
      if ($p.kind -eq 'unreadable-work-queue') { throw 'ownership-unknown' }
      if ($p.kind -eq 'work-claim' -and $p.branch -eq $topic) { throw 'destination-claimed' }
    }
    $stateDir=Join-Path $queueDir 'recovery'; Assert-PlainPath $stateDir; [void][IO.Directory]::CreateDirectory($stateDir)
    $receiptPath=Join-Path $stateDir ($q.run_id+'.json'); Assert-PlainPath $receiptPath
    $indexPath=(Invoke-Git @('rev-parse','--git-path','index')).output.Trim()
    if (-not [IO.Path]::IsPathRooted($indexPath)) { $indexPath=Join-Path $rootPath $indexPath }
    $indexHash=Get-FileDigest $indexPath
    $receipt=$null
    if (Test-Path -LiteralPath $receiptPath) {
      $receipt=[IO.File]::ReadAllText($receiptPath) | ConvertFrom-Json
      if ($receipt.request_sha256 -ne $requestHash -or $receipt.destination -ne $dest -or $receipt.source_index_sha256 -ne $indexHash) { throw 'receipt-refresh-required' }
    } else {
      if ($Action -eq 'verify') { throw 'receipt-missing' }
      if ((Test-Path -LiteralPath $dest) -or (Invoke-Git @('show-ref','--verify','--quiet',('refs/heads/'+$topic))).ok) { throw 'destination-occupied' }
      $receipt=[pscustomobject]@{schema_version=1;run_id=$q.run_id;objective=$q.objective;request_sha256=$requestHash;repository_id=$Survey.repository_id;source_head=$Survey.head;source_dirty_fingerprint=$Survey.dirty_fingerprint;source_index_sha256=$indexHash;baseline_sha=$baseline;branch=$topic;destination=$dest;state='reserved';artifacts=@($q.artifacts);next_action='continue-independent-work'}
      Write-RecoveryReceipt $receiptPath $receipt
    }
    if ((Get-FileDigest $Request) -ne $requestHash) { throw 'authority-request-changed' }
    if (-not (Test-Path -LiteralPath $dest)) {
      if ($Action -eq 'verify') { throw 'destination-missing' }
      # No checkout means no hooks or source-configured smudge/process execution.
      $reservedBranch=Invoke-Git @('rev-parse','--verify',('refs/heads/'+$topic))
      if ($reservedBranch.ok) {
        if ($receipt.state -ne 'reserved' -or $reservedBranch.output.Trim() -ne $baseline) { throw 'reserved-branch-changed' }
        $created=Invoke-Git @('worktree','add','--no-checkout','--',$dest,$topic)
      } else { $created=Invoke-Git @('worktree','add','--no-checkout','-b',$topic,'--',$dest,$baseline) }
      if (-not $created.ok) { throw 'worktree-create-failed-preserved' }
    }
    Assert-PlainPath $dest
    $savedRoot=$rootPath
    try {
      $rootPath=$dest
      $actualCommon=(Invoke-Git @('rev-parse','--git-common-dir')).output.Trim()
      if (-not [IO.Path]::IsPathRooted($actualCommon)) { $actualCommon=Join-Path $dest $actualCommon }
      if ([IO.Path]::GetFullPath($actualCommon) -ne $common -or (Invoke-Git @('rev-parse','HEAD')).output.Trim() -ne $baseline -or (Invoke-Git @('branch','--show-current')).output.Trim() -ne $topic) { throw 'destination-identity-changed' }
      if ($receipt.state -eq 'reserved') {
        if ($Action -eq 'verify') { throw 'preparation-incomplete' }
        $safe=@('-c','core.symlinks=false')
        if ((Get-FileDigest $Request) -ne $requestHash) { throw 'authority-request-changed' }
        $destinationIndex=(Invoke-Git @('rev-parse','--git-path','index')).output.Trim()
        if (-not [IO.Path]::IsPathRooted($destinationIndex)) { $destinationIndex=Join-Path $dest $destinationIndex }
        if (Test-Path -LiteralPath $destinationIndex) {
          if (-not (Invoke-Git @('diff','--cached','--quiet','HEAD','--')).ok) { throw 'destination-staged-work-preserved' }
        } elseif (-not (Invoke-Git @('read-tree',$baseline)).ok) { throw 'baseline-index-failed' }
        if (-not (Invoke-Git @('diff','--quiet','--diff-filter=ACMRTUXB','HEAD','--')).ok) { throw 'partial-baseline-checkout-modified' }
        # Materialize missing baseline files only; never force over resumed edits.
        $missing=Invoke-Git @('ls-files','--deleted','-z')
        foreach ($path in $missing.output.Split([char]0)) {
          if ($path) {
            if ((Get-FileDigest $Request) -ne $requestHash) { throw 'authority-request-changed' }
            [void](Get-ContainedPath $dest $path)
            if (-not (Invoke-Git ($safe+@('checkout-index','--',$path))).ok) { throw 'partial-baseline-checkout-preserved' }
          }
        }
        $receipt.state='baseline-ready'; Write-RecoveryReceipt $receiptPath $receipt
      }
      $tracked=Invoke-Git @('diff','--quiet','HEAD','--')
      if (-not $tracked.ok) { throw 'destination-source-changed' }
    } finally { $rootPath=$savedRoot }
    foreach ($a in $artifacts) {
      Assert-PlainPath $a.source; Assert-PlainPath $a.target
      if ((Get-FileDigest $a.source) -ne $a.sha256 -or (Get-FileDigest $Request) -ne $requestHash) { throw 'source-or-authority-changed' }
      if (Test-Path -LiteralPath $a.target) {
        if (-not (Test-Path -LiteralPath $a.target -PathType Leaf) -or (Get-FileDigest $a.target) -ne $a.sha256) { throw 'destination-artifact-changed' }
      } else {
        if ($Action -eq 'verify') { throw 'artifact-copy-incomplete' }
        [void][IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($a.target)); Assert-PlainPath $a.target
        $temp=$a.target+'.'+$q.run_id+'.copy'
        if (Test-Path -LiteralPath $temp) { if ((Get-FileDigest $temp) -ne $a.sha256) { throw 'partial-copy-needs-review' } }
        else { [IO.File]::Copy($a.source,$temp,$false) }
        if ((Get-FileDigest $temp) -ne $a.sha256 -or (Get-FileDigest $a.source) -ne $a.sha256 -or (Get-FileDigest $Request) -ne $requestHash) { throw 'copy-race-preserved' }
        Assert-PlainPath $a.target; [IO.File]::Move($temp,$a.target)
      }
    }
    $after=(& $entrypoint -Action survey -Root $rootPath -Json | Out-String) | ConvertFrom-Json
    if ($after.head -ne $Survey.head -or $after.dirty_fingerprint -ne $Survey.dirty_fingerprint -or (Get-FileDigest $indexPath) -ne $indexHash -or $after.authority.baseline_sha -ne $baseline -or (Get-FileDigest $Request) -ne $requestHash) { throw 'source-baseline-or-authority-refresh-required' }
    $receipt.state='prepared'; $receipt.next_action='kb-work'
    if ($Action -ne 'verify') { Write-RecoveryReceipt $receiptPath $receipt }
    return @{schema_version=1;action=$Action;status='prepared';receipt=$receiptPath;destination=$dest;baseline_sha=$baseline;next_action='kb-work';source_preserved=$true}
  } catch { $failure.reason=$_.Exception.Message; $failure.receipt=$receiptPath; return $failure }
  finally { if ($lock) { $lock.Dispose() } }
}
