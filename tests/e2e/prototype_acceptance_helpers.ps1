$ErrorActionPreference = "Stop"

function Write-AcceptanceStep {
  param([string]$Message)

  Write-Host "==> $Message"
}

function Invoke-AcceptanceBuild {
  param([pscustomobject]$Context)

  & (Join-Path $Context.RepoRoot "scripts\build-aetrad.ps1") -Binary $Context.Binary
}

function Invoke-AcceptanceLocalnetScript {
  param(
    [pscustomobject]$Context,
    [string]$ScriptName,
    [hashtable]$Extra = @{}
  )

  $args = @{
    OutputDir      = $Context.OutputDir
    Binary         = $Context.Binary
    ChainId        = $Context.ChainId
    ValidatorCount = $Context.ValidatorCount
    BaseP2PPort    = $Context.BaseP2PPort
    BaseRPCPort    = $Context.BaseRPCPort
    BaseRESTPort   = $Context.BaseRESTPort
    BaseGRPCPort   = $Context.BaseGRPCPort
    BasePprofPort  = $Context.BasePprofPort
    PortStride     = $Context.PortStride
    TimeoutCommit  = $Context.TimeoutCommit
    LogLevel       = $Context.LogLevel
    Profile        = if ($Context.Profile) { $Context.Profile } else { "base" }
    EnableAPI      = $Context.EnableAPI
    EnableGRPC     = $Context.EnableGRPC
    EnableRPC      = $Context.EnableRPC
  }
  foreach ($key in $Extra.Keys) {
    $args[$key] = $Extra[$key]
  }

  & (Join-Path $Context.RepoRoot "scripts\localnet\$ScriptName") @args
}

function Invoke-AcceptanceQueryCliJson {
  param(
    [pscustomobject]$Context,
    [string[]]$Arguments
  )

  return Invoke-LocalnetCliJson -Binary $Context.Binary -Arguments ($Arguments + @("--node", $Context.RpcNode, "--output", "json"))
}

function Invoke-AcceptanceQueryGrpcJson {
  param(
    [pscustomobject]$Context,
    [string[]]$Arguments
  )

  return Invoke-LocalnetCliJson -Binary $Context.Binary -Arguments ($Arguments + @(
      "--grpc-addr", $Context.GrpcAddr,
      "--grpc-insecure",
      "--node", $Context.RpcNode,
      "--output", "json"
    ))
}

function Invoke-AcceptanceRestJson {
  param(
    [pscustomobject]$Context,
    [string]$Path
  )

  return Invoke-RestMethod -Uri "$($Context.RestBase)$Path" -TimeoutSec 5
}

function New-AcceptanceSignedTxArgs {
  param(
    [pscustomobject]$Context,
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [AllowEmptyString()][string]$Fees = ""
  )

  if ([string]::IsNullOrWhiteSpace($Fees)) {
    $Fees = $Context.Fees
  }

  return $ActionArgs + @(
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $Context.ChainId,
    "--keyring-backend", "test",
    "--fees", $Fees,
    "--yes",
    "--broadcast-mode", "sync",
    "--node", $Context.RpcNode,
    "--output", "json"
  )
}

function Send-AcceptanceTx {
  param(
    [pscustomobject]$Context,
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [AllowEmptyString()][string]$Fees = "",
    [switch]$ExpectFailure,
    [string]$ExpectedLog = ""
  )

  return Send-LocalnetTx `
    -Binary $Context.Binary `
    -Arguments (New-AcceptanceSignedTxArgs -Context $Context -ActionArgs $ActionArgs -FromHome $FromHome -FromKey $FromKey -Fees $Fees) `
    -RPCPort $Context.Node0Ports.RPC `
    -TimeoutSeconds $Context.TimeoutSeconds `
    -ExpectFailure:$ExpectFailure `
    -ExpectedLog $ExpectedLog
}

