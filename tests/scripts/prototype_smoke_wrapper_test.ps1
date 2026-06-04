param(
  [string]$Script = "tests\e2e\prototype_smoke.ps1"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$ScriptPath = if ([System.IO.Path]::IsPathRooted($Script)) { $Script } else { Join-Path $RepoRoot $Script }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) {
    throw $Message
  }
}

function Assert-NotContains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -match $Pattern) {
    throw $Message
  }
}

[scriptblock]::Create((Get-Content -Raw -LiteralPath $ScriptPath)) | Out-Null
$text = Get-Content -Raw -LiteralPath $ScriptPath

Assert-Contains -Text $text -Pattern 'Profile\s*=\s*"Smoke"' -Message "prototype smoke wrapper must force Smoke profile"
Assert-Contains -Text $text -Pattern 'prototype_acceptance\.ps1' -Message "prototype smoke wrapper must call acceptance suite"
foreach ($paramName in @(
    "OutputDir",
    "Binary",
    "ValidatorCount",
    "TimeoutSeconds",
    "BaseRPCPort",
    "Node",
    "Fees",
    "WrongFees",
    "SkipBuild",
    "KeepLogsOnFailure"
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($paramName)) -Message "prototype smoke wrapper missing $paramName passthrough"
}

Assert-NotContains -Text $text -Pattern 'Remove-Item|Stop-Process' -Message "prototype smoke wrapper must leave cleanup to localnet scripts"
Assert-NotContains -Text $text -Pattern '(?i)mnemonic|private[_-]?key|redis://|postgresql://' -Message "prototype smoke wrapper must not contain secrets"

Write-Host "prototype smoke wrapper test passed"
