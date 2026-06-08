param(
  [string]$Doc = "docs\architecture\final-target-formula.md",
  [string]$Policy = "app\params\final_target_formula.go",
  [string]$Tests = "app\params\final_target_formula_test.go"
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
    'Final Target Formula',
    'Aetra =',
    'CometBFT BFT PoS',
    'Cosmos SDK',
    'AVM-only genesis smart contracts',
    '100-300 active validators over time',
    '5-8 second block time',
    '<= 120 second worst acceptable finality target',
    'strict objective slashing',
    'validator effective power cap',
    'anti-concentration rewards',
    'dynamic low/moderate inflation',
    'fee burn',
    'protocol treasury',
    'mandatory tests for every feature',
    'Aetra should be a chain people can trust',
    'not a chain optimized only for speed or short-term APR',
    'product direction prioritizes trust over speed-only or short-term APR positioning'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "final target formula doc missing: $term"
}

foreach ($term in @(
    'AetraFinalTargetFormula',
    'FinalTargetFormulaReport',
    'DefaultAetraFinalTargetFormula',
    'ValidateAetraFinalTargetFormula',
    'BuildFinalTargetFormulaReport',
    'FinalTargetConsensusCometBFTBFTPoS',
    'FinalTargetCosmosSDK',
    'FinalTargetAVMOnlyGenesis',
    'FinalTargetValidatorSetRange',
    'FinalTargetBlockTimeRange',
    'FinalTargetWorstFinality',
    'FinalTargetObjectiveSlashing',
    'FinalTargetValidatorPowerCap',
    'FinalTargetAntiConcentrationRewards',
    'FinalTargetDynamicLowModerateInflation',
    'FinalTargetFeeBurn',
    'FinalTargetProtocolTreasury',
    'FinalTargetMandatoryFeatureTests',
    'FinalTargetTrustProductDecision',
    'FinalTargetMinActiveValidators',
    'FinalTargetMaxActiveValidators',
    'FinalTargetMinBlockTimeSeconds',
    'FinalTargetMaxBlockTimeSeconds',
    'FinalTargetWorstFinalitySeconds'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "final target formula policy missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraFinalTargetFormulaMatchesTarget',
    'TestAetraFinalTargetFormulaRejectsMissingCoreStack',
    'TestAetraFinalTargetFormulaRejectsUnsafePerformanceTargets',
    'TestAetraFinalTargetFormulaRejectsMissingTrustEconomicsAndTestPolicy'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "final target formula tests missing: $term"
}

Write-Host "final target formula doc test passed"
