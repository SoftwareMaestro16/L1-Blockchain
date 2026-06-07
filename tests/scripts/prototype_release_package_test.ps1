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

$fakeBin = Join-Path $OutputRoot "fake-aetrad.exe"
"fake aetrad binary for package test" | Set-Content -LiteralPath $fakeBin

Push-Location $RepoRoot
try {
  & .\scripts\release\prototype-package.ps1 `
    -Version "prototype-test" `
    -Commit "deadbeef0000" `
    -TargetOS windows `
    -TargetArch amd64 `
    -OutputRoot $OutputRoot `
    -Binary $fakeBin `
    -SkipBuild `
    -AllowDirty | Out-Host
} finally {
  Pop-Location
}

$packageDir = Join-Path $OutputRoot "prototype-test\aetra-prototype-test-windows-amd64"
$archive = Join-Path $OutputRoot "prototype-test\aetra-prototype-test-windows-amd64.zip"
$archiveSha = "$archive.sha256"

Assert-True (Test-Path -LiteralPath (Join-Path $packageDir "bin\aetrad.exe")) "missing packaged binary"
Assert-True (Test-Path -LiteralPath (Join-Path $packageDir "release-manifest.json")) "missing manifest"
Assert-True (Test-Path -LiteralPath (Join-Path $packageDir "QUICKSTART.md")) "missing quickstart"
Assert-True (Test-Path -LiteralPath (Join-Path $packageDir "RELEASE-NOTES.md")) "missing release notes"
Assert-True (Test-Path -LiteralPath (Join-Path $packageDir "SHA256SUMS.txt")) "missing checksums"
Assert-True (Test-Path -LiteralPath (Join-Path $packageDir "docs\operator-troubleshooting.md")) "missing troubleshooting runbook"
Assert-True (Test-Path -LiteralPath (Join-Path $packageDir "docs\release\prototype-limitations.md")) "missing limitations doc"
Assert-True (Test-Path -LiteralPath $archive) "missing archive"
Assert-True (Test-Path -LiteralPath $archiveSha) "missing archive checksum"

$manifest = Get-Content -Raw -LiteralPath (Join-Path $packageDir "release-manifest.json") | ConvertFrom-Json
Assert-True ($manifest.version -eq "prototype-test") "manifest version mismatch"
Assert-True ($manifest.commit -eq "deadbeef0000") "manifest commit mismatch"
Assert-True ($manifest.status -match "not mainnet-ready") "manifest must state prototype status"
Assert-True ($null -ne $manifest.dirty) "manifest must include dirty flag"
Assert-True (($manifest.excluded -join " ") -match "\.localnet") "manifest must list excluded runtime state"
Assert-True (($manifest.required_checks -join " ") -match "prototype-audit") "manifest must list security audit check"

$checksumText = Get-Content -Raw -LiteralPath (Join-Path $packageDir "SHA256SUMS.txt")
Assert-True ($checksumText -match "bin/aetrad.exe") "binary checksum missing"
Assert-True ($checksumText -match "release-manifest.json") "manifest checksum missing"
Assert-True ($checksumText -match "QUICKSTART.md") "quickstart checksum missing"
Assert-True ($checksumText -match "RELEASE-NOTES.md") "release notes checksum missing"
Assert-True ($checksumText -match "docs/operator-troubleshooting.md") "troubleshooting runbook checksum missing"
Assert-True ($checksumText -match "docs/release/prototype-limitations.md") "limitations doc checksum missing"

$binaryHash = (Get-FileHash -LiteralPath (Join-Path $packageDir "bin\aetrad.exe") -Algorithm SHA256).Hash.ToLowerInvariant()
Assert-True ($checksumText -match [regex]::Escape("$binaryHash  bin/aetrad.exe")) "binary checksum does not match packaged binary"

$archiveHash = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
$archiveHashText = Get-Content -Raw -LiteralPath $archiveSha
Assert-True ($archiveHashText -match [regex]::Escape($archiveHash)) "archive checksum mismatch"

$forbidden = Get-ChildItem -LiteralPath $packageDir -Recurse -Force |
  Where-Object { $_.FullName -match '\\(\.work|\.localnet|keyring|\.env|diagnostics|priv_validator_key\.json|priv_validator_state\.json|node_key\.json)$' }
Assert-True (@($forbidden).Count -eq 0) "package contains forbidden local/private material"

$notes = Get-Content -Raw -LiteralPath (Join-Path $packageDir "RELEASE-NOTES.md")
Assert-True ($notes -match "dirty tree at package time") "release notes must include dirty flag"
Assert-True ($notes -match "Known Limitations") "release notes must include limitations"
Assert-True ($notes -match "prototype-limitations\.md") "release notes must link limitations doc"
Assert-True ($notes -match "Critical/High") "release notes must separate security blockers from limitations"
Assert-True ($notes -match "prototype-audit") "release notes must list security evidence"

Write-Host "prototype release package script test passed"
