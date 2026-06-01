param(
  [string]$Root = ".",
  [int]$MaxFiles = 1200,
  [switch]$Json
)

$ErrorActionPreference = "Stop"

function Resolve-RepoRoot {
  param([string]$Path)
  $resolved = (Resolve-Path $Path).Path
  try {
    $gitRoot = (& git -C $resolved rev-parse --show-toplevel 2>$null)
    if ($LASTEXITCODE -eq 0 -and $gitRoot) {
      return (Resolve-Path $gitRoot.Trim()).Path
    }
  } catch {
    # Fall through to the resolved path when git is unavailable.
  }
  return $resolved
}

function Test-IgnoredPath {
  param([string]$Relative)
  $normalized = $Relative -replace "\\", "/"
  return $normalized -match '(^|/)(\.git|node_modules|vendor|dist|build|target|bin|obj|\.gradle|\.idea|\.venv|venv|__pycache__|coverage|out)(/|$)'
}

function Add-Item {
  param(
    [System.Collections.Generic.List[object]]$List,
    [object]$Value
  )
  $List.Add($Value) | Out-Null
}

function Get-AvailableLanguageServers {
  $candidates = @(
    @{ language = "go"; command = "gopls" },
    @{ language = "python"; command = "pyright-langserver" },
    @{ language = "typescript"; command = "typescript-language-server" },
    @{ language = "kotlin"; command = "kotlin-language-server" },
    @{ language = "java"; command = "jdtls" },
    @{ language = "csharp"; command = "OmniSharp" },
    @{ language = "rust"; command = "rust-analyzer" }
  )
  foreach ($candidate in $candidates) {
    $cmd = Get-Command $candidate.command -ErrorAction SilentlyContinue
    [pscustomobject]@{
      language = $candidate.language
      command = $candidate.command
      available = [bool]$cmd
      path = if ($cmd) { $cmd.Source } else { "" }
    }
  }
}

function Get-SymbolsForFile {
  param(
    [string]$FullName,
    [string]$Relative,
    [string]$Extension
  )
  $symbols = [System.Collections.Generic.List[object]]::new()
  $lineNo = 0
  Get-Content -LiteralPath $FullName -ErrorAction SilentlyContinue | ForEach-Object {
    $lineNo++
    $line = $_
    $matches = @()
    switch ($Extension.ToLowerInvariant()) {
      ".py" {
        if ($line -match '^\s*(class|def)\s+([A-Za-z_][A-Za-z0-9_]*)') {
          $matches += @{ kind = $Matches[1]; name = $Matches[2] }
        }
      }
      ".go" {
        if ($line -match '^\s*func\s+(?:\([^)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)\s*\(') {
          $matches += @{ kind = "func"; name = $Matches[1] }
        } elseif ($line -match '^\s*type\s+([A-Za-z_][A-Za-z0-9_]*)\s+(struct|interface)') {
          $matches += @{ kind = $Matches[2]; name = $Matches[1] }
        }
      }
      { $_ -in @(".ts", ".tsx", ".js", ".jsx") } {
        if ($line -match '^\s*(?:export\s+)?(?:async\s+)?function\s+([A-Za-z_$][A-Za-z0-9_$]*)') {
          $matches += @{ kind = "function"; name = $Matches[1] }
        } elseif ($line -match '^\s*(?:export\s+)?class\s+([A-Za-z_$][A-Za-z0-9_$]*)') {
          $matches += @{ kind = "class"; name = $Matches[1] }
        } elseif ($line -match '^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*(?:async\s*)?(?:\([^)]*\)|[A-Za-z_$][A-Za-z0-9_$]*)\s*=>') {
          $matches += @{ kind = "function"; name = $Matches[1] }
        }
      }
      { $_ -in @(".kt", ".kts") } {
        if ($line -match '^\s*(class|object|interface|fun)\s+([A-Za-z_][A-Za-z0-9_]*)') {
          $matches += @{ kind = $Matches[1]; name = $Matches[2] }
        }
      }
      { $_ -in @(".java", ".cs", ".swift", ".rs", ".rb", ".php") } {
        if ($line -match '^\s*(?:public|private|protected|internal|static|final|open|export|\s)*\s*(class|interface|struct|enum|func|function|def|trait|impl)\s+([A-Za-z_][A-Za-z0-9_]*)') {
          $matches += @{ kind = $Matches[1]; name = $Matches[2] }
        }
      }
    }
    foreach ($match in $matches) {
      Add-Item $symbols ([pscustomobject]@{
        name = $match.name
        kind = $match.kind
        file = $Relative
        line = $lineNo
      })
    }
  }
  return $symbols
}

