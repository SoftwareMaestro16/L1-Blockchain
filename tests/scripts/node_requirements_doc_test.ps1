param(
  [string]$Doc = "docs\architecture\network-node-requirements.md",
  [string]$ValidatorDoc = "docs\validator-onboarding.md",
  [string]$Policy = "app\params\node_requirements.go",
  [string]$Tests = "app\params\node_requirements_test.go"
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
$validatorDocText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $ValidatorDoc)
$policyText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Policy)
$testText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Tests)

foreach ($term in @(
    'Network and Node Requirements',
    'medium hardware class',
    'CPU: 4-8 modern cores',
    'RAM: 16-32 GB',
    'Storage: NVMe SSD',
    'Network: stable 100 Mbps+, low packet loss',
    'OS: Linux recommended, Windows local tooling supported for development',
    'Mainnet requirements should be finalized after load testing',
    'state sync support',
    'snapshots',
    'pruning profiles',
    'archive node profile',
    'export/import reliability',
    'restart safety',
    'deterministic app hash across restarts',
    'documented validator setup',
    'documented sentry architecture',
    'public peers <-> sentry nodes <-> private validator node'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "network node requirements doc missing: $term"
}

foreach ($term in @(
    'Hardware Target',
    'CPU: 4-8 modern cores',
    'RAM: 16-32 GB',
    'Storage: NVMe SSD',
    'Network: stable 100 Mbps+, low packet loss',
    'OS: Linux recommended, Windows local tooling supported for development',
    'state sync support',
    'snapshots',
    'pruning profiles',
    'archive node profile',
    'export/import reliability',
    'restart safety',
    'deterministic app hash',
    'Sentry Architecture',
    'public peers <-> sentry nodes <-> private validator node'
  )) {
  Assert-Contains -Text $validatorDocText -Pattern ([regex]::Escape($term)) -Message "validator onboarding doc missing: $term"
}

foreach ($term in @(
    'NodeHardwareProfile',
    'NodeStateManagementReadiness',
    'DefaultPublicTestnetHardwareProfile',
    'DefaultLocalDevelopmentHardwareProfile',
    'ValidatePublicTestnetHardwareProfile',
    'ValidateLocalDevelopmentHardwareProfile',
    'ValidateMainnetHardwareProfile',
    'ValidateNodeStateManagementReadiness',
    'AetraValidatorCPUCoreMin',
    'AetraValidatorRAMMinGB',
    'AetraValidatorMinNetworkMbps',
    'AetraValidatorStorageClass',
    'AetraValidatorRecommendedOS',
    'AetraValidatorDevelopmentOSSupport',
    'AetraPruningProfileDefault',
    'AetraPruningProfileArchive',
    'AetraPruningProfileAggressive'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "node requirements policy missing: $term"
}

foreach ($term in @(
    'TestDefaultPublicTestnetHardwareProfileMatchesMediumTarget',
    'TestPublicTestnetHardwareRejectsExtremeOrWeakProfiles',
    'TestWindowsIsLocalDevelopmentOnly',
    'TestMainnetHardwareRequiresCompletedLoadTesting',
    'TestDefaultStateManagementReadinessCoversRequiredNodeFeatures',
    'TestStateManagementReadinessRejectsMissingRequiredGates'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "node requirements tests missing: $term"
}

Write-Host "node requirements doc test passed"
