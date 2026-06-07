param(
  [string]$Standard = "docs\standards\aft-44.md",
  [string]$Boundaries = "docs\module-boundaries.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$StandardPath = if ([System.IO.Path]::IsPathRooted($Standard)) { $Standard } else { Join-Path $RepoRoot $Standard }
$BoundariesPath = if ([System.IO.Path]::IsPathRooted($Boundaries)) { $Boundaries } else { Join-Path $RepoRoot $Boundaries }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$standardText = Get-Content -Raw -LiteralPath $StandardPath
$boundariesText = Get-Content -Raw -LiteralPath $BoundariesPath

foreach ($term in @(
    'Aetra Fungible Token Standard',
    'AFT-44',
    'token_master',
    'token_wallet',
    'Native `AET`/`naet` is not an `AFT-44` token',
    'Wallet contract address is derived',
    'master.total_supply == sum(wallet.balance)',
    'admin controls metadata',
    'Token balances are never accepted as protocol fee payment',
    'Replayed wallet query ids are rejected',
    'Bounce/finalize handling',
    'x/aetravm/standards/aft',
    'go test ./x/aetravm/standards/aft'
  )) {
  Assert-Contains -Text $standardText -Pattern ([regex]::Escape($term)) -Message "AFT-44 standard doc missing: $term"
}

foreach ($term in @(
    'x/aetravm/standards',
    'No chain state',
    'AFT-44 fungible token master/wallet contract model',
    'Native `AET`/`naet` is not a user token standard instance',
    'User token balances cannot satisfy base-chain protocol fees'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing AFT-44 boundary: $term"
}

Write-Host "aft44 standard doc test passed"
