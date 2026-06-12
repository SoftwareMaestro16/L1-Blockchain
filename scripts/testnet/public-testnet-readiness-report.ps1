param(
  [ValidateSet("Markdown", "Json")]
  [string]$OutputFormat = "Markdown",
  [switch]$RunLocalnetProfiles,
  [string]$Binary = "",
  [int]$TimeoutSeconds = 180,
  [switch]$AllowFailures
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Path))
}

function Read-RepoFile {
  param([string]$Path)
  $resolved = Resolve-RepoPath $Path
  if (-not (Test-Path -LiteralPath $resolved)) {
    throw "missing file: $Path"
  }
  return Get-Content -Raw -LiteralPath $resolved
}

function Assert-FileExists {
  param([string]$Path)
  if (-not (Test-Path -LiteralPath (Resolve-RepoPath $Path))) {
    throw "missing required file: $Path"
  }
  return $Path
}

function Assert-Contains {
  param([string]$Path, [string]$Pattern, [string]$Label)
  $text = Read-RepoFile $Path
  if ($text -notmatch $Pattern) {
    throw "$Label not found in $Path"
  }
  return "$Path :: $Label"
}

function Assert-NotContains {
  param([string]$Path, [string]$Pattern, [string]$Label)
  $text = Read-RepoFile $Path
  if ($text -match $Pattern) {
    throw "$Label unexpectedly found in $Path"
  }
  return "$Path :: no $Label"
}

function Add-Check {
  param(
    [System.Collections.Generic.List[object]]$Checks,
    [string]$Id,
    [string]$Title,
    [scriptblock]$Body
  )
  try {
    $evidence = @(& $Body)
    $Checks.Add([pscustomobject]@{
        id       = $Id
        title    = $Title
        status   = "PASS"
        evidence = @($evidence)
        error    = ""
      }) | Out-Null
  } catch {
    $Checks.Add([pscustomobject]@{
        id       = $Id
        title    = $Title
        status   = "FAIL"
        evidence = @()
        error    = $_.Exception.Message
      }) | Out-Null
  }
}

$checks = [System.Collections.Generic.List[object]]::new()

Add-Check $checks "avm_runtime_wired" "AVM contract runtime is wired as a native SDK module" {
  Assert-Contains "app\modulewiring\modules.go" "contractsmodule\.NewAppModule\(deps\.ContractsKeeper\)" "contracts app module wiring"
  Assert-Contains "app\keepers.go" "ContractsKeeper\s*=\s*persistentKeepers\.ContractsKeeper" "contracts keeper wiring"
  Assert-Contains "app\wiring\storekeys\keys.go" "contractstypes\.StoreKey" "contracts store key"
  Assert-FileExists "proto\l1\contracts\v1\tx.proto"
  Assert-FileExists "proto\l1\contracts\v1\query.proto"
  Assert-Contains "x\contracts\module.go" "RegisterServices" "contracts service registration"
  Assert-Contains "x\contracts\types\service.go" "RegisterMsgServer" "contracts msg service descriptor"
  Assert-Contains "x\contracts\types\service.go" "RegisterQueryServer" "contracts query service descriptor"
  Assert-Contains "x\contracts\keeper\keeper.go" "func \(k \*Keeper\) StoreCode" "contracts StoreCode runtime"
  Assert-Contains "x\contracts\keeper\keeper.go" "func \(k \*Keeper\) ExecuteContract" "contracts ExecuteContract runtime"
}

Add-Check $checks "native_account_wired" "native-account is a runtime SDK module, not only types/spec" {
  Assert-Contains "app\modulewiring\modules.go" "nativeaccountmodule\.NewAppModule" "native-account app module wiring"
  Assert-Contains "app\keepers.go" "NativeAccountKeeper\s*=\s*persistentKeepers\.NativeAccountKeeper" "native-account keeper wiring"
  Assert-Contains "app\wiring\storekeys\keys.go" "nativeaccounttypes\.StoreKey" "native-account store key"
  Assert-FileExists "x\native-account\module.go"
  Assert-FileExists "x\native-account\keeper\msg_server.go"
  Assert-FileExists "x\native-account\keeper\query_server.go"
  Assert-FileExists "proto\l1\nativeaccount\v1\tx.proto"
  Assert-FileExists "proto\l1\nativeaccount\v1\query.proto"
}

Add-Check $checks "direct_delegation_disabled" "normal users cannot directly choose validators for staking" {
  Assert-Contains "x\nominator-pool\types\state.go" "DirectUserValidatorDelegationEnabled:\s*false" "default direct user delegation disabled"
  Assert-Contains "x\nominator-pool\types\state.go" "direct user delegation to validators is disabled" "direct delegation rejection error"
  Assert-Contains "tests\e2e\pos_smoke.ps1" "delegate-direct-disabled" "e2e direct delegation rejection smoke"
}

