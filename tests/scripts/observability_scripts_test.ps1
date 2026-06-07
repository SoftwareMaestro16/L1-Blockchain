param(
  [string]$OutputDir = ".work\observability-script-test"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$OutputDir = if ([System.IO.Path]::IsPathRooted($OutputDir)) {
  [System.IO.Path]::GetFullPath($OutputDir)
} else {
  [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $OutputDir))
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

Assert-TestPathInsideRepo -Path $OutputDir
if (Test-Path -LiteralPath $OutputDir) {
  Remove-Item -LiteralPath $OutputDir -Recurse -Force
}

$nodeHome = Join-Path $OutputDir "node0\aetrad"
$configDir = Join-Path $nodeHome "config"
$logDir = Join-Path $OutputDir "logs"
$keyringDir = Join-Path $nodeHome "keyring-test"
New-Item -ItemType Directory -Force -Path $configDir, $logDir, $keyringDir | Out-Null

'{"genesis_time":"2026-06-04T00:00:00Z","chain_id":"aetra-test"}' | Set-Content -LiteralPath (Join-Path $configDir "genesis.json")
@'
[api]
enable = false
address = "tcp://0.0.0.0:1317"

[grpc]
enable = false
address = "127.0.0.1:9090"

[telemetry]
enabled = false
sink = ""
'@ | Set-Content -LiteralPath (Join-Path $configDir "app.toml")
@'
[rpc]
laddr = ""
'@ | Set-Content -LiteralPath (Join-Path $configDir "config.toml")
'{"priv_key":{"type":"placeholder","value":"do-not-copy"}}' | Set-Content -LiteralPath (Join-Path $configDir "priv_validator_key.json")
'mnemonic: redaction fixture phrase only' | Set-Content -LiteralPath (Join-Path $keyringDir "leak.txt")
'panic recovered; private_key = placeholder-redaction-value; fee denom uatom not accepted; account sequence mismatch' | Set-Content -LiteralPath (Join-Path $logDir "node0.err.log")

Push-Location $RepoRoot
try {
  $health = & .\scripts\localnet\health.ps1 `
    -OutputDir $OutputDir `
    -ValidatorCount 1 `
    -EnableRPC $false `
    -EnableAPI $false `
    -EnableGRPC $false `
    -Json | ConvertFrom-Json

  Assert-True ($health.rpc -eq "disabled") "RPC should be disabled in telemetry-only health mode"
  Assert-True ($health.rest -eq "disabled") "REST should be disabled in telemetry-only health mode"
  Assert-True ($health.grpc -eq "disabled") "gRPC should be disabled in telemetry-only health mode"
  Assert-True ($health.telemetry[0].telemetry.enabled -eq $false) "telemetry disabled mode was not reported"
  Assert-True ($health.recent_logs[0].recent -match "fee denom uatom not accepted") "recent logs should include tx failure clues"
  Assert-True ($health.recent_logs[0].recent -notmatch "placeholder-redaction-value") "recent logs must be redacted"

  try {
    & .\scripts\localnet\health.ps1 -OutputDir $OutputDir -ValidatorCount 1 -TimeoutSeconds 1 -EnableAPI $false -EnableGRPC $false | Out-Null
    throw "health should fail when RPC endpoint is missing"
  } catch {
    Assert-True ($_.Exception.Message -match "RPC port") "missing endpoint failure should mention RPC port"
  }

  $bundle = Join-Path $OutputDir "bundle"
  & .\scripts\localnet\diagnostics.ps1 `
    -OutputDir $OutputDir `
    -BundleDir $bundle `
    -ValidatorCount 1 `
    -EnableRPC $false `
    -EnableAPI $false `
    -EnableGRPC $false | Out-Null

  Assert-True (Test-Path -LiteralPath (Join-Path $bundle "manifest.json")) "diagnostic manifest missing"
  Assert-True (Test-Path -LiteralPath (Join-Path $bundle "health.json")) "diagnostic health output missing"
  Assert-True (Test-Path -LiteralPath (Join-Path $bundle "processes.json")) "diagnostic process list missing"
  Assert-True (-not (Test-Path -LiteralPath (Join-Path $bundle "node0\config\priv_validator_key.json"))) "private validator key must not be copied"
  Assert-True (-not (Test-Path -LiteralPath (Join-Path $bundle "node0\keyring-test"))) "keyring directory must not be copied"

  $bundleLog = Get-Content -Raw -LiteralPath (Join-Path $bundle "logs\node0.err.log")
  Assert-True ($bundleLog -notmatch "placeholder-redaction-value") "diagnostic logs must redact private key values"
  Assert-True ($bundleLog -match "\[REDACTED\]") "diagnostic logs should show redaction marker"
} finally {
  Pop-Location
}

Write-Host "observability script test passed"
