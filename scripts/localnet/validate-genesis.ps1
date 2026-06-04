param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "orbitalis-local-1",
  [int]$ValidatorCount = 3
)

$ErrorActionPreference = "Stop"

$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
if ($OutputDir -eq "") { $OutputDir = Join-Path $RepoRoot ".localnet" }
if ($Binary -eq "") { $Binary = Join-Path $RepoRoot "build\orbitalisd.exe" }
if ($ValidatorCount -lt 1) { throw "ValidatorCount must be at least 1" }

if (!(Test-Path $Binary) -or !(Test-Path $OutputDir)) {
  & (Join-Path $PSScriptRoot "init.ps1") -OutputDir $OutputDir -Binary $Binary -ValidatorCount $ValidatorCount
}

$nodes = Get-ChildItem -LiteralPath $OutputDir -Directory -Filter "node*" |
  Where-Object { Test-Path (Join-Path $_.FullName "orbitalisd\config\genesis.json") } |
  Sort-Object {
    if ($_.Name -match '^node(\d+)$') { [int]$Matches[1] } else { [int]::MaxValue }
  }

if ($nodes.Count -ne $ValidatorCount) {
  throw "Expected $ValidatorCount validator nodes under $OutputDir, found $($nodes.Count)"
}

$firstHash = $null
$secretPattern = '(?i)\b(mnemonic|private[_-]?key|priv_validator|secret|seed|wallet)\b'

foreach ($node in $nodes) {
  $nodeHome = Join-Path $node.FullName "orbitalisd"
  $genesisPath = Join-Path $nodeHome "config\genesis.json"

  & $Binary genesis validate-genesis $genesisPath --home $nodeHome
  if ($LASTEXITCODE -ne 0) {
    throw "genesis validation failed for $genesisPath"
  }

  $raw = Get-Content -Raw -LiteralPath $genesisPath
  if ($raw -match $secretPattern) {
    throw "genesis for $($node.Name) contains secret-like material"
  }

  $hash = (Get-FileHash -LiteralPath $genesisPath -Algorithm SHA256).Hash
  if ($null -eq $firstHash) {
    $firstHash = $hash
  } elseif ($hash -ne $firstHash) {
    throw "genesis hash mismatch for $($node.Name): expected $firstHash, got $hash"
  }

  $doc = $raw | ConvertFrom-Json
  if ($doc.chain_id -ne $ChainId) {
    throw "unexpected chain-id for $($node.Name): $($doc.chain_id)"
  }

  $appState = $doc.app_state
  if ($null -eq $appState) {
    throw "missing app_state for $($node.Name)"
  }

  $bankMetadata = @($appState.bank.denom_metadata | Where-Object { $_.base -eq "norb" })
  if ($bankMetadata.Count -ne 1 -or $bankMetadata[0].display -ne "ORB") {
    throw "native token metadata for norb/ORB is missing or invalid"
  }

  if ($appState.staking.params.bond_denom -ne "norb") {
    throw "staking bond denom is not norb"
  }

  if ($appState.mint.params.mint_denom -ne "norb") {
    throw "mint denom is not norb"
  }

  $feeDenoms = @($appState.fees.params.allowed_fee_denoms)
  if ($feeDenoms.Count -ne 1 -or $feeDenoms[0] -ne "norb") {
    throw "fees module does not restrict fees to norb"
  }

  if (@($appState.tokenfactory.denoms).Count -ne 0) {
    throw "tokenfactory genesis is expected to start with no factory denoms"
  }

  if ([int64]$appState.dex.next_pool_id -ne 1 -or @($appState.dex.pools).Count -ne 0) {
    throw "dex genesis is expected to start with next_pool_id=1 and no pools"
  }

  $genTxs = @($appState.genutil.gen_txs)
  if ($genTxs.Count -ne $ValidatorCount) {
    throw "expected $ValidatorCount gentxs, found $($genTxs.Count)"
  }
}

Write-Host "Validated $ValidatorCount-node genesis for $ChainId at $OutputDir"
Write-Host "genesis sha256: $firstHash"
