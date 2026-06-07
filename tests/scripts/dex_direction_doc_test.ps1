param(
  [string]$Doc = "docs\architecture\dex-direction.md",
  [string]$Boundaries = "docs\module-boundaries.md",
  [string]$Checklist = "docs\security\cosmos-security-checklist.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$BoundariesPath = if ([System.IO.Path]::IsPathRooted($Boundaries)) { $Boundaries } else { Join-Path $RepoRoot $Boundaries }
$ChecklistPath = if ([System.IO.Path]::IsPathRooted($Checklist)) { $Checklist } else { Join-Path $RepoRoot $Checklist }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$boundariesText = Get-Content -Raw -LiteralPath $BoundariesPath
$checklistText = Get-Content -Raw -LiteralPath $ChecklistPath

foreach ($term in @(
    'DEX Direction',
    'prototype language is now a migration reference',
    'Production DEX pools and routers target AVM contracts',
    'Historical Native Module',
    'historical reference implementation',
    'native `naet`',
    'Pool creator must be a valid non-zero user address',
    'Liquidity provider must be a valid non-zero user address',
    'Liquidity withdrawer must be a valid non-zero user address',
    'Swap trader is also the current swap recipient',
    'Pool asset denoms must not spoof native AET metadata',
    'Factory denoms whose subdenom spoofs `naet`, `AET`, or `Aetra` are rejected',
    'recorded reserves match the `dex` module account balances',
    'LP supply matches `pool.total_shares`',
    'swaps preserve constant-product constraints',
    'slippage bounds are enforced before state mutation',
    'pool_contract',
    'router_contract',
    'AFT-44 LP token master/wallet contracts',
    'async_swap_settlement',
    'deterministic bounce',
    'refund/excess',
    'go test ./x/dex/types ./x/dex/keeper'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "DEX direction doc missing: $term"
}

foreach ($term in @(
    'historical `x/dex` prototype language',
    'Future contract-based pools/routers',
    'Native AET metadata cannot be spoofed',
    'Recorded reserves must match module account balances',
    'LP supply must'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing DEX direction: $term"
}

foreach ($term in @(
    'Native token assumptions use `naet`',
    'Pool creation rejects native AET spoofing',
    'real native base denom `naet`',
    'Ante fee policy accepts prototype tx fees only in allowed denom `naet`'
  )) {
  Assert-Contains -Text $checklistText -Pattern ([regex]::Escape($term)) -Message "security checklist missing DEX/native denom update: $term"
}

Write-Host "dex direction doc test passed"
