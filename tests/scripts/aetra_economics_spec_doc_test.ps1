param(
  [string]$Doc = "docs\architecture\aetra-economics-spec.md",
  [string]$Policy = "app\params\aetra_economics_spec.go",
  [string]$Tests = "app\params\aetra_economics_spec_test.go"
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
    'x/aetra-economics Module Specification',
    'Purpose: low/moderate inflation, fee burn, treasury allocation, reward smoothing, and transparent APR model.',
    'economic-control module of Aetra',
    'implement low/moderate inflation',
    'implement fee burn',
    'implement treasury allocation',
    'implement reward smoothing',
    'expose a transparent APR model',
    'The implementation gate is `app/params/aetra_economics_spec.go`',
    '`AetraEconomicsModuleName` must be `x/aetra-economics`',
    'wrong or missing module identity must fail validation'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "aetra economics spec doc missing: $term"
}

foreach ($term in @(
    'AetraEconomicsModuleName',
    'AetraEconomicsSpecEvidence',
    'AetraEconomicsSpecReport',
    'DefaultAetraEconomicsSpecEvidence',
    'ValidateAetraEconomicsSpec',
    'BuildAetraEconomicsSpecReport',
    'AetraEconomicsPurposeLowModerateInflation',
    'AetraEconomicsPurposeFeeBurn',
    'AetraEconomicsPurposeTreasuryAllocation',
    'AetraEconomicsPurposeRewardSmoothing',
    'AetraEconomicsPurposeTransparentAPRModel'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "aetra economics spec policy missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraEconomicsSpecCoversModulePurpose',
    'TestAetraEconomicsSpecRejectsMissingPurposeComponents',
    'TestAetraEconomicsSpecRejectsWrongModuleIdentity'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "aetra economics spec tests missing: $term"
}

Write-Host "aetra economics spec doc test passed"
