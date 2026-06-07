param(
  [string]$Binary = "",
  [string]$Version = "",
  [string]$Commit = "",
  [ValidateSet("windows", "linux")]
  [string]$TargetOS = "windows",
  [ValidateSet("amd64", "arm64")]
  [string]$TargetArch = "amd64",
  [string]$GoBinary = "",
  [double]$MinFreeGB = 4,
  [switch]$Force,
  [switch]$Clean,
  [switch]$SkipModVerify,
  [switch]$SkipVersionRun
)

$ErrorActionPreference = "Stop"

function Get-BuildRepoRoot {
  return [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
}

function Normalize-BuildRelativePath {
  param([string]$Path)
  $sep = [string][System.IO.Path]::DirectorySeparatorChar
  return $Path.Replace('\', $sep).Replace('/', $sep)
}

function Resolve-BuildPath {
  param([string]$Path, [string]$DefaultRelativePath)
  $repoRoot = Get-BuildRepoRoot
  if ([string]::IsNullOrWhiteSpace($Path)) {
    $Path = Join-Path $repoRoot (Normalize-BuildRelativePath $DefaultRelativePath)
  } elseif (-not [System.IO.Path]::IsPathRooted($Path)) {
    $Path = Join-Path $repoRoot (Normalize-BuildRelativePath $Path)
  }
  return [System.IO.Path]::GetFullPath($Path)
}

function Assert-BuildWorkspacePath {
  param([string]$Path, [string]$Purpose)
  $repoRoot = (Get-BuildRepoRoot).TrimEnd('\', '/')
  $fullPath = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  $prefix = $repoRoot + [System.IO.Path]::DirectorySeparatorChar
  if ($fullPath.Equals($repoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use repository root as $Purpose`: $fullPath"
  }
  if (-not $fullPath.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use path outside repository as $Purpose`: $fullPath"
  }
}

function Get-DefaultBinaryRelativePath {
  param([string]$OS, [string]$Arch)
  if ($OS -eq "windows" -and $Arch -eq "amd64") {
    return "build\aetrad.exe"
  }
  $suffix = if ($OS -eq "windows") { ".exe" } else { "" }
  return "build\aetrad-$OS-$Arch$suffix"
}

function Get-GoTool {
  param([string]$Requested)
  if (-not [string]::IsNullOrWhiteSpace($Requested)) {
    if ($Requested -notmatch '[\\/]') {
      return $Requested
    }
    return Resolve-BuildPath -Path $Requested -DefaultRelativePath $Requested
  }
  $repoRoot = Get-BuildRepoRoot
  $localGo = Join-Path $repoRoot ".work\tools\go1.25.11\go\bin\go.exe"
  if (Test-Path -LiteralPath $localGo) {
    $env:PATH = "$(Split-Path $localGo);$env:PATH"
    return $localGo
  }
  return "go"
}

function Assert-GoVersionMatchesMod {
  param([string]$Go, [string]$GoVersionOutput)
  $goMod = Get-Content -Raw -LiteralPath (Join-Path (Get-BuildRepoRoot) "go.mod")
  if ($goMod -notmatch '(?m)^go\s+([0-9]+)\.([0-9]+)') {
    throw "Could not read Go version from go.mod"
  }
  $required = "$($Matches[1]).$($Matches[2])"
  if ($GoVersionOutput -notmatch "go$required\.") {
    throw "Go toolchain mismatch: go.mod requires $required.x, got: $GoVersionOutput"
  }
}

function Assert-FreeSpace {
  param([string]$Path, [double]$RequiredGB)
  if ($RequiredGB -le 0) { return }
  $root = [System.IO.Path]::GetPathRoot([System.IO.Path]::GetFullPath($Path))
  $drive = [System.IO.DriveInfo]::new($root)
  $freeGB = [math]::Round($drive.AvailableFreeSpace / 1GB, 2)
  if ($freeGB -lt $RequiredGB) {
    throw "Not enough free space on $root for build: ${freeGB}GB available, ${RequiredGB}GB required"
  }
}

function Get-GitValue {
  param([string[]]$CommandArgs, [string]$Fallback)
  try {
    $value = (& git @CommandArgs 2>$null).Trim()
    if (-not [string]::IsNullOrWhiteSpace($value)) { return $value }
  } catch {}
  return $Fallback
}

function Get-CurrentGOOS {
  if ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows)) {
    return "windows"
  }
  if ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Linux)) {
    return "linux"
  }
  return "unknown"
}

function Invoke-GoModVerify {
  param([string]$Go)
  $oldErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $verifyOutput = & $Go mod verify 2>&1
    $verifyCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $oldErrorActionPreference
  }
  if ($verifyCode -eq 0) {
    $verifyOutput | ForEach-Object { Write-Host $_ }
    return
  }

  $verifyText = ($verifyOutput | Out-String)
  $knownCometWindowsMismatch = (Get-CurrentGOOS) -eq "windows" -and
    $verifyText -match 'github\.com/cometbft/cometbft v0\.39\.3: dir has been modified'
  if (-not $knownCometWindowsMismatch) {
    $verifyOutput | ForEach-Object { Write-Error $_ }
    throw "go mod verify failed"
  }

  $moduleDir = Join-Path $env:GOMODCACHE "github.com\cometbft\cometbft@v0.39.3"
  Assert-BuildWorkspacePath -Path $moduleDir -Purpose "CometBFT module cache"
  if (Test-Path -LiteralPath $moduleDir) {
    Write-Warning "go mod verify hit the known Windows extracted-tree mismatch for CometBFT v0.39.3; verifying cached module zip after removing only $moduleDir"
    Remove-Item -LiteralPath $moduleDir -Recurse -Force
  }

  $ErrorActionPreference = "Continue"
  try {
    $retryOutput = & $Go mod verify 2>&1
    $retryCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $oldErrorActionPreference
  }
  if ($retryCode -ne 0) {
    $retryOutput | ForEach-Object { Write-Error $_ }
    throw "go mod verify failed after CometBFT Windows cache fallback"
  }
  $retryOutput | ForEach-Object { Write-Host $_ }
}

