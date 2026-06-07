param(
  [string]$Doc = "docs\architecture\cosmwasm-policy-spec.md",
  [string]$Policy = "app\params\cosmwasm_policy_spec.go",
  [string]$Tests = "app\params\cosmwasm_policy_spec_test.go",
  [string]$RuntimePolicy = "app\wasmconfig\types.go",
  [string]$RuntimeTests = "app\wasmconfig\policy_test.go"
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

$docText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Doc)
$policyText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Policy)
$testText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Tests)
$runtimePolicyText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $RuntimePolicy)
$runtimeTestText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $RuntimeTests)

foreach ($term in @(
    'CosmWasm Policy Specification',
    'Aetra''s native smart contract runtime is AVM',
    'CosmWasm is an optional gated compatibility layer',
    'The implementation gate is `app/params/cosmwasm_policy_spec.go`',
    'The runtime policy gate is `app/wasmconfig`',
    '28.1 Contract Permissions',
    'early testnet:',
    'permissioned code upload or governance-gated upload',
    'later testnet:',
    'permissionless upload with strong fees/deposits',
    'mainnet:',
    'policy decided after security review',
    'AVM remains the primary native contract runtime',
    'Contract migration must require explicit migration authority rules',
    'Pinned code must be disabled by default',
    'BuildAetraCosmWasmLaunchPolicyReport',
    '28.2 Gas And Storage',
    'max wasm code size',
    'max instantiate gas',
    'max execute gas per tx',
    'max query gas',
    'storage rent or storage pricing',
    'contract upload fee',
    'contract migration authority rules',
    'pinned code policy if used',
    '`app/wasmconfig.Policy` must expose explicit bounded fields',
    'deterministic integer accounting',
    'native `naet` fee policy',
    'BuildAetraCosmWasmGasStorageReport',
    '28.3 Contract Security Tests',
    'infinite loop contract hits gas limit',
    'large storage write bounded',
    'failed contract does not corrupt state',
    'contract cannot access reserved module funds',
    'migration authorization enforced',
    'reply/submessage behavior deterministic',
    'event emission stable',
    'export/import with contracts',
    'contract query does not mutate state',
    'Infinite loop and gas exhaustion tests must prove deterministic failure without chain halt',
    'Large storage writes must be bounded by gas, storage pricing, and max storage/write policy',
    'Failed execute/migrate paths must roll back contract state',
    'Contract bank access must be mediated by normal SDK permissions',
    'Smart query execution must not mutate state',
    'BuildAetraCosmWasmContractSecurityTestReport',
    'Required Tests',
    'launch policy tests',
    'gas/storage limit tests',
    'upload fee tests',
    'storage pricing tests',
    'migration authority tests',
    'pinned code policy tests',
    'AVM/CosmWasm boundary tests',
    'contract security tests',
    'BuildAetraCosmWasmTestReport',
    'enable CosmWasm by default',
    'make CosmWasm the primary Aetra runtime'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "cosmwasm policy spec doc missing: $term"
}

foreach ($term in @(
    'AetraCosmWasmPolicyModuleName',
    'AetraCosmWasmRoleOptionalCompatibility',
    'AetraCosmWasmRoleAVMPrimaryRuntime',
    'AetraCosmWasmLaunchPhaseEarlyTestnet',
    'AetraCosmWasmLaunchPhaseLaterTestnet',
    'AetraCosmWasmLaunchPhaseMainnet',
    'AetraCosmWasmUploadPermissionedOrGovernanceGated',
    'AetraCosmWasmUploadPermissionlessWithStrongFeesDeposits',
    'AetraCosmWasmMainnetAfterSecurityReview',
    'AetraCosmWasmGasStorageMaxWasmCodeSize',
    'AetraCosmWasmGasStorageMaxInstantiateGas',
    'AetraCosmWasmGasStorageMaxExecuteGasPerTx',
    'AetraCosmWasmGasStorageMaxQueryGas',
    'AetraCosmWasmGasStorageStorageRentOrPricing',
    'AetraCosmWasmGasStorageContractUploadFee',
    'AetraCosmWasmGasStorageMigrationAuthorityRules',
    'AetraCosmWasmGasStoragePinnedCodePolicyIfUsed',
    'AetraCosmWasmGasStorageGovernanceConfigurable',
    'AetraCosmWasmGasStorageDeterministicAccounting',
    'AetraCosmWasmGasStorageSecurityAndBenchmarkGates',
    'AetraCosmWasmSecurityTestInfiniteLoopGasLimit',
    'AetraCosmWasmSecurityTestLargeStorageWriteBounded',
    'AetraCosmWasmSecurityTestFailedContractStateSafe',
    'AetraCosmWasmSecurityTestReservedModuleFundsDenied',
    'AetraCosmWasmSecurityTestMigrationAuthorization',
    'AetraCosmWasmSecurityTestReplySubmessageDeterminism',
    'AetraCosmWasmSecurityTestStableEvents',
    'AetraCosmWasmSecurityTestExportImportContracts',
    'AetraCosmWasmSecurityTestQueryNoStateMutation',
    'DefaultAetraCosmWasmLaunchPolicyEvidence',
    'ValidateAetraCosmWasmLaunchPolicy',
    'BuildAetraCosmWasmLaunchPolicyReport',
    'DefaultAetraCosmWasmGasStorageEvidence',
    'ValidateAetraCosmWasmGasStorage',
    'BuildAetraCosmWasmGasStorageReport',
    'DefaultAetraCosmWasmTestEvidence',
    'ValidateAetraCosmWasmTests',
    'BuildAetraCosmWasmTestReport',
    'DefaultAetraCosmWasmContractSecurityTestEvidence',
    'ValidateAetraCosmWasmContractSecurityTests',
    'BuildAetraCosmWasmContractSecurityTestReport'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "cosmwasm policy spec gate missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraCosmWasmLaunchPolicyCoversSection281',
    'TestAetraCosmWasmLaunchPolicyRejectsMissingPhaseAndAVMBoundary',
    'TestAetraCosmWasmLaunchPolicyRejectsWrongDuplicateAndUnexpectedPhase',
    'TestDefaultAetraCosmWasmGasStorageCoversSection282',
    'TestAetraCosmWasmGasStorageRejectsMissingControlsAndSafetyGates',
    'TestDefaultAetraCosmWasmTestsCoverImplementationGate',
    'TestAetraCosmWasmTestsRejectMissingDuplicateUnexpectedAndWrongModule',
    'TestDefaultAetraCosmWasmContractSecurityTestsCoverSection283',
    'TestAetraCosmWasmContractSecurityTestsRejectMissingDuplicateUnexpectedAndWrongModule'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "cosmwasm policy spec tests missing: $term"
}

foreach ($term in @(
    'DefaultMaxInstantiateGas',
    'DefaultMaxExecuteGasPerTx',
    'DefaultContractUploadFeeNaet',
    'DefaultStoragePricePerByteEpochNaet',
    'MaxInstantiateGas',
    'MaxExecuteGasPerTx',
    'ContractUploadFeeNaet',
    'StoragePricePerByteEpochNaet'
  )) {
  Assert-Contains -Text $runtimePolicyText -Pattern ([regex]::Escape($term)) -Message "wasm runtime policy missing: $term"
}

foreach ($term in @(
    'TestPolicyDefinesPhase11ReadinessSurface',
    'TestPolicyRejectsUnsafeLimits',
    'TestGasCodeQueryAndPinnedCodePolicies',
    'TestUploadFeeAndStoragePricingPolicy',
    'MaxInstantiateGas',
    'MaxExecuteGasPerTx',
    'ContractUploadFeeNaet',
    'StoragePricePerByteEpochNaet'
  )) {
  Assert-Contains -Text $runtimeTestText -Pattern ([regex]::Escape($term)) -Message "wasm runtime tests missing: $term"
}

Write-Host "cosmwasm policy spec doc test passed"
