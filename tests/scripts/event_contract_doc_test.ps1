param(
  [string]$EventContract = "docs\event-contract.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$EventContractPath = if ([System.IO.Path]::IsPathRooted($EventContract)) { $EventContract } else { Join-Path $RepoRoot $EventContract }

function Assert-Contains {
  param(
    [string]$Text,
    [string]$Pattern,
    [string]$Message
  )
  if ($Text -notmatch $Pattern) {
    throw $Message
  }
}

function Assert-NotContains {
  param(
    [string]$Text,
    [string]$Pattern,
    [string]$Message
  )
  if ($Text -match $Pattern) {
    throw $Message
  }
}

$text = Get-Content -Raw -LiteralPath $EventContractPath

foreach ($eventType in @(
    'contract-assets_create_denom',
    'contract-assets_mint',
    'contract-assets_burn',
    'contract-assets_change_admin',
    'dex_create_pool',
    'dex_add_liquidity',
    'dex_remove_liquidity',
    'dex_swap_exact_amount_in',
    'fees_update_params'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($eventType)) -Message "event contract missing event type: $eventType"
}

foreach ($attribute in @(
    'denom',
    'creator',
    'admin',
    'sender',
    'amount',
    'mint_to_address',
    'burn_from_address',
    'new_admin',
    'pool_id',
    'lp_denom',
    'minted_shares',
    'token_in',
    'token_out',
    'allowed_fee_denom',
    'validator_rewards_ratio',
    'community_pool_ratio'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($attribute)) -Message "event contract missing attribute: $attribute"
}

foreach ($rule in @(
    'do not replace final state queries',
    'deterministic',
    'Failed tx paths must not emit success events',
    'Rejected fee-denom txs',
    'must not include mnemonics',
    'must not iterate maps',
    'Assert-LocalnetTxEvent'
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($rule)) -Message "event contract missing rule: $rule"
}

Assert-NotContains -Text $text -Pattern '(?i)private[_-]?key\s*[:=]|mnemonic:\s*[a-z]|C:\\Users\\|\.localnet\\node\d+\\aetrad' -Message "event contract must not include secrets or local runtime paths"

Write-Host "event contract doc test passed"
