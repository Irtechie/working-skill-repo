# Exact accepted local scope. Forge actions are delegated, never synthesized here.
function Get-RecoveryDeliveryPolicy {
  $value=@{mode='pr';merge='manual';known=$true}
  $path=Join-Path $rootPath 'docs/context/operations/kb-routing.yaml'
  if (-not (Test-Path -LiteralPath $path -PathType Leaf)) { return $value }
  $text=[IO.File]::ReadAllText($path)
  if (@([regex]::Matches($text,'(?m)^\s*delivery\s*:')).Count -gt 1 -or $text -match '(?m)^[ \t]+delivery\s*:') { $value.known=$false;return $value }
  # Blank/comment-only lines remain inside the mapping; a substantive column-one
  # key ends it. Do not let a separator hide a later local mode or duplicate.
  if ($text -match '(?m)^delivery:[ \t]*\r?\n((?:(?:[ \t]+[^\r\n]*|#[^\r\n]*|)(?:\r?\n|$))*)') {
    $section=$Matches[1]
    if ($section -match '(&|\*|<<\s*:)' ) { $value.known=$false;return $value }
    foreach ($key in @('mode','merge')) {
      if (@([regex]::Matches($section,('(?m)^\s+'+$key+':'))).Count -gt 1) { $value.known=$false;return $value }
      if ($section -match ('(?m)^\s+'+$key+':\s*["\x27]?([a-z-]+)["\x27]?\s*(?:#.*)?$')) { $value[$key]=$Matches[1] }
      elseif ($section -match ('(?m)^\s+'+$key+':')) { $value.known=$false }
    }
    if ($value.mode -notin @('pr','local','direct') -or $value.merge -notin @('manual','auto-after-checks')) { $value.known=$false }
  } elseif ($text -match '(?m)^delivery:') { $value.known=$false }
  return $value
}
function Get-RecoveryPolicyHash {
  $values=@()
  foreach ($p in @('config/rehab-policy.json','docs/context/operations/kb-routing.yaml')) {
    $full=Join-Path $rootPath $p; Assert-PlainPath $full
    if (Test-Path -LiteralPath $full -PathType Leaf) { $values+=($p+'|'+(Get-FileDigest $full)) }
  }
  return Get-Hash ($values -join '|')
}
function Remove-ConfirmedGeneratedArtifact($Item,$Survey,[string]$StateRoot,[string]$RequestHash,[string]$PolicyHash) {
  if ($Item.rejected -isnot [bool] -or -not $Item.rejected -or $Item.path -notmatch '^\.kb/(generated|tmp)/[^:]+$') { throw 'generated-output-rejection-required' }
  if (@($Survey.protections | Where-Object { $_.kind -in @('work-claim','unreadable-work-queue') }).Count -gt 0) { throw 'live-artifact-ownership-preserved' }
  $source=Get-ContainedPath $rootPath ([string]$Item.path)
  $provenance=Get-ContainedPath $rootPath ([string]$Item.provenance.path)
  if ($Item.provenance.sha256 -notmatch '^[a-f0-9]{64}$' -or (Get-FileDigest $provenance) -ne $Item.provenance.sha256) { throw 'generated-provenance-unverified' }
  $record=[IO.File]::ReadAllText($provenance) | ConvertFrom-Json
  if ($record.kind -ne 'generated-output' -or $record.repository_id -ne $Survey.repository_id -or $record.path -ne $Item.path -or $record.sha256 -ne $Item.sha256 -or $record.producer -isnot [string] -or -not $record.producer) { throw 'generated-provenance-mismatch' }
  $archive=Join-Path $StateRoot ((Get-Hash ($Item.path+'|'+$Item.sha256)).Substring(0,32)+'.artifact')
  Assert-PlainPath $archive;[void][IO.Directory]::CreateDirectory($archive)
  $copy=Join-Path $archive 'content';$restore=Join-Path $archive 'restored-content'
  Assert-PlainPath $copy;Assert-PlainPath $restore
  if ((Get-FileDigest $Request) -ne $RequestHash -or (Get-RecoveryPolicyHash) -ne $PolicyHash) { throw 'acceptance-or-policy-changed' }
  if (Test-Path -LiteralPath $source) {
    if ((Get-Item -LiteralPath $source).Length -gt 5MB -or (Get-FileDigest $source) -ne $Item.sha256) { throw 'generated-artifact-changed' }
    if (-not (Test-Path -LiteralPath $copy)) { [IO.File]::Copy($source,$copy,$false) }
  }
  if ((Get-FileDigest $copy) -ne $Item.sha256) { throw 'generated-archive-changed' }
  if (-not (Test-Path -LiteralPath $restore)) { [IO.File]::Copy($copy,$restore,$false) }
  if ((Get-FileDigest $restore) -ne $Item.sha256) { throw 'generated-restore-failed' }
  $receiptPath=Join-Path $archive 'receipt.json'
  $receipt=@{path=$Item.path;sha256=$Item.sha256;request_sha256=$RequestHash;provenance_sha256=$Item.provenance.sha256;state='archive-verified';archive=$archive}
  Write-RecoveryReceipt $receiptPath $receipt
  # Refresh ownership and accepted bytes immediately before the exact deletion.
  $fresh=(& $entrypoint -Action survey -Root $rootPath -Json | Out-String) | ConvertFrom-Json
  if (@($fresh.protections | Where-Object { $_.kind -in @('work-claim','unreadable-work-queue') }).Count -gt 0 -or (Get-FileDigest $Request) -ne $RequestHash -or (Get-RecoveryPolicyHash) -ne $PolicyHash) { throw 'ownership-acceptance-or-policy-changed' }
  Assert-PlainPath $source
  if (Test-Path -LiteralPath $source) {
    if ((Get-FileDigest $source) -ne $Item.sha256) { throw 'generated-artifact-changed' }
    [IO.File]::Delete($source)
  }
  $receipt.state='removed';Write-RecoveryReceipt $receiptPath $receipt
  return $archive
}
function Assert-CurrentDisposition($Survey,[string]$RequestHash,[string]$PolicyHash,[string]$Ref,[string]$Tip) {
  if ((Get-FileDigest $Request) -ne $RequestHash) { throw 'acceptance-changed' }
  if ((Get-RecoveryPolicyHash) -ne $PolicyHash) { throw 'policy-changed' }
  $fresh=(& $entrypoint -Action survey -Root $rootPath -Json | Out-String) | ConvertFrom-Json
  if ($fresh.repository_id -ne $Survey.repository_id -or $fresh.head -ne $Survey.head -or $fresh.index_sha256 -ne $Survey.index_sha256 -or $fresh.dirty_fingerprint -ne $Survey.dirty_fingerprint) { throw 'source-changed-preserved' }
  if ($fresh.authority.status -ne 'verified' -or $fresh.authority.baseline_sha -ne $Survey.authority.baseline_sha) { throw 'remote-baseline-refresh-required' }
  $branch=@($fresh.local_branches | Where-Object { $_.ref -eq $Ref })
  if ($branch.Count -ne 1 -or $branch[0].tip -ne $Tip) { throw 'accepted-tip-changed' }
  if ($branch[0].protected -or $Ref -eq $fresh.authority.default_ref -or $Ref -eq ('refs/heads/'+$fresh.branch)) { throw 'live-or-default-work-preserved' }
  if (@($fresh.protections | Where-Object { $_.kind -eq 'unreadable-work-queue' }).Count -gt 0) { throw 'ownership-unknown' }
}
function Restore-RecoveryArchive([string]$Archive,[string]$Ref,[string]$Tip,$Artifacts,[string]$ExpectedBundleHash) {
  $bundle=Join-Path $Archive 'branch.bundle'; Assert-PlainPath $bundle
  if ((Get-FileDigest $bundle) -ne $ExpectedBundleHash) { throw 'archive-bundle-changed' }
  $heads=Invoke-Git @('bundle','list-heads',$bundle)
  if (-not $heads.ok -or $heads.output.Trim() -ne ($Tip+' '+$Ref)) { throw 'archive-head-mismatch' }
  $restore=Join-Path $Archive 'restore.git'; Assert-PlainPath $restore
  if (-not (Test-Path -LiteralPath $restore)) { if (-not (Invoke-Git @('init','--bare','--',$restore)).ok) { throw 'restore-init-failed' } }
  $savedRoot=$rootPath
  try {
    $rootPath=$restore
    if (-not (Invoke-Git @('bundle','verify',$bundle)).ok) { throw 'archive-restore-verification-failed' }
    $existing=Invoke-Git @('rev-parse','--verify',$Ref)
    if ($existing.ok -and $existing.output.Trim() -ne $Tip) { throw 'restore-ref-changed' }
    if (-not $existing.ok -and -not (Invoke-Git @('fetch','--quiet','--no-tags','--no-write-fetch-head','--refmap=','--',$bundle,($Ref+':'+$Ref))).ok) { throw 'archive-restore-fetch-failed' }
    if ((Invoke-Git @('rev-parse','--verify',$Ref)).output.Trim() -ne $Tip -or -not (Invoke-Git @('fsck','--full','--no-reflogs')).ok) { throw 'archive-restore-verification-failed' }
  } finally { $rootPath=$savedRoot }
  foreach ($a in $Artifacts) {
    $archived=Get-ContainedPath (Join-Path $Archive 'artifacts') $a.path
    $restored=Get-ContainedPath (Join-Path $Archive 'restored-artifacts') $a.path
    if ((Get-FileDigest $archived) -ne $a.sha256) { throw 'artifact-archive-changed' }
    [void][IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($restored)); Assert-PlainPath $restored
    if (-not (Test-Path -LiteralPath $restored)) { [IO.File]::Copy($archived,$restored,$false) }
    if ((Get-FileDigest $restored) -ne $a.sha256) { throw 'artifact-restore-verification-failed' }
  }
}
function Invoke-Disposition($Survey) {
  $result=@{schema_version=1;action='dispose';status='retained-blocked';items=@();next_action='continue-independent-work';reason='request-required'}
  $lock=$null
  try {
    if (-not $Request -or -not (Test-Path -LiteralPath $Request -PathType Leaf)) { return $result }
    Assert-PlainPath $Request; $requestHash=Get-FileDigest $Request
    $q=[IO.File]::ReadAllText($Request) | ConvertFrom-Json
    if ($q.schema_version -ne 1 -or $q.run_id -isnot [string] -or $q.run_id -notmatch '^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$' -or $q.repository_id -ne $Survey.repository_id -or $q.objective -isnot [string] -or -not $q.objective) { throw 'invalid-disposition-identity' }
    $accept=$q.acceptance
    if ($accept.source -notin @('current-user-reply','explicit-kb-rehab') -or $accept.accepted -isnot [bool] -or -not $accept.accepted -or ($null -ne $accept.revoked -and ($accept.revoked -isnot [bool] -or $accept.revoked)) -or ($null -ne $accept.paused -and ($accept.paused -isnot [bool] -or $accept.paused))) { throw 'current-acceptance-required' }
    $policyHash=Get-RecoveryPolicyHash
    $deliveryPolicy=Get-RecoveryDeliveryPolicy
    $nativePresent=$Survey.capabilities.native_kbcheck -or (Test-Path -LiteralPath (Join-Path $rootPath 'cmd/kbreconcile')) -or $null -ne (Get-Command kbreconcile -CommandType Application -ErrorAction SilentlyContinue)
    $queue=Join-Path $common '.copilot-kb'; Assert-PlainPath $queue; [void][IO.Directory]::CreateDirectory($queue)
    $lockPath=Join-Path $queue 'work-queue.lock'; Assert-PlainPath $lockPath
    try { $lock=[IO.File]::Open($lockPath,[IO.FileMode]::OpenOrCreate,[IO.FileAccess]::ReadWrite,[IO.FileShare]::None) } catch { throw 'ownership-lock-busy' }
    $stateRoot=Join-Path $queue ('recovery/dispositions/'+$q.run_id); Assert-PlainPath $stateRoot; [void][IO.Directory]::CreateDirectory($stateRoot)
    foreach ($item in @($q.items)) {
      $row=@{ref=$item.ref;path=$item.path;disposition='review-needed';reason='missing-ownership-or-proof';owner='kb-rehab';completed=$false}
      try {
        $scope=@($accept.items | Where-Object { $_.kind -eq $item.kind -and (($item.kind -eq 'branch' -and $_.ref -eq $item.ref -and $_.tip -eq $item.tip) -or ($item.kind -eq 'artifact' -and $_.path -eq $item.path -and $_.sha256 -eq $item.sha256)) })
        if ($scope.Count -ne 1) { throw 'item-outside-accepted-inventory' }
        if ($item.kind -eq 'artifact') {
          $path=Get-ContainedPath $rootPath ([string]$item.path)
          if ($item.path -match '(?i)(^|/)(\.git|\.env|credentials?|secrets?)(/|\.|$)|(?i)(token|password|id_rsa|id_ed25519|\.pem|\.key)' ) { $row.disposition='preserve-live';$row.reason='credential-preserved-in-place' }
          elseif ($item.decision -eq 'discard-generated') {
            if ($nativePresent -or $q.native_refusal -or $Survey.policy.status -eq 'invalid') { throw 'native-owner-or-valid-policy-required' }
            $row.archive=Remove-ConfirmedGeneratedArtifact $item $Survey $stateRoot $requestHash $policyHash
            $row.disposition='discard-confirmed-junk';$row.reason='explicitly-rejected-generated-output-restored';$row.completed=$true
          }
          elseif (-not (Test-Path -LiteralPath $path -PathType Leaf) -or (Get-FileDigest $path) -ne $item.sha256) { throw 'artifact-changed' }
          else { $row.disposition='salvage';$row.reason='source-artifact-preserved';$row.owner='kb-complete' }
        } elseif ($item.kind -eq 'branch' -and $item.ref -match '^refs/heads/[A-Za-z0-9_./-]+$' -and $item.tip -match '^[a-f0-9]{40,64}$') {
          $key=(Get-Hash ($item.ref+'|'+$item.tip)).Substring(0,32)
          $receiptPath=Join-Path $stateRoot ($key+'.json'); Assert-PlainPath $receiptPath
          $receipt=$null; if (Test-Path -LiteralPath $receiptPath) { $receipt=[IO.File]::ReadAllText($receiptPath) | ConvertFrom-Json }
          $expectedArchive=Join-Path $stateRoot ($key+'.archive')
          if ($receipt -and ($receipt.archive -ne $expectedArchive -or $receipt.ref -ne $item.ref -or $receipt.tip -ne $item.tip -or $receipt.repository_id -ne $Survey.repository_id)) { throw 'archive-receipt-identity-mismatch' }
          $current=Invoke-Git @('rev-parse','--verify',$item.ref)
          if ($receipt -and $receipt.state -in @('archive-verified','retired') -and -not $current.ok) {
            if ($receipt.request_sha256 -ne $requestHash) { throw 'retirement-receipt-scope-changed' }
            Restore-RecoveryArchive $receipt.archive $item.ref $item.tip @($receipt.artifacts) $receipt.bundle_sha256
            if ($receipt.state -ne 'retired') { $receipt.state='retired';Write-RecoveryReceipt $receiptPath $receipt }
            $row.disposition='discard-confirmed-junk';$row.reason='previously-retired-restorable';$row.completed=$true;$row.archive=$receipt.archive
          } elseif (-not $current.ok -or $current.output.Trim() -ne $item.tip) { throw 'accepted-tip-changed' }
          elseif ($item.decision -eq 'reject' -and (@($Survey.local_branches | Where-Object { $_.ref -eq $item.ref -and $_.protected }).Count -gt 0 -or $item.ref -eq $Survey.authority.default_ref)) { $row.disposition='preserve-live';$row.reason='occupied-or-default-branch' }
          elseif ($item.decision -in @('deliver','merge')) {
            if (-not $item.manifest -or $item.manifest_sha256 -notmatch '^[a-f0-9]{64}$') { throw 'owning-manifest-required' }
            $manifest=Get-ContainedPath $rootPath ([string]$item.manifest)
            if ((Get-FileDigest $manifest) -ne $item.manifest_sha256) { throw 'owning-manifest-changed' }
            $manifestText=[IO.File]::ReadAllText($manifest)
            $refPattern='(?m)^\s*(?:ref|branch):\s*["\x27]?(?:'+[regex]::Escape([string]$item.ref)+'|'+[regex]::Escape(([string]$item.ref).Substring(11))+')["\x27]?\s*$'
            $tipPattern='(?m)^\s*(?:tip|head):\s*["\x27]?'+[regex]::Escape([string]$item.tip)+'["\x27]?\s*$'
            if ($manifestText -notmatch $refPattern -or $manifestText -notmatch $tipPattern) { $row.reason='owning-manifest-ref-tip-unverified' }
            else {
              $row.manifest=$manifest;$row.tip=$item.tip
              if (-not $deliveryPolicy.known -or $deliveryPolicy.mode -eq 'local') { $row.disposition='retained-blocked';$row.owner='kb-complete';$row.reason='publishing-withheld-by-current-local-or-unknown-policy' }
              elseif ($item.decision -eq 'merge' -and (($accept.merge -is [bool] -and $accept.merge) -or $deliveryPolicy.merge -eq 'auto-after-checks')) {
                $row.disposition='merge-eligible';$row.owner='kb-land';$row.reason='authorized-pending-forge-observation';$row.forge_checks_required=$true
              } else { $row.disposition='deliver-pr';$row.owner='kb-complete';$row.reason='review-and-proof-required-in-own-workspace' }
            }
          } elseif ($item.decision -eq 'reject' -and $item.rejected -is [bool] -and $item.rejected) {
            if ($Survey.policy.status -eq 'invalid') { throw 'invalid-project-policy' }
            if ($nativePresent -or $q.native_refusal) { throw 'native-owner-required-no-weaker-retry' }
            Assert-CurrentDisposition $Survey $requestHash $policyHash $item.ref $item.tip
            $archive=Join-Path $stateRoot ($key+'.archive'); Assert-PlainPath $archive
            [void][IO.Directory]::CreateDirectory($archive)
            $artifacts=@();$excluded=@()
            foreach ($a in @($item.artifacts)) {
              if (-not $a) { continue }
              if ($a.path -match '(?i)(^|/)(\.git|\.env|credentials?|secrets?)(/|\.|$)|(token|password|id_rsa|id_ed25519|\.pem|\.key)') { $excluded+=@{path=$a.path;reason='credential-preserved-in-place'};continue }
              if (@($accept.items | Where-Object { $_.kind -eq 'artifact' -and $_.path -eq $a.path -and $_.sha256 -eq $a.sha256 }).Count -ne 1) { throw 'archive-artifact-outside-accepted-scope' }
              if ($a.sha256 -notmatch '^[a-f0-9]{64}$') { throw 'artifact-hash-required' }
              $source=Get-ContainedPath $rootPath ([string]$a.path)
              if (-not (Test-Path -LiteralPath $source -PathType Leaf) -or (Get-Item -LiteralPath $source).Length -gt 5MB -or (Get-FileDigest $source) -ne $a.sha256) { throw 'archive-artifact-changed-or-too-large' }
              $target=Get-ContainedPath (Join-Path $archive 'artifacts') ([string]$a.path)
              [void][IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($target));Assert-PlainPath $target
              if ((Get-FileDigest $Request) -ne $requestHash -or (Get-RecoveryPolicyHash) -ne $policyHash) { throw 'acceptance-or-policy-changed' }
              if (-not (Test-Path -LiteralPath $target)) { [IO.File]::Copy($source,$target,$false) }
              if ((Get-FileDigest $target) -ne $a.sha256 -or (Get-FileDigest $source) -ne $a.sha256) { throw 'artifact-copy-changed' }
              $artifacts+=@{path=[string]$a.path;sha256=[string]$a.sha256}
            }
            $bundle=Join-Path $archive 'branch.bundle';Assert-PlainPath $bundle
            if (-not (Test-Path -LiteralPath $bundle)) {
              Assert-CurrentDisposition $Survey $requestHash $policyHash $item.ref $item.tip
              if (-not (Invoke-Git @('bundle','create',$bundle,$item.ref)).ok) { throw 'archive-bundle-create-failed' }
            }
            $bundleHash=Get-FileDigest $bundle
            if ($receipt -and ($receipt.request_sha256 -ne $requestHash -or $receipt.bundle_sha256 -ne $bundleHash -or $receipt.policy_sha256 -ne $policyHash)) { throw 'archive-acceptance-or-policy-changed' }
            Restore-RecoveryArchive $archive $item.ref $item.tip $artifacts $bundleHash
            $receipt=@{schema_version=1;state='archive-verified';repository_id=$Survey.repository_id;request_sha256=$requestHash;policy_sha256=$policyHash;ref=$item.ref;tip=$item.tip;archive=$archive;bundle_sha256=$bundleHash;artifacts=$artifacts;excluded=$excluded}
            Write-RecoveryReceipt $receiptPath $receipt
            Assert-CurrentDisposition $Survey $requestHash $policyHash $item.ref $item.tip
            if (-not (Invoke-Git @('update-ref','-d',$item.ref,$item.tip)).ok) { throw 'ref-compare-and-swap-refused' }
            $receipt.state='retired';Write-RecoveryReceipt $receiptPath $receipt
            $row.disposition='discard-confirmed-junk';$row.reason='explicitly-rejected-restoration-verified';$row.completed=$true;$row.archive=$archive;$row.receipt=$receiptPath;$row.excluded=$excluded
          }
        } else { throw 'unsupported-item-kind-or-ref' }
      } catch { $row.disposition='retained-blocked';$row.reason=$_.Exception.Message }
      $result.items+= $row
    }
    $result.status='classified';$result.reason=$null
    return $result
  } catch { $result.reason=$_.Exception.Message;return $result }
  finally { if ($lock) { $lock.Dispose() } }
}