$RepoRoot = Get-BuildRepoRoot
$defaultBinary = Get-DefaultBinaryRelativePath -OS $TargetOS -Arch $TargetArch
$Binary = Resolve-BuildPath -Path $Binary -DefaultRelativePath $defaultBinary
$binaryIsDefault = $Binary.Equals((Resolve-BuildPath -Path "" -DefaultRelativePath $defaultBinary), [System.StringComparison]::OrdinalIgnoreCase)
Assert-BuildWorkspacePath -Path $Binary -Purpose "build binary"
Assert-FreeSpace -Path $Binary -RequiredGB $MinFreeGB

$Go = Get-GoTool -Requested $GoBinary
$GoVersion = (& $Go version).Trim()
Assert-GoVersionMatchesMod -Go $Go -GoVersionOutput $GoVersion

if ([string]::IsNullOrWhiteSpace($Commit)) {
  $Commit = Get-GitValue -CommandArgs @("rev-parse", "--short=12", "HEAD") -Fallback "unknown"
}
if ([string]::IsNullOrWhiteSpace($Version)) {
  $Version = if ($Commit -eq "unknown") { "dev" } else { "dev-$Commit" }
}
if ($Version -notmatch '^[A-Za-z0-9._-]+$') {
  throw "Version may contain only letters, numbers, dot, underscore, or dash: $Version"
}

$dirty = if ((git status --porcelain --untracked-files=no) -eq $null) { "false" } else { "true" }
$buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$goCache = Join-Path $RepoRoot ".work\gocache"
$goTmp = Join-Path $RepoRoot ".work\gotmp"
New-Item -ItemType Directory -Force -Path $goCache, $goTmp, (Split-Path $Binary) | Out-Null
Assert-BuildWorkspacePath -Path $goCache -Purpose "Go build cache"
Assert-BuildWorkspacePath -Path $goTmp -Purpose "Go temp directory"

if ((Test-Path -LiteralPath $Binary) -and -not $binaryIsDefault -and -not $Force -and -not $Clean) {
  throw "Refusing to overwrite custom binary without -Force: $Binary"
}
if ($Clean -and (Test-Path -LiteralPath $Binary)) {
  Remove-Item -LiteralPath $Binary -Force
}

Push-Location $RepoRoot
$oldGoCache = $env:GOCACHE
$oldGoTmp = $env:GOTMPDIR
$oldGoModCache = $env:GOMODCACHE
$oldGoOS = $env:GOOS
$oldGoArch = $env:GOARCH
$oldCgoEnabled = $env:CGO_ENABLED
try {
  $env:GOCACHE = $goCache
  $env:GOTMPDIR = $goTmp
  $env:GOMODCACHE = Join-Path $RepoRoot ".work\gomodcache"
  $env:GOOS = $TargetOS
  $env:GOARCH = $TargetArch
  if ($TargetOS -ne "windows") {
    $env:CGO_ENABLED = "0"
  }

  New-Item -ItemType Directory -Force -Path $env:GOMODCACHE | Out-Null
  Assert-BuildWorkspacePath -Path $env:GOMODCACHE -Purpose "Go module cache"

  Write-Host "Go: $GoVersion"
  Write-Host "Go cache: $goCache"
  Write-Host "Go temp: $goTmp"
  Write-Host "Go module cache: $env:GOMODCACHE"
  if (-not $SkipModVerify) {
    & $Go mod download
    if ($LASTEXITCODE -ne 0) { throw "go mod download failed" }
    Invoke-GoModVerify -Go $Go
  }

  $ldflags = "-X github.com/sovereign-l1/l1/cmd/l1d/cmd.appVersion=$Version -X github.com/sovereign-l1/l1/cmd/l1d/cmd.gitCommit=$Commit -X github.com/sovereign-l1/l1/cmd/l1d/cmd.buildDate=$buildDate -X github.com/sovereign-l1/l1/cmd/l1d/cmd.dirty=$dirty"
  $buildArgs = @("build", "-mod=readonly", "-trimpath", "-p=1", "-ldflags", $ldflags, "-o", $Binary, "./cmd/l1d")
  & $Go @buildArgs
  if ($LASTEXITCODE -ne 0) { throw "go build failed for $TargetOS/$TargetArch" }

  Write-Host "Binary: $Binary"
  Write-Host "Version: $Version"
  Write-Host "Commit: $Commit"
  Write-Host "Dirty: $dirty"
  Write-Host "Build date: $buildDate"

  $canRun = (Get-CurrentGOOS) -eq $TargetOS
  if (-not $SkipVersionRun -and $canRun) {
    & $Binary version --long --output json
    if ($LASTEXITCODE -ne 0) { throw "built binary version check failed" }
  } elseif (-not $canRun) {
    Write-Host "Skipping version run for cross-built $TargetOS/$TargetArch binary"
  }
} finally {
  $env:GOCACHE = $oldGoCache
  $env:GOTMPDIR = $oldGoTmp
  $env:GOMODCACHE = $oldGoModCache
  $env:GOOS = $oldGoOS
  $env:GOARCH = $oldGoArch
  $env:CGO_ENABLED = $oldCgoEnabled
  Pop-Location
}
