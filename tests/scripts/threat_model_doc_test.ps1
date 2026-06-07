param(
  [string]$Doc = "docs\architecture\threat-model.md",
  [string]$Policy = "app\params\threat_model_spec.go",
  [string]$Tests = "app\params\threat_model_spec_test.go"
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
    'Threat Model',
    'This document defines section 29 of the Aetra architecture backlog',
    'The implementation gate is `app/params/threat_model_spec.go`',
    '29.1 Validator Cartel',
    'several validators coordinate censorship or governance capture',
    '100-300 validator target',
    'validator power cap',
    'top-N monitoring',
    'commission floor',
    'identity transparency',
    'governance participation metrics',
    'delegation warnings',
    'top-10 concentration simulation',
    'split-identity validator simulation',
    'delegation overflow simulation',
    'governance capture threshold analysis',
    'Cartel detection must use objective chain data',
    'economic signals, warnings, caps, and metrics',
    'Identity transparency must not become mandatory KYC',
    'Concentration warnings must not halt staking',
    'Governance capture threshold analysis must model proposal quorum',
    'Split-identity simulation must assume one operator can run multiple validators',
    'Delegation overflow simulation must prove over-cap stake cannot increase effective voting power',
    'BuildAetraValidatorCartelThreatReport'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "threat model doc missing: $term"
}

foreach ($term in @(
    'AetraThreatModelModuleName',
    'AetraThreatValidatorCartel',
    'AetraThreatControlValidatorSetTarget',
    'AetraThreatControlValidatorPowerCap',
    'AetraThreatControlTopNMonitoring',
    'AetraThreatControlCommissionFloor',
    'AetraThreatControlIdentityTransparency',
    'AetraThreatControlGovernanceParticipationMetrics',
    'AetraThreatControlDelegationWarnings',
    'AetraThreatSimulationTop10Concentration',
    'AetraThreatSimulationSplitIdentityValidator',
    'AetraThreatSimulationDelegationOverflow',
    'AetraThreatSimulationGovernanceCaptureThreshold',
    'AetraValidatorCartelThreatEvidence',
    'DefaultAetraValidatorCartelThreatEvidence',
    'ValidateAetraValidatorCartelThreat',
    'BuildAetraValidatorCartelThreatReport',
    'UsesObjectiveChainData',
    'UsesEconomicSignals',
    'AvoidsMandatoryValidatorKYC',
    'DoesNotHaltStakingOnWarning'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "threat model gate missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraValidatorCartelThreatCoversSection291',
    'TestAetraValidatorCartelThreatRejectsMissingControlsSimulationsAndSafety'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "threat model tests missing: $term"
}

Write-Host "threat model doc test passed"
