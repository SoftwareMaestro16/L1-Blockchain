param(
  [string]$Doc = "docs\native-account-staking-reputation.md",
  [string]$Readme = "README.md",
  [string]$ApplicationDoc = "docs\architecture\application-module-architecture.md",
  [string]$Boundaries = "docs\module-boundaries.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) { return $Path }
  return Join-Path $RepoRoot $Path
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

function Assert-NotContains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -cmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Doc)
$readmeText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Readme)
$applicationText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $ApplicationDoc)
$boundariesText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Boundaries)

foreach ($term in @(
    'Native Account, Staking, Reputation, And Rent Model',
    'native wallet/account',
    'Activation Flow',
    'MsgActivateAccount',
    '`AE...` is the only user-facing address family',
    '`4:...` is raw/internal',
    'private-key exclusion',
    'seed phrase exclusion',
    'Private keys and seed phrases are never accepted',
    'User -> Liquid Staking Contract -> Pool Contract -> Validators',
    'Direct user validator delegation is disabled',
    'network staking index',
    'Staking, Pool, And Allocation Params',
    'MaxValidatorCount',
    'MinValidatorStake',
    'MinPoolDeposit',
    'PoolFeeBps',
    'StorageRentRate',
    'Pool Shares, Rewards, And Lazy Reward Index Math',
    'minted_shares = deposit_amount * total_shares / pool_value',
    'reward_index += reward_delta / total_shares',
    'lazy',
    'Validator Versus Pool-User Economics',
    'validator commission',
    'pool fee',
    'operator bonus',
    'infrastructure costs',
    'slashing losses',
    'Allocation Engine Math',
    'reputation',
    'uptime',
    'commission',
    'allocation limits',
    'stake efficiency',
    'slashing risk',
    'network load',
    'Stake Reputation Math',
    'Reputation cannot increase without stake-time',
    'Proof Model',
    'ProofMetadata',
    'Storage Rent Model',
    'Storage rent applies to every active account and persistent on-chain state',
    'official liquid staking pool contracts and pool accounting records',
    'Official Pool Rent And `frozen_limited`',
    'pool fee/reserve payer policy',
    '`frozen_limited`',
    'System And Protocol Storage-Rent Accounting',
    'protocol-payer accounting',
    'deterministic system rent top-up runs before user freeze processing',
    'protocol-critical modules are not frozen',
    'Auth Policy Modes',
    '`single_key`',
    '`multisig`',
    '`threshold`',
    '`weighted`',
    '`two_device`',
    '`timelock`',
    '`recovery`',
    '`spending_limits`',
    'Upgrade And Migration Model',
    'lazy migration',
    'Batched migration jobs',
    'Address derivation does not change',
    'Native Versus Contract Boundary',
    'Tokens, NFTs, DEX pools, markets, auctions, and application-specific assets are contracts or contract standards',
    'not new native `x/` asset modules',
    'Historical native tokenfactory or native DEX prototype docs are migration references only',
    'AFT-44',
    'ANFT-66',
    'TestLazyMigrationPreservesExistingAccountAndTouchesSingleKey',
    'TestMigrateAccountV1ToV2DeterministicGolden',
    'TestNativeAccountInvariantRegistryIncludesEveryRequiredInvariant'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "native account staking doc missing: $term"
}

foreach ($term in @(
    'docs/native-account-staking-reputation.md',
    'Native Account, Staking, Reputation, And Rent Model'
  )) {
  Assert-Contains -Text $readmeText -Pattern ([regex]::Escape($term)) -Message "README missing native account staking doc link: $term"
}

foreach ($term in @(
    'Historical DEX Prototype Boundary',
    'production target is contract-only DEX behavior',
    'must not be reintroduced as active native asset modules'
  )) {
  Assert-Contains -Text $applicationText -Pattern ([regex]::Escape($term)) -Message "application module architecture missing historical DEX wording: $term"
}

foreach ($term in @(
    'Historical `x/tokenfactory` Prototype Boundary',
    'Production fungible token behavior belongs in AVM contracts such as AFT-44',
    'Historical `x/dex` Prototype Boundary',
    'Production DEX pools and routers belong in AVM contracts',
    'token, NFT, market, and DEX application logic must not be reintroduced as active native asset modules'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing historical asset-module wording: $term"
}

foreach ($pattern in @(
    'x/dex` remains the native DEX module',
    'native DEX is the current source of truth',
    'native token/NFT/DEX module design'
  )) {
  Assert-NotContains -Text $applicationText -Pattern ([regex]::Escape($pattern)) -Message "application module architecture still has active native token/NFT/DEX wording: $pattern"
}

foreach ($pattern in @(
    'Purpose: create and manage custom denoms without EVM dependency',
    'Purpose: deterministic constant-product AMM',
    'x/dex` remains a native module while async contract execution and VM'
  )) {
  Assert-NotContains -Text $boundariesText -Pattern ([regex]::Escape($pattern)) -Message "module boundaries still has active native asset-module wording: $pattern"
}

Assert-NotContains -Text $docText -Pattern '```(powershell|bash|sh|go|json)' -Message "native account staking doc examples must stay text-only or gain command validation"

Write-Host "native account staking doc test passed"
