param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$NodeIndex = 0,
  [int]$Height = 0,
  [string]$ArchivePath = ""
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$manifest = Read-LocalnetManifest -OutputDir $OutputDir
if ($null -eq $manifest) {
  throw "localnet manifest not found in $OutputDir"
}
if ($NodeIndex -lt 0 -or $NodeIndex -ge [int]$manifest.validator_count) {
  throw "node index $NodeIndex out of range for $($manifest.validator_count) validators"
}

$nodeHome = Get-NodeHome -OutputDir $OutputDir -Index $NodeIndex
$exportArgs = @("snapshots", "export", "--home", $nodeHome)
if ($Height -gt 0) {
  $exportArgs += @("--height", "$Height")
}
Invoke-ExternalChecked -FilePath $Binary -Arguments $exportArgs -FailureMessage "snapshot export failed" | Out-Null

$listOutput = Invoke-ExternalChecked -FilePath $Binary -Arguments @("snapshots", "list", "--home", $nodeHome) -FailureMessage "snapshot list failed"
$listOutput | ForEach-Object { Write-Host $_ }

if (![string]::IsNullOrWhiteSpace($ArchivePath) -and $Height -gt 0) {
  $ArchivePath = ConvertTo-AbsolutePath -Path $ArchivePath
  New-Item -ItemType Directory -Force -Path (Split-Path -Parent $ArchivePath) | Out-Null
  Invoke-ExternalChecked -FilePath $Binary -Arguments @("snapshots", "dump", "$Height", "1", "--home", $nodeHome, "--output", $ArchivePath) -FailureMessage "snapshot dump failed" | Out-Null
  Write-Host "Snapshot archive written to $ArchivePath"
}
