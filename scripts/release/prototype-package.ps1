param(
  [string]$Version = "",
  [string]$Commit = "",
  [ValidateSet("windows", "linux")]
  [string]$TargetOS = "windows",
  [ValidateSet("amd64", "arm64")]
  [string]$TargetArch = "amd64",
  [string]$OutputRoot = "",
  [string]$Binary = "",
  [string[]]$EvidencePath = @(),
  [switch]$SkipBuild,
  [switch]$AllowDirty,
  [switch]$RunChecks,
  [switch]$RunAcceptanceSmoke
)

$ErrorActionPreference = "Stop"

function Get-ReleaseRepoRoot {
  return [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
}

function Resolve-ReleasePath {
  param([string]$Path, [string]$DefaultRelativePath)
  $repoRoot = Get-ReleaseRepoRoot
  if ([string]::IsNullOrWhiteSpace($Path)) {
    $Path = Join-Path $repoRoot (Normalize-ReleaseRelativePath $DefaultRelativePath)
  } elseif (-not [System.IO.Path]::IsPathRooted($Path)) {
    $Path = Join-Path $repoRoot (Normalize-ReleaseRelativePath $Path)
  }
  return [System.IO.Path]::GetFullPath($Path)
}

function Normalize-ReleaseRelativePath {
  param([string]$Path)
  $sep = [string][System.IO.Path]::DirectorySeparatorChar
  return $Path.Replace('\', $sep).Replace('/', $sep)
}

function Assert-ReleaseWorkspacePath {
  param([string]$Path, [string]$Purpose)
  $repoRoot = (Get-ReleaseRepoRoot).TrimEnd('\', '/')
  $fullPath = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  $prefix = $repoRoot + [System.IO.Path]::DirectorySeparatorChar
  if ($fullPath.Equals($repoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use repository root as $Purpose`: $fullPath"
  }
  if (-not $fullPath.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use path outside repository as $Purpose`: $fullPath"
  }
}

function Copy-ReleaseItem {
  param([string]$Source, [string]$Destination)
  if (!(Test-Path -LiteralPath $Source)) { return }
  New-Item -ItemType Directory -Force -Path (Split-Path $Destination) | Out-Null
  Copy-Item -LiteralPath $Source -Destination $Destination -Recurse -Force
}

function Get-ReleaseRelativePath {
  param([string]$Root, [string]$Path)
  $rootFull = [System.IO.Path]::GetFullPath($Root).TrimEnd('\', '/') + [System.IO.Path]::DirectorySeparatorChar
  $full = [System.IO.Path]::GetFullPath($Path)
  if (-not $full.StartsWith($rootFull, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Path is outside package root: $full"
  }
  return $full.Substring($rootFull.Length).Replace('\', '/')
}

function Get-ReleaseGitStatus {
  Push-Location (Get-ReleaseRepoRoot)
  try {
    return @(& git status --porcelain --untracked-files=all)
  } finally {
    Pop-Location
  }
}

function Assert-ReleaseCleanTree {
  param([switch]$AllowDirty)
  $status = @(Get-ReleaseGitStatus)
  if ($status.Count -gt 0 -and -not $AllowDirty) {
    $preview = ($status | Select-Object -First 20) -join "`n"
    throw "Release package requires a clean git tree. Commit, stash, or remove changes before packaging.`n$preview"
  }
  return $status
}

function Invoke-ReleaseCheck {
  param([string]$Name, [scriptblock]$Script)
  Write-Host "==> $Name"
  & $Script
  if ($LASTEXITCODE -ne 0) {
    throw "$Name failed"
  }
}

function Invoke-ReleaseCompressArchive {
  param(
    [string]$Path,
    [string]$DestinationPath,
    [int]$Attempts = 5
  )

  for ($i = 1; $i -le $Attempts; $i++) {
    try {
      Compress-Archive -Path $Path -DestinationPath $DestinationPath -Force -ErrorAction Stop
      return
    } catch {
      if ($i -eq $Attempts) {
        throw
      }
      Start-Sleep -Milliseconds (250 * $i)
    }
  }
}

$RepoRoot = Get-ReleaseRepoRoot
Push-Location $RepoRoot
try {
  if ([string]::IsNullOrWhiteSpace($Commit)) {
    $Commit = (& git rev-parse --short=12 HEAD).Trim()
  }
  if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = "prototype-$Commit"
  }
  if ($Version -notmatch '^[A-Za-z0-9._-]+$') {
    throw "Version may contain only letters, numbers, dot, underscore, or dash: $Version"
  }
  $gitStatus = @(Assert-ReleaseCleanTree -AllowDirty:$AllowDirty)
  $dirty = $gitStatus.Count -gt 0

  if ($RunChecks) {
    Invoke-ReleaseCheck -Name "go test" -Script { go test -p=1 ./... }
    Invoke-ReleaseCheck -Name "go vet" -Script { go vet -p=1 ./... }
    Invoke-ReleaseCheck -Name "buf lint" -Script { buf lint }
    Invoke-ReleaseCheck -Name "prototype audit" -Script { & .\scripts\security\prototype-audit.ps1 -Profile Fast }
  }
  if ($RunAcceptanceSmoke) {
    Invoke-ReleaseCheck -Name "prototype acceptance smoke" -Script { & .\tests\e2e\prototype_acceptance.ps1 -Profile Smoke }
  }

  $OutputRoot = Resolve-ReleasePath -Path $OutputRoot -DefaultRelativePath "dist\prototype"
  Assert-ReleaseWorkspacePath -Path $OutputRoot -Purpose "release output root"
  $packageName = "aetra-$Version-$TargetOS-$TargetArch"
  $packageRoot = Join-Path $OutputRoot $Version
  $packageDir = Join-Path $packageRoot $packageName
  Assert-ReleaseWorkspacePath -Path $packageRoot -Purpose "release package root"
  Assert-ReleaseWorkspacePath -Path $packageDir -Purpose "release package directory"

  if (Test-Path -LiteralPath $packageDir) {
    Remove-Item -LiteralPath $packageDir -Recurse -Force
  }
  New-Item -ItemType Directory -Force -Path (Join-Path $packageDir "bin") | Out-Null
  New-Item -ItemType Directory -Force -Path (Join-Path $packageDir "docs") | Out-Null
  New-Item -ItemType Directory -Force -Path (Join-Path $packageDir "evidence") | Out-Null

  $binName = if ($TargetOS -eq "windows") { "aetrad.exe" } else { "aetrad" }
  $packageBinary = Join-Path (Join-Path $packageDir "bin") $binName

  if ($SkipBuild) {
    $Binary = Resolve-ReleasePath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
    if (!(Test-Path -LiteralPath $Binary)) { throw "Binary not found: $Binary" }
    Copy-Item -LiteralPath $Binary -Destination $packageBinary -Force
  } else {
    & (Join-Path $RepoRoot "scripts\build-aetrad.ps1") `
      -Version $Version `
      -Commit $Commit `
      -TargetOS $TargetOS `
      -TargetArch $TargetArch `
      -Binary $packageBinary `
      -Force
  }

  foreach ($doc in @(
      "README.md",
      "docs\operator-commands.md",
      "docs\operator-troubleshooting.md",
      "docs\validator-onboarding.md",
      "docs\public-testnet-production-gates.md",
      "docs\public-testnet-e2e-smoke-commands.md",
      "docs\state-export-import.md",
      "docs\upgrade-playbook.md",
      "docs\prototype-acceptance-suite.md",
      "docs\security\prototype-audit-gate.md",
      "docs\query-surface.md",
      "docs\observability.md",
      "docs\release\prototype-package.md",
      "docs\release\prototype-limitations.md"
    )) {
    $source = Resolve-ReleasePath -Path $doc -DefaultRelativePath $doc
    $destination = Join-Path $packageDir (Normalize-ReleaseRelativePath $doc)
    Copy-ReleaseItem -Source $source -Destination $destination
  }

  foreach ($path in $EvidencePath) {
    $resolved = Resolve-ReleasePath -Path $path -DefaultRelativePath $path
    if (!(Test-Path -LiteralPath $resolved)) { throw "Evidence path not found: $resolved" }
    $leaf = Split-Path $resolved -Leaf
    Copy-ReleaseItem -Source $resolved -Destination (Join-Path $packageDir "evidence\$leaf")
  }

  $quickstart = @"
# Aetra Prototype Quickstart

This is a prerelease prototype artifact, not mainnet validator software.

## Verify

````powershell
Get-FileHash .\bin\$binName -Algorithm SHA256
Get-Content .\SHA256SUMS.txt
````

## Run A Local Node

Use the repository operator guide for full localnet setup and command examples:

- `README.md`
- `docs/operator-commands.md`
- `docs/operator-troubleshooting.md`
- `docs/validator-onboarding.md`
- `docs/public-testnet-production-gates.md`
- `docs/state-export-import.md`
- `docs/upgrade-playbook.md`
- `docs/observability.md`
- `docs/security/prototype-audit-gate.md`
- `docs/release/prototype-limitations.md`

Prototype tx fees use `naet`, for example `--fees 1000000naet`.
"@
  $quickstart | Set-Content -LiteralPath (Join-Path $packageDir "QUICKSTART.md")

  $manifest = [ordered]@{
    name            = "Aetra prototype release package"
    version         = $Version
    commit          = $Commit
    dirty           = $dirty
    target_os       = $TargetOS
    target_arch     = $TargetArch
    binary          = "bin/$binName"
    created_utc     = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    status          = "prototype prerelease; not mainnet-ready"
    required_checks = @(
      "go test -p=1 ./...",
      "go vet -p=1 ./...",
      "buf lint",
      "scripts/security/prototype-audit.ps1 -Profile Fast or stronger",
      "tests/e2e/prototype_acceptance.ps1 -Profile Smoke"
    )
    checks_executed = [ordered]@{
      run_checks           = [bool]$RunChecks
      run_acceptance_smoke = [bool]$RunAcceptanceSmoke
    }
    evidence        = @($EvidencePath | ForEach-Object { Split-Path $_ -Leaf })
    excluded        = @(".work", ".localnet", "keyrings", "mnemonics", "validator private keys", "node keys", ".env files", "diagnostic bundles")
  }
  $manifest | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath (Join-Path $packageDir "release-manifest.json")

  $notes = @"
# Aetra $Version Prototype Release Notes

- commit: `$Commit`
- dirty tree at package time: `$dirty`
- target: `$TargetOS/$TargetArch`
- status: prototype prerelease; not mainnet-ready

## Test Evidence

Required before publishing:

- `go test -p=1 ./...`
- `go vet -p=1 ./...`
- `buf lint`
- `scripts\security\prototype-audit.ps1 -Profile Fast` or stronger
- `tests\e2e\prototype_acceptance.ps1 -Profile Smoke`

Packaged evidence files are copied under `evidence/` when passed through `-EvidencePath`.

## Known Limitations

- See `docs/release/prototype-limitations.md` for the versioned non-goals, accepted limitations, and blockers.
- Prototype only; no mainnet launch or validator onboarding guarantees.
- IBC/external bridge, production governance economics, exchange-grade DEX behavior, public faucet, full external audit, and explorer/API SLA are non-goals.
- Localnet uses local-only test keyrings under ignored directories.
- Query list endpoints have prototype caps; pagination/load work remains before public high-cardinality use.
- Vote extension behavior is prototype-only and must be replaced or disabled before a public validator network.
- Current dependency/security triage must match `docs/security/prototype-audit-gate.md` for each release run.
- Untriaged Critical/High security findings are release blockers, not accepted limitations.
"@
  $notes | Set-Content -LiteralPath (Join-Path $packageDir "RELEASE-NOTES.md")

  $checksumPath = Join-Path $packageDir "SHA256SUMS.txt"
  Get-ChildItem -LiteralPath $packageDir -File -Recurse |
    Where-Object { $_.FullName -ne $checksumPath } |
    Sort-Object FullName |
    ForEach-Object {
      $relative = Get-ReleaseRelativePath -Root $packageDir -Path $_.FullName
      "$((Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256).Hash.ToLowerInvariant())  $relative"
    } | Set-Content -LiteralPath $checksumPath

  $archive = Join-Path $packageRoot "$packageName.zip"
  if (Test-Path -LiteralPath $archive) { Remove-Item -LiteralPath $archive -Force }
  Invoke-ReleaseCompressArchive -Path (Join-Path $packageDir "*") -DestinationPath $archive
  $archiveHash = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
  "$archiveHash  $(Split-Path $archive -Leaf)" | Set-Content -LiteralPath "$archive.sha256"

  Write-Host "Package directory: $packageDir"
  Write-Host "Package archive: $archive"
  Write-Host "Archive checksum: $archive.sha256"
} finally {
  Pop-Location
}