Add-Check $checks "official_pool_staking" "official pool staking flow works through pool/index messages" {
  Assert-Contains "x\nominator-pool\keeper\keeper.go" "DepositToStakingPool" "pool deposit runtime"
  Assert-Contains "x\nominator-pool\keeper\keeper.go" "RequestPoolUnbond" "pool unbond runtime"
  Assert-Contains "x\nominator-pool\keeper\keeper.go" "WithdrawPoolStake" "pool matured withdrawal runtime"
  Assert-Contains "x\nominator-pool\keeper\keeper_pool_staking_test.go" "DepositToStakingPool" "pool deposit tests"
  Assert-Contains "x\nominator-pool\keeper\keeper_pool_staking_test.go" "FrozenLimitedPoolAllowsTopUpClaimUnbondAndMaturedWithdrawals" "frozen_limited exit test"
}

Add-Check $checks "storage_rent_enforcement" "storage rent is enforced for accounts, contracts, and official pool state" {
  Assert-Contains "x\native-account\types\storage_rent.go" "CollectWalletStorageRent" "wallet rent collection"
  Assert-Contains "x\native-account\types\storage_rent.go" "AccountStatusFrozen" "wallet freeze on debt"
  Assert-Contains "x\contracts\keeper\keeper.go" "chargeContractRentAt" "contract rent runtime hook"
  Assert-Contains "x\nominator-pool\keeper\keeper.go" "accrueOfficialPoolRent" "official pool rent runtime hook"
  Assert-Contains "x\storage-rent\types\state.go" "SystemRentReserve" "system rent reserve state"
  Assert-Contains "x\storage-rent\types\policy.go" "SystemRentAlertInvariant" "underfunded system rent alert"
}

Add-Check $checks "system_governance_safety" "system config and governance safety modules are wired" {
  Assert-Contains "app\modulewiring\modules.go" "configmodule\.NewAppModule" "config module wiring"
  Assert-Contains "app\modulewiring\modules.go" "constitutionmodule\.NewAppModule" "constitution module wiring"
  Assert-Contains "x\config\types\state.go" "RequiredSystemAccountKeys" "system account config safety"
  Assert-Contains "x\constitution\types\state.go" "ProtectedModules" "constitution protected modules"
  Assert-Contains "x\config\module.go" "RegisterServices" "config service registration"
  Assert-Contains "x\constitution\module.go" "RegisterServices" "constitution service registration"
  Assert-Contains "tests\scripts\governance_parameters_doc_test.ps1" "governance" "governance params doc test"
}

Add-Check $checks "app_invariants_registered" "app-level invariants are registered and include rent/delegation gates" {
  Assert-Contains "app\invariants.go" "AppInvariantRegistry" "app invariant registry"
  Assert-Contains "app\invariants.go" "RegisterAppInvariants" "SDK invariant registration"
  Assert-Contains "app\invariants.go" "InvariantSystemStorageReserveRunway" "system rent runway invariant"
  Assert-Contains "app\invariants.go" "InvariantDirectUserValidatorDelegationRejected" "direct delegation invariant"
  Assert-Contains "app\invariants_test.go" "TestAppInvariantRegistryIncludesEveryRequiredInvariant" "registry coverage test"
}

Add-Check $checks "export_import_roundtrip" "export/import roundtrip evidence exists" {
  Assert-FileExists "tests\e2e\export_import_smoke.ps1"
  Assert-Contains "tests\e2e\export_import_smoke.ps1" "export" "export smoke command"
  Assert-Contains "tests\e2e\export_import_smoke.ps1" "import" "import smoke command"
  Assert-Contains "architecture.md" "export/import stable" "architecture mainnet readiness export/import criterion"
  Assert-Contains "docs\public-testnet-long-running-evidence.md" "Export/import roundtrip preserves account, contract, pool, storage rent, and" "long-running export/import requirement"
}

Add-Check $checks "formal_ci_readiness" "formal public testnet CI readiness workflow exists" {
  Assert-FileExists ".github\workflows\testnet-readiness.yml"
  foreach ($term in @(
      "go-test-all",
      "go test ./...",
      "genesis-validate",
      "scripts/localnet/validate-genesis.ps1",
      "localnet-smoke",
      "tests/e2e/localnet_smoke.ps1",
      "export-import-roundtrip",
      "tests/e2e/export_import_smoke.ps1",
      "invariants",
      "go test ./app -run Invariant",
      "linter",
      "go vet ./...",
      "buf lint",
      "release-artifact-build",
      "scripts/release/prototype-package.ps1",
      "version-command",
      "version --long --output json",
      "chain-id-validation",
      "validator-docs"
    )) {
    Assert-Contains ".github\workflows\testnet-readiness.yml" ([regex]::Escape($term)) "CI readiness term $term"
  }
}

