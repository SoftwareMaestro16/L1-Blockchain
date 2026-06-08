param(
  [string]$Doc = "docs\architecture\engineering-priorities.md",
  [string]$Catalog = "app\params\engineering_priorities.go",
  [string]$Tests = "app\params\engineering_priorities_test.go"
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
$catalogText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Catalog)
$testText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Tests)

foreach ($term in @(
    'Engineering Priorities',
    'Priority order',
    'P0:',
    'consensus safety',
    'deterministic state',
    'staking correctness',
    'slashing correctness',
    'supply invariants',
    'export/import',
    'P1:',
    'validator power cap',
    'fee burn/economics',
    'validator score',
    'nomination pool safety',
    'governance bounds',
    'P2:',
    'AVM production hardening',
    'observability',
    'dashboards',
    'load tests',
    'public testnet docs',
    'P3:',
    'advanced anti-cartel analytics',
    'AVM language research',
    'MEV policy',
    'encrypted mempool research',
    'higher validator cap experiments',
    'Do not start P3 until P0 and P1 are stable',
    'P0 is the foundation',
    'P1 is the core Aetra differentiation layer',
    'P2 prepares the public testnet surface',
    'P3 is research and advanced policy',
    'DefaultAetraEngineeringPrioritiesEvidence'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "engineering priorities doc missing: $term"
}

foreach ($term in @(
    'AetraEngineeringPriorityP0',
    'AetraEngineeringPriorityP1',
    'AetraEngineeringPriorityP2',
    'AetraEngineeringPriorityP3',
    'AetraEngineeringP0ConsensusSafety',
    'AetraEngineeringP0DeterministicState',
    'AetraEngineeringP0StakingCorrectness',
    'AetraEngineeringP0SlashingCorrectness',
    'AetraEngineeringP0SupplyInvariants',
    'AetraEngineeringP0ExportImport',
    'AetraEngineeringP1ValidatorPowerCap',
    'AetraEngineeringP1FeeBurnEconomics',
    'AetraEngineeringP1ValidatorScore',
    'AetraEngineeringP1NominationPoolSafety',
    'AetraEngineeringP1GovernanceBounds',
    'AetraEngineeringP2AVMHardening',
    'AetraEngineeringP2Observability',
    'AetraEngineeringP2Dashboards',
    'AetraEngineeringP2LoadTests',
    'AetraEngineeringP2PublicTestnetDocs',
    'AetraEngineeringP3AntiCartelAnalytics',
    'AetraEngineeringP3AVMLanguageResearch',
    'AetraEngineeringP3MEVPolicy',
    'AetraEngineeringP3EncryptedMempoolResearch',
    'AetraEngineeringP3HigherValidatorCapExperiments',
    'AetraEngineeringPriorityEvidence',
    'AetraEngineeringPrioritiesReport',
    'DefaultAetraEngineeringPrioritiesEvidence',
    'ValidateAetraEngineeringPriorities',
    'BuildAetraEngineeringPrioritiesReport',
    'RequiredAetraEngineeringP0Items',
    'RequiredAetraEngineeringP1Items',
    'RequiredAetraEngineeringP2Items',
    'RequiredAetraEngineeringP3Items'
  )) {
  Assert-Contains -Text $catalogText -Pattern ([regex]::Escape($term)) -Message "engineering priorities catalog missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraEngineeringPrioritiesCoverSection35',
    'TestAetraEngineeringPrioritiesRejectP3BeforeP0P1Stable',
    'TestAetraEngineeringPrioritiesRejectMissingDuplicateUnexpectedItems',
    'TestAetraEngineeringPrioritiesRejectDuplicatePriority'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "engineering priorities tests missing: $term"
}

Write-Host "engineering priorities doc test passed"
