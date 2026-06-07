param(
  [string]$Doc = "docs\architecture\definition-of-done.md",
  [string]$Catalog = "app\params\definition_of_done.go",
  [string]$Tests = "app\params\definition_of_done_test.go"
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
    'Definition of Done',
    'No task is complete until',
    'code is implemented',
    'params are validated',
    'genesis import/export works',
    'query surface exists',
    'events exist where operationally relevant',
    'unit tests pass',
    'integration tests pass',
    'e2e/localnet test exists for user-facing flow',
    'docs describe operator/user impact',
    'failure modes are documented',
    'security implications are reviewed',
    'For consensus/economics/staking changes, also required',
    'adversarial tests',
    'invariant tests',
    'export/import test',
    'deterministic restart test',
    'migration test if state changed',
    'DefaultAetraDefinitionOfDoneEvidence',
    'Every feature must be reviewed as a full delivery unit',
    'Security review must be proportional to risk'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "definition of done doc missing: $term"
}

foreach ($term in @(
    'AetraDoDRequirementCodeImplemented',
    'AetraDoDRequirementParamsValidated',
    'AetraDoDRequirementGenesisImportExport',
    'AetraDoDRequirementQuerySurface',
    'AetraDoDRequirementOperationalEvents',
    'AetraDoDRequirementUnitTests',
    'AetraDoDRequirementIntegrationTests',
    'AetraDoDRequirementE2ELocalnetUserFlow',
    'AetraDoDRequirementOperatorUserDocs',
    'AetraDoDRequirementFailureModesDocumented',
    'AetraDoDRequirementSecurityReviewed',
    'AetraDoDCriticalRequirementAdversarialTests',
    'AetraDoDCriticalRequirementInvariantTests',
    'AetraDoDCriticalRequirementExportImportTest',
    'AetraDoDCriticalRequirementDeterministicRestart',
    'AetraDoDCriticalRequirementMigrationIfState',
    'AetraDefinitionOfDoneEvidence',
    'AetraDefinitionOfDoneReport',
    'DefaultAetraDefinitionOfDoneEvidence',
    'ValidateAetraDefinitionOfDone',
    'BuildAetraDefinitionOfDoneReport',
    'RequiredAetraDoDRequirements',
    'RequiredAetraDoDCriticalRequirements'
  )) {
  Assert-Contains -Text $catalogText -Pattern ([regex]::Escape($term)) -Message "definition of done catalog missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraDefinitionOfDoneCoversSection33BaseTask',
    'TestDefaultAetraDefinitionOfDoneCoversConsensusEconomicsStakingTask',
    'TestAetraDefinitionOfDoneRejectsMissingBaseRequirements',
    'TestAetraDefinitionOfDoneRejectsMissingCriticalRequirements'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "definition of done tests missing: $term"
}

Write-Host "definition of done doc test passed"
