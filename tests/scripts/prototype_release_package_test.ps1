param(
  [string]$OutputRoot = ".work\release-package-test"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$OutputRoot = if ([System.IO.Path]::IsPathRooted($OutputRoot)) {
  [System.IO.Path]::GetFullPath($OutputRoot)
} else {
  [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $OutputRoot))
}

function Assert-TestPathInsideRepo {
  param([string]$Path)
  $repo = $RepoRoot.TrimEnd('\', '/')
  $full = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  if ($full -eq $repo -or -not $full.StartsWith($repo + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing test path outside repo: $full"
  }
}

function Assert-True {
  param([bool]$Condition, [string]$Message)
  if (-not $Condition) { throw $Message }
}

Assert-TestPathInsideRepo -Path $OutputRoot
if (Test-Path -LiteralPath $OutputRoot) {
  Remove-Item -LiteralPath $OutputRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $OutputRoot | Out-Null

$fakeBin = Join-Path $OutputRoot "fake-orbitalisd.exe"
"fake orbitalisd binary for package test" | Set-Content -LiteralPath $fakeBin

Push-Location $RepoRoot
try {
  & .\scripts\release\prototype-package.ps1 `
    -Version "prototype-test" `
    -Commit "deadbeef0000" `
    -TargetOS windows `
    -TargetArch amd64 `
    -OutputRoot $OutputRoot `
    -Binary $fakeBin `
    -SkipBuild | Out-Host
} finally {
  Pop-Location
}

$packageDir = Join-Path $OutputRoot "prototype-test\orbitalis-prototype-test-windows-amd64"
$archive = Join-Path $OutputRoot "prototype-test\orbitalis-prototype-test-windows-amd64.zip"
$archiveSha = "$archive.sha256"

Assert-True (Test-Path -LiteralPath (Join-Path $packageDir "bin\orbitalisd.exe")) "missing packaged binary"
Assert-True (Test-Path -LiteralPath (Join-Path $packageDir "release-manifest.json")) "missing manifest"
Assert-True (Test-Path -LiteralPath (Join-Path $packageDir "SHA256SUMS.txt")) "missing checksums"
Assert-True (Test-Path -LiteralPath $archive) "missing archive"
Assert-True (Test-Path -LiteralPath $archiveSha) "missing archive checksum"

$manifest = Get-Content -Raw -LiteralPath (Join-Path $packageDir "release-manifest.json") | ConvertFrom-Json
Assert-True ($manifest.version -eq "prototype-test") "manifest version mismatch"
Assert-True ($manifest.commit -eq "deadbeef0000") "manifest commit mismatch"
Assert-True ($manifest.status -match "not mainnet-ready") "manifest must state prototype status"

$checksumText = Get-Content -Raw -LiteralPath (Join-Path $packageDir "SHA256SUMS.txt")
Assert-True ($checksumText -match "bin/orbitalisd.exe") "binary checksum missing"
Assert-True ($checksumText -match "release-manifest.json") "manifest checksum missing"

$forbidden = Get-ChildItem -LiteralPath $packageDir -Recurse -Force |
  Where-Object { $_.FullName -match '\\(\.work|\.localnet|keyring|priv_validator_key\.json|priv_validator_state\.json|node_key\.json)$' }
Assert-True (@($forbidden).Count -eq 0) "package contains forbidden local/private material"

Write-Host "prototype release package script test passed"
