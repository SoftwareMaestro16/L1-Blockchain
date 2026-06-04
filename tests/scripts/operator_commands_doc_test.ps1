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
    "Tokenfactory",
    "DEX",
    "Diagnose",
    "Troubleshooting",
    "Required Command Checks"
  )) {
  Assert-Contains -Text $guideText -Pattern "## $([regex]::Escape($heading))" -Message "operator guide missing heading: $heading"
}

foreach ($required in @(
    '\.\\scripts\\build-orbitalisd\.ps1',
    '\.\\scripts\\localnet\\init\.ps1',
    '\.\\scripts\\localnet\\validate-genesis\.ps1',
    '\.\\scripts\\localnet\\start\.ps1',
    '\.\\scripts\\localnet\\health\.ps1',
    '\.\\scripts\\localnet\\diagnostics\.ps1',
    '\.\\scripts\\localnet\\stop\.ps1',
    '\.\\scripts\\localnet\\reset\.ps1',
    '\.\\tests\\e2e\\prototype_smoke\.ps1',
    '\.\\tests\\e2e\\pos_smoke\.ps1',
    '\.\\tests\\e2e\\tokenfactory_smoke\.ps1',
    'build\\orbitalisd\.exe version --long --output json',
    'build\\orbitalisd\.exe query block',
    'build\\orbitalisd\.exe query bank balance',
    'build\\orbitalisd\.exe query staking validators',
    'build\\orbitalisd\.exe query fees params',
    'build\\orbitalisd\.exe query tokenfactory denom',
    'build\\orbitalisd\.exe query dex pool',
    'build\\orbitalisd\.exe tx bank send',
    'build\\orbitalisd\.exe tx staking delegate',
    'build\\orbitalisd\.exe tx tokenfactory create-denom',
    'build\\orbitalisd\.exe tx tokenfactory mint',
    'build\\orbitalisd\.exe tx tokenfactory burn',
    'build\\orbitalisd\.exe tx tokenfactory change-admin',
    'build\\orbitalisd\.exe tx dex create-pool',
    'build\\orbitalisd\.exe tx dex add-liquidity',
    'build\\orbitalisd\.exe tx dex swap-exact-in',
    'build\\orbitalisd\.exe tx dex remove-liquidity'
  )) {
  Assert-Contains -Text $guideText -Pattern $required -Message "operator guide missing command pattern: $required"
}

foreach ($variable in @('$CHAIN_ID', '$NODE', '$GRPC', '$REST', '$HOME', '$FROM', '$FEES', '$KEYRING')) {
  Assert-Contains -Text $guideText -Pattern ([regex]::Escape($variable)) -Message "operator guide missing reusable variable $variable"
}

Assert-Contains -Text $guideText -Pattern '--keyring-backend test.*local|local.*--keyring-backend test' -Message "operator guide must mark test keyring local-only"
Assert-Contains -Text $guideText -Pattern 'norb' -Message "operator guide must use norb examples"
Assert-Contains -Text $guideText -Pattern 'ORB.*display metadata only|display metadata only.*ORB' -Message "operator guide must clarify ORB display-only status"
Assert-Contains -Text $guideText -Pattern 'observability\.md' -Message "operator guide must link troubleshooting observability docs"
Assert-Contains -Text $guideText -Pattern 'operator-troubleshooting\.md' -Message "operator guide must link troubleshooting runbook"
Assert-Contains -Text $guideText -Pattern 'prototype-audit\.ps1 -Profile Fast' -Message "operator guide must include security audit check"

Assert-NotContains -Text $guideText -Pattern '(?i)print-mnemonic\s+\$?true|--print-mnemonic|mnemonic:\s*[a-z]' -Message "operator guide must not tell operators to print mnemonics"
Assert-NotContains -Text $guideText -Pattern '(?i)private[_-]?key\s*[:=]' -Message "operator guide must not include private key material"
Assert-NotContains -Text $guideText -Pattern '1000000ORB|--fees\s+\d+ORB' -Message "operator guide must not use ORB as tx fee denom"

Assert-Contains -Text $readmeText -Pattern 'docs/operator-commands\.md' -Message "README must link operator guide"
Assert-Contains -Text $readmeText -Pattern 'README keeps only the shortest probes' -Message "README must avoid duplicating the full operator runbook"

Write-Host "operator command doc test passed"