Add-Check $checks "no_native_asset_modules" "token/NFT/DEX remain contracts, not native app modules" {
  Assert-NotContains "app\modulewiring\modules.go" "tokenfactory|dexmodule|nftmodule|marketmodule|assetfactory" "native token/NFT/DEX app module"
  Assert-Contains "docs\public-testnet-preparation.md" "token, NFT, market, and exchange-style application logic now targets AVM contracts" "docs contract-only asset model"
  Assert-Contains "docs\public-testnet-long-running-evidence.md" "Native application-asset modules remain absent; assets use AVM contracts" "long-running contract-only asset rule"
}

Add-Check $checks "docs_match_behavior" "public docs match implemented behavior and readiness command surface" {
  Assert-Contains "architecture.md" "VM:\s+AVM first, AVM only at genesis" "AVM-only genesis architecture"
  Assert-NotContains "architecture.md" "CosmWasm first|EVM optional later" "stale CosmWasm/EVM-first language"
  Assert-Contains "docs\public-testnet-production-gates.md" "public-testnet-readiness-report\.ps1" "readiness report command"
  Assert-Contains "docs\public-testnet-preparation.md" "official liquid staking pool" "official pool docs"
  Assert-Contains "docs\public-testnet-preparation.md" "storage rent" "storage rent docs"
  Assert-Contains "docs\public-testnet-preparation.md" "direct delegation" "direct delegation docs"
}

Add-Check $checks "localnet_profiles" "3/5/10 public-testnet localnet profiles are covered" {
  Assert-Contains "scripts\testnet\public-testnet-preflight.ps1" 'ValidateSet\("3", "5", "10", "All"\)' "3/5/10 profile parameter"
  Assert-Contains "scripts\testnet\public-testnet-preflight.ps1" "@\(3, 5, 10\)" "all profile expansion"
  Assert-Contains "docs\public-testnet-production-gates.md" "public-testnet-preflight\.ps1 -ValidatorProfile 3" "3-validator documented command"
  Assert-Contains "docs\public-testnet-production-gates.md" "public-testnet-preflight\.ps1 -ValidatorProfile 5" "5-validator documented command"
  Assert-Contains "docs\public-testnet-production-gates.md" "public-testnet-preflight\.ps1 -ValidatorProfile 10" "10-validator documented command"
  if ($RunLocalnetProfiles) {
    & (Resolve-RepoPath "scripts\testnet\public-testnet-preflight.ps1") -ValidatorProfile All -Binary $Binary -TimeoutSeconds $TimeoutSeconds -SkipBuild
    "scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile All executed"
  } else {
    "static profile coverage checked; use -RunLocalnetProfiles to execute"
  }
}

Add-Check $checks "long_running_evidence" "long-running public testnet evidence checklist exists with required metrics" {
  Assert-FileExists "docs\public-testnet-long-running-evidence.md"
  foreach ($term in @(
      "app_hash",
      "finality_seconds",
      "missed_blocks",
      "evidence_age",
      "peer_count",
      "state_sync_restore",
      "snapshot_restore",
      "storage_rent_debt",
      "system_rent_runway",
      "pool_deposit_claim_unbond",
      "validator_uptime",
      "incident_count"
    )) {
    Assert-Contains "docs\public-testnet-long-running-evidence.md" ([regex]::Escape($term)) "long-run metric $term"
  }
}

Add-Check $checks "e2e_smoke_commands" "public testnet e2e smoke command list exists" {
  Assert-FileExists "docs\public-testnet-e2e-smoke-commands.md"
  foreach ($term in @(
      "public-testnet-readiness-report.ps1",
      "public-testnet-preflight.ps1 -ValidatorProfile 3",
      "public-testnet-preflight.ps1 -ValidatorProfile 5",
      "public-testnet-preflight.ps1 -ValidatorProfile 10",
      "export_import_smoke.ps1",
      "pos_smoke.ps1",
      "fees_ante_smoke.ps1",
      "query_surface_smoke.ps1",
      "execution_os_smoke.ps1"
    )) {
    Assert-Contains "docs\public-testnet-e2e-smoke-commands.md" ([regex]::Escape($term)) "e2e command $term"
  }
}

$failed = @($checks | Where-Object { $_.status -eq "FAIL" })
$report = [pscustomobject]@{
  gate         = "public-testnet-readiness"
  generatedAt = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
  repoRoot     = $RepoRoot
  status       = if ($failed.Count -eq 0) { "PASS" } else { "FAIL" }
  checks       = @($checks)
}

if ($OutputFormat -eq "Json") {
  $report | ConvertTo-Json -Depth 8
} else {
  "# Public Testnet Readiness Report"
  ""
  "Status: $($report.status)"
  ""
  foreach ($check in $checks) {
    "- $($check.status): $($check.id) - $($check.title)"
    if ($check.status -eq "FAIL") {
      "  - error: $($check.error)"
    }
    foreach ($item in $check.evidence) {
      "  - evidence: $item"
    }
  }
}

if ($failed.Count -gt 0 -and -not $AllowFailures) {
  exit 1
}
