param(
  [string]$Doc = "docs\architecture\repository-work-breakdown.md",
  [string]$Catalog = "app\params\repository_work_breakdown.go",
  [string]$Tests = "app\params\repository_work_breakdown_test.go"
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
    'Repository-Level Work Breakdown',
    'This section maps work to likely repository areas',
    '32.1 `proto/`',
    'define protobuf messages for new modules',
    'define query services',
    'define tx services',
    'define genesis messages',
    'define params messages',
    'run code generation',
    'add proto breaking-change checks if available',
    'generated code compiles',
    'proto lint passes if configured',
    'query/tx service registration tested',
    'The `proto/` tree owns public wire contracts',
    'Code generation must be reproducible',
    'Service registration tests must prove that query and tx services are reachable',
    '32.2 `x/`',
    'implement keepers',
    'implement message servers',
    'implement query servers',
    'implement genesis',
    'implement params validation',
    'implement invariants',
    'implement hooks where needed',
    'implement events',
    'implement module interfaces',
    'keeper unit tests',
    'msg server tests',
    'query server tests',
    'genesis tests',
    'invariant tests',
    'fuzz/property tests for math',
    'The `x/` tree owns module behavior',
    'Keepers must be deterministic',
    'Message servers must validate signers',
    'Query servers must return stable response shapes',
    'Fuzz/property tests are required for math-heavy logic',
    'DefaultAetraRepoXWorkEvidence',
    'DefaultAetraRepoProtoWorkEvidence'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "repository work breakdown doc missing: $term"
}

foreach ($term in @(
    'AetraRepoAreaProto',
    'AetraRepoAreaX',
    'AetraRepoProtoTaskDefineMessages',
    'AetraRepoProtoTaskDefineQueryServices',
    'AetraRepoProtoTaskDefineTxServices',
    'AetraRepoProtoTaskDefineGenesis',
    'AetraRepoProtoTaskDefineParams',
    'AetraRepoProtoTaskRunCodeGeneration',
    'AetraRepoProtoTaskBreakingChangeChecks',
    'AetraRepoProtoTestGeneratedCodeCompiles',
    'AetraRepoProtoTestLintPasses',
    'AetraRepoProtoTestServiceRegistration',
    'AetraRepoXTaskImplementKeepers',
    'AetraRepoXTaskImplementMsgServers',
    'AetraRepoXTaskImplementQueryServers',
    'AetraRepoXTaskImplementGenesis',
    'AetraRepoXTaskImplementParamsValidation',
    'AetraRepoXTaskImplementInvariants',
    'AetraRepoXTaskImplementHooks',
    'AetraRepoXTaskImplementEvents',
    'AetraRepoXTaskImplementModuleInterfaces',
    'AetraRepoXTestKeeperUnit',
    'AetraRepoXTestMsgServer',
    'AetraRepoXTestQueryServer',
    'AetraRepoXTestGenesis',
    'AetraRepoXTestInvariant',
    'AetraRepoXTestFuzzPropertyMath',
    'AetraRepoWorkAreaEvidence',
    'AetraRepoWorkAreaReport',
    'DefaultAetraRepoProtoWorkEvidence',
    'DefaultAetraRepoXWorkEvidence',
    'ValidateAetraRepoProtoWork',
    'ValidateAetraRepoXWork',
    'BuildAetraRepoProtoWorkReport',
    'BuildAetraRepoXWorkReport',
    'RequiredAetraRepoProtoTasks',
    'RequiredAetraRepoProtoTests',
    'RequiredAetraRepoXTasks',
    'RequiredAetraRepoXTests'
  )) {
  Assert-Contains -Text $catalogText -Pattern ([regex]::Escape($term)) -Message "repository work breakdown catalog missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraRepoProtoWorkCoversSection321',
    'TestAetraRepoProtoWorkRejectsMissingTasksAndTests',
    'TestDefaultAetraRepoXWorkCoversSection322',
    'TestAetraRepoXWorkRejectsMissingTasksAndTests'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "repository work breakdown tests missing: $term"
}

Write-Host "repository work breakdown doc test passed"
