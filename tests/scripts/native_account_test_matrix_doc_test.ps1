param(
  [string]$Matrix = "docs\test-matrix.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$MatrixPath = if ([System.IO.Path]::IsPathRooted($Matrix)) {
  [System.IO.Path]::GetFullPath($Matrix)
} else {
  [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Matrix))
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$text = Get-Content -Raw -LiteralPath $MatrixPath

foreach ($heading in @(
    'Native Account, Staking, Rent, Proof, And Event Matrix',
    'Wallet Tests',
    'Auth Policy Tests',
    'Staking Tests',
    'Reputation Tests',
    'Storage Rent Tests',
    'Proof Tests',
    'Event Tests'
  )) {
  Assert-Contains -Text $text -Pattern "###+ $([regex]::Escape($heading))|## $([regex]::Escape($heading))" -Message "native account test matrix missing heading: $heading"
}

foreach ($term in @(
    'Activate account success.',
    'Activate duplicate rejected.',
    'Address mismatch from `derive(pubkey)` rejected.',
    'Private key not serialized/exported.',
    'Seed phrase not serialized/exported.',
    '`AE...` roundtrip.',
    '`4:...` roundtrip.',
    '`AE... <-> 4:...` roundtrip.',
    'Sequence starts deterministic.',
    'Account number starts deterministic.',
    'Export/import preserves account.',
    'Lazy migration preserves existing account.',
    'Unsupported version rejected safely.',
    'Auth policy update requires authorization.',
    'Frozen account cannot spend.',
    'Archived account keeps minimal required state.',
    'Closed account allowed only with zero obligations.'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($term)) -Message "wallet matrix missing: $term"
}

foreach ($term in @(
    'Single-key success.',
    'Multisig insufficient signatures rejected.',
    'Threshold exact threshold accepted.',
    'Weighted threshold deterministic.',
    'Two-device policy requires device signature for protected action.',
    'Spending limit allows small transfer.',
    'Spending limit rejects large transfer.',
    'Timelock blocks early recovery.',
    'Recovery rotates key without changing address.',
    'Auth policy never serializes private key/seed/2FA secrets.'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($term)) -Message "auth policy matrix missing: $term"
}

foreach ($term in @(
    'User deposits AET into official liquid staking contract and receives pool shares.',
    'Normal user staking deposit does not include validator selection.',
    'Low pool deposit below `MinPoolDeposit` rejected.',
    'Validator entry below `1_000_000 AET` rejected.',
    'Pool-backed validator below `400_000 AET` self-stake rejected.',
    'Pool-backed validator with more than `600_000 AET` nominator stake toward minimum entry rejected.',
    'Direct user delegation to validator rejected.',
    '`MaxValidatorCount` enforced.',
    'Validator count below `100` rejected for production/mainnet params unless explicit testnet override is active.',
    'Validator count above `300` rejected.',
    '`TargetValidatorCount` deterministic behavior.',
    'Allocation engine weights validators by reputation, uptime, commission, limits, stake efficiency, slashing risk, and network load.',
    'Allocation engine excludes jailed/slashed validators from new positive allocation.',
    'Pool injects aggregated stake into multiple validators.',
    'Pool rebalance deterministic and bounded.',
    'Pool unbonding delay of `18 days` enforced.',
    'Reward index deterministic rounding.',
    'Pool-user income equals gross staking rewards minus validator commission, pool fee, and slashing losses.',
    'Validator income equals self-stake rewards plus commission/operator bonus minus infrastructure cost and slashing/jail losses.',
    '300,000 AET pool user illustrative annual math fixture at 14.4% gross yield, 10% validator commission, and 1% pool fee returns `38,491.2 AET`.',
    '300,000 AET validator illustrative annual math fixture with 300,000 AET pool allocation at 14.4% gross yield and 10% commission returns `47,520 AET` before infrastructure cost.',
    'Staking economics tests verify the same formulas for multiple configurable gross yield inputs, not only 14.4%.',
    'Claim rewards updates only caller pool-share state.',
    'No full scan path in reward calculation.',
    'Million-user style scalability test.',
    'Export/import preserves pools/shares/allocations/rewards/unbondings.',
    'Unauthorized params update rejected.',
    'Jailed validators receive no positive validator bonus.'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($term)) -Message "staking matrix missing: $term"
}

