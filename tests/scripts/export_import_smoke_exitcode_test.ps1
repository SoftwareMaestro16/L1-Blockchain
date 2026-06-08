$ErrorActionPreference = "Stop"

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$scriptPath = Join-Path $repoRoot "tests\e2e\export_import_smoke.ps1"
$source = Get-Content -Raw -LiteralPath $scriptPath

if ($source -notmatch 'function\s+Assert-NativeCommandFails\s*\{') {
  throw "Assert-NativeCommandFails helper not found in export_import_smoke.ps1"
}

$functionStart = $source.IndexOf("function Assert-NativeCommandFails")
$functionEnd = $source.IndexOf("Push-Location", $functionStart)
if ($functionEnd -le $functionStart) {
  throw "could not isolate Assert-NativeCommandFails helper body"
}

$helper = $source.Substring($functionStart, $functionEnd - $functionStart)
if ($helper -notmatch '\$global:LASTEXITCODE\s*=\s*0') {
  throw "Assert-NativeCommandFails must clear `$LASTEXITCODE after expected native command failures"
}

if ($helper.IndexOf('$global:LASTEXITCODE = 0') -lt $helper.IndexOf('command failed, but output did not match')) {
  throw "Assert-NativeCommandFails must clear `$LASTEXITCODE only after validating the expected failure"
}

if ($source -notmatch 'function\s+Invoke-NativeCommand\s*\{') {
  throw "Invoke-NativeCommand helper not found in export_import_smoke.ps1"
}

$nativeFunctionStart = $source.IndexOf("function Invoke-NativeCommand")
$nativeFunctionEnd = $source.IndexOf("Push-Location", $nativeFunctionStart)
if ($nativeFunctionEnd -le $nativeFunctionStart) {
  throw "could not isolate Invoke-NativeCommand helper body"
}

$nativeHelper = $source.Substring($nativeFunctionStart, $nativeFunctionEnd - $nativeFunctionStart)
foreach ($required in @(
  '$ErrorActionPreference = "Continue"',
  '$exitCode = $LASTEXITCODE',
  '$global:LASTEXITCODE = 0'
)) {
  if ($nativeHelper -notmatch [regex]::Escape($required)) {
    throw "Invoke-NativeCommand missing required native command guard: $required"
  }
}

if ($source -notmatch 'Invoke-NativeCommand\s+-Arguments\s+@\("keys",\s*"add",\s*\$exportUserKey') {
  throw "export user key creation must use Invoke-NativeCommand"
}

Write-Host "export_import_smoke native command helpers guard expected failures and successful stderr output"
