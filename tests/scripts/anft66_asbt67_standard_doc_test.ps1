param(
  [string]$Standard = "docs\standards\anft-66-asbt-67.md",
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
    'Aetra NFT Standard',
    'Aetra Soulbound Token Standard',
    'ANFT-66',
    'ASBT-67',
    'nft_collection',
    'nft_item',
    'sbt_item',
    'Collection is the source of truth for item address derivation',
    'NFT transfer requires current owner authorization',
    'SBT owner is immutable after mint',
    'SBT revoke does not transfer ownership',
    'Metadata and royalties are bounded',
    'Metadata must not spoof native AET metadata',
    'royalty policy bounded',
    'Batch minting must have strict limits',
    'x/aetravm/standards/anft',
    'go test ./x/aetravm/standards/anft'
  )) {
  Assert-Contains -Text $standardText -Pattern ([regex]::Escape($term)) -Message "ANFT-66/ASBT-67 standard doc missing: $term"
}

foreach ($term in @(
    'x/aetravm/standards',
    'No chain state',
    'ANFT-66 NFT collection/item model and ASBT-67 soulbound extension',
    'NFT collection/item membership and SBT non-transferability'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing ANFT-66/ASBT-67 boundary: $term"
}

Write-Host "anft/asbt standard doc test passed"
