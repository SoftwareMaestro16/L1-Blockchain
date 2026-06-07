param(
  [string]$Gates = "docs\public-testnet-production-gates.md",
  [string]$Preparation = "docs\public-testnet-preparation.md",
  [string]$SecurityGate = "docs\security\prototype-audit-gate.md",
  [string]$Preflight = "scripts\testnet\public-testnet-preflight.ps1"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Path))
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$gatesText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Gates)
$prepText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Preparation)
$securityText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $SecurityGate)
$preflightText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Preflight)

foreach ($term in @(
    'Public Testnet And Production Gates',
    'Test And Production Gates',
    'test-production-gates.md',
    'go test ./...',
    'go vet ./...',
    'buf lint',
    'govulncheck',
    'gosec',
    'CodeQL',
    'gitleaks',
    'dependency review',
    'determinism-gate.ps1',
    'ValidatorProfile 3',
    'ValidatorProfile 5',
    'ValidatorProfile 10',
    'ValidatorProfile All',
    'Snapshot and state-sync',
    'Validator onboarding',
    'Faucet plan is implemented or explicitly deferred',
    'Explorer/indexer plan is implemented or explicitly deferred',
    'Incident response and rollback docs are tested',
    'cosmwasm_smoke.ps1 -EnableWasm',
    'go test ./x/aetravm/avm ./x/aetravm/async',
    'go test ./x/aetravm/standards/...',
    'untriaged `Critical` or `High`',
    'Production Gate',
    'long-running public testnet',
    'Validator set can upgrade safely',
    'Staking, fees, DEX, AVM, and contract standards have adversarial tests',
    'State export/import is deterministic',
    'Independent audit is completed',
    'Emergency governance and halt/restart process is tested',
    'sharding R&D',
    'Immediate Build Order',
    'Finish base-chain safety',
    'Finish PoS/staking production hardening',
    'Build deterministic async queue without AVM first',
    'Build minimal AVM with a counter contract',
    'Implement AW-5 wallet',
    'Implement AFT-44 token master/wallet',
    'Implement ANFT-66 NFT collection/item',
    'Implement ASBT-67 soulbound item',
    'Gate CosmWasm behind explicit config and tests',
    'Start sharding simulator and spec',
    'prototype masterchain/workchain/shardchain',
    'Phase 12 Modular Execution Public Testnet Gate',
    'Base chain hardening complete',
    'Determinism gate passes',
    'Export/import gate passes',
    'Genesis migration gate passes',
    '3-validator localnet long-run passes',
    '5-validator localnet long-run passes',
    '10-validator stress profile passes',
    'State-sync and snapshot restore pass',
    'Load score and routing simulator pass',
    'aetrad execution-os smoke --profile execution-os-sim',
    'Sharding simulator pass',
    'Mesh simulator pass',
    'Identity executable spec pass',
    'VM readiness tests pass',
    'Independent audit findings are triaged',
    'Public docs do not overclaim production sharding',
    'Public testnet can advertise only the features that passed their gate',
    'Any feature still behind R&D remains documented as experimental',
    'Suggested branch order',
    'feature/load-score-spec',
    'feature/routing-engine-spec',
    'feature/zone-registry-sim',
    'feature/load-driven-sharding-sim',
    'feature/aether-mesh-sim',
    'feature/identity-zone-spec',
    'feature/contract-zone-readiness',
    'prototype/execution-os-keepers',
    'prototype/aether-core-routing-wiring',
    'tooling/execution-os-localnet',
    'security/execution-os-invariants',
    'testnet/modular-execution-gate',
    'powershell -NoProfile -ExecutionPolicy Bypass -File tests\scripts\determinism_gate_test.ps1',
    'Traceability matrix',
    'Aether Core as control plane',
    'Execution Zones',
    'Compute Shards',
    'Deterministic `LOAD_SCORE`',
    'Aether Mesh',
    '`.aet` Identity Layer',
    'Trilemma claim support',
    'Do not wire production Execution Zones',
    'production contract execution into Aether Core'
  )) {
  Assert-Contains -Text $gatesText -Pattern ([regex]::Escape($term)) -Message "public testnet/production gate doc missing: $term"
}

foreach ($term in @(
    'Launch Checklist',
    'Faucet Plan',
    'Explorer And Indexer Plan',
    'Snapshot And State-Sync Plan',
    'Rollback And Restart Procedure'
  )) {
  Assert-Contains -Text $prepText -Pattern ([regex]::Escape($term)) -Message "public testnet prep missing linked gate term: $term"
}

foreach ($term in @(
    'Public testnet cannot proceed',
    'untriaged High/Critical',
    'fund-safety',
    'consensus-safety',
    'secret-leak'
  )) {
  Assert-Contains -Text $securityText -Pattern ([regex]::Escape($term)) -Message "security gate missing production/public gate term: $term"
}

foreach ($term in @(
    'ValidateSet("3", "5", "10", "All")',
    '@(3, 5, 10)',
    'prototype_acceptance.ps1',
    'cosmwasm_smoke.ps1'
  )) {
  Assert-Contains -Text $preflightText -Pattern ([regex]::Escape($term)) -Message "preflight missing gate term: $term"
}

Write-Host "public testnet production gates doc test passed"
