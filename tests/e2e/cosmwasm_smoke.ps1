param(
  [string]$Binary = "build\aetrad.exe",
  [string]$Node = "tcp://127.0.0.1:26657",
  [string]$ChainId = "aetra-local-1",
  [string]$AppHome = ".localnet\node0\aetrad",
  [string]$From = "node0",
  [string]$ContractWasm = "",
  [string]$InstantiateMsg = "{}",
  [string]$ExecuteMsg = "{}",
  [string]$QueryMsg = "{}",
  [string]$MigrateWasm = "",
  [string]$MigrateMsg = "{}",
  [string]$Label = "aetra-smoke",
  [string]$Admin = "",
  [int]$TxWaitSeconds = 4,
  [switch]$EnableWasm
)

$ErrorActionPreference = "Stop"

function Resolve-SmokePath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path (Get-Location) $Path))
}

$binaryPath = Resolve-SmokePath $Binary
if (-not (Test-Path -LiteralPath $binaryPath)) {
  throw "aetrad binary not found: $binaryPath"
}

if (-not $EnableWasm) {
  $oldErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  & $binaryPath query wasm params --node $Node *>$null
  $queryExitCode = $LASTEXITCODE
  $ErrorActionPreference = $oldErrorActionPreference
  if ($queryExitCode -eq 0) {
    throw "CosmWasm query unexpectedly succeeded while the feature gate is disabled"
  }
  Write-Host "CosmWasm disabled-by-default smoke passed"
  exit 0
}

if ([string]::IsNullOrWhiteSpace($ContractWasm)) {
  throw "ContractWasm is required when -EnableWasm is set"
}

$contractPath = Resolve-SmokePath $ContractWasm
if (-not (Test-Path -LiteralPath $contractPath)) {
  throw "contract wasm not found: $contractPath"
}

if ([string]::IsNullOrWhiteSpace($Admin)) {
  $Admin = (& $binaryPath keys show $From -a --home $AppHome --keyring-backend test).Trim()
  if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($Admin)) {
    throw "failed to resolve admin address for $From"
  }
}

$storeOutputText = & $binaryPath tx wasm store $contractPath `
  --from $From `
  --home $AppHome `
  --keyring-backend test `
  --chain-id $ChainId `
  --node $Node `
  --fees 1000000naet `
  --output json `
  -y
if ($LASTEXITCODE -ne 0) {
  throw "wasm store failed"
}

$storeOutput = $storeOutputText | ConvertFrom-Json
$txHash = $storeOutput.txhash
if ([string]::IsNullOrWhiteSpace($txHash)) {
  throw "wasm store output did not include txhash"
}

Start-Sleep -Seconds $TxWaitSeconds

$txOutputText = & $binaryPath query tx $txHash --node $Node --output json
if ($LASTEXITCODE -ne 0) {
  throw "failed to query wasm store tx $txHash"
}

$txOutput = $txOutputText | ConvertFrom-Json
$codeId = $null
foreach ($event in $txOutput.events) {
  foreach ($attr in $event.attributes) {
    if ($attr.key -eq "code_id") {
      $codeId = $attr.value
      break
    }
  }
  if (-not [string]::IsNullOrWhiteSpace($codeId)) {
    break
  }
}
if ([string]::IsNullOrWhiteSpace($codeId)) {
  throw "wasm store tx did not emit code_id"
}

& $binaryPath tx wasm instantiate $codeId $InstantiateMsg `
  --from $From `
  --label $Label `
  --admin $Admin `
  --home $AppHome `
  --keyring-backend test `
  --chain-id $ChainId `
  --node $Node `
  --fees 1000000naet `
  --output json `
  -y
if ($LASTEXITCODE -ne 0) {
  throw "wasm instantiate failed for code_id $codeId"
}

Start-Sleep -Seconds $TxWaitSeconds

$contractsOutputText = & $binaryPath query wasm list-contract-by-code $codeId --node $Node --output json
if ($LASTEXITCODE -ne 0) {
  throw "failed to query contracts for code_id $codeId"
}
$contractsOutput = $contractsOutputText | ConvertFrom-Json
$contractAddress = $null
if ($contractsOutput.contracts -and $contractsOutput.contracts.Count -gt 0) {
  $contractAddress = $contractsOutput.contracts[0]
}
if ([string]::IsNullOrWhiteSpace($contractAddress)) {
  throw "wasm instantiate did not create a queryable contract address"
}

& $binaryPath tx wasm execute $contractAddress $ExecuteMsg `
  --from $From `
  --home $AppHome `
  --keyring-backend test `
  --chain-id $ChainId `
  --node $Node `
  --fees 1000000naet `
  --output json `
  -y
if ($LASTEXITCODE -ne 0) {
  throw "wasm execute failed for contract $contractAddress"
}

Start-Sleep -Seconds $TxWaitSeconds

& $binaryPath query wasm contract-state smart $contractAddress $QueryMsg --node $Node --output json *>$null
if ($LASTEXITCODE -ne 0) {
  throw "wasm smart query failed for contract $contractAddress"
}

if (-not [string]::IsNullOrWhiteSpace($MigrateWasm)) {
  $migratePath = Resolve-SmokePath $MigrateWasm
  if (-not (Test-Path -LiteralPath $migratePath)) {
    throw "migrate wasm not found: $migratePath"
  }
  $migrateStoreOutputText = & $binaryPath tx wasm store $migratePath `
    --from $From `
    --home $AppHome `
    --keyring-backend test `
    --chain-id $ChainId `
    --node $Node `
    --fees 1000000naet `
    --output json `
    -y
  if ($LASTEXITCODE -ne 0) {
    throw "wasm migrate code store failed"
  }
  $migrateStoreOutput = $migrateStoreOutputText | ConvertFrom-Json
  $migrateTxHash = $migrateStoreOutput.txhash
  Start-Sleep -Seconds $TxWaitSeconds
  $migrateTxOutputText = & $binaryPath query tx $migrateTxHash --node $Node --output json
  if ($LASTEXITCODE -ne 0) {
    throw "failed to query wasm migrate code store tx $migrateTxHash"
  }
  $migrateTxOutput = $migrateTxOutputText | ConvertFrom-Json
  $migrateCodeId = $null
  foreach ($event in $migrateTxOutput.events) {
    foreach ($attr in $event.attributes) {
      if ($attr.key -eq "code_id") {
        $migrateCodeId = $attr.value
        break
      }
    }
    if (-not [string]::IsNullOrWhiteSpace($migrateCodeId)) {
      break
    }
  }
  if ([string]::IsNullOrWhiteSpace($migrateCodeId)) {
    throw "wasm migrate code store tx did not emit code_id"
  }
  & $binaryPath tx wasm migrate $contractAddress $migrateCodeId $MigrateMsg `
    --from $From `
    --home $AppHome `
    --keyring-backend test `
    --chain-id $ChainId `
    --node $Node `
    --fees 1000000naet `
    --output json `
    -y
  if ($LASTEXITCODE -ne 0) {
    throw "wasm migrate failed for contract $contractAddress"
  }
}

Write-Host "CosmWasm deployment smoke passed with code_id=$codeId contract=$contractAddress admin=$Admin"
