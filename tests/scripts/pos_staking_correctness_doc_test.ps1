param(
  [string]$Doc = "docs\security\pos-staking-correctness.md",
  [string]$NativeDoc = "docs\native-token-lifecycle.md",
  [string]$ValidateGenesis = "scripts\localnet\validate-genesis.ps1"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$NativeDocPath = if ([System.IO.Path]::IsPathRooted($NativeDoc)) { $NativeDoc } else { Join-Path $RepoRoot $NativeDoc }
$ValidateGenesisPath = if ([System.IO.Path]::IsPathRooted($ValidateGenesis)) { $ValidateGenesis } else { Join-Path $RepoRoot $ValidateGenesis }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

function Assert-NotContains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -match $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$nativeDocText = Get-Content -Raw -LiteralPath $NativeDocPath
$validateText = Get-Content -Raw -LiteralPath $ValidateGenesisPath

foreach ($term in @(
    "Phase 6 Aetra",
    "Aetra Slashing System",
    "slashing-system.md",
    "Production Staking Policy",
    "naet",
    "AET has no fixed max supply",
    "max_supply = 0",
    "ValidatorCount",
    "3-validator",
    "5-validator",
    "FinalizeBlock",
    "export/import restart",
    "slashing params",
    "delegator reward withdrawal"
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "PoS correctness doc missing: $term"
}

foreach ($term in @(
    'bond_denom = "naet"',
    "validator self-delegation",
    "commission max rate",
    "unbonding time",
    "redelegation limits",
    "jailed validators",
    "staking, slashing, distribution, mint, and fee params"
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "PoS production policy missing: $term"
}

foreach ($term in @(
    "validator edit, jail, unjail, and inactive validator removal",
    "unbonding completion after the unbonding period",
    "redelegation completion and redelegation limit rejection",
    "downtime jailing behavior",
    "rewards and commission accounting",
    "staking pool invariants"
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "PoS production test backlog missing: $term"
}

foreach ($term in @(
    "Dev",
    "Smoke",
    "Rehearsal",
    "Stress",
    "Long run",
    "10-validator",
    "20-validator"
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "PoS localnet profile matrix missing: $term"
}

foreach ($term in @(
    "stop/start after delegation",
    "stop/start after unbonding",
    "stop/start after slashing state changes",
    "state-sync join after staking changes",
    "snapshot restore after staking changes",
    "kill one validator in a 3-validator profile",
    "kill two validators in a 5-validator profile",
    "stale local state",
    "same staking state root"
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "PoS recovery or chaos drill missing: $term"
}

foreach ($term in @(
    "bonded, unbonded, and not-bonded pools",
    "validator tokens and delegator shares",
    "distribution outstanding rewards",
    'total `naet` supply changes only through mint, reward, slash, and explicit',
    "create validator",
    "delegate",
    "redelegate",
    "unbond",
    "validator set update",
    "reward withdrawal",
    "slashing missed-block window scan"
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "PoS invariant or benchmark missing: $term"
}

foreach ($term in @(
    "Supply: uncapped PoS supply",
    "mint.params.max_supply = `"0`"",
    "aetrad.exe",
    "aetra-local-1",
    "1000000naet",
    "-ValidatorCount 5"
  )) {
  Assert-Contains -Text $nativeDocText -Pattern ([regex]::Escape($term)) -Message "native token lifecycle doc missing: $term"
}

foreach ($pattern in @(
    '\[int\]\$ValidatorCount = 3',
    '\$genTxs\.Count -ne \$ValidatorCount',
    '\$balances\.Count -ne \$ValidatorCount',
    'MsgCreateValidator',
    'expectedSelfDelegation = "100000000"',
    'denom":"naet"',
    'staking bond denom is not naet',
    'mint denom is not naet',
    'fees module does not restrict fees to naet'
  )) {
  Assert-Contains -Text $validateText -Pattern $pattern -Message "validate-genesis missing localnet genesis guard: $pattern"
}

Assert-NotContains -Text $docText -Pattern 'norb|Orbitalis|orbitalisd|ORB' -Message "PoS correctness doc contains old token/network terms"
Assert-NotContains -Text $nativeDocText -Pattern 'norb|Orbitalis|orbitalisd|ORB' -Message "native token lifecycle doc contains old token/network terms"

Write-Host "pos staking correctness doc test passed"
