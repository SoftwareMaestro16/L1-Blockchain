param(
  [ValidateSet("list", "show", "apply")]
  [string]$Action = "show",
  [string]$OutputDir = "",
  [ValidateSet("base", "execution-os-sim", "zones-prototype", "mesh-prototype", "identity-prototype")]
  [string]$Profile = "base",
  [string]$ChainId = "aetra-local-1",
  [int]$ValidatorCount = 3
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"

switch ($Action) {
  "list" {
    Get-LocalnetProfiles | ForEach-Object { Write-Host $_ }
  }
  "show" {
    $manifest = Read-LocalnetManifest -OutputDir $OutputDir
    $profilePath = Join-Path $OutputDir "profile.json"
    if (Test-Path -LiteralPath $profilePath) {
      Get-Content -Raw -LiteralPath $profilePath
    } elseif ($manifest) {
      $manifest | ConvertTo-Json -Depth 20
    } else {
      [pscustomobject]@{ profile = "base"; output_dir = $OutputDir; production_live = $false } | ConvertTo-Json -Depth 10
    }
  }
  "apply" {
    Set-LocalnetProfileGenesis -OutputDir $OutputDir -Profile $Profile
    Write-LocalnetProfileManifest -OutputDir $OutputDir -Profile $Profile -ValidatorCount $ValidatorCount -ChainId $ChainId
    Write-Host "Applied localnet profile $Profile to $OutputDir"
  }
}