function Get-EntryPointHints {
  param([string]$RepoRoot, [object[]]$Files)
  $hints = [System.Collections.Generic.List[object]]::new()
  $relativeSet = @{}
  foreach ($file in $Files) {
    $relativeSet[$file.relative] = $true
  }

  $knownFiles = @(
    "package.json", "pyproject.toml", "requirements.txt", "go.mod", "Cargo.toml",
    "pom.xml", "build.gradle", "build.gradle.kts", "settings.gradle.kts",
    "Program.cs", "main.go", "manage.py", "vite.config.ts", "next.config.js"
  )
  foreach ($known in $knownFiles) {
    if ($relativeSet.ContainsKey($known)) {
      Add-Item $hints ([pscustomobject]@{ kind = "manifest-or-entry"; file = $known; reason = "known project entry/config file" })
    }
  }

  foreach ($file in $Files) {
    $path = $file.relative -replace "\\", "/"
    if ($path -match '(^|/)cmd/[^/]+/main\.go$') {
      Add-Item $hints ([pscustomobject]@{ kind = "cli-entry"; file = $file.relative; reason = "Go cmd main" })
    } elseif ($path -match '(^|/)(pages|app|routes)/.+\.(tsx|ts|jsx|js|vue|svelte)$') {
      Add-Item $hints ([pscustomobject]@{ kind = "route-or-screen"; file = $file.relative; reason = "frontend route/screen convention" })
    } elseif ($path -match 'MainActivity\.kt$') {
      Add-Item $hints ([pscustomobject]@{ kind = "android-entry"; file = $file.relative; reason = "Android activity" })
    } elseif ($path -match '^\.github/workflows/.+\.ya?ml$') {
      Add-Item $hints ([pscustomobject]@{ kind = "ci-workflow"; file = $file.relative; reason = "GitHub Actions workflow" })
    }
  }
  return $hints
}

$repoRoot = Resolve-RepoRoot $Root
$extensions = @(".go", ".py", ".ts", ".tsx", ".js", ".jsx", ".kt", ".kts", ".java", ".cs", ".rs", ".rb", ".php", ".swift")
$files = [System.Collections.Generic.List[object]]::new()

Get-ChildItem -LiteralPath $repoRoot -Recurse -File -Force -ErrorAction SilentlyContinue |
  Sort-Object FullName |
  ForEach-Object {
    if ($files.Count -ge $MaxFiles) {
      return
    }
    $relative = $_.FullName.Substring($repoRoot.Length + 1)
    if (Test-IgnoredPath $relative) {
      return
    }
    if ($extensions -contains $_.Extension.ToLowerInvariant() -or $_.Name -in @("package.json", "pyproject.toml", "go.mod", "Cargo.toml", "pom.xml", "build.gradle", "build.gradle.kts", "settings.gradle.kts", "requirements.txt")) {
      Add-Item $files ([pscustomobject]@{
        relative = $relative
        extension = $_.Extension.ToLowerInvariant()
        bytes = $_.Length
      })
    }
  }

$symbols = [System.Collections.Generic.List[object]]::new()
foreach ($file in $files | Where-Object { $extensions -contains $_.extension }) {
  $fullPath = Join-Path $repoRoot $file.relative
  foreach ($symbol in (Get-SymbolsForFile -FullName $fullPath -Relative $file.relative -Extension $file.extension)) {
    Add-Item $symbols $symbol
  }
}

$byExtension = $files |
  Group-Object extension |
  Sort-Object Name |
  ForEach-Object {
    [pscustomobject]@{
      extension = if ($_.Name) { $_.Name } else { "[none]" }
      files = $_.Count
      bytes = ($_.Group | Measure-Object bytes -Sum).Sum
    }
  }

$result = [pscustomobject]@{
  ok = $true
  generated_at = (Get-Date).ToString("o")
  root = $repoRoot
  note = "Static code-intel inventory. Language-server availability is reported, but LSP calls are not required for this helper."
  language_servers = @(Get-AvailableLanguageServers)
  file_count = $files.Count
  files_by_extension = @($byExtension)
  entry_points = @(Get-EntryPointHints -RepoRoot $repoRoot -Files $files | Select-Object -First 80)
  symbols = @($symbols | Select-Object -First 250)
  largest_files = @($files | Sort-Object bytes -Descending | Select-Object -First 25)
}

if ($Json) {
  $result | ConvertTo-Json -Depth 8
} else {
  Write-Host "Code intel: $($result.file_count) files, $($result.symbols.Count) symbols sampled"
  Write-Host ""
  Write-Host "Language servers:"
  foreach ($server in $result.language_servers) {
    $status = if ($server.available) { "available" } else { "missing" }
    Write-Host "- $($server.language): $($server.command) ($status)"
  }
  Write-Host ""
  Write-Host "Files by extension:"
  foreach ($row in $result.files_by_extension) {
    Write-Host "- $($row.extension): $($row.files)"
  }
  Write-Host ""
  Write-Host "Entry point hints:"
  foreach ($entry in $result.entry_points) {
    Write-Host "- [$($entry.kind)] $($entry.file) - $($entry.reason)"
  }
  Write-Host ""
  Write-Host "Top symbols:"
  foreach ($symbol in ($result.symbols | Select-Object -First 60)) {
    Write-Host "- $($symbol.name) ($($symbol.kind)) $($symbol.file):$($symbol.line)"
  }
}
