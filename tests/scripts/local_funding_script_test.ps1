param(
  [string]$Script = "scripts\localnet\fund.ps1",
  [string]$Docs = "docs\local-funding.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$ScriptPath = if ([System.IO.Path]::IsPathRooted($Script)) { $Script } else { Join-Path $RepoRoot $Script }
$DocsPath = if ([System.IO.Path]::IsPathRooted($Docs)) { $Docs } else { Join-Path $RepoRoot $Docs }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

function Assert-NotContains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -match $Pattern) { throw $Message }
}

$scriptText = Get-Content -Raw -LiteralPath $ScriptPath
$docsText = Get-Content -Raw -LiteralPath $DocsPath
$helperText = Get-Content -Raw -LiteralPath (Join-Path $RepoRoot "scripts\localnet\lib\queries.ps1")

Assert-Contains -Text $scriptText -Pattern 'non-local chain-id' -Message "funding script must reject non-local chain ids"
Assert-Contains -Text $scriptText -Pattern 'Wait-LocalnetRpc' -Message "funding script must verify local RPC status"
Assert-Contains -Text $scriptText -Pattern 'Send-LocalnetBankTx' -Message "funding script must use normal bank send helper"
Assert-Contains -Text $helperText -Pattern '"--keyring-backend",\s*"test"' -Message "funding path helper must use local test keyring"
Assert-NotContains -Text $scriptText -Pattern 'MintCoins|contract-assets mint|add-genesis-account|gentx|mnemonic|private_key' -Message "funding script must not mint, edit genesis, or expose secrets"

foreach ($term in @(
    "local-only",
    "does not mint",
    "genesis-funded",
    "aetra-local-1",
    "no mnemonics",
    "funding_smoke.ps1"
  )) {
  Assert-Contains -Text $docsText -Pattern ([regex]::Escape($term)) -Message "funding docs missing: $term"
}

Write-Host "local funding script doc test passed"
