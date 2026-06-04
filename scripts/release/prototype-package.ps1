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
  [switch]$SkipBuild
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

  $OutputRoot = Resolve-ReleasePath -Path $OutputRoot -DefaultRelativePath "dist\prototype"
  Assert-ReleaseWorkspacePath -Path $OutputRoot -Purpose "release output root"
  $packageName = "orbitalis-$Version-$TargetOS-$TargetArch"
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

  $binName = if ($TargetOS -eq "windows") { "orbitalisd.exe" } else { "orbitalisd" }
  $packageBinary = Join-Path (Join-Path $packageDir "bin") $binName

  if ($SkipBuild) {
    $Binary = Resolve-ReleasePath -Path $Binary -DefaultRelativePath "build\orbitalisd.exe"
    if (!(Test-Path -LiteralPath $Binary)) { throw "Binary not found: $Binary" }
    Copy-Item -LiteralPath $Binary -Destination $packageBinary -Force
  } else {
    $go = Join-Path $RepoRoot ".work\tools\go1.25.11\go\bin\go.exe"
    if (!(Test-Path -LiteralPath $go)) { $go = "go" }
    $goCache = Join-Path $RepoRoot ".work\gocache"
    $goTmp = Join-Path $RepoRoot ".work\gotmp"
    New-Item -ItemType Directory -Force -Path $goCache, $goTmp | Out-Null
    $env:GOCACHE = $goCache
    $env:GOTMPDIR = $goTmp
    $env:GOOS = $TargetOS
    $env:GOARCH = $TargetArch
    $env:CGO_ENABLED = "0"
    $buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    $dirty = if ((git status --porcelain --untracked-files=no) -eq $null) { "false" } else { "true" }
    $ldflags = "-X github.com/sovereign-l1/l1/cmd/l1d/cmd.appVersion=$Version -X github.com/sovereign-l1/l1/cmd/l1d/cmd.gitCommit=$Commit -X github.com/sovereign-l1/l1/cmd/l1d/cmd.buildDate=$buildDate -X github.com/sovereign-l1/l1/cmd/l1d/cmd.dirty=$dirty"
    & $go build -trimpath -p=1 -ldflags $ldflags -o $packageBinary ./cmd/l1d
    if ($LASTEXITCODE -ne 0) { throw "go build failed for $TargetOS/$TargetArch" }
  }

  foreach ($doc in @(
      "README.md",
      "docs\operator-commands.md",
      "docs\prototype-acceptance-suite.md",
      "docs\security\prototype-audit-gate.md",
      "docs\query-surface.md",
      "docs\observability.md",
      "docs\release\prototype-package.md"
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

  $manifest = [ordered]@{
    name         = "Orbitalis prototype release package"
    version      = $Version
    commit       = $Commit
    target_os    = $TargetOS
    target_arch  = $TargetArch
    binary       = "bin/$binName"
    created_utc  = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    status       = "prototype prerelease; not mainnet-ready"
  }
  $manifest | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath (Join-Path $packageDir "release-manifest.json")

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
  Compress-Archive -Path (Join-Path $packageDir "*") -DestinationPath $archive -Force
  $archiveHash = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
  "$archiveHash  $(Split-Path $archive -Leaf)" | Set-Content -LiteralPath "$archive.sha256"

  Write-Host "Package directory: $packageDir"
  Write-Host "Package archive: $archive"
  Write-Host "Archive checksum: $archive.sha256"
} finally {
  Pop-Location
}
