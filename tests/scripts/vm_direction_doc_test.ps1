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
    'CosmWasm remains the gated near-term VM candidate',
    'CosmWasm is disabled by default',
    'enabled only by explicit config or feature gate',
    'Upload permissions are explicit',
    'Instantiate permissions are explicit',
    'Admin and migration policy is explicit',
    'Gas limits are explicit and bounded',
    'Contract size limits are explicit and bounded',
    'Memory/cache limits are explicit and bounded',
    'Query limits are explicit and bounded',
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
    'Contract standards can be tested independent of the VM choice'
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
    'CosmWasm is the near-term gated VM candidate',
    'CosmWasm remains disabled by default',
    'Pinned code is disabled by default',
    'CosmWasm cannot bypass `naet` fee policy',
    'future Aetra VM requires a written binary serialization spec',
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
