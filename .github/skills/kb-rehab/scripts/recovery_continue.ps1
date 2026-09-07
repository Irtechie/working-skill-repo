# Advisory state and phase handoff only. Cleanup and forge execution stay with their owners.
function Invoke-Continuation($Survey) {
  $lock=$null
  $output=@{schema_version=1;action='continue';status='dependency-needed';next_action='continue-independent-work';notice=$null;cleanup=@{status='not-authorized';completed=$false;merge_authorized=$false;items=@()};preparation=$null;reason='request-required'}
  try {
    if (-not $Request -or -not (Test-Path -LiteralPath $Request -PathType Leaf)) { return $output }
    Assert-PlainPath $Request; $requestHash=Get-FileDigest $Request
    $q=[IO.File]::ReadAllText($Request) | ConvertFrom-Json
    if ($q.schema_version -ne 1 -or $q.run_id -isnot [string] -or $q.run_id -notmatch '^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$' -or $q.objective -isnot [string] -or -not $q.objective) { throw 'invalid-continuation-identity' }
    if ($q.authority.source -ne 'current-run' -or $q.authority.continue -isnot [bool] -or -not $q.authority.continue -or $q.authority.objective -ne $q.objective -or ($null -ne $q.authority.paused -and ($q.authority.paused -isnot [bool] -or $q.authority.paused))) { throw 'current-continuation-authority-required' }
    if (-not (Test-Path -LiteralPath $q.preparation_request -PathType Leaf)) { throw 'preparation-request-required' }
    Assert-PlainPath $q.preparation_request
    $preparationHash=Get-FileDigest $q.preparation_request
    $prep=[IO.File]::ReadAllText($q.preparation_request) | ConvertFrom-Json
    if ($prep.run_id -ne $q.run_id -or $prep.objective -ne $q.objective) { throw 'preparation-scope-mismatch' }
    # Omit this objective's own branch, so creating it cannot repeat the old offer.
    $items=@()
    foreach ($b in @($Survey.local_branches)) {
      if ($b.ref -eq ('refs/heads/'+$prep.branch) -or $b.ref -eq $Survey.authority.default_ref) { continue }
      if ($b.divergence.status -eq 'verified' -and $b.divergence.ahead -eq 0) { continue }
      $items+=[ordered]@{id=(Get-Hash ($b.ref+'|'+$b.tip));kind='branch';ref=$b.ref;tip=$b.tip;protected=$b.protected}
    }
    foreach ($a in @($Survey.dirty_paths)) {
      if (@($prep.artifacts | Where-Object { $_.path -eq $a.path }).Count -gt 0) { continue }
      $items+=[ordered]@{id=(Get-Hash ($Survey.head+'|'+$a.path+'|'+$a.sha256+'|'+$a.status));kind='artifact';path=$a.path;sha256=$a.sha256;source_head=$Survey.head;protected=$true}
    }
    $offerId=Get-Hash ($Survey.repository_id+'|'+$q.objective+'|'+(ConvertTo-Json -InputObject @($items) -Depth 6 -Compress))
    try {
    $queueDir=Join-Path $common '.copilot-kb'; Assert-PlainPath $queueDir; [void][IO.Directory]::CreateDirectory($queueDir)
    $lockPath=Join-Path $queueDir 'work-queue.lock'; Assert-PlainPath $lockPath
    try { $lock=[IO.File]::Open($lockPath,[IO.FileMode]::OpenOrCreate,[IO.FileAccess]::ReadWrite,[IO.FileShare]::None) } catch { throw 'ownership-lock-busy' }
    $stateDir=Join-Path $queueDir 'recovery/notices'; Assert-PlainPath $stateDir; [void][IO.Directory]::CreateDirectory($stateDir)
    $statePath=Join-Path $stateDir ($q.run_id+'.json'); Assert-PlainPath $statePath
    $state=@{schema_version=1;objective=$q.objective;repository_id=$Survey.repository_id;offers=@()}
    if (Test-Path -LiteralPath $statePath) {
      $state=[IO.File]::ReadAllText($statePath) | ConvertFrom-Json
      if ($state.objective -ne $q.objective -or $state.repository_id -ne $Survey.repository_id) { throw 'notice-scope-mismatch' }
    }
    $seen=@($state.offers | Where-Object { $_.id -eq $offerId }).Count -gt 0
    $output.notice=@{emit=(-not $seen -and $items.Count -gt 0);offer_id=$offerId;items=$items;message='I found unfinished work in the listed scope. I can review it, discard confirmed junk, or deliver useful work under current policy. I will continue this task separately.'}
    if (-not $seen -and $items.Count -gt 0) {
      if (@($state.offers).Count -ge 256) { throw 'notice-history-limit-review-needed' }
      $state.offers=@($state.offers)+@(@{id=$offerId;items=$items})
      if ((Get-FileDigest $Request) -ne $requestHash) { throw 'continuation-authority-changed' }
      Write-RecoveryReceipt $statePath $state
    }
    $reply=$q.reply
    if ($reply -and $reply.response -in @('revoked','pause','paused')) {
      $output.cleanup.status='withdrawn-or-paused'
    } elseif ($reply -and $reply.response -notin @('unanswered','decline','review','cleanup','accept')) {
      $output.cleanup.status='review-needed-no-acceptance'
    } elseif ($reply -and $reply.response -in @('review','cleanup','accept')) {
      $prior=@($state.offers | Where-Object { $_.id -eq $reply.offer_id })
      if ($reply.source -ne 'current-user-reply' -or $prior.Count -ne 1) { $output.cleanup.status='unbound-reply-preserved' }
      else {
        $replyIds=@($reply.item_ids | Where-Object { $_ -is [string] -and $_ })
        $mode='review'; if ($reply.response -eq 'cleanup' -and $replyIds.Count -gt 0) { $mode='cleanup' }
        $selected=@($prior[0].items)
        if ($replyIds.Count -gt 0) { $selected=@($selected | Where-Object { $replyIds -contains $_.id }) }
        $accepted=@(); $changed=@()
        foreach ($item in $selected) {
          $current=@($items | Where-Object { $_.id -eq $item.id })
          if ($current.Count -eq 1) { $accepted+=$current[0] } else { $changed+=$item.id }
        }
        $output.cleanup=@{status='accepted-scope';owner='kb-rehab';mode=$mode;offer_id=$reply.offer_id;items=$accepted;changed_item_ids=$changed;completed=$false;merge_authorized=$false;authority_source='current-user-reply'}
        if ($accepted.Count -eq 0) { $output.cleanup.status='changed-scope-preserved' }
      }
    }
    } catch {
      $output.notice=@{emit=$false;status='advisory-unavailable';reason=$_.Exception.Message}
      $output.cleanup=@{status='not-authorized';completed=$false;merge_authorized=$false;items=@()}
    } finally { if ($lock) { $lock.Dispose(); $lock=$null } }
    if ((Get-FileDigest $Request) -ne $requestHash) { throw 'continuation-authority-changed' }
    if ((Get-FileDigest $q.preparation_request) -ne $preparationHash) { throw 'preparation-request-changed' }
    $prepared=(& $entrypoint -Action prepare -Root $rootPath -Request $q.preparation_request -Json | Out-String) | ConvertFrom-Json
    if ((Get-FileDigest $Request) -ne $requestHash) { throw 'continuation-authority-changed' }
    $output.preparation=$prepared; $output.reason=$prepared.reason
    if ($prepared.status -eq 'prepared') {
      $manifest=Get-ContainedPath ([string]$prepared.destination) ([string]$q.manifest)
      if (-not (Test-Path -LiteralPath $manifest -PathType Leaf)) { throw 'prepared-manifest-missing' }
      $output.status='ready'; $output.next_action='kb-work'; $output.manifest=$manifest; $output.workspace=$prepared.destination
      $output.gates='revalidate-existing-manifest-gates'; $output.reason=$null
    } else { $output.status='dependency-needed'; $output.next_action='continue-other-ready-work' }
    return $output
  } catch {
    $output.reason=$_.Exception.Message
    $output.cleanup=@{status='not-authorized';completed=$false;merge_authorized=$false;items=@()}
    return $output
  }
  finally { if ($lock) { $lock.Dispose() } }
}
