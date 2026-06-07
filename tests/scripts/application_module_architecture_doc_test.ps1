param(
  [string]$Doc = "docs\architecture\application-module-architecture.md",
  [string]$Boundaries = "docs\module-boundaries.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$BoundariesPath = if ([System.IO.Path]::IsPathRooted($Boundaries)) { $Boundaries } else { Join-Path $RepoRoot $Boundaries }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$boundariesText = Get-Content -Raw -LiteralPath $BoundariesPath

foreach ($term in @(
    'Application Module Architecture',
    'Historical DEX Prototype Boundary',
    'production target is contract-only DEX behavior',
    'token, NFT, market, and DEX application logic must not be reintroduced as active native asset modules',
    'pool creator, liquidity provider, withdrawer, swap trader, and swap recipient',
    'native denom spoofing',
    'reserves match the DEX module account balances',
    'LP supply matches pool shares',
    'Contract pool migration',
    '`x/identity` is the planned `.aet` domain registry',
    'source of truth',
    'DomainRecord',
    'lowercase ASCII `a-z`, digits `0-9`, `-`, and `_`',
    'invisible Unicode',
    'mixed-script spoofing',
    'reject the zero',
    'available -> auction -> active -> expired -> available/auction',
    'default auction duration is `24h`',
    'ownership duration is `365 days`',
    'length 1-3: reserved or governance-only',
    'premium_start_price = 10_000 AET',
    'high_start_price    = 1_000 AET',
    'medium_start_price = 100 AET',
    'low_start_price     = 10 AET',
    'default bid increment is `5%`',
    'last `10m` extend the auction',
    'capped at `6` extensions',
    'highest valid bid wins',
    'Failed or premature finalization cannot assign',
    'burn:      40%',
    'treasury:  40%',
    'rewards:   20%',
    'renewal_fee = start_price(name_length) * 0.10',
    'ResolverRecord',
    'domains to wallets',
    '`wallet`',
    '`contract`',
    '`multisig`',
    '`nft`',
    '`dex`',
    'Custom resolver keys are lowercase ASCII',
    'fail "domain not resolved"',
    'registry owner can update resolver records',
    'delegated resolver manager can update only granted keys',
    'unresolved domain transfer fails before funds move',
    'expired domain resolution fails',
    'resolver updates emit deterministic events',
    'domain NFT owner and registry owner must stay consistent',
    '`dex.alice.aet`',
    'explicit extensions',
    'multi-resolver records',
    '`x/workflow` is the planned orchestration module',
    'resolver-based payment',
    'domain auction finalization',
    'token mint and wallet deploy',
    'NFT mint and metadata attach',
    'contract deployment plus first message',
    '`naet` fee checks',
    '`x/identity/types`: domain name and record validation',
    '`x/workflow/types`: bounded workflow and step validation'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "application architecture doc missing: $term"
}

foreach ($pattern in @(
    'x/dex` remains the native DEX module',
    'The native DEX is the current source of truth'
  )) {
  if ($docText -match [regex]::Escape($pattern)) {
    throw "application architecture doc contains active native DEX wording: $pattern"
  }
}

foreach ($term in @(
    'Application module architecture',
    '`x/identity`',
    '`.aet` domain registry',
    '`x/workflow`',
    'bounded multi-step orchestration'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing application architecture: $term"
}

Write-Host "application module architecture doc test passed"
