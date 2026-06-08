param(
  [string]$ScriptsDir = "scripts\devnet"
)

$ErrorActionPreference = "Stop"
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$ScriptsPath = if ([System.IO.Path]::IsPathRooted($ScriptsDir)) { $ScriptsDir } else { Join-Path $RepoRoot $ScriptsDir }

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$one = Get-Content -Raw -LiteralPath (Join-Path $ScriptsPath "start-1-node.ps1")
$four = Get-Content -Raw -LiteralPath (Join-Path $ScriptsPath "start-4-node.ps1")
$smoke = Get-Content -Raw -LiteralPath (Join-Path $ScriptsPath "smoke.ps1")

Assert-Contains -Text $one -Pattern 'ValidatorCount\s*=\s*1' -Message "1-node devnet script must force ValidatorCount 1"
Assert-Contains -Text $one -Pattern 'scripts\\localnet\\init\.ps1' -Message "1-node devnet script must use localnet init"
Assert-Contains -Text $one -Pattern 'scripts\\localnet\\start\.ps1' -Message "1-node devnet script must use localnet start"
Assert-Contains -Text $one -Pattern 'aetra-devnet-1' -Message "1-node devnet chain id missing"

Assert-Contains -Text $four -Pattern 'ValidatorCount\s*=\s*4' -Message "4-node devnet script must force ValidatorCount 4"
Assert-Contains -Text $four -Pattern 'scripts\\localnet\\init\.ps1' -Message "4-node devnet script must use localnet init"
Assert-Contains -Text $four -Pattern 'scripts\\localnet\\start\.ps1' -Message "4-node devnet script must use localnet start"
Assert-Contains -Text $four -Pattern 'aetra-devnet-4' -Message "4-node devnet chain id missing"

Assert-Contains -Text $smoke -Pattern 'tests\\e2e\\localnet_smoke\.ps1' -Message "devnet smoke must run localnet e2e smoke"
Assert-Contains -Text $smoke -Pattern 'SkipBankTx' -Message "devnet smoke must support single-node no bank tx path"
Assert-Contains -Text $smoke -Pattern 'MinHeight\s*=\s*3' -Message "devnet smoke must wait for post-start blocks"

Write-Host "devnet script surface test passed"
