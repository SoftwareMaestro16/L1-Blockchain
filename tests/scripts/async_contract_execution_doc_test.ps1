param(
  [string]$Doc = "docs\architecture\async-smart-contract-execution.md",
  [string]$Boundaries = "docs\module-boundaries.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DocPath = if ([System.IO.Path]::IsPathRooted($Doc)) { $Doc } else { Join-Path $RepoRoot $Doc }
$BoundariesPath = if ([System.IO.Path]::IsPathRooted($Boundaries)) { $Boundaries } else { Join-Path $RepoRoot $Boundaries }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath $DocPath
$boundariesText = Get-Content -Raw -LiteralPath $BoundariesPath

foreach ($term in @(
    'Async Smart Contract Execution',
    'Cosmos SDK rule that transaction delivery is synchronous',
    'Production partitioning or sharding remains a later R&D track',
    'contract account',
    'incoming message',
    'outgoing message',
    'message queue',
    'bounce message',
    'logical time',
    'address = sha256',
    'source',
    'destination',
    'value in `naet`',
    'opcode',
    'query_id',
    'body',
    'bounce flag',
    'created logical time',
    'expiration/deadline block',
    'gas limit',
    'forward fee in `naet`',
    'tx index',
    'message index',
    'source logical time',
    'destination address key',
    'sequence tie-breaker',
    'next_tx_index',
    'source logical time drift',
    'destination key drift',
    'Contract Logical Time',
    'recipient mutation is not committed',
    'does not create another refund',
    'execution gas per message',
    'per-message gas limit',
    'storage fee per committed state byte',
    'message forwarding fee',
    'per-message forward fee in `naet`',
    'contract deployment cost',
    'max messages per tx',
    'max messages per block',
    'max recursion/depth',
    'max body size',
    'max state size',
    'max emitted messages per execution',
    'max storage writes per execution',
    'queued messages',
    'processed messages',
    'bounced messages',
    'failed executions',
    'queue lag',
    'contract state size',
    'x/aetravm/async',
    'export/import state shape that preserves queue state exactly'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "async execution doc missing: $term"
}

foreach ($term in @(
    'x/aetravm/async',
    'deterministic asynchronous contract message semantics',
    'Cosmos SDK delivered transactions remain synchronous',
    'Production partitioning or sharding is a later R&D track',
    'Queue ordering must use tx index, message index, source logical time',
    'double-refund loops',
    'All protocol fee and message value accounting is native `naet` only',
    'Per-message gas limit and forward fee validation must be explicit',
    'Export/import must preserve queued messages'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing async boundary: $term"
}

Write-Host "async contract execution doc test passed"
