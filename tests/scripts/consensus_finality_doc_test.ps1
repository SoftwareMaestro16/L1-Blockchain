param(
  [string]$Doc = "docs\architecture\consensus-finality.md",
  [string]$Policy = "app\params\consensus_finality.go",
  [string]$Profile = "app\params\network_profile.go",
  [string]$Tests = "app\params\consensus_finality_test.go",
  [string]$VoteExtension = "app\abcihandlers\vote_extension.go",
  [string]$VoteExtensionTests = "app\abcihandlers\vote_extension_test.go"
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
$profileText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Profile)
$testText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Tests)
$voteExtensionText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $VoteExtension)
$voteExtensionTestText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $VoteExtensionTests)

foreach ($term in @(
    'Consensus, Block Production, and Finality',
    'CometBFT BFT consensus',
    'avoids 1-2 second blocks',
    '100-128 validators: block time 5-6 seconds',
    '150-200 validators: block time 6 seconds',
    '250-300 validators: block time 7-8 seconds',
    'normal finality: 5-15 seconds',
    'network stress finality: 20-90 seconds',
    'worst acceptable target: <= 120 seconds',
    'localnet with 100 validators must remain stable',
    'localnet/load profile must demonstrate block production under configured',
    'degraded validator scenarios must preserve liveness when >= 2/3 voting power',
    'finality measurements must be included in testnet reports',
    'ValidateConsensusFinalityReport',
    'Vote Extensions',
    'validator telemetry summary',
    'oracle-like future extensions',
    'encrypted mempool shares if implemented later',
    'keep vote extensions small',
    'verify signatures before trusting extension data',
    'avoid large payloads that hurt consensus latency',
    'avoid non-deterministic validation',
    'cover handlers with tests'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "consensus finality doc missing: $term"
}

foreach ($term in @(
    'ConsensusFinalityReport',
    'ValidateConsensusFinalityReport',
    'AetraHealthyVotingPowerBps',
    '100-128 validator localnet must remain stable',
    '1-2 second block targets are not allowed',
    'degraded scenario must preserve liveness when >= 2/3 voting power is healthy',
    'finality measurements must be included in testnet reports'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "consensus finality policy missing: $term"
}

foreach ($term in @(
    'BlockTimeTargetRange',
    'BlockTimeMinSeconds',
    'BlockTimeMaxSeconds',
    'AetraWorstFinalityTargetSeconds',
    'AetraHealthyVotingPowerBps'
  )) {
  Assert-Contains -Text $profileText -Pattern ([regex]::Escape($term)) -Message "network profile missing consensus finality term: $term"
}

foreach ($term in @(
    'TestConsensusFinalityReportAcceptsRequiredTargets',
    'TestConsensusFinalityReportRejectsUnstableHundredValidatorLocalnet',
    'TestConsensusFinalityReportRejectsOneSecondBlocks',
    'TestConsensusFinalityReportRejectsFinalityOutsideBounds',
    'TestConsensusFinalityReportRequiresDegradedLivenessWithHealthyTwoThirds',
    'TestConsensusFinalityReportRequiresTestnetReportMeasurements'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "consensus finality tests missing: $term"
}

foreach ($term in @(
    'VoteExtensionKindValidatorTelemetrySummary',
    'VoteExtensionKindOracleFutureExtension',
    'VoteExtensionKindEncryptedMempoolShare',
    'MaxVoteExtensionBytes',
    'MaxVoteExtensionDataBytes',
    'AllowedVoteExtensionKind',
    'ValidatorAddress',
    'DeterministicVoteExtensionData'
  )) {
  Assert-Contains -Text $voteExtensionText -Pattern ([regex]::Escape($term)) -Message "vote extension handler missing: $term"
}

foreach ($term in @(
    'TestVoteExtensionHandlerIsDeterministicAndRejectsTampering',
    'TestVoteExtensionPolicyRejectsUnsignedOversizedAndUnknownKinds',
    'TestVoteExtensionAllowedKindsAreExplicitAndSmall'
  )) {
  Assert-Contains -Text $voteExtensionTestText -Pattern ([regex]::Escape($term)) -Message "vote extension tests missing: $term"
}

Write-Host "consensus finality doc test passed"