foreach ($term in @(
    'Stake-time reputation increases with pool share exposure times duration.',
    'No stake-time means no reputation.',
    'Reputation claim deterministic.',
    'Slashed/jailed validator cannot receive bonus.',
    'Export/import preserves accumulator.',
    'Reputation is account-owned and non-transferable.'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($term)) -Message "reputation matrix missing: $term"
}

foreach ($term in @(
    'Active wallet pays tiny lazy rent.',
    'Unactivated address pays no rent.',
    'Empty/no-state address pays no rent.',
    'Rent is computed from `code_bytes + data_bytes`.',
    'Rent increases with elapsed seconds of storage duration.',
    'Rent is collected automatically during account tx or contract/pool action.',
    'Rent is deducted from the balance of the account/contract/pool/module/protocol payer that owns or funds the state.',
    'Wallet tx effective fee includes gas fee, storage rent delta, and unpaid storage debt.',
    'Wallet with enough balance pays debt and remains `active`.',
    'Wallet with insufficient balance becomes `frozen`.',
    'System/module accounts are not storage-rent exempt and are charged through protocol accounting.',
    'System storage reserve runway is computed deterministically.',
    'System storage reserve warning threshold emits alert.',
    'System storage reserve critical threshold triggers deterministic top-up from fee collector/treasury.',
    'System rent top-up executes before user-account freeze logic.',
    'Protocol-critical system account cannot enter `frozen`, `archived`, or `deleted` status because of storage rent.',
    'Underfunded system rent raises invariant/alert but consensus-critical modules continue executing.',
    'Official liquid staking pool contract has stable `AE...` and `4:...` addresses.',
    'Official liquid staking pool pays storage rent from pool fee/reserve/treasury policy.',
    'Official pool enters `frozen_limited` instead of trapping funds when rent debt exceeds threshold.',
    '`frozen_limited` pool rejects new deposits but allows claims, unbond requests, matured withdrawals, and governance/top-up recovery.',
    'Normal frozen contract rejects execute and new state writes.',
    'Normal frozen contract allows top-up, debt payment, unfreeze, read-only query, and proof query.',
    'Frozen wallet preserves balance, state, sequence, auth/recovery policy, ownership, reputation, pool shares, domains, and pending rewards.',
    'Frozen status never wipes state, resets sequence, deletes code/data, or deletes ownership.',
    'Wallet cannot be automatically deleted for storage debt.',
    'Wallet can be archived only when no balance, stake, pool shares, domains, pending rewards, ownership obligations, or required reputation state remain.',
    'Close rejected while obligations exist.',
    'Export/import preserves rent state.'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($term)) -Message "storage rent matrix missing: $term"
}

foreach ($term in @(
    'Pool deposit proof metadata stable.',
    'Pool share proof metadata stable.',
    'Pool allocation proof metadata stable.',
    'Reward claim proof metadata stable.',
    'Reputation proof metadata stable.',
    'State key deterministic.',
    'Proof query returns height/key/root/path metadata.',
    'Proofs do not require scanning all users.'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($term)) -Message "proof matrix missing: $term"
}

foreach ($term in @(
    '`AccountActivated` stable.',
    '`PoolStakeDeposited` stable.',
    '`PoolSharesMinted` stable.',
    '`PoolAllocationUpdated` stable.',
    '`PoolUnbondingRequested` stable.',
    '`PoolUnbondingCompleted` stable.',
    '`PoolRewardsClaimed` stable.',
    '`StakeReputationClaimed` stable.',
    '`ValidatorRegistered` stable.',
    '`ValidatorUpdated` stable.',
    '`UnbondingStarted` stable.',
    '`UnbondingCompleted` stable.',
    'Events never include private key, seed phrase, or 2FA secrets.'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($term)) -Message "event matrix missing: $term"
}

Write-Host "native account test matrix doc test passed"
