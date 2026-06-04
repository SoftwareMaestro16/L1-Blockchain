param(
  [string]$Output = "",
  [string]$Version = "",
  [string]$Goos = "",
  [string]$Goarch = "",
  [switch]$Force,
  [switch]$SkipModVerify,
  [switch]$SkipChecksum
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$BuilderId = "scripts/build-orbitalisd.ps1"
$BinaryName = "orbitalisd"
$VersionPackage = "github.com/cosmos/cosmos-sdk/version"
$CmdPackage = "github.com/sovereign-l1/l1/cmd/l1d/cmd"

function Resolve-Go {
  $localGo = Join-Path $RepoRoot ".work\tools\go1.25.11\go\bin\go.exe"
  if (Test-Path -LiteralPath $localGo) {
    return (Resolve-Path -LiteralPath $localGo).Path
  }

  $goCommand = Get-Command go -ErrorAction SilentlyContinue
  if ($null -eq $goCommand) {
    throw "go was not found. Install Go 1.25.x or place it under .work\tools\go1.25.11\go\bin."
  }
  return $goCommand.Source
}

function Invoke-GitOrDefault {
  param(
    [string[]]$Arguments,
    [string]$Default
  )

  $output = & git @Arguments 2>$null
  if ($LASTEXITCODE -ne 0 -or $null -eq $output) {
    return $Default
  }
  return (($output -join "`n").Trim())
}

function Resolve-BuildDate {
  if (-not [string]::IsNullOrWhiteSpace($env:SOURCE_DATE_EPOCH)) {
    return [DateTimeOffset]::FromUnixTimeSeconds([int64]$env:SOURCE_DATE_EPOCH).UtcDateTime.ToString("yyyy-MM-ddTHH:mm:ssZ")
  }
  return [DateTime]::UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ")
}

function Resolve-OutputPath {
  param(
    [string]$RequestedOutput,
    [string]$TargetGoos,
    [string]$TargetGoarch,
    [bool]$ExplicitTarget
  )

  if (-not [string]::IsNullOrWhiteSpace($RequestedOutput)) {
    if ([System.IO.Path]::IsPathRooted($RequestedOutput)) {
      return $RequestedOutput
    }
    return (Join-Path $RepoRoot $RequestedOutput)
  }

  $extension = ""
  if ($TargetGoos -eq "windows") {
    $extension = ".exe"
  }

  $fileName = "$BinaryName$extension"
  if ($ExplicitTarget) {
    $fileName = "$BinaryName-$TargetGoos-$TargetGoarch$extension"
  }
  return (Join-Path $RepoRoot (Join-Path "build" $fileName))
}

Push-Location $RepoRoot
try {
  $go = Resolve-Go
  $goBin = Split-Path -Parent $go
  $toolBin = Join-Path $RepoRoot ".work\tools\bin"
  if (Test-Path -LiteralPath $toolBin) {
    $env:PATH = "$goBin;$toolBin;$env:PATH"
  } else {
    $env:PATH = "$goBin;$env:PATH"
  }

  $goVersion = (& $go version)
  if ($LASTEXITCODE -ne 0) {
    throw "go version failed"
  }
  if (($goVersion -join " ") -notmatch "go1\.25\.") {
    throw "expected Go 1.25.x, got: $($goVersion -join ' ')"
  }

  if ([string]::IsNullOrWhiteSpace($Goos)) {
    $Goos = ((& $go env GOOS) -join "").Trim()
  }
  if ([string]::IsNullOrWhiteSpace($Goarch)) {
    $Goarch = ((& $go env GOARCH) -join "").Trim()
  }
  if ([string]::IsNullOrWhiteSpace($Goos) -or [string]::IsNullOrWhiteSpace($Goarch)) {
    throw "could not resolve GOOS/GOARCH"
  }

  $explicitTarget = $PSBoundParameters.ContainsKey("Goos") -or $PSBoundParameters.ContainsKey("Goarch")
  $outputPath = Resolve-OutputPath -RequestedOutput $Output -TargetGoos $Goos -TargetGoarch $Goarch -ExplicitTarget $explicitTarget
  $outputDir = Split-Path -Parent $outputPath
  New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

  $manifestPath = "$outputPath.build.json"
  if ((Test-Path -LiteralPath $outputPath) -and -not $Force) {
    if (-not (Test-Path -LiteralPath $manifestPath)) {
      throw "refusing to overwrite existing $outputPath without -Force because no build manifest was found"
    }

    $existingManifest = Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
    if ($existingManifest.builder -ne $BuilderId) {
      throw "refusing to overwrite existing $outputPath without -Force because it was not created by $BuilderId"
    }
  }

  if (-not $SkipModVerify) {
    & $go mod verify
    if ($LASTEXITCODE -ne 0) {
      throw "go mod verify failed; use a clean module cache before release-like builds"
    }
  }

  $commit = Invoke-GitOrDefault -Arguments @("rev-parse", "HEAD") -Default "unknown"
  & git diff --quiet --ignore-submodules --
  $dirty = "false"
  if ($LASTEXITCODE -ne 0) {
    $dirty = "true"
  }

  if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = Invoke-GitOrDefault -Arguments @("describe", "--tags", "--always", "--dirty") -Default "dev"
  }

  $buildDate = Resolve-BuildDate
  $buildTags = "$Goos,$Goarch"
  $ldflags = @(
    "-X $VersionPackage.Name=Orbitalis",
    "-X $VersionPackage.AppName=$BinaryName",
    "-X $VersionPackage.Version=$Version",
    "-X $VersionPackage.Commit=$commit",
    "-X $VersionPackage.BuildTags=$buildTags",
    "-X $CmdPackage.BuildDate=$buildDate",
    "-X $CmdPackage.Dirty=$dirty"
  ) -join " "

  $previousGoos = $env:GOOS
  $previousGoarch = $env:GOARCH
  $env:GOOS = $Goos
  $env:GOARCH = $Goarch

  $tmpOutput = "$outputPath.tmp-$PID"
  if (Test-Path -LiteralPath $tmpOutput) {
    Remove-Item -LiteralPath $tmpOutput -Force
  }

  try {
    & $go build -trimpath -mod=readonly -ldflags $ldflags -o $tmpOutput ./cmd/l1d
    if ($LASTEXITCODE -ne 0) {
      throw "go build failed"
    }

    if (Test-Path -LiteralPath $outputPath) {
      Remove-Item -LiteralPath $outputPath -Force
    }
    Move-Item -LiteralPath $tmpOutput -Destination $outputPath
  }
  finally {
    $env:GOOS = $previousGoos
    $env:GOARCH = $previousGoarch
    if (Test-Path -LiteralPath $tmpOutput) {
      Remove-Item -LiteralPath $tmpOutput -Force
    }
  }

  $checksum = $null
  if (-not $SkipChecksum) {
    $hash = Get-FileHash -Algorithm SHA256 -LiteralPath $outputPath
    $checksum = $hash.Hash.ToLowerInvariant()
    "$checksum  $(Split-Path -Leaf $outputPath)" | Set-Content -LiteralPath "$outputPath.sha256"
  }

  $relativeOutput = [System.IO.Path]::GetRelativePath($RepoRoot, $outputPath)
  $manifest = [ordered]@{
    builder = $BuilderId
    output = $relativeOutput
    version = $Version
    commit = $commit
    dirty = $dirty
    buildDate = $buildDate
    target = "$Goos/$Goarch"
    goVersion = ($goVersion -join " ")
    modVerify = (-not $SkipModVerify)
    trimpath = $true
    checksumSha256 = $checksum
  }
  $manifest | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath $manifestPath

  Write-Host "Built $relativeOutput"
  if ($checksum) {
    Write-Host "SHA256 $checksum"
  }
  & $outputPath version --long --output json
}
finally {
  Pop-Location
}
