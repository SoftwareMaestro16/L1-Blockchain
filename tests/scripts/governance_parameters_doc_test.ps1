param(
  [string]$Doc = "docs\architecture\governance-parameters.md",
  [string]$Policy = "app\params\governance_parameters.go",
  [string]$Tests = "app\params\governance_parameters_test.go"
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
    'Governance and Parameters',
    'validator set size',
    'validator power cap',
    'commission floor/max',
    'max commission change',
    'inflation min/max',
    'target bonded ratio',
    'fee split',
    'slashing fractions',
    'downtime windows',
    'CosmWasm upload policy',
    'treasury spend policy',
    'params must have min/max validation',
    'unsafe params must be rejected at proposal execution',
    'genesis validation must reject invalid params',
    'parameter changes must emit events',
    'critical changes should use longer voting period or higher quorum'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "governance parameters doc missing: $term"
}

foreach ($term in @(
    'GovernanceParameterSpec',
    'GovernanceParamChange',
    'GovernanceParameterSafetyReport',
    'DefaultGovernanceParameterSpecs',
    'DefaultGovernanceGenesisParams',
    'ValidateGovernanceParameterSpecs',
    'ValidateGovernanceParamChange',
    'ValidateGovernanceGenesisParams',
    'GovernanceParamValidatorSetSize',
    'GovernanceParamValidatorPowerCap',
    'GovernanceParamCommissionFloor',
    'GovernanceParamCommissionMax',
    'GovernanceParamCommissionMaxChange',
    'GovernanceParamInflationMin',
    'GovernanceParamInflationMax',
    'GovernanceParamTargetBondedRatio',
    'GovernanceParamFeeBurnShare',
    'GovernanceParamFeeRewardShare',
    'GovernanceParamFeeTreasuryShare',
    'GovernanceParamDoubleSignSlash',
    'GovernanceParamDowntimeSlash',
    'GovernanceParamDowntimeWindow',
    'GovernanceParamCosmWasmUploadPolicy',
    'GovernanceParamTreasurySpendPolicy',
    'GovernanceCriticalVotingPeriodBlocks',
    'GovernanceCriticalQuorumBps'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "governance parameters policy missing: $term"
}

foreach ($term in @(
    'TestDefaultGovernanceParameterSpecsCoverSection13',
    'TestGovernanceParamChangeRejectsUnsafeExecution',
    'TestCriticalGovernanceParamChangesRequireLongerVotingAndHigherQuorum',
    'TestNonCriticalGovernanceParamChangeUsesNormalVotingBounds',
    'TestGovernanceGenesisValidationRejectsInvalidParams',
    'TestGovernanceEnumParamsAreBounded',
    'TestGovernanceSafetyReportDetectsMissingBoundsGenesisAndEvents'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "governance parameters tests missing: $term"
}

Write-Host "governance parameters doc test passed"
