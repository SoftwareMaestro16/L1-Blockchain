$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$scriptPath = Join-Path $RepoRoot "scripts\demo\prototype-demo.ps1"
$docPath = Join-Path $RepoRoot "docs\prototype-demo.md"

function Assert-True {
  param(
    [bool]$Condition,
    [string]$Message
  )

  if (-not $Condition) {
    throw $Message
  }
}

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

Assert-True (Test-Path -LiteralPath $scriptPath) "missing prototype demo script"
Assert-True (Test-Path -LiteralPath $docPath) "missing prototype demo docs"

$scriptText = Get-Content -Raw -LiteralPath $scriptPath
Assert-Contains -Text $scriptText -Pattern '\[switch\]\$Check' -Message "demo script must expose -Check mode"
Assert-Contains -Text $scriptText -Pattern 'keyring-backend", "test"' -Message "demo script must use local test keyring"
Assert-Contains -Text $scriptText -Pattern "ChainId must contain 'local'" -Message "demo script must enforce local chain id"
Assert-Contains -Text $scriptText -Pattern 'contract-assets' -Message "demo script must include contract-assets flow"
Assert-Contains -Text $scriptText -Pattern 'dex", "swap-exact-in' -Message "demo script must include DEX swap flow"
Assert-Contains -Text $scriptText -Pattern 'REST pool 1' -Message "demo script must show REST query"
Assert-Contains -Text $scriptText -Pattern 'stop\.ps1' -Message "demo script must stop localnet by default"

$checkOutput = & $scriptPath -Check 2>&1
if ($null -ne $LASTEXITCODE -and $LASTEXITCODE -ne 0) {
  throw "prototype demo -Check failed: $($checkOutput -join "`n")"
}
$checkText = $checkOutput -join "`n"
foreach ($needle in @(
    "Aetra prototype demo check",
    "local-only: true",
    "build aetrad",
    "start 3-validator localnet",
    "send bank tx",
    "create and mint contract-assets denom",
    "create DEX pool and swap",
    "stop localnet"
  )) {
  Assert-Contains -Text $checkText -Pattern ([regex]::Escape($needle)) -Message "demo -Check output missing $needle"
}

$docText = Get-Content -Raw -LiteralPath $docPath
Assert-Contains -Text $docText -Pattern 'local-only' -Message "demo docs must mark local-only scope"
Assert-Contains -Text $docText -Pattern 'not a substitute for the e2e acceptance suite' -Message "demo docs must separate demo from tests"
Assert-Contains -Text $docText -Pattern 'tests/e2e/prototype_acceptance\.ps1' -Message "demo docs must link acceptance suite"

Write-Host "prototype demo script test passed"
