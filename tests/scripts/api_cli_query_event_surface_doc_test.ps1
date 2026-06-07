param(
  [string]$Doc = "docs\architecture\api-cli-query-event-surface.md",
  [string]$Catalog = "observability\api_surface.go",
  [string]$Tests = "observability\api_surface_test.go"
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
    'API, CLI, Query, and Event Surface',
    'Every Aetra module must expose enough surface for validators, wallets, explorers, dashboards, monitoring, governance tooling, and incident response',
    '30.1 CLI Requirements',
    'aetrad query aetra-staking-policy ...',
    'aetrad query aetra-economics ...',
    'aetrad query aetra-validator-score ...',
    'aetrad tx aetra-staking-policy ...',
    'aetrad tx aetra-economics ...',
    'aetrad tx aetra-validator-score ...',
    'json output',
    'height query where applicable',
    'pagination where applicable',
    'clear errors',
    'examples in docs',
    '30.2 gRPC/REST Requirements',
    'Every query must have',
    'protobuf definition',
    'gRPC service',
    'REST gateway mapping if project supports it',
    'response examples',
    'tests where feasible',
    'The protobuf definition is the canonical contract for machine clients',
    'The gRPC service is the primary typed query API',
    'CLI query commands should call the same query service path',
    'CLI query commands',
    'CLI tx commands',
    'protobuf query definitions',
    'gRPC query services',
    'REST gateway routes where applicable',
    'query tests where feasible',
    'deterministic event names and bounded attributes',
    'DefaultAPISurfaceModuleSpecs'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "api surface doc missing: $term"
}

foreach ($term in @(
    'RequiredAPIModuleStakingPolicy',
    'RequiredAPIModuleEconomics',
    'RequiredAPIModuleValidatorScore',
    'CommandCategoryQuery',
    'CommandCategoryTx',
    'RequiredAPISurfaceCLIQuery',
    'RequiredAPISurfaceCLITx',
    'RequiredAPISurfaceProtobuf',
    'RequiredAPISurfaceGRPCService',
    'RequiredAPISurfaceGRPCQuery',
    'RequiredAPISurfaceRESTGateway',
    'RequiredAPISurfaceRESTQuery',
    'RequiredAPISurfaceEvents',
    'RequiredAPISurfaceResponseExample',
    'RequiredAPISurfaceQueryTests',
    'RequiredAPISurfaceExamplesInDocs',
    'RequiredAPISurfaceJSONOutput',
    'RequiredAPISurfaceClearErrors',
    'RequiredAPISurfaceHeightQuery',
    'RequiredAPISurfacePagination',
    'APISurfaceModuleSpec',
    'CLICommandSpec',
    'APISurfaceReadinessReport',
    'DefaultAPISurfaceModuleSpecs',
    'ValidateAPISurfaceReadiness',
    'BuildAPISurfaceReadinessReport'
  )) {
  Assert-Contains -Text $catalogText -Pattern ([regex]::Escape($term)) -Message "api surface catalog missing: $term"
}

foreach ($term in @(
    'TestDefaultAPISurfaceCoversSection30RequiredModules',
    'TestAPISurfaceRequiresQueryAndTxCommands',
    'TestAPISurfaceRejectsMissingCLIBehavior',
    'TestAPISurfaceRejectsMissingTxValidation',
    'TestAPISurfaceRejectsMissingGRPCRestEventsAndDocs',
    'TestAPISurfaceSection302RequiresProtoGrpcRestExamplesAndTests',
    'TestAPISurfaceRejectsMissingRequiredModule'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "api surface tests missing: $term"
}

Write-Host "api cli query event surface doc test passed"
