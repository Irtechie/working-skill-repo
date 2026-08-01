[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [ValidateSet("list", "claim", "update", "release")]
    [string]$Action,

    [string]$WorkId,
    [string]$SessionId,
    [string]$Branch,
    [string]$Summary,
    [string]$Scope,
    [Alias("Resource")]
    [string[]]$SemanticResources = @(),
    [string]$ResumeCondition,

    [ValidateSet("queued", "in_progress", "active", "paused", "awaiting-review", "blocked", "quarantined", "local-durable", "delivery-integrated", "superseded", "done", "retired")]
    [string]$Status = "in_progress",

    [ValidateRange(1, 20)]
    [int]$RepoWipLimit = 3,

    [ValidateRange(5, 10080)]
    [int]$StaleMinutes = 60
)

$ErrorActionPreference = "Stop"

function Get-RepositoryPaths {
    $root = (& git rev-parse --show-toplevel 2>$null).Trim()
    if (-not $root) {
        throw "not_in_git_repository"
    }
    $common = (& git rev-parse --git-common-dir 2>$null).Trim()
    if (-not [System.IO.Path]::IsPathRooted($common)) {
        $common = Join-Path $root $common
    }
    $queueDir = Join-Path ([System.IO.Path]::GetFullPath($common)) ".copilot-kb"
    [pscustomobject]@{
        Root = [System.IO.Path]::GetFullPath($root)
        QueueDir = $queueDir
        QueuePath = Join-Path $queueDir "work-queue.json"
        LockPath = Join-Path $queueDir "work-queue.lock"
    }
}

function Read-Queue([string]$Path) {
    if (-not (Test-Path $Path)) {
        return @()
    }
    $raw = Get-Content $Path -Raw
    if (-not $raw.Trim()) {
        return @()
    }
    $value = $raw | ConvertFrom-Json
    return @($value)
}

function Write-Queue([string]$Path, [object[]]$Queue) {
    $temp = "$Path.$PID.tmp"
    @($Queue) | ConvertTo-Json -Depth 6 | Set-Content -Path $temp -Encoding utf8
    Move-Item -Force $temp $Path
}

function Invoke-WithQueueLock([scriptblock]$Body) {
    $paths = Get-RepositoryPaths
    New-Item -ItemType Directory -Force -Path $paths.QueueDir | Out-Null
    $lock = $null
    for ($attempt = 0; $attempt -lt 50 -and -not $lock; $attempt++) {
        try {
            $lock = [System.IO.File]::Open(
                $paths.LockPath,
                [System.IO.FileMode]::OpenOrCreate,
                [System.IO.FileAccess]::ReadWrite,
                [System.IO.FileShare]::None
            )
        }
        catch [System.IO.IOException] {
            Start-Sleep -Milliseconds 100
        }
    }
    if (-not $lock) {
        throw "work_queue_lock_timeout"
    }
    try {
        & $Body $paths
    }
    finally {
        $lock.Dispose()
    }
}

function Normalize-SemanticResources([string[]]$Values) {
    $normalized = foreach ($value in @($Values)) {
        $candidate = ([string]$value).Trim().ToLowerInvariant()
        if ($candidate -notmatch "^(code|publisher|release-manifest|deploy):[^:\s][^\r\n]*$") {
            throw "semantic_resource_must_be_supported_kind_and_value"
        }
        $parts = $candidate.Split(":", 2)
        $escaped = [Uri]::EscapeDataString($parts[1].Trim())
        if (-not $escaped -or $escaped.Length -gt 768) {
            throw "semantic_resource_value_invalid"
        }
        "$($parts[0]):$escaped"
    }
    return @($normalized | Sort-Object -Unique)
}

function Set-QueueProperty([object]$Item, [string]$Name, [object]$Value) {
    if ($Item.PSObject.Properties.Name -contains $Name) {
        $Item.$Name = $Value
    }
    else {
        $Item | Add-Member -NotePropertyName $Name -NotePropertyValue $Value
    }
}

function Require-Identity {
    if ($WorkId -notmatch "^[a-z0-9][a-z0-9-]{1,63}$") {
        throw "work_id_must_be_kebab_case"
    }
    if (-not $SessionId) {
        throw "session_id_required"
    }
}