function Get-AcceptanceBalanceAmount {
  param(
    [pscustomobject]$Context,
    [string]$Address,
    [string]$Denom
  )

  $balance = Get-LocalnetBankBalance -Binary $Context.Binary -Address $Address -Denom $Denom -RPCPort $Context.Node0Ports.RPC
  if (-not $balance.amount) {
    return [int64]0
  }
  return [int64]$balance.amount
}

function Get-AcceptanceFactoryDenom {
  param(
    [pscustomobject]$Context,
    [string]$Subdenom
  )

  $denoms = Invoke-AcceptanceQueryGrpcJson -Context $Context -Arguments @("query", "contract-assets", "denoms", "--limit", "100")
  $matches = @($denoms.denoms | Where-Object { [string]$_.denom -like "factory/*/$Subdenom" })
  if ($matches.Count -ne 1) {
    throw "expected exactly one factory denom for subdenom $Subdenom, got $($matches.Count)"
  }
  return [string]$matches[0].denom
}

function Assert-AcceptanceFeesParams {
  param([object]$Params)

  $allowed = @($Params.allowed_fee_denoms)
  if ($allowed.Count -ne 1 -or $allowed[0] -ne "naet") {
    throw "fees params must allow only naet, got $($allowed -join ',')"
  }
}

function Assert-AcceptanceNativeMetadata {
  param([object]$Metadata)

  if ($Metadata.base -ne "naet") {
    throw "native metadata base must be naet, got $($Metadata.base)"
  }
  if ($Metadata.display -ne "AET" -or $Metadata.symbol -ne "AET") {
    throw "native metadata display/symbol must be AET, got $($Metadata.display)/$($Metadata.symbol)"
  }

  $baseUnit = @($Metadata.denom_units | Where-Object { $_.denom -eq "naet" })
  $displayUnit = @($Metadata.denom_units | Where-Object { $_.denom -eq "AET" })
  if ($baseUnit.Count -ne 1 -or (Get-AcceptanceDenomUnitExponent -DenomUnit $baseUnit[0]) -ne 0) {
    throw "native metadata must include naet exponent 0"
  }
  if ($displayUnit.Count -ne 1 -or (Get-AcceptanceDenomUnitExponent -DenomUnit $displayUnit[0]) -ne 9) {
    throw "native metadata must include AET exponent 9"
  }
}

function Get-AcceptanceDenomUnitExponent {
  param([object]$DenomUnit)

  if ($DenomUnit.PSObject.Properties.Name -contains "exponent") {
    return [int]$DenomUnit.exponent
  }
  return 0
}

function Assert-AcceptanceBondedValidator {
  param([object]$Validator)

  $status = [string]$Validator.status
  if ($status -ne "BOND_STATUS_BONDED" -and $status -ne "3") {
    throw "validator $($Validator.operator_address) is not bonded: $status"
  }
}

function Invoke-AcceptanceDiagnostics {
  param(
    [pscustomobject]$Context,
    [string]$Reason
  )

  $safeReason = ($Reason -replace '[^A-Za-z0-9_.-]', '-').Trim('-')
  if ([string]::IsNullOrWhiteSpace($safeReason)) {
    $safeReason = "failure"
  }

  $bundle = Join-Path $Context.RepoRoot ".work\diagnostics\acceptance-$safeReason-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
  try {
    & (Join-Path $Context.RepoRoot "scripts\localnet\diagnostics.ps1") `
      -OutputDir $Context.OutputDir `
      -BundleDir $bundle `
      -ValidatorCount $Context.ValidatorCount `
      -BaseP2PPort $Context.BaseP2PPort `
      -BaseRPCPort $Context.BaseRPCPort `
      -BaseRESTPort $Context.BaseRESTPort `
      -BaseGRPCPort $Context.BaseGRPCPort `
      -BasePprofPort $Context.BasePprofPort `
      -PortStride $Context.PortStride `
      -EnableAPI $Context.EnableAPI `
      -EnableGRPC $Context.EnableGRPC `
      -EnableRPC $Context.EnableRPC `
      -TimeoutSeconds 10 | Out-Host
  } catch {
    Write-Host "Diagnostic collection failed: $($_.Exception.Message)"
  }
}
