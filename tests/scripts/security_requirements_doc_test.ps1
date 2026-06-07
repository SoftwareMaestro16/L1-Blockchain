param(
  [string]$Doc = "docs\architecture\security-requirements.md",
  [string]$Policy = "app\params\security_requirements.go",
  [string]$Tests = "app\params\security_requirements_test.go"
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
    'Security Requirements',
    'Consensus Safety',
    'deterministic state transitions',
    'no non-deterministic external calls in consensus path',
    'no wall-clock dependency in app state transitions except consensus-provided block time',
    'no floating point accounting',
    'no unordered map iteration affecting state',
    'deterministic serialization',
    'export/import equality tests',
    'app hash stability tests',
    'Economic Safety',
    'no unbounded mint',
    'no unauthorized module account mint/burn',
    'supply invariants',
    'fee split invariants',
    'delegation share invariants',
    'reward distribution invariants',
    'slashing cannot underflow stake',
    'jailed validators cannot receive active validator rewards incorrectly',
    'Permission Safety',
    'module account permissions validated at startup',
    'reserved addresses cannot sign user txs',
    'blocked addresses cannot receive normal user funds unless explicitly allowed',
    'governance authority checked',
    'params authority checked',
    'keeper wiring tests'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "security requirements doc missing: $term"
}

foreach ($term in @(
    'ConsensusSafetyRequirements',
    'EconomicSafetyRequirements',
    'PermissionSafetyRequirements',
    'ValidatorRewardEligibility',
    'SecurityRequirementsReport',
    'DefaultConsensusSafetyRequirements',
    'DefaultEconomicSafetyRequirements',
    'DefaultPermissionSafetyRequirements',
    'ValidateSecurityRequirements',
    'BuildSecurityRequirementsReport',
    'ValidateSlashingDoesNotUnderflowStake',
    'ValidateActiveValidatorRewardEligibility',
    'SecurityRequirementDeterministicStateTransitions',
    'SecurityRequirementNoExternalConsensusCalls',
    'SecurityRequirementNoWallClockStateTransitions',
    'SecurityRequirementNoFloatingPointAccounting',
    'SecurityRequirementNoUnorderedMapStateEffects',
    'SecurityRequirementDeterministicSerialization',
    'SecurityRequirementExportImportEqualityTests',
    'SecurityRequirementAppHashStabilityTests',
    'SecurityRequirementNoUnboundedMint',
    'SecurityRequirementNoUnauthorizedModuleMintBurn',
    'SecurityRequirementSupplyInvariants',
    'SecurityRequirementFeeSplitInvariants',
    'SecurityRequirementDelegationShareInvariants',
    'SecurityRequirementRewardDistributionInvariants',
    'SecurityRequirementSlashingNoStakeUnderflow',
    'SecurityRequirementJailedValidatorRewardExclusion',
    'SecurityRequirementModulePermissionsStartup',
    'SecurityRequirementReservedCannotSignUserTxs',
    'SecurityRequirementBlockedCannotReceiveFunds',
    'SecurityRequirementGovernanceAuthorityChecked',
    'SecurityRequirementParamsAuthorityChecked',
    'SecurityRequirementKeeperWiringTests'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "security requirements policy missing: $term"
}

foreach ($term in @(
    'TestDefaultSecurityRequirementsPass',
    'TestConsensusSafetyRejectsNonDeterministicRisks',
    'TestEconomicSafetyRejectsInvariantGaps',
    'TestPermissionSafetyRejectsAuthorityAndWiringGaps',
    'TestPermissionSafetyRequiresExplicitAllowlistAndReservedCatalog',
    'TestSlashingCannotUnderflowStake',
    'TestJailedValidatorsCannotReceiveActiveRewards'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "security requirements tests missing: $term"
}

Write-Host "security requirements doc test passed"
