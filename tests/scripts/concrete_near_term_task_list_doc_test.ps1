param(
  [string]$Doc = "docs\architecture\concrete-near-term-task-list.md",
  [string]$Catalog = "app\params\near_term_task_list.go",
  [string]$Tests = "app\params\near_term_task_list_test.go"
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
    'Concrete Near-Term Task List',
    'Near-term implementation should be split into small pull requests',
    'Audit existing validator/economics modules and map them to this spec',
    'Add missing params validation for validator power cap and commission policy',
    'Add effective power and overflow stake queries',
    'Add concentration snapshot query',
    'Add cap math tests for 100/150/200/300 validator scenarios',
    'Add fee split accounting tests',
    'Add inflation bounds tests',
    'Add supply invariant tests',
    'Add validator score state and query tests',
    'Add progressive downtime design or document why standard slashing is enough for v1',
    'Add nomination pool accounting tests',
    'Add AVM smoke and malicious contract tests',
    'Add public testnet finality measurement script',
    'Add documentation for validators and delegators',
    'Add CI gate for critical unit/integration tests',
    'what consensus/economics behavior changes',
    'what params are added or changed',
    'what tests were added',
    'what migration risk exists',
    'whether public docs need updates',
    'Validator/economics audit comes first',
    'DefaultAetraNearTermTaskListEvidence'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "near-term task list doc missing: $term"
}

foreach ($term in @(
    'AetraNearTermTaskAuditExistingModules',
    'AetraNearTermTaskPowerCapCommissionParams',
    'AetraNearTermTaskEffectivePowerQueries',
    'AetraNearTermTaskConcentrationSnapshotQuery',
    'AetraNearTermTaskCapMathTests',
    'AetraNearTermTaskFeeSplitAccountingTests',
    'AetraNearTermTaskInflationBoundsTests',
    'AetraNearTermTaskSupplyInvariantTests',
    'AetraNearTermTaskValidatorScoreStateQueries',
    'AetraNearTermTaskProgressiveDowntimeDecision',
    'AetraNearTermTaskNominationPoolAccounting',
    'AetraNearTermTaskAVMSmokeMalicious',
    'AetraNearTermTaskFinalityMeasurementScript',
    'AetraNearTermTaskValidatorDelegatorDocs',
    'AetraNearTermTaskCriticalCIGate',
    'AetraNearTermChecklistConsensusEconomicsBehavior',
    'AetraNearTermChecklistParamsAddedChanged',
    'AetraNearTermChecklistTestsAdded',
    'AetraNearTermChecklistMigrationRisk',
    'AetraNearTermChecklistPublicDocs',
    'AetraNearTermTaskListEvidence',
    'AetraNearTermTaskListReport',
    'DefaultAetraNearTermTaskListEvidence',
    'ValidateAetraNearTermTaskList',
    'BuildAetraNearTermTaskListReport',
    'RequiredAetraNearTermTasks',
    'RequiredAetraNearTermChecklist'
  )) {
  Assert-Contains -Text $catalogText -Pattern ([regex]::Escape($term)) -Message "near-term task list catalog missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraNearTermTaskListCoversSection36',
    'TestAetraNearTermTaskListRejectsMissingDuplicateUnexpectedTasks'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "near-term task list tests missing: $term"
}

Write-Host "concrete near-term task list doc test passed"
