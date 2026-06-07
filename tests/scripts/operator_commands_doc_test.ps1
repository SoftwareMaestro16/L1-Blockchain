param(
  [string]$Guide = "docs\operator-commands.md",
  [string]$Readme = "README.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$GuidePath = if ([System.IO.Path]::IsPathRooted($Guide)) { $Guide } else { Join-Path $RepoRoot $Guide }
$ReadmePath = if ([System.IO.Path]::IsPathRooted($Readme)) { $Readme } else { Join-Path $RepoRoot $Readme }

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

$guideText = Get-Content -Raw -LiteralPath $GuidePath
$readmeText = Get-Content -Raw -LiteralPath $ReadmePath

foreach ($heading in @(
    "Build And Version",
    "Localnet",
    "Common Flags",
    "Queries",
    "Staking Tx",
    "Bank Tx",
    "Contract assets",
    "DEX",
    "Diagnose",
    "Troubleshooting",
    "Required Command Checks"
  )) {
  Assert-Contains -Text $guideText -Pattern "## $([regex]::Escape($heading))" -Message "operator guide missing heading: $heading"
}

foreach ($required in @(
    '\.\\scripts\\build-aetrad\.ps1',
    '\.\\scripts\\localnet\\init\.ps1',
    '\.\\scripts\\localnet\\validate-genesis\.ps1',
    '\.\\scripts\\localnet\\start\.ps1',
    '\.\\scripts\\localnet\\health\.ps1',
    '\.\\scripts\\localnet\\diagnostics\.ps1',
    '\.\\scripts\\localnet\\stop\.ps1',
    '\.\\scripts\\localnet\\reset\.ps1',
    '\.\\tests\\e2e\\prototype_smoke\.ps1',
    '\.\\tests\\e2e\\pos_smoke\.ps1',
    '\.\\tests\\e2e\\contract-assets_smoke\.ps1',
    'build\\aetrad\.exe version --long --output json',
    'build\\aetrad\.exe query block',
    'build\\aetrad\.exe query bank balance',
    'build\\aetrad\.exe query staking validators',
    'build\\aetrad\.exe query fees params',
    'build\\aetrad\.exe query contract-assets denom',
    'build\\aetrad\.exe query dex pool',
    'build\\aetrad\.exe tx bank send',
    'build\\aetrad\.exe tx staking delegate',
    'build\\aetrad\.exe tx contract-assets create-denom',
    'build\\aetrad\.exe tx contract-assets mint',
    'build\\aetrad\.exe tx contract-assets burn',
    'build\\aetrad\.exe tx contract-assets change-admin',
    'build\\aetrad\.exe tx dex create-pool',
    'build\\aetrad\.exe tx dex add-liquidity',
    'build\\aetrad\.exe tx dex swap-exact-in',
    'build\\aetrad\.exe tx dex remove-liquidity'
  )) {
  Assert-Contains -Text $guideText -Pattern $required -Message "operator guide missing command pattern: $required"
}

foreach ($variable in @('$CHAIN_ID', '$NODE', '$GRPC', '$REST', '$HOME', '$FROM', '$FEES', '$KEYRING')) {
  Assert-Contains -Text $guideText -Pattern ([regex]::Escape($variable)) -Message "operator guide missing reusable variable $variable"
}

Assert-Contains -Text $guideText -Pattern '--keyring-backend test.*local|local.*--keyring-backend test' -Message "operator guide must mark test keyring local-only"
Assert-Contains -Text $guideText -Pattern 'naet' -Message "operator guide must use naet examples"
Assert-Contains -Text $guideText -Pattern 'AET.*display metadata only|display metadata only.*AET' -Message "operator guide must clarify AET display-only status"
Assert-Contains -Text $guideText -Pattern 'observability\.md' -Message "operator guide must link troubleshooting observability docs"
Assert-Contains -Text $guideText -Pattern 'operator-troubleshooting\.md' -Message "operator guide must link troubleshooting runbook"
Assert-Contains -Text $guideText -Pattern 'prototype-audit\.ps1 -Profile Fast' -Message "operator guide must include security audit check"

Assert-NotContains -Text $guideText -Pattern '(?i)print-mnemonic\s+\$?true|--print-mnemonic|mnemonic:\s*[a-z]' -Message "operator guide must not tell operators to print mnemonics"
Assert-NotContains -Text $guideText -Pattern '(?i)private[_-]?key\s*[:=]' -Message "operator guide must not include private key material"
Assert-NotContains -Text $guideText -Pattern '1000000AET|--fees\s+\d+AET' -Message "operator guide must not use AET as tx fee denom"

Assert-Contains -Text $readmeText -Pattern 'docs/operator-commands\.md' -Message "README must link operator guide"
Assert-Contains -Text $readmeText -Pattern 'README keeps only the shortest probes' -Message "README must avoid duplicating the full operator runbook"

Write-Host "operator command doc test passed"
