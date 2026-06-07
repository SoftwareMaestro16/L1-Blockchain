param(
  [string]$Doc = "docs\architecture\detailed-engineering-scope.md",
  [string]$Policy = "app\params\engineering_scope.go",
  [string]$Tests = "app\params\engineering_scope_test.go"
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

foreach ($term in @(
    'Detailed Engineering Scope',
    'production-grade functionality, not as a placeholder',
    'feature = code + params + genesis validation + queries + events + tests + docs',
    'If a feature has code but does not have tests, query surface, genesis validation, or acceptance criteria, the feature is not complete',
    'Core Chain Configuration',
    'define chain id naming policy for devnet, testnet, and mainnet',
    'define staking denom `naet`',
    'define display denom `AET`',
    'verify coin metadata for `naet/AET`',
    'verify address prefix and reserved system address policy',
    'verify module account permissions',
    'verify blocked address policy',
    'verify mint authority',
    'verify burn authority',
    'verify fee collector authority',
    'verify treasury authority',
    'verify genesis validation for all Aetra modules',
    'verify app export/import with all modules enabled',
    '`app` wiring review',
    'genesis params table',
    'module accounts table',
    'authority matrix',
    'CLI command matrix',
    'query matrix',
    'event matrix',
    'tests for startup validation',
    'app boots with default genesis',
    'app rejects invalid denom metadata',
    'app rejects missing module accounts',
    'app rejects duplicate reserved addresses',
    'app rejects unsafe module account permissions',
    'export/import preserves app hash where expected',
    'simulation or integration test covers module initialization order',
    'Consensus Parameter Policy',
    'define target block time range',
    'define max block bytes',
    'define max block gas',
    'define evidence max age by blocks',
    'define evidence max age by duration',
    'define validator public key types',
    'define CometBFT timeout profile for 100, 200, and 300 validators',
    'define snapshot interval',
    'define state sync parameters',
    'define pruning profiles',
    'Recommended initial values must be conservative',
    'Do not maximize block size early',
    '100 validators: 5-6s',
    '200 validators: 6s',
    '300 validators: 7-8s',
    'start conservative',
    'increase only after load tests',
    'keep below values that increase propagation delay',
    'change only through governance after testnet evidence',
    'conservative initial values',
    'block time target table for 100, 200, and 300 validators',
    'max block gas bounds',
    'max block bytes bounds',
    'evidence window table',
    'timeout profile table',
    'state sync, snapshot, and pruning table',
    'governance safety bounds',
    'localnet remains stable under configured timeout profile',
    'oversized blocks are rejected',
    'invalid consensus params are rejected',
    'governance cannot set unsafe block gas/bytes outside bounds',
    'evidence remains valid through configured evidence period'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "engineering scope doc missing: $term"
}

foreach ($term in @(
    'EngineeringScopeCoreChainConfiguration',
    'EngineeringScopeConsensusParameterPolicy',
    'FeatureCompletionEvidence',
    'EngineeringScopeItem',
    'EngineeringScopePlan',
    'EngineeringScopeReport',
    'ValidateFeatureCompletion',
    'BuildFeatureCompletionReport',
    'DefaultCoreChainConfigurationScopePlan',
    'ValidateEngineeringScopePlan',
    'BuildEngineeringScopeReport',
    'FeatureCompletionCode',
    'FeatureCompletionParams',
    'FeatureCompletionGenesisValidation',
    'FeatureCompletionQueries',
    'FeatureCompletionEvents',
    'FeatureCompletionTests',
    'FeatureCompletionDocs',
    'CoreChainTaskChainIDNamingPolicy',
    'CoreChainTaskStakingDenomNaet',
    'CoreChainTaskDisplayDenomAET',
    'CoreChainTaskCoinMetadata',
    'CoreChainTaskAddressPrefixReserved',
    'CoreChainTaskModuleAccountPermissions',
    'CoreChainTaskBlockedAddressPolicy',
    'CoreChainTaskMintAuthority',
    'CoreChainTaskBurnAuthority',
    'CoreChainTaskFeeCollectorAuthority',
    'CoreChainTaskTreasuryAuthority',
    'CoreChainTaskAetraGenesisValidation',
    'CoreChainTaskAllModulesExportImport',
    'CoreChainDeliverableAppWiringReview',
    'CoreChainDeliverableGenesisParamsTable',
    'CoreChainDeliverableModuleAccountsTable',
    'CoreChainDeliverableAuthorityMatrix',
    'CoreChainDeliverableCLICommandMatrix',
    'CoreChainDeliverableQueryMatrix',
    'CoreChainDeliverableEventMatrix',
    'CoreChainDeliverableStartupValidationTests',
    'CoreChainTestDefaultGenesisBoots',
    'CoreChainTestRejectInvalidDenomMetadata',
    'CoreChainTestRejectMissingModuleAccounts',
    'CoreChainTestRejectDuplicateReservedAddress',
    'CoreChainTestRejectUnsafeModulePermissions',
    'CoreChainTestExportImportAppHash',
    'CoreChainTestModuleInitializationOrder',
    'ConsensusParamTaskBlockTimeRange',
    'ConsensusParamTaskMaxBlockBytes',
    'ConsensusParamTaskMaxBlockGas',
    'ConsensusParamTaskEvidenceMaxAgeBlocks',
    'ConsensusParamTaskEvidenceMaxAgeDuration',
    'ConsensusParamTaskValidatorPubKeyTypes',
    'ConsensusParamTaskTimeoutProfiles',
    'ConsensusParamTaskSnapshotInterval',
    'ConsensusParamTaskStateSyncParameters',
    'ConsensusParamTaskPruningProfiles',
    'ConsensusParamDeliverableConservativeInitialValues',
    'ConsensusParamDeliverableBlockTimeTable',
    'ConsensusParamDeliverableBlockGasBounds',
    'ConsensusParamDeliverableBlockBytesBounds',
    'ConsensusParamDeliverableEvidenceWindowTable',
    'ConsensusParamDeliverableTimeoutProfileTable',
    'ConsensusParamDeliverableStateSyncSnapshotPruning',
    'ConsensusParamDeliverableGovernanceSafetyBounds',
    'ConsensusParamTestLocalnetTimeoutStability',
    'ConsensusParamTestOversizedBlocksRejected',
    'ConsensusParamTestInvalidParamsRejected',
    'ConsensusParamTestGovernanceBounds',
    'ConsensusParamTestEvidencePeriod',
    'DefaultConsensusParameterPolicyScopePlan'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "engineering scope policy missing: $term"
}

foreach ($term in @(
    'TestFeatureCompletionRequiresCodeParamsGenesisQueriesEventsTestsDocs',
    'TestFeatureCompletionRejectsMissingFeatureID',
    'TestDefaultCoreChainConfigurationScopePlanCoversTasksDeliverablesAndTests',
    'TestDefaultConsensusParameterPolicyScopePlanCoversTasksDeliverablesAndTests',
    'TestEngineeringScopeRejectsMissingEvidenceAndRequiredItems',
    'TestEngineeringScopeRejectsUnknownScopeAndUnexpectedItems'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "engineering scope tests missing: $term"
}

Write-Host "engineering scope doc test passed"
