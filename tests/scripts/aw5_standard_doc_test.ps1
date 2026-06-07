param(
  [string]$Standard = "docs\standards\aw-5.md",
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
    'Aetra Wallet Standard',
    'AW-5',
    'signature_allowed',
    'seqno',
    'wallet_id',
    'valid_until',
    'Signature validation happens before protocol fee acceptance',
    'Extension auth is explicit and revocable',
    'Multi-send is bounded',
    'Recovery policy is explicit and bounded',
    'Wallet cannot silently pay protocol fees in non-`naet`',
    'A relayer can submit a command, but on-chain fee payment is still `naet`',
    'x/aetravm/standards/aw',
    'recovery policy bounded',
    'go test ./x/aetravm/standards/aw'
  )) {
  Assert-Contains -Text $standardText -Pattern ([regex]::Escape($term)) -Message "AW-5 standard doc missing: $term"
}

foreach ($term in @(
    'x/aetravm/standards',
    'No chain state',
    'AW-5 replay-safe contract wallet model',
    'Wallet `seqno`, `wallet_id`, `valid_until`, signature, extension auth, and',
    'native-only fee rules must fail before state mutation'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing AW-5 boundary: $term"
}

Write-Host "aw5 standard doc test passed"