Invoke-WithQueueLock {
    param($paths)

    $queue = @(Read-Queue $paths.QueuePath)
    $now = [DateTimeOffset]::UtcNow
    $activeStatuses = @("queued", "in_progress", "active")

    if ($Action -eq "list") {
        $result = foreach ($item in $queue) {
            $updated = [DateTimeOffset]::Parse($item.updated_at)
            [pscustomobject]@{
                work_id = $item.work_id
                status = $item.status
                session_id = $item.session_id
                branch = $item.branch
                worktree = $item.worktree
                summary = $item.summary
                scope = $item.scope
                semantic_resources = @($item.semantic_resources)
                active_owner = ($activeStatuses -contains $item.status)
                global_authority = $false
                resume_condition = $item.resume_condition
                started_at = $item.started_at
                updated_at = $item.updated_at
                stale = (
                    $activeStatuses -contains $item.status -and
                    ($now - $updated).TotalMinutes -ge $StaleMinutes
                )
            }
        }
        @($result) | ConvertTo-Json -Depth 5
        return
    }

    Require-Identity
    if ($Action -in @("claim", "update") -and $Status -in @("done", "superseded")) {
        throw "terminal_status_requires_release"
    }

    $existing = @($queue | Where-Object { $_.work_id -eq $WorkId })
    $owned = @($existing | Where-Object { $_.session_id -eq $SessionId }) | Select-Object -First 1
    if ($Action -in @("claim", "update") -and $owned -and $owned.status -in @("done", "superseded")) {
        throw "terminal_claim_cannot_be_reopened"
    }

    $normalizedResources = Normalize-SemanticResources $SemanticResources

    if ($Action -in @("claim", "update") -and $activeStatuses -contains $Status) {
        $resourceConflict = @(
            $queue | Where-Object {
                if ($activeStatuses -notcontains $_.status -or $_.session_id -eq $SessionId) {
                    return $false
                }
                $existingResources = @($_.semantic_resources)
                return @($normalizedResources | Where-Object { $existingResources -contains $_ }).Count -gt 0
            }
        ) | Select-Object -First 1
        if ($resourceConflict) {
            [pscustomobject]@{
                result = "resource_conflict"
                session_id = $resourceConflict.session_id
                work_id = $resourceConflict.work_id
                semantic_resources = @($normalizedResources | Where-Object { @($resourceConflict.semantic_resources) -contains $_ })
                global_authority = $false
            } | ConvertTo-Json -Depth 5 -Compress
            exit 5
        }
    }

    if ($Action -eq "claim") {
        $conflict = @(
            $existing | Where-Object {
                $activeStatuses -contains $_.status -and $_.session_id -ne $SessionId
            }
        ) | Select-Object -First 1
        if ($conflict) {
            [pscustomobject]@{
                result = "conflict"
                work_id = $conflict.work_id
                status = $conflict.status
                session_id = $conflict.session_id
                branch = $conflict.branch
                worktree = $conflict.worktree
                updated_at = $conflict.updated_at
            } | ConvertTo-Json -Compress
            exit 3
        }

        $activeOwners = @(
            $queue | Where-Object {
                $activeStatuses -contains $_.status -and $_.session_id -ne $SessionId
            } | Select-Object -ExpandProperty session_id -Unique
        )
        if ($activeOwners.Count -ge $RepoWipLimit) {
            [pscustomobject]@{
                result = "repo_wip_limit"
                limit = $RepoWipLimit
                active_owners = $activeOwners.Count
                sessions = @($queue | Where-Object {
                    $activeStatuses -contains $_.status -and $activeOwners -contains $_.session_id
                } | Select-Object work_id, session_id, branch, updated_at)
            } | ConvertTo-Json -Depth 5 -Compress
            exit 4
        }

        if (-not $owned) {
            $owned = [pscustomobject]@{
                work_id = $WorkId
                status = $Status
                session_id = $SessionId
                branch = $Branch
                worktree = $paths.Root
                summary = $Summary
                scope = $Scope
                semantic_resources = $normalizedResources
                resume_condition = $ResumeCondition
                global_authority = $false
                started_at = $now.ToString("o")
                updated_at = $now.ToString("o")
                completed_at = $null
            }
            $queue += $owned
        }
    }
    elseif (-not $owned) {
        throw "work_claim_not_owned"
    }

    if ($Action -in @("claim", "update")) {
        $owned.status = $Status
        $owned.updated_at = $now.ToString("o")
        $owned.completed_at = $null
        if ($Branch) { $owned.branch = $Branch }
        if ($Summary) { $owned.summary = $Summary }
        if ($Scope) { $owned.scope = $Scope }
        if ($PSBoundParameters.ContainsKey("SemanticResources")) { Set-QueueProperty $owned "semantic_resources" $normalizedResources }
        if ($ResumeCondition) { Set-QueueProperty $owned "resume_condition" $ResumeCondition }
    }
    elseif ($Action -eq "release") {
        if ($Status -notin @("paused", "awaiting-review", "blocked", "quarantined", "local-durable", "delivery-integrated", "done", "superseded", "retired")) {
            throw "release_status_must_be_terminal"
        }
        $owned.status = $Status
        $owned.updated_at = $now.ToString("o")
        $owned.completed_at = $now.ToString("o")
    }

    Write-Queue $paths.QueuePath $queue
    [pscustomobject]@{
        result = "ok"
        work_id = $owned.work_id
        status = $owned.status
        session_id = $owned.session_id
        branch = $owned.branch
        worktree = $owned.worktree
        semantic_resources = @($owned.semantic_resources)
        active_owner = ($activeStatuses -contains $owned.status)
        global_authority = $false
        resume_condition = $owned.resume_condition
        updated_at = $owned.updated_at
    } | ConvertTo-Json -Compress
}
