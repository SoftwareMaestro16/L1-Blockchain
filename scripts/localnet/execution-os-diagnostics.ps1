param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [ValidateSet("base", "execution-os-sim", "zones-prototype", "mesh-prototype", "identity-prototype")]
  [string]$Profile = "base",
  [int]$NodeIndex = 0,
  [switch]$Json
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"
Assert-LocalnetProfile -Profile $Profile

if (-not (Test-Path -LiteralPath $Binary)) {
  throw "Binary not found at $Binary"
}

$nodeHome = Get-NodeHome -OutputDir $OutputDir -Index $NodeIndex
$genesisPath = Join-Path $nodeHome "config\genesis.json"
if (-not (Test-Path -LiteralPath $genesisPath)) {
  throw "Genesis file not found at $genesisPath"
}

$diagnostics = Invoke-LocalnetCliJson -Binary $Binary -Arguments @(
  "execution-os", "diagnostics",
  "--profile", $Profile,
  "--genesis", $genesisPath
)

if ($Json) {
  $diagnostics | ConvertTo-Json -Depth 100
} else {
  Write-Host "profile: $($diagnostics.profile)"
  Write-Host "source: $($diagnostics.source)"
  Write-Host "current_load_score_bps: $($diagnostics.current_load_score_bps)"
  Write-Host "active_zones: $(@($diagnostics.active_zones) -join ',')"
  Write-Host "active_shards: $(@($diagnostics.active_shards | ForEach-Object { "$($_.zone_id)=$($_.active_shards)" }) -join ',')"
  Write-Host "pending_mesh_messages: $($diagnostics.pending_mesh_messages)"
  Write-Host "replay_marker_count: $($diagnostics.replay_marker_count)"
  Write-Host "zone_commitment_roots: $(@($diagnostics.zone_commitment_roots) -join ',')"
}
