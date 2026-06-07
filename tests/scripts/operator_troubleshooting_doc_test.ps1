param(
  [string]$Runbook = "docs\operator-troubleshooting.md",
  [string]$Readme = "README.md",
  [string]$OperatorGuide = "docs\operator-commands.md",
  [string]$Observability = "docs\observability.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$RunbookPath = if ([System.IO.Path]::IsPathRooted($Runbook)) { $Runbook } else { Join-Path $RepoRoot $Runbook }
$ReadmePath = if ([System.IO.Path]::IsPathRooted($Readme)) { $Readme } else { Join-Path $RepoRoot $Readme }
$OperatorGuidePath = if ([System.IO.Path]::IsPathRooted($OperatorGuide)) { $OperatorGuide } else { Join-Path $RepoRoot $OperatorGuide }
$ObservabilityPath = if ([System.IO.Path]::IsPathRooted($Observability)) { $Observability } else { Join-Path $RepoRoot $Observability }

function Assert-Contains {
  param(
    [string]$Text,
    [string]$Pattern,
    [string]$Message
  )
  if ($Text -notmatch $Pattern) {
    throw $Message
  }
}

function Assert-NotContains {
  param(
    [string]$Text,
    [string]$Pattern,
    [string]$Message
  )
  if ($Text -match $Pattern) {
    throw $Message
  }
}

if (!(Test-Path -LiteralPath $RunbookPath)) {
  throw "missing operator troubleshooting runbook"
}

$runbookText = Get-Content -Raw -LiteralPath $RunbookPath
$readmeText = Get-Content -Raw -LiteralPath $ReadmePath
$operatorText = Get-Content -Raw -LiteralPath $OperatorGuidePath
$observabilityText = Get-Content -Raw -LiteralPath $ObservabilityPath

foreach ($heading in @(
    "First Triage",
    "Common Failures",
    "Port Profiles",
    "Evidence And Links"
  )) {
  Assert-Contains -Text $runbookText -Pattern "## $([regex]::Escape($heading))" -Message "runbook missing heading: $heading"
}

foreach ($failure in @(
    "Binary missing",
    "Port in use",
    "No blocks",
    "No peers",
    "Wrong fee denom",
    "Insufficient funds",
    "Sequence mismatch",
    "DEX slippage",
    "Unauthorized contract-assets",
    "REST down"
  )) {
  Assert-Contains -Text $runbookText -Pattern $failure -Message "runbook missing common failure: $failure"
}

foreach ($command in @(
    '\.\\scripts\\build-aetrad\.ps1',
    '\.\\scripts\\localnet\\health\.ps1',
    '\.\\scripts\\localnet\\diagnostics\.ps1',
    '\.\\scripts\\localnet\\stop\.ps1',
    '\.\\scripts\\localnet\\reset\.ps1',
    '\.\\scripts\\localnet\\fund\.ps1',
    'build\\aetrad\.exe query fees params',
    'build\\aetrad\.exe query bank balance',
    'build\\aetrad\.exe query auth account',
    'build\\aetrad\.exe query dex pool',
    'build\\aetrad\.exe query contract-assets denom',
    'Invoke-RestMethod'
  )) {
  Assert-Contains -Text $runbookText -Pattern $command -Message "runbook missing troubleshooting command: $command"
}

foreach ($safety in @(
    'Never paste mnemonics',
    'redact',
    'exclude keyrings',
    'Destructive reset',
    '--keyring-backend test',
    'local prototype',
    '3-validator',
    '5-validator',
    '-BaseRPCPort',
    '-BaseRESTPort',
    '-BaseGRPCPort'
  )) {
  Assert-Contains -Text $runbookText -Pattern $safety -Message "runbook missing safety or scale guidance: $safety"
}

foreach ($evidence in @(
    'localnet_smoke\.ps1',
    'fees_ante_smoke\.ps1',
    'mempool_negative_smoke\.ps1',
    'dex_smoke\.ps1',
    'observability_scripts_test\.ps1'
  )) {
  Assert-Contains -Text $runbookText -Pattern $evidence -Message "runbook missing evidence link: $evidence"
}

Assert-NotContains -Text $runbookText -Pattern '(?i)--print-mnemonic|print-mnemonic|mnemonic:\s*[a-z]|private[_-]?key\s*[:=]' -Message "runbook must not tell operators to expose secrets"
Assert-NotContains -Text $runbookText -Pattern '(?i)redis://|postgresql://|password\s*=' -Message "runbook must not include environment secrets"

Assert-Contains -Text $readmeText -Pattern 'docs/operator-troubleshooting\.md' -Message "README must link troubleshooting runbook"
Assert-Contains -Text $operatorText -Pattern 'operator-troubleshooting\.md' -Message "operator guide must link troubleshooting runbook"
Assert-Contains -Text $observabilityText -Pattern 'operator-troubleshooting\.md' -Message "observability doc must link troubleshooting runbook"

Write-Host "operator troubleshooting doc test passed"
