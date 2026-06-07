param(
  [string]$OutputRoot = ".work\build-script-test"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$OutputRoot = if ([System.IO.Path]::IsPathRooted($OutputRoot)) {
  [System.IO.Path]::GetFullPath($OutputRoot)
} else {
  [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $OutputRoot))
}

function Assert-TestPathInsideRepo {
  param([string]$Path)
  $repo = $RepoRoot.TrimEnd('\', '/')
  $full = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  if ($full -eq $repo -or -not $full.StartsWith($repo + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing test path outside repo: $full"
  }
}

function Assert-True {
  param([bool]$Condition, [string]$Message)
  if (-not $Condition) { throw $Message }
}

Assert-TestPathInsideRepo -Path $OutputRoot
if (Test-Path -LiteralPath $OutputRoot) {
  Remove-Item -LiteralPath $OutputRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $OutputRoot | Out-Null

$fakeGo = Join-Path $OutputRoot "fake-go.ps1"
$fakeBinary = Join-Path $OutputRoot "aetrad.ps1"
@'
param([Parameter(ValueFromRemainingArguments = $true)][string[]]$GoArgs)

if ($GoArgs.Count -eq 1 -and $GoArgs[0] -eq "version") {
  Write-Output "go version go1.25.11 windows/amd64"
  exit 0
}

if ($GoArgs.Count -eq 2 -and $GoArgs[0] -eq "mod" -and $GoArgs[1] -eq "verify") {
  Write-Output "all modules verified"
  exit 0
}

if ($GoArgs.Count -eq 2 -and $GoArgs[0] -eq "mod" -and $GoArgs[1] -eq "download") {
  Write-Output "modules downloaded"
  exit 0
}

if ($GoArgs.Count -gt 0 -and $GoArgs[0] -eq "build") {
  $outIndex = [Array]::IndexOf($GoArgs, "-o")
  if ($outIndex -lt 0 -or $outIndex -eq ($GoArgs.Count - 1)) {
    Write-Error "missing -o"
    exit 1
  }
  $out = $GoArgs[$outIndex + 1]
  New-Item -ItemType Directory -Force -Path (Split-Path $out) | Out-Null
  @(
    'param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Args)',
    'if ($Args.Count -ge 1 -and $Args[0] -eq "version") {',
    '  Write-Output ''{"name":"Aetra","server_name":"aetrad","version":"prototype-test","commit":"deadbeef0000","extra_info":{"dirty":"false","cosmos_sdk_version":"v0.54.3","cometbft_version":"v0.39.3"}}''',
    '  exit 0',
    '}',
    'Write-Output "fake aetrad"'
  ) | Set-Content -LiteralPath $out
  exit 0
}

Write-Error "unexpected fake go args: $($GoArgs -join ' ')"
exit 1
'@ | Set-Content -LiteralPath $fakeGo

Push-Location $RepoRoot
try {
  $output = & .\scripts\build-aetrad.ps1 `
    -GoBinary $fakeGo `
    -Binary $fakeBinary `
    -Version prototype-test `
    -Commit deadbeef0000 `
    -MinFreeGB 0 `
    -Force *>&1
} finally {
  Pop-Location
}

Assert-True (Test-Path -LiteralPath $fakeBinary) "build script did not create binary"
$text = ($output | Out-String)
Assert-True ($text -match "go version go1.25.11") "go version was not printed"
Assert-True ($text -match "modules downloaded") "go mod download was not run"
Assert-True ($text -match "all modules verified") "go mod verify was not run"
Assert-True ($text -match [regex]::Escape($fakeBinary)) "binary path was not printed"
Assert-True ($text -match "prototype-test") "version metadata was not printed"
Assert-True ($text -match "Aetra") "built binary version was not executed"

$fallbackGo = Join-Path $OutputRoot "fake-go-fallback.ps1"
$fallbackBinary = Join-Path $OutputRoot "aetrad-fallback.ps1"
$fallbackState = Join-Path $OutputRoot "verify-fallback-state.txt"
@"
param([Parameter(ValueFromRemainingArguments = `$true)][string[]]`$GoArgs)

if (`$GoArgs.Count -eq 1 -and `$GoArgs[0] -eq "version") {
  Write-Output "go version go1.25.11 windows/amd64"
  exit 0
}

if (`$GoArgs.Count -eq 2 -and `$GoArgs[0] -eq "mod" -and `$GoArgs[1] -eq "download") {
  New-Item -ItemType Directory -Force -Path (Join-Path `$env:GOMODCACHE "github.com\cometbft\cometbft@v0.39.3") | Out-Null
  Write-Output "modules downloaded"
  exit 0
}

if (`$GoArgs.Count -eq 2 -and `$GoArgs[0] -eq "mod" -and `$GoArgs[1] -eq "verify") {
  if (!(Test-Path -LiteralPath "$fallbackState")) {
    Set-Content -LiteralPath "$fallbackState" -Value "seen"
    Write-Output "github.com/cometbft/cometbft v0.39.3: dir has been modified (`$env:GOMODCACHE\github.com\cometbft\cometbft@v0.39.3)"
    exit 1
  }
  Write-Output "all modules verified"
  exit 0
}

if (`$GoArgs.Count -gt 0 -and `$GoArgs[0] -eq "build") {
  `$outIndex = [Array]::IndexOf(`$GoArgs, "-o")
  `$out = `$GoArgs[`$outIndex + 1]
  New-Item -ItemType Directory -Force -Path (Split-Path `$out) | Out-Null
  @(
    'param([Parameter(ValueFromRemainingArguments = `$true)][string[]]`$Args)',
    'if (`$Args.Count -ge 1 -and `$Args[0] -eq "version") { Write-Output ''{"name":"Aetra"}''; exit 0 }'
  ) | Set-Content -LiteralPath `$out
  exit 0
}

Write-Error "unexpected fake go args: `$(`$GoArgs -join ' ')"
exit 1
"@ | Set-Content -LiteralPath $fallbackGo

Push-Location $RepoRoot
try {
  $fallbackOutput = & .\scripts\build-aetrad.ps1 `
    -GoBinary $fallbackGo `
    -Binary $fallbackBinary `
    -Version prototype-test `
    -Commit deadbeef0000 `
    -MinFreeGB 0 `
    -Force *>&1
} finally {
  Pop-Location
}

$fallbackText = ($fallbackOutput | Out-String)
Assert-True (Test-Path -LiteralPath $fallbackBinary) "fallback build did not create binary"
Assert-True ($fallbackText -match "known Windows extracted-tree mismatch") "CometBFT Windows verify fallback did not run"

Write-Host "build-aetrad script test passed"
