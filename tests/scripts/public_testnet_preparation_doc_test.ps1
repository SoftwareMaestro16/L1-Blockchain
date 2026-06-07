param(
  [string]$Runbook = "docs\public-testnet-preparation.md",
  [string]$Validator = "docs\validator-onboarding.md",
  [string]$Incident = "docs\testnet-incident-response.md",
  [string]$Preflight = "scripts\testnet\public-testnet-preflight.ps1",
  [string]$Start = "scripts\localnet\start.ps1",
  [string]$Ports = "scripts\localnet\lib\ports.ps1",
  [string]$Paths = "scripts\localnet\lib\paths.ps1",
  [string]$AdversarialWorkflow = ".github\workflows\adversarial-e2e.yml",
  [string]$PrototypeWorkflow = ".github\workflows\prototype-release.yml"
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

function Assert-NotContains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -match $Pattern) { throw $Message }
}

$runbookText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Runbook)
$validatorText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Validator)
$incidentText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Incident)
$preflightText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Preflight)
$startText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Start)
$portsText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Ports)
$pathsText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Paths)
$adversarialText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $AdversarialWorkflow)
$prototypeText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $PrototypeWorkflow)

foreach ($term in @(
    "Phase 16",
    "ValidatorProfile All",
    "ValidatorProfile 3",
    "ValidatorProfile 5",
    "ValidatorProfile 10",
    "10-validator profile is the stress profile",
    "Localnet Hardening",
    "Faucet Plan",
    "Explorer And Indexer Plan",
    "Minimum Hardware",
    "Snapshot And State-Sync Plan",
    "CosmWasm Test Contract",
    "async contracts are enabled",
    "Rollback And Restart Procedure",
    "Launch Checklist",
    "go test -p=1 ./...",
    "go vet -p=1 ./...",
    "buf lint",
    "1000000naet",
    "aetrad"
  )) {
  Assert-Contains -Text $runbookText -Pattern ([regex]::Escape($term)) -Message "public testnet runbook missing: $term"
}

foreach ($term in @(
    "Validator Onboarding",
    "scripts\build-aetrad.ps1",
    ".aetra",
    "100000000naet",
    "1000000naet",
    "catching_up",
    "CosmWasm Contract Smoke"
  )) {
  Assert-Contains -Text $validatorText -Pattern ([regex]::Escape($term)) -Message "validator onboarding missing: $term"
}

foreach ($term in @(
    "Testnet Incident Response",
    "Critical",
    "Consensus Halt",
    "Faucet Incident",
    "Snapshot Or State-Sync Incident",
    "Communication",
    "aetrad"
  )) {
  Assert-Contains -Text $incidentText -Pattern ([regex]::Escape($term)) -Message "incident runbook missing: $term"
}

foreach ($term in @(
    'ValidateSet("3", "5", "10", "All")',
    '@(3, 5, 10)',
    "aetra-testnet-preflight-1",
    "build\aetrad.exe",
    "scripts\build-aetrad.ps1",
    "ValidatorCount `$validators",
    "cosmwasm_smoke.ps1"
  )) {
  Assert-Contains -Text $preflightText -Pattern ([regex]::Escape($term)) -Message "public testnet preflight missing: $term"
}

foreach ($term in @(
    "Assert-LocalnetWorkspacePath",
    "Assert-LocalnetPortsAvailable",
    "startup-timing.json",
    "aetrad"
  )) {
  Assert-Contains -Text $startText -Pattern ([regex]::Escape($term)) -Message "localnet start missing hardening/timing term: $term"
}

foreach ($term in @(
    'NodeDaemonHome = "aetrad"',
    'MinimumGasPrices = "0naet"',
    'node$i\aetrad'
  )) {
  Assert-Contains -Text $portsText -Pattern ([regex]::Escape($term)) -Message "localnet ports helper missing Aetra term: $term"
}

Assert-Contains -Text $pathsText -Pattern ([regex]::Escape('node$Index\aetrad')) -Message "localnet paths helper must use aetrad homes"

foreach ($workflowText in @($adversarialText, $prototypeText)) {
  Assert-Contains -Text $workflowText -Pattern "aetrad|Aetra|AETRA_BINARY" -Message "workflow missing Aetra runtime naming"
}

$oldNamingPattern = "Orbitalis|orbitalisd|ORBITALIS|norb|ORB\b|orbitalis-local-1|\.orbitalis"
foreach ($entry in @(
    @{ Name = $Runbook; Text = $runbookText },
    @{ Name = $Validator; Text = $validatorText },
    @{ Name = $Incident; Text = $incidentText },
    @{ Name = $Preflight; Text = $preflightText },
    @{ Name = $Start; Text = $startText },
    @{ Name = $Ports; Text = $portsText },
    @{ Name = $Paths; Text = $pathsText },
    @{ Name = $AdversarialWorkflow; Text = $adversarialText },
    @{ Name = $PrototypeWorkflow; Text = $prototypeText }
  )) {
  Assert-NotContains -Text $entry.Text -Pattern $oldNamingPattern -Message "$($entry.Name) contains old runtime naming"
}

Write-Host "public testnet preparation doc test passed"
