param(
  [string]$Direction = "docs\architecture\vm-direction.md",
  [string]$CosmWasm = "docs\security\cosmwasm-readiness.md",
  [string]$Boundaries = "docs\module-boundaries.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$DirectionPath = if ([System.IO.Path]::IsPathRooted($Direction)) { $Direction } else { Join-Path $RepoRoot $Direction }
$CosmWasmPath = if ([System.IO.Path]::IsPathRooted($CosmWasm)) { $CosmWasm } else { Join-Path $RepoRoot $CosmWasm }
$BoundariesPath = if ([System.IO.Path]::IsPathRooted($Boundaries)) { $Boundaries } else { Join-Path $RepoRoot $Boundaries }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$directionText = Get-Content -Raw -LiteralPath $DirectionPath
$cosmWasmText = Get-Content -Raw -LiteralPath $CosmWasmPath
$boundariesText = Get-Content -Raw -LiteralPath $BoundariesPath

foreach ($term in @(
    'VM Direction',
    'Aetra uses AVM as the genesis smart contract runtime',
    'AVM is the primary VM',
    'AVM production enablement requires complete security review',
    'complete gas model review',
    'state growth benchmarks',
    'interaction tests with fee burn',
    'interaction tests with staking rewards',
    'export/import tests',
    'adversarial contract tests',
    'CosmWasm remains an explicitly gated optional compatibility layer',
    'CosmWasm is disabled by default',
    'enabled only by explicit config or feature gate',
    'Instantiate, execute, query, and migrate support',
    'Upload permissions are explicit',
    'Instantiate permissions are explicit',
    'Admin and migration policy is explicit',
    'Gas limits are explicit and bounded',
    'Contract size limits are explicit and bounded',
    'Contract storage rent or storage pricing is explicit',
    'Memory/cache limits are explicit and bounded',
    'Query limits are explicit and bounded',
    'Events must be compatible with indexers',
    'Localnet smoke tests are required',
    'Malicious contract tests are required',
    'Export/import tests with active contracts are required',
    'Pinned code is disabled by default',
    'Governance authority for enabling/disabling CosmWasm is explicit',
    'cannot bypass `naet` fee policy',
    'binary serialization spec',
    'message ABI',
    'storage ABI',
    'gas schedule',
    'deterministic execution proof',
    'fuzz tests',
    'upgrade/migration policy',
    'adversarial audit',
    'explicit storage schema',
    'explicit inbound messages',
    'explicit outbound messages',
    'explicit getters',
    'explicit unknown-message policy',
    'explicit bounce behavior',
    'explicit fee behavior',
    'explicit deployment behavior',
    'Contract standards can be tested independent of optional compatibility layers'
  )) {
  Assert-Contains -Text $directionText -Pattern ([regex]::Escape($term)) -Message "VM direction doc missing: $term"
}

foreach ($term in @(
    'CosmWasm stays disabled in Aetra',
    'wasm.enabled = false',
    'no `wasm` store key in the app',
    'governance-only',
    'code-owner-only',
    'Migration is admin-only',
    'smart query gas limit',
    'simulation gas limit',
    'memory cache',
    'smart query response limit',
    'smart query depth limit',
    'Pinned code',
    'Only the configured governance authority can enable or disable the gate',
    'non-`naet` fee rejection',
    'zero admin, zero recipient, and zero contract address rejection',
    'Query limits are part of the feature gate',
    'aetrad',
    'naet'
  )) {
  Assert-Contains -Text $cosmWasmText -Pattern ([regex]::Escape($term)) -Message "CosmWasm readiness doc missing: $term"
}

foreach ($term in @(
    'app/wasmconfig',
    'AVM is the primary VM',
    'CosmWasm is an optional gated compatibility layer',
    'CosmWasm remains disabled by default',
    'Pinned code is disabled by default',
    'CosmWasm cannot bypass `naet` fee policy',
    'future non-AVM runtime requires a written binary serialization spec',
    'Contract standards must remain testable independent'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing VM direction boundary: $term"
}

foreach ($standard in @(
    'docs\standards\aft-44.md',
    'docs\standards\anft-66-asbt-67.md',
    'docs\standards\aw-5.md'
  )) {
  $standardPath = Join-Path $RepoRoot $standard
  $standardText = Get-Content -Raw -LiteralPath $standardPath
  foreach ($term in @(
      'VM-Independent Contract Requirements',
      'Explicit storage schema',
      'Explicit inbound messages',
      'Explicit outbound messages',
      'Explicit getters',
      'Explicit unknown-message policy',
      'Explicit bounce behavior',
      'Explicit fee behavior',
      'Explicit deployment behavior'
    )) {
    Assert-Contains -Text $standardText -Pattern ([regex]::Escape($term)) -Message "$standard missing contract requirement: $term"
  }
}

Write-Host "vm direction doc test passed"
