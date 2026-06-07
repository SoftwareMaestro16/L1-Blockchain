param(
  [string]$Doc = "docs\architecture\aetra-modular-execution-os.md",
  [string]$Readme = "README.md",
  [string]$Sharding = "docs\architecture\sharding-rd.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$ReadmePath = if ([System.IO.Path]::IsPathRooted($Readme)) { $Readme } else { Join-Path $RepoRoot $Readme }
$ShardingPath = if ([System.IO.Path]::IsPathRooted($Sharding)) { $Sharding } else { Join-Path $RepoRoot $Sharding }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

if (-not (Test-Path -LiteralPath $DocPath)) {
  throw "architecture design doc missing: $DocPath"
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$readmeText = Get-Content -Raw -LiteralPath $ReadmePath
$shardingText = Get-Content -Raw -LiteralPath $ShardingPath

foreach ($term in @(
    'Aetra Modular L1 Execution OS',
    'Architecture Diagram',
    'Layer-By-Layer Design',
    'Aether Core',
    'Execution Zones',
    'Compute Shards',
    'Deterministic Load Detection',
    'Load Lifecycle',
    'Routing Algorithm Specification',
    'Aether Mesh',
    'Identity Layer',
    'Security Model Summary',
    'Failure Scenarios And Mitigations',
    'Trilemma Model',
    'Aether Core must not execute smart contracts',
    'Aether Core must not process application-specific business logic',
    'Financial Zone',
    'Identity Zone',
    'Application Zone',
    'Contract Zone',
    'local mempool policy',
    'local fee market parameters',
    'state_root',
    'receipt_root',
    'message_root',
    'shard_id = hash(zone_id || primary_actor || routing_epoch) % active_shards(zone_id)',
    'N = 60 blocks',
    'alpha = 2 / (N + 1)',
    'MAX_DELTA = 0.05',
    'EMA_t(metric) = alpha * value_t + (1 - alpha) * EMA_{t-1}(metric)',
    'LOAD_SCORE = 0.20*mempool_size_score + 0.30*block_utilization_score + 0.20*tx_latency_score + 0.10*failure_rate_score + 0.20*execution_time_score',
    '0.0 <= LOAD_SCORE < 0.3',
    '0.3 <= LOAD_SCORE < 0.7',
    '0.7 <= LOAD_SCORE <= 1.0',
    'critical system transactions',
    'high fee transactions',
    'normal user transactions',
    'low fee or spam-like transactions',
    'cross-zone messaging',
    'single-use receipt markers',
    '.aet',
    'NFT-based ownership model',
    'available -> commit -> reveal/register -> active -> renewal window -> expired -> available',
    'PoS staking',
    'deterministic evidence',
    'low base fees',
    'dynamic load response',
    'No double spend'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "modular execution OS doc missing: $term"
}

$finalStatement = 'Aetra is a modular Layer 1 execution operating system where consensus, execution, and scaling are separated into independent but cryptographically connected layers to achieve secure and scalable decentralized computation.'
if ($docText.TrimEnd() -notlike "*$finalStatement") {
  throw "modular execution OS doc must end with the required final statement"
}

Assert-Contains `
  -Text $readmeText `
  -Pattern ([regex]::Escape('docs/architecture/aetra-modular-execution-os.md')) `
  -Message "README must link to modular execution OS doc"

Assert-Contains `
  -Text $shardingText `
  -Pattern ([regex]::Escape('aetra-modular-execution-os.md')) `
  -Message "sharding R&D doc must cross-link modular execution OS doc"

Write-Host "aetra modular execution OS doc test passed"
