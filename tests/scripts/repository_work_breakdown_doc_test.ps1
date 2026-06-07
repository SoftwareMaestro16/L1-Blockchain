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
    '32.3 `app/`',
    'wire keepers',
    'wire modules',
    'wire module account permissions',
    'wire begin/end/preblock order',
    'wire simulation manager if used',
    'wire API routes',
    'wire AutoCLI if used',
    'validate startup',
    'app startup',
    'module account permissions',
    'begin/end order',
    'export/import',
    'deterministic restart',
    'API service registration',
    'The `app/` tree owns whole-chain assembly',
    'Keeper wiring must pass the exact keeper dependencies used by modules',
    'Module account permissions are consensus-sensitive',
    'Begin/end/preblock order must be explicit',
    'Startup validation must reject unsafe module account permissions',
    '32.4 `tests/`',
    'integration test suites',
    'e2e localnet smoke tests',
    'adversarial tests',
    'load profile tests',
    'documentation path tests',
    'CI scripts',
    'tests must be runnable from documented commands',
    'Windows PowerShell local scripts should remain usable if current project supports them',
    'Linux CI path should remain primary for production confidence',
    'The `tests/` tree owns cross-module confidence',
    'Integration suites must prove keeper, app, module, governance, staking, economics, slashing, contract, and migration behavior',
    'E2E localnet smoke tests must prove that a node can start',
    'Adversarial tests must cover malformed txs',
    'Load profile tests must measure block time',
    'Documentation path tests must keep architecture docs',
    'DefaultAetraRepoTestsWorkEvidence',
    'DefaultAetraRepoAppWorkEvidence',
    'DefaultAetraRepoXWorkEvidence',
    'DefaultAetraRepoProtoWorkEvidence'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "repository work breakdown doc missing: $term"
}

foreach ($term in @(
    'AetraRepoAreaProto',
    'AetraRepoAreaX',
    'AetraRepoAreaApp',
    'AetraRepoAreaTests',
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
    'AetraRepoAppTaskWireKeepers',
    'AetraRepoAppTaskWireModules',
    'AetraRepoAppTaskWireModuleAccountPermissions',
    'AetraRepoAppTaskWireBeginEndPreblockOrder',
    'AetraRepoAppTaskWireSimulationManager',
    'AetraRepoAppTaskWireAPIRoutes',
    'AetraRepoAppTaskWireAutoCLI',
    'AetraRepoAppTaskValidateStartup',
    'AetraRepoAppTestStartup',
    'AetraRepoAppTestModuleAccountPermissions',
    'AetraRepoAppTestBeginEndOrder',
    'AetraRepoAppTestExportImport',
    'AetraRepoAppTestDeterministicRestart',
    'AetraRepoAppTestAPIServiceRegistration',
    'AetraRepoTestsTaskIntegrationSuites',
    'AetraRepoTestsTaskE2ELocalnetSmoke',
    'AetraRepoTestsTaskAdversarial',
    'AetraRepoTestsTaskLoadProfiles',
    'AetraRepoTestsTaskDocumentationPath',
    'AetraRepoTestsTaskCIScripts',
    'AetraRepoTestsRequirementDocumentedCommands',
    'AetraRepoTestsRequirementWindowsPowerShell',
    'AetraRepoTestsRequirementLinuxCIPrimary',
    'AetraRepoWorkAreaEvidence',
    'AetraRepoWorkAreaReport',
    'DefaultAetraRepoProtoWorkEvidence',
    'DefaultAetraRepoXWorkEvidence',
    'DefaultAetraRepoAppWorkEvidence',
    'DefaultAetraRepoTestsWorkEvidence',
    'ValidateAetraRepoProtoWork',
    'ValidateAetraRepoXWork',
    'ValidateAetraRepoAppWork',
    'ValidateAetraRepoTestsWork',
    'BuildAetraRepoProtoWorkReport',
    'BuildAetraRepoXWorkReport',
    'BuildAetraRepoAppWorkReport',
    'BuildAetraRepoTestsWorkReport',
    'RequiredAetraRepoProtoTasks',
    'RequiredAetraRepoProtoTests',
    'RequiredAetraRepoXTasks',
    'RequiredAetraRepoXTests',
    'RequiredAetraRepoAppTasks',
    'RequiredAetraRepoAppTests',
    'RequiredAetraRepoTestsTasks',
    'RequiredAetraRepoTestsRequirements'
  )) {
  Assert-Contains -Text $catalogText -Pattern ([regex]::Escape($term)) -Message "repository work breakdown catalog missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraRepoProtoWorkCoversSection321',
    'TestAetraRepoProtoWorkRejectsMissingTasksAndTests',
    'TestDefaultAetraRepoXWorkCoversSection322',
    'TestAetraRepoXWorkRejectsMissingTasksAndTests',
    'TestDefaultAetraRepoAppWorkCoversSection323',
    'TestAetraRepoAppWorkRejectsMissingTasksAndTests',
    'TestDefaultAetraRepoTestsWorkCoversSection324',
    'TestAetraRepoTestsWorkRejectsMissingTasksAndRequirements'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "repository work breakdown tests missing: $term"
}

Write-Host "repository work breakdown doc test passed"
