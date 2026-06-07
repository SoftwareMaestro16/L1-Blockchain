param(
  [string]$Doc = "docs\architecture\nomination-pool-spec.md",
  [string]$Policy = "app\params\nomination_pool_spec.go",
  [string]$Tests = "app\params\nomination_pool_spec_test.go"
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
    'Nomination Pool Detailed Specification',
    'Nomination pools are important for accessibility',
    'centralization risks',
    'The implementation gate is `app/params/nomination_pool_spec.go`',
    '26. Nomination Pool Detailed Specification',
    '26.1 Pool Model',
    'Pool:',
    'PoolId',
    'OperatorAddress',
    'ValidatorAddress',
    'TotalBonded',
    'TotalShares',
    'CommissionBps',
    'Status',
    'CreatedHeight',
    'UnbondingEntries',
    'PoolDelegation:',
    'DelegatorAddress',
    'PrincipalEstimate',
    'RewardsAccrued',
    'AetraNominationPoolModuleName',
    'BuildAetraNominationPoolModelReport',
    'ValidateAetraNominationPoolModel',
    'accessibility',
    'deterministic accounting',
    'centralization risks',
    'Current Implementation Mapping',
    'PoolID',
    'PoolOperator',
    'ValidatorTarget',
    'TotalBondedStake',
    'PoolCommissionBps',
    'UnbondingQueue',
    'DelegatorShare',
    'PendingRewards',
    'TotalShares',
    'sum of all delegator shares',
    'deposits mint shares using deterministic integer math',
    'withdrawals burn shares and create unbonding entries',
    'slash losses must reduce pool bonded value without corrupting share supply',
    'export/import must preserve sorted pools, delegations, and unbonding entries',
    'no mandatory KYC should be embedded into consensus pool admission',
    '26.2 Pool Requirements',
    'users deposit native staking denom',
    'pool mints shares deterministically',
    'pool delegates to validator',
    'pool distributes rewards pro-rata',
    'pool commission bounded',
    'pool withdrawal follows unbonding period',
    'pool slashing reduces share value',
    'pool operator cannot withdraw user principal',
    'pool cannot bypass validator power cap',
    'pool must expose risk warnings',
    'AetraNominationPoolRequirementsEvidence',
    'BuildAetraNominationPoolRequirementsReport',
    'ValidateAetraNominationPoolRequirements',
    'share-index accounting',
    'operator permissions must not include user principal withdrawal authority',
    'validator power-cap rules must apply to stake routed through pools',
    '26.3 Pool Tests',
    'first deposit share price',
    'subsequent deposit share price',
    'reward distribution',
    'commission deduction',
    'partial withdrawal',
    'full withdrawal',
    'slashing pool validator',
    'jailed validator',
    'redelegation if allowed',
    'pool operator abuse attempt',
    'export/import with active unbonding entries',
    'rounding dust handling',
    'AetraNominationPoolTestingEvidence',
    'BuildAetraNominationPoolTestingReport',
    'ValidateAetraNominationPoolTesting'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "nomination pool spec doc missing: $term"
}

foreach ($term in @(
    'AetraNominationPoolModuleName',
    'AetraNominationPoolModelEvidence',
    'AetraNominationPoolModelReport',
    'DefaultAetraNominationPoolModelEvidence',
    'ValidateAetraNominationPoolModel',
    'BuildAetraNominationPoolModelReport',
    'AetraNominationPoolFieldPoolID',
    'AetraNominationPoolFieldOperatorAddress',
    'AetraNominationPoolFieldValidatorAddress',
    'AetraNominationPoolFieldTotalBonded',
    'AetraNominationPoolFieldTotalShares',
    'AetraNominationPoolFieldCommissionBps',
    'AetraNominationPoolFieldStatus',
    'AetraNominationPoolFieldCreatedHeight',
    'AetraNominationPoolFieldUnbondingEntries',
    'AetraNominationPoolFieldDelegatorAddress',
    'AetraNominationPoolFieldDelegationPoolID',
    'AetraNominationPoolFieldShares',
    'AetraNominationPoolFieldPrincipalEstimate',
    'AetraNominationPoolFieldRewardsAccrued',
    'AetraNominationPoolRiskAccessibility',
    'AetraNominationPoolRiskAccounting',
    'AetraNominationPoolRiskCentralization',
    'AetraNominationPoolImplementationMap',
    'AetraNominationPoolRequirementsEvidence',
    'AetraNominationPoolRequirementsReport',
    'DefaultAetraNominationPoolRequirementsEvidence',
    'ValidateAetraNominationPoolRequirements',
    'BuildAetraNominationPoolRequirementsReport',
    'AetraNominationPoolRequirementNativeStakingDenom',
    'AetraNominationPoolRequirementDeterministicShareMint',
    'AetraNominationPoolRequirementDelegatesToValidator',
    'AetraNominationPoolRequirementProRataRewards',
    'AetraNominationPoolRequirementCommissionBounded',
    'AetraNominationPoolRequirementWithdrawalUnbonding',
    'AetraNominationPoolRequirementSlashingReducesShare',
    'AetraNominationPoolRequirementOperatorNoPrincipalTheft',
    'AetraNominationPoolRequirementCannotBypassPowerCap',
    'AetraNominationPoolRequirementRiskWarnings',
    'AetraNominationPoolTestingEvidence',
    'AetraNominationPoolTestingReport',
    'DefaultAetraNominationPoolTestingEvidence',
    'ValidateAetraNominationPoolTesting',
    'BuildAetraNominationPoolTestingReport',
    'AetraNominationPoolTestFirstDepositSharePrice',
    'AetraNominationPoolTestSubsequentDepositSharePrice',
    'AetraNominationPoolTestRewardDistribution',
    'AetraNominationPoolTestCommissionDeduction',
    'AetraNominationPoolTestPartialWithdrawal',
    'AetraNominationPoolTestFullWithdrawal',
    'AetraNominationPoolTestSlashingPoolValidator',
    'AetraNominationPoolTestJailedValidator',
    'AetraNominationPoolTestRedelegationIfAllowed',
    'AetraNominationPoolTestOperatorAbuseAttempt',
    'AetraNominationPoolTestExportImportActiveUnbonding',
    'AetraNominationPoolTestRoundingDustHandling'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "nomination pool spec policy missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraNominationPoolModelCoversSection261',
    'TestAetraNominationPoolModelRejectsMissingRequiredFieldsAndRisks',
    'TestAetraNominationPoolModelRejectsDuplicateUnexpectedAndWrongModule',
    'TestDefaultAetraNominationPoolRequirementsCoverSection262',
    'TestAetraNominationPoolRequirementsRejectMissingRequiredItems',
    'TestDefaultAetraNominationPoolTestingCoversSection263',
    'TestAetraNominationPoolTestingRejectsMissingRequiredItems',
    'module_name_required',
    'CreatedHeight',
    'PrincipalEstimate',
    'OperatorKycStatus',
    'LocalUiEstimate'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "nomination pool spec tests missing: $term"
}

Write-Host "nomination pool spec doc test passed"
